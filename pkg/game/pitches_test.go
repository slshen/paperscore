package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPitches(t *testing.T) {
	assert := assert.New(t)
	for _, tc := range []struct {
		in    string
		b, s  int
		count string
	}{
		{"", 0, 0, "0-0"},
		{"B", 1, 0, "1-0"},
		{"CSBB", 2, 2, "2-2"},
		{"C.X", 0, 1, "0-1"},
	} {
		ps := Pitches(tc.in)
		assert.Equal(tc.b, ps.Balls())
		assert.Equal(tc.s, ps.Strikes())
		assert.Equal(tc.count, ps.Count())
	}
	assert.Equal("X", Pitches("CX").Last())
}
