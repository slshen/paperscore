package tournament

import (
	"fmt"
	"strings"
	"time"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Group struct {
	Date  time.Time
	Name  string
	Games []*game.Game
}

type Report struct {
	Us    string
	Group *Group

	gs *stats.GameStats
}

func NewReport(us string, re stats.RunExpectancy, group *Group) (*Report, error) {
	r := &Report{
		Us:    us,
		Group: group,
		gs:    stats.NewGameStats(re),
	}
	r.gs = stats.NewGameStats(re)
	for _, g := range r.Group.Games {
		if err := r.gs.Read(g); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Report) GetBestAndWorstRE24(n int) *dataframe.Data {
	dat := r.gs.GetRE24Data()
	dat = stats.GetBiggestRE24(dat, n)
	dat.Name = r.noteGames(fmt.Sprintf("%s Plays", r.Group.Name))
	return dat
}

func (r *Report) noteGames(s string) string {
	if len(r.Group.Games) > 1 {
		return fmt.Sprintf("%s (%d games)", s, len(r.Group.Games))
	}
	return s
}

func (r *Report) GetBattingData() *dataframe.Data {
	dat := r.gs.GetBattingData()
	idx := dat.GetIndex()
	dat.Name = r.noteGames(r.Group.Name)
	dat = dat.RFilter(func(row int) bool {
		t := strings.ToLower(idx.GetString(row, "Team"))
		return strings.HasPrefix(t, r.Us)
	})
	dat = dat.Select(
		dataframe.Col("Name").WithFormat("%14s"),
		dataframe.Col("PA").WithSummary(dataframe.Sum),
		dataframe.Col("AB").WithSummary(dataframe.Sum),
		dataframe.Rename("Hits", "H").WithSummary(dataframe.Sum),
		dataframe.Col("LOPH").WithSummary(dataframe.Sum),
		dataframe.DeriveInts("XBH", func(idx *dataframe.Index, i int) int {
			return idx.GetInt(i, "Doubles") + idx.GetInt(i, "Triples") + idx.GetInt(i, "HRs")
		}).WithFormat("%3d").WithSummary(dataframe.Sum),
		dataframe.Rename("Walks", "BB").WithSummary(dataframe.Sum),
		dataframe.Rename("StrikeOuts", "K").WithSummary(dataframe.Sum),
		dataframe.Rename("RunsScored", "RS").WithFormat("%2d").WithSummary(dataframe.Sum),
		dataframe.Col("RE24").WithSummary(dataframe.Sum),
		dataframe.DeriveInts("OBP", stats.Thousands(stats.OnBase)).WithPCT(),
		dataframe.DeriveInts("SLG", stats.Thousands(stats.Slugging)).WithPCT(),
		dataframe.DeriveInts("OPS", stats.Thousands(stats.OPS)).WithPCT(),
		dataframe.DeriveInts("AVG", stats.Thousands(stats.AVG)).WithPCT(),
		dataframe.DeriveInts("LAVG", stats.Thousands(stats.LAVG)).WithPCT(),
		dataframe.DeriveFloats("SWS%", func(idx *dataframe.Index, i int) float64 {
			m := idx.GetInt(i, "Misses")
			p := idx.GetInt(i, "PitchesSeen")
			if p > 0 {
				return 100.0 * float64(m) / float64(p)
			}
			return 0
		}).WithFormat("%4.1f").WithSummary(dataframe.Average),
		dataframe.DeriveFloats("C%", func(idx *dataframe.Index, i int) float64 {
			ab := idx.GetInt(i, "AB")
			k := idx.GetInt(i, "StrikeOuts")
			if ab > 0 {
				return 100.0 * float64(ab-k) / float64(ab)
			}
			return 0
		}).WithFormat("%5.1f").WithSummary(dataframe.Average),
		dataframe.DeriveFloats("LOOK", func(idx *dataframe.Index, i int) float64 {
			cs := idx.GetInt(i, "CalledStrikes")
			s := idx.GetInt(i, "Strikes")
			if s > 0 {
				return 100.0 * float64(cs) / float64(s)
			}
			return 0
		}).WithFormat("%4.1f").WithSummary(dataframe.Average),
		dataframe.DeriveFloats("SB2%", func(idx *dataframe.Index, i int) (res float64) {
			sb2 := idx.GetInt(i, "SB2")
			sb2opp := idx.GetInt(i, "SB2Opp")
			if sb2opp > 0 {
				res = 100.0 * float64(sb2) / float64(sb2opp)
			}
			return
		}).WithFormat("%5.1f").WithSummary(dataframe.Average),
		dataframe.Rename("SB2Opp", "S2O").WithSummary(dataframe.Sum),
	)
	idx = dat.GetIndex()
	dat = dat.RSort(func(r1, r2 int) bool {
		return idx.GetInt(r1, "OPS") > idx.GetInt(r2, "OPS")
	})
	return dat
}

func (r *Report) GetPitchingData() *dataframe.Data {
	return r.gs.GetPitchingData()
}

func (r *Report) GetAltData() *dataframe.Data {
	return r.gs.GetAltData()
}

func (r *Report) GetRE24Data() *dataframe.Data {
	return r.gs.GetRE24Data()
}
