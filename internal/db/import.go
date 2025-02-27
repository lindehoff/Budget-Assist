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

// importCategoryType imports a single category type and its translations
func importCategoryType(tx *gorm.DB, ct struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsMultiple   bool   `json:"isMultiple"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}) error {
	var existingType CategoryType
	if err := tx.Where("name = ?", ct.Name).First(&existingType).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check for existing category type: %w", err)
		}
		// Create new category type if it doesn't exist
		categoryType := CategoryType{
			Name:        ct.Name,
			Description: ct.Description,
			IsMultiple:  ct.IsMultiple,
		}
		if err := tx.Create(&categoryType).Error; err != nil {
			return fmt.Errorf("failed to create category type: %w", err)
		}
		existingType = categoryType
	}

	// Create translations for category type if they don't exist
	for lang, transl := range ct.Translations {
		if err := createTranslationIfNotExists(tx, existingType.ID, EntityTypeCategoryType, lang, transl.Name, transl.Description); err != nil {
			return err
		}
	}
	return nil
}

// importSubcategory imports a single subcategory and its translations
func importSubcategory(tx *gorm.DB, categoryID uint, typeID uint, subcat struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}) error {
	var existingSubcategory Subcategory
	if err := tx.Where("name = ? AND category_type_id = ?", subcat.Name, typeID).First(&existingSubcategory).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check for existing subcategory: %w", err)
		}
		// Create new subcategory if it doesn't exist
		subcategory := Subcategory{
			Name:           subcat.Name,
			Description:    subcat.Description,
			CategoryTypeID: typeID,
			IsActive:       true,
		}
		if err := tx.Create(&subcategory).Error; err != nil {
			return fmt.Errorf("failed to create subcategory: %w", err)
		}
		existingSubcategory = subcategory

		// Link subcategory to category
		link := CategorySubcategory{
			CategoryID:    categoryID,
			SubcategoryID: subcategory.ID,
			IsActive:      true,
		}
		if err := tx.Create(&link).Error; err != nil {
			return fmt.Errorf("failed to create category-subcategory link: %w", err)
		}
	}

	// Create translations for subcategory if they don't exist
	for lang, transl := range subcat.Translations {
		if err := createTranslationIfNotExists(tx, existingSubcategory.ID, EntityTypeSubcategory, lang, transl.Name, transl.Description); err != nil {
			return err
		}
	}
	return nil
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
		if err := importCategoryTypes(tx, defaultData.CategoryTypes); err != nil {
			return err
		}

		// Import categories and their subcategories
		return importCategories(tx, defaultData.Categories)
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

// importCategoryTypes imports all category types from the default data
func importCategoryTypes(tx *gorm.DB, categoryTypes []struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsMultiple   bool   `json:"isMultiple"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}) error {
	for _, ct := range categoryTypes {
		if err := importCategoryType(tx, ct); err != nil {
			return err
		}
	}
	return nil
}

// importCategories imports all categories and their subcategories
func importCategories(tx *gorm.DB, categories []struct {
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
}) error {
	for _, cat := range categories {
		// Create or get existing category
		existingCategory, err := createOrGetCategory(tx, cat)
		if err != nil {
			return err
		}

		// Create translations for category
		if err := createCategoryTranslations(tx, existingCategory.ID, cat.Translations); err != nil {
			return err
		}

		// Create subcategories
		if err := createSubcategories(tx, existingCategory.ID, cat.TypeID, cat.Subcategories); err != nil {
			return err
		}
	}
	return nil
}

// createOrGetCategory creates a new category or gets an existing one
func createOrGetCategory(tx *gorm.DB, cat struct {
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
}) (*Category, error) {
	var existingCategory Category
	if err := tx.Where("name = ? AND type_id = ?", cat.Name, cat.TypeID).First(&existingCategory).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check for existing category: %w", err)
		}
		// Create new category if it doesn't exist
		category := Category{
			Name:        cat.Name,
			Description: cat.Description,
			TypeID:      cat.TypeID,
			IsActive:    true,
		}
		if err := tx.Create(&category).Error; err != nil {
			return nil, fmt.Errorf("failed to create category: %w", err)
		}
		return &category, nil
	}
	return &existingCategory, nil
}

// createCategoryTranslations creates translations for a category
func createCategoryTranslations(tx *gorm.DB, categoryID uint, translations map[string]struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}) error {
	for lang, transl := range translations {
		if err := createTranslationIfNotExists(tx, categoryID, EntityTypeCategory, lang, transl.Name, transl.Description); err != nil {
			return err
		}
	}
	return nil
}

// createSubcategories creates subcategories for a category
func createSubcategories(tx *gorm.DB, categoryID uint, typeID uint, subcategories []struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Translations map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translations"`
}) error {
	for _, subcat := range subcategories {
		if err := importSubcategory(tx, categoryID, typeID, subcat); err != nil {
			return err
		}
	}
	return nil
}

// ImportDefaultPrompts imports the default prompts from the prompts.json file
func ImportDefaultPrompts(ctx context.Context, db *gorm.DB) error {
	// TODO: Implement default prompts import when prompts.json is ready
	return nil
}
