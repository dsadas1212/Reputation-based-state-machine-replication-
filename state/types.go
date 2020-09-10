package state

// State refers to a generic state which we can query to get balances, and other state information
type State interface{}

// Machine is an object that takes a previous state and applies the change
type Machine interface {
	ApplyState(State) State
}
