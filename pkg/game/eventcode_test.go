package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEventCode(t *testing.T) {
	assert := assert.New(t)
	p := eventCodeParser{}
	assert.Equal(Play("W"), p.parseEvent("W.B-1"))
	assert.Equal("B-1", p.advancesCode)
	advs, err := parseAdvances(p.advancesCode)
	assert.NoError(err)
	assert.NotNil(advs["B"])
	play := p.parseEvent("SB2;SB3")
	assert.Equal(Play("SB2;SB3"), play)
}
