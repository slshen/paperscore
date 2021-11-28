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
	outs    int
	runners string
	count   int
	runs    int
}

func (re *ObservedRunExpectancy) Read(g *game.Game) error {
	if re.states == nil {
		re.states = make([]*stateObservation, 24)
		for i := 0; i < 24; i++ {
			re.states[i] = &stateObservation{
				outs:    i / 8,
				runners: fmt.Sprintf("%d%d%d", (i>>2)&1, (i>>1)&1, i&1),
			}
		}
	}
	states := g.GetStates()
	observed := re.reset()
	for _, state := range states {
		if state.Outs == 3 {
			observed = re.reset()
			continue
		}
		for _, state24 := range observed {
			state24.runs += len(state.ScoringRunners)
		}
		index := re.getIndex(state.Outs, GetRunners(state))
		state24 := re.states[index]
		observed = append(observed, state24)
		state24.count++
	}
	return nil
}

func (re *ObservedRunExpectancy) WriteYAML(w io.Writer) error {
	for i := 0; i < 8; i++ {
		key := []rune{'_', '_', '_'}
		if (i & 1) != 0 {
			key[2] = '1'
		}
		if (i & 2) != 0 {
			key[1] = '2'
		}
		if (i & 4) != 0 {
			key[0] = '3'
		}
		_, err := fmt.Fprintf(w, "\"%s\": [ %.3f, %.3f, %.3f ]\n", string(key),
			re.states[i].getExpectedRuns(),
			re.states[i+8].getExpectedRuns(),
			re.states[i+16].getExpectedRuns())
		if err != nil {
			return err
		}
	}
	return nil
}

func (re *ObservedRunExpectancy) reset() []*stateObservation {
	nono := re.states[0]
	nono.count++
	return []*stateObservation{nono}
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
	state24 := re.states[re.getIndex(outs, runrs)]
	return state24.getExpectedRuns()
}

func (re *ObservedRunExpectancy) GetExpectedRunsCount(outs int, runrs Runners) int {
	state24 := re.states[re.getIndex(outs, runrs)]
	return state24.count
}

func (state24 *stateObservation) getExpectedRuns() (runs float64) {
	if state24.count > 0 {
		runs = float64(state24.runs) / float64(state24.count)
	}
	return
}
