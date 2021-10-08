package boxscore

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/table"
)

type Lineup struct {
	TeamName     string
	Team         *game.Team
	Order        []game.PlayerID
	Pitchers     []game.PlayerID
	BattingData  map[game.PlayerID]*BattingData
	PitchingData map[game.PlayerID]*PitchingData
	Total        struct {
		AB, Hits, RunsScored, LOB, Errors, StrikeOuts int
		E                                             []int `yaml:",flow"`
	}
}

type BattingData struct {
	Player                         *game.Player `yaml:"-"`
	AB, Runs, Hits /*RBI,*/, Walks int
	StrikeOuts                     int
	StrikeOutsLooking              int
	RunsScored                     int
	Singles, Doubles, Triples, HRs int
	StolenBases, CaughtStealing    int
	LOB                            int
	PitchesSeen, Swings, Misses    int
}

type PitchingData struct {
	Player                             *game.Player `yaml:"-"`
	Pitches, Strikes, Balls            int
	Swings, Misses                     int
	Hits, Doubles, Triples, HRs, Walks int
	StrikeOuts, StrikeOutsLooking      int
	Outs                               int
	WP, HP                             int
	BattersFaced                       int
}

func newLineup(teamName string, team *game.Team) *Lineup {
	return &Lineup{
		TeamName:     teamName,
		Team:         team,
		BattingData:  make(map[game.PlayerID]*BattingData),
		PitchingData: make(map[game.PlayerID]*PitchingData),
	}
}

func (lineup *Lineup) getBatterData(player game.PlayerID) *BattingData {
	data := lineup.BattingData[player]
	if data == nil {
		data = &BattingData{
			Player: lineup.Team.GetPlayer(player),
		}
		lineup.BattingData[player] = data
	}
	return data
}

func (lineup *Lineup) getPitchingData(player game.PlayerID) *PitchingData {
	data := lineup.PitchingData[player]
	if data == nil {
		data = &PitchingData{
			Player: lineup.Team.GetPlayer(player),
		}
		lineup.PitchingData[player] = data
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
			{Header: " H", Width: 2},
			{Header: " R", Width: 2},
			{Header: " K", Width: 2},
		},
	}
	s.WriteString(tab.Header())
	for _, player := range lineup.Order {
		data := lineup.getBatterData(player)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, data.Player.NameOrQ(), data.AB, data.Hits,
			data.RunsScored, data.StrikeOuts)
	}
	fmt.Fprintf(s, tab.Format(), "", "", "--", "--", "--", "--")
	fmt.Fprintf(s, tab.Format(), "", "",
		lineup.Total.AB, lineup.Total.Hits, lineup.Total.RunsScored, lineup.Total.StrikeOuts)
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
			{Header: "XBH", Width: 3},
			{Header: "WHFF", Width: 4},
			{Header: "SWST", Width: 4},
		},
	}
	s.WriteString(tab.Header())
	for _, pitcher := range lineup.Pitchers {
		data := lineup.getPitchingData(pitcher)
		ip := fmt.Sprintf("%d.%d", data.Outs/3, data.Outs%3)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, ip,
			data.Hits, data.Walks, data.StrikeOuts, data.Doubles+data.Triples+data.HRs,
			data.Whiff(), data.SwStr())
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

func (lineup *Lineup) battingCounts(get func(*BattingData) int) string {
	var counts []string
	for _, player := range lineup.Order {
		data := lineup.getBatterData(player)
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

func (lineup *Lineup) pitchingCounts(get func(*PitchingData) int) string {
	var counts []string
	for _, player := range lineup.Pitchers {
		data := lineup.getPitchingData(player)
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
	return lineup.battingCounts(func(pd *BattingData) int { return pd.Singles })
}

func (lineup *Lineup) Doubles() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.Doubles })
}

func (lineup *Lineup) Triples() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.Triples })
}

func (lineup *Lineup) HRs() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.HRs })
}

func (lineup *Lineup) SBs() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.StolenBases })
}

func (lineup *Lineup) StrikeOuts() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.StrikeOuts })
}

func (lineup *Lineup) StrikeOutsLooking() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.StrikeOutsLooking })
}

func (lineup *Lineup) Walks() string {
	return lineup.battingCounts(func(pd *BattingData) int { return pd.Walks })
}

func (lineup *Lineup) PitcherHPs() string {
	return lineup.pitchingCounts(func(pd *PitchingData) int { return pd.HP })
}

func (lineup *Lineup) PitcherWPs() string {
	return lineup.pitchingCounts(func(pd *PitchingData) int { return pd.WP })
}

func (lineup *Lineup) recordSteal(runner game.PlayerID) {
	data := lineup.getBatterData(runner)
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
	data := lineup.getBatterData(runner)
	data.RunsScored++
	lineup.Total.RunsScored++
}

func (lineup *Lineup) recordOffensePA(state *game.State) {
	if !state.Complete {
		return
	}
	data := lineup.getBatterData(state.Batter)
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
	data := lineup.getPitchingData(state.Pitcher)
	data.Outs += state.OutsOnPlay
	if lastState != nil && (lastState.Batter != state.Batter || lastState.Pitcher != state.Pitcher) {
		data.BattersFaced++
	}
	if state.Play.WildPitch() {
		data.WP++
	}
	if state.Complete || state.Incomplete {
		data.Pitches += len(state.Pitches)
		data.Strikes += state.Pitches.StrikesThrown()
		data.Balls += state.Pitches.Balls()
		data.Swings += state.Pitches.Swings()
		data.Misses += state.Pitches.Misses()
		if state.Play.Walk() {
			data.Walks++
		}
		if state.Play.Hit() {
			data.Hits++
		}
		if state.Play.HitByPitch() {
			data.HP++
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
}

func (pd PitchingData) Whiff() string {
	// misses/swings
	if pd.Swings > 0 {
		return fmt.Sprintf("%.03f", float64(pd.Misses)/float64(pd.Swings))[1:]
	}
	return ""
}

func (pd PitchingData) SwStr() string {
	// % pitches swung & miss
	if pd.Pitches > 0 {
		return fmt.Sprintf("%.03f", float64(pd.Misses)/float64(pd.Pitches))[1:]
	}
	return ""
}
