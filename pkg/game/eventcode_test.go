package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEventCode(t *testing.T) {
	assert := assert.New(t)
	p := eventCodeParser{}
	p.parseEvent("W.B-1")
	assert.Equal("W", p.playCode)
	assert.Equal("B-1", p.advancesCode)
	p.parseEvent("SB2;SB3")
	assert.Equal("SB2;SB3", p.playCode)
}
