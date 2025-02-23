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
    name_sv TEXT NOT NULL,
    name_en TEXT NOT NULL,
    instance_identifier TEXT,  -- e.g., "Vehicle: ABC123", "Property: Summer House"
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subcategories (
    id INTEGER PRIMARY KEY,
    category_type_id INTEGER REFERENCES category_types(id),
    name_sv TEXT NOT NULL,
    name_en TEXT NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT true  -- false for user-defined subcategories
);
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

### Income (Inkomster)
```json
{
  "type": "INCOME",
  "is_multiple": false,
  "subcategories": [
    "Lön",
    "Försäljning",
    "Gåva"
  ]
}
```

### Property (Bostad)
```json
{
  "type": "PROPERTY",
  "is_multiple": true,
  "subcategories": [
    "Värme",
    "Hushållsel",
    "Vatten och avlopp",
    "Renhållningsavgift",
    "Underhåll/reparationer",
    "Övrigt",
    "Villahemförsäkring",
    "Fastighetsskatt",
    "Lån Ränta",
    "Lån Amortering"
  ]
}
```

### Vehicle (Fordon)
```json
{
  "type": "VEHICLE",
  "is_multiple": true,
  "subcategories": [
    "Skatt",
    "Besiktning",
    "Försäkring",
    "Däck",
    "Reparationer och service",
    "Parkering",
    "Övrigt",
    "Lån Ränta",
    "Lån Amortering"
  ]
}
```

[Additional category definitions continue...] 