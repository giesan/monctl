// Package version holds the build version shared by the CLI and window.
package version

// Value is replaced by GoReleaser for packaged builds.
var Value = "dev"

func WindowTitle() string {
	return "Monitor Control " + Value
}
