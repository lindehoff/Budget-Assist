package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// DefaultCategoriesData represents the structure of the categories.json file
type DefaultCategoriesData struct {
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
			Name         string `json:"name"`
			Description  string `json:"description"`
			Translations map[string]struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"translations"`
		} `json:"subcategories"`
	} `json:"categories"`
}

// ImportDefaultCategories imports the default categories from the categories.json file
func ImportDefaultCategories(ctx context.Context, db *gorm.DB) error {
	// Read the categories.json file
	data, err := os.ReadFile(filepath.Join("defaults", "categories.json"))
	if err != nil {
		return fmt.Errorf("failed to read categories.json: %w", err)
	}

	// Parse the JSON data
	var defaultData DefaultCategoriesData
	if err := json.Unmarshal(data, &defaultData); err != nil {
		return fmt.Errorf("failed to parse categories.json: %w", err)
	}

	// Start a transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Import category types
		for _, ct := range defaultData.CategoryTypes {
			categoryType := CategoryType{
				Name:        ct.Name,
				Description: ct.Description,
				IsMultiple:  ct.IsMultiple,
			}

			if err := tx.Create(&categoryType).Error; err != nil {
				return fmt.Errorf("failed to create category type: %w", err)
			}

			// Create translations for category type
			for lang, transl := range ct.Translations {
				translation := Translation{
					EntityID:     categoryType.ID,
					EntityType:   string(EntityTypeCategoryType),
					LanguageCode: lang,
					Name:         transl.Name,
					Description:  transl.Description,
				}
				if err := tx.Create(&translation).Error; err != nil {
					return fmt.Errorf("failed to create category type translation: %w", err)
				}
			}
		}

		// Import categories and subcategories
		for _, cat := range defaultData.Categories {
			category := Category{
				TypeID:   cat.TypeID,
				IsActive: true,
			}

			if err := tx.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to create category: %w", err)
			}

			// Create translations for category
			for lang, transl := range cat.Translations {
				translation := Translation{
					EntityID:     category.ID,
					EntityType:   string(EntityTypeCategory),
					LanguageCode: lang,
					Name:         transl.Name,
					Description:  transl.Description,
				}
				if err := tx.Create(&translation).Error; err != nil {
					return fmt.Errorf("failed to create category translation: %w", err)
				}
			}

			// Create subcategories
			for _, sub := range cat.Subcategories {
				subcategory := Subcategory{
					CategoryTypeID: category.TypeID,
					IsActive:       true,
				}

				if err := tx.Create(&subcategory).Error; err != nil {
					return fmt.Errorf("failed to create subcategory: %w", err)
				}

				// Create translations for subcategory
				for lang, transl := range sub.Translations {
					translation := Translation{
						EntityID:     subcategory.ID,
						EntityType:   string(EntityTypeSubcategory),
						LanguageCode: lang,
						Name:         transl.Name,
						Description:  transl.Description,
					}
					if err := tx.Create(&translation).Error; err != nil {
						return fmt.Errorf("failed to create subcategory translation: %w", err)
					}
				}

				// Link subcategory to category
				categorySubcategory := CategorySubcategory{
					CategoryID:    category.ID,
					SubcategoryID: subcategory.ID,
					IsActive:      true,
				}
				if err := tx.Create(&categorySubcategory).Error; err != nil {
					return fmt.Errorf("failed to create category-subcategory link: %w", err)
				}
			}
		}

		return nil
	})
}

// ImportDefaultPrompts imports the default prompts from the prompts.json file
func ImportDefaultPrompts(ctx context.Context, db *gorm.DB) error {
	// TODO: Implement default prompts import when prompts.json is ready
	return nil
}
