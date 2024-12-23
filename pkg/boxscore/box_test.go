package boxscore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestBox(t *testing.T) {
	assert := assert.New(t)
	for _, f := range []string{ /*"20211016-3.yaml",*/
		"../gamefile/testdata/test.gm",
		"20211119-2.yaml" /*"20211009-1.yaml" "20210926-1.yaml", "20210925-3.yaml"*/} {
		path := f
		if f[0] != '.' {
			path = filepath.Join("../../data/2021", f)
		}
		g, err := game.ReadGameFile(path)
		if !assert.Nil(err) {
			return
		}
		box, err := NewBoxScore(g, nil)
		assert.NoError(err)
		assert.NotNil(box)
		assert.NoError(box.Render(os.Stdout))
	}
}
