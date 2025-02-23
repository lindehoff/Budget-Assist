# Data Model Design

## Core Entities

### Categories
```sql
CREATE TABLE category_types (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,  -- e.g., "VEHICLE", "PROPERTY", "FIXED_COST"
    is_multiple BOOLEAN DEFAULT false,  -- true for vehicle/property
    description TEXT
);

CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    type_id INTEGER REFERENCES category_types(id),
    name TEXT NOT NULL,
    instance_identifier TEXT,  -- e.g., "Vehicle: ABC123", "Property: Summer House"
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subcategories (
    id INTEGER PRIMARY KEY,
    category_type_id INTEGER REFERENCES category_types(id),
    name TEXT NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT true  -- false for user-defined subcategories
);

-- Translations for category types, categories, and subcategories
CREATE TABLE translations (
    id INTEGER PRIMARY KEY,
    entity_type TEXT NOT NULL,  -- 'category_type', 'category', 'subcategory'
    entity_id INTEGER NOT NULL,
    language_code TEXT NOT NULL,  -- e.g., 'sv', 'de', 'es'
    name TEXT NOT NULL,
    UNIQUE(entity_type, entity_id, language_code)
);

CREATE INDEX idx_translations_lookup ON translations(entity_type, entity_id, language_code);
```

### Transactions
```sql
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY,
    amount DECIMAL NOT NULL,
    currency TEXT CHECK(currency IN ('SEK', 'EUR', 'USD')),
    transaction_date DATETIME NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES categories(id),
    subcategory_id INTEGER REFERENCES subcategories(id),
    raw_data TEXT,
    ai_analysis TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Predefined Categories

### Income
```json
{
  "type": "INCOME",
  "name": "Income",
  "is_multiple": true,
  "translations": {
    "sv": "Inkomster"
  },
  "subcategories": [
    {
      "name": "Salary",
      "translations": {
        "sv": "Lön"
      }
    },
    {
      "name": "Sales",
      "translations": {
        "sv": "Försäljning"
      }
    },
    {
      "name": "Gift",
      "translations": {
        "sv": "Gåva"
      }
    }
  ]
}
```

### Property
```json
{
  "type": "PROPERTY",
  "name": "Property",
  "is_multiple": true,
  "translations": {
    "sv": "Bostad"
  },
  "subcategories": [
    {
      "name": "Heating",
      "translations": {
        "sv": "Värme"
      }
    },
    {
      "name": "Electricity",
      "translations": {
        "sv": "Hushållsel"
      }
    },
    {
      "name": "Water and Sewage",
      "translations": {
        "sv": "Vatten och avlopp"
      }
    },
    {
      "name": "Waste Management",
      "translations": {
        "sv": "Renhållningsavgift"
      }
    },
    {
      "name": "Maintenance/Repairs",
      "translations": {
        "sv": "Underhåll/reparationer"
      }
    },
    {
      "name": "Other",
      "translations": {
        "sv": "Övrigt"
      }
    },
    {
      "name": "Home Insurance",
      "translations": {
        "sv": "Villahemförsäkring"
      }
    },
    {
      "name": "Property Tax",
      "translations": {
        "sv": "Fastighetsskatt"
      }
    },
    {
      "name": "Loan Interest",
      "translations": {
        "sv": "Lån Ränta"
      }
    },
    {
      "name": "Loan Amortization",
      "translations": {
        "sv": "Lån Amortering"
      }
    }
  ]
}
```

### Vehicle
```json
{
  "type": "VEHICLE",
  "name": "Vehicle",
  "is_multiple": true,
  "translations": {
    "sv": "Fordon"
  },
  "subcategories": [
    {
      "name": "Tax",
      "translations": {
        "sv": "Skatt"
      }
    },
    {
      "name": "Inspection",
      "translations": {
        "sv": "Besiktning"
      }
    },
    {
      "name": "Insurance",
      "translations": {
        "sv": "Försäkring"
      }
    },
    {
      "name": "Tires",
      "translations": {
        "sv": "Däck"
      }
    },
    {
      "name": "Repairs and Service",
      "translations": {
        "sv": "Reparationer och service"
      }
    },
    {
      "name": "Parking",
      "translations": {
        "sv": "Parkering"
      }
    },
    {
      "name": "Other",
      "translations": {
        "sv": "Övrigt"
      }
    },
    {
      "name": "Loan Interest",
      "translations": {
        "sv": "Lån Ränta"
      }
    },
    {
      "name": "Loan Amortization",
      "translations": {
        "sv": "Lån Amortering"
      }
    }
  ]
}
```

### Fixed Costs
```json
{
  "type": "FIXED_COSTS",
  "name": "Fixed Costs",
  "is_multiple": false,
  "translations": {
    "sv": "Fasta kostnader"
  },
  "subcategories": [
    {
      "name": "Media",
      "translations": {
        "sv": "Medier"
      }
    },
    {
      "name": "Other Insurance",
      "translations": {
        "sv": "Övriga försäkringar"
      }
    },
    {
      "name": "Unemployment Insurance and Union Fees",
      "translations": {
        "sv": "A-kassa och fackavgift"
      }
    },
    {
      "name": "Childcare & Maintenance",
      "translations": {
        "sv": "Barnomsorg & underhållsbidrag"
      }
    },
    {
      "name": "Student Loan Repayment",
      "translations": {
        "sv": "Återbetalning CSN-lån"
      }
    }
  ]
}
```

### Variable Costs
```json
{
  "type": "VARIABLE_COSTS",
  "name": "Variable Costs",
  "is_multiple": false,
  "translations": {
    "sv": "Rörliga kostnader"
  },
  "subcategories": [
    {
      "name": "Groceries",
      "translations": {
        "sv": "Livsmedel"
      }
    },
    {
      "name": "Clothes and Shoes",
      "translations": {
        "sv": "Kläder och skor"
      }
    },
    {
      "name": "Leisure/Play",
      "translations": {
        "sv": "Fritid/lek"
      }
    },
    {
      "name": "Mobile Phone",
      "translations": {
        "sv": "Mobiltelefon"
      }
    },
    {
      "name": "Hygiene Products",
      "translations": {
        "sv": "Hygienartiklar"
      }
    },
    {
      "name": "Consumables",
      "translations": {
        "sv": "Förbrukningsvaror"
      }
    },
    {
      "name": "Home Equipment",
      "translations": {
        "sv": "Hemutrustning"
      }
    },
    {
      "name": "Lunch Expenses",
      "translations": {
        "sv": "Lunchkostnader"
      }
    },
    {
      "name": "Public Transport",
      "translations": {
        "sv": "Kollektivresor"
      }
    },
    {
      "name": "Medical, Dental, Medicine",
      "translations": {
        "sv": "Läkare, tandläkare, medicin"
      }
    },
    {
      "name": "Savings",
      "translations": {
        "sv": "Sparande"
      }
    },
    {
      "name": "Other Expenses",
      "translations": {
        "sv": "Andra kostnader"
      }
    }
  ]
} 