package database

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// JoinType represents the type of SQL JOIN operation
type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullJoin
)

// String returns the SQL representation of the join type
func (jt JoinType) String() string {
	switch jt {
	case InnerJoin:
		return "INNER JOIN"
	case LeftJoin:
		return "LEFT JOIN"
	case RightJoin:
		return "RIGHT JOIN"
	case FullJoin:
		return "FULL JOIN"
	default:
		return "INNER JOIN"
	}
}

// QueryBuilder provides a fluent, type-safe API for building database queries
type QueryBuilder[T any] struct {
	db        *DB
	ctx       context.Context
	model     *T
	tableName string

	// Query clauses
	selectCols  []string
	joins       []*JoinClause
	wheres      []*WhereClause
	whereGroups []*WhereGroup
	orders      []*OrderClause
	groupBys    []string
	havings     []*WhereClause
	limitVal    *int
	offsetVal   *int

	// Relations to preload
	relations []string

	// Options
	distinct  bool
	forUpdate bool

	// Transaction
	tx any

	// Timeout
	timeout time.Duration
}

// JoinClause represents a SQL JOIN operation
type JoinClause struct {
	Type       JoinType
	Table      string
	Alias      string
	Conditions []*JoinCondition
}

// JoinCondition represents a condition in a JOIN clause
type JoinCondition struct {
	Left     string
	Operator string
	Right    string
	IsValue  bool // If true, Right is a value; otherwise it's a column
}

// WhereClause represents a WHERE condition
type WhereClause struct {
	Column   string
	Operator string
	Value    any
	IsRaw    bool
	RawSQL   string
	RawArgs  []any
	Negate   bool // For NOT conditions
}

// WhereGroup represents a grouped WHERE condition (for OR/AND grouping)
type WhereGroup struct {
	Conditions []*WhereClause
	Groups     []*WhereGroup
	Connector  string // "AND" or "OR"
	Negate     bool
}

// OrderClause represents an ORDER BY clause
type OrderClause struct {
	Column    string
	Direction string // "ASC" or "DESC"
}

// OrderDirection represents sort direction
type OrderDirection string

const (
	ASC  OrderDirection = "ASC"
	DESC OrderDirection = "DESC"
)

// JoinBuilder provides a fluent API for building JOIN clauses
type JoinBuilder[T any] struct {
	parent *QueryBuilder[T]
	clause *JoinClause
}

// WhereGroupBuilder provides a fluent API for building grouped WHERE clauses
type WhereGroupBuilder[T any] struct {
	parent *QueryBuilder[T]
	group  *WhereGroup
}

// QueryResult represents the result of a database operation
type QueryResult[T any] struct {
	Data          []T           `json:"data,omitempty"`
	Single        *T            `json:"single,omitempty"`
	Count         int64         `json:"count"`
	Success       bool          `json:"success"`
	Error         error         `json:"error,omitempty"`
	ExecutionTime time.Duration `json:"execution_time"`
	Query         string        `json:"query,omitempty"`
	Args          []any         `json:"args,omitempty"`
}

// Query creates a new QueryBuilder instance
func Query[T any](db *DB) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		db:          db,
		ctx:         context.Background(),
		selectCols:  []string{},
		joins:       []*JoinClause{},
		wheres:      []*WhereClause{},
		whereGroups: []*WhereGroup{},
		orders:      []*OrderClause{},
		groupBys:    []string{},
		havings:     []*WhereClause{},
		relations:   []string{},
	}
}

// Context sets the context for the query
func (q *QueryBuilder[T]) Context(ctx context.Context) *QueryBuilder[T] {
	q.ctx = ctx
	return q
}

// Table sets the table name explicitly
func (q *QueryBuilder[T]) Table(name string) *QueryBuilder[T] {
	q.tableName = name
	return q
}

// Select specifies the columns to select
func (q *QueryBuilder[T]) Select(columns ...string) *QueryBuilder[T] {
	q.selectCols = append(q.selectCols, columns...)
	return q
}

// Distinct adds DISTINCT to the query
func (q *QueryBuilder[T]) Distinct() *QueryBuilder[T] {
	q.distinct = true
	return q
}

