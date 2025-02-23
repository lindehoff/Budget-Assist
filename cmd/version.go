package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Budget-Assist",
	Long: `Print the version number of Budget-Assist.
This command displays the current version of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Budget-Assist v%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
} 