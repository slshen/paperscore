package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type GameStats struct {
	TeamStats map[string]*TeamStats
	RE        RunExpectancy

	red   *REData
	teams map[string]*game.Team
}

func NewGameStats(re RunExpectancy) *GameStats {
	return &GameStats{
		TeamStats: make(map[string]*TeamStats),
		RE:        re,
		red:       NewREData(re),
		teams:     make(map[string]*game.Team),
	}
}

func (gs *GameStats) Read(g *game.Game) error {
	gs.teams[g.Home] = g.HomeTeam
	gs.teams[g.Visitor] = g.VisitorTeam
	states := g.GetStates()
	for i, state := range states {
		var battingTeam, fieldingTeam *game.Team
		if state.Top() {
			battingTeam = g.VisitorTeam
			fieldingTeam = g.HomeTeam
		} else {
			battingTeam = g.HomeTeam
			fieldingTeam = g.VisitorTeam
		}
		battingTeamStats := gs.GetStats(battingTeam)
		fieldingTeamStats := gs.GetStats(fieldingTeam)
		lastState := getLastState(states, i)
		var advances game.Advances
		if !state.Complete {
			advances = state.Advances
		}
		reChange := gs.red.Record(g.ID, state, lastState, advances)
		battingTeamStats.RecordBatting(g, state, lastState, reChange)
		fieldingTeamStats.RecordFielding(g, state, lastState)
	}
	return nil
}

func getLastState(states []*game.State, i int) *game.State {
	if i == 0 {
		return nil
	}
	state := states[i]
	lastState := states[i-1]
	if state.InningNumber == lastState.InningNumber && state.Half == lastState.Half {
		return lastState
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

func (gs *GameStats) GetXRAData() *dataframe.Data {
	var dat *dataframe.Data
	for _, stats := range gs.TeamStats {
		if dat == nil {
			dat = stats.GetXRAData()
		} else {
			dat.Append(stats.GetXRAData())
		}
	}
	return dat
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
