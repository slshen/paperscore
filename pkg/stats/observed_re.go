package stats

import (
	"fmt"
	"io"

	"github.com/slshen/sb/pkg/game"
)

type ObservedRunExpectancy struct {
	totals     []*stateObservation
	inProgress []*stateObservation
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
		re.inProgress[0] = &stateObservation{}
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
		index := re.getIndex(state.Outs, GetRunners(state))
		if re.inProgress[index] == nil {
			re.inProgress[index] = &stateObservation{}
		}
		re.inProgress[index].count++
	}
	return nil
}

func (re *ObservedRunExpectancy) getIndex(outs int, runrs Runners) int {
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

func (re *ObservedRunExpectancy) GetExpectedRuns(outs int, runrs Runners) float64 {
	if re.totals == nil {
		return 0
	}
	state24 := re.totals[re.getIndex(outs, runrs)]
	return state24.getExpectedRuns()
}

func (re *ObservedRunExpectancy) GetExpectedRunsCount(outs int, runrs Runners) int {
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
	for i, runrs := range RunnersValues {
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
