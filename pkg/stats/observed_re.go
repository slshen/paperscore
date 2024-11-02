package stats

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type ObservedRunExpectancy struct {
	totals     []*stateObservation
	inProgress []*stateObservation
	runData    *dataframe.Data
}

var _ RunExpectancy = (*ObservedRunExpectancy)(nil)
var _ RunExpectancyCounts = (*ObservedRunExpectancy)(nil)

type stateObservation struct {
	count int
	runs  int
}

func (re *ObservedRunExpectancy) Read(g *game.Game) error {
	if re.totals == nil {
		re.totals = make([]*stateObservation, 24)
		re.inProgress = make([]*stateObservation, 24)
		for i := 0; i < 24; i++ {
			re.totals[i] = &stateObservation{}
		}
		re.inProgress[0] = &stateObservation{count: 1}
		re.runData = &dataframe.Data{
			Columns: []*dataframe.Column{
				dataframe.NewColumn("Index", "%d", dataframe.EmptyInts),
				dataframe.NewColumn("Runs", "%4d", dataframe.EmptyInts),
			},
		}
	}
	states := g.GetStates()
	for _, state := range states {
		for _, state24 := range re.inProgress {
			if state24 != nil {
				state24.runs += len(state.ScoringRunners)
			}
		}
		if state.Outs == 3 {
			for i, p := range re.inProgress {
				if p != nil {
					re.totals[i].count += p.count
					re.totals[i].runs += p.runs
					re.runData.Columns[0].AppendInt(i)
					// cap # of runs to 9
					runs := p.runs
					if runs > 9 {
						runs = 9
					}
					re.runData.Columns[1].AppendInt(runs)
					if i == 0 {
						re.inProgress[i].count = 1
						re.inProgress[i].runs = 0
					} else {
						re.inProgress[i] = nil
					}
				}
			}
			continue
		}
		index := re.getIndex(state.Outs, GetOccupiedBases(state))
		if re.inProgress[index] == nil {
			re.inProgress[index] = &stateObservation{
				count: 1,
			}
		}
		re.inProgress[index].count++
	}
	return nil
}

func (re *ObservedRunExpectancy) getIndex(outs int, runrs OccupiedBases) int {
	if outs == 3 {
		return 0
	}
	index := outs * 8
	if runrs[2] != '_' {
		index |= 1
	}
	if runrs[1] != '_' {
		index |= 2
	}
	if runrs[0] != '_' {
		index |= 4
	}
	return index
}

func (re *ObservedRunExpectancy) GetExpectedRuns(outs int, runrs OccupiedBases) float64 {
	if re.totals == nil {
		return 0
	}
	state24 := re.totals[re.getIndex(outs, runrs)]
	return state24.getExpectedRuns()
}

func (re *ObservedRunExpectancy) GetExpectedRunsCount(outs int, runrs OccupiedBases) int {
	if re.totals == nil {
		return 0
	}
	state24 := re.totals[re.getIndex(outs, runrs)]
	return state24.count
}

func (state24 *stateObservation) getExpectedRuns() (runs float64) {
	if state24.count > 0 {
		runs = float64(state24.runs) / float64(state24.count)
	}
	return
}

func (re *ObservedRunExpectancy) WriteYAML(w io.Writer) error {
	for i, runrs := range OccupedBasesValues {
		_, err := fmt.Fprintf(w, "\"%s\": [ %.3f, %.3f, %.3f ]\n", runrs,
			re.totals[i].getExpectedRuns(),
			re.totals[i+8].getExpectedRuns(),
			re.totals[i+16].getExpectedRuns())
		if err != nil {
			return err
		}
	}
	return nil
}

func (re *ObservedRunExpectancy) GetRunData() *dataframe.Data {
	return re.runData.Select(
		deriveState24(),
		dataframe.Col("Runs").WithFormat("%4d"),
	)
}

type RunFrequency struct {
	dataframe.Data
}

