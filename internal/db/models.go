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

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CategoryType represents different types of categories (e.g., VEHICLE, PROPERTY)
type CategoryType struct {
	ID          uint       `gorm:"primarykey"`
	Name        string     `gorm:"not null;unique;size:100"`
	Description string     `gorm:"size:500"`
	IsMultiple  bool       `gorm:"not null"`
	Categories  []Category `gorm:"foreignKey:TypeID"`
}

// Category represents a main category
type Category struct {
	ID                 uint   `gorm:"primarykey"`
	Name               string `gorm:"not null;size:100"`
	Description        string `gorm:"size:500"`
	TypeID             uint   `gorm:"not null"`
	Type               string `gorm:"not null;size:100"` // Reference to CategoryType.Name
	InstanceIdentifier string
	IsActive           bool `gorm:"default:true"`
	Subcategories      []CategorySubcategory
}

// Subcategory represents a subcategory that can be linked to multiple categories
type Subcategory struct {
	ID          uint   `gorm:"primarykey"`
	Name        string `gorm:"not null;size:100"`
	Description string `gorm:"size:500"`
	IsActive    bool   `gorm:"default:true"`
	IsSystem    bool   `gorm:"default:false"`
	Tags        []Tag  `gorm:"many2many:subcategory_tags"`
	Categories  []CategorySubcategory
}

// Tag represents a label that can be attached to subcategories
type Tag struct {
	ID            uint   `gorm:"primarykey"`
	Name          string `gorm:"uniqueIndex;not null"`
	Description   string
	Subcategories []Subcategory `gorm:"many2many:subcategory_tags"`
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
	ID              uint `gorm:"primarykey"`
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
	return fmt.Sprintf("%s %s", t.Amount.StringFixed(2), t.Currency)
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

// Budget represents a budget plan for a specific category
type Budget struct {
	ID             uint `gorm:"primarykey"`
	CategoryID     uint `gorm:"not null"`
	SubcategoryID  *uint
	Amount         float64      `gorm:"not null"`
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
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"not null"`
	Type        string    // "spending", "income", "budget-comparison", etc.
	Parameters  string    // JSON string of report parameters
	GeneratedAt time.Time `gorm:"not null"`
	Data        string    // JSON string of report data
}

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

// PromptType represents different types of prompts
type PromptType string

const (
	BillAnalysisPrompt              PromptType = "bill_analysis"
	ReceiptAnalysisPrompt           PromptType = "receipt_analysis"
	BankStatementAnalysisPrompt     PromptType = "bank_statement_analysis"
	TransactionCategorizationPrompt PromptType = "transaction_categorization"
)

// Prompt represents an AI prompt template
type Prompt struct {
	ID           uint       `gorm:"primarykey"`
	Type         PromptType `gorm:"not null;size:50;uniqueIndex:idx_prompt_type_active"`
	Name         string     `gorm:"not null;size:100"`
	Description  string     `gorm:"size:500"`
	SystemPrompt string     `gorm:"not null;type:text"`
	UserPrompt   string     `gorm:"not null;type:text"`
	Version      string     `gorm:"not null;size:20"`
	IsActive     bool       `gorm:"not null;uniqueIndex:idx_prompt_type_active"`
}
