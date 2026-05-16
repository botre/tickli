package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime/debug"
)

// Version is set at build time using -ldflags
var Version = "dev"

// CurrentVersion returns the running tickli version. It prefers the build-time
// ldflags value and falls back to the module version recorded in the binary's
// build info (set automatically for `go install`-ed binaries).
func CurrentVersion() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version",
		Long: `Display the current version of the Tickli CLI tool.
This command shows the version number, which can be set at build time
or automatically detected from build information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("tickli version", CurrentVersion())
		},
	}
	return cmd
}
