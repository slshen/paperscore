package gamefile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(Parser)
	f, err := ParseFile("testdata/test.gm")
	if !assert.NoError(err) {
		return
	}
	assert.NotNil(f)
	assert.Equal("pride-2022", f.Properties["visitorid"])
	assert.NotNil(f.VisitorEvents)
	assert.NotNil(f.HomeEvents)
	events := f.VisitorEvents
	assert.Equal("2", events[0].Pitcher)
	if assert.Greater(len(events), 30) {
		play := events[1].Play
		if assert.NotNil(play) {
			assert.Equal(1, play.PlateAppearance.Int())
			assert.NoError(err)
			assert.Equal("7", play.Batter)
			assert.Equal("CSFS", play.PitchSequence)
			assert.Equal("K", play.Code)
		}
		event := events[9]
		assert.NotNil(event)
		if assert.NotNil(event.Alternative) {
			assert.Equal("routine ground ball", event.Alternative.Comment)
		}
		event = events[8]
		if assert.Len(event.Play.Afters, 1) && assert.NotNil(event.Play.Afters[0].Conference) {
			assert.True(*event.Play.Afters[0].Conference)
		}
		event = events[3]
		assert.Equal("9", *event.Play.Afters[0].CourtesyRunner)
		event = events[19]
		if assert.Len(event.Play.Afters, 1) {
			assert.Equal("3", event.Play.Afters[0].Sub.Enter)
			assert.Equal("2", event.Play.Afters[0].Sub.Exit)
		}
	}
}
