package stats

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestREAnalysis(t *testing.T) {
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
	rea := NewREAnalysis(re)
	fmt.Println(rea.Run())
	// t.Fail()
}
