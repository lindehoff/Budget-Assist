package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		Translations map[string]TranslationData `json:"translations"`
	} `json:"categoryTypes"`
	Categories []struct {
		Name          string                     `json:"name"`
		Description   string                     `json:"description"`
		TypeID        uint                       `json:"typeId"`
		Translations  map[string]TranslationData `json:"translations"`
		Subcategories []struct {
			Name         string                     `json:"name"`
			Description  string                     `json:"description"`
			Translations map[string]TranslationData `json:"translations"`
		} `json:"subcategories"`
	} `json:"categories"`
}

// ImportDefaultCategories imports the default categories from the categories.json file
func ImportDefaultCategories(ctx context.Context, db *gorm.DB, importer CategoryImporter) error {
	// Read and parse the categories.json file
	defaultData, err := readDefaultCategoriesFile()
	if err != nil {
		return err
	}

	// Check if categories already exist to avoid duplicate import
	var count int64
	if err := db.Model(&Category{}).Count(&count).Error; err == nil && count > 0 {
		return nil
	}

	// Import category types first
	for _, ct := range defaultData.CategoryTypes {
		var categoryType CategoryType
		// Check if category type already exists
		if err := db.Where("name = ?", ct.Name).First(&categoryType).Error; err != nil {
			// Not found, so create new category type
			categoryType = CategoryType{
				Name:        ct.Name,
				Description: ct.Description,
				IsMultiple:  ct.IsMultiple,
			}
			if err := importer.CreateCategoryType(ctx, &categoryType); err != nil {
				return fmt.Errorf("failed to create category type: %w", err)
			}
			// Create translations for category type
			for lang, trans := range ct.Translations {
				translation := &Translation{
					EntityID:     categoryType.ID,
					EntityType:   string(EntityTypeCategoryType),
					LanguageCode: lang,
					Name:         trans.Name,
					Description:  trans.Description,
				}
				if err := importer.CreateTranslation(ctx, translation); err != nil {
					return fmt.Errorf("failed to create translation: %w", err)
				}
			}
		} else {
			// Category type already exists, skip creation
		}
	}

	// Create subcategories first to get their IDs
	subcategoryMap := make(map[string]uint) // name -> ID mapping
	for _, cat := range defaultData.Categories {
		for _, subcat := range cat.Subcategories {
			// Use English translation if available, otherwise use first available translation with non-empty name
			var enName, enDesc string
			if trans, ok := subcat.Translations[LangEN]; ok && trans.Name != "" {
				enName = trans.Name
				enDesc = trans.Description
			} else {
				for _, trans := range subcat.Translations {
					if trans.Name != "" {
						enName = trans.Name
						enDesc = trans.Description
						break
					}
				}
			}
			if enName == "" {
				return fmt.Errorf("failed to create subcategory: translation missing required name field")
			}
			subcatObj, err := importer.CreateSubcategory(ctx, enName, enDesc, true, subcat.Translations)
			if err != nil {
				return fmt.Errorf("failed to create subcategory %q: %w", enName, err)
			}
			subcategoryMap[enName] = subcatObj.ID
		}
	}

	// Create categories and link subcategories
	for _, cat := range defaultData.Categories {
		var subcategoryIDs []uint
		for _, subcat := range cat.Subcategories {
			enName := subcat.Translations[LangEN].Name
			if id, ok := subcategoryMap[enName]; ok {
				subcategoryIDs = append(subcategoryIDs, id)
			}
		}

		if _, err := importer.CreateCategory(ctx, cat.Name, cat.Description, cat.TypeID, cat.Translations, subcategoryIDs); err != nil {
			return fmt.Errorf("failed to create category %q: %w", cat.Name, err)
		}
	}

	return nil
}

// readDefaultCategoriesFile reads and parses the categories.json file
func readDefaultCategoriesFile() (*DefaultCategoriesData, error) {
	data, err := os.ReadFile(filepath.Join("defaults", "categories.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read categories.json: %w", err)
	}

	var defaultData DefaultCategoriesData
	if err := json.Unmarshal(data, &defaultData); err != nil {
		return nil, fmt.Errorf("failed to parse categories.json: %w", err)
	}

	return &defaultData, nil
}

// ImportDefaultPrompts imports the default prompts from the prompts.json file
func ImportDefaultPrompts(ctx context.Context, db *gorm.DB) error {
	// TODO: Implement default prompts import when prompts.json is ready
	return nil
}
