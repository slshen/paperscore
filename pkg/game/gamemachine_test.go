package game

import (
	"testing"

	"github.com/slshen/paperscore/pkg/gamefile"
	"github.com/stretchr/testify/assert"
)

func TestCE2(t *testing.T) {
	assert := assert.New(t)
	var (
		gm    gameMachine
		state State
		play  = &gamefile.ActualPlay{
			PitchSequence: "N",
			Code:          "C/E2",
		}
	)
	assert.NoError(gm.handlePlayCode(play, &state))
	assert.Equal(state.Type, CatcherInterference)
}
