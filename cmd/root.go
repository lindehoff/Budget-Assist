/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	userHomeDir string
	debugFlag   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "budgetassist",
	Short: "A CLI tool for managing personal finances",
	Long: `Budget-Assist is a command-line interface tool that helps you
manage your personal finances efficiently. It provides features for
importing transactions, categorizing expenses, and generating reports.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging(cmd)
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
	// Set the default home directory to the user's home directory
	if userHomeDir == "" {
		userHomeDir, _ = os.UserHomeDir()
	}

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.budgetassist.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug mode")
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

// setupLogging configures the global logger with file and optional console output
func setupLogging(cmd *cobra.Command) {
	// Determine log level from configuration
	logLevel := slog.LevelInfo // Default level

	// Check configuration file for log level
	configLogLevel := strings.ToLower(viper.GetString("logging.level"))
	switch configLogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	// Override with debug flag if specified
	if debugFlag {
		logLevel = slog.LevelDebug
	}

	// Override with environment variable if specified
	if envLevel := os.Getenv("BUDGET_ASSIST_LOG_LEVEL"); envLevel != "" {
		switch strings.ToLower(envLevel) {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		}
	}

	// Create log directory if needed
	logDir := viper.GetString("logging.directory")
	if logDir == "" {
		logDir = filepath.Join(userHomeDir, ".budgetassist", "logs")
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the log directory, we'll fall back to console-only logging
		fmt.Fprintf(os.Stderr, "Error creating log directory: %v\n", err)
	}

	// Configure log file
	logFile := viper.GetString("logging.file")
	if logFile == "" {
		// Create a default log file name with timestamp
		timeStr := time.Now().Format("2006-01-02")
		logFile = fmt.Sprintf("budgetassist-%s.log", timeStr)
	}

	logFilePath := filepath.Join(logDir, logFile)

	// Create or open log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to console-only logging if file can't be opened
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
	}

	// Set up handlers based on configuration
	var handler slog.Handler

	// Create file handler if file was opened successfully
	if file != nil {
		fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})

		// If debug flag is set, use a multi handler for both console and file
		if debugFlag {
			// Create console handler with a more distinguishable format
			consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: logLevel,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					// Format the time for better readability
					if a.Key == slog.TimeKey {
						return slog.Attr{
							Key:   a.Key,
							Value: slog.StringValue(time.Now().Format("15:04:05")),
						}
					}
					// Add color to log level if terminal supports it
					if a.Key == slog.LevelKey {
						level := a.Value.Any().(slog.Level)
						levelStr := level.String()

						// Color the level based on severity
						switch level {
						case slog.LevelDebug:
							levelStr = "\033[36mDEBUG\033[0m" // Cyan
						case slog.LevelInfo:
							levelStr = "\033[32mINFO\033[0m" // Green
						case slog.LevelWarn:
							levelStr = "\033[33mWARN\033[0m" // Yellow
						case slog.LevelError:
							levelStr = "\033[31mERROR\033[0m" // Red
						}

						return slog.Attr{
							Key:   a.Key,
							Value: slog.StringValue(levelStr),
						}
					}
					return a
				},
			})

			// Create a multi-handler that sends logs to both console and file
			handler = NewMultiHandler(consoleHandler, fileHandler)
		} else {
			// Use only file handler when debug flag is not set
			handler = fileHandler
		}
	} else {
		// Fall back to console-only logging if file couldn't be opened
		// Only show logs in console if debug flag is set
		if debugFlag {
			handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: logLevel,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					// Format the time for better readability
					if a.Key == slog.TimeKey {
						return slog.Attr{
							Key:   a.Key,
							Value: slog.StringValue(time.Now().Format("15:04:05")),
						}
					}
					// Add color to log level if terminal supports it
					if a.Key == slog.LevelKey {
						level := a.Value.Any().(slog.Level)
						levelStr := level.String()

						// Color the level based on severity
						switch level {
						case slog.LevelDebug:
							levelStr = "\033[36mDEBUG\033[0m" // Cyan
						case slog.LevelInfo:
							levelStr = "\033[32mINFO\033[0m" // Green
						case slog.LevelWarn:
							levelStr = "\033[33mWARN\033[0m" // Yellow
						case slog.LevelError:
							levelStr = "\033[31mERROR\033[0m" // Red
						}

						return slog.Attr{
							Key:   a.Key,
							Value: slog.StringValue(levelStr),
						}
					}
					return a
				},
			})
		} else {
			// Create a null handler that discards all logs when debug is off and file logging failed
			handler = &NullHandler{level: logLevel}
		}
	}

	// Set global logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log startup information
	slog.Info("Application started",
		"version", viper.GetString("version"),
		"log_level", logLevel.String(),
		"config_file", viper.ConfigFileUsed())
}

// NullHandler is a slog.Handler that discards all logs
type NullHandler struct {
	level slog.Level
}

func (h *NullHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *NullHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (h *NullHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *NullHandler) WithGroup(name string) slog.Handler {
	return h
}

// MultiHandler is a custom slog.Handler that sends logs to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a handler that duplicates log records to multiple handlers
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records with lower levels.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle handles the Record. It will be called only for records that have
// a level greater than or equal to the handler's minimum level.
func (h *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of both the
// receiver's attributes and the arguments.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var handlers []slog.Handler
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return NewMultiHandler(handlers...)
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	var handlers []slog.Handler
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	return NewMultiHandler(handlers...)
}
