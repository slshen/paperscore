package stats

import (
	"slices"

	"github.com/slshen/paperscore/pkg/game"
)

type FieldingStats struct {
	FieldingByPosition []*Fielding
	PositionsByPlayer  map[game.PlayerID][]int
	ErrorsByPlayer     map[game.PlayerID]int
	Errors             int
}

type Fielding struct {
	Position int
	Errors   int
}

func newFieldingStats() *FieldingStats {
	fs := make([]*Fielding, 9)
	for i := 0; i < 9; i++ {
		fs[i] = &Fielding{
			Position: i + 1,
		}
	}
	return &FieldingStats{
		FieldingByPosition: fs,
		PositionsByPlayer:  map[game.PlayerID][]int{},
		ErrorsByPlayer:     map[game.PlayerID]int{},
	}
}

func (stats *FieldingStats) recordFielder(pos int, player game.PlayerID) {
	positions := stats.PositionsByPlayer[player]
	if !slices.Contains(positions, pos) {
		stats.PositionsByPlayer[player] = append(positions, pos)
	}
}

func (stats *FieldingStats) recordError(state *game.State, e game.FieldingError) {
	f := stats.FieldingByPosition[e.Fielder-1]
	player := state.Defense[e.Fielder-1]
	if player != "" {
		stats.ErrorsByPlayer[player]++
	}
	f.Errors++
	stats.Errors++
}
