package tournament

import (
	"fmt"
	"testing"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestGroupBy(t *testing.T) {
	assert := assert.New(t)
	gs, err := game.ReadGamesDir("../../data/2021")
	assert.NoError(err)
	grs := GroupByTournament(gs)
	for _, gr := range grs {
		fmt.Printf("%s - %d\n", gr.Name, len(gr.Games))
	}
	// assert.FailNow("")
}
