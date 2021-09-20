package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdvance(t *testing.T) {
	assert := assert.New(t)
	a, err := parseAdvance("B-1")
	assert.NoError(err)
	assert.Equal("B", a.From)
	a, err = parseAdvance("1X2(64)")
	if assert.NoError(err) {
		assert.True(a.Out)
		assert.Equal([]int{6, 4}, a.Fielders)
	}
	a, err = parseAdvance("3-H(E2/TH)")
	if assert.NoError(err) {
		assert.False(a.Out)
		if assert.NotNil(a.FieldingError) {
			assert.Equal(2, a.FieldingError.Fielder)
		}
	}
}
