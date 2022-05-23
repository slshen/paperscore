package stats

import (
	"path/filepath"
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestRE24(t *testing.T) {
	assert := assert.New(t)
	re := &ObservedRunExpectancy{}
	files, err := filepath.Glob("../../data/2021/2021*.yaml")
	if !assert.NoError(err) {
		return
	}
	for _, f := range files {
		g, err := game.ReadGameFile(f)
		assert.Nil(err)
		assert.NoError(re.Read(g))
	}
	assert.NotNil(GetRunExpectancyData(re))
	for outs := 0; outs < 3; outs++ {
		for _, runrs := range OccupedBasesValues {
			assert.Greater(re.GetExpectedRuns(outs, runrs), 0.0)
		}
	}
}
