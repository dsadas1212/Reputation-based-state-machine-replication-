package consensus

import "errors"

var (
	errOlderBlockTime  = errors.New("timestamp older than parent")
	errInvalidView     = errors.New("Invalid View: Block has a different view than the consensus Engine")
	errInvalidCoinbase = errors.New("Invalid Coinbase: The coinbase found in the block does not match the leader in the consensus engine")
	errUnknownBlock    = errors.New("unknown block")
)
