package cmd

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/lindehoff/Budget-Assist/internal/category"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/spf13/cobra"
)

// CategoryError represents category command-related errors
type CategoryError struct {
	Err       error
	Operation string
	Category  string
}

func (e CategoryError) Error() string {
	return fmt.Sprintf("category %s failed for %q: %v", e.Operation, e.Category, e.Err)
}

var (
	categoryManager *category.Manager
)

// categoryCmd represents the category command
var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage transaction categories",
	Long: `Manage transaction categories and their settings.
	
This command allows you to list, add, update, and remove transaction categories.
Categories can be organized hierarchically and include rules for automatic categorization.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if categoryManager == nil {
			store, err := getStore()
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}
			aiService, err := getAIService()
			if err != nil {
				return fmt.Errorf("failed to initialize AI service: %w", err)
			}
			categoryManager = category.NewManager(store, aiService, slog.Default())
		}
		return nil
	},
}

// categoryListCmd represents the category list subcommand
var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	Long: `Display all available transaction categories.
	
The output includes category details such as:
- Name and description
- Parent category (if any)
- Active status
- Associated rules`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")

		categories, err := categoryManager.ListCategories(cmd.Context(), nil)
		if err != nil {
			return &CategoryError{
				Operation: "list",
				Err:       err,
			}
		}

		switch format {
		case "json":
			return outputJSON(categories)
		case "table":
			return outputTable(categories)
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	},
}

// categoryAddCmd represents the category add subcommand
var categoryAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new category",
	Long: `Create a new transaction category with the specified details.
	
The category can include:
- Name and description
- Parent category
- Rules for automatic categorization
- Monthly budget amount`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		parent, _ := cmd.Flags().GetString("parent")

		req := category.CreateCategoryRequest{
			Name:        name,
			Description: description,
			TypeID:      1, // Default to expense type
		}

		if parent != "" {
			// TODO: Implement parent category lookup
			return fmt.Errorf("parent category support not yet implemented")
		}

		cat, err := categoryManager.CreateCategory(cmd.Context(), req)
		if err != nil {
			return &CategoryError{
				Operation: "create",
				Category:  name,
				Err:       err,
			}
		}

		fmt.Printf("Successfully created category %q with ID %d\n", cat.Name, cat.ID)
		return nil
	},
}

// categoryUpdateCmd represents the category update subcommand
var categoryUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an existing category",
	Long: `Modify an existing transaction category.
	
You can update:
- Name and description
- Parent category
- Active status
- Budget amount and color`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return &CategoryError{
				Operation: "update",
				Category:  args[0],
				Err:       fmt.Errorf("invalid category ID: %w", err),
			}
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		active, _ := cmd.Flags().GetBool("active")
		parent, _ := cmd.Flags().GetString("parent")

		// Ensure at least one flag is set
		if name == "" && description == "" && !cmd.Flags().Changed("active") &&
			parent == "" {
			return &CategoryError{
				Operation: "update",
				Category:  args[0],
				Err:       fmt.Errorf("at least one field must be specified for update"),
			}
		}

		req := category.UpdateCategoryRequest{
			Name:        name,
			Description: description,
		}

		if cmd.Flags().Changed("active") {
			req.IsActive = &active
		}

		if parent != "" {
			// TODO: Implement parent category support
			return fmt.Errorf("parent category support not yet implemented")
		}

		cat, err := categoryManager.UpdateCategory(cmd.Context(), id, req)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully updated category %q (ID: %d)\n", cat.Name, cat.ID)
		return nil
	},
}

// categoryDeleteCmd represents the category delete subcommand
var categoryDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a category",
	Long: `Remove a transaction category.
	
This will mark the category as inactive rather than permanently deleting it.
This ensures that historical transactions maintain their categorization.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return &CategoryError{
				Operation: "delete",
				Category:  args[0],
				Err:       fmt.Errorf("invalid category ID: %w", err),
			}
		}

		// Get the category first to show its name in the confirmation
		cat, err := categoryManager.GetCategoryByID(cmd.Context(), id)
		if err != nil {
			return err
		}

		// Ask for confirmation unless --force is used
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete category %q (ID: %d)? [y/N] ", cat.Name, cat.ID)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Operation cancelled")
				return nil
			}
		}

		// Delete the category (mark as inactive)
		inactive := false
		req := category.UpdateCategoryRequest{
			IsActive: &inactive,
		}

		cat, err = categoryManager.UpdateCategory(cmd.Context(), id, req)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully deleted category %q (ID: %d)\n", cat.Name, cat.ID)
		return nil
	},
}

func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryAddCmd)
	categoryCmd.AddCommand(categoryUpdateCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)
	rootCmd.AddCommand(categoryCmd)

	// Add flags for the list command
	categoryListCmd.Flags().StringP("format", "f", "table", "Output format (table|json)")

	// Add flags for the add command
	categoryAddCmd.Flags().StringP("description", "d", "", "Category description")
	categoryAddCmd.Flags().StringP("parent", "p", "", "Parent category name")
	categoryAddCmd.Flags().Float64P("budget", "b", 0, "Monthly budget amount")
	categoryAddCmd.Flags().StringP("color", "c", "", "Category color (hex)")

	// Mark required flags
	if err := categoryAddCmd.MarkFlagRequired("description"); err != nil {
		fmt.Printf("failed to mark description flag as required: %v\n", err)
	}

	// Add flags for the update command
	categoryUpdateCmd.Flags().StringP("name", "n", "", "New category name")
	categoryUpdateCmd.Flags().StringP("description", "d", "", "New category description")
	categoryUpdateCmd.Flags().BoolP("active", "a", true, "Set category active status")
	categoryUpdateCmd.Flags().StringP("parent", "p", "", "New parent category name")
	categoryUpdateCmd.Flags().Float64P("budget", "b", 0, "New monthly budget amount")
	categoryUpdateCmd.Flags().StringP("color", "c", "", "New category color (hex)")

	// Add flags for the delete command
	categoryDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func outputJSON(categories []db.Category) error {
	return printJSON(categories)
}

func outputTable(categories []db.Category) error {
	table := newTable()
	table.SetHeader([]string{"ID", "Name", "Description", "Type", "Active"})

	for _, cat := range categories {
		active := "✓"
		if !cat.IsActive {
			active = "✗"
		}
		table.Append([]string{
			fmt.Sprintf("%d", cat.ID),
			cat.Name,
			cat.Description,
			fmt.Sprintf("%d", cat.TypeID),
			active,
		})
	}

	table.Render()
	return nil
}
