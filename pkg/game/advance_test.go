package game

import (
	"testing"

	"github.com/slshen/sb/pkg/gamefile"
	"github.com/stretchr/testify/assert"
)

func TestAdvance(t *testing.T) {
	assert := assert.New(t)
	a, err := parseAdvance(&gamefile.ActualPlay{}, "B-1")
	assert.NoError(err)
	assert.Equal("B", a.From)
	a, err = parseAdvance(&gamefile.ActualPlay{}, "1X2(64)")
	if assert.NoError(err) {
		assert.True(a.Out)
		assert.Equal([]int{6, 4}, a.Fielders)
	}
	a, err = parseAdvance(&gamefile.ActualPlay{}, "3-H(E2/TH)")
	if assert.NoError(err) {
		assert.False(a.Out)
		if assert.NotNil(a.FieldingError) {
			assert.Equal(2, a.FieldingError.Fielder)
		}
	}
	as, err := parseAdvances(&gamefile.ActualPlay{Advances: []string{"B-1", "1-2", "2-3"}},
		PlayerID("b1"), [3]PlayerID{"r1", "r2"})
	assert.NoError(err)
	assert.Equal(3, len(as))
	assert.NotNil(as.From("B"))
	assert.Equal(PlayerID("b1"), as.From("B").Runner)
	assert.Equal(PlayerID("r2"), as.From("2").Runner)
}
