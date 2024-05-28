package boxscore

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/stats"
	"github.com/slshen/sb/pkg/text"
)

type Lineup struct {
	*stats.TeamStats
}

func (lineup *Lineup) BattingTable() *dataframe.Data {
	dat := lineup.GetBattingData().Select(
		dataframe.Rename("Name", "#").WithFormat("%-14s"),
		dataframe.Col("AB"),
		dataframe.Rename("Hits", "H"),
		dataframe.Col("LOPH"),
		dataframe.Rename("StrikeOuts", "K"),
		dataframe.Rename("Walks", "BB"),
	)
	idx := dat.GetIndex()
	idx.GetColumn("AB").Summary = dataframe.Sum
	idx.GetColumn("H").Summary = dataframe.Sum
	idx.GetColumn("K").Summary = dataframe.Sum
	idx.GetColumn("BB").Summary = dataframe.Sum
	idx.GetColumn("LOPH").Summary = dataframe.Sum
	return dat
}

func (lineup *Lineup) PitchingTable() *dataframe.Data {
	dat := lineup.GetPitchingData().Select(
		dataframe.Rename("Name", "Pitcher"),
		dataframe.Col("IP"),
		dataframe.Rename("BattersFaced", "BF"),
		dataframe.Rename("Hits", "H"),
		dataframe.Rename("Walks", "BB"),
		dataframe.Rename("StrikeOuts", "K"),
		dataframe.Col("HP"),
		dataframe.Col("WP"),
		dataframe.Rename("SwStr", "SWS"),
	)
	return dat
}

func (lineup *Lineup) ErrorsList() string {
	s := &strings.Builder{}
	for _, f := range lineup.FieldingByPosition {
		if f.Errors > 0 {
			fmt.Fprintf(s, " E%d:%d", f.Position, f.Errors)
		}
	}
	return s.String()
}

func (lineup *Lineup) battingCounts(get func(*stats.Batting) int) string {
	var counts []string
	for _, player := range lineup.Batters {
		data := lineup.GetBatting(nil, player)
		n := get(data)
		if n > 0 {
			if n == 1 {
				counts = append(counts, data.Name)
			} else {
				counts = append(counts, fmt.Sprintf("%s(%d)", data.Name, n))
			}
		}
	}
	return text.Wrap(strings.Join(counts, ", "), 30)
}

func (lineup *Lineup) pitchingCounts(get func(*stats.Pitching) int) string {
	var counts []string
	for _, player := range lineup.Pitchers {
		data := lineup.GetPitching(player)
		n := get(data)
		if n > 0 {
			if n == 1 {
				counts = append(counts, data.Name)
			} else {
				counts = append(counts, fmt.Sprintf("%s(%d)", data.Name, n))
			}
		}
	}
	return text.Wrap(strings.Join(counts, ", "), 30)
}

func (lineup *Lineup) battingSum(get func(*stats.Batting) int) int {
	sum := 0
	for _, data := range lineup.Batting {
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
	return lineup.battingSum(func(b *stats.Batting) int { return b.LOB }) + lineup.LOB
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

func (lineup *Lineup) Bunts() string {
	return lineup.battingCounts(func(b *stats.Batting) int { return b.BuntHits + b.BuntSacrifices })
}

func (lineup *Lineup) MissedOrFoulBunts() string {
	return lineup.battingCounts(func(b *stats.Batting) int {
		return b.FoulBunts + b.MissedBunts
	})
}

func (lineup *Lineup) PitcherHPs() string {
	return lineup.pitchingCounts(func(pd *stats.Pitching) int { return pd.HP })
}

func (lineup *Lineup) PitcherWPs() string {
	return lineup.pitchingCounts(func(pd *stats.Pitching) int { return pd.WP })
}
