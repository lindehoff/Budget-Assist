package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/lindehoff/Budget-Assist/internal/category"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Output format constants
const (
	outputFormatJSON  = "json"
	outputFormatTable = "table"
)

// Status symbols
const (
	statusActive   = "✓"
	statusInactive = "✗"
)

// CategoryError represents category command-related errors
type CategoryError struct {
	Operation string
	Resource  string
	Err       error
}

func (e CategoryError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s operation failed for %q: %v", e.Operation, e.Resource, e.Err)
	}
	return fmt.Sprintf("%s operation failed: %v", e.Operation, e.Err)
}

var (
	categoryManager *category.Manager
)

// categoryCmd represents the category command
var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage transaction categories and subcategories",
	Long: `Manage transaction categories, subcategories, and their settings.
	
This command allows you to:
- List, add, update, and remove categories and subcategories
- Link subcategories to categories
- Manage translations for both categories and subcategories
- Import category hierarchies from JSON files`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Call the parent's PersistentPreRun function first
		if parent := cmd.Root(); parent != nil && parent.PersistentPreRun != nil {
			parent.PersistentPreRun(parent, args)
		}

		if categoryManager == nil {
			store, err := getStore()
			if err != nil {
				return &CategoryError{
					Operation: "initialize",
					Resource:  "store",
					Err:       err,
				}
			}
			aiService, err := getAIService()
			if err != nil {
				return &CategoryError{
					Operation: "initialize",
					Resource:  "ai_service",
					Err:       err,
				}
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
	Long: `Display all available transaction categories and their subcategories.
	
The output includes:
- Category ID, name, and description
- Associated subcategories
- Active status
- Type information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		showSubcategories, _ := cmd.Flags().GetBool("subcategories")

		// Add debug logging
		slog.Debug("Executing category list command",
			"format", format,
			"show_subcategories", showSubcategories)

		categories, err := categoryManager.ListCategories(cmd.Context(), nil)
		if err != nil {
			slog.Error("Failed to list categories", "error", err)
			return &CategoryError{
				Operation: "list",
				Resource:  "categories",
				Err:       err,
			}
		}

		if showSubcategories {
			subcategories, err := categoryManager.ListSubcategories(cmd.Context())
			if err != nil {
				return &CategoryError{
					Operation: "list",
					Resource:  "subcategories",
					Err:       err,
				}
			}
			return outputWithSubcategories(categories, subcategories, format)
		}

		switch format {
		case outputFormatJSON:
			return outputJSON(categories)
		case outputFormatTable:
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
	
Required:
- Name
- Description
- Type
- Type ID

Optional:
- Instance identifier
- Subcategories to link`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		typeID, _ := cmd.Flags().GetUint("type-id")
		typeName, _ := cmd.Flags().GetString("type")
		instanceID, _ := cmd.Flags().GetString("instance-id")
		subcategories, _ := cmd.Flags().GetStringArray("subcategory")

		// Add debug logging
		slog.Debug("Executing category add command",
			"name", name,
			"description", description,
			"type_id", typeID,
			"type", typeName,
			"instance_id", instanceID,
			"subcategories", subcategories)

		req := category.CreateCategoryRequest{
			Name:               name,
			Description:        description,
			TypeID:             typeID,
			Type:               typeName,
			InstanceIdentifier: instanceID,
			Subcategories:      subcategories,
		}

		cat, err := categoryManager.CreateCategory(cmd.Context(), req)
		if err != nil {
			slog.Error("Failed to create category", "name", name, "error", err)
			return &CategoryError{
				Operation: "create",
				Resource:  name,
				Err:       err,
			}
		}

		slog.Info("Category created successfully", "name", cat.Name, "id", cat.ID)
		fmt.Printf("Successfully created category %q with ID %d\n", cat.Name, cat.ID)
		return nil
	},
}

