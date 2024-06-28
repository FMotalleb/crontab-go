package parser

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestExportEnv_SingleMatch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "Single match",
			input:    "MY_VAR=value",
			expected: map[string]string{"MY_VAR": "value"},
			wantErr:  false,
		},
		{
			name:     "No match",
			input:    "no_match",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "Special characters",
			input:    "VAR_WITH_UNDERSCORE=value_with_underscore",
			expected: map[string]string{"VAR_WITH_UNDERSCORE": "value_with_underscore"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := cronLine{string: tt.input}
			got, err := cl.exportEnv()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
