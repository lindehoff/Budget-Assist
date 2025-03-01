package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"gorm.io/gorm"
)

// TranslationData represents the data needed for a translation
type TranslationData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CategoryImporter defines the interface for importing categories
type CategoryImporter interface {
	CreateCategory(ctx context.Context, name, description string, typeID uint, translations map[string]TranslationData, subcategoryIDs []uint) (*Category, error)
	CreateSubcategory(ctx context.Context, name, description string, isSystem bool, translations map[string]TranslationData) (*Subcategory, error)
	CreateCategoryType(ctx context.Context, categoryType *CategoryType) error
	CreateTranslation(ctx context.Context, translation *Translation) error
}

// DefaultCategoriesData represents the structure of the categories.json file
type DefaultCategoriesData struct {
	CategoryTypes []struct {
		Name         string                     `json:"name"`
		Description  string                     `json:"description"`
		IsMultiple   bool                       `json:"isMultiple"`
		Translations map[string]TranslationData `json:"translations,omitempty"`
	} `json:"categoryTypes"`
	Subcategories []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	} `json:"subcategories"`
	Categories []struct {
		Name          string                     `json:"name"`
		Description   string                     `json:"description"`
		Type          string                     `json:"type"`
		Translations  map[string]TranslationData `json:"translations,omitempty"`
		Subcategories []string                   `json:"subcategories"`
	} `json:"categories"`
}

// DefaultPromptsData represents the structure of the prompts.json file
type DefaultPromptsData struct {
	Prompts []struct {
		Type         string `json:"type"`
		Name         string `json:"name"`
		Translations map[string]struct {
			Name         string `json:"name"`
			SystemPrompt string `json:"system_prompt"`
			UserPrompt   string `json:"user_prompt"`
		} `json:"translations"`
		Version  string `json:"version"`
		IsActive bool   `json:"is_active"`
	} `json:"prompts"`
}

// importCategoryTypes imports category types and returns a map of type names to IDs
func importCategoryTypes(db *gorm.DB, types []struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	IsMultiple   bool                       `json:"isMultiple"`
	Translations map[string]TranslationData `json:"translations,omitempty"`
}) (map[string]uint, error) {
	typeMap := make(map[string]uint)
	for _, ct := range types {
		// Check if a category type with the same name already exists
		var existingType CategoryType
		if err := db.Where("name = ?", ct.Name).First(&existingType).Error; err == nil {
			return nil, fmt.Errorf("failed to create category type: duplicate name %s", ct.Name)
		}

		categoryType := CategoryType{
			Name:        ct.Name,
			Description: ct.Description,
			IsMultiple:  ct.IsMultiple,
		}
		if err := db.Create(&categoryType).Error; err != nil {
			return nil, fmt.Errorf("failed to create category type: %w", err)
		}
		typeMap[ct.Name] = categoryType.ID
	}
	return typeMap, nil
}

// importSubcategories imports subcategories and their tags and returns a map of subcategory names to IDs
func importSubcategories(db *gorm.DB, subcategories []struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}) (map[string]uint, error) {
	subcategoryMap := make(map[string]uint)
	tagMap := make(map[string]uint)

	// First, load all existing tags into the map
	var existingTags []Tag
	if err := db.Find(&existingTags).Error; err != nil {
		return nil, fmt.Errorf("failed to list existing tags: %w", err)
	}
	for _, tag := range existingTags {
		tagMap[tag.Name] = tag.ID
	}

	for _, sc := range subcategories {
		// Check if subcategory already exists
		var existingSubcategory Subcategory
		if err := db.Where("name = ?", sc.Name).First(&existingSubcategory).Error; err == nil {
			// If it exists, use its ID
			subcategoryMap[sc.Name] = existingSubcategory.ID
			continue
		}

		// Create subcategory
		subcategory := Subcategory{
			Name:        sc.Name,
			Description: sc.Description,
			IsSystem:    true,
			IsActive:    true,
		}
		if err := db.Create(&subcategory).Error; err != nil {
			return nil, fmt.Errorf("failed to create subcategory %s: %w", sc.Name, err)
		}
		subcategoryMap[sc.Name] = subcategory.ID

		// Create or get tags and link them to subcategory
		for _, tagName := range sc.Tags {
			tagID, ok := tagMap[tagName]
			if !ok {
				// Create new tag if it doesn't exist
				tag := Tag{
					Name: tagName,
				}
				if err := db.Create(&tag).Error; err != nil {
					return nil, fmt.Errorf("failed to create tag %s: %w", tagName, err)
				}
				tagID = tag.ID
				tagMap[tagName] = tagID
			}

			// Link tag to subcategory
			if err := db.Exec("INSERT INTO subcategory_tags (subcategory_id, tag_id) VALUES (?, ?)", subcategory.ID, tagID).Error; err != nil {
				return nil, fmt.Errorf("failed to link tag %s to subcategory %s: %w", tagName, sc.Name, err)
			}
		}
	}
	return subcategoryMap, nil
}

