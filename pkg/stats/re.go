package stats

import (
	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type RunExpectancy interface {
	GetExpectedRuns(outs int, runrs OccupiedBases) float64
}

type RunExpectancyCounts interface {
	GetExpectedRunsCount(outs int, runrs OccupiedBases) int
}

type OccupiedBases string

var (
	OccupedBasesValues      = []OccupiedBases{"___", "__1", "_2_", "_21", "3__", "3_1", "32_", "321"}
	BasesEmpty              = OccupedBasesValues[0]
	RunnerOnFirst           = OccupedBasesValues[1]
	RunnerOnSecond          = OccupedBasesValues[2]
	RunnerOnFirstAndSecond  = OccupedBasesValues[3]
	RunnerOnThird           = OccupedBasesValues[4]
	RunnersOnFirstAndThird  = OccupedBasesValues[5]
	RunnersOnSecondAndThird = OccupedBasesValues[6]
	BasesLoaded             = OccupedBasesValues[7]
)

func GetOccupiedBases(state *game.State) OccupiedBases {
	if state == nil {
		return BasesEmpty
	}
	k := []rune{'_', '_', '_'}
	if state.Runners[0] != "" {
		k[2] = '1'
	}
	if state.Runners[1] != "" {
		k[1] = '2'
	}
	if state.Runners[2] != "" {
		k[0] = '3'
	}
	return OccupiedBases(k)
}

func GetExpectedRuns(re RunExpectancy, state *game.State) float64 {
	if state == nil || state.Outs == 3 {
		return re.GetExpectedRuns(0, BasesEmpty)
	}
	return re.GetExpectedRuns(state.Outs, GetOccupiedBases(state))
}

func GetExpectedRunsChange(re RunExpectancy, state *game.State) (runsBefore float64, runsAfter float64, runsScored int, change float64) {
	runsBefore = GetExpectedRuns(re, state.LastState)
	if state.Outs < 3 {
		runsAfter = GetExpectedRuns(re, state)
	}
	runsScored = len(state.ScoringRunners)
	change = runsAfter - runsBefore + float64(runsScored)
	return
}

func GetRunExpectancyData(re RunExpectancy) *dataframe.Data {
	var (
		runners    = &dataframe.Column{Name: "Runr", Format: "%4s"}
		_0out      = &dataframe.Column{Name: "0Out", Format: "%5.3f"}
		_1out      = &dataframe.Column{Name: "1Out", Format: "%5.3f"}
		_2out      = &dataframe.Column{Name: "2Out", Format: "%5.3f"}
		_0outCount *dataframe.Column
		_1outCount *dataframe.Column
		_2outCount *dataframe.Column
	)
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{runners, _0out, _1out, _2out},
	}
	count, _ := re.(RunExpectancyCounts)
	if count != nil {
		_0outCount = &dataframe.Column{Name: "0OutCount"}
		_1outCount = &dataframe.Column{Name: "1OutCount"}
		_2outCount = &dataframe.Column{Name: "2OutCount"}
		dat.Columns = append(dat.Columns, _0outCount, _1outCount, _2outCount)
	}
	for _, runrs := range OccupedBasesValues {
		runners.AppendString(string(runrs))
		_0out.AppendFloat(re.GetExpectedRuns(0, runrs))
		_1out.AppendFloat(re.GetExpectedRuns(1, runrs))
		_2out.AppendFloat(re.GetExpectedRuns(2, runrs))
		if count != nil {
			_0outCount.AppendInt(count.GetExpectedRunsCount(0, runrs))
			_1outCount.AppendInt(count.GetExpectedRunsCount(1, runrs))
			_2outCount.AppendInt(count.GetExpectedRunsCount(2, runrs))
		}
	}
	return dat
}
