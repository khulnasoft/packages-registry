package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type dummy struct {
	field string
}

func TestOrderedMapKeysOf(t *testing.T) {
	tests := []struct {
		name     string
		in       interface{}
		expected []string
	}{
		{
			name:     "with map",
			in:       map[string]int{"foo": 1, "bar": 2},
			expected: []string{"bar", "foo"},
		},
		{
			name:     "with complex map",
			in:       map[string]dummy{"foo": {field: "1"}, "bar": {field: "2"}},
			expected: []string{"bar", "foo"},
		},
		{
			name:     "with no map",
			in:       2,
			expected: []string{},
		},
		{
			name:     "with nil",
			in:       nil,
			expected: []string{},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			result := OrderedMapKeysOf(spec.in)

			require.Equal(t, spec.expected, result)
		})
	}
}
