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
	assert.Len(f.TeamEvents, 2)
	assert.Contains([]string{"pride-fall-2021", "norcal-starz"}, f.TeamEvents[0].TeamID)
	for _, team := range f.TeamEvents {
		if team.TeamID == "pride-fall-2021" {
			assert.Greater(len(team.Events), 48)
			assert.Equal("2", team.Events[0].Pitcher)
			play := team.Events[1].Play
			if assert.NotNil(play) {
				assert.Equal(Numbers("17"), play.Batter)
				assert.Equal("BBBB", play.PitchSequence)
				assert.Equal("W", play.Code)
				if assert.Len(play.Advances, 1) {
					assert.Equal("B-1", play.Advances[0])
				}
			}
			play = team.Events[2].Play
			if assert.NotNil(play) {
				assert.Equal("6", play.Batter.String())
				assert.Equal("C", play.PitchSequence)
				assert.Equal("SB2", play.Code)
			}
			play = team.Events[36].Play
			if assert.NotNil(play) {
				assert.Equal("00", play.Batter.String())
				assert.Equal("advance on throw", play.Comment, play)
			}
		}
	}
}
