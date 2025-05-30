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
		{"BCCFBX", 2, 2, "2-2"},
		{"TLBB", 2, 2, "2-2"},
		{"MM", 0, 2, "0-2"},
		{"MCL", 0, 3, "0-3"},
	} {
		ps := Pitches(tc.in)
		known, count, balls, strikes := ps.Count()
		assert.True(known)
		assert.Equal(tc.count, count)
		assert.Equal(tc.b, balls)
		assert.Equal(tc.s, strikes)
	}
	assert.Equal('X', Pitches("CX").Last())
}
