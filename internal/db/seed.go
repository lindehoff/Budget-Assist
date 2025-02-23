package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// CategoryTypeData represents the data structure for seeding category types
type CategoryTypeData struct {
	Name          string
	IsMultiple    bool
	Description   string
	Translations  map[string]string
	Subcategories []SubcategoryData
}

// SubcategoryData represents the data structure for seeding subcategories
type SubcategoryData struct {
	Name         string
	Description  string
	Translations map[string]string
}

// predefinedCategories contains all the predefined category types and their subcategories
var predefinedCategories = []CategoryTypeData{
	{
		Name:        "Income",
		IsMultiple:  true,
		Description: "Income sources",
		Translations: map[string]string{
			LangSV: "Inkomster",
		},
		Subcategories: []SubcategoryData{
			{
				Name: "Salary",
				Translations: map[string]string{
					LangSV: "Lön",
				},
			},
			{
				Name: "Sales",
				Translations: map[string]string{
					LangSV: "Försäljning",
				},
			},
			{
				Name: "Gift",
				Translations: map[string]string{
					LangSV: "Gåva",
				},
			},
		},
	},
	{
		Name:        "Property",
		IsMultiple:  true,
		Description: "Property related expenses",
		Translations: map[string]string{
			LangSV: "Bostad",
		},
		Subcategories: []SubcategoryData{
			{
				Name: "Heating",
				Translations: map[string]string{
					LangSV: "Värme",
				},
			},
			{
				Name: "Electricity",
				Translations: map[string]string{
					LangSV: "Hushållsel",
				},
			},
			{
				Name: "Water and Sewage",
				Translations: map[string]string{
					LangSV: "Vatten och avlopp",
				},
			},
			{
				Name: "Waste Management",
				Translations: map[string]string{
					LangSV: "Renhållningsavgift",
				},
			},
			{
				Name: "Maintenance/Repairs",
				Translations: map[string]string{
					LangSV: "Underhåll/reparationer",
				},
			},
			{
				Name: "Other",
				Translations: map[string]string{
					LangSV: "Övrigt",
				},
			},
			{
				Name: "Home Insurance",
				Translations: map[string]string{
					LangSV: "Villahemförsäkring",
				},
			},
			{
				Name: "Property Tax",
				Translations: map[string]string{
					LangSV: "Fastighetsskatt",
				},
			},
			{
				Name: "Loan Interest",
				Translations: map[string]string{
					LangSV: "Lån Ränta",
				},
			},
			{
				Name: "Loan Amortization",
				Translations: map[string]string{
					LangSV: "Lån Amortering",
				},
			},
		},
	},
	{
		Name:        "Vehicle",
		IsMultiple:  true,
		Description: "Vehicle related expenses",
		Translations: map[string]string{
			LangSV: "Fordon",
		},
		Subcategories: []SubcategoryData{
			{
				Name: "Tax",
				Translations: map[string]string{
					LangSV: "Skatt",
				},
			},
			{
				Name: "Inspection",
				Translations: map[string]string{
					LangSV: "Besiktning",
				},
			},
			{
				Name: "Insurance",
				Translations: map[string]string{
					LangSV: "Försäkring",
				},
			},
			{
				Name: "Tires",
				Translations: map[string]string{
					LangSV: "Däck",
				},
			},
			{
				Name: "Repairs and Service",
				Translations: map[string]string{
					LangSV: "Reparationer och service",
				},
			},
			{
				Name: "Parking",
				Translations: map[string]string{
					LangSV: "Parkering",
				},
			},
			{
				Name: "Other",
				Translations: map[string]string{
					LangSV: "Övrigt",
				},
			},
			{
				Name: "Loan Interest",
				Translations: map[string]string{
					LangSV: "Lån Ränta",
				},
			},
			{
				Name: "Loan Amortization",
				Translations: map[string]string{
					LangSV: "Lån Amortering",
				},
			},
		},
	},
	{
		Name:        "Fixed Costs",
		IsMultiple:  false,
		Description: "Regular fixed expenses",
		Translations: map[string]string{
			LangSV: "Fasta kostnader",
		},
		Subcategories: []SubcategoryData{
			{
				Name: "Media",
				Translations: map[string]string{
					LangSV: "Medier",
				},
			},
			{
				Name: "Other Insurance",
				Translations: map[string]string{
					LangSV: "Övriga försäkringar",
				},
			},
			{
				Name: "Unemployment Insurance and Union Fees",
				Translations: map[string]string{
					LangSV: "A-kassa och fackavgift",
				},
			},
			{
				Name: "Childcare & Maintenance",
				Translations: map[string]string{
					LangSV: "Barnomsorg & underhållsbidrag",
				},
			},
			{
				Name: "Student Loan Repayment",
				Translations: map[string]string{
					LangSV: "Återbetalning CSN-lån",
				},
			},
		},
	},
	{
		Name:        "Variable Costs",
		IsMultiple:  false,
		Description: "Variable expenses",
		Translations: map[string]string{
			LangSV: "Rörliga kostnader",
		},
		Subcategories: []SubcategoryData{
			{
				Name: "Groceries",
				Translations: map[string]string{
					LangSV: "Livsmedel",
				},
			},
			{
				Name: "Clothes and Shoes",
				Translations: map[string]string{
					LangSV: "Kläder och skor",
				},
			},
			{
				Name: "Leisure/Play",
				Translations: map[string]string{
					LangSV: "Fritid/lek",
				},
			},
			{
				Name: "Mobile Phone",
				Translations: map[string]string{
					LangSV: "Mobiltelefon",
				},
			},
			{
				Name: "Hygiene Products",
				Translations: map[string]string{
					LangSV: "Hygienartiklar",
				},
			},
			{
				Name: "Consumables",
				Translations: map[string]string{
					LangSV: "Förbrukningsvaror",
				},
			},
			{
				Name: "Home Equipment",
				Translations: map[string]string{
					LangSV: "Hemutrustning",
				},
			},
			{
				Name: "Lunch Expenses",
				Translations: map[string]string{
					LangSV: "Lunchkostnader",
				},
			},
			{
				Name: "Public Transport",
				Translations: map[string]string{
					LangSV: "Kollektivresor",
				},
			},
			{
				Name: "Medical, Dental, Medicine",
				Translations: map[string]string{
					LangSV: "Läkare, tandläkare, medicin",
				},
			},
			{
				Name: "Savings",
				Translations: map[string]string{
					LangSV: "Sparande",
				},
			},
			{
				Name: "Other Expenses",
				Translations: map[string]string{
					LangSV: "Andra kostnader",
				},
			},
		},
	},
}

