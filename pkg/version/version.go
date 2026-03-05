package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Version information - injected at build time via ldflags.
// When installed via "go install ...@vX.Y.Z", the version is
// automatically detected from Go's embedded build info.
var (
	// Version is the semantic version number
	Version = "dev"
	// GitCommit is the git commit hash
	GitCommit = ""
	// BuildDate is the build date
	BuildDate = ""
)

func init() {
	// If ldflags already set the version, nothing to do.
	if Version != "dev" {
		return
	}

	// Try to get version from Go's build info (set by go install ...@vX.Y.Z).
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	// Module version is "(devel)" for local builds, or "vX.Y.Z" for go install.
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = strings.TrimPrefix(info.Main.Version, "v")

		// Also extract VCS info if available.
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if len(s.Value) > 7 {
					GitCommit = s.Value[:7]
				} else {
					GitCommit = s.Value
				}
			case "vcs.time":
				BuildDate = s.Value
			}
		}
	}
}

// GetVersion returns a formatted version string.
func GetVersion() string {
	if Version == "dev" || (GitCommit == "" && BuildDate == "") {
		return fmt.Sprintf("aimgr version %s", Version)
	}

	return fmt.Sprintf("aimgr version %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}
