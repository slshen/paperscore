package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrajectory(t *testing.T) {
	assert := assert.New(t)
	for _, tc := range []struct {
		play, code string
		tr         Trajectory
	}{
		{"53/G5/B/SH 1-2 2-3", "53", Bunt},
		{"FC1/BG1 B-1 1-2", "FC1", BuntGrounder},
	} {
		p := playCodeParser{}
		p.parsePlayCode(tc.play)
		assert.Equal(tc.code, p.playCode)
		assert.Equal(tc.tr, p.modifiers.Trajectory())
	}
}
