package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/uptrace/bun/driver/pgdriver"
)

// Database errors
var (
	ErrConflict            = errors.New("conflict")
	ErrNotFound            = errors.New("not found")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrDatabaseConnection  = errors.New("database connection error")
)

// Auth errors
var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// DatabaseError represents a detailed database error with context
type DatabaseError struct {
	Type          string // "unique_violation", "foreign_key_violation", etc.
	Message       string // User-friendly message
	Detail        string // Technical detail (for logging)
	Constraint    string // Constraint name that was violated
	Table         string // Table name
	Column        string // Column name (if available)
	OriginalError error  // Original pgx error
}

func (e *DatabaseError) Error() string {
	return e.Message
}

// UniqueViolationError represents a unique constraint violation
type UniqueViolationError struct {
	DatabaseError
	Field string // The field that caused the conflict (email, username, etc.)
	Value string // The conflicting value (sanitized for logging)
}

func (e *UniqueViolationError) Error() string {
	return e.Message
}

// MapPgError maps pgx/PostgreSQL errors to custom application errors with detailed context
func MapPgError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's a pgdriver.Error (used by Bun)
	var pgDriverErr pgdriver.Error
	if !errors.As(err, &pgDriverErr) {
		// Not a PostgreSQL error, return as-is
		return err
	}

	// Extract fields from pgdriver.Error
	code := pgDriverErr.Field('C')
	message := pgDriverErr.Field('M')
	detail := pgDriverErr.Field('D')
	constraintName := pgDriverErr.Field('n')
	tableName := pgDriverErr.Field('t')
	columnName := pgDriverErr.Field('c')

	switch code {
	case "23505": // unique_violation
		return handleUniqueViolation(detail, constraintName, tableName, columnName, err)
	case "23503": // foreign_key_violation
		return handleForeignKeyViolation(detail, constraintName, tableName, err)
	case "23502": // not_null_violation
		return handleNotNullViolation(message, tableName, columnName, err)
	case "23514": // check_violation
		return handleCheckViolation(message, constraintName, tableName, err)
	case "P0002": // no_data_found (PostgreSQL procedure)
		return ErrNotFound
	case "02000": // no_data (SQL standard)
		return ErrNotFound
	case "08000", "08003", "08006": // connection errors
		return &DatabaseError{
			Type:          "connection_error",
			Message:       "Database connection error. Please try again.",
			Detail:        message,
			OriginalError: err,
		}
	case "40001": // serialization_failure
		return &DatabaseError{
			Type:          "serialization_failure",
			Message:       "Database operation conflict. Please try again.",
			Detail:        message,
			OriginalError: err,
		}
	case "53300": // too_many_connections
		return &DatabaseError{
			Type:          "too_many_connections",
			Message:       "Service is currently busy. Please try again in a moment.",
			Detail:        message,
			OriginalError: err,
		}
	default:
		// Unknown PostgreSQL error
		return &DatabaseError{
			Type:          "unknown",
			Message:       "A database error occurred. Please try again.",
			Detail:        fmt.Sprintf("Code: %s, Message: %s", code, message),
			OriginalError: err,
		}
	}
}

// handleUniqueViolation processes unique constraint violations and provides detailed context
func handleUniqueViolation(detail, constraintName, tableName, columnName string, originalErr error) error {
	// Extract field name from constraint or detail
	field := extractFieldFromConstraint(constraintName, detail)
	if field == "value" && columnName != "" {
		field = columnName
	}

	// Create user-friendly message based on the field
	userMessage := createUniqueViolationMessage(field, tableName)

	return &UniqueViolationError{
		DatabaseError: DatabaseError{
			Type:          "unique_violation",
			Message:       userMessage,
			Detail:        detail,
			Constraint:    constraintName,
			Table:         tableName,
			Column:        field,
			OriginalError: originalErr,
		},
		Field: field,
	}
}

// handleForeignKeyViolation processes foreign key constraint violations
func handleForeignKeyViolation(detail, constraintName, tableName string, originalErr error) error {
	// Determine if it's a deletion or insertion violation
	var userMessage string
	if strings.Contains(detail, "still referenced") {
		userMessage = "Cannot delete this item because it is being used elsewhere."
	} else if strings.Contains(detail, "not present") {
		userMessage = "The referenced item does not exist."
	} else {
		userMessage = "Database relationship constraint violated."
	}

	return &DatabaseError{
		Type:          "foreign_key_violation",
		Message:       userMessage,
		Detail:        detail,
		Constraint:    constraintName,
		Table:         tableName,
		OriginalError: originalErr,
	}
}

// handleNotNullViolation processes NOT NULL constraint violations
func handleNotNullViolation(message, tableName, columnName string, originalErr error) error {
	userMessage := fmt.Sprintf("Required field '%s' cannot be empty.", humanizeColumnName(columnName))

	return &DatabaseError{
		Type:          "not_null_violation",
		Message:       userMessage,
		Detail:        message,
		Table:         tableName,
		Column:        columnName,
		OriginalError: originalErr,
	}
}

