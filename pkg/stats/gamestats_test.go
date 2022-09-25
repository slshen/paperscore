package stats

import (
	"fmt"
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
}

func TestAlt(t *testing.T) {
	assert := assert.New(t)
	re, err := ReadREMatrix("../../data/tweaked_re.csv")
	assert.NoError(err)
	gs := NewGameStats(re)
	games, err := game.ReadGameFiles([]string{"../gamefile/testdata/test.gm"})
	if !assert.NoError(err) {
		return
	}
	assert.Len(games, 1)
	assert.NoError(gs.Read(games[0]))
	fmt.Println(gs.GetAltData())
	// t.Fail()
}