// importCategories imports categories and their subcategories
func importCategories(db *gorm.DB, categories []struct {
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	Type          string                     `json:"type"`
	Translations  map[string]TranslationData `json:"translations,omitempty"`
	Subcategories []string                   `json:"subcategories"`
}, typeMap map[string]uint, subcategoryMap map[string]uint) error {
	for _, cat := range categories {
		// Get the category type ID from the map
		typeID, ok := typeMap[cat.Type]
		if !ok {
			return fmt.Errorf("failed to create category %s: category type %s not found", cat.Name, cat.Type)
		}

		// Check if a category with the same name already exists
		var existingCategory Category
		if err := db.Where("name = ?", cat.Name).First(&existingCategory).Error; err == nil {
			// If it exists, skip creating it
			continue
		}

		category := Category{
			Name:        cat.Name,
			Description: cat.Description,
			TypeID:      typeID,
			IsActive:    true,
		}

		if err := db.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to create category %s: %w", cat.Name, err)
		}

		// Create and link subcategories
		for _, subcatName := range cat.Subcategories {
			// Get the subcategory ID from the map
			subcatID, ok := subcategoryMap[subcatName]
			if !ok {
				// Create the subcategory if it doesn't exist
				subcategory := Subcategory{
					Name:        subcatName,
					Description: subcatName, // Use name as description for now
					IsSystem:    true,
					IsActive:    true,
				}
				if err := db.Create(&subcategory).Error; err != nil {
					return fmt.Errorf("failed to create subcategory %s: %w", subcatName, err)
				}
				subcatID = subcategory.ID
				subcategoryMap[subcatName] = subcatID
			}

			// Link the subcategory to the category
			link := CategorySubcategory{
				CategoryID:    category.ID,
				SubcategoryID: subcatID,
				IsActive:      true,
			}
			if err := db.Create(&link).Error; err != nil {
				return fmt.Errorf("failed to link subcategory %s to category %s: %w", subcatName, cat.Name, err)
			}
		}
	}
	return nil
}

// ImportDefaultCategories imports the default categories from the categories.json file
// only if no categories exist in the database
func ImportDefaultCategories(ctx context.Context, db *gorm.DB) error {
	// Check if any category types already exist
	var count int64
	if err := db.Model(&CategoryType{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing categories: %w", err)
	}

	// If categories already exist, skip import
	if count > 0 {
		return nil
	}

	// Read and parse the categories.json file
	defaultData, err := readDefaultCategoriesFile()
	if err != nil {
		return err
	}

	// Import category types
	typeMap, err := importCategoryTypes(db, defaultData.CategoryTypes)
	if err != nil {
		return err
	}

	// Import subcategories
	subcategoryMap, err := importSubcategories(db, defaultData.Subcategories)
	if err != nil {
		return err
	}

	// Import categories and their subcategories
	if err := importCategories(db, defaultData.Categories, typeMap, subcategoryMap); err != nil {
		return err
	}

	return nil
}

func readDefaultCategoriesFile() (*DefaultCategoriesData, error) {
	// Try to find the categories.json file in different locations
	locations := []string{
		"defaults/categories.json", // Check in the defaults directory first
		"categories.json",          // Then check in the current directory
	}

	var data []byte
	var lastErr error
	fileFound := false
	for _, loc := range locations {
		data, lastErr = os.ReadFile(loc)
		if lastErr == nil {
			fileFound = true
			break
		}
	}

	if !fileFound {
		return nil, fmt.Errorf("failed to read categories file: no file found in any location")
	}

	var categoriesData DefaultCategoriesData
	if err := json.Unmarshal(data, &categoriesData); err != nil {
		return nil, fmt.Errorf("failed to parse categories file: %w", err)
	}

	// Validate the data
	if len(categoriesData.CategoryTypes) == 0 {
		return nil, fmt.Errorf("failed to read categories file: no category types found")
	}

	return &categoriesData, nil
}

// ImportDefaultPrompts imports the default prompts from the prompts.json file
// only if no prompts exist in the database
func ImportDefaultPrompts(ctx context.Context, db *gorm.DB) error {
	// Check if any prompts already exist
	var count int64
	if err := db.Model(&Prompt{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing prompts: %w", err)
	}

	// If prompts already exist, skip import
	if count > 0 {
		return nil
	}

	// Read and parse the prompts.json file
	defaultData, err := readDefaultPromptsFile()
	if err != nil {
		return err
	}

	// Import prompts
	for _, p := range defaultData.Prompts {
		// Use English translation as default
		enTrans := p.Translations["en"]
		if enTrans.Name == "" {
			// If no English translation, use the default name
			enTrans.Name = p.Name
		}

		prompt := Prompt{
			Type:         p.Type,
			Name:         enTrans.Name,
			Description:  "", // No description in the current format
			SystemPrompt: enTrans.SystemPrompt,
			UserPrompt:   enTrans.UserPrompt,
			Version:      p.Version,
			IsActive:     p.IsActive,
		}
		if err := db.Create(&prompt).Error; err != nil {
			return fmt.Errorf("failed to create prompt %s: %w", p.Name, err)
		}
	}

	return nil
}

func readDefaultPromptsFile() (*DefaultPromptsData, error) {
	// Try to find the prompts.json file in different locations
	locations := []string{
		"defaults/prompts.json", // Check in the defaults directory first
		"prompts.json",          // Then check in the current directory
	}

	var data []byte
	var lastErr error
	fileFound := false
	for _, loc := range locations {
		data, lastErr = os.ReadFile(loc)
		if lastErr == nil {
			fileFound = true
			break
		}
	}

	if !fileFound {
		return nil, fmt.Errorf("failed to read prompts file: no file found in any location")
	}

	var promptsData DefaultPromptsData
	if err := json.Unmarshal(data, &promptsData); err != nil {
		return nil, fmt.Errorf("failed to parse prompts file: %w", err)
	}

	// Validate the data
	if len(promptsData.Prompts) == 0 {
		return nil, fmt.Errorf("failed to read prompts file: no prompts found")
	}

	return &promptsData, nil
}
