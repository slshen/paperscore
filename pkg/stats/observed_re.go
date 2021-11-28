package stats

import (
	"fmt"
	"io"

	"github.com/slshen/sb/pkg/game"
)

type ObservedRunExpectancy struct {
	states []*stateObservation
}

var _ RunExpectancy = (*ObservedRunExpectancy)(nil)
var _ RunExpectancyCounts = (*ObservedRunExpectancy)(nil)

type stateObservation struct {
	outs  int
	count int
	runs  int
}

func (re *ObservedRunExpectancy) Read(g *game.Game) error {
	if re.states == nil {
		re.states = make([]*stateObservation, 24)
		for i := 0; i < 24; i++ {
			re.states[i] = &stateObservation{
				outs: i / 8,
			}
		}
	}
	states := g.GetStates()
	observed := re.reset(nil)
	for _, state := range states {
		for _, state24 := range observed {
			if state24 != nil {
				state24.runs += len(state.ScoringRunners)
			}
		}
		if state.Outs == 3 {
			observed = re.reset(observed)
			continue
		}
		index := re.getIndex(state.Outs, GetRunners(state))
		if observed[index] == nil {
			observed[index] = &stateObservation{}
		}
		state24 := observed[index]
		state24.count++
	}
	return nil
}

func (re *ObservedRunExpectancy) reset(observed []*stateObservation) []*stateObservation {
	for i, obs := range observed {
		if obs != nil {
			st := re.states[i]
			st.count += obs.count
			st.runs += obs.runs
		}
	}
	res := make([]*stateObservation, 24)
	res[0] = &stateObservation{
		count: 1,
	}
	return res
}

func (re *ObservedRunExpectancy) getIndex(outs int, runrs Runners) int {
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
	if re.states == nil {
		return 0
	}
	state24 := re.states[re.getIndex(outs, runrs)]
	return state24.getExpectedRuns()
}

func (re *ObservedRunExpectancy) GetExpectedRunsCount(outs int, runrs Runners) int {
	if re.states == nil {
		return 0
	}
	state24 := re.states[re.getIndex(outs, runrs)]
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
			re.states[i].getExpectedRuns(),
			re.states[i+8].getExpectedRuns(),
			re.states[i+16].getExpectedRuns())
		if err != nil {
			return err
		}
	}
	return nil
}
