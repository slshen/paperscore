package report

import (
	"fmt"
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	assert := assert.New(t)
	re, err := stats.ReadREMatrix("../../data/tweaked_re.csv")
	assert.NoError(err)
	gs, err := game.ReadGamesDir("../../data")
	assert.NoError(err)
	grs := GroupByTournament(gs)
	for _, gr := range grs {
		r := &Report{
			Us:    "pride",
			Group: gr,
		}
		err := r.Run(re)
		assert.NoError(err)
		fmt.Println(r.GetBattingData())
	}
	// t.Fail()
}
