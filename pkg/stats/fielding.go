package stats

import (
	"github.com/slshen/sb/pkg/game"
)

type FieldingStats struct {
	FieldingByPosition []*Fielding
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
	}
}

func (stats *FieldingStats) recordError(e game.FieldingError) {
	f := stats.FieldingByPosition[e.Fielder-1]
	f.Errors++
	stats.Errors++
}
