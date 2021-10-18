package stats

import (
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestGameStats(t *testing.T) {
	assert := assert.New(t)
	gs := NewGameStats()
	g, err := game.ReadGameFile("../../data/20210911-1.yaml")
	if !assert.NoError(err) {
		return
	}
	assert.NoError(gs.Read(g))
	assert.NotNil(gs.GetBattingData())
	assert.NotNil(gs.GetPitchingData())
}
