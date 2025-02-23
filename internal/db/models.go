package db

import (
	"database/sql"
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
	gorm.Model
	EntityType   EntityType `gorm:"not null"`
	EntityID     uint       `gorm:"not null"`
	LanguageCode string     `gorm:"not null"`
	Name         string     `gorm:"not null"`
}

// CategoryType represents different types of categories (e.g., VEHICLE, PROPERTY)
type CategoryType struct {
	gorm.Model
	Name         string `gorm:"not null"`
	IsMultiple   bool   `gorm:"default:false"`
	Description  string
	Categories   []Category    `gorm:"foreignKey:TypeID"`
	Translations []Translation `gorm:"-"`
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
	gorm.Model
	TypeID             uint          `gorm:"not null"`
	CategoryType       CategoryType  `gorm:"foreignKey:TypeID"`
	Name               string        `gorm:"not null"`
	InstanceIdentifier string        // e.g., "Vehicle: ABC123"
	IsActive           bool          `gorm:"default:true"`
	Subcategories      []Subcategory `gorm:"many2many:category_subcategories;"`
	Transactions       []Transaction `gorm:"foreignKey:CategoryID"`
	Translations       []Translation `gorm:"-"`
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
	gorm.Model
	CategoryTypeID uint         `gorm:"not null"`
	CategoryType   CategoryType `gorm:"foreignKey:CategoryTypeID"`
	Name           string       `gorm:"not null"`
	Description    string
	IsSystem       bool          `gorm:"default:true"`
	Categories     []Category    `gorm:"many2many:category_subcategories;"`
	Transactions   []Transaction `gorm:"foreignKey:SubcategoryID"`
	Translations   []Translation `gorm:"-"`
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
	gorm.Model
	Amount          float64   `gorm:"not null"`
	Currency        string    `gorm:"not null;check:currency IN ('SEK', 'EUR', 'USD')"`
	TransactionDate time.Time `gorm:"not null"`
	Description     string
	CategoryID      *uint
	Category        *Category `gorm:"foreignKey:CategoryID"`
	SubcategoryID   *uint
	Subcategory     *Subcategory `gorm:"foreignKey:SubcategoryID"`
	RawData         string       // Original import data
	AIAnalysis      string       // AI-generated insights
	UpdatedAt       sql.NullTime
	Tags            []Tag `gorm:"many2many:transaction_tags;"`
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
	gorm.Model
	CategoryID    uint     `gorm:"not null"`
	Category      Category `gorm:"foreignKey:CategoryID"`
	SubcategoryID *uint
	Subcategory   *Subcategory `gorm:"foreignKey:SubcategoryID"`
	Amount        float64      `gorm:"not null"`
	Currency      string       `gorm:"not null;check:currency IN ('SEK', 'EUR', 'USD')"`
	Period        string       // "monthly", "yearly", etc.
	StartDate     time.Time
	EndDate       time.Time
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
