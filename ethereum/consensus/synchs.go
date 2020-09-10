// Package consensus implements Sync HotStuff for the ethereum state
// machine using ethereum source code go-ethereum
//
// We use block.difficulty, a big.Int type in ethereum to represent
// the view number v for Sync HotStuff
package consensus

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	// Importing our logging module
	"github.com/adithyabhatkajake/libchatter/log"

	// Ethereum imports
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	ethcons "github.com/ethereum/go-ethereum/consensus"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"
	ethrlp "github.com/ethereum/go-ethereum/rlp"
	ethtrie "github.com/ethereum/go-ethereum/trie"

	// Sha3 imports
	"golang.org/x/crypto/sha3"
)

var (
	allowedFutureBlockTime = 15 * time.Second
)

// Inspired from the clique implementation
var (
	extraVanity = 0                      // Number of bytes of extraData used. Currently 0
	extraSeal   = crypto.SignatureLength // We use the seal to contain the signature of the proposer

	uncleHash = ethtypes.CalcUncleHash(nil) // We always use this empty uncleHash, since uncles are meaningless in the context of Sync HotStuff
)

// ChainContext implements SyncHotStuff as an Ethereum Consensus Engine Interface
type ChainContext struct {
	// Protocol Variables
	leader ethcmn.Address // Current leader
	view   uint64         // Current view

	// Protocol Constants
	nodes map[uint64]ethcmn.Address

	// Chain information
	headersByNumber map[uint64]*ethtypes.Header
	// Ensures thread safety while dealing with multiple operations
	lock sync.Mutex
}

// GetLeader returns the leader of the current view
func (cc *ChainContext) GetLeader() ethcmn.Address {
	return cc.leader
}

// New generates new ChainContext based on Ethereum's core.ChainContext and
// consensus.Engine interfaces in order to process Ethereum transactions.
//
// TODO: Fix this to take a config as parameter and set values accordingly
func New() *ChainContext {
	return &ChainContext{
		headersByNumber: make(map[uint64]*ethtypes.Header),
		nodes:           make(map[uint64]ethcmn.Address),
	}
}

