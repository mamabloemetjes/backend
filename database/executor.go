package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// All executes the query and returns all matching records with automatic retry
func (q *QueryBuilder[T]) All(ctx context.Context) ([]T, error) {
	start := time.Now()
	var data []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		data = nil // Reset on retry

		// When relations are being preloaded, we need to use Model() with the slice
		// This is required for has-many and many-to-many relationships
		if len(q.relations) > 0 {
			query := q.buildBunQueryWithModel(&data)
			return query.Scan(ctx)
		}

		// No relations, use the regular buildBunQuery approach
		query := q.buildBunQuery()
		return query.Scan(ctx, &data)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// First executes the query and returns the first matching record with automatic retry
func (q *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	start := time.Now()
	var data T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		query := q.buildBunQuery().Limit(1)
		return query.Scan(ctx, &data)
	})

	if err != nil {
		// Return nil for no rows instead of error
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to execute first query: %w (took %v)", err, time.Since(start))
	}

	return &data, nil
}

// Count executes the query and returns the count of matching records with automatic retry
func (q *QueryBuilder[T]) Count(ctx context.Context) (int, error) {
	start := time.Now()
	var count int

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		query := q.buildBunQuery()
		var err error
		count, err = query.Count(ctx)
		return err
	})

	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w (took %v)", err, time.Since(start))
	}

	return count, nil
}

// Exists checks if any records match the query
func (q *QueryBuilder[T]) Exists(ctx context.Context) (bool, error) {
	count, err := q.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Insert inserts a new record and returns it with automatic retry
func (q *QueryBuilder[T]) Insert(ctx context.Context, data *T) (*T, error) {
	start := time.Now()

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		query := q.db.NewInsert().Model(data)

		if q.tableName != "" {
			query = query.Table(q.tableName)
		}

		_, err := query.Exec(ctx)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute insert query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// InsertMany inserts multiple records with automatic retry
func (q *QueryBuilder[T]) InsertMany(ctx context.Context, data []T) ([]T, error) {
	start := time.Now()

	if len(data) == 0 {
		return data, nil
	}

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		query := q.db.NewInsert().Model(&data)

		if q.tableName != "" {
			query = query.Table(q.tableName)
		}

		_, err := query.Exec(ctx)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk insert query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// Update updates records matching the query with automatic retry
func (q *QueryBuilder[T]) Update(ctx context.Context, data any) (int, error) {
	start := time.Now()
	var rowsAffected int64

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		var model T
		query := q.db.NewUpdate().Model(&model)

		if q.tableName != "" {
			query = query.Table(q.tableName)
		}

		// Apply WHERE conditions using the same helper
		query = q.applyWhereConditionsToUpdate(query)

		// Handle data based on type
		switch v := data.(type) {
		case map[string]any:
			// Update with map
			for key, value := range v {
				query = query.Set("? = ?", bun.Ident(key), value)
			}
		case *T:
			// Update with struct
			query = query.Model(v)
		default:
			return fmt.Errorf("unsupported data type for update: %T", data)
		}

		res, err := query.Exec(ctx)
		if err != nil {
			return err
		}
		rowsAffected, _ = res.RowsAffected()
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to execute update query: %w (took %v)", err, time.Since(start))
	}

	return int(rowsAffected), nil
}

// UpdateReturning updates records and returns them with automatic retry
func (q *QueryBuilder[T]) UpdateReturning(ctx context.Context, data any) ([]T, error) {
	start := time.Now()
	var results []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		results = nil // Reset on retry
		var model T
		query := q.db.NewUpdate().Model(&model)

		if q.tableName != "" {
			query = query.Table(q.tableName)
		}

		// Apply WHERE conditions
		query = q.applyWhereConditionsToUpdate(query)

		// Handle data based on type
		switch v := data.(type) {
		case map[string]any:
			for key, value := range v {
				query = query.Set("? = ?", bun.Ident(key), value)
			}
		case *T:
			query = query.Model(v)
		default:
			return fmt.Errorf("unsupported data type for update: %T", data)
		}

		// Add RETURNING *
		query = query.Returning("*")

		_, err := query.Exec(ctx, &results)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w (took %v)", err, time.Since(start))
	}

	return results, nil
}

// Delete deletes records matching the query with automatic retry
func (q *QueryBuilder[T]) Delete(ctx context.Context) (int, error) {
	start := time.Now()
	var rowsAffected int64

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		var model T
		query := q.db.NewDelete().Model(&model)

		if q.tableName != "" {
			query = query.Table(q.tableName)
		}

		// Apply WHERE conditions
		query = q.applyWhereConditionsToDelete(query)

		res, err := query.Exec(ctx)
		if err != nil {
			return err
		}
		rowsAffected, _ = res.RowsAffected()
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to execute delete query: %w (took %v)", err, time.Since(start))
	}

	return int(rowsAffected), nil
}

