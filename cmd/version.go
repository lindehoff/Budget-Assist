package cmd

import (
	"encoding/json"
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

// VersionInfo represents the version information in a structured format
type VersionInfo struct {
	Version    string `json:"version"`
	CommitHash string `json:"commitHash"`
	BuildTime  string `json:"buildTime"`
	BuildUser  string `json:"buildUser"`
	GoVersion  string `json:"goVersion"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build information",
	Long: `Print detailed version and build information for Budget-Assist.
This includes version number, build time, commit hash, and build environment.
Use --json flag to get the output in JSON format for programmatic use.
Use --short flag to only display the version number.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		shortOutput, _ := cmd.Flags().GetBool("short")

		info := VersionInfo{
			Version:    Version,
			CommitHash: CommitHash,
			BuildTime:  BuildTime,
			BuildUser:  BuildUser,
			GoVersion:  runtime.Version(),
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
		}

		if shortOutput {
			fmt.Println(info.Version)
			return
		}

		if jsonOutput {
			jsonData, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				fmt.Printf("Error generating JSON output: %v\n", err)
				return
			}
			fmt.Println(string(jsonData))
			return
		}

		fmt.Printf("Budget-Assist Version Information\n")
		fmt.Printf("--------------------------------\n")
		fmt.Printf("Version:     %s\n", info.Version)
		fmt.Printf("Commit:      %s\n", info.CommitHash)
		fmt.Printf("Built:       %s\n", info.BuildTime)
		fmt.Printf("Built by:    %s\n", info.BuildUser)
		fmt.Printf("Go version:  %s\n", info.GoVersion)
		fmt.Printf("OS/Arch:     %s/%s\n", info.OS, info.Arch)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("json", false, "Output version information in JSON format")
	versionCmd.Flags().Bool("short", false, "Only display the version number")
}
