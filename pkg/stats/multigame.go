package stats

import (
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/slshen/sb/pkg/game"
)

type MultiGame struct {
	TeamStats map[string]*Stats
	OnlyTeam  string
}

func NewMultiGame() *MultiGame {
	return &MultiGame{
		TeamStats: make(map[string]*Stats),
	}
}

func (mg *MultiGame) Read(g *game.Game) error {
	states, err := g.GetStates()
	if err != nil {
		return err
	}
	for i, state := range states {
		var battingTeam, fieldingTeam *game.Team
		if state.Top() {
			battingTeam = g.VisitorTeam
			fieldingTeam = g.HomeTeam
		} else {
			battingTeam = g.HomeTeam
			fieldingTeam = g.VisitorTeam
		}
		batting := mg.GetStats(battingTeam)
		pitching := mg.GetStats(fieldingTeam)
		batting.RecordBatting(g, state)
		pitching.RecordPitching(g, state, lastState(states, i))
	}
	return nil
}

func lastState(states []*game.State, i int) *game.State {
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

func (mg *MultiGame) GetStats(team *game.Team) *Stats {
	stats := mg.TeamStats[team.Name]
	if stats == nil {
		stats = NewStats(team)
		mg.TeamStats[team.Name] = stats
	}
	return stats
}

type dataMaker struct {
	columns map[string]int
	data    *Data
}

func (dm *dataMaker) addRow(m map[string]interface{}) {
	if dm.columns == nil {
		dm.data = &Data{Width: map[string]int{
			"Name": 20,
			"Team": 20,
		}}
		dm.columns = make(map[string]int)
		for _, k := range sortKeys(m) {
			dm.columns[k] = len(dm.columns)
			dm.data.Columns = append(dm.data.Columns, k)
		}
	}
	row := make([]interface{}, len(dm.columns))
	for k, v := range m {
		row[dm.columns[k]] = v
	}
	dm.data.Rows = append(dm.data.Rows, row)
}

func (mg *MultiGame) GetPitchingData() *Data {
	var dm dataMaker
	for team, stats := range mg.TeamStats {
		if mg.filterExclude(team, stats) {
			continue
		}
		for _, pitching := range stats.Pitching {
			var m map[string]interface{}
			if err := mapstructure.Decode(pitching, &m); err != nil {
				panic(err)
			}
			mg.adjustRowValues(len(pitching.Games), team, pitching.Player, m)
			dm.addRow(m)
		}
	}
	return dm.data
}

func (mg *MultiGame) GetBattingData() *Data {
	var dm dataMaker
	for team, stats := range mg.TeamStats {
		if mg.filterExclude(team, stats) {
			continue
		}
		for _, batting := range stats.Batting {
			var m map[string]interface{}
			if err := mapstructure.Decode(batting, &m); err != nil {
				panic(err)
			}
			mg.adjustRowValues(len(batting.Games), team, batting.Player, m)
			dm.addRow(m)
		}
	}
	return dm.data
}

func (mg *MultiGame) adjustRowValues(gameCount int, team string, player *game.Player, m map[string]interface{}) {
	delete(m, "Games")
	m["Games"] = gameCount
	if mg.OnlyTeam == "" {
		m["Team"] = team
	}
	m["Name"] = player.NameOrNumber()
	delete(m, "PlayerID")
	delete(m, "Number")
}

func (mg *MultiGame) filterExclude(team string, stats *Stats) bool {
	if mg.OnlyTeam != "" &&
		!strings.HasPrefix(strings.ToLower(team), strings.ToLower(mg.OnlyTeam)) {
		return true
	}
	return false
}

func sortKeys(m map[string]interface{}) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
