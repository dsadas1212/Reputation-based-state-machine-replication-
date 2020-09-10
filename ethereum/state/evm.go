package state

// Here we implement ethereum state machine

import (
	"github.com/adithyabhatkajake/libsynchs/state"
	evm "github.com/ethereum/go-ethereum/core/vm"
)

// EthereumSM implements our generic State Machine object
type EthereumSM struct {
	ethVM evm.EVM
}

func (e *EthereumSM) Apply(oldState state.State) state.State {
	return oldState
}
