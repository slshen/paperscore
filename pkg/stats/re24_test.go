package stats

import (
	"path/filepath"
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestRE24(t *testing.T) {
	assert := assert.New(t)
	re24 := NewRE24()
	for _, f := range []string{
		"20211009-1.yaml", "20210911-1.yaml", "20210925-3.yaml", "20210926-1.yaml", "20210925-2.yaml",
	} {
		g, err := game.ReadGameFile(filepath.Join("../../data", f))
		assert.Nil(err)
		assert.NoError(re24.Read(g))
	}
	assert.NotNil(re24.GetData())
}
