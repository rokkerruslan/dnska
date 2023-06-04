package endpoints

import (
	"net/netip"
)

var DefaultAddrPort = netip.MustParseAddrPort("127.0.0.1:53")
