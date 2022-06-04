package stats

import (
	"fmt"
	"io"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
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
					re.runData.Columns[0].AppendInts(i)
					re.runData.Columns[1].AppendInts(p.runs)
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
			re.inProgress[index] = &stateObservation{}
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

func (re *ObservedRunExpectancy) GetRunExpectancyFrequency() *dataframe.Data {
	// group by Index, Runs, Count(*) as Obs
	dat := re.runData.GroupBy("Index", "Runs").Aggregate(
		dataframe.Aggregation{
			Column: dataframe.NewColumn("Obs", "%4d", dataframe.EmptyInts),
			Func: func(col *dataframe.Column, group *dataframe.Group) {
				col.AppendInts(len(group.Rows))
			},
		},
	)
	// group by Index, Sum(Obs)
	obsIndexDat := dat.GroupBy("Index").Aggregate(
		dataframe.Aggregation{
			Column: dataframe.NewColumn("ObsIndex", "%d", dataframe.EmptyInts),
			Func: func(col *dataframe.Column, group *dataframe.Group) {
				obs := 0
				for _, row := range group.Rows {
					obs += dat.Columns[2].GetInt(row)
				}
				col.AppendInts(obs)
			},
		},
	)
	sumObsPerIndex := map[int]int{}
	for i, index := range obsIndexDat.Columns[0].GetInts() {
		sumObsPerIndex[index] = obsIndexDat.Columns[1].GetInt(i)
	}
	dat = dat.Select(
		dataframe.DeriveStrings("Rnrs", func(idx *dataframe.Index, i int) string {
			return string(OccupedBasesValues[idx.GetInt(i, "Index")%8])
		}).WithFormat("%4s"),
		dataframe.DeriveInts("Outs", func(idx *dataframe.Index, i int) int {
			return idx.GetInt(i, "Index") / 8
		}).WithFormat("%4d"),
		dataframe.Col("Runs"),
		dataframe.Col("Obs"),
		dataframe.DeriveFloats("Freq", func(idx *dataframe.Index, i int) float64 {
			index := idx.GetInt(i, "Index")
			sum := sumObsPerIndex[index]
			return float64(idx.GetInt(i, "Obs")) / float64(sum)
		}).WithFormat("%5.3f"),
	)
	return dat.RSort(dataframe.Less(
		dataframe.Descending(dataframe.CompareString(dat.Columns[0])),
		dataframe.CompareInt(dat.Columns[1]),
		dataframe.CompareInt(dat.Columns[2]),
	))
}