func (re *ObservedRunExpectancy) GetRunExpectancyFrequency() *RunFrequency {
	// group by Index, Runs, Count(*) as Obs
	dat := re.runData.GroupBy("Index", "Runs").Aggregate(
		dataframe.ACount("Obs").WithFormat("%4d"),
	)
	dat = dat.RSort(dataframe.Less(
		dataframe.CompareInt(dat.Columns[0]),
		dataframe.CompareInt(dat.Columns[1]),
	))
	// group by Index, Sum(Obs)
	obsIndexDat := dat.GroupBy("Index").Aggregate(
		dataframe.ASum("ObsIndex", dat.Columns[2]),
	)
	sumObsPerIndex := map[int]int{}
	for i, index := range obsIndexDat.Columns[0].GetInts() {
		sumObsPerIndex[index] = obsIndexDat.Columns[1].GetInt(i)
	}
	result := &dataframe.Data{
		Columns: []*dataframe.Column{
			dataframe.NewEmptyColumn("Index", dataframe.Int),
			dataframe.NewEmptyColumn("Runs", dataframe.Int),
			dataframe.NewEmptyColumn("Obs", dataframe.Int),
			dataframe.NewEmptyColumn("Tot", dataframe.Int),
			dataframe.NewEmptyColumn("Freq", dataframe.Float),
		},
	}
	for i := 0; i < 24; i++ {
		irow := sort.SearchInts(dat.Columns[0].GetInts(), i)
		ip1row := sort.SearchInts(dat.Columns[0].GetInts(), i+1)
		for runs := 0; runs < 10; runs++ {
			result.Columns[0].AppendInt(i)
			result.Columns[1].AppendInt(runs)
			runValues := dat.Columns[1].GetInts()[irow:ip1row]
			runRow := sort.SearchInts(runValues, runs)
			var obs int
			if runRow < len(runValues) && runValues[runRow] == runs {
				obs = dat.Columns[2].GetInt(irow + runRow)
			}
			result.Columns[2].AppendInt(obs)
			tot := sumObsPerIndex[i]
			result.Columns[3].AppendInt(tot)
			var freq float64
			if tot > 0 {
				freq = float64(obs) / float64(tot)
			}
			result.Columns[4].AppendFloat(freq)
		}
	}
	result = result.Select(
		deriveState24(),
		dataframe.Rename("Runs", "R").WithFormat("%1d"),
		dataframe.Col("Obs").WithFormat("%3d"),
		dataframe.Col("Tot").WithFormat("%3d"),
		dataframe.Col("Freq").WithFormat("%5.3f"),
	)
	result = result.RSort(dataframe.Less(
		dataframe.Descending(cmpSt24(result.Columns[0])),
		dataframe.CompareInt(result.Columns[1]),
	))
	return &RunFrequency{*result}
}

func deriveState24() dataframe.Selection {
	return dataframe.DeriveStrings("St24", func(idx *dataframe.Index, i int) string {
		return fmt.Sprintf("%s/%d", OccupedBasesValues[idx.GetInt(i, "Index")%8],
			idx.GetInt(i, "Index")/8)
	}).WithFormat("%5s")
}

func (dat *RunFrequency) Pivot() *dataframe.Data {
	idx := dat.GetIndex()
	runsCol := idx.GetColumn("R")
	obsCol := idx.GetColumn("Obs")
	totCol := idx.GetColumn("Tot")
	aggs := make([]dataframe.Aggregation, 20)
	for i := 0; i < 10; i++ {
		runs := i
		aggs[i] = dataframe.AFunc(fmt.Sprintf("PcRn%d", i), dataframe.Float,
			func(acol *dataframe.Column, group *dataframe.Group) {
				for _, row := range group.Rows {
					if runsCol.GetInt(row) == runs {
						freq := float64(obsCol.GetInt(row)) / float64(totCol.GetInt(row))
						acol.AppendFloat(100 * freq)
						return
					}
				}
				acol.AppendFloat(0)
			}).WithFormat("%5.1f")
		aggs[10+i] = dataframe.AFunc(fmt.Sprintf("CnRn%d", i), dataframe.Int,
			func(acol *dataframe.Column, group *dataframe.Group) {
				for _, row := range group.Rows {
					if runsCol.GetInt(row) == runs {
						acol.AppendInt(obsCol.GetInt(row))
						return
					}
				}
				acol.AppendInt(0)
			}).WithFormat("%5d")
	}
	pivot := dat.GroupBy("St24").Aggregate(aggs...)
	return pivot.RSort(dataframe.Less(
		dataframe.Descending(cmpSt24(pivot.Columns[0])),
	))
}

func cmpSt24(col *dataframe.Column) func(i, j int) int {
	return func(i, j int) int {
		s1 := col.GetString(i)
		s2 := col.GetString(j)
		c := strings.Compare(s1[0:3], s2[0:3])
		if c != 0 {
			return c
		}
		o1, _ := strconv.Atoi(s1[4:])
		o2, _ := strconv.Atoi(s2[4:])
		return o2 - o1
	}
}
