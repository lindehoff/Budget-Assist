# API Design

## RESTful Endpoints

### Transaction Management
```go
// Base path: /api/v1

// List transactions with filtering
GET /transactions
Query Parameters:
  - start_date: string (YYYY-MM-DD)
  - end_date: string (YYYY-MM-DD)
  - category_id: int
  - subcategory_id: int
  - min_amount: float
  - max_amount: float
  - search: string
  - page: int
  - per_page: int

// Get single transaction
GET /transactions/:id

// Update transaction
PATCH /transactions/:id
{
    "category_id": int,
    "subcategory_id": int,
    "notes": string,
    "tags": []string
}

// Bulk update transactions
POST /transactions/bulk
{
    "transaction_ids": []int,
    "updates": {
        "category_id": int,
        "subcategory_id": int
    }
}
```

### Category Management
```go
// List categories
GET /categories
Query Parameters:
  - type: string (VEHICLE, PROPERTY, etc.)
  - include_inactive: boolean

// Create new category instance
POST /categories
{
    "type": "VEHICLE",
    "name_sv": "Bil ABC123",
    "name_en": "Car ABC123"
}

// Update category
PATCH /categories/:id
{
    "name_sv": string,
    "name_en": string,
    "is_active": boolean
}
```

### Reports & Analysis
```go
// Get monthly summary
GET /reports/monthly
Query Parameters:
  - year: int
  - month: int
  - category_type: string

// Get category breakdown
GET /reports/categories
Query Parameters:
  - start_date: string
  - end_date: string
  - group_by: string (category, subcategory)

// Get budget analysis
GET /reports/budget-analysis
Query Parameters:
  - period: string (monthly, yearly)
```

## Response Formats

### Standard Response Envelope
```json
{
  "status": "success",
  "data": {},
  "metadata": {
    "page": 1,
    "per_page": 20,
    "total": 100
  },
  "errors": null
}
```

### Error Response
```json
{
  "status": "error",
  "data": null,
  "errors": [
    {
      "code": "VALIDATION_ERROR",
      "field": "amount",
      "message": "Amount must be positive"
    }
  ]
}
```

## WebSocket Events

### Real-time Updates
```go
// Connection endpoint
WS /ws/updates

// Event types
type WSEvent struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Event examples:
{
    "type": "transaction_created",
    "payload": { "id": 123, "amount": 100.00 }
}

{
    "type": "category_updated",
    "payload": { "id": 456, "name": "Updated Name" }
}
```

## API Versioning

### Version Headers
```http
Accept: application/json; version=1
X-API-Version: 1
```

### URL Versioning
```
/api/v1/transactions
/api/v2/transactions
```

## Authentication & Authorization

### JWT Structure
```go
type Claims struct {
    UserID    int64    `json:"uid"`
    Role      string   `json:"role"`
    Features  []string `json:"features"`
    ExpiresAt int64    `json:"exp"`
}
```

### Required Headers
```http
Authorization: Bearer <jwt_token>
X-Request-ID: <uuid>
```

## Rate Limiting

### Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### Configuration
```go
type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize        int
    UserSpecific     bool
}
```

## API Documentation
- OpenAPI/Swagger Specification
- Interactive documentation at `/api/docs`
- Example requests and responses
- Authentication guide
- Rate limiting details 