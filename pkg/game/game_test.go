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
		"20211030-1.yaml",
		"20210912-3.yaml",
		"20211030-2.yaml",
		"20210911-3.yaml",
		"20211009-1.yaml",
		"20210911-1.yaml",
		"20210925-3.yaml",
		"20210926-1.yaml",
		"20210925-2.yaml",
	} {
		g, err := ReadGameFile(filepath.Join("../../data/2021", f))
		if !assert.NoError(err, f) {
			continue
		}
		if assert.NotNil(g) {
			// assert.Greater(len(g.VisitorPlays), 10)
			states := g.GetStates()
			for _, state := range states {
				d, _ := yaml.Marshal(state)
				fmt.Println(string(d))
			}
		}
	}
}

func TestAlternativeStates(t *testing.T) {
	assert := assert.New(t)
	g, err := ReadGameFile("../gamefile/testdata/test.gm")
	assert.NoError(err)
	assert.NotNil(g)
	states := g.visitorStates
	assert.Len(states, 27)
	pa6 := states[5]
	assert.Equal("E4/G4", pa6.PlayCode)
	alts := g.GetAlternativeStates(pa6)
	assert.Len(alts, 1)
	assert.Equal("43/G4", alts[0].PlayCode)
}
