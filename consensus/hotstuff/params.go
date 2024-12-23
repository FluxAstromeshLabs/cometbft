package hotstuff

const (
	FooChannel  = byte(0x80)
	BarChannel  = byte(0x81)
	PingChannel = byte(0x82)
	PongChannel = byte(0x83)
	MaxMsgSize  = 1048576 // 1MB
)
