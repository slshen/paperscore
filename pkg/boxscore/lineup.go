package boxscore

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/slshen/sb/pkg/text"
)

type Lineup struct {
	TeamName         string
	Team             *game.Team
	Order            []game.PlayerID
	Pitchers         []game.PlayerID
	Stats            *stats.TeamStats
	TeamLOB          int
	Errors           int
	ErrorsByPosition []int `yaml:",flow"`
}

func newLineup(teamName string, team *game.Team, re stats.RunExpectancy) *Lineup {
	return &Lineup{
		TeamName: teamName,
		Team:     team,
		Stats:    stats.NewStats(team, re),
	}
}

func (lineup *Lineup) insertBatter(batter game.PlayerID) {
	for _, player := range lineup.Order {
		if player == batter {
			return
		}
	}
	lineup.Order = append(lineup.Order, batter)
	if len(lineup.Team.Players) > 0 && lineup.Team.Players[batter] == nil {
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

func (lineup *Lineup) battingSum(get func(*stats.Batting) int) int {
	sum := 0
	for _, data := range lineup.Stats.Batting {
		sum += get(data)
	}
	return sum
}

func (lineup *Lineup) TotalAB() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.AB })
}
func (lineup *Lineup) TotalHits() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.Hits })
}
func (lineup *Lineup) TotalRunsScored() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.RunsScored })
}
func (lineup *Lineup) TotalStrikeOuts() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.StrikeOuts })
}
func (lineup *Lineup) TotalWalks() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.Walks })
}
func (lineup *Lineup) TotalGroundOuts() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.GroundOuts })
}
func (lineup *Lineup) TotalLineDrives() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.LineDrives })
}

func (lineup *Lineup) BattingTable() string {
	s := &strings.Builder{}
	tab := text.Table{
		Columns: []text.Column{
			{Header: " #", Width: 2},
			{Header: firstWord(lineup.TeamName, 20), Width: 20, Left: true},
			{Header: "AB", Width: 2},
			{Header: " H", Width: 2},
			{Header: " K", Width: 2},
			{Header: "BB", Width: 2},
			{Header: " L", Width: 2},
			{Header: "GO", Width: 2},
		},
	}
	s.WriteString(tab.Header())
	for _, player := range lineup.Order {
		data := lineup.Stats.GetBatting(player)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, data.Player.NameOrQ(), data.AB, data.Hits,
			data.StrikeOuts, data.Walks, data.LineDrives, data.GroundOuts)
	}
	fmt.Fprintf(s, tab.Format(), "", "", "--", "--", "--", "--", "--", "--")
	fmt.Fprintf(s, tab.Format(), "", "",
		lineup.TotalAB(), lineup.TotalHits(), lineup.TotalStrikeOuts(),
		lineup.TotalWalks(), lineup.TotalLineDrives(), lineup.TotalGroundOuts())
	return strings.TrimRight(s.String(), "\n")
}

func (lineup *Lineup) PitchingTable() string {
	s := &strings.Builder{}
	tab := text.Table{
		Columns: []text.Column{
			{Header: " #", Width: 2},
			{Header: "  IP", Width: 4},
			{Header: "BF", Width: 2},
			{Header: " H", Width: 2},
			{Header: "BB", Width: 2},
			{Header: " K", Width: 2},
			{Header: "GO", Width: 2},
			{Header: "FO", Width: 2},
			{Header: "XBH", Width: 3},
			{Header: "SWST", Width: 4},
			{Header: "SB", Width: 2},
		},
	}
	s.WriteString(tab.Header())
	for _, pitcher := range lineup.Pitchers {
		data := lineup.Stats.GetPitching(pitcher)
		ip := fmt.Sprintf("%d.%d", data.Outs/3, data.Outs%3)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, ip, data.BattersFaced,
			data.Hits, data.Walks, data.StrikeOuts, data.GroundOuts, data.FlyOuts,
			data.Doubles+data.Triples+data.HRs,
			data.SwStr(), data.StolenBases)
	}
	return strings.TrimRight(s.String(), "\n")
}

func (lineup *Lineup) ErrorsList() string {
	s := &strings.Builder{}
	for i, count := range lineup.ErrorsByPosition {
		if count > 0 {
			fmt.Fprintf(s, " E%d:%d", i+1, count)
		}
	}
	return s.String()
}

