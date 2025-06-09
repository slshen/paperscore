package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameShorten(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"Go Programming Language", "GoP Language"},
		{"Hello World", "H World"},
		{"Single", "Single"},
		{"", ""},
		{"Angie S", "Angie S"},
		{"multiple   spaces   here", "ms here"},
		{"  leading and trailing  ", "la trailing"},
		{"OneWord", "OneWord"},
		{"Kh Simmons", "Kh Simmons"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.out, NameShorten(tc.in), "input: %q", tc.in)
	}
}

func TestInitialize(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"Oberlin College", "OC"},
		{"NC Wesleyan", "NCW"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.out, Initialize(tc.in), "input: %q", tc.in)
	}
}
