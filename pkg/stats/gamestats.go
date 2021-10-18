package stats

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/slshen/sb/pkg/game"
)

type GameStats struct {
	TeamStats map[string]*Stats
	RE        *RunExpectancy
	Filter
}

func NewGameStats(re *RunExpectancy) *GameStats {
	return &GameStats{
		TeamStats: make(map[string]*Stats),
		RE:        re,
	}
}

func (gs *GameStats) Read(g *game.Game) error {
	states := g.GetStates()
	for i, state := range states {
		if gs.filterOut(g, state) {
			continue
		}
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
		battingTeamStats.RecordBatting(g, state, lastState, gs.RE)
		fieldingTeamStats.RecordPitching(g, state, lastState)
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

func (gs *GameStats) GetStats(team *game.Team) *Stats {
	stats := gs.TeamStats[team.Name]
	if stats == nil {
		stats = NewStats(team)
		gs.TeamStats[team.Name] = stats
	}
	return stats
}

func (gs *GameStats) GetPitchingData() *Data {
	dm := newDataMaker("PIT")
	for team, stats := range gs.TeamStats {
		var players []game.PlayerID
		for player := range stats.Pitching {
			players = append(players, player)
		}
		for _, player := range sortPlayers(players) {
			pitching := stats.Pitching[player]
			var m map[string]interface{}
			if err := mapstructure.Decode(pitching, &m); err != nil {
				panic(err)
			}
			gs.adjustRowValues(len(pitching.Games), team, pitching.Player, m)
			dm.addRow(m)
		}
	}
	return dm.data
}

func sortPlayers(players []game.PlayerID) []game.PlayerID {
	sort.Slice(players, func(i, j int) bool {
		return strings.Compare(string(players[i]), string(players[j])) < 0
	})
	return players
}

func (gs *GameStats) GetBattingData() *Data {
	dm := newDataMaker("BAT")
	for team, stats := range gs.TeamStats {
		var players []game.PlayerID
		for player := range stats.Batting {
			players = append(players, player)
		}
		for _, player := range sortPlayers(players) {
			batting := stats.Batting[player]
			var m map[string]interface{}
			if err := mapstructure.Decode(batting, &m); err != nil {
				panic(err)
			}
			gs.adjustRowValues(len(batting.Games), team, batting.Player, m)
			dm.addRow(m)
		}
	}
	return dm.data
}

func (gs *GameStats) adjustRowValues(gameCount int, team string, player *game.Player, m map[string]interface{}) {
	delete(m, "Games")
	m["Games"] = gameCount
	if gs.Team != "" {
		m["Name"] = player.NameOrNumber()
	} else {
		m["Name"] = fmt.Sprintf("%s/%s", team, player.NameOrNumber())
	}
	delete(m, "PlayerID")
	delete(m, "Number")
}
