package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set during build using -ldflags
var (
	// Version is the current version of Budget-Assist
	Version = "dev"
	// CommitHash is the git commit hash of the build
	CommitHash = "unknown"
	// BuildTime is the time when the binary was built
	BuildTime = "unknown"
	// BuildUser is the user who built the binary
	BuildUser = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build information",
	Long: `Print detailed version and build information for Budget-Assist.
This includes version number, build time, commit hash, and build environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Budget-Assist Version Information\n")
		fmt.Printf("--------------------------------\n")
		fmt.Printf("Version:     %s\n", Version)
		fmt.Printf("Commit:      %s\n", CommitHash)
		fmt.Printf("Built:       %s\n", BuildTime)
		fmt.Printf("Built by:    %s\n", BuildUser)
		fmt.Printf("Go version:  %s\n", runtime.Version())
		fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
