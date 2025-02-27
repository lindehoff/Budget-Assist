// Package db provides database models and operations.
//
// Note: The struct field alignment in this package is optimized for database schema
// rather than memory usage. This is intentional as these models are primarily used
// for database operations with GORM, and the schema design takes precedence over
// memory optimization.
package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// EntityType represents the type of entity that can be translated
type EntityType string

const (
	EntityTypeCategoryType EntityType = "category_type"
	EntityTypeCategory     EntityType = "category"
	EntityTypeSubcategory  EntityType = "subcategory"
)

// Translation represents a localized name for an entity
type Translation struct {
	gorm.Model
	EntityID     uint   `gorm:"not null"`
	EntityType   string `gorm:"not null"`
	LanguageCode string `gorm:"not null"`
	Name         string `gorm:"not null"`
	Description  string
}

// CategoryType represents different types of categories (e.g., VEHICLE, PROPERTY)
type CategoryType struct {
	ID            uint          `gorm:"primarykey"`
	CreatedAt     time.Time     `gorm:"not null"`
	UpdatedAt     time.Time     `gorm:"not null"`
	Name          string        `gorm:"not null;unique;size:100"`
	Description   string        `gorm:"size:500"`
	IsMultiple    bool          `gorm:"not null"`
	Categories    []Category    `gorm:"foreignKey:TypeID"`
	Subcategories []Subcategory `gorm:"foreignKey:CategoryTypeID"`
	Translations  []Translation `gorm:"polymorphic:Entity;polymorphicValue:category_type"`
}

// GetTranslation returns the translated name for the specified language code
func (ct *CategoryType) GetTranslation(langCode string) string {
	for _, t := range ct.Translations {
		if t.LanguageCode == langCode {
			return t.Name
		}
	}
	return ct.Name // Fallback to English name
}

// Category represents a main category
type Category struct {
	gorm.Model
	Name               string `gorm:"not null;size:100"`
	Description        string `gorm:"size:500"`
	TypeID             uint   `gorm:"not null"`
	InstanceIdentifier string
	IsActive           bool          `gorm:"default:true"`
	Subcategories      []Subcategory `gorm:"many2many:category_subcategories;"`
	Translations       []Translation `gorm:"polymorphic:Entity;polymorphicValue:category"`
}

// BeforeCreate validates the category before creation
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.Name == "" && len(c.Translations) == 0 {
		return fmt.Errorf("either name or at least one translation is required")
	}
	if c.TypeID == 0 {
		return fmt.Errorf("type ID is required")
	}
	return nil
}

// GetName returns the translated name for the given language code
func (c *Category) GetName(langCode string) string {
	if langCode == LangEN && c.Name != "" {
		return c.Name
	}
	for _, t := range c.Translations {
		if t.LanguageCode == langCode {
			return t.Name
		}
	}
	// Return English name as fallback
	if c.Name != "" {
		return c.Name
	}
	// Return first available translation if no English name
	if len(c.Translations) > 0 {
		return c.Translations[0].Name
	}
	return ""
}

// GetDescription returns the translated description for the given language code
func (c *Category) GetDescription(langCode string) string {
	if langCode == LangEN && c.Description != "" {
		return c.Description
	}
	for _, t := range c.Translations {
		if t.LanguageCode == langCode {
			return t.Description
		}
	}
	// Return English description as fallback
	if c.Description != "" {
		return c.Description
	}
	// Return first available translation if no English description
	if len(c.Translations) > 0 {
		return c.Translations[0].Description
	}
	return ""
}

// Subcategory represents a subcategory that can be linked to multiple categories
type Subcategory struct {
	gorm.Model
	Name               string `gorm:"not null;size:100"`
	Description        string `gorm:"size:500"`
	CategoryTypeID     uint   `gorm:"not null"`
	InstanceIdentifier string
	IsActive           bool          `gorm:"default:true"`
	IsSystem           bool          `gorm:"default:false"`
	Tags               string        `gorm:"size:500"` // JSON string of tags for this subcategory
	Categories         []Category    `gorm:"many2many:category_subcategories;"`
	Translations       []Translation `gorm:"polymorphic:Entity;polymorphicValue:subcategory"`
}

// GetName returns the translated name for the given language code
func (s *Subcategory) GetName(langCode string) string {
	if langCode == LangEN && s.Name != "" {
		return s.Name
	}
	for _, t := range s.Translations {
		if t.LanguageCode == langCode {
			return t.Name
		}
	}
	// Return English name as fallback
	if s.Name != "" {
		return s.Name
	}
	// Return first available translation if no English name
	if len(s.Translations) > 0 {
		return s.Translations[0].Name
	}
	return ""
}

