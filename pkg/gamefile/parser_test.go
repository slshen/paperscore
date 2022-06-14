package gamefile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(parser)
	f, err := ParseFile("testdata/test.gm")
	assert.NoError(err)
	assert.NotNil(f)
	assert.Equal("pride-2022", f.Properties["us"])
	if !assert.Len(f.TeamEvents, 2) {
		return
	}
	events := f.TeamEvents[0].Events
	assert.Equal("2", events[0].Pitcher)
	if assert.Greater(len(events), 30) {
		play := events[1].Play
		if assert.NotNil(play) {
			assert.Equal(1, play.PlateAppearance.Int())
			assert.NoError(err)
			assert.Equal("7", play.Batter.String())
			assert.Equal("CSFS", play.PitchSequence)
			assert.Equal("K", play.Code)
		}
	}
	assert.NotNil(f.GetVisitorEvents())
	assert.NotNil(f.GetHomeEvents())
}
