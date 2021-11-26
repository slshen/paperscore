package stats

import (
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type RunExpectancy interface {
	GetExpectedRuns(outs int, first, second, third bool) float64
}

type RunExpectancyCounts interface {
	GetExpectedRunsCount(outs int, first, second, third bool) int
}

var reRunnersKey = []string{"___", "__1", "_2_", "_21", "3__", "3_1", "32_", "321"}

func GetExpectedRuns(re RunExpectancy, state *game.State) float64 {
	if state == nil || state.Outs == 3 {
		return re.GetExpectedRuns(0, false, false, false)
	}
	return re.GetExpectedRuns(state.Outs,
		state.Runners[0] != "", state.Runners[1] != "", state.Runners[2] != "")
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
	for i := 0; i < 8; i++ {
		runners.AppendString(reRunnersKey[i])
		first := (i & 1) != 0
		second := (i & 2) != 0
		third := (i & 4) != 0
		_0out.AppendFloats(re.GetExpectedRuns(0, first, second, third))
		_1out.AppendFloats(re.GetExpectedRuns(1, first, second, third))
		_2out.AppendFloats(re.GetExpectedRuns(2, first, second, third))
		if count != nil {
			_0outCount.AppendInts(count.GetExpectedRunsCount(0, first, second, third))
			_1outCount.AppendInts(count.GetExpectedRunsCount(1, first, second, third))
			_2outCount.AppendInts(count.GetExpectedRunsCount(2, first, second, third))
		}
	}
	return dat
}
