package app

import "runtime/debug"

var (
	version = "dev"     // set by goreleaser
	commit  = "none"    //nolint:unused // set by goreleaser
	date    = "unknown" //nolint:unused // set by goreleaser
)

// GetVersion returns the version string, attempting to get it from VCS info if available
func GetVersion() string {
	// If version was set by goreleaser, use it
	if version != "dev" {
		return version
	}

	// Try to get version from build info
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				// Use short commit hash (first 8 characters)
				if len(setting.Value) >= 8 {
					return "dev-" + setting.Value[:8]
				}
				return "dev-" + setting.Value
			}
		}
	}

	// Fallback to original behavior
	return version
}
