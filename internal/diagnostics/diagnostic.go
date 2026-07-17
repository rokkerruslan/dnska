package diagnostics

import (
	"runtime/debug"
	"time"
)

var _StartUpTime time.Time

func init() {
	_StartUpTime = time.Now()
}

// These variables are filled at build time via -ldflags. They carry only the
// facts that runtime/debug.BuildInfo does not already provide (the build
// environment) plus a human-friendly version from `git describe`. Everything
// else (revision, commit time, Go version, platform) is read back from
// debug.BuildInfo instead of being injected twice.
var (
	_Version   string
	_Host      string
	_MachineID string
	_User      string
)

type Info struct {
	// Version is `git describe` output: a tag when available, otherwise the
	// short commit hash.
	Version string

	// Build environment — not present in debug.BuildInfo.
	BuildHost      string
	BuildMachineID string
	BuildUser      string

	// Filled by the toolchain and read back from debug.BuildInfo.
	GoVersion string
	Revision  string
	Committed string
	Modified  bool
	OS        string
	Arch      string

	Uptime string
}

func CollectInfo() Info {
	bi := &debug.BuildInfo{}
	if v, ok := debug.ReadBuildInfo(); ok {
		bi = v
	}

	return Info{
		Version:        _Version,
		BuildHost:      _Host,
		BuildMachineID: _MachineID,
		BuildUser:      _User,

		GoVersion: bi.GoVersion,
		Revision:  setting(bi, "vcs.revision"),
		Committed: setting(bi, "vcs.time"),
		Modified:  setting(bi, "vcs.modified") == "true",
		OS:        setting(bi, "GOOS"),
		Arch:      setting(bi, "GOARCH"),

		Uptime: time.Since(_StartUpTime).Round(time.Millisecond).String(),
	}
}

// setting returns the value of a debug.BuildInfo setting, or "" if absent.
func setting(bi *debug.BuildInfo, key string) string {
	for _, s := range bi.Settings {
		if s.Key == key {
			return s.Value
		}
	}

	return ""
}
