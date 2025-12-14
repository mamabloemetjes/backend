package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
)

// RawQuery executes a raw SQL query and returns results
func RawQuery[T any](db *DB, ctx context.Context, sql string, args ...any) ([]T, error) {
	start := time.Now()
	var data []T

	_, err := db.QueryContext(ctx, &data, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute raw query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// RawQueryOne executes a raw SQL query and returns a single result
func RawQueryOne[T any](db *DB, ctx context.Context, sql string, args ...any) (*T, error) {
	start := time.Now()
	var data T

	_, err := db.QueryOneContext(ctx, &data, sql, args...)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to execute raw query: %w (took %v)", err, time.Since(start))
	}

	return &data, nil
}

// RawExec executes a raw SQL command (INSERT, UPDATE, DELETE) without returning data
func RawExec(db *DB, ctx context.Context, sql string, args ...any) (int, error) {
	start := time.Now()

	res, err := db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute raw command: %w (took %v)", err, time.Since(start))
	}

	return res.RowsAffected(), nil
}

// Transaction executes a function within a database transaction
func Transaction(ctx context.Context, fn func(*pg.Tx) error) error {
	db := GetInstance()
	if db == nil {
		return fmt.Errorf("database instance not initialized")
	}

	return db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return fn(tx)
	})
}

