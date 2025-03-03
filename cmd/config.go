package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Constants for log levels
const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
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

		// Print each section in a structured way
		for section, values := range settings {
			fmt.Printf("\n[%s]\n", section)
			if valuesMap, ok := values.(map[string]interface{}); ok {
				for key, value := range valuesMap {
					// Mask sensitive data
					if section == "ai" && key == "api_key" {
						if str, ok := value.(string); ok && len(str) > 8 {
							value = str[:8] + "..." + str[len(str)-4:]
						}
					}
					fmt.Printf("  %s: %v\n", key, value)
				}
			} else {
				fmt.Printf("  %v\n", values)
			}
		}
		return nil
	},
}

// configListCmd is an alias for configViewCmd
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long:  `Display all current configuration settings. This is an alias for 'config view'.`,
	RunE:  configViewCmd.RunE,
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
		viper.SetDefault("database.import_default_categories", true)
		viper.SetDefault("database.import_default_prompts", true)
		viper.SetDefault("import.default_currency", "SEK")
		viper.SetDefault("export.format", "csv")
		viper.SetDefault("ai.enabled", true)
		viper.SetDefault("ai.timeout", "10s")
		viper.SetDefault("ai.model", "gpt-4-turbo")
		viper.SetDefault("logging.level", "info")
		viper.SetDefault("logging.directory", filepath.Join(userHomeDir, ".budgetassist", "logs"))
		viper.SetDefault("logging.file", fmt.Sprintf("budgetassist-%s.log", time.Now().Format("2006-01-02")))

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

// configSetCmd represents the config set subcommand
var configSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set a configuration value",
	Long: `Set a configuration value in the configuration file.
	
For example:
  budgetassist config set logging.level debug
  budgetassist config set ai.model gpt-4-turbo
  budgetassist config set import.default_currency USD`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Special handling for certain keys
		switch key {
		case "logging.level":
			level := strings.ToLower(value)
			if level != logLevelDebug && level != logLevelInfo && level != logLevelWarn && level != logLevelError {
				return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error", value)
			}
			viper.Set(key, level)
		case "database.import_default_categories", "database.import_default_prompts", "ai.enabled":
			// Boolean values
			if strings.ToLower(value) == "true" {
				viper.Set(key, true)
			} else if strings.ToLower(value) == "false" {
				viper.Set(key, false)
			} else {
				return &ConfigError{
					Operation: "set",
					Key:       key,
					Err:       fmt.Errorf("value must be true or false"),
				}
			}
		case "ai.max_retries":
			// Integer values
			var intValue int
			_, err := fmt.Sscanf(value, "%d", &intValue)
			if err != nil {
				return &ConfigError{
					Operation: "set",
					Key:       key,
					Err:       fmt.Errorf("value must be an integer"),
				}
			}
			viper.Set(key, intValue)
		default:
			// String values
			viper.Set(key, value)
		}

		// Write the updated configuration to file
		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			configPath = filepath.Join(userHomeDir, ".budgetassist.yaml")
		}

		if err := viper.WriteConfig(); err != nil {
			// If the config file doesn't exist, create it
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfigAs(configPath); err != nil {
					return &ConfigError{
						Operation: "set",
						Key:       key,
						Err:       fmt.Errorf("failed to create config file: %w", err),
					}
				}
			} else {
				return &ConfigError{
					Operation: "set",
					Key:       key,
					Err:       fmt.Errorf("failed to write config: %w", err),
				}
			}
		}

		slog.Info("Configuration updated", "key", key, "value", value)
		fmt.Printf("Configuration updated: %s = %v\n", key, value)
		return nil
	},
}

// configGetCmd represents the config get subcommand
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get a specific configuration value.
	
This command allows you to retrieve a specific configuration value.
See 'config set --help' for a list of available configuration keys.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Check if the key exists
		if !viper.IsSet(key) {
			return &ConfigError{
				Operation: "get",
				Key:       key,
				Err:       fmt.Errorf("configuration key not found"),
			}
		}

		// Get the value
		value := viper.Get(key)

		// Mask sensitive data
		if key == "ai.api_key" {
			if str, ok := value.(string); ok && len(str) > 8 {
				value = str[:8] + "..." + str[len(str)-4:]
			}
		}

		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

