# Database V2 - Usage Examples

This document provides comprehensive examples of how to use the new database system.

## Table of Contents
- [Basic Queries](#basic-queries)
- [Type-Safe Joins](#type-safe-joins)
- [Where Clauses](#where-clauses)
- [Inserting Data](#inserting-data)
- [Updating Data](#updating-data)
- [Deleting Data](#deleting-data)
- [Transactions](#transactions)
- [Pagination](#pagination)
- [Relations & Preloading](#relations--preloading)
- [Raw SQL](#raw-sql)
- [Helper Functions](#helper-functions)
- [Advanced Patterns](#advanced-patterns)

---

## Basic Queries

### Select All Records

```go
users, err := database.Query[User](db).All(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Select First Record

```go
user, err := database.Query[User](db).
    Where("email", "user@example.com").
    First(ctx)
if err != nil {
    log.Fatal(err)
}
if user == nil {
    log.Println("User not found")
}
```

### Count Records

```go
count, err := database.Query[Order](db).
    Where("status", "pending").
    Count(ctx)
```

### Check if Records Exist

```go
exists, err := database.Query[Product](db).
    Where("sku", "ABC123").
    Exists(ctx)
```

### Select Specific Columns

```go
users, err := database.Query[User](db).
    Select("id", "name", "email").
    Where("active", true).
    All(ctx)
```

---

## Type-Safe Joins

### Simple Inner Join

```go
orders, err := database.Query[Order](db).
    Join("users", "u").
        On("orders.user_id", "=", "u.id").
        End().
    Where("u.active", true).
    Select("orders.*", "u.name as user_name").
    All(ctx)
```

### Multiple Joins

```go
orders, err := database.Query[Order](db).
    Join("users", "u").
        On("orders.user_id", "=", "u.id").
        End().
    LeftJoin("products", "p").
        On("orders.product_id", "=", "p.id").
        End().
    LeftJoin("shipping_addresses", "sa").
        On("orders.shipping_address_id", "=", "sa.id").
        End().
    Where("orders.status", "pending").
    Select("orders.*", "u.name", "p.title", "sa.address").
    All(ctx)
```

### Join with Multiple Conditions

```go
orders, err := database.Query[Order](db).
    Join("users", "u").
        On("orders.user_id", "=", "u.id").
        And("u.active", "=", "true").
        End().
    All(ctx)
```

### Left/Right/Full Joins

```go
// Left Join
users, err := database.Query[User](db).
    LeftJoin("orders", "o").
        On("users.id", "=", "o.user_id").
        End().
    Select("users.*", "COUNT(o.id) as order_count").
    GroupBy("users.id").
    All(ctx)

// Right Join
orders, err := database.Query[Order](db).
    RightJoin("products", "p").
        On("orders.product_id", "=", "p.id").
        End().
    All(ctx)

// Full Join
data, err := database.Query[Data](db).
    FullJoin("other_data", "od").
        On("data.id", "=", "od.data_id").
        End().
    All(ctx)
```

---

## Where Clauses

### Basic Where

```go
users, err := database.Query[User](db).
    Where("active", true).
    Where("age", 18).
    All(ctx)
```

### Where with Operators

```go
products, err := database.Query[Product](db).
    WhereOp("price", ">", 100).
    WhereOp("stock", "<=", 10).
    All(ctx)
```

### Where NOT

```go
users, err := database.Query[User](db).
    WhereNot("status", "banned").
    All(ctx)
```

### Where IN / NOT IN

```go
users, err := database.Query[User](db).
    WhereIn("id", []any{1, 2, 3, 4, 5}).
    All(ctx)

products, err := database.Query[Product](db).
    WhereNotIn("category", []any{"archived", "deleted"}).
    All(ctx)
```

### Where NULL / NOT NULL

```go
users, err := database.Query[User](db).
    WhereNull("deleted_at").
    All(ctx)

users, err := database.Query[User](db).
    WhereNotNull("verified_at").
    All(ctx)
```

### Where LIKE

```go
users, err := database.Query[User](db).
    WhereLike("email", "%@gmail.com").
    All(ctx)
```

### Where Raw (Complex Conditions)

```go
orders, err := database.Query[Order](db).
    WhereRaw("created_at > NOW() - INTERVAL '30 days'").
    All(ctx)

products, err := database.Query[Product](db).
    WhereRaw("price BETWEEN ? AND ?", 10, 100).
    All(ctx)
```

### OR Conditions (Grouped)

```go
users, err := database.Query[User](db).
    Or().
        Where("status", "active").
        Where("status", "pending").
        End().
    All(ctx)
```

---

## Inserting Data

### Insert Single Record

```go
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}

created, err := database.Query[User](db).Insert(ctx, user)
if err != nil {
    log.Fatal(err)
}
// created now has the ID populated (if using auto-increment)
```

### Insert Multiple Records (Bulk Insert)

```go
users := []User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob", Email: "bob@example.com"},
    {Name: "Charlie", Email: "charlie@example.com"},
}

created, err := database.Query[User](db).InsertMany(ctx, users)
if err != nil {
    log.Fatal(err)
}
```

### Upsert (Insert or Update on Conflict)

```go
user := &User{
    Email: "john@example.com",
    Name:  "John Updated",
    Age:   31,
}

// If email already exists, update name and age
result, err := database.Upsert(db, ctx, user, "email", "name", "age")
```

### Bulk Upsert

```go
users := []User{
    {Email: "alice@example.com", Name: "Alice Updated"},
    {Email: "bob@example.com", Name: "Bob Updated"},
}

result, err := database.BulkUpsert(db, ctx, users, "email", "name")
```

### Using Helpers

```go
user := &User{Name: "Jane", Email: "jane@example.com"}
created, err := database.Create(db, ctx, user)

// Create many
users := []User{{Name: "A"}, {Name: "B"}}
created, err := database.CreateMany(db, ctx, users)
```

---

## Updating Data

### Update with Map

```go
affected, err := database.Query[User](db).
    Where("id", userId).
    Update(ctx, map[string]any{
        "name":       "Updated Name",
        "updated_at": time.Now(),
    })
```

### Update Multiple Records

```go
affected, err := database.Query[Order](db).
    Where("status", "pending").
    WhereRaw("created_at < NOW() - INTERVAL '1 hour'").
    Update(ctx, map[string]any{
        "status": "cancelled",
    })
```

### Update and Return Updated Records

```go
users, err := database.Query[User](db).
    Where("active", false).
    UpdateReturning(ctx, map[string]any{
        "active": true,
    })
// users contains all updated records
```

### Update by ID Helper

```go
affected, err := database.UpdateByID[User](db, ctx, userId, map[string]any{
    "last_login": time.Now(),
})
```

---

## Deleting Data

### Delete Records

```go
affected, err := database.Query[User](db).
    Where("active", false).
    WhereRaw("last_login < NOW() - INTERVAL '1 year'").
    Delete(ctx)
```

### Delete and Return Deleted Records

```go
deleted, err := database.Query[Order](db).
    Where("status", "cancelled").
    DeleteReturning(ctx)
```

### Delete by ID Helper

```go
affected, err := database.DeleteByID[User](db, ctx, userId)
```

### Soft Delete

```go
affected, err := database.SoftDelete[User](db, ctx, userId)
// Sets deleted_at to current timestamp
```

### Restore Soft Deleted

```go
affected, err := database.Restore[User](db, ctx, userId)
// Sets deleted_at back to NULL
```

### Exclude Soft Deleted Records

```go
users, err := database.ExcludeSoftDeleted(
    database.Query[User](db),
).All(ctx)
```

### Only Soft Deleted Records

```go
users, err := database.OnlySoftDeleted(
    database.Query[User](db),
).All(ctx)
```

---

## Transactions

### Simple Transaction

```go
err := database.Transaction(ctx, func(tx *pg.Tx) error {
    // Create user
    user := &User{Name: "John", Email: "john@example.com"}
    if _, err := tx.Model(user).Insert(); err != nil {
        return err // Will rollback
    }

    // Create order
    order := &Order{UserID: user.ID, Total: 100.00}
    if _, err := tx.Model(order).Insert(); err != nil {
        return err // Will rollback
    }

    return nil // Will commit
})
```

### Transaction with Result

```go
order, err := database.TransactionWithResult(ctx, func(tx *pg.Tx) (*Order, error) {
    // Create user
    user := &User{Name: "Jane", Email: "jane@example.com"}
    if _, err := tx.Model(user).Insert(); err != nil {
        return nil, err
    }

    // Create order
    order := &Order{UserID: user.ID, Total: 200.00}
    if _, err := tx.Model(order).Insert(); err != nil {
        return nil, err
    }

    return order, nil
})
```

---

## Pagination

### Offset-Based Pagination

```go
page := 1
pageSize := 20

result, err := database.Paginate(
    database.Query[User](db).Where("active", true),
    ctx,
    page,
    pageSize,
)

// result.Data contains the users
// result.Pagination.Total contains total count
// result.Pagination.Page contains current page
// result.Pagination.PageSize contains page size
```

### Manual Pagination

```go
users, err := database.Query[User](db).
    Where("active", true).
    OrderBy("created_at", database.DESC).
    Limit(20).
    Offset(40). // Page 3
    All(ctx)
```

---

## Relations & Preloading

### Preload Relations (go-pg style)

```go
// Preload a single relation
user, err := database.Query[User](db).
    Where("id", userId).
    With("Orders").
    First(ctx)

// Preload nested relations
user, err := database.Query[User](db).
    Where("id", userId).
    With("Orders").
    With("Orders.Products").
    First(ctx)

// Preload multiple relations
users, err := database.Query[User](db).
    With("Orders").
    With("Profile").
    With("Addresses").
    All(ctx)
```

---

## Raw SQL

### Raw Query Returning Multiple Results

```go
type OrderStats struct {
    UserID       int     `json:"user_id"`
    OrderCount   int     `json:"order_count"`
    TotalRevenue float64 `json:"total_revenue"`
}

stats, err := database.RawQuery[OrderStats](db, ctx, `
    SELECT 
        user_id,
        COUNT(*) as order_count,
        SUM(total) as total_revenue
    FROM orders
    WHERE created_at > $1
    GROUP BY user_id
    HAVING COUNT(*) > $2
    ORDER BY total_revenue DESC
`, startDate, minOrders)
```

### Raw Query Returning Single Result

```go
result, err := database.RawQueryOne[OrderStats](db, ctx, `
    SELECT 
        COUNT(*) as order_count,
        SUM(total) as total_revenue
    FROM orders
    WHERE user_id = $1
`, userId)
```

### Raw Execute (No Results)

```go
affected, err := database.RawExec(db, ctx, `
    UPDATE users 
    SET last_active = NOW() 
    WHERE id = $1
`, userId)
```

---

## Helper Functions

### Find by ID

```go
user, err := database.FindByID[User](db, ctx, userId)
```

### Find by Multiple IDs

```go
users, err := database.FindByIDs[User](db, ctx, []any{1, 2, 3, 4, 5})
```

### Batch Processing

```go
err := database.BatchProcess(ctx, 
    database.Query[User](db).Where("active", true),
    100, // batch size
    func(users []User) error {
        // Process each batch
        for _, user := range users {
            // Send email, etc.
        }
        return nil
    },
)
```

### Chunk Processing

```go
err := database.Chunk(ctx,
    database.Query[Order](db).Where("status", "pending"),
    50, // chunk size
    func(orders []Order, chunkNumber int) error {
        log.Printf("Processing chunk %d with %d orders", chunkNumber, len(orders))
        // Process chunk
        return nil
    },
)
```

---

## Advanced Patterns

### Complex Query with Everything

```go
orders, err := database.Query[Order](db).
    Select("orders.*", "u.name as user_name", "p.title as product_name").
    Join("users", "u").
        On("orders.user_id", "=", "u.id").
        End().
    LeftJoin("products", "p").
        On("orders.product_id", "=", "p.id").
        End().
    Where("orders.status", "completed").
    WhereOp("orders.total", ">=", 100).
    WhereRaw("orders.created_at > NOW() - INTERVAL '30 days'").
    Or().
        Where("orders.priority", "high").
        Where("orders.priority", "urgent").
        End().
    OrderBy("orders.created_at", database.DESC).
    GroupBy("orders.id", "u.name", "p.title").
    Limit(50).
    Distinct().
    All(ctx)
```

### Query with Timeout

```go
users, err := database.Query[User](db).
    Where("active", true).
    Timeout(5 * time.Second).
    All(ctx)
```

### Query with Row Locking

```go
user, err := database.Query[User](db).
    Where("id", userId).
    ForUpdate(). // Locks the row for update
    First(ctx)
```

### Subqueries (Using Raw SQL)

```go
users, err := database.Query[User](db).
    WhereRaw(`id IN (
        SELECT user_id 
        FROM orders 
        WHERE total > ? 
        GROUP BY user_id 
        HAVING COUNT(*) > ?
    )`, 1000, 5).
    All(ctx)
```

### Dynamic Query Building

```go
query := database.Query[Product](db)

// Conditionally add filters
if categoryID != 0 {
    query = query.Where("category_id", categoryID)
}

if minPrice > 0 {
    query = query.WhereOp("price", ">=", minPrice)
}

if maxPrice > 0 {
    query = query.WhereOp("price", "<=", maxPrice)
}

if searchTerm != "" {
    query = query.WhereLike("name", "%"+searchTerm+"%")
}

products, err := query.
    OrderBy("created_at", database.DESC).
    Limit(20).
    All(ctx)
```

### Aggregations with Group By

```go
type CategoryStats struct {
    CategoryID   int     `json:"category_id"`
    ProductCount int     `json:"product_count"`
    AvgPrice     float64 `json:"avg_price"`
}

stats, err := database.RawQuery[CategoryStats](db, ctx, `
    SELECT 
        category_id,
        COUNT(*) as product_count,
        AVG(price) as avg_price
    FROM products
    WHERE active = true
    GROUP BY category_id
    HAVING COUNT(*) > 10
    ORDER BY product_count DESC
`)
```

---

## Best Practices

1. **Always use context**: Pass context for cancellation and timeout support
2. **Use transactions for related operations**: Ensure data consistency
3. **Prefer type-safe joins over raw SQL**: More maintainable and safer
4. **Use helpers for simple operations**: `FindByID`, `Create`, etc.
5. **Handle nil results**: `First()` returns nil if no record found
6. **Use pagination for large datasets**: Prevent memory issues
7. **Use soft deletes when appropriate**: Allows data recovery
8. **Add timeouts for long-running queries**: Prevent hanging
9. **Use batch/chunk processing for bulk operations**: Better performance

---

## Migration from Old System

### Before (Old System)

```go
result, err := database.ExecuteQuery[User](
    structs.NewQuery().
        SetOperation("select").
        SetTable("users").
        AddWhere("active", true).
        AddJoin("JOIN orders ON orders.user_id = users.id").
        AddOrder("created_at DESC").
        SetLimit(10),
)
```

### After (New System)

```go
users, err := database.Query[User](db).
    Join("orders", "o").
        On("o.user_id", "=", "users.id").
        End().
    Where("active", true).
    OrderBy("created_at", database.DESC).
    Limit(10).
    All(ctx)
```

Much more natural and type-safe! 🎉
