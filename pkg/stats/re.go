package stats

import (
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type RunExpectancy interface {
	GetExpectedRuns(outs int, runrs Runners) float64
}

type RunExpectancyCounts interface {
	GetExpectedRunsCount(outs int, runrs Runners) int
}

type Runners string

var RunnersValues = []Runners{"___", "__1", "_2_", "_21", "3__", "3_1", "32_", "321"}
var NoOneOnNoOuts = RunnersValues[0]

func GetRunners(state *game.State) Runners {
	if state == nil {
		return NoOneOnNoOuts
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
	return Runners(k)
}

func GetExpectedRuns(re RunExpectancy, state *game.State) float64 {
	if state == nil || state.Outs == 3 {
		return re.GetExpectedRuns(0, NoOneOnNoOuts)
	}
	return re.GetExpectedRuns(state.Outs, GetRunners(state))
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
	for _, runrs := range RunnersValues {
		runners.AppendString(string(runrs))
		_0out.AppendFloats(re.GetExpectedRuns(0, runrs))
		_1out.AppendFloats(re.GetExpectedRuns(1, runrs))
		_2out.AppendFloats(re.GetExpectedRuns(2, runrs))
		if count != nil {
			_0outCount.AppendInts(count.GetExpectedRunsCount(0, runrs))
			_1outCount.AppendInts(count.GetExpectedRunsCount(1, runrs))
			_2outCount.AppendInts(count.GetExpectedRunsCount(2, runrs))
		}
	}
	return dat
}