// SeedPredefinedCategories initializes the database with predefined categories
func SeedPredefinedCategories(ctx context.Context, db *gorm.DB) error {
	op := NewDBOperation(ctx)
	op.logger.Info("starting to seed predefined categories")

	for _, ctData := range predefinedCategories {
		// Validate category type data
		if ctData.Name == "" {
			op.logger.Error("invalid category type data",
				"error", ErrInvalidCategoryType,
				"operation", "validate",
				"entity", "category_type")
			return &DatabaseOperationError{
				Operation: "validate",
				Entity:    "category_type",
				Err:       ErrInvalidCategoryType,
			}
		}

		// Check if category type already exists
		var existing CategoryType
		if err := db.WithContext(ctx).Where("name = ?", ctData.Name).First(&existing).Error; err == nil {
			op.logger.Info("category type already exists, skipping",
				"name", ctData.Name)
			continue
		}

		op.logger.Info("creating category type",
			"name", ctData.Name,
			"is_multiple", ctData.IsMultiple)

		// Create category type
		categoryType := &CategoryType{
			Name:        ctData.Name,
			IsMultiple:  ctData.IsMultiple,
			Description: ctData.Description,
		}
		if err := db.WithContext(ctx).Create(categoryType).Error; err != nil {
			op.logger.Error("failed to create category type",
				"error", err,
				"name", ctData.Name)
			return &DatabaseOperationError{
				Operation: "create",
				Entity:    "category_type",
				Err:       err,
			}
		}

		// Create translations for category type
		for lang, translation := range ctData.Translations {
			op.logger.Info("adding translation for category type",
				"language", lang,
				"name", ctData.Name,
				"translation", translation)

			if err := db.WithContext(ctx).Create(&Translation{
				EntityType:   EntityTypeCategoryType,
				EntityID:     categoryType.ID,
				LanguageCode: lang,
				Name:         translation,
			}).Error; err != nil {
				op.logger.Error("failed to create translation",
					"error", err,
					"entity_type", "category_type",
					"entity_id", categoryType.ID)
				return &DatabaseOperationError{
					Operation: "create_translation",
					Entity:    fmt.Sprintf("category_type_%d", categoryType.ID),
					Err:       err,
				}
			}
		}

		// Create subcategories
		for _, subData := range ctData.Subcategories {
			// Validate subcategory data
			if subData.Name == "" {
				op.logger.Error("invalid subcategory data",
					"error", "empty name",
					"category_type", ctData.Name)
				return &DatabaseOperationError{
					Operation: "validate",
					Entity:    "subcategory",
					Err:       fmt.Errorf("empty subcategory name for category type %s", ctData.Name),
				}
			}

			op.logger.Info("creating subcategory",
				"name", subData.Name,
				"category_type", ctData.Name)

			subcategory := &Subcategory{
				CategoryTypeID: categoryType.ID,
				Name:           subData.Name,
				Description:    subData.Description,
				IsSystem:       true,
			}
			if err := db.WithContext(ctx).Create(subcategory).Error; err != nil {
				op.logger.Error("failed to create subcategory",
					"error", err,
					"name", subData.Name)
				return &DatabaseOperationError{
					Operation: "create",
					Entity:    "subcategory",
					Err:       err,
				}
			}

			// Create translations for subcategory
			for lang, translation := range subData.Translations {
				op.logger.Info("adding translation for subcategory",
					"language", lang,
					"name", subData.Name,
					"translation", translation)

				if err := db.WithContext(ctx).Create(&Translation{
					EntityType:   EntityTypeSubcategory,
					EntityID:     subcategory.ID,
					LanguageCode: lang,
					Name:         translation,
				}).Error; err != nil {
					op.logger.Error("failed to create translation",
						"error", err,
						"entity_type", "subcategory",
						"entity_id", subcategory.ID)
					return &DatabaseOperationError{
						Operation: "create_translation",
						Entity:    fmt.Sprintf("subcategory_%d", subcategory.ID),
						Err:       err,
					}
				}
			}
		}
	}

	op.logger.Info("successfully seeded predefined categories")
	return nil
}