// GetDescription returns the translated description for the given language code
func (s *Subcategory) GetDescription(langCode string) string {
	if langCode == LangEN && s.Description != "" {
		return s.Description
	}
	for _, t := range s.Translations {
		if t.LanguageCode == langCode {
			return t.Description
		}
	}
	// Return English description as fallback
	if s.Description != "" {
		return s.Description
	}
	// Return first available translation if no English description
	if len(s.Translations) > 0 {
		return s.Translations[0].Description
	}
	return ""
}

// BeforeCreate validates the subcategory before creation
func (s *Subcategory) BeforeCreate(tx *gorm.DB) error {
	if s.Name == "" && len(s.Translations) == 0 {
		return fmt.Errorf("either name or at least one translation is required")
	}
	return nil
}

// CategorySubcategory represents the many-to-many relationship between categories and subcategories
type CategorySubcategory struct {
	CategoryID    uint        `gorm:"primaryKey"`
	SubcategoryID uint        `gorm:"primaryKey"`
	IsActive      bool        `gorm:"default:true"`
	Category      Category    `gorm:"constraint:OnDelete:CASCADE"`
	Subcategory   Subcategory `gorm:"constraint:OnDelete:CASCADE"`
}

// Transaction represents a financial transaction
type Transaction struct {
	gorm.Model
	Date            time.Time
	TransactionDate time.Time
	Amount          decimal.Decimal
	Description     string
	CategoryID      *uint
	SubcategoryID   *uint
	Category        *Category    `gorm:"foreignKey:CategoryID"`
	Subcategory     *Subcategory `gorm:"foreignKey:SubcategoryID"`
	Source          string
	Reference       string
	RawData         string `gorm:"type:text"`
	AIAnalysis      string `gorm:"type:text"`
	Metadata        string `gorm:"type:json"`
	Currency        string `gorm:"not null;size:3;default:'SEK'"`
}

// FormatAmount returns the amount formatted with the currency
func (t *Transaction) FormatAmount() string {
	return fmt.Sprintf("%s %s", t.Amount.String(), t.Currency)
}

// BeforeCreate hook to validate the currency
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	switch t.Currency {
	case CurrencySEK, CurrencyEUR, CurrencyUSD:
		return nil
	default:
		return fmt.Errorf("invalid currency: %s", t.Currency)
	}
}

// Tag represents a label that can be attached to transactions
type Tag struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex;not null"`
	Description  string
	Transactions []Transaction `gorm:"many2many:transaction_tags;"`
}

// Budget represents a budget plan for a specific category
type Budget struct {
	ID             uint `gorm:"primarykey"`
	CategoryID     uint `gorm:"not null"`
	SubcategoryID  *uint
	Amount         float64      `gorm:"not null"`
	CreatedAt      time.Time    `gorm:"not null"`
	UpdatedAt      time.Time    `gorm:"not null"`
	StartDate      time.Time    `gorm:"not null"`
	EndDate        time.Time    `gorm:"not null"`
	Description    string       `gorm:"size:500"`
	RecurrenceRule string       `gorm:"size:100"`
	Currency       string       `gorm:"not null;size:3"`
	IsRecurring    bool         `gorm:"not null"`
	Category       Category     `gorm:"foreignKey:CategoryID"`
	Subcategory    *Subcategory `gorm:"foreignKey:SubcategoryID"`
}

// Report represents saved analysis reports
type Report struct {
	gorm.Model
	Name        string    `gorm:"not null"`
	Type        string    // "spending", "income", "budget-comparison", etc.
	Parameters  string    // JSON string of report parameters
	GeneratedAt time.Time `gorm:"not null"`
	Data        string    // JSON string of report data
}

// Language code constants
const (
	LangEN = "en" // English (default)
	LangSV = "sv" // Swedish
)

// Currency constants
const (
	CurrencySEK = "SEK"
	CurrencyEUR = "EUR"
	CurrencyUSD = "USD"
)

// Period constants
const (
	PeriodMonthly = "monthly"
	PeriodYearly  = "yearly"
)

// Prompt represents an AI prompt template
type Prompt struct {
	gorm.Model
	Type         string `gorm:"not null;size:50"`
	Name         string `gorm:"not null;size:100"`
	Description  string `gorm:"size:500"`
	SystemPrompt string `gorm:"not null;type:text"`
	UserPrompt   string `gorm:"not null;type:text"`
	Examples     string `gorm:"type:text"` // JSON string of examples
	Rules        string `gorm:"type:text"` // JSON string of rules
	Version      string `gorm:"not null;size:20"`
	IsActive     bool   `gorm:"not null"`
}

// GetTags returns the tags as a string slice
func (s *Subcategory) GetTags() []string {
	if s.Tags == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(s.Tags), &tags); err != nil {
		// If there's an error, return an empty slice
		return []string{}
	}
	return tags
}

// HasTag checks if the subcategory has a specific tag
func (s *Subcategory) HasTag(tag string) bool {
	tags := s.GetTags()
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
