package boxscore

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/table"
)

type Lineup struct {
	TeamName   string
	Team       *game.Team
	Order      []game.PlayerID
	Pitchers   []game.PlayerID
	PlayerData map[game.PlayerID]*PlayerData
	Total      struct {
		AB, Hits, RunsScored, LOB, Errors, StrikeOuts int
		E                                             []int `yaml:",flow"`
	}
}

type PlayerData struct {
	Player                         *game.Player `yaml:"-"`
	AB, Runs, Hits /*RBI,*/, Walks int
	StrikeOuts                     int
	StrikeOutsLooking              int
	RunsScored                     int
	Singles, Doubles, Triples, HRs int
	StolenBases, CaughtStealing    int
	LOB                            int
	PitchesSeen, Swings, Misses    int
	Pitching                       PitchingData
}

type PitchingData struct {
	Pitches, Strikes, Balls                     int
	Swings, Misses                              int
	Hits, Singles, Doubles, Triples, HRs, Walks int
	StrikeOuts, StrikeOutsLooking               int
	Outs                                        int
	WP, HP                                      int
	BattersFaced                                int
}

func newLineup(teamName string, team *game.Team) *Lineup {
	return &Lineup{
		TeamName:   teamName,
		Team:       team,
		PlayerData: make(map[game.PlayerID]*PlayerData),
	}
}

func (lineup *Lineup) getData(player game.PlayerID) *PlayerData {
	data := lineup.PlayerData[player]
	if data == nil {
		data = &PlayerData{
			Player: lineup.Team.GetPlayer(player),
		}
		lineup.PlayerData[player] = data
	}
	return data
}

func (lineup *Lineup) insertBatter(batter game.PlayerID) {
	for _, player := range lineup.Order {
		if player == batter {
			return
		}
	}
	lineup.Order = append(lineup.Order, batter)
	if lineup.Team != nil && lineup.Team.Players[batter] == nil {
		fmt.Printf("batter %s does not have a team entry\n", batter)
	}
}

func (lineup *Lineup) insertPitcher(pitcher game.PlayerID) {
	for _, player := range lineup.Pitchers {
		if player == pitcher {
			return
		}
	}
	lineup.Pitchers = append(lineup.Pitchers, pitcher)
}

func (lineup *Lineup) BattingTable() string {
	s := &strings.Builder{}
	tab := table.Table{
		Columns: []table.Column{
			{Header: " #", Width: 2},
			{Header: firstWord(lineup.TeamName, 20), Width: 20, Left: true},
			{Header: "AB", Width: 2},
			{Header: " R", Width: 2},
			{Header: " H", Width: 2},
			{Header: " K", Width: 2},
		},
	}
	s.WriteString(tab.Header())
	for _, player := range lineup.Order {
		data := lineup.getData(player)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, data.Player.NameOrQ(), data.AB, data.RunsScored,
			data.Hits, data.StrikeOuts)
	}
	fmt.Fprintf(s, tab.Format(), "", "", "--", "--", "--", "--")
	fmt.Fprintf(s, tab.Format(), "", "",
		lineup.Total.AB, lineup.Total.RunsScored, lineup.Total.Hits, lineup.Total.StrikeOuts)
	return s.String()
}

func (lineup *Lineup) PitchingTable() string {
	s := &strings.Builder{}
	tab := table.Table{
		Columns: []table.Column{
			{Header: " #", Width: 2},
			{Header: "  IP", Width: 4},
			{Header: " H", Width: 2},
			{Header: "BB", Width: 2},
			{Header: " K", Width: 2},
			{Header: "BF", Width: 2},
			{Header: " P", Width: 2},
			{Header: " S", Width: 2},
		},
	}
	s.WriteString(tab.Header())
	for _, pitcher := range lineup.Pitchers {
		data := lineup.getData(pitcher)
		ip := fmt.Sprintf("%d.%d", data.Pitching.Outs/3, data.Pitching.Outs%3)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, ip, data.Pitching.Hits, data.Pitching.Walks,
			data.Pitching.StrikeOuts, data.Pitching.BattersFaced,
			data.Pitching.Pitches, data.Pitching.Strikes)
	}
	return s.String()

}

func (lineup *Lineup) ErrorsList() string {
	s := &strings.Builder{}
	for i, count := range lineup.Total.E {
		if count > 0 {
			fmt.Fprintf(s, " E%d:%d", i+1, count)
		}
	}
	return s.String()
}

