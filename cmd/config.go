package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigError represents configuration-related errors
type ConfigError struct {
	Err       error
	Operation string
	Key       string
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("config %s failed for key %q: %v", e.Operation, e.Key, e.Err)
}

var (
	// ErrConfigNotFound indicates that the configuration file was not found
	ErrConfigNotFound = fmt.Errorf("configuration file not found")

	// ErrInvalidConfigFormat indicates an invalid configuration format
	ErrInvalidConfigFormat = fmt.Errorf("invalid configuration format")
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Budget-Assist configuration",
	Long: `Manage Budget-Assist configuration settings.
	
This command allows you to view, edit, and reset your configuration settings.
The configuration file is stored in your home directory as .budgetassist.yaml.`,
}

// configViewCmd represents the config view subcommand
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long:  `Display all current configuration settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := viper.AllSettings()
		if len(settings) == 0 {
			fmt.Println("No configuration settings found.")
			return nil
		}

		for k, v := range settings {
			fmt.Printf("%s: %v\n", k, v)
		}
		return nil
	},
}

// configInitCmd represents the config init subcommand
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration",
	Long:  `Create a new configuration file with default settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set default values
		viper.SetDefault("database.type", "sqlite")
		viper.SetDefault("database.path", filepath.Join(userHomeDir, ".budgetassist.db"))
		viper.SetDefault("import.default_currency", "USD")
		viper.SetDefault("export.format", "csv")

		configPath := filepath.Join(userHomeDir, ".budgetassist.yaml")
		if err := viper.SafeWriteConfigAs(configPath); err != nil {
			if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
				return &ConfigError{
					Operation: "init",
					Err:       fmt.Errorf("configuration file already exists at %s", configPath),
				}
			}
			return &ConfigError{
				Operation: "init",
				Err:       err,
			}
		}

		slog.Info("Configuration initialized", "path", configPath)
		fmt.Printf("Configuration file created at: %s\n", configPath)
		return nil
	},
}

// configResetCmd represents the config reset subcommand
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long:  `Remove existing configuration and create a new one with default settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := filepath.Join(userHomeDir, ".budgetassist.yaml")

		// Remove existing config if it exists
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return &ConfigError{
				Operation: "reset",
				Err:       err,
			}
		}

		// Initialize new config
		return configInitCmd.RunE(cmd, args)
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)

	// Initialize Viper configuration
	viper.SetConfigName(".budgetassist")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(userHomeDir)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("Failed to read config", "error", err)
		}
	}
}
