package playbyplay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestPlayByPlay(t *testing.T) {
	assert := assert.New(t)
	for _, f := range []string{"20211009-1.yaml" /*"20210926-1.yaml", "20210925-3.yaml"*/} {
		g, err := game.ReadGameFile(filepath.Join("../../data/2021", f))
		if !assert.NoError(err) {
			return
		}
		pbp := &Generator{
			Game: g,
		}
		err = pbp.Generate(os.Stdout)
		assert.NoError(err)
	}
}
