package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean input",
			input:    "Toyota Camry",
			expected: "Toyota Camry",
		},
		{
			name:     "input with script tags",
			input:    "<script>alert('xss')</script>Toyota",
			expected: "Toyota",
		},
		{
			name:     "input with NoSQL operators",
			input:    "$ne Toyota",
			expected: "Toyota",
		},
		{
			name:     "input with brackets",
			input:    "{$gt: 1000}",
			expected: ": 1000",
		},
		{
			name:     "input with HTML tags",
			input:    "<div>Toyota</div>",
			expected: "Toyota",
		},
		{
			name:     "input with whitespace",
			input:    "  Toyota  ",
			expected: "Toyota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveScriptTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "script tag lowercase",
			input:    "<script>alert(1)</script>test",
			expected: "test",
		},
		{
			name:     "script tag uppercase",
			input:    "<SCRIPT>alert(1)</SCRIPT>test",
			expected: "test",
		},
		{
			name:     "mixed case script tag",
			input:    "<ScRiPt>alert(1)</ScRiPt>test",
			expected: "test",
		},
		{
			name:     "HTML tags",
			input:    "<div><p>test</p></div>",
			expected: "test",
		},
		{
			name:     "no tags",
			input:    "clean text",
			expected: "clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeScriptTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveNoSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MongoDB $ne operator",
			input:    "$ne test",
			expected: " test",
		},
		{
			name:     "MongoDB $gt operator",
			input:    "$gt 1000",
			expected: " 1000",
		},
		{
			name:     "object notation",
			input:    "{price: {$gt: 1000}}",
			expected: "price: : 1000",
		},
		{
			name:     "array notation",
			input:    "[1, 2, 3]",
			expected: "1, 2, 3",
		},
		{
			name:     "clean input",
			input:    "Toyota",
			expected: "Toyota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeNoSQLInjection(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"page", "limit", "status"}

	tests := []struct {
		name     string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			item:     "page",
			expected: true,
		},
		{
			name:     "item exists middle",
			item:     "limit",
			expected: true,
		},
		{
			name:     "item exists end",
			item:     "status",
			expected: true,
		},
		{
			name:     "item not exists",
			item:     "unknown",
			expected: false,
		},
		{
			name:     "empty string",
			item:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}
