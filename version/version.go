package version

var (
	// GitCommit is the current HEAD set using ldflags.
	GitCommit string

	// Version is the built softwares version.
	Version = CoreSemVer
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}

const (
	// CoreSemVer is the current version of Core.
	// It's the Semantic Version of the software.
	// XXX: Don't change the name of this variable or you will break
	// automation :)
	CoreSemVer = "0.1.0"
)

var (
	// P2PProtocol versions all p2p behaviour and msgs.
	// This includes proposer selection.
	P2PProtocol uint64 = 20

	// BlockProtocol versions all block data structures and processing.
	// This includes validity of blocks and state updates.
	BlockProtocol uint64 = 30
)
