package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type BattingCountSituations struct {
	Us    string
	NotUs string
	sits  []*CountSituation
}

type CountSituation struct {
	Count string
	Batting
}

func NewBattingByCount() *BattingCountSituations {
	sits := make([]*CountSituation, 4*3)
	for balls := 0; balls <= 3; balls++ {
		for strikes := 0; strikes <= 2; strikes++ {
			sits[strikes*4+balls] = &CountSituation{
				Count: fmt.Sprintf("%d-%d", balls, strikes),
			}
		}
	}
	return &BattingCountSituations{
		sits: sits,
	}
}

func (bc *BattingCountSituations) Read(g *game.Game) {
	states := g.GetStates()
	if bc.Us != "" {
		if g.HomeID != "" && strings.Contains(g.HomeID, bc.Us) {
			states = g.GetHomeStates()
		} else if g.VisitorID != "" && strings.Contains(g.VisitorID, bc.Us) {
			states = g.GetVisitorStates()
		}
	}
	if bc.NotUs != "" {
		if g.HomeID == "" || !strings.Contains(g.HomeID, bc.NotUs) {
			states = g.GetHomeStates()
		} else if g.VisitorID == "" || !strings.Contains(g.VisitorID, bc.NotUs) {
			states = g.GetVisitorStates()
		}
	}
	for _, state := range states {
		bc.record(state)
	}
}

func (bc *BattingCountSituations) record(state *game.State) {
	if state.Complete {
		sits := map[string]*CountSituation{}
		for i := 0; i < len(state.Pitches); i++ {
			count, balls, strikes := state.Pitches[0:i].Count()
			if balls < 4 && strikes < 3 && sits[count] == nil {
				sit := bc.sits[strikes*4+balls]
				sits[count] = sit
			}
		}
		for _, sit := range sits {
			sit.Record(state)
		}
	}
}

func (bc *BattingCountSituations) GetData() *dataframe.Data {
	dat := &dataframe.Data{}
	var idx *dataframe.Index
	for _, sit := range bc.sits {
		idx = dat.MustAppendStruct(idx, sit.Batting)
	}
	return dat.Select(
		dataframe.DeriveStrings("Count", func(idx *dataframe.Index, i int) string {
			return bc.sits[i].Count
		}).WithFormat("%5s"),
		dataframe.DeriveFloats("AVG", AVG).WithFormat("%6.3f"),
		dataframe.DeriveFloats("LAVG", LAVG).WithFormat("%6.3f"),
		dataframe.DeriveFloats("OBS", OnBase).WithFormat("%6.3f"),
		dataframe.DeriveFloats("SLUG", Slugging).WithFormat("%6.3f"),
		dataframe.DeriveFloats("OPS", OPS).WithFormat("%6.3f"),
		dataframe.DeriveFloats("K%", func(idx *dataframe.Index, i int) float64 {
			ab := idx.GetInt(i, "AB")
			k := idx.GetInt(i, "StrikeOuts")
			if ab > 0 {
				return float64(k) / float64(ab)
			}
			return 0
		}).WithFormat("%6.3f"),
		dataframe.DeriveFloats("BB%", func(idx *dataframe.Index, i int) float64 {
			ab := idx.GetInt(i, "PA")
			bb := idx.GetInt(i, "Walks")
			if ab > 0 {
				return float64(bb) / float64(ab)
			}
			return 0
		}).WithFormat("%6.3f"),
		dataframe.Col("PA").WithFormat("%4d"),
	)
}
