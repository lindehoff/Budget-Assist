// Package db provides database models and operations.
//
// Note: The struct field alignment in this package is optimized for database schema
// rather than memory usage. This is intentional as these models are primarily used
// for database operations with GORM, and the schema design takes precedence over
// memory optimization.
package db

import (
	"fmt"
	"time"

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
	ID           uint      `gorm:"primarykey"`
	EntityID     uint      `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
	Name         string    `gorm:"not null;size:100"`
	Description  string    `gorm:"size:500"`
	EntityType   string    `gorm:"not null;size:50"`
	LanguageCode string    `gorm:"not null;size:5"`
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

// Category represents a specific instance of a category type
type Category struct {
	ID                 uint      `gorm:"primarykey"`
	TypeID             uint      `gorm:"not null"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
	Name               string    `gorm:"not null;size:100"`
	Description        string    `gorm:"size:500"`
	InstanceIdentifier string    `gorm:"size:100"`
	IsActive           bool      `gorm:"not null"`
	Type               CategoryType
	Transactions       []Transaction
	Translations       []Translation `gorm:"polymorphic:Entity;polymorphicValue:category"`
}

// GetTranslation returns the translated name for the specified language code
func (c *Category) GetTranslation(langCode string) string {
	for _, t := range c.Translations {
		if t.LanguageCode == langCode {
			return t.Name
		}
	}
	return c.Name // Fallback to English name
}

// Subcategory represents subcategories within a category type
type Subcategory struct {
	ID             uint      `gorm:"primarykey"`
	CategoryTypeID uint      `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
	Name           string    `gorm:"not null;size:100"`
	Description    string    `gorm:"size:500"`
	IsSystem       bool      `gorm:"not null"`
	CategoryType   CategoryType
	Transactions   []Transaction
	Translations   []Translation `gorm:"polymorphic:Entity;polymorphicValue:subcategory"`
}

// GetTranslation returns the translated name for the specified language code
func (s *Subcategory) GetTranslation(langCode string) string {
	for _, t := range s.Translations {
		if t.LanguageCode == langCode {
			return t.Name
		}
	}
	return s.Name // Fallback to English name
}

// Transaction represents a financial transaction
type Transaction struct {
	ID              uint `gorm:"primarykey"`
	CategoryID      *uint
	SubcategoryID   *uint
	Amount          float64      `gorm:"not null"`
	CreatedAt       time.Time    `gorm:"not null"`
	UpdatedAt       time.Time    `gorm:"not null"`
	TransactionDate time.Time    `gorm:"not null"`
	Description     string       `gorm:"size:500"`
	RawData         string       `gorm:"size:1000"`
	AIAnalysis      string       `gorm:"size:1000"`
	Currency        string       `gorm:"not null;size:3"`
	Category        *Category    `gorm:"foreignKey:CategoryID"`
	Subcategory     *Subcategory `gorm:"foreignKey:SubcategoryID"`
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
