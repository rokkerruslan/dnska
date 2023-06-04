package limits

// Size limits
//
// Various objects and parameters in the DNS have size
// limits. They are listed below.  Some could be easily
// changed, others are more fundamental.

const (
	// UDPPayloadSizeLimit limit max payload size. In DNS Protocol design, UDP transport
	// Block size (payload size) has been limited to 512-Bytes to optimize performance
	// whilst generating minimal network traffic.
	// Longer messages are truncated and the TC bit is set in the header.
	UDPPayloadSizeLimit = 512
	MaxLabelSize        = 63
	MaxNameSize         = 255
)