// Join starts building an INNER JOIN clause
func (q *QueryBuilder[T]) Join(table, alias string) *JoinBuilder[T] {
	clause := &JoinClause{
		Type:       InnerJoin,
		Table:      table,
		Alias:      alias,
		Conditions: []*JoinCondition{},
	}
	return &JoinBuilder[T]{
		parent: q,
		clause: clause,
	}
}

// LeftJoin starts building a LEFT JOIN clause
func (q *QueryBuilder[T]) LeftJoin(table, alias string) *JoinBuilder[T] {
	clause := &JoinClause{
		Type:       LeftJoin,
		Table:      table,
		Alias:      alias,
		Conditions: []*JoinCondition{},
	}
	return &JoinBuilder[T]{
		parent: q,
		clause: clause,
	}
}

// RightJoin starts building a RIGHT JOIN clause
func (q *QueryBuilder[T]) RightJoin(table, alias string) *JoinBuilder[T] {
	clause := &JoinClause{
		Type:       RightJoin,
		Table:      table,
		Alias:      alias,
		Conditions: []*JoinCondition{},
	}
	return &JoinBuilder[T]{
		parent: q,
		clause: clause,
	}
}

// FullJoin starts building a FULL JOIN clause
func (q *QueryBuilder[T]) FullJoin(table, alias string) *JoinBuilder[T] {
	clause := &JoinClause{
		Type:       FullJoin,
		Table:      table,
		Alias:      alias,
		Conditions: []*JoinCondition{},
	}
	return &JoinBuilder[T]{
		parent: q,
		clause: clause,
	}
}

// Where adds a simple WHERE condition (column = value)
func (q *QueryBuilder[T]) Where(column string, value any) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "=",
		Value:    value,
	})
	return q
}

// WhereOp adds a WHERE condition with a custom operator
func (q *QueryBuilder[T]) WhereOp(column, operator string, value any) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return q
}

// WhereNot adds a WHERE NOT condition
func (q *QueryBuilder[T]) WhereNot(column string, value any) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "=",
		Value:    value,
		Negate:   true,
	})
	return q
}

// WhereIn adds a WHERE IN condition
func (q *QueryBuilder[T]) WhereIn(column string, values []any) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "IN",
		Value:    values,
	})
	return q
}

// WhereNotIn adds a WHERE NOT IN condition
func (q *QueryBuilder[T]) WhereNotIn(column string, values []any) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "IN",
		Value:    values,
		Negate:   true,
	})
	return q
}

// WhereNull adds a WHERE IS NULL condition
func (q *QueryBuilder[T]) WhereNull(column string) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "IS NULL",
	})
	return q
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (q *QueryBuilder[T]) WhereNotNull(column string) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "IS NOT NULL",
	})
	return q
}

// WhereLike adds a WHERE LIKE condition
func (q *QueryBuilder[T]) WhereLike(column, pattern string) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		Column:   column,
		Operator: "LIKE",
		Value:    pattern,
	})
	return q
}

// WhereRaw adds a raw WHERE condition
func (q *QueryBuilder[T]) WhereRaw(sql string, args ...any) *QueryBuilder[T] {
	q.wheres = append(q.wheres, &WhereClause{
		IsRaw:   true,
		RawSQL:  sql,
		RawArgs: args,
	})
	return q
}

// WhereGroup starts building a grouped WHERE clause
func (q *QueryBuilder[T]) WhereGroup(connector string) *WhereGroupBuilder[T] {
	group := &WhereGroup{
		Conditions: []*WhereClause{},
		Groups:     []*WhereGroup{},
		Connector:  connector,
	}
	return &WhereGroupBuilder[T]{
		parent: q,
		group:  group,
	}
}

// Or starts an OR group
func (q *QueryBuilder[T]) Or() *WhereGroupBuilder[T] {
	return q.WhereGroup("OR")
}

// OrderBy adds an ORDER BY clause
func (q *QueryBuilder[T]) OrderBy(column string, direction OrderDirection) *QueryBuilder[T] {
	q.orders = append(q.orders, &OrderClause{
		Column:    column,
		Direction: string(direction),
	})
	return q
}

