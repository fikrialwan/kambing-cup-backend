package helper_test

import (
	"kambing-cup-backend/helper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Kambing Cup", "kambing-cup"},
		{"MULTIPLE   SPACES", "multiple---spaces"},
		{"LowerCase", "lowercase"},
		{"123 Testing", "123-testing"},
		{"", ""},
	}

	for _, test := range tests {
		result := helper.FormatSlug(test.input)
		assert.Equal(t, test.expected, result)
	}
}
