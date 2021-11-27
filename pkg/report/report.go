package report

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
	dat.Columns = append(dat.Columns, obp(dat))
	dat.Columns = append(dat.Columns, slg(dat))
	dat.Columns = append(dat.Columns, xbh(dat))
	dat.Columns = append(dat.Columns, ops(dat))
	dat.Columns = append(dat.Columns, sws(dat))
	dat.Columns = append(dat.Columns, contact(dat))
	dat.Columns = append(dat.Columns, look(dat))
	dat.Columns = append(dat.Columns, avg(dat))
	dat.Columns = append(dat.Columns, lavg(dat))
	dat = dat.Select(
		dataframe.Col("Name"),
		dataframe.Col("PA").WithSummary(dataframe.Sum),
		dataframe.Col("AB").WithSummary(dataframe.Sum),
		dataframe.Rename("Hits", "H").WithSummary(dataframe.Sum),
		dataframe.Col("LOPH").WithSummary(dataframe.Sum),
		dataframe.Col("XBH"),
		dataframe.Rename("Walks", "BB").WithSummary(dataframe.Sum),
		dataframe.Col("RE24").WithSummary(dataframe.Average),
		dataframe.Col("OBP"),
		dataframe.Col("SLG"),
		dataframe.Col("OPS"),
		dataframe.Col("AVG"),
		dataframe.Col("LAVG"),
		dataframe.Col("SWS%"),
		dataframe.Col("C%"),
		dataframe.Col("LOOK"),
	)
	idx = dat.GetIndex()
	dat = dat.RSort(func(r1, r2 int) bool {
		return idx.GetInt(r1, "OPS") > idx.GetInt(r2, "OPS")
	})

	return dat
}

func deriveIntValues(dat *dataframe.Data, f func(idx *dataframe.Index, r int) int) []int {
	idx := dat.GetIndex()
	values := make([]int, dat.RowCount())
	for i := range values {
		values[i] = f(idx, i)
	}
	return values
}

func deriveFloatValues(dat *dataframe.Data, f func(idx *dataframe.Index, r int) float64) []float64 {
	idx := dat.GetIndex()
	values := make([]float64, dat.RowCount())
	for i := range values {
		values[i] = f(idx, i)
	}
	return values
}

func xbh(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:    "XBH",
		Summary: dataframe.Sum,
		Format:  "%3d",
		Values: deriveIntValues(dat, func(idx *dataframe.Index, i int) int {
			return idx.GetInt(i, "Doubles") + idx.GetInt(i, "Triples") + idx.GetInt(i, "HRs")
		}),
	}
}

func slg(dat *dataframe.Data) *dataframe.Column {
	values := deriveIntValues(dat, func(idx *dataframe.Index, i int) (val int) {
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
	})
	return &dataframe.Column{
		Name:          "SLG",
		Summary:       dataframe.Average,
		SummaryFormat: "%4.0f",
		Format:        "%4d",
		Values:        values,
	}
}

func ops(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:          "OPS",
		Summary:       dataframe.Average,
		SummaryFormat: "%4.0f",
		Format:        "%4d",
		Values: deriveIntValues(dat, func(idx *dataframe.Index, i int) int {
			return idx.GetInt(i, "OBP") + idx.GetInt(i, "SLG")
		}),
	}
}

func obp(dat *dataframe.Data) *dataframe.Column {
	values := deriveIntValues(dat, func(idx *dataframe.Index, i int) int {
		hbp := idx.GetInt(i, "HitByPitch")
		h := idx.GetInt(i, "Hits")
		bb := idx.GetInt(i, "Walks")
		ab := idx.GetInt(i, "AB")
		sf := idx.GetInt(i, "SacrificeFlys")
		obp := float64(h+bb+hbp) / float64(ab+bb+hbp+sf)
		return int(1000 * obp)
	})
	return &dataframe.Column{
		Name:          "OBP",
		Summary:       dataframe.Average,
		SummaryFormat: "%4.0f",
		Format:        "%4d",
		Values:        values,
	}
}

func sws(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:    "SWS%",
		Format:  "%4.1f",
		Summary: dataframe.Average,
		Values: deriveFloatValues(dat, func(idx *dataframe.Index, r int) float64 {
			m := idx.GetInt(r, "Misses")
			p := idx.GetInt(r, "PitchesSeen")
			if p > 0 {
				return 100.0 * float64(m) / float64(p)
			}
			return 0
		}),
	}
}

func contact(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:    "C%",
		Format:  "%5.1f",
		Summary: dataframe.Average,
		Values: deriveFloatValues(dat, func(idx *dataframe.Index, r int) float64 {
			ab := idx.GetInt(r, "AB")
			k := idx.GetInt(r, "StrikeOuts")
			if ab > 0 {
				return 100.0 * float64(ab-k) / float64(ab)
			}
			return 0
		}),
	}
}

func look(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:    "LOOK",
		Format:  "%5.1f",
		Summary: dataframe.Average,
		Values: deriveFloatValues(dat, func(idx *dataframe.Index, r int) float64 {
			cs := idx.GetInt(r, "CalledStrikes")
			s := idx.GetInt(r, "Strikes")
			if s > 0 {
				return 100.0 * float64(cs) / float64(s)
			}
			return 0
		}),
	}
}

func lavg(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:          "LAVG",
		Summary:       dataframe.Average,
		SummaryFormat: "%4.0f",
		Format:        "%4d",
		Values: deriveIntValues(dat, func(idx *dataframe.Index, r int) int {
			h := idx.GetInt(r, "Hits")
			lo := idx.GetInt(r, "LineDriveOuts")
			ab := idx.GetInt(r, "AB")
			if ab > 0 {
				return int(1000.0 * float64(h+lo) / float64(ab))
			}
			return 0
		}),
	}
}

func avg(dat *dataframe.Data) *dataframe.Column {
	return &dataframe.Column{
		Name:          "AVG",
		Summary:       dataframe.Average,
		SummaryFormat: "%4.0f",
		Format:        "%4d",
		Values: deriveIntValues(dat, func(idx *dataframe.Index, r int) int {
			h := idx.GetInt(r, "Hits")
			ab := idx.GetInt(r, "AB")
			if ab > 0 {
				return int(1000.0 * float64(h) / float64(ab))
			}
			return 0
		}),
	}
}
