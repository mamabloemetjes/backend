package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
)

// All executes the query and returns all matching records
func (q *QueryBuilder[T]) All(ctx context.Context) ([]T, error) {
	start := time.Now()
	var data []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	pgQuery := q.db.ModelContext(ctx, &data)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply SELECT columns
	if len(q.selectCols) > 0 {
		for _, col := range q.selectCols {
			pgQuery = pgQuery.Column(col)
		}
	}

	// Apply DISTINCT
	if q.distinct {
		pgQuery = pgQuery.Distinct()
	}

	// Apply JOINs
	for _, join := range q.joins {
		pgQuery = pgQuery.Join(join.toSQL())
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Apply GROUP BY
	for _, groupBy := range q.groupBys {
		pgQuery = pgQuery.Group(groupBy)
	}

	// Apply HAVING
	for _, having := range q.havings {
		if having.IsRaw {
			pgQuery = pgQuery.Having(having.RawSQL, having.RawArgs...)
		} else {
			pgQuery = pgQuery.Having(fmt.Sprintf("%s %s ?", having.Column, having.Operator), having.Value)
		}
	}

	// Apply ORDER BY
	for _, order := range q.orders {
		pgQuery = pgQuery.Order(fmt.Sprintf("%s %s", order.Column, order.Direction))
	}

	// Apply LIMIT
	if q.limitVal != nil {
		pgQuery = pgQuery.Limit(*q.limitVal)
	}

	// Apply OFFSET
	if q.offsetVal != nil {
		pgQuery = pgQuery.Offset(*q.offsetVal)
	}

	// Apply relations (preloading)
	for _, relation := range q.relations {
		pgQuery = pgQuery.Relation(relation)
	}

	// Apply FOR UPDATE
	if q.forUpdate {
		pgQuery = pgQuery.For("UPDATE")
	}

	// Execute query
	err := pgQuery.Select()
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// First executes the query and returns the first matching record
func (q *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	start := time.Now()
	var data T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	pgQuery := q.db.ModelContext(ctx, &data)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply SELECT columns
	if len(q.selectCols) > 0 {
		for _, col := range q.selectCols {
			pgQuery = pgQuery.Column(col)
		}
	}

	// Apply JOINs
	for _, join := range q.joins {
		pgQuery = pgQuery.Join(join.toSQL())
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Apply ORDER BY
	for _, order := range q.orders {
		pgQuery = pgQuery.Order(fmt.Sprintf("%s %s", order.Column, order.Direction))
	}

	// Apply relations (preloading)
	for _, relation := range q.relations {
		pgQuery = pgQuery.Relation(relation)
	}

	// Execute query
	err := pgQuery.First()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Return nil without error for no results
		}
		return nil, fmt.Errorf("failed to execute first query: %w (took %v)", err, time.Since(start))
	}

	return &data, nil
}

// Count executes the query and returns the count of matching records
func (q *QueryBuilder[T]) Count(ctx context.Context) (int, error) {
	start := time.Now()
	var data []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	pgQuery := q.db.ModelContext(ctx, &data)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply JOINs
	for _, join := range q.joins {
		pgQuery = pgQuery.Join(join.toSQL())
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Execute count query
	count, err := pgQuery.Count()
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

// Insert inserts a new record and returns it
func (q *QueryBuilder[T]) Insert(ctx context.Context, data *T) (*T, error) {
	start := time.Now()

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	pgQuery := q.db.ModelContext(ctx, data)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Execute insert
	_, err := pgQuery.Insert()
	if err != nil {
		return nil, fmt.Errorf("failed to execute insert query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// InsertMany inserts multiple records
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

	// Build the query
	pgQuery := q.db.ModelContext(ctx, &data)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Execute bulk insert
	_, err := pgQuery.Insert()
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk insert query: %w (took %v)", err, time.Since(start))
	}

	return data, nil
}

// Update updates records matching the query
func (q *QueryBuilder[T]) Update(ctx context.Context, data any) (int, error) {
	start := time.Now()

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	var model T
	pgQuery := q.db.ModelContext(ctx, &model)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Handle data based on type
	switch v := data.(type) {
	case map[string]any:
		// Update with map
		for key, value := range v {
			pgQuery = pgQuery.Set("? = ?", pg.Ident(key), value)
		}
	case *T:
		// Update with struct
		pgQuery = pgQuery.Model(v)
	default:
		return 0, fmt.Errorf("unsupported data type for update: %T", data)
	}

	// Execute update
	res, err := pgQuery.Update()
	if err != nil {
		return 0, fmt.Errorf("failed to execute update query: %w (took %v)", err, time.Since(start))
	}

	return res.RowsAffected(), nil
}

// UpdateReturning updates records and returns them
func (q *QueryBuilder[T]) UpdateReturning(ctx context.Context, data any) ([]T, error) {
	start := time.Now()
	var results []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	pgQuery := q.db.ModelContext(ctx, &results)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Handle data based on type
	switch v := data.(type) {
	case map[string]any:
		// Update with map
		for key, value := range v {
			pgQuery = pgQuery.Set("? = ?", pg.Ident(key), value)
		}
	case *T:
		// Update with struct
		pgQuery = pgQuery.Model(v)
	default:
		return nil, fmt.Errorf("unsupported data type for update: %T", data)
	}

	// Add RETURNING *
	pgQuery = pgQuery.Returning("*")

	// Execute update
	_, err := pgQuery.Update()
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w (took %v)", err, time.Since(start))
	}

	return results, nil
}

// Delete deletes records matching the query
func (q *QueryBuilder[T]) Delete(ctx context.Context) (int, error) {
	start := time.Now()

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	var model T
	pgQuery := q.db.ModelContext(ctx, &model)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Execute delete
	res, err := pgQuery.Delete()
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete query: %w (took %v)", err, time.Since(start))
	}

	return res.RowsAffected(), nil
}

// DeleteReturning deletes records and returns them
func (q *QueryBuilder[T]) DeleteReturning(ctx context.Context) ([]T, error) {
	start := time.Now()
	var results []T

	// Apply timeout if specified
	if q.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	// Build the query
	pgQuery := q.db.ModelContext(ctx, &results)

	// Apply table name if specified
	if q.tableName != "" {
		pgQuery = pgQuery.Table(q.tableName)
	}

	// Apply WHERE conditions
	pgQuery = q.applyWhereConditions(pgQuery)

	// Add RETURNING *
	pgQuery = pgQuery.Returning("*")

	// Execute delete
	_, err := pgQuery.Delete()
	if err != nil {
		return nil, fmt.Errorf("failed to execute delete query: %w (took %v)", err, time.Since(start))
	}

	return results, nil
}

// applyWhereConditions applies WHERE conditions to a go-pg query
func (q *QueryBuilder[T]) applyWhereConditions(pgQuery *pg.Query) *pg.Query {
	// Apply simple WHERE conditions
	for _, where := range q.wheres {
		if where.IsRaw {
			pgQuery = pgQuery.Where(where.RawSQL, where.RawArgs...)
		} else {
			var condition string
			if where.Negate {
				condition = fmt.Sprintf("NOT (%s %s ?)", where.Column, where.Operator)
			} else {
				if where.Operator == "IS NULL" || where.Operator == "IS NOT NULL" {
					condition = fmt.Sprintf("%s %s", where.Column, where.Operator)
					pgQuery = pgQuery.Where(condition)
					continue
				}
				condition = fmt.Sprintf("%s %s ?", where.Column, where.Operator)
			}
			pgQuery = pgQuery.Where(condition, where.Value)
		}
	}

	// Apply WHERE groups
	for _, group := range q.whereGroups {
		pgQuery = q.applyWhereGroup(pgQuery, group)
	}

	return pgQuery
}

// applyWhereGroup applies a WHERE group to a go-pg query
func (q *QueryBuilder[T]) applyWhereGroup(pgQuery *pg.Query, group *WhereGroup) *pg.Query {
	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return pgQuery
	}

	var conditions []string
	var args []any

	// Build conditions
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

	// Build group SQL
	if len(conditions) > 0 {
		groupSQL := "(" + joinStrings(conditions, " "+group.Connector+" ") + ")"
		if group.Negate {
			groupSQL = "NOT " + groupSQL
		}
		pgQuery = pgQuery.Where(groupSQL, args...)
	}

	return pgQuery
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
