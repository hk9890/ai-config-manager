package version

import "fmt"

// Version information - injected at build time
var (
	// Version is the semantic version number
	Version = "1.13.0"
	// GitCommit is the git commit hash
	GitCommit = ""
	// BuildDate is the build date
	BuildDate = ""
)

// GetVersion returns a formatted version string
func GetVersion() string {
	if Version == "dev" || (GitCommit == "" && BuildDate == "") {
		return fmt.Sprintf("aimgr version %s", Version)
	}

	return fmt.Sprintf("aimgr version %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}
