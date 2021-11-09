package stats

import (
	"fmt"

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

func GetRunExpectancyData(re RunExpectancy) *Data {
	data := &Data{
		Columns: []string{"Runners", "0Out", "1Out", "2Out"},
		Width:   map[string]int{"0Out": 5, "1Out": 5, "2Out": 5},
	}
	count, _ := re.(RunExpectancyCounts)
	if count != nil {
		data.Columns = append(data.Columns, "0OutCount", "1OutCount", "2OutCount")
	}
	for i := 0; i < 8; i++ {
		runners := reRunnersKey[i]
		first := (i & 1) != 0
		second := (i & 2) != 0
		third := (i & 4) != 0
		row := Row{
			runners,
			fmt.Sprintf("%0.3f", re.GetExpectedRuns(0, first, second, third)),
			fmt.Sprintf("%0.3f", re.GetExpectedRuns(1, first, second, third)),
			fmt.Sprintf("%0.3f", re.GetExpectedRuns(2, first, second, third)),
		}
		if count != nil {
			row = append(row,
				count.GetExpectedRunsCount(0, first, second, third),
				count.GetExpectedRunsCount(1, first, second, third),
				count.GetExpectedRunsCount(2, first, second, third),
			)
		}
		data.Rows = append(data.Rows, row)
	}
	return data
}