// Author returns the address of the node who created the block.
func (cc *ChainContext) Author(header *ethtypes.Header) (ethcmn.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
func (cc *ChainContext) VerifyHeader(chain ethcons.ChainHeaderReader, header *ethtypes.Header, _ bool) error {
	// // Short circuit if the header is known, or its parent is not known
	// number := header.Number.Uint64()
	// if chain.GetHeader(header.Hash(), number) != nil {
	// 	return nil
	// }
	// parent := chain.GetHeader(header.ParentHash, number-1)
	// if parent == nil {
	// 	return ethcons.ErrUnknownAncestor
	// }
	return cc.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (cc *ChainContext) VerifyHeaders(chain ethcons.ChainHeaderReader, headers []*ethtypes.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := cc.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

func (cc *ChainContext) verifyHeader(chain ethcons.ChainHeaderReader, header *ethtypes.Header, parents []*ethtypes.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		return consensus.ErrFutureBlock
	}

	if header.Time <= parent.Time {
		return errOlderBlockTime
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / ethparams.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < ethparams.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return ethcons.ErrInvalidNumber
	}
	// Verify the engine specific seal securing the block
	if seal {
		if err := cc.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	return nil
}

// Engine implements Ethereum's core.ChainContext interface. As a ChainContext
// implements the consensus.Engine interface, it is simply returned.
func (cc *ChainContext) Engine() ethcons.Engine {
	return cc
}

// SetHeader implements Ethereum's core.ChainContext interface. It sets the
// header for the given block number.
func (cc *ChainContext) SetHeader(number uint64, header *ethtypes.Header) {
	cc.headersByNumber[number] = header
}

// GetHeader implements Ethereum's core.ChainContext interface.
func (cc *ChainContext) GetHeader(_ ethcmn.Hash, number uint64) *ethtypes.Header {
	if header, ok := cc.headersByNumber[number]; ok {
		return header
	}
	return nil
}

// CalcDifficulty implements Ethereum's consensus.Engine interface. It currently
// performs a no-op since SyncHS does not need it.
//
// For example, in ethash implementation, Ethash calculates the difficulty of
// the next block through the CalcDifficulty function, and creates different
// difficulty calculation methods for the difficulty of different stages.
func (cc *ChainContext) CalcDifficulty(_ ethcons.ChainHeaderReader, _ uint64, _ *ethtypes.Header) *big.Int {
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and
// uncle rewards, setting the final state on the header
// It currently performs a no-op.
func (cc *ChainContext) Finalize(
	_ ethcons.ChainHeaderReader, _ *ethtypes.Header, _ *ethstate.StateDB,
	_ []*ethtypes.Transaction, _ []*ethtypes.Header) {
}

// FinalizeAndAssemble implements consensus.Engine, accumulating the block and
// uncle rewards, setting the final state and assembling the block.
//
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (cc *ChainContext) FinalizeAndAssemble(_ ethcons.ChainHeaderReader, header *ethtypes.Header, state *ethstate.StateDB, txs []*ethtypes.Transaction,
	uncles []*ethtypes.Header, receipts []*ethtypes.Receipt) (*ethtypes.Block, error) {
	// We have always implemented EIP158
	header.Root = state.IntermediateRoot(true)

	// We did not do anything, just set the root by deleting empty objects
	return ethtypes.NewBlock(header, txs, uncles, receipts, new(ethtrie.Trie)), nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the ethash protocol. The changes are done inline.
//
// NOTE: Nothing to do in the context of SyncHS
func (cc *ChainContext) Prepare(_ ethcons.ChainHeaderReader, header *ethtypes.Header) error {
	// Set the view number in the Difficulty field of the header
	header.Difficulty.SetUint64(cc.view)
	return nil
}

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
//
// NOTE: Nothing to do in the context of SyncHS
func (cc *ChainContext) Seal(chain ethcons.ChainHeaderReader, block *ethtypes.Block, results chan<- *ethtypes.Block, stop <-chan struct{}) error {
	header := block.Header()
	header.Nonce, header.MixDigest = ethtypes.BlockNonce{}, ethcmn.Hash{}
	select {
	case results <- block.WithSeal(header):
	default:
		log.Warn("Miner is not reading the Seal result")
	}

	return nil
}

// SealHash returns the hash of a block before it is sealed.
//
// NOTE: Nothing to do in the context of SyncHS
func (cc *ChainContext) SealHash(header *ethtypes.Header) (hash ethcmn.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	ethrlp.Encode(hasher, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra,
	})
	hasher.Sum(hash[:0])
	return hash
}

// VerifySeal implements Ethereum's consensus.Engine interface.
// In ethash, it checks to ensure that the PoW was computed correctly.
// In SyncHS, we check if the correct leader, signed it in the correct view.
func (cc *ChainContext) VerifySeal(_ ethcons.ChainHeaderReader, header *ethtypes.Header) error {
	// Check if it was signed by the correct leader in the correct view
	// header.Difficulty *big.Int => We use this to store our view number
	// header.Nonce ethtypes.Nonce
	// header.Extra []byte
	blkView := header.Difficulty.Uint64()
	if blkView != cc.view {
		return errInvalidView
	}
	if header.Coinbase != cc.leader {
		return errInvalidCoinbase
	}
	return nil
}

// VerifyUncles implements Ethereum's consensus.Engine interface. It currently
// performs a no-op.
// In SyncHS, we do not use Uncles, so we do nothing.
func (cc *ChainContext) VerifyUncles(_ ethcons.ChainReader, _ *ethtypes.Block) error {
	return nil
}

// Close implements Ethereum's consensus.Engine interface. It terminates any
// background threads maintained by the consensus engine. It currently performs
// a no-op.
func (cc *ChainContext) Close() error {
	return nil
}
