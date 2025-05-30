package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type BattingCountSituations struct {
	Us     string
	NotUs  string
	Direct bool

	sits []*CountSituation
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
		if strings.HasPrefix(g.Home.Name, bc.Us) {
			states = g.GetHomeStates()
		} else {
			states = g.GetVisitorStates()
		}
	}
	if bc.NotUs != "" {
		if !strings.HasPrefix(g.Home.Name, bc.NotUs) {
			states = g.GetHomeStates()
		} else {
			states = g.GetVisitorStates()
		}
	}
	for _, state := range states {
		if bc.Direct {
			bc.recordDirect(state)
		} else {
			bc.recordPassingThrough(state)
		}
	}
}

func (bc *BattingCountSituations) recordDirect(state *game.State) {
	if state.Complete {
		known, _, balls, strikes := state.Pitches[0 : len(state.Pitches)-1].Count()
		if known {
			sit := bc.sits[strikes*4+balls]
			sit.Record(state)
		}
	}
}

func (bc *BattingCountSituations) recordPassingThrough(state *game.State) {
	if state.Complete {
		sits := map[string]*CountSituation{}
		for i := 0; i < len(state.Pitches); i++ {
			known, count, balls, strikes := state.Pitches[0:i].Count()
			if !known {
				continue
			}
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
		idx = dat.AppendStruct(idx, sit.Batting)
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
		dataframe.DeriveFloats("K%", KPCT).WithFormat("%6.3f"),
		dataframe.DeriveFloats("BB%", BBPCT).WithFormat("%6.3f"),
		dataframe.DeriveFloats("PGO%", PGO).WithFormat("%6.3f"),
		dataframe.Col("PA").WithFormat("%4d"),
		dataframe.Col("AB").WithFormat("%4d"),
	)
}
