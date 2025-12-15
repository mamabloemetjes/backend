package lib

import (
	"errors"

	"github.com/go-pg/pg/v10"
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

func MapPgError(err error) error {
	var pgErr pg.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Field('C') { // SQLSTATE
		case "23505": // unique_violation
			return ErrConflict
		case "P0002": // no_data_found
			return ErrNotFound
		}
	}
	return err
}
