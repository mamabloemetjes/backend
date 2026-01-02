package lib

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// FieldError represents a clean validation error for APIs
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError is a structured validation error
type ValidationError struct {
	Errors []FieldError `json:"errors"`
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// ExtractAndValidateBody extracts and validates the request body into the provided struct type T
func ExtractAndValidateBody[T any](r *http.Request) (*T, error) {
	defer r.Body.Close()

	var body T

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&body); err != nil {
		return nil, err
	}

	if err := validate.Struct(body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return nil, mapValidationErrors(ve)
		}
		return nil, err
	}

	return &body, nil
}

func mapValidationErrors(errs validator.ValidationErrors) *ValidationError {
	out := &ValidationError{}

	for _, e := range errs {
		field := strings.ToLower(e.Field())

		var message string
		switch e.Tag() {
		case "required":
			message = "is required"
		case "email":
			message = "must be a valid email address"
		case "url":
			message = "must be a valid URL"
		case "uuid4":
			message = "must be a valid UUID"
		case "min":
			message = "must be at least " + e.Param() + " characters"
		case "max":
			message = "must be at most " + e.Param() + " characters"
		case "len":
			message = "must be exactly " + e.Param() + " characters"
		case "gte":
			message = "must be greater than or equal to " + e.Param()
		case "lte":
			message = "must be less than or equal to " + e.Param()
		case "oneof":
			message = "must be one of: " + e.Param()
		case "dive":
			// dive is a nested validation tag, skip it as the actual error will be reported by the nested field
			continue
		default:
			message = "is invalid"
		}

		out.Errors = append(out.Errors, FieldError{
			Field:   field,
			Message: message,
		})
	}

	return out
}
