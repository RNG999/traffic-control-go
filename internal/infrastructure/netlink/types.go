package netlink

// DetailedQdiscStats represents detailed qdisc statistics
type DetailedQdiscStats struct {
	BasicStats   QdiscStats
	QueueLength  uint32
	Backlog      uint32
	BacklogBytes uint64
	// Rate information
	BytesPerSecond   uint64
	PacketsPerSecond uint64
	// HTB specific
	HTBStats *HTBQdiscStats
}

// HTBQdiscStats represents HTB-specific statistics
type HTBQdiscStats struct {
	DirectPackets uint32
	Version       uint32
}

// DetailedClassStats represents detailed class statistics
type DetailedClassStats struct {
	BasicStats ClassStats
	// HTB specific
	HTBStats *HTBClassStats
}

// HTBClassStats represents HTB class-specific statistics
type HTBClassStats struct {
	Lends   uint32
	Borrows uint32
	Giants  uint32
	Tokens  uint32
	CTokens uint32
	Rate    uint64
	Ceil    uint64
	Level   uint32
}
