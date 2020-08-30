package msg

const (
	// MaxMsgSize defines the biggest message to be ever recived in the system
	MaxMsgSize = 10000
)

// IsValid checks if the message is syntactically correct
func (bl *NoProgressBlame) IsValid() bool {
	var isValid = true
	return isValid
}

// IsValid checks if the message is syntactically correct
func (bl *EquivocationBlame) IsValid() bool {
	var isValid = true
	return isValid
}
