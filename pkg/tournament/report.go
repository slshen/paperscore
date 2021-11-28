package tournament

import (
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Group struct {
	Key   interface{}
	Name  string
	Games []*game.Game
}

type Report struct {
	Us    string
	Group *Group

	gs *stats.GameStats
}

func (r *Report) Run(re stats.RunExpectancy) error {
	r.gs = stats.NewGameStats(re)
	for _, g := range r.Group.Games {
		if err := r.gs.Read(g); err != nil {
			return err
		}
	}
	return nil
}

func (r *Report) GetBattingData() *dataframe.Data {
	dat := r.gs.GetBattingData()
	idx := dat.GetIndex()
	dat.Name = r.Group.Name
	dat = dat.RFilter(func(row int) bool {
		t := strings.ToLower(idx.GetString(row, "Team"))
		return strings.HasPrefix(t, r.Us)
	})
	dat = dat.Select(
		dataframe.Col("Name"),
		dataframe.Col("PA").WithSummary(dataframe.Sum),
		dataframe.Col("AB").WithSummary(dataframe.Sum),
		dataframe.Rename("Hits", "H").WithSummary(dataframe.Sum),
		dataframe.Col("LOPH").WithSummary(dataframe.Sum),
		dataframe.DeriveInts("XBH", func(idx *dataframe.Index, i int) int {
			return idx.GetInt(i, "Doubles") + idx.GetInt(i, "Triples") + idx.GetInt(i, "HRs")
		}).WithFormat("%3d").WithSummary(dataframe.Sum),
		dataframe.Rename("Walks", "BB").WithSummary(dataframe.Sum),
		dataframe.Rename("StrikeOuts", "K").WithSummary(dataframe.Sum),
		dataframe.Col("RE24").WithSummary(dataframe.Average),
		dataframe.DeriveInts("OBP", obp).WithPCT(),
		dataframe.DeriveInts("SLG", slg).WithPCT(),
		dataframe.DeriveInts("OPS", func(idx *dataframe.Index, i int) int {
			return obp(idx, i) + slg(idx, i)
		}).WithPCT(),
		dataframe.DeriveInts("AVG", func(idx *dataframe.Index, i int) int {
			h := idx.GetInt(i, "Hits")
			ab := idx.GetInt(i, "AB")
			if ab > 0 {
				return int(1000.0 * float64(h) / float64(ab))
			}
			return 0
		}).WithPCT(),
		dataframe.DeriveInts("LAVG", func(idx *dataframe.Index, i int) int {
			h := idx.GetInt(i, "Hits")
			lo := idx.GetInt(i, "LineDriveOuts")
			ab := idx.GetInt(i, "AB")
			if ab > 0 {
				return int(1000.0 * float64(h+lo) / float64(ab))
			}
			return 0
		}).WithPCT(),
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

func obp(idx *dataframe.Index, i int) int {
	hbp := idx.GetInt(i, "HitByPitch")
	h := idx.GetInt(i, "Hits")
	bb := idx.GetInt(i, "Walks")
	ab := idx.GetInt(i, "AB")
	sf := idx.GetInt(i, "SacrificeFlys")
	obp := float64(h+bb+hbp) / float64(ab+bb+hbp+sf)
	return int(1000 * obp)
}

func slg(idx *dataframe.Index, i int) (val int) {
	ab := idx.GetInt(i, "AB")
	if ab > 0 {
		s := idx.GetInt(i, "Singles")
		d := idx.GetInt(i, "Doubles")
		t := idx.GetInt(i, "Triples")
		h := idx.GetInt(i, "HRs")
		// fmt.Printf("%s %d %d %d %d\n", idx.GetString(i, "Name"), s, d, t, h)
		slg := float64(s+2*d+3*t+4*h) / float64(ab)
		val = int(1000.0 * slg)
	}
	return
}
