package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEventCode(t *testing.T) {
	assert := assert.New(t)
	p := playCodeParser{}
	p.parsePlayCode("W")
	assert.Equal("W", p.playCode)
	p.parsePlayCode("SB2;SB3")
	assert.Equal("SB2;SB3", p.playCode)
	p.parsePlayCode("CSH(252)")
	assert.True(p.playIs("CS%($$$)"))
}
