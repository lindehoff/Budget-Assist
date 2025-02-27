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
	Subcategories []struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Tags         []string `json:"tags,omitempty"`
		Translations map[string]struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"translations"`
	} `json:"subcategories"`
	Categories []struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Type         string `json:"type"`
		Translations map[string]struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"translations"`
		Subcategories []string `json:"subcategories"`
	} `json:"categories"`
}

// importCategoryType imports a single category type and its translations
func importCategoryType(tx *gorm.DB, ct struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsMultiple   bool   `json:"isMultiple"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}) (*CategoryType, error) {
	var existingType CategoryType
	if err := tx.Where("name = ?", ct.Name).First(&existingType).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check for existing category type: %w", err)
		}
		// Create new category type if it doesn't exist
		categoryType := CategoryType{
			Name:        ct.Name,
			Description: ct.Description,
			IsMultiple:  ct.IsMultiple,
		}
		if err := tx.Create(&categoryType).Error; err != nil {
			return nil, fmt.Errorf("failed to create category type: %w", err)
		}
		existingType = categoryType
	}

	// Create translations for category type if they don't exist
	for lang, transl := range ct.Translations {
		if err := createTranslationIfNotExists(tx, existingType.ID, EntityTypeCategoryType, lang, transl.Name, transl.Description); err != nil {
			return nil, err
		}
	}
	return &existingType, nil
}

// importSubcategory imports a single subcategory and its translations
func importSubcategory(tx *gorm.DB, sub struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags,omitempty"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}, categoryTypeID uint) (*Subcategory, error) {
	var existingSubcategory Subcategory
	if err := tx.Where("name = ? AND category_type_id = ?", sub.Name, categoryTypeID).First(&existingSubcategory).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check for existing subcategory: %w", err)
		}
		// Create new subcategory if it doesn't exist
		subcategory := Subcategory{
			Name:           sub.Name,
			Description:    sub.Description,
			CategoryTypeID: categoryTypeID,
			IsActive:       true,
			IsSystem:       true,
		}

		// Convert tags to JSON string
		if len(sub.Tags) > 0 {
			tagsJSON, err := json.Marshal(sub.Tags)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tags: %w", err)
			}
			subcategory.Tags = string(tagsJSON)
		}

		if err := tx.Create(&subcategory).Error; err != nil {
			return nil, fmt.Errorf("failed to create subcategory: %w", err)
		}
		existingSubcategory = subcategory
	}

	// Create translations for subcategory
	for lang, transl := range sub.Translations {
		if err := createTranslationIfNotExists(tx, existingSubcategory.ID, EntityTypeSubcategory, lang, transl.Name, transl.Description); err != nil {
			return nil, err
		}
	}
	return &existingSubcategory, nil
}

// createTranslationIfNotExists creates a translation if it doesn't exist
func createTranslationIfNotExists(tx *gorm.DB, entityID uint, entityType EntityType, lang, name, description string) error {
	var existingTransl Translation
	if err := tx.Where("entity_type = ? AND entity_id = ? AND language_code = ?",
		entityType, entityID, lang).First(&existingTransl).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check for existing translation: %w", err)
		}
		// Create new translation if it doesn't exist
		translation := Translation{
			EntityID:     entityID,
			EntityType:   string(entityType),
			LanguageCode: lang,
			Name:         name,
			Description:  description,
		}
		if err := tx.Create(&translation).Error; err != nil {
			return fmt.Errorf("failed to create translation: %w", err)
		}
	}
	return nil
}

// importCategoryTypes imports all category types and returns a map of names to IDs
func importCategoryTypes(tx *gorm.DB, categoryTypes []struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsMultiple   bool   `json:"isMultiple"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}) (map[string]uint, error) {
	categoryTypeMap := make(map[string]uint)
	for _, ct := range categoryTypes {
		categoryType, err := importCategoryType(tx, ct)
		if err != nil {
			return nil, err
		}
		categoryTypeMap[ct.Name] = categoryType.ID
	}
	return categoryTypeMap, nil
}

// buildSubcategoryMap collects all unique subcategories and creates them if needed
func buildSubcategoryMap(tx *gorm.DB, defaultData *DefaultCategoriesData, categoryTypeMap map[string]uint) (map[string]*Subcategory, error) {
	subcategoryMap := make(map[string]*Subcategory)

	for _, cat := range defaultData.Categories {
		typeID, ok := categoryTypeMap[cat.Type]
		if !ok {
			return nil, fmt.Errorf("category type %q not found", cat.Type)
		}

		for _, subName := range cat.Subcategories {
			if _, exists := subcategoryMap[subName]; !exists {
				// Find matching subcategory definition
				var subDef *struct {
					Name         string   `json:"name"`
					Description  string   `json:"description"`
					Tags         []string `json:"tags,omitempty"`
					Translations map[string]struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"translations"`
				}
				for _, s := range defaultData.Subcategories {
					if s.Name == subName {
						subDef = &s
						break
					}
				}
				if subDef == nil {
					return nil, fmt.Errorf("subcategory %q not found in subcategories list", subName)
				}

				// Import the subcategory
				sub, err := importSubcategory(tx, *subDef, typeID)
				if err != nil {
					return nil, err
				}
				subcategoryMap[subName] = sub
			}
		}
	}

	return subcategoryMap, nil
}

// createCategoriesWithSubcategories creates categories and links them to their subcategories
func createCategoriesWithSubcategories(tx *gorm.DB, categories []struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
	Subcategories []string `json:"subcategories"`
}, categoryTypeMap map[string]uint, subcategoryMap map[string]*Subcategory) error {
	for _, cat := range categories {
		typeID := categoryTypeMap[cat.Type]

		// Check if category already exists
		var existingCategory Category
		if err := tx.Where("name = ? AND type_id = ?", cat.Name, typeID).First(&existingCategory).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to check for existing category: %w", err)
			}
			// Create the category only if it doesn't exist
			category := Category{
				Name:        cat.Name,
				Description: cat.Description,
				TypeID:      typeID,
				IsActive:    true,
			}
			if err := tx.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to create category: %w", err)
			}

			// Create translations for category
			for lang, transl := range cat.Translations {
				if err := createTranslationIfNotExists(tx, category.ID, EntityTypeCategory, lang, transl.Name, transl.Description); err != nil {
					return err
				}
			}

			// Link subcategories
			for _, subName := range cat.Subcategories {
				sub := subcategoryMap[subName]
				link := CategorySubcategory{
					CategoryID:    category.ID,
					SubcategoryID: sub.ID,
					IsActive:      true,
				}
				if err := tx.Create(&link).Error; err != nil {
					return fmt.Errorf("failed to create category-subcategory link: %w", err)
				}
			}
		}
	}

	return nil
}

// ImportDefaultCategories imports the default categories from the categories.json file
func ImportDefaultCategories(ctx context.Context, db *gorm.DB) error {
	// Read and parse the categories.json file
	defaultData, err := readDefaultCategoriesFile()
	if err != nil {
		return err
	}

	// Start a transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Import category types
		categoryTypeMap, err := importCategoryTypes(tx, defaultData.CategoryTypes)
		if err != nil {
			return err
		}

		// Build subcategory map
		subcategoryMap, err := buildSubcategoryMap(tx, defaultData, categoryTypeMap)
		if err != nil {
			return err
		}

		// Create categories and link to subcategories
		return createCategoriesWithSubcategories(tx, defaultData.Categories, categoryTypeMap, subcategoryMap)
	})
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
