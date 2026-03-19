package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name    string
		error   ValidationError
		wantErr error
	}{
		{
			name:    "valid Snippet",
			error:   ValidationError{Field: "Command", Message: "Missing value"},
			wantErr: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorString := tt.error.Error()
			assert.Equal(t, "Error occurred on field Command with message: Missing value", errorString)
			if tt.wantErr != nil {
				assert.ErrorIs(t, tt.error.Unwrap(), tt.wantErr)
			}
		})
	}
}