func (lineup *Lineup) playCounts(get func(*PlayerData) int) string {
	var counts []string
	for _, player := range lineup.Order {
		data := lineup.getData(player)
		n := get(data)
		if n > 0 {
			if n == 1 {
				counts = append(counts, data.Player.NameOrNumber())
			} else {
				counts = append(counts, fmt.Sprintf("%s(%d)", data.Player.NameOrNumber(), n))
			}
		}
	}
	return wordwrap.WrapString(strings.Join(counts, ", "), 30)
}

func (lineup *Lineup) Singles() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.Singles })
}

func (lineup *Lineup) Doubles() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.Doubles })
}

func (lineup *Lineup) Triples() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.Triples })
}

func (lineup *Lineup) HRs() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.HRs })
}

func (lineup *Lineup) SBs() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.StolenBases })
}

func (lineup *Lineup) StrikeOuts() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.StrikeOuts })
}

func (lineup *Lineup) StrikeOutsLooking() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.StrikeOutsLooking })
}

func (lineup *Lineup) Walks() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.Walks })
}

func (lineup *Lineup) PitcherHPs() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.Pitching.HP })
}

func (lineup *Lineup) PitcherWPs() string {
	return lineup.playCounts(func(pd *PlayerData) int { return pd.Pitching.WP })
}

func (lineup *Lineup) recordSteal(runner game.PlayerID) {
	data := lineup.getData(runner)
	data.StolenBases++
}

func (lineup *Lineup) recordError(e *game.FieldingError) {
	for len(lineup.Total.E) < e.Fielder {
		lineup.Total.E = append(lineup.Total.E, 0)
	}
	lineup.Total.E[e.Fielder-1]++
	lineup.Total.Errors++
}
func (lineup *Lineup) recordRunScored(runner game.PlayerID) {
	data := lineup.getData(runner)
	data.RunsScored++
	lineup.Total.RunsScored++
}

func (lineup *Lineup) recordOffensePA(state *game.State) {
	if !state.Complete {
		return
	}
	data := lineup.getData(state.Batter)
	if state.Play.Hit() {
		data.Hits++
		lineup.Total.Hits++
		if state.Play.Single() {
			data.Singles++
		}
		if state.Play.Double() {
			data.Doubles++
		}
		if state.Play.Triple() {
			data.Triples++
		}
		if state.Play.HomeRun() {
			data.HRs++
		}
	}
	if state.Play.StrikeOut() {
		data.StrikeOuts++
		lineup.Total.StrikeOuts++
	}
	if state.Play.Walk() {
		data.Walks++
	}
	if !(state.Play.Walk() || state.Play.HitByPitch() ||
		state.Play.CatcherInterference() ||
		(state.Play.ReachedOnError() != nil && state.Modifiers.Contains(game.Obstruction)) ||
		state.Modifiers.Contains(game.SacrificeFly, game.SacrificeHit)) {
		data.AB++
		lineup.Total.AB++
	}
	data.PitchesSeen = state.Pitches.Balls() + state.Pitches.Strikes()
	data.Swings = state.Pitches.Swings()
	data.Misses = state.Pitches.Misses()
}

func (lineup *Lineup) recordDefensePA(state *game.State) {
	if !state.Complete {
		return
	}
	if e := state.Play.ReachedOnError(); e != nil {
		lineup.recordError(e)
	}
}

func (lineup *Lineup) recordPitching(state *game.State, lastState *game.State) {
	data := lineup.getData(state.Pitcher)
	pdata := &data.Pitching
	pdata.Outs += state.OutsOnPlay
	if lastState != nil && (lastState.Batter != state.Batter || lastState.Pitcher != state.Pitcher) {
		pdata.BattersFaced++
	}
	if state.Play.WildPitch() {
		pdata.WP++
	}
	if state.Complete || state.Incomplete {
		pdata.Pitches += len(state.Pitches)
		pdata.Strikes += state.Pitches.Strikes()
		pdata.Balls += state.Pitches.Balls()
		pdata.Swings += state.Pitches.Swings()
		pdata.Misses += state.Pitches.Misses()
		if state.Play.Walk() {
			pdata.Walks++
		}
		if state.Play.Hit() {
			pdata.Hits++
		}
		if state.Play.HitByPitch() {
			pdata.HP++
		}
	}
}
