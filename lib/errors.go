package lib

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Database errors
var (
	ErrConflict = errors.New("conflict")
	ErrNotFound = errors.New("not found")
)

// Auth errors
var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// MapPgError maps pgx/PostgreSQL errors to custom application errors
func MapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrConflict
		case "23503": // foreign_key_violation
			return ErrConflict
		case "P0002": // no_data_found
			return ErrNotFound
		case "02000": // no_data
			return ErrNotFound
		}
	}

	return err
}
