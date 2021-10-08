package boxscore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestBox(t *testing.T) {
	assert := assert.New(t)
	for _, f := range []string{"20210926-1.yaml", "20210925-3.yaml"} {
		g, err := game.ReadGameFile(filepath.Join("../../data", f))
		if !assert.Nil(err) {
			return
		}
		box, err := NewBoxScore(g)
		assert.NoError(err)
		assert.NotNil(box)
		assert.NoError(box.Render(os.Stdout))
	}
}