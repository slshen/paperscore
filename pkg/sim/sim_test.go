package sim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSim(t *testing.T) {
	t.Skip()
	assert := assert.New(t)
	sim, err := NewSimulation("../../data/sim.yaml")
	assert.NoError(err)
	assert.NotNil(sim)
}