// TransactionWithResult executes a function within a transaction and returns a result
func TransactionWithResult[T any](ctx context.Context, fn func(*pg.Tx) (T, error)) (T, error) {
	var result T
	db := GetInstance()
	if db == nil {
		return result, fmt.Errorf("database instance not initialized")
	}

	err := db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// PaginationResult wraps paginated data with metadata
type PaginationResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Paginate applies pagination to a query builder and returns results with metadata
func Paginate[T any](q *QueryBuilder[T], ctx context.Context, page, pageSize int) (*PaginationResult[T], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	// Get total count
	total, err := q.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated data
	data, err := q.Limit(pageSize).Offset(offset).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated data: %w", err)
	}

	return &PaginationResult[T]{
		Data: data,
		Pagination: Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

// CursorPagination represents cursor-based pagination parameters
type CursorPagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	PageSize   int    `json:"page_size"`
}

// CursorPaginationResult wraps cursor-paginated data with metadata
type CursorPaginationResult[T any] struct {
	Data       []T              `json:"data"`
	Pagination CursorPagination `json:"pagination"`
}

// FindByID is a helper to find a record by ID
func FindByID[T any](db *DB, ctx context.Context, id any) (*T, error) {
	return Query[T](db).Where("id", id).First(ctx)
}

// FindByIDs is a helper to find multiple records by IDs
func FindByIDs[T any](db *DB, ctx context.Context, ids []any) ([]T, error) {
	return Query[T](db).WhereIn("id", ids).All(ctx)
}

// Create is a helper to insert a single record
func Create[T any](db *DB, ctx context.Context, data *T) (*T, error) {
	return Query[T](db).Insert(ctx, data)
}

// CreateMany is a helper to insert multiple records
func CreateMany[T any](db *DB, ctx context.Context, data []T) ([]T, error) {
	return Query[T](db).InsertMany(ctx, data)
}

// UpdateByID is a helper to update a record by ID
func UpdateByID[T any](db *DB, ctx context.Context, id any, data map[string]any) (int, error) {
	return Query[T](db).Where("id", id).Update(ctx, data)
}

// DeleteByID is a helper to delete a record by ID
func DeleteByID[T any](db *DB, ctx context.Context, id any) (int, error) {
	return Query[T](db).Where("id", id).Delete(ctx)
}

// SoftDelete performs a soft delete by setting deleted_at timestamp
func SoftDelete[T any](db *DB, ctx context.Context, id any) (int, error) {
	return Query[T](db).
		Where("id", id).
		Update(ctx, map[string]any{
			"deleted_at": time.Now(),
		})
}

// Restore restores a soft-deleted record
func Restore[T any](db *DB, ctx context.Context, id any) (int, error) {
	return Query[T](db).
		Where("id", id).
		Update(ctx, map[string]any{
			"deleted_at": nil,
		})
}

// ExcludeSoftDeleted adds a WHERE clause to exclude soft-deleted records
func ExcludeSoftDeleted[T any](q *QueryBuilder[T]) *QueryBuilder[T] {
	return q.WhereNull("deleted_at")
}

// OnlySoftDeleted adds a WHERE clause to only include soft-deleted records
func OnlySoftDeleted[T any](q *QueryBuilder[T]) *QueryBuilder[T] {
	return q.WhereNotNull("deleted_at")
}

// BatchProcess processes records in batches
func BatchProcess[T any](ctx context.Context, query *QueryBuilder[T], batchSize int, fn func([]T) error) error {
	if batchSize < 1 {
		batchSize = 100
	}

	offset := 0
	for {
		batch, err := query.Limit(batchSize).Offset(offset).All(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch batch at offset %d: %w", offset, err)
		}

		if len(batch) == 0 {
			break
		}

		if err := fn(batch); err != nil {
			return fmt.Errorf("batch processing failed at offset %d: %w", offset, err)
		}

		if len(batch) < batchSize {
			break
		}

		offset += batchSize
	}

	return nil
}

// Upsert performs an INSERT ... ON CONFLICT DO UPDATE operation
func Upsert[T any](db *DB, ctx context.Context, data *T, conflictColumn string, updateColumns ...string) (*T, error) {
	start := time.Now()

	// Build the base insert query
	pgQuery := db.ModelContext(ctx, data)

	// Build ON CONFLICT clause
	conflictClause := fmt.Sprintf("(%s) DO UPDATE", conflictColumn)

	// Add SET clause for update columns
	if len(updateColumns) > 0 {
		setClause := " SET "
		for i, col := range updateColumns {
			if i > 0 {
				setClause += ", "
			}
			setClause += fmt.Sprintf("%s = EXCLUDED.%s", col, col)
		}
		conflictClause += setClause
	} else {
		// If no columns specified, update all columns
		conflictClause += " DO NOTHING"
	}

	pgQuery = pgQuery.OnConflict(conflictClause)

	// Execute upsert
	_, err := pgQuery.Insert()
	if err != nil {
		return nil, fmt.Errorf("failed to execute upsert: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// BulkUpsert performs bulk INSERT ... ON CONFLICT DO UPDATE
func BulkUpsert[T any](db *DB, ctx context.Context, data []T, conflictColumn string, updateColumns ...string) ([]T, error) {
	start := time.Now()

	if len(data) == 0 {
		return data, nil
	}

	// Build the base insert query
	pgQuery := db.ModelContext(ctx, &data)

	// Build ON CONFLICT clause
	conflictClause := fmt.Sprintf("(%s) DO UPDATE", conflictColumn)

	// Add SET clause for update columns
	if len(updateColumns) > 0 {
		setClause := " SET "
		for i, col := range updateColumns {
			if i > 0 {
				setClause += ", "
			}
			setClause += fmt.Sprintf("%s = EXCLUDED.%s", col, col)
		}
		conflictClause += setClause
	} else {
		// If no columns specified, update all columns
		conflictClause += " DO NOTHING"
	}

	pgQuery = pgQuery.OnConflict(conflictClause)

	// Execute bulk upsert
	_, err := pgQuery.Insert()
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk upsert: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// Chunk executes a callback for each chunk of results
func Chunk[T any](ctx context.Context, query *QueryBuilder[T], chunkSize int, fn func([]T, int) error) error {
	if chunkSize < 1 {
		chunkSize = 100
	}

	offset := 0
	chunkNumber := 0

	for {
		chunk, err := query.Limit(chunkSize).Offset(offset).All(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch chunk at offset %d: %w", offset, err)
		}

		if len(chunk) == 0 {
			break
		}

		if err := fn(chunk, chunkNumber); err != nil {
			return fmt.Errorf("chunk processing failed at chunk %d: %w", chunkNumber, err)
		}

		if len(chunk) < chunkSize {
			break
		}

		offset += chunkSize
		chunkNumber++
	}

	return nil
}

// Pluck extracts a single column from query results
func Pluck[T any, R any](ctx context.Context, query *QueryBuilder[T], column string) ([]R, error) {
	var results []R

	// Execute query with only the specified column
	_, err := query.Select(column).All(ctx)
	if err != nil {
		return nil, err
	}

	// This is a simplified version - in practice you'd need reflection
	// to extract the column value from each struct
	return results, fmt.Errorf("pluck not fully implemented - use Select() with specific columns")
}

// FirstOrCreate finds the first record matching conditions or creates it
func FirstOrCreate[T any](db *DB, ctx context.Context, conditions map[string]any, defaults map[string]any) (*T, error) {
	// Try to find existing record
	q := Query[T](db)
	for key, value := range conditions {
		q = q.Where(key, value)
	}

	existing, err := q.First(ctx)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	// Record doesn't exist, create it
	// Merge conditions and defaults
	data := make(map[string]any)
	for k, v := range conditions {
		data[k] = v
	}
	for k, v := range defaults {
		data[k] = v
	}

	// Note: This is simplified - in practice you'd need to convert map to struct
	return nil, fmt.Errorf("firstOrCreate not fully implemented - use Insert() directly")
}

// UpdateOrCreate updates an existing record or creates a new one
func UpdateOrCreate[T any](db *DB, ctx context.Context, conditions map[string]any, data map[string]any) (*T, error) {
	// Try to find and update existing record
	q := Query[T](db)
	for key, value := range conditions {
		q = q.Where(key, value)
	}

	results, err := q.UpdateReturning(ctx, data)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		return &results[0], nil
	}

	// Record doesn't exist, create it
	// Merge conditions and data
	merged := make(map[string]any)
	for k, v := range conditions {
		merged[k] = v
	}
	for k, v := range data {
		merged[k] = v
	}

	// Note: This is simplified - in practice you'd need to convert map to struct
	return nil, fmt.Errorf("updateOrCreate not fully implemented - use Insert() directly")
}
