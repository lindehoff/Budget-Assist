package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/olekukonko/tablewriter"
)

// getStore returns a new database store instance
func getStore() (db.Store, error) {
	// TODO: Get configuration from viper
	dbPath := os.ExpandEnv("${HOME}/.budgetassist.db")
	config := &db.Config{
		DBPath: dbPath,
	}

	gormDB, err := db.Initialize(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db.NewStore(gormDB, slog.Default()), nil
}

// getAIService returns a new AI service instance
func getAIService() (ai.Service, error) {
	// TODO: Get configuration from viper
	config := ai.Config{
		BaseURL:        "https://api.openai.com",
		APIKey:         os.Getenv("OPENAI_API_KEY"),
		RequestTimeout: 30,
		MaxRetries:     3,
	}

	return ai.NewOpenAIService(config, slog.Default()), nil
}

// printJSON prints data as formatted JSON
func printJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// newTable creates a new table writer with default settings
func newTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
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
	return table
}