// GroupBy adds a GROUP BY clause
func (q *QueryBuilder[T]) GroupBy(columns ...string) *QueryBuilder[T] {
	q.groupBys = append(q.groupBys, columns...)
	return q
}

// Having adds a HAVING clause
func (q *QueryBuilder[T]) Having(column, operator string, value any) *QueryBuilder[T] {
	q.havings = append(q.havings, &WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return q
}

// Limit sets the LIMIT clause
func (q *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	q.limitVal = &limit
	return q
}

// Offset sets the OFFSET clause
func (q *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	q.offsetVal = &offset
	return q
}

// With specifies a relation to preload (go-pg style)
func (q *QueryBuilder[T]) With(relation string) *QueryBuilder[T] {
	q.relations = append(q.relations, relation)
	return q
}

// ForUpdate adds FOR UPDATE clause (for row locking)
func (q *QueryBuilder[T]) ForUpdate() *QueryBuilder[T] {
	q.forUpdate = true
	return q
}

// Timeout sets a timeout for the query
func (q *QueryBuilder[T]) Timeout(duration time.Duration) *QueryBuilder[T] {
	q.timeout = duration
	return q
}

// JoinBuilder methods

// On adds a JOIN condition
func (j *JoinBuilder[T]) On(left, operator, right string) *JoinBuilder[T] {
	j.clause.Conditions = append(j.clause.Conditions, &JoinCondition{
		Left:     left,
		Operator: operator,
		Right:    right,
		IsValue:  false,
	})
	return j
}

// OnValue adds a JOIN condition with a value instead of a column
func (j *JoinBuilder[T]) OnValue(left, operator string, value any) *JoinBuilder[T] {
	j.clause.Conditions = append(j.clause.Conditions, &JoinCondition{
		Left:     left,
		Operator: operator,
		Right:    fmt.Sprintf("%v", value),
		IsValue:  true,
	})
	return j
}

// And is an alias for On to make chaining more readable
func (j *JoinBuilder[T]) And(left, operator, right string) *JoinBuilder[T] {
	return j.On(left, operator, right)
}

// End completes the join builder and returns to the query builder
func (j *JoinBuilder[T]) End() *QueryBuilder[T] {
	j.parent.joins = append(j.parent.joins, j.clause)
	return j.parent
}

// WhereGroupBuilder methods

// Where adds a condition to the group
func (w *WhereGroupBuilder[T]) Where(column string, value any) *WhereGroupBuilder[T] {
	w.group.Conditions = append(w.group.Conditions, &WhereClause{
		Column:   column,
		Operator: "=",
		Value:    value,
	})
	return w
}

// WhereOp adds a condition with an operator to the group
func (w *WhereGroupBuilder[T]) WhereOp(column, operator string, value any) *WhereGroupBuilder[T] {
	w.group.Conditions = append(w.group.Conditions, &WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return w
}

// WhereRaw adds a raw condition to the group
func (w *WhereGroupBuilder[T]) WhereRaw(sql string, args ...any) *WhereGroupBuilder[T] {
	w.group.Conditions = append(w.group.Conditions, &WhereClause{
		IsRaw:   true,
		RawSQL:  sql,
		RawArgs: args,
	})
	return w
}

// End completes the group builder and returns to the query builder
func (w *WhereGroupBuilder[T]) End() *QueryBuilder[T] {
	w.parent.whereGroups = append(w.parent.whereGroups, w.group)
	return w.parent
}

// Helper function to build JOIN SQL
func (j *JoinClause) toSQL() string {
	var sb strings.Builder

	sb.WriteString(j.Type.String())
	sb.WriteString(" ")
	sb.WriteString(j.Table)

	if j.Alias != "" {
		sb.WriteString(" ")
		sb.WriteString(j.Alias)
	}

	if len(j.Conditions) > 0 {
		sb.WriteString(" ON ")
		for i, cond := range j.Conditions {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(cond.Left)
			sb.WriteString(" ")
			sb.WriteString(cond.Operator)
			sb.WriteString(" ")
			if cond.IsValue {
				sb.WriteString("?")
			} else {
				sb.WriteString(cond.Right)
			}
		}
	}

	return sb.String()
}
