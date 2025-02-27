package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

// getStore returns a new database store instance
func getStore() (db.Store, error) {
	config := &db.Config{
		DBPath:                  viper.GetString("database.path"),
		ImportDefaultCategories: viper.GetBool("database.import_default_categories"),
		ImportDefaultPrompts:    viper.GetBool("database.import_default_prompts"),
	}

	gormDB, err := db.Initialize(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db.NewStore(gormDB, slog.Default()), nil
}

// getAIService returns a new AI service instance
func getAIService() (ai.Service, error) {
	config := ai.Config{
		BaseURL:        "https://api.openai.com",
		APIKey:         viper.GetString("ai.api_key"),
		RequestTimeout: viper.GetDuration("ai.timeout"),
		MaxRetries:     3,
	}

	store, err := getStore()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	return ai.NewOpenAIService(config, store, slog.Default()), nil
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
