package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrDuplicate  = errors.New("duplicate entry")
	ErrValidation = errors.New("validation failed")
)

type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("Error occurred on field %s with message: %s", v.Field, v.Message)
}

func (v *ValidationError) Unwrap() error {

	return ErrValidation
}
