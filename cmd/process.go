package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/docprocess"
	"github.com/lindehoff/Budget-Assist/internal/pipeline"
	"github.com/lindehoff/Budget-Assist/internal/processor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var processCmd = &cobra.Command{
	Use:   "process [path]",
	Short: "Process documents for transaction extraction",
	Long: `Process documents (PDF invoices, receipts, bank statements, CSV files) for transaction extraction.
The command will:
1. Extract text from documents
2. Identify transactions using AI
3. Categorize transactions using AI
4. Store results in the database`,
	Args: cobra.ExactArgs(1),
	RunE: runProcess,
}

func init() {
	rootCmd.AddCommand(processCmd)
	processCmd.Flags().Bool("no-ai", false, "Skip AI categorization")
}

func runProcess(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Validate path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Initialize services
	logger := slog.Default()

	// Get database connection from global config
	store, err := getStore()
	if err != nil {
		return fmt.Errorf("failed to get database store: %w", err)
	}

	// Check if AI should be skipped
	skipAI, _ := cmd.Flags().GetBool("no-ai")
	var aiService ai.Service

	if !skipAI {
		// Get OpenAI config from environment or config file
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = viper.GetString("openai.api_key")
		}

		aiConfig := ai.Config{
			BaseURL:        viper.GetString("openai.base_url"),
			APIKey:         apiKey,
			RequestTimeout: viper.GetDuration("openai.request_timeout"),
			MaxRetries:     viper.GetInt("openai.max_retries"),
		}

		// Set defaults if not configured
		if aiConfig.BaseURL == "" {
			aiConfig.BaseURL = "https://api.openai.com/v1"
		}
		if aiConfig.RequestTimeout == 0 {
			aiConfig.RequestTimeout = 30 * time.Second
		}
		if aiConfig.MaxRetries == 0 {
			aiConfig.MaxRetries = 3
		}

		if aiConfig.APIKey == "" {
			return fmt.Errorf("OpenAI API key not found in environment variable OPENAI_API_KEY or config file (openai.api_key)")
		}

		// Initialize AI service
		aiService = ai.NewOpenAIService(aiConfig, store, logger)
	}

	// Initialize processors
	pdfProcessor := docprocess.NewPDFProcessor(logger)
	csvProcessor := processor.NewSEBProcessor(logger)

	// Create processing pipeline
	p := pipeline.NewPipeline(pdfProcessor, csvProcessor, aiService, store, logger)

	// Process documents
	results, err := p.ProcessDocuments(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("failed to process documents: %w", err)
	}

	// Print results
	fmt.Printf("\nProcessing Results:\n")
	fmt.Printf("==================\n")

	var totalTransactions int
	var failures int

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s: %v\n", filepath.Base(result.FilePath), result.Error)
			failures++
		} else {
			fmt.Printf("✅ %s: Found %d transactions\n",
				filepath.Base(result.FilePath),
				result.TransactionsFound)
			totalTransactions += result.TransactionsFound
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Files processed: %d\n", len(results))
	fmt.Printf("- Successful: %d\n", len(results)-failures)
	fmt.Printf("- Failed: %d\n", failures)
	fmt.Printf("- Total transactions found: %d\n", totalTransactions)

	return nil
}