// DeleteReturning deletes records and returns them with automatic retry
func (q *QueryBuilder[T]) DeleteReturning(ctx context.Context) ([]T, error) {
	start := time.Now()
	var results []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	err := WithRetry(ctx, func() error {
		results = nil // Reset on retry
		var model T
		query := q.db.NewDelete().Model(&model)

		if q.tableName != "" {
			query = query.Table(q.tableName)
		}

		// Apply WHERE conditions
		query = q.applyWhereConditionsToDelete(query)

		// Add RETURNING *
		query = query.Returning("*")

		_, err := query.Exec(ctx, &results)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute delete query: %w (took %v)", err, time.Since(start))
	}

	return results, nil
}

// applyWhereConditionsToUpdate applies WHERE conditions to a Bun UpdateQuery
func (q *QueryBuilder[T]) applyWhereConditionsToUpdate(query *bun.UpdateQuery) *bun.UpdateQuery {
	// Apply simple WHERE conditions
	for _, where := range q.wheres {
		if where.IsRaw {
			query = query.Where(where.RawSQL, where.RawArgs...)
		} else {
			var condition string
			if where.Negate {
				condition = fmt.Sprintf("NOT (%s %s ?)", where.Column, where.Operator)
			} else {
				if where.Operator == "IS NULL" || where.Operator == "IS NOT NULL" {
					condition = fmt.Sprintf("%s %s", where.Column, where.Operator)
					query = query.Where(condition)
					continue
				}
				condition = fmt.Sprintf("%s %s ?", where.Column, where.Operator)
			}
			query = query.Where(condition, where.Value)
		}
	}

	// Apply WHERE groups
	for _, group := range q.whereGroups {
		query = q.applyWhereGroupToUpdate(query, group)
	}

	return query
}

// applyWhereGroupToUpdate applies a WHERE group to a Bun UpdateQuery
func (q *QueryBuilder[T]) applyWhereGroupToUpdate(query *bun.UpdateQuery, group *WhereGroup) *bun.UpdateQuery {
	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return query
	}

	var conditions []string
	var args []any

	for _, cond := range group.Conditions {
		if cond.IsRaw {
			conditions = append(conditions, cond.RawSQL)
			args = append(args, cond.RawArgs...)
		} else {
			var condStr string
			if cond.Operator == "IS NULL" || cond.Operator == "IS NOT NULL" {
				condStr = fmt.Sprintf("%s %s", cond.Column, cond.Operator)
			} else {
				condStr = fmt.Sprintf("%s %s ?", cond.Column, cond.Operator)
				args = append(args, cond.Value)
			}
			conditions = append(conditions, condStr)
		}
	}

	if len(conditions) > 0 {
		groupSQL := "(" + joinStrings(conditions, " "+group.Connector+" ") + ")"
		if group.Negate {
			groupSQL = "NOT " + groupSQL
		}
		query = query.Where(groupSQL, args...)
	}

	return query
}

// applyWhereConditionsToDelete applies WHERE conditions to a Bun DeleteQuery
func (q *QueryBuilder[T]) applyWhereConditionsToDelete(query *bun.DeleteQuery) *bun.DeleteQuery {
	// Apply simple WHERE conditions
	for _, where := range q.wheres {
		if where.IsRaw {
			query = query.Where(where.RawSQL, where.RawArgs...)
		} else {
			var condition string
			if where.Negate {
				condition = fmt.Sprintf("NOT (%s %s ?)", where.Column, where.Operator)
			} else {
				if where.Operator == "IS NULL" || where.Operator == "IS NOT NULL" {
					condition = fmt.Sprintf("%s %s", where.Column, where.Operator)
					query = query.Where(condition)
					continue
				}
				condition = fmt.Sprintf("%s %s ?", where.Column, where.Operator)
			}
			query = query.Where(condition, where.Value)
		}
	}

	// Apply WHERE groups
	for _, group := range q.whereGroups {
		query = q.applyWhereGroupToDelete(query, group)
	}

	return query
}

// applyWhereGroupToDelete applies a WHERE group to a Bun DeleteQuery
func (q *QueryBuilder[T]) applyWhereGroupToDelete(query *bun.DeleteQuery, group *WhereGroup) *bun.DeleteQuery {
	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return query
	}

	var conditions []string
	var args []any

	for _, cond := range group.Conditions {
		if cond.IsRaw {
			conditions = append(conditions, cond.RawSQL)
			args = append(args, cond.RawArgs...)
		} else {
			var condStr string
			if cond.Operator == "IS NULL" || cond.Operator == "IS NOT NULL" {
				condStr = fmt.Sprintf("%s %s", cond.Column, cond.Operator)
			} else {
				condStr = fmt.Sprintf("%s %s ?", cond.Column, cond.Operator)
				args = append(args, cond.Value)
			}
			conditions = append(conditions, condStr)
		}
	}

	if len(conditions) > 0 {
		groupSQL := "(" + joinStrings(conditions, " "+group.Connector+" ") + ")"
		if group.Negate {
			groupSQL = "NOT " + groupSQL
		}
		query = query.Where(groupSQL, args...)
	}

	return query
}
