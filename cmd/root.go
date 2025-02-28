/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	userHomeDir string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "budgetassist",
	Short: "A CLI tool for managing personal finances",
	Long: `Budget-Assist is a command-line interface tool that helps you
manage your personal finances efficiently. It provides features for
importing transactions, categorizing expenses, and generating reports.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configure logging based on debug flag
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo, // Default level
		}

		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			opts.Level = slog.LevelDebug
		}

		// Configure global logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
		slog.SetDefault(logger)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.budgetassist.yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home := userHomeDir

		// Search config in home directory with name ".budgetassist" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".budgetassist")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
