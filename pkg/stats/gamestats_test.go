package stats

import (
	"path/filepath"
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestGameStats(t *testing.T) {
	assert := assert.New(t)
	gs := NewGameStats(nil)
	files, err := filepath.Glob("../../data/2021/2021*.yaml")
	if !assert.NoError(err) {
		return
	}
	games, err := game.ReadGameFiles(files)
	if !assert.NoError(err) {
		return
	}
	for _, g := range games {
		assert.NoError(gs.Read(g))
	}
	assert.NotNil(gs.GetBattingData())
	assert.NotNil(gs.GetPitchingData())
	assert.NotNil(gs.GetXRAData())
}
