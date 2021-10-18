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
		"20210912-3.yaml", "20210911-3.yaml", "20211009-1.yaml", "20210911-1.yaml",
		"20210925-3.yaml", "20210926-1.yaml", "20210925-2.yaml",
	} {
		g, err := ReadGameFile(filepath.Join("../../data", f))
		assert.NoError(err, f)
		if assert.NotNil(g) {
			// assert.Greater(len(g.VisitorPlays), 10)
			states := g.GetStates()
			if err != nil {
				for _, state := range states {
					d, _ := yaml.Marshal(state)
					fmt.Println(string(d))
				}
			}
		}
	}
}
