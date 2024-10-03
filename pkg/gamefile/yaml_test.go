package gamefile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseYAML(t *testing.T) {
	assert := assert.New(t)
	f, err := ParseYAMLFile("../../data/2021/20210912-3.yaml")
	assert.NoError(err)
	assert.NotNil(f)
	assert.Equal("9/12/21", f.Properties["date"])
	assert.Equal("pride-fall-2021", f.Properties["visitorid"])
	events := f.VisitorEvents
	assert.Greater(len(events), 48)
	assert.Equal("2", events[0].Pitcher)
	play := events[1].Play
	if assert.NotNil(play) {
		assert.Equal("17", play.Batter)
		assert.Equal("BBBB", play.PitchSequence)
		assert.Equal("W", play.Code)
		if assert.Len(play.Advances, 1) {
			assert.Equal("B-1", play.Advances[0])
		}
	}
	play = events[2].Play
	if assert.NotNil(play) {
		assert.Equal("6", play.Batter)
		assert.Equal("C", play.PitchSequence)
		assert.Equal("SB2", play.Code)
	}
	play = events[36].Play
	if assert.NotNil(play) {
		assert.Equal("00", play.Batter)
		assert.Equal("advance on throw", events[36].Comment, play)
	}
}
