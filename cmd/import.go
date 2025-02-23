package cmd

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ImportError represents import-related errors
type ImportError struct {
	Operation string
	Source    string
	Err       error
}

func (e ImportError) Error() string {
	return fmt.Sprintf("import %s failed for source %q: %v", e.Operation, e.Source, e.Err)
}

var (
	// ErrUnsupportedFormat indicates that the import format is not supported
	ErrUnsupportedFormat = fmt.Errorf("unsupported import format")

	// ErrInvalidSource indicates that the import source is invalid
	ErrInvalidSource = fmt.Errorf("invalid import source")

	// ErrParsingFailed indicates that parsing the import file failed
	ErrParsingFailed = fmt.Errorf("failed to parse import file")
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [flags] <file>",
	Short: "Import financial data from various sources",
	Long: `Import financial transactions and data from various sources.
	
Supported formats:
- CSV files from major banks
- QIF (Quicken Interchange Format)
- OFX (Open Financial Exchange)
- PDF bank statements (experimental)

Example:
  budgetassist import --format=csv --bank=chase statement.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return &ImportError{
				Operation: "validate",
				Source:    "args",
				Err:       fmt.Errorf("file path is required"),
			}
		}

		format, _ := cmd.Flags().GetString("format")
		bank, _ := cmd.Flags().GetString("bank")
		currency, _ := cmd.Flags().GetString("currency")

		// Use default currency from config if not specified
		if currency == "" {
			currency = viper.GetString("import.default_currency")
		}

		filePath := args[0]
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(userHomeDir, filePath)
		}

		slog.Info("Starting import",
			"file", filePath,
			"format", format,
			"bank", bank,
			"currency", currency,
		)

		// TODO: Implement actual import logic in internal/core package
		fmt.Printf("Import not yet implemented. Would import %s file from %s bank in %s currency\n",
			format, bank, currency)

		return nil
	},
}

// importListCmd represents the import list subcommand
var importListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported import formats and banks",
	Long: `Display all supported import formats and bank templates.
This helps you identify the correct format and bank for your import.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Supported formats:")
		fmt.Println("  - csv  : Comma-Separated Values")
		fmt.Println("  - qif  : Quicken Interchange Format")
		fmt.Println("  - ofx  : Open Financial Exchange")
		fmt.Println("  - pdf  : PDF Bank Statements (experimental)")
		fmt.Println("\nSupported banks:")
		fmt.Println("  - chase    : Chase Bank")
		fmt.Println("  - bofa     : Bank of America")
		fmt.Println("  - wells    : Wells Fargo")
		fmt.Println("  - citi     : Citibank")
		fmt.Println("  - discover : Discover Card")
		fmt.Println("  - amex     : American Express")
		fmt.Println("  - generic  : Generic CSV format")
	},
}

func init() {
	importCmd.AddCommand(importListCmd)
	rootCmd.AddCommand(importCmd)

	// Add flags for the import command
	importCmd.Flags().StringP("format", "f", "csv", "Import format (csv, qif, ofx, pdf)")
	importCmd.Flags().StringP("bank", "b", "generic", "Bank template to use for parsing")
	importCmd.Flags().StringP("currency", "c", "", "Currency code (default from config)")
	importCmd.Flags().BoolP("dry-run", "d", false, "Validate import without saving")

	// Mark required flags
	importCmd.MarkFlagRequired("format")
}
