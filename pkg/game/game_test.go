package game

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	assert := assert.New(t)
	for _, f := range []string{
		"20210925-3.yaml", //	"20210926-1.yaml", "20210925-2.yaml",
	} {
		g, err := ReadGameFile(filepath.Join("../../data", f))
		assert.Nil(err)
		if assert.NotNil(g) {
			//assert.Greater(len(g.VisitorPlays), 10)
			states, err := g.GetStates()
			assert.Nil(err)
			for _, state := range states {
				d, _ := yaml.Marshal(state)
				fmt.Println(string(d))
			}
		}
	}
	assert.FailNow("")
}