// subcategoryAddCmd represents the subcategory add subcommand
var subcategoryAddCmd = &cobra.Command{
	Use:   "subcategory-add [name]",
	Short: "Add a new subcategory",
	Long: `Create a new subcategory with the specified details.
	
Required:
- Name
- Description

Optional:
- Instance identifier
- Categories to link
- Tags to attach
- System flag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		instanceID, _ := cmd.Flags().GetString("instance-id")
		categories, _ := cmd.Flags().GetStringArray("category")
		tags, _ := cmd.Flags().GetStringArray("tag")
		isSystem, _ := cmd.Flags().GetBool("system")

		// Add debug logging
		slog.Debug("Executing subcategory add command",
			"name", name,
			"description", description,
			"instance_id", instanceID,
			"categories", categories,
			"tags", tags,
			"is_system", isSystem)

		req := category.CreateSubcategoryRequest{
			Name:               name,
			Description:        description,
			InstanceIdentifier: instanceID,
			Categories:         categories,
			Tags:               tags,
			IsSystem:           isSystem,
		}

		subcat, err := categoryManager.CreateSubcategory(cmd.Context(), req)
		if err != nil {
			slog.Error("Failed to create subcategory", "name", name, "error", err)
			return &CategoryError{
				Operation: "create",
				Resource:  name,
				Err:       err,
			}
		}

		slog.Info("Subcategory created successfully", "name", subcat.Name, "id", subcat.ID, "categories", categories)
		fmt.Printf("Successfully created subcategory %q with ID %d\n", subcat.Name, subcat.ID)
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
- Active status
- Instance identifier
- Translations
- Linked subcategories`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return &CategoryError{
				Operation: "update",
				Resource:  args[0],
				Err:       fmt.Errorf("invalid category ID: %w", err),
			}
		}

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		active, _ := cmd.Flags().GetBool("active")
		instanceID, _ := cmd.Flags().GetString("instance-id")
		addSubs, _ := cmd.Flags().GetStringArray("add-subcategory")
		removeSubs, _ := cmd.Flags().GetStringArray("remove-subcategory")

		// Add debug logging
		slog.Debug("Executing category update command",
			"id", id,
			"name", name,
			"description", description,
			"active", active,
			"instance_id", instanceID,
			"add_subcategories", addSubs,
			"remove_subcategories", removeSubs)

		// Ensure at least one flag is set
		if name == "" && description == "" && !cmd.Flags().Changed("active") &&
			instanceID == "" && len(addSubs) == 0 && len(removeSubs) == 0 {
			slog.Error("Category update failed", "error", "no fields specified for update", "id", id)
			return &CategoryError{
				Operation: "update",
				Resource:  args[0],
				Err:       fmt.Errorf("at least one field must be specified for update"),
			}
		}

		// Convert subcategory IDs to names
		var addSubNames, removeSubNames []string
		for _, idStr := range addSubs {
			id, err := parseID(idStr)
			if err != nil {
				return &CategoryError{
					Operation: "update",
					Resource:  fmt.Sprintf("subcategory_id=%s", idStr),
					Err:       err,
				}
			}
			subcat, err := categoryManager.GetSubcategoryByID(cmd.Context(), id)
			if err != nil {
				return &CategoryError{
					Operation: "update",
					Resource:  fmt.Sprintf("subcategory_id=%d", id),
					Err:       err,
				}
			}
			addSubNames = append(addSubNames, subcat.Name)
		}
		for _, idStr := range removeSubs {
			id, err := parseID(idStr)
			if err != nil {
				return &CategoryError{
					Operation: "update",
					Resource:  fmt.Sprintf("subcategory_id=%s", idStr),
					Err:       err,
				}
			}
			subcat, err := categoryManager.GetSubcategoryByID(cmd.Context(), id)
			if err != nil {
				return &CategoryError{
					Operation: "update",
					Resource:  fmt.Sprintf("subcategory_id=%d", id),
					Err:       err,
				}
			}
			removeSubNames = append(removeSubNames, subcat.Name)
		}

		req := category.UpdateCategoryRequest{
			Name:                name,
			Description:         description,
			InstanceIdentifier:  instanceID,
			AddSubcategories:    addSubNames,
			RemoveSubcategories: removeSubNames,
		}

		if cmd.Flags().Changed("active") {
			req.IsActive = &active
		}

		cat, err := categoryManager.UpdateCategory(cmd.Context(), id, req)
		if err != nil {
			slog.Error("Failed to update category", "id", id, "error", err)
			return &CategoryError{
				Operation: "update",
				Resource:  fmt.Sprintf("id=%d", id),
				Err:       err,
			}
		}

		slog.Info("Successfully updated category", "id", id, "name", cat.Name)
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
			slog.Error("Category delete failed", "error", "invalid ID", "id_arg", args[0], "error_details", err)
			return &CategoryError{
				Operation: "delete",
				Resource:  args[0],
				Err:       fmt.Errorf("invalid category ID: %w", err),
			}
		}

		// Add debug logging
		slog.Debug("Executing category delete command", "id", id)

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
			_, err := fmt.Scanln(&response)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}
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
			slog.Error("Failed to delete category", "id", id, "error", err)
			return err
		}

		slog.Info("Successfully deleted category", "id", id, "name", cat.Name)
		fmt.Printf("Successfully deleted category %q (ID: %d)\n", cat.Name, cat.ID)
		return nil
	},
}