func (lineup *Lineup) battingCounts(get func(*stats.Batting) int) string {
	var counts []string
	for _, player := range lineup.Order {
		data := lineup.Stats.GetBatting(player)
		n := get(data)
		if n > 0 {
			if n == 1 {
				counts = append(counts, data.Player.NameOrNumber())
			} else {
				counts = append(counts, fmt.Sprintf("%s(%d)", data.Player.NameOrNumber(), n))
			}
		}
	}
	return text.Wrap(strings.Join(counts, ", "), 30)
}

func (lineup *Lineup) pitchingCounts(get func(*stats.Pitching) int) string {
	var counts []string
	for _, player := range lineup.Pitchers {
		data := lineup.Stats.GetPitching(player)
		n := get(data)
		if n > 0 {
			if n == 1 {
				counts = append(counts, data.Player.NameOrNumber())
			} else {
				counts = append(counts, fmt.Sprintf("%s(%d)", data.Player.NameOrNumber(), n))
			}
		}
	}
	return text.Wrap(strings.Join(counts, ", "), 30)
}

func (lineup *Lineup) Singles() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.Singles })
}

func (lineup *Lineup) Doubles() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.Doubles })
}

func (lineup *Lineup) Triples() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.Triples })
}

func (lineup *Lineup) HRs() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.HRs })
}

func (lineup *Lineup) SBs() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.StolenBases })
}

func (lineup *Lineup) CaughtStealings() string {
	return lineup.battingCounts(func(b *stats.Batting) int { return b.CaughtStealing })
}

func (lineup *Lineup) TotalLOB() int {
	return lineup.battingSum(func(b *stats.Batting) int { return b.LOB }) + lineup.TeamLOB
}

func (lineup *Lineup) StrikeOuts() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.StrikeOuts })
}

func (lineup *Lineup) StrikeOutsLooking() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.StrikeOutsLooking })
}

func (lineup *Lineup) Walks() string {
	return lineup.battingCounts(func(pd *stats.Batting) int { return pd.Walks })
}

func (lineup *Lineup) PitcherHPs() string {
	return lineup.pitchingCounts(func(pd *stats.Pitching) int { return pd.HP })
}

func (lineup *Lineup) PitcherWPs() string {
	return lineup.pitchingCounts(func(pd *stats.Pitching) int { return pd.WP })
}

func (lineup *Lineup) recordSteal(runner game.PlayerID) {
	data := lineup.Stats.GetBatting(runner)
	data.StolenBases++
}

func (lineup *Lineup) recordError(e *game.FieldingError) {
	for len(lineup.ErrorsByPosition) < e.Fielder {
		lineup.ErrorsByPosition = append(lineup.ErrorsByPosition, 0)
	}
	lineup.ErrorsByPosition[e.Fielder-1]++
	lineup.Errors++
}

func (lineup *Lineup) recordOffense(state, lastState *game.State) {
	data := lineup.Stats.GetBatting(state.Batter)
	lineup.TeamLOB += data.Record(state)
	switch state.Play.Type {
	case game.StolenBase:
		for _, runner := range state.Play.Runners {
			lineup.recordSteal(runner)
		}
	case game.CaughtStealing:
		fallthrough
	case game.PickedOff:
		if !state.NotOutOnPlay {
			runner := state.Play.Runners[0]
			runnerData := lineup.Stats.GetBatting(runner)
			if state.Play.Type == game.PickedOff {
				runnerData.PickedOff++
			} else {
				runnerData.CaughtStealing++
			}
		}
	}
}

func (lineup *Lineup) recordDefense(state *game.State) error {
	switch state.Play.Type {
	case game.ReachedOnError:
		lineup.recordError(state.Play.FieldingError)
	case game.PickedOff:
		fallthrough
	case game.CaughtStealing:
		if state.NotOutOnPlay && state.Play.FieldingError != nil {
			lineup.recordError(state.Play.FieldingError)
		}
	case game.FoulFlyError:
		lineup.recordError(state.Play.FieldingError)
	}
	return nil
}

func (lineup *Lineup) recordPitching(state *game.State, lastState *game.State) {
	data := lineup.Stats.GetPitching(state.Pitcher)
	data.Record(state, lastState)
}