// handleCheckViolation processes CHECK constraint violations
func handleCheckViolation(message, constraintName, tableName string, originalErr error) error {
	// Try to extract meaningful message from constraint name
	userMessage := createCheckViolationMessage(constraintName)

	return &DatabaseError{
		Type:          "check_violation",
		Message:       userMessage,
		Detail:        message,
		Constraint:    constraintName,
		Table:         tableName,
		OriginalError: originalErr,
	}
}

// extractFieldFromConstraint attempts to extract the field name from constraint or detail
func extractFieldFromConstraint(constraint, detail string) string {
	// Common patterns in constraint names: idx_users_email, users_email_key, etc.
	parts := strings.Split(constraint, "_")

	// Try to find common field names
	commonFields := []string{"email", "username", "name", "sku", "token", "phone", "url"}
	for _, field := range commonFields {
		for _, part := range parts {
			if strings.EqualFold(part, field) {
				return field
			}
		}
	}

	// Try to extract from detail message: "Key (email)=(user@example.com) already exists."
	if detail != "" && strings.Contains(detail, "Key (") {
		start := strings.Index(detail, "Key (") + 5
		end := strings.Index(detail[start:], ")")
		if end > 0 {
			return detail[start : start+end]
		}
	}

	// If constraint has more than 2 parts, assume last part is the field
	if len(parts) > 2 {
		return parts[len(parts)-1]
	}

	return "value"
}

// createUniqueViolationMessage creates a user-friendly message for unique violations
func createUniqueViolationMessage(field, table string) string {
	switch strings.ToLower(field) {
	case "email":
		return "An account with this email address already exists."
	case "username":
		return "This username is already taken. Please choose a different one."
	case "sku":
		return "A product with this SKU already exists."
	case "token":
		return "This token has already been used."
	case "phone":
		return "This phone number is already registered."
	case "url":
		return "This URL is already in use."
	case "name":
		if strings.Contains(strings.ToLower(table), "product") {
			return "A product with this name already exists."
		}
		return "An item with this name already exists."
	default:
		return fmt.Sprintf("This %s already exists. Please use a different value.", humanizeColumnName(field))
	}
}

// createCheckViolationMessage creates a user-friendly message for check violations
func createCheckViolationMessage(constraint string) string {
	lower := strings.ToLower(constraint)

	// Common check constraint patterns
	if strings.Contains(lower, "email") && strings.Contains(lower, "format") {
		return "Please provide a valid email address."
	}
	if strings.Contains(lower, "username") && strings.Contains(lower, "length") {
		return "Username must be between 3 and 50 characters."
	}
	if strings.Contains(lower, "price") || strings.Contains(lower, "amount") {
		return "Price or amount must be a positive value."
	}
	if strings.Contains(lower, "stock") {
		return "Stock quantity must be zero or greater."
	}
	if strings.Contains(lower, "expires") {
		return "Expiration date must be in the future."
	}

	return "The provided value does not meet the required constraints."
}

// humanizeColumnName converts a column name to a human-readable format
func humanizeColumnName(column string) string {
	// Remove common prefixes
	column = strings.TrimPrefix(column, "is_")
	column = strings.TrimPrefix(column, "has_")

	// Replace underscores with spaces
	column = strings.ReplaceAll(column, "_", " ")

	// Capitalize first letter
	if len(column) > 0 {
		column = strings.ToUpper(column[:1]) + column[1:]
	}

	return column
}

// IsUniqueViolation checks if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	var uniqueErr *UniqueViolationError
	return errors.As(err, &uniqueErr)
}

// IsForeignKeyViolation checks if the error is a foreign key violation
func IsForeignKeyViolation(err error) bool {
	var dbErr *DatabaseError
	return errors.As(err, &dbErr) && dbErr.Type == "foreign_key_violation"
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// GetUserMessage extracts a user-friendly message from any error
func GetUserMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for our custom database errors
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr.Message
	}

	var uniqueErr *UniqueViolationError
	if errors.As(err, &uniqueErr) {
		return uniqueErr.Message
	}

	// Check for known errors
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return "Invalid email or password."
	case errors.Is(err, ErrInvalidToken):
		return "Invalid or expired token."
	case errors.Is(err, ErrExpiredToken):
		return "This link has expired. Please request a new one."
	case errors.Is(err, ErrNotFound):
		return "The requested item was not found."
	case errors.Is(err, ErrConflict):
		return "This item already exists."
	default:
		// Generic error message
		return "An error occurred. Please try again."
	}
}

// GetDetailForLogging extracts detailed information for logging
func GetDetailForLogging(err error) string {
	if err == nil {
		return ""
	}

	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return fmt.Sprintf(
			"Type: %s, Table: %s, Column: %s, Constraint: %s, Detail: %s",
			dbErr.Type,
			dbErr.Table,
			dbErr.Column,
			dbErr.Constraint,
			dbErr.Detail,
		)
	}

	return err.Error()
}
