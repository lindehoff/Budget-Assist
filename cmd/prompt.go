package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/spf13/cobra"
)

// PromptError represents prompt command-related errors
type PromptError struct {
	Err       error
	Operation string
	Prompt    string
}

func (e PromptError) Error() string {
	return fmt.Sprintf("prompt %s failed for %q: %v", e.Operation, e.Prompt, e.Err)
}

var (
	promptManager *ai.PromptManager
	aiService     ai.Service
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage AI prompt templates",
	Long: `Manage AI prompt templates and their settings.
	
This command allows you to list, add, update, and test prompt templates.
Prompts are used for transaction categorization and document analysis.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if promptManager == nil {
			store, err := getStore()
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}
			aiService, err = getAIService()
			if err != nil {
				return fmt.Errorf("failed to initialize AI service: %w", err)
			}
			promptManager = ai.NewPromptManager(store, slog.Default())
		}
		return nil
	},
}

// promptListCmd represents the prompt list subcommand
var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all prompt templates",
	Long: `Display all available prompt templates.
	
The output includes template details such as:
- Type and name
- Version and status
- System and user prompts (when --show-prompts is used)
- Associated rules and examples`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		showPrompts, _ := cmd.Flags().GetBool("show-prompts")

		prompts, err := promptManager.ListPrompts(cmd.Context())
		if err != nil {
			return &PromptError{
				Operation: "list",
				Prompt:    "all",
				Err:       err,
			}
		}
		if len(prompts) == 0 {
			fmt.Println("No prompt templates found")
			return nil
		}

		switch format {
		case "json":
			return printJSON(prompts)
		case "table":
			return outputPromptTable(prompts, showPrompts)
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	},
}

// promptAddCmd represents the prompt add subcommand
var promptAddCmd = &cobra.Command{
	Use:   "add [type] [name]",
	Short: "Add a new prompt template",
	Long: `Create a new prompt template with the specified details.
	
The template requires:
- Type (e.g., bill_analysis, receipt_analysis)
- Name and description
- System and user prompts`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		promptType := db.PromptType(args[0])
		name := args[1]
		systemPrompt, _ := cmd.Flags().GetString("system")
		userPrompt, _ := cmd.Flags().GetString("user")

		template := &ai.PromptTemplate{
			Type:         promptType,
			Name:         name,
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			Version:      "1.0.0",
			IsActive:     true,
		}

		if err := promptManager.UpdatePrompt(cmd.Context(), template); err != nil {
			return &PromptError{
				Operation: "create",
				Prompt:    name,
				Err:       err,
			}
		}

		fmt.Printf("Successfully created prompt template %q of type %q\n", name, promptType)
		return nil
	},
}

// promptUpdateCmd represents the prompt update subcommand
var promptUpdateCmd = &cobra.Command{
	Use:   "update [type]",
	Short: "Update an existing prompt template",
	Long: `Modify an existing prompt template.
	
You can update:
- System and user prompts
- Active status`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		promptType := db.PromptType(args[0])

		// Get the existing template
		template, err := promptManager.GetPrompt(cmd.Context(), promptType)
		if err != nil {
			return &PromptError{
				Operation: "update",
				Prompt:    string(promptType),
				Err:       err,
			}
		}

		// Update fields if flags are set
		if cmd.Flags().Changed("system") {
			system, _ := cmd.Flags().GetString("system")
			template.SystemPrompt = system
		}
		if cmd.Flags().Changed("user") {
			user, _ := cmd.Flags().GetString("user")
			template.UserPrompt = user
		}
		if cmd.Flags().Changed("active") {
			active, _ := cmd.Flags().GetBool("active")
			template.IsActive = active
		}

		if err := promptManager.UpdatePrompt(cmd.Context(), template); err != nil {
			return &PromptError{
				Operation: "update",
				Prompt:    string(promptType),
				Err:       err,
			}
		}

		fmt.Printf("Successfully updated prompt template %q\n", template.Name)
		return nil
	},
}

// promptTestCmd represents the prompt test subcommand
var promptTestCmd = &cobra.Command{
	Use:   "test [type]",
	Short: "Test a prompt template",
	Long: `Test a prompt template with sample data.
	
This command allows you to:
- Test prompt execution with sample data
- View the generated prompt
- Validate template syntax`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		promptType := db.PromptType(args[0])
		data, _ := cmd.Flags().GetString("data")

		// Parse the JSON data
		var templateData interface{}
		if err := json.Unmarshal([]byte(data), &templateData); err != nil {
			return &PromptError{
				Operation: "test",
				Prompt:    string(promptType),
				Err:       fmt.Errorf("invalid JSON data: %w", err),
			}
		}

		// Get the template
		template, err := promptManager.GetPrompt(cmd.Context(), promptType)
		if err != nil {
			return &PromptError{
				Operation: "test",
				Prompt:    string(promptType),
				Err:       err,
			}
		}

		// Execute the template with the data
		result, err := ai.ExecuteTemplate(template.SystemPrompt, templateData)
		if err != nil {
			return &PromptError{
				Operation: "test",
				Prompt:    string(promptType),
				Err:       err,
			}
		}

		fmt.Println("Generated prompt:")
		fmt.Println("----------------")
		fmt.Println(result)

		// Test with AI service if available
		if aiService != nil {
			fmt.Println("\nAI Service response:")
			fmt.Println("------------------")
			switch promptType {
			case db.TransactionCategorizationPrompt:
				// Extract description from the JSON data
				data, ok := templateData.(map[string]interface{})
				if !ok {
					return &PromptError{
						Operation: "test",
						Prompt:    string(promptType),
						Err:       fmt.Errorf("invalid data format: expected JSON object"),
					}
				}
				description, ok := data["Description"].(string)
				if !ok {
					return &PromptError{
						Operation: "test",
						Prompt:    string(promptType),
						Err:       fmt.Errorf("missing or invalid Description field"),
					}
				}
				matches, err := aiService.SuggestCategories(cmd.Context(), description)
				if err != nil {
					return &PromptError{
						Operation: "test",
						Prompt:    string(promptType),
						Err:       err,
					}
				}
				for _, match := range matches {
					fmt.Printf("Category: %s (confidence: %.2f)\n", match.Category, match.Confidence)
				}
			case db.BillAnalysisPrompt:
				// Extract content from the JSON data
				data, ok := templateData.(map[string]interface{})
				if !ok {
					return &PromptError{
						Operation: "test",
						Prompt:    string(promptType),
						Err:       fmt.Errorf("invalid data format: expected JSON object"),
					}
				}
				content, ok := data["Content"].(string)
				if !ok {
					return &PromptError{
						Operation: "test",
						Prompt:    string(promptType),
						Err:       fmt.Errorf("missing or invalid Content field"),
					}
				}
				extraction, err := aiService.ExtractDocument(cmd.Context(), &ai.Document{Content: []byte(content)})
				if err != nil {
					return &PromptError{
						Operation: "test",
						Prompt:    string(promptType),
						Err:       err,
					}
				}
				fmt.Printf("Extracted content: %s\n", extraction.Content)
			default:
				fmt.Println("Testing not implemented for this prompt type")
			}
		}

		return nil
	},
}

// promptImportCmd represents the prompt import subcommand
var promptImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import prompt templates from a JSON file",
	Long: `Import prompt templates from a JSON file.
	
The file should contain a JSON array of prompt templates with:
- Type and name
- System and user prompts
- Language translations`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Read and parse the JSON file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return &PromptError{
				Operation: "import",
				Prompt:    filePath,
				Err:       fmt.Errorf("failed to read file: %w", err),
			}
		}

		var templates struct {
			Prompts []struct {
				Type         string `json:"type"`
				Name         string `json:"name"`
				Description  string `json:"description"`
				Translations map[string]struct {
					Name         string `json:"name"`
					SystemPrompt string `json:"system_prompt"`
					UserPrompt   string `json:"user_prompt"`
				} `json:"translations"`
				Version  string `json:"version"`
				IsActive bool   `json:"is_active"`
			} `json:"prompts"`
		}
		if err := json.Unmarshal(data, &templates); err != nil {
			return &PromptError{
				Operation: "import",
				Prompt:    filePath,
				Err:       fmt.Errorf("failed to parse JSON: %w", err),
			}
		}

		// Import each prompt template
		for _, p := range templates.Prompts {
			// Use English as the default language if available, otherwise use Swedish
			var defaultLang string
			if _, ok := p.Translations["en"]; ok {
				defaultLang = "en"
			} else if _, ok := p.Translations["sv"]; ok {
				defaultLang = "sv"
			} else {
				return &PromptError{
					Operation: "import",
					Prompt:    p.Name,
					Err:       fmt.Errorf("no valid translation found (requires either English or Swedish)"),
				}
			}

			template := &ai.PromptTemplate{
				Type:         db.PromptType(p.Type),
				Name:         p.Name,
				Description:  p.Description,
				SystemPrompt: p.Translations[defaultLang].SystemPrompt,
				UserPrompt:   p.Translations[defaultLang].UserPrompt,
				Version:      p.Version,
				IsActive:     p.IsActive,
			}

			if err := promptManager.UpdatePrompt(cmd.Context(), template); err != nil {
				return &PromptError{
					Operation: "import",
					Prompt:    p.Name,
					Err:       err,
				}
			}
			fmt.Printf("Successfully imported prompt template %q of type %q\n", template.Name, template.Type)
		}

		return nil
	},
}

func outputPromptTable(prompts []*ai.PromptTemplate, showPrompts bool) error {
	table := newTable()
	if showPrompts {
		table.SetHeader([]string{"Type", "Name", "Version", "Active", "System Prompt", "User Prompt"})
	} else {
		table.SetHeader([]string{"Type", "Name", "Version", "Active"})
	}

	for _, p := range prompts {
		active := "✓"
		if !p.IsActive {
			active = "✗"
		}

		if showPrompts {
			table.Append([]string{
				string(p.Type),
				p.Name,
				p.Version,
				active,
				p.SystemPrompt,
				p.UserPrompt,
			})
		} else {
			table.Append([]string{
				string(p.Type),
				p.Name,
				p.Version,
				active,
			})
		}
	}

	table.Render()
	return nil
}

func init() {
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptAddCmd)
	promptCmd.AddCommand(promptUpdateCmd)
	promptCmd.AddCommand(promptTestCmd)
	promptCmd.AddCommand(promptImportCmd)
	rootCmd.AddCommand(promptCmd)

	// Add flags for the list command
	promptListCmd.Flags().StringP("format", "f", "table", "Output format (table|json)")
	promptListCmd.Flags().BoolP("show-prompts", "p", false, "Show the actual prompt templates")

	// Add flags for the add command
	promptAddCmd.Flags().StringP("system", "s", "", "System prompt text")
	promptAddCmd.Flags().StringP("user", "u", "", "User prompt template")

	// Mark required flags
	if err := promptAddCmd.MarkFlagRequired("system"); err != nil {
		fmt.Printf("failed to mark system flag as required: %v\n", err)
	}
	if err := promptAddCmd.MarkFlagRequired("user"); err != nil {
		fmt.Printf("failed to mark user flag as required: %v\n", err)
	}

	// Add flags for the update command
	promptUpdateCmd.Flags().StringP("system", "s", "", "New system prompt text")
	promptUpdateCmd.Flags().StringP("user", "u", "", "New user prompt template")
	promptUpdateCmd.Flags().BoolP("active", "a", true, "Set prompt active status")

	// Add flags for the test command
	promptTestCmd.Flags().StringP("data", "d", "", "Sample data in JSON format")
	if err := promptTestCmd.MarkFlagRequired("data"); err != nil {
		fmt.Printf("failed to mark data flag as required: %v\n", err)
	}
}
