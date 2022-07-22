package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlayerID(t *testing.T) {
	assert := assert.New(t)
	team, err := GetTeam("../../data/2022", "", "pride-2022")
	assert.NoError(err)
	if !assert.NotNil(team) {
		return
	}
	assert.Equal(PlayerID("mf17"), team.parsePlayerID("17"))
}