// categoryImportCmd represents the category import subcommand
var categoryImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import categories from a JSON file",
	Long: `Import categories and subcategories from a JSON file.
	
The file should contain:
- Category types with translations
- Categories with translations
- Subcategories with translations
- Category-subcategory relationships`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		importFile := args[0]

		slog.Debug("Starting category import", "file", importFile)

		data, err := os.ReadFile(importFile)
		if err != nil {
			slog.Error("Failed to read import file", "file", importFile, "error", err)
			return &CategoryError{
				Operation: "read_import_file",
				Resource:  importFile,
				Err:       err,
			}
		}

		var importData struct {
			CategoryTypes []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				IsMultiple  bool   `json:"isMultiple"`
			} `json:"categoryTypes"`
			Categories []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Type        string `json:"type"`
				Instance    string `json:"instance,omitempty"`
			} `json:"categories"`
			Subcategories []struct {
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Tags        []string `json:"tags,omitempty"`
				Categories  []string `json:"categories,omitempty"`
				IsSystem    bool     `json:"isSystem,omitempty"`
			} `json:"subcategories"`
		}

		if err := json.Unmarshal(data, &importData); err != nil {
			slog.Error("Failed to parse import data", "file", importFile, "error", err)
			return &CategoryError{
				Operation: "parse_import_data",
				Resource:  importFile,
				Err:       err,
			}
		}

		slog.Debug("Parsed import file",
			"category_types", len(importData.CategoryTypes),
			"categories", len(importData.Categories),
			"subcategories", len(importData.Subcategories))

		// First create category types
		typeMap := make(map[string]uint) // name -> ID mapping
		for _, ct := range importData.CategoryTypes {
			categoryType := &db.CategoryType{
				Name:        ct.Name,
				Description: ct.Description,
				IsMultiple:  ct.IsMultiple,
			}

			// Create the category type
			if err := categoryManager.CreateCategoryType(cmd.Context(), categoryType); err != nil {
				return &CategoryError{
					Operation: "create_category_type",
					Resource:  ct.Name,
					Err:       err,
				}
			}

			typeMap[ct.Name] = categoryType.ID
			fmt.Printf("Successfully imported category type %q\n", categoryType.Name)
		}

		// Then create subcategories
		subcategoryMap := make(map[string]uint) // name -> ID mapping
		for _, subcat := range importData.Subcategories {
			req := category.CreateSubcategoryRequest{
				Name:        subcat.Name,
				Description: subcat.Description,
				IsSystem:    subcat.IsSystem,
				Tags:        subcat.Tags,
			}

			subcatObj, err := categoryManager.CreateSubcategory(cmd.Context(), req)
			if err != nil {
				return &CategoryError{
					Operation: "import_subcategory",
					Resource:  subcat.Name,
					Err:       err,
				}
			}
			subcategoryMap[subcat.Name] = subcatObj.ID
			fmt.Printf("Successfully imported subcategory %q\n", subcatObj.Name)
		}

		// Add debug logging
		slog.Debug("Successfully imported subcategories",
			"count", len(importData.Subcategories),
			"subcategory_map", subcategoryMap)

		// Finally create categories and link subcategories
		for _, cat := range importData.Categories {
			typeID, ok := typeMap[cat.Type]
			if !ok {
				return &CategoryError{
					Operation: "import_category",
					Resource:  cat.Name,
					Err:       fmt.Errorf("category type %q not found", cat.Type),
				}
			}

			req := category.CreateCategoryRequest{
				Name:               cat.Name,
				Description:        cat.Description,
				Type:               cat.Type,
				TypeID:             typeID,
				InstanceIdentifier: cat.Instance,
			}

			mainCat, err := categoryManager.CreateCategory(cmd.Context(), req)
			if err != nil {
				return &CategoryError{
					Operation: "import_category",
					Resource:  cat.Name,
					Err:       err,
				}
			}
			fmt.Printf("Successfully imported category %q\n", mainCat.Name)
		}

		// Link subcategories to categories after all entities are created
		for _, subcat := range importData.Subcategories {
			if len(subcat.Categories) == 0 {
				continue
			}

			subcatID, ok := subcategoryMap[subcat.Name]
			if !ok {
				return &CategoryError{
					Operation: "link_subcategory",
					Resource:  subcat.Name,
					Err:       fmt.Errorf("subcategory not found"),
				}
			}

			for _, catName := range subcat.Categories {
				// Get the store from the category manager
				store := categoryManager.GetStore()
				cat, err := store.GetCategoryByName(cmd.Context(), catName)
				if err != nil {
					return &CategoryError{
						Operation: "link_subcategory",
						Resource:  fmt.Sprintf("%s -> %s", subcat.Name, catName),
						Err:       fmt.Errorf("category not found: %s", catName),
					}
				}

				link := &db.CategorySubcategory{
					CategoryID:    cat.ID,
					SubcategoryID: subcatID,
					IsActive:      true,
				}
				if err := store.CreateCategorySubcategory(cmd.Context(), link); err != nil {
					return &CategoryError{
						Operation: "link_subcategory",
						Resource:  fmt.Sprintf("%s -> %s", subcat.Name, catName),
						Err:       err,
					}
				}
				fmt.Printf("Successfully linked subcategory %q to category %q\n", subcat.Name, cat.Name)
			}
		}

		return nil
	},
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryAddCmd)
	categoryCmd.AddCommand(subcategoryAddCmd)
	categoryCmd.AddCommand(categoryUpdateCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)
	categoryCmd.AddCommand(categoryImportCmd)
	rootCmd.AddCommand(categoryCmd)

	// List command flags
	categoryListCmd.Flags().StringP("format", "f", outputFormatTable, "Output format (table|json)")
	categoryListCmd.Flags().BoolP("subcategories", "s", false, "Include subcategories in output")

	// Add category flags
	categoryAddCmd.Flags().StringP("description", "d", "", "Category description")
	categoryAddCmd.Flags().UintP("type-id", "i", 0, "Category type ID")
	categoryAddCmd.Flags().StringP("type", "t", "", "Category type name")
	categoryAddCmd.Flags().StringP("instance-id", "n", "", "Instance identifier")
	categoryAddCmd.Flags().StringArrayP("subcategory", "s", nil, "Subcategories to link")

	// Add subcategory flags
	subcategoryAddCmd.Flags().StringP("description", "d", "", "Subcategory description")
	subcategoryAddCmd.Flags().StringP("instance-id", "i", "", "Instance identifier")
	subcategoryAddCmd.Flags().BoolP("system", "s", false, "Mark as system subcategory")
	subcategoryAddCmd.Flags().StringArrayP("category", "c", nil, "Categories to link to")
	subcategoryAddCmd.Flags().StringArrayP("tag", "t", nil, "Tags to attach")

	// Update flags
	categoryUpdateCmd.Flags().StringP("name", "n", "", "New category name")
	categoryUpdateCmd.Flags().StringP("description", "d", "", "New category description")
	categoryUpdateCmd.Flags().BoolP("active", "a", true, "Set category active status")
	categoryUpdateCmd.Flags().StringP("instance-id", "i", "", "New instance identifier")
	categoryUpdateCmd.Flags().StringArrayP("add-subcategory", "", nil, "Subcategories to add")
	categoryUpdateCmd.Flags().StringArrayP("remove-subcategory", "", nil, "Subcategories to remove")

	// Delete flags
	categoryDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Mark required flags
	if err := categoryAddCmd.MarkFlagRequired("description"); err != nil {
		fmt.Printf("failed to mark description flag as required: %v\n", err)
	}
	if err := categoryAddCmd.MarkFlagRequired("type"); err != nil {
		fmt.Printf("failed to mark type flag as required: %v\n", err)
	}
	if err := categoryAddCmd.MarkFlagRequired("type-id"); err != nil {
		fmt.Printf("failed to mark type-id flag as required: %v\n", err)
	}
	if err := subcategoryAddCmd.MarkFlagRequired("description"); err != nil {
		fmt.Printf("failed to mark description flag as required: %v\n", err)
	}
}

