package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlayerID(t *testing.T) {
	assert := assert.New(t)
	team, err := ReadTeamFile("pride", "../../data/2022/pride-2022.yaml")
	assert.NoError(err)
	if !assert.NotNil(team) {
		return
	}
	assert.Equal(PlayerID("mf17"), team.parsePlayerID("17"))
}
