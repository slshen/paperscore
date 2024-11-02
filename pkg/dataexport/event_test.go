package dataexport

import (
	"testing"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestGetBaseOutCode(t *testing.T) {
	assert := assert.New(t)
	for _, k := range []struct {
		outs       int
		r1, r2, r3 string
		code       string
	}{
		{0, "r1", "r2", "r3", "0321"},
		{0, "", "", "", "0xxx"},
		{1, "", "", "r3", "13xx"},
		{2, "", "r2", "", "2x2x"},
		{3, "", "r2", "", "3xxx"},
	} {
		code := getBaseOutCode(&game.State{
			Outs: k.outs,
			Runners: [3]game.PlayerID{game.PlayerID(k.r1),
				game.PlayerID(k.r2), game.PlayerID(k.r3)},
		})
		assert.Equal(k.code, code, k)
	}
}