func outputJSON(categories []db.Category) error {
	return printJSON(categories)
}

func outputTable(categories []db.Category) error {
	table := newTable()
	table.SetHeader([]string{"ID", "Name", "Description", "Type", "Active"})

	for _, cat := range categories {
		active := formatActive(cat.IsActive)
		table.Append([]string{
			fmt.Sprintf("%d", cat.ID),
			cat.Name,
			cat.Description,
			cat.Type,
			active,
		})
	}

	table.Render()
	return nil
}

func outputWithSubcategories(categories []db.Category, subcategories []db.Subcategory, format string) error {
	type output struct {
		Categories    []db.Category    `json:"categories"`
		Subcategories []db.Subcategory `json:"subcategories"`
	}

	data := output{
		Categories:    categories,
		Subcategories: subcategories,
	}

	switch format {
	case outputFormatJSON:
		return printJSON(data)
	case outputFormatTable:
		table := newTable()
		table.SetHeader([]string{"Name", "Description", "Active", "Tags"})
		table.SetAutoWrapText(true)
		table.SetColWidth(100)
		table.SetColumnAlignment([]int{
			tablewriter.ALIGN_LEFT,   // Name
			tablewriter.ALIGN_LEFT,   // Description
			tablewriter.ALIGN_CENTER, // Active
			tablewriter.ALIGN_LEFT,   // Tags
		})

		for _, cat := range categories {
			active := formatActive(cat.IsActive)
			table.Append([]string{
				cat.Name,
				cat.Description,
				active,
				"",
			})

			// Add subcategories under their parent category
			for _, catSub := range cat.Subcategories {
				if !catSub.IsActive {
					continue
				}
				sub := catSub.Subcategory
				if sub.IsActive {
					active := formatActive(sub.IsActive)
					var tags []string
					for _, tag := range sub.Tags {
						tags = append(tags, tag.Name)
					}
					table.Append([]string{
						"  ↳ " + sub.Name, // Indented with arrow to show hierarchy
						sub.Description,
						active,
						strings.Join(tags, ", "),
					})
				}
			}
		}

		table.Render()
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func formatActive(isActive bool) string {
	if isActive {
		return statusActive
	}
	return statusInactive
}

func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
