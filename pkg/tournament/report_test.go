package tournament

import (
	"fmt"
	"testing"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/stats"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	assert := assert.New(t)
	re, err := stats.ReadREMatrix("../../data/tweaked_re.csv")
	assert.NoError(err)
	gs, err := game.ReadGamesDir("../../data/2021")
	assert.NoError(err)
	grs := GroupByTournament(gs)
	for _, gr := range grs {
		r, err := NewReport("pride", re, gr)
		assert.NoError(err)
		fmt.Println(r.GetBattingData())
	}
	// t.Fail()
}
