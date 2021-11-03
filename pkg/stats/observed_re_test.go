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
	files, err := filepath.Glob("../../data/2021*.yaml")
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
		assert.Greater(re.GetExpectedRuns(outs, false, false, false), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, true, false, false), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, false, true, false), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, true, true, false), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, false, false, true), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, true, false, true), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, false, true, true), 0.0)
		assert.Greater(re.GetExpectedRuns(outs, true, true, true), 0.0)
	}
}
