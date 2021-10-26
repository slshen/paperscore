package sim

import (
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestSim(t *testing.T) {
	assert := assert.New(t)
	team, err := game.ReadTeamFile("", "../../data/pride-fall-2021.yaml")
	assert.NoError(err)
	GenerateGame(team, nil)
}