// configOptionsCmd represents the config options subcommand
var configOptionsCmd = &cobra.Command{
	Use:   "options",
	Short: "List all available configuration options",
	Long: `Display all available configuration options with their descriptions, 
default values, and current values.

This command helps you understand what configuration options are available
and how they can be set using the 'config set' command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Define all available configuration options
		options := []ConfigOption{
			{
				Key:          "database.type",
				Description:  "Database type to use",
				DefaultValue: "sqlite",
				CurrentValue: viper.GetString("database.type"),
				Type:         "string",
				Example:      "sqlite",
			},
			{
				Key:          "database.path",
				Description:  "Path to the database file",
				DefaultValue: filepath.Join(userHomeDir, ".budgetassist.db"),
				CurrentValue: viper.GetString("database.path"),
				Type:         "string",
				Example:      "~/.budgetassist/data.db",
			},
			{
				Key:          "database.import_default_categories",
				Description:  "Import default categories when initializing the database",
				DefaultValue: "true",
				CurrentValue: viper.GetBool("database.import_default_categories"),
				Type:         "boolean",
				Example:      "true or false",
			},
			{
				Key:          "database.import_default_prompts",
				Description:  "Import default prompts when initializing the database",
				DefaultValue: "true",
				CurrentValue: viper.GetBool("database.import_default_prompts"),
				Type:         "boolean",
				Example:      "true or false",
			},
			{
				Key:          "import.default_currency",
				Description:  "Default currency for imports",
				DefaultValue: "SEK",
				CurrentValue: viper.GetString("import.default_currency"),
				Type:         "string",
				Example:      "SEK, USD, EUR",
			},
			{
				Key:          "export.format",
				Description:  "Default export format",
				DefaultValue: "csv",
				CurrentValue: viper.GetString("export.format"),
				Type:         "string",
				Example:      "csv, json",
			},
			{
				Key:          "ai.enabled",
				Description:  "Enable AI features",
				DefaultValue: "true",
				CurrentValue: viper.GetBool("ai.enabled"),
				Type:         "boolean",
				Example:      "true or false",
			},
			{
				Key:          "ai.api_key",
				Description:  "API key for AI service",
				DefaultValue: "",
				CurrentValue: maskSensitiveValue(viper.GetString("ai.api_key")),
				Type:         "string",
				Example:      "sk-...",
			},
			{
				Key:          "ai.model",
				Description:  "AI model to use",
				DefaultValue: "gpt-4-turbo",
				CurrentValue: viper.GetString("ai.model"),
				Type:         "string",
				Example:      "gpt-4-turbo, gpt-4o-mini",
			},
			{
				Key:          "ai.timeout",
				Description:  "Timeout for AI requests",
				DefaultValue: "10s",
				CurrentValue: viper.GetString("ai.timeout"),
				Type:         "duration",
				Example:      "10s, 30s, 1m",
			},
			{
				Key:          "ai.base_url",
				Description:  "Base URL for AI service",
				DefaultValue: "",
				CurrentValue: viper.GetString("ai.base_url"),
				Type:         "string",
				Example:      "https://api.openai.com/v1",
			},
			{
				Key:          "ai.max_retries",
				Description:  "Maximum number of retries for AI requests",
				DefaultValue: "3",
				CurrentValue: viper.GetInt("ai.max_retries"),
				Type:         "integer",
				Example:      "3, 5, 10",
			},
			{
				Key:          "logging.level",
				Description:  "Logging level",
				DefaultValue: "info",
				CurrentValue: viper.GetString("logging.level"),
				Type:         "string",
				Example:      "debug, info, warn, error",
			},
			{
				Key:          "logging.directory",
				Description:  "Directory for log files",
				DefaultValue: filepath.Join(userHomeDir, ".budgetassist", "logs"),
				CurrentValue: viper.GetString("logging.directory"),
				Type:         "string",
				Example:      "~/.budgetassist/logs",
			},
			{
				Key:          "logging.file",
				Description:  "Log file name",
				DefaultValue: fmt.Sprintf("budgetassist-%s.log", time.Now().Format("2006-01-02")),
				CurrentValue: viper.GetString("logging.file"),
				Type:         "string",
				Example:      "budgetassist.log",
			},
		}

		// Get output format
		format, _ := cmd.Flags().GetString("format")

		// Output in the requested format
		switch format {
		case "json":
			return outputOptionsAsJSON(options)
		default:
			return outputOptionsAsTable(options)
		}
	},
}

// ConfigOption represents a configuration option
type ConfigOption struct {
	Key          string
	Description  string
	DefaultValue string
	CurrentValue interface{}
	Type         string
	Example      string
}

// Helper function to mask sensitive values like API keys
func maskSensitiveValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

// Output options as a JSON document
func outputOptionsAsJSON(options []ConfigOption) error {
	jsonData, err := json.MarshalIndent(options, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal options to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// Output options as a formatted table
func outputOptionsAsTable(options []ConfigOption) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Description", "Type", "Default", "Current", "Example"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, option := range options {
		currentValue := fmt.Sprintf("%v", option.CurrentValue)

		// Handle special cases for display
		if option.Key == "database.path" || option.Key == "logging.directory" {
			// Replace home directory with ~ for better readability
			currentValue = strings.Replace(currentValue, userHomeDir, "~", 1)
			option.DefaultValue = strings.Replace(option.DefaultValue, userHomeDir, "~", 1)
		}

		table.Append([]string{
			option.Key,
			option.Description,
			option.Type,
			option.DefaultValue,
			currentValue,
			option.Example,
		})
	}

	table.Render()
	return nil
}

func init() {
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configOptionsCmd)

	// Add flags for the options command
	configOptionsCmd.Flags().String("format", "table", "Output format (table or json)")

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
