package consensus

var (
	// DefaultLeaderID is the ID of the Replica that the protocol starts with
	DefaultLeaderID uint64 = 1
)

func (shs *SyncHS) changeLeader() {
	shs.leader = (shs.leader + 1) % shs.config.GetNumNodes()
}
