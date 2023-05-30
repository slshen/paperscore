package expr

import (
	"fmt"
	"testing"

	"github.com/slshen/sb/pkg/markov"
	"github.com/stretchr/testify/assert"
)

func TestParseModel(t *testing.T) {
	t.SkipNow()
	assert := assert.New(t)
	f, err := ParseFile("testdata/simple.mat")
	assert.NoError(err)
	m, err := NewModel(f)
	assert.NoError(err)
	assert.NotNil(m)
	assert.Equal(0.45, m.pEvent["S"])
	assert.Equal(1-0.45, m.pEvent["O"])
}

func TestInning(t *testing.T) {
	t.SkipNow()
	assert := assert.New(t)
	f, err := ParseFile("testdata/simple.mat")
	if !assert.NoError(err) {
		return
	}
	m, diag := NewModel(f)
	assert.NoError(diag.ErrorOrNil())
	sim := markov.Simulation{
		Model: m,
	}
	err = sim.RunInning()
	assert.NoError(err)
	fmt.Println(sim.Runs)
	t.Fail()
}
