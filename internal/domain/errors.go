package domain

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested resource does not exist.
// ErrDuplicate is returned when a resource already exists.
// ErrValidation is returned when input fails validation checks.
var (
	ErrNotFound   = errors.New("not found")
	ErrDuplicate  = errors.New("duplicate entry")
	ErrValidation = errors.New("validation failed")
)

// ValidationError provides detailed information about a failed validation check.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the formatted validation error message.
func (v *ValidationError) Error() string {
	return fmt.Sprintf("Error occurred on field %s with message: %s", v.Field, v.Message)
}

// Unwrap returns the underlying ErrValidation sentinel for use with errors.Is.
func (v *ValidationError) Unwrap() error {

	return ErrValidation
}
