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
	Stats            *stats.Stats
	TeamLOB          int
	Errors           int
	ErrorsByPosition []int `yaml:",flow"`
}

func newLineup(teamName string, team *game.Team) *Lineup {
	return &Lineup{
		TeamName: teamName,
		Team:     team,
		Stats:    stats.NewStats(team),
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

func (lineup *Lineup) BattingTable() string {
	s := &strings.Builder{}
	tab := text.Table{
		Columns: []text.Column{
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
		data := lineup.Stats.GetBatting(player)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, data.Player.NameOrQ(), data.AB, data.Hits,
			data.RunsScored, data.StrikeOuts)
	}
	fmt.Fprintf(s, tab.Format(), "", "", "--", "--", "--", "--")
	fmt.Fprintf(s, tab.Format(), "", "",
		lineup.TotalAB(), lineup.TotalHits(), lineup.TotalRunsScored(), lineup.TotalStrikeOuts())
	return strings.TrimRight(s.String(), "\n")
}

func (lineup *Lineup) PitchingTable() string {
	s := &strings.Builder{}
	tab := text.Table{
		Columns: []text.Column{
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
		data := lineup.Stats.GetPitching(pitcher)
		ip := fmt.Sprintf("%d.%d", data.Outs/3, data.Outs%3)
		fmt.Fprintf(s, tab.Format(), data.Player.Number, ip,
			data.Hits, data.Walks, data.StrikeOuts, data.Doubles+data.Triples+data.HRs,
			data.Whiff(), data.SwStr())
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
	if state.Play.StolenBase() {
		for _, base := range state.Play.StolenBases() {
			switch base {
			case "2":
				lineup.recordSteal(lastState.Runners[0])
			case "3":
				lineup.recordSteal(lastState.Runners[1])
			case "H":
				lineup.recordSteal(lastState.Runners[2])
			}
		}
	}
	if state.Play.CaughtStealing() {
		if !state.NotOutOnPlay {
			base := game.PreviousBase[string(state.Play)[2:3]]
			runner := lastState.Runners[game.BaseNumber[base]]
			runnerData := lineup.Stats.GetBatting(runner)
			runnerData.CaughtStealing++
		}
	}
}

func (lineup *Lineup) recordDefense(state *game.State) error {
	if !state.Complete {
		return nil
	}
	if state.Play.ReachedOnError() {
		fe, err := state.Play.FieldingError()
		if err != nil {
			return err
		}
		lineup.recordError(fe)
	}
	return nil
}

func (lineup *Lineup) recordPitching(state *game.State, lastState *game.State) {
	data := lineup.Stats.GetPitching(state.Pitcher)
	data.Record(state, lastState)
}
