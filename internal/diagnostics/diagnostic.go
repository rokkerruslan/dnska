package diagnostics

import (
	"runtime/debug"
	"strconv"
	"time"
)

var _StartUpTime time.Time

func init() {
	_StartUpTime = time.Now()
}

// That variables will be filled at build time.
var (
	_Version   string
	_Timestamp string
	_Host      string
	_MachineID string
	_User      string
	_IsDirty   string
)

type Info struct {
	Version        string
	Date           string
	BuildHost      string
	BuildMachineID string
	BuildUser      string
	BuildIsDirty   string

	Ext debug.BuildInfo

	Uptime string
}

func CollectInfo() Info {
	i, _ := strconv.ParseInt(_Timestamp, 10, 64)

	bi := debug.BuildInfo{}
	if v, ok := debug.ReadBuildInfo(); ok {
		bi = *v
	}

	return Info{
		Version:        _Version,
		Date:           time.Unix(i, 0).UTC().Format(time.RFC1123),
		BuildMachineID: _MachineID,
		BuildHost:      _Host,
		BuildUser:      _User,
		BuildIsDirty:   _IsDirty,

		Ext: bi,

		Uptime: time.Since(_StartUpTime).
			Round(time.Millisecond).
			String(),
	}
}
