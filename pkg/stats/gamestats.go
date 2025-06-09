package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type GameStats struct {
	TeamStats map[string]*TeamStats
	RE        RunExpectancy

	red   *REData
	alt   *AltData
	teams map[string]*game.Team
}

func NewGameStats(re RunExpectancy) *GameStats {
	return &GameStats{
		TeamStats: make(map[string]*TeamStats),
		RE:        re,
		red:       NewREData(re),
		alt:       NewAltData(re),
		teams:     make(map[string]*game.Team),
	}
}

func (gs *GameStats) Read(g *game.Game) error {
	gs.teams[g.Home.Name] = g.Home
	gs.teams[g.Visitor.Name] = g.Visitor
	states := g.GetStates()
	for _, state := range states {
		var battingTeam, fieldingTeam *game.Team
		if state.Top() {
			battingTeam = g.Visitor
			fieldingTeam = g.Home
		} else {
			battingTeam = g.Home
			fieldingTeam = g.Visitor
		}
		battingTeamStats := gs.GetStats(battingTeam)
		fieldingTeamStats := gs.GetStats(fieldingTeam)
		reChange := gs.red.Record(g.ID, state)
		if alt := g.GetAlternativeState(state); alt != nil {
			gs.alt.Record(g.ID, alt)
		}
		battingTeamStats.RecordBatting(g, state, reChange)
		fieldingTeamStats.RecordFielding(g, state)
	}
	return nil
}

func (gs *GameStats) GetStats(team *game.Team) *TeamStats {
	stats := gs.TeamStats[team.Name]
	if stats == nil {
		stats = NewStats(team, gs.RE)
		gs.TeamStats[team.Name] = stats
	}
	return stats
}

func (gs *GameStats) GetPitchingData() *dataframe.Data {
	var dat *dataframe.Data
	for _, stats := range gs.TeamStats {
		if dat == nil {
			dat = stats.GetPitchingData()
		} else {
			dat.Append(stats.GetPitchingData())
		}
	}
	idx := dat.GetIndex()
	return dat.RSort(func(r1, r2 int) bool {
		return comparePlayers(idx, r1, r2)
	})
}

func comparePlayers(idx *dataframe.Index, r1, r2 int) bool {
	n1 := fmt.Sprintf("%v/%v", idx.GetValue(r1, "Team"), idx.GetValue(r1, "Name"))
	n2 := fmt.Sprintf("%v/%v", idx.GetValue(r2, "Team"), idx.GetValue(r2, "Name"))
	return strings.Compare(n1, n2) < 0
}

func (gs *GameStats) GetBattingData() *dataframe.Data {
	return gs.getBattingData(false)
}

func (gs *GameStats) GetAllBattingData() *dataframe.Data {
	return gs.getBattingData(true)
}

func (gs *GameStats) GetRE24Data() *dataframe.Data {
	return gs.red.GetData()
}

func (gs *GameStats) GetAltData() *dataframe.Data {
	return gs.alt.GetData()
}

func (gs *GameStats) GetPerPlayerAltData() *dataframe.Data {
	return gs.alt.GetPerPlayerData()
}

func (gs *GameStats) getBattingData(includeInactiveBatters bool) *dataframe.Data {
	var dat *dataframe.Data
	for _, stats := range gs.TeamStats {
		if dat == nil {
			dat = stats.GetBattingData()
		} else {
			dat.Append(stats.GetBattingData())
		}
	}
	idx := dat.GetIndex()
	if !includeInactiveBatters {
		dat = dat.RFilter(func(row int) bool {
			return idx.GetValue(row, "Inactive").(int) == 0
		})
	}
	return dat.RSort(func(r1, r2 int) bool {
		return comparePlayers(idx, r1, r2)
	})
}
