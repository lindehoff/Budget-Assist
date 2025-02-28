package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/lindehoff/Budget-Assist/internal/category"
	"github.com/lindehoff/Budget-Assist/internal/db"
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

		categories, err := categoryManager.ListCategories(cmd.Context(), nil)
		if err != nil {
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
- Type ID

Optional:
- Translations (will be prompted)
- Instance identifier
- Subcategories to link`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		typeID, _ := cmd.Flags().GetUint("type")
		instanceID, _ := cmd.Flags().GetString("instance-id")

		translations := make(map[string]category.TranslationData)
		// Always add English as default
		translations[db.LangEN] = category.TranslationData{
			Name:        name,
			Description: description,
		}

		req := category.CreateCategoryRequest{
			Name:               name,
			Description:        description,
			TypeID:             typeID,
			InstanceIdentifier: instanceID,
			Translations:       translations,
		}

		cat, err := categoryManager.CreateCategory(cmd.Context(), req)
		if err != nil {
			return &CategoryError{
				Operation: "create",
				Resource:  name,
				Err:       err,
			}
		}

		fmt.Printf("Successfully created category %q with ID %d\n", cat.Name, cat.ID)

		// Prompt for additional translations
		fmt.Print("Do you want to add translations for this category? (y/n): ")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			return &CategoryError{
				Operation: "read_input",
				Resource:  "translation_prompt",
				Err:       err,
			}
		}

		if response == "y" || response == "Y" {
			if err := addTranslations(cmd.Context(), cat.ID, "category"); err != nil {
				return err
			}
		}

		return nil
	},
}

// subcategoryAddCmd represents the subcategory add subcommand
var subcategoryAddCmd = &cobra.Command{
	Use:   "add-sub [name]",
	Short: "Add a new subcategory",
	Long: `Create a new subcategory that can be linked to multiple categories.
	
Required:
- Name
- Description
- At least one translation

Optional:
- Instance identifier
- Categories to link
- System flag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		instanceID, _ := cmd.Flags().GetString("instance-id")
		isSystem, _ := cmd.Flags().GetBool("system")
		categories, _ := cmd.Flags().GetStringArray("category")

		translations := make(map[string]category.TranslationData)
		// Always add English as default
		translations[db.LangEN] = category.TranslationData{
			Name:        name,
			Description: description,
		}

		// Convert category names to IDs
		var categoryIDs []uint
		if len(categories) > 0 {
			// TODO: Implement category lookup by name
			return fmt.Errorf("category lookup by name not yet implemented")
		}

		req := category.CreateSubcategoryRequest{
			IsSystem:           isSystem,
			InstanceIdentifier: instanceID,
			Categories:         categoryIDs,
			Translations:       translations,
		}

		subcat, err := categoryManager.CreateSubcategory(cmd.Context(), req)
		if err != nil {
			return &CategoryError{
				Operation: "create_subcategory",
				Resource:  name,
				Err:       err,
			}
		}

		fmt.Printf("Successfully created subcategory %q with ID %d\n", subcat.GetName(db.LangEN), subcat.ID)

		// Prompt for additional translations
		fmt.Print("Do you want to add translations for this subcategory? (y/n): ")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			return &CategoryError{
				Operation: "read_input",
				Resource:  "translation_prompt",
				Err:       err,
			}
		}

		if response == "y" || response == "Y" {
			if err := addTranslations(cmd.Context(), subcat.ID, "subcategory"); err != nil {
				return err
			}
		}

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

		// Ensure at least one flag is set
		if name == "" && description == "" && !cmd.Flags().Changed("active") &&
			instanceID == "" && len(addSubs) == 0 && len(removeSubs) == 0 {
			return &CategoryError{
				Operation: "update",
				Resource:  args[0],
				Err:       fmt.Errorf("at least one field must be specified for update"),
			}
		}

		// Convert subcategory names to IDs
		var addSubIDs, removeSubIDs []uint
		// TODO: Implement subcategory lookup by name

		req := category.UpdateCategoryRequest{
			Name:                name,
			Description:         description,
			InstanceIdentifier:  instanceID,
			AddSubcategories:    addSubIDs,
			RemoveSubcategories: removeSubIDs,
		}

		if cmd.Flags().Changed("active") {
			req.IsActive = &active
		}

		cat, err := categoryManager.UpdateCategory(cmd.Context(), id, req)
		if err != nil {
			return &CategoryError{
				Operation: "update",
				Resource:  fmt.Sprintf("id=%d", id),
				Err:       err,
			}
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
				Resource:  args[0],
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
			return err
		}

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
		filePath := args[0]

		data, err := os.ReadFile(filePath)
		if err != nil {
			return &CategoryError{
				Operation: "import",
				Resource:  filePath,
				Err:       fmt.Errorf("failed to read file: %w", err),
			}
		}

		var importData struct {
			CategoryTypes []struct {
				Name         string `json:"name"`
				Description  string `json:"description"`
				IsMultiple   bool   `json:"isMultiple"`
				Translations map[string]struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"translations"`
			} `json:"categoryTypes"`
			Categories []struct {
				Name         string `json:"name"`
				Description  string `json:"description"`
				TypeID       uint   `json:"typeId"`
				Translations map[string]struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"translations"`
				Subcategories []struct {
					Translations map[string]struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"translations"`
				} `json:"subcategories"`
			} `json:"categories"`
		}

		if err := json.Unmarshal(data, &importData); err != nil {
			return &CategoryError{
				Operation: "import",
				Resource:  filePath,
				Err:       fmt.Errorf("failed to parse JSON: %w", err),
			}
		}

		// First create category types
		typeMap := make(map[string]uint) // name -> ID mapping
		for _, ct := range importData.CategoryTypes {
			// Create translations map
			translations := make(map[string]category.TranslationData)
			for lang, trans := range ct.Translations {
				translations[lang] = category.TranslationData{
					Name:        trans.Name,
					Description: trans.Description,
				}
			}

			// Add English translation from the name/description fields
			translations[db.LangEN] = category.TranslationData{
				Name:        ct.Name,
				Description: ct.Description,
			}

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

			// Create translations
			for lang, trans := range translations {
				translation := &db.Translation{
					EntityID:     categoryType.ID,
					EntityType:   string(db.EntityTypeCategoryType),
					LanguageCode: lang,
					Name:         trans.Name,
					Description:  trans.Description,
				}
				if err := categoryManager.CreateTranslation(cmd.Context(), translation); err != nil {
					return &CategoryError{
						Operation: "create_translation",
						Resource:  fmt.Sprintf("category_type_%d", categoryType.ID),
						Err:       err,
					}
				}
			}

			typeMap[ct.Name] = categoryType.ID
			fmt.Printf("Successfully imported category type %q\n", categoryType.GetTranslation(db.LangEN))
		}

		// Then create subcategories
		subcategoryMap := make(map[string]uint) // name -> ID mapping
		for _, cat := range importData.Categories {
			for _, subcat := range cat.Subcategories {
				translations := make(map[string]category.TranslationData)
				for lang, trans := range subcat.Translations {
					translations[lang] = category.TranslationData{
						Name:        trans.Name,
						Description: trans.Description,
					}
				}

				// Get the English name for reference
				enName := subcat.Translations[db.LangEN].Name

				req := category.CreateSubcategoryRequest{
					IsSystem:     true,
					Translations: translations,
				}

				subcatObj, err := categoryManager.CreateSubcategory(cmd.Context(), req)
				if err != nil {
					return &CategoryError{
						Operation: "import_subcategory",
						Resource:  enName,
						Err:       err,
					}
				}
				subcategoryMap[enName] = subcatObj.ID
				name := subcatObj.GetName(db.LangEN)
				fmt.Printf("Successfully imported subcategory %q\n", name)
			}
		}

		// Finally create categories and link subcategories
		for _, cat := range importData.Categories {
			translations := make(map[string]category.TranslationData)
			for lang, trans := range cat.Translations {
				translations[lang] = category.TranslationData{
					Name:        trans.Name,
					Description: trans.Description,
				}
			}

			// Add English translation from the name/description fields
			translations[db.LangEN] = category.TranslationData{
				Name:        cat.Name,
				Description: cat.Description,
			}

			var subcategoryIDs []uint
			for _, subcat := range cat.Subcategories {
				enName := subcat.Translations[db.LangEN].Name
				if id, ok := subcategoryMap[enName]; ok {
					subcategoryIDs = append(subcategoryIDs, id)
				}
			}

			req := category.CreateCategoryRequest{
				TypeID:        cat.TypeID,
				Translations:  translations,
				Subcategories: subcategoryIDs,
			}

			mainCat, err := categoryManager.CreateCategory(cmd.Context(), req)
			if err != nil {
				return &CategoryError{
					Operation: "import_category",
					Resource:  translations[db.LangEN].Name,
					Err:       err,
				}
			}
			name := mainCat.GetName(db.LangEN)
			fmt.Printf("Successfully imported category %q with %d subcategories\n", name, len(subcategoryIDs))
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
	categoryAddCmd.Flags().UintP("type", "t", 1, "Category type ID")
	categoryAddCmd.Flags().StringP("instance-id", "i", "", "Instance identifier")

	// Add subcategory flags
	subcategoryAddCmd.Flags().StringP("description", "d", "", "Subcategory description")
	subcategoryAddCmd.Flags().StringP("instance-id", "i", "", "Instance identifier")
	subcategoryAddCmd.Flags().BoolP("system", "s", false, "Mark as system subcategory")
	subcategoryAddCmd.Flags().StringArrayP("category", "c", nil, "Categories to link to")

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
			cat.GetName(db.LangEN),
			cat.GetDescription(db.LangEN),
			fmt.Sprintf("%d", cat.TypeID),
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
		table.SetHeader([]string{"Type", "ID", "Name", "Description", "Active"})

		for _, cat := range categories {
			active := formatActive(cat.IsActive)
			table.Append([]string{
				"Category",
				fmt.Sprintf("%d", cat.ID),
				cat.GetName(db.LangEN),
				cat.GetDescription(db.LangEN),
				active,
			})

			// Add subcategories under their parent category
			for _, catSub := range cat.Subcategories {
				if !catSub.IsActive {
					continue
				}
				sub := catSub.Subcategory
				if sub.IsActive {
					active := formatActive(sub.IsActive)
					table.Append([]string{
						"  ↳ Subcategory", // Indented with arrow to show hierarchy
						fmt.Sprintf("%d", sub.ID),
						sub.GetName(db.LangEN),
						sub.GetDescription(db.LangEN),
						active,
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

func addTranslations(ctx context.Context, entityID uint, entityType string) error {
	fmt.Println("Adding translations (enter empty language code to finish):")
	for {
		var langCode, name, description string

		fmt.Print("Language code (e.g., sv): ")
		_, err := fmt.Scanln(&langCode)
		if err != nil || langCode == "" {
			break
		}

		fmt.Print("Translated name: ")
		_, err = fmt.Scanln(&name)
		if err != nil {
			return &CategoryError{
				Operation: "read_input",
				Resource:  "translation_name",
				Err:       err,
			}
		}

		fmt.Print("Translated description: ")
		_, err = fmt.Scanln(&description)
		if err != nil {
			return &CategoryError{
				Operation: "read_input",
				Resource:  "translation_description",
				Err:       err,
			}
		}

		translation := &db.Translation{
			EntityID:     entityID,
			EntityType:   entityType,
			LanguageCode: langCode,
			Name:         name,
			Description:  description,
		}

		if err := categoryManager.CreateTranslation(ctx, translation); err != nil {
			return &CategoryError{
				Operation: "create_translation",
				Resource:  fmt.Sprintf("%s_%d", entityType, entityID),
				Err:       err,
			}
		}
	}

	return nil
}

func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func formatActive(isActive bool) string {
	if isActive {
		return statusActive
	}
	return statusInactive
}
