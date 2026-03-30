package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		expectedKeys []string
	}{
		{
			name:         "no template variables",
			command:      "ls -la",
			expectedKeys: nil,
		},
		{
			name:         "single variable",
			command:      "ssh {{.host}}",
			expectedKeys: []string{"host"},
		},
		{
			name:         "multiple variables",
			command:      "ssh {{.user}}@{{.host}} -p {{.port}}",
			expectedKeys: []string{"user", "host", "port"},
		},
		{
			name:         "variable with extra whitespace",
			command:      "ssh {{ .host }}",
			expectedKeys: []string{"host"},
		},
		{
			name:         "duplicate variables",
			command:      "echo {{.name}} {{.name}}",
			expectedKeys: []string{"name", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, keys, err := Parse(tt.command)

			require.NoError(t, err)
			assert.NotNil(t, tmpl)
			assert.Equal(t, tt.expectedKeys, keys)
		})
	}
}

func TestParse_InvalidTemplate(t *testing.T) {
	_, _, err := Parse("ssh {{.host")

	assert.Error(t, err)
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		values   map[string]string
		expected string
	}{
		{
			name:     "no variables",
			command:  "ls -la",
			values:   map[string]string{},
			expected: "ls -la",
		},
		{
			name:     "single variable",
			command:  "ssh {{.host}}",
			values:   map[string]string{"host": "example.com"},
			expected: "ssh example.com",
		},
		{
			name:    "multiple variables",
			command: "ssh {{.user}}@{{.host}} -p {{.port}}",
			values: map[string]string{
				"user": "root",
				"host": "example.com",
				"port": "2222",
			},
			expected: "ssh root@example.com -p 2222",
		},
		{
			name:     "empty value",
			command:  "echo {{.name}}",
			values:   map[string]string{"name": ""},
			expected: "echo ",
		},
		{
			name:     "value with spaces",
			command:  "echo {{.msg}}",
			values:   map[string]string{"msg": "hello world"},
			expected: "echo hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, _, err := Parse(tt.command)
			require.NoError(t, err)

			result, err := Resolve(tmpl, &tt.values)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
