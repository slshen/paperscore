package markov

import (
	"fmt"
	"math/rand"
	"time"
)

type Model interface {
	NextState(rnd *rand.Rand, batter string, current BaseOutState) (next BaseOutState, event string, runs float64)
}

type Event struct {
	State, Next BaseOutState
	Event       string
	Runs        float64
}

type Simulation struct {
	Model             Model
	Runs              float64
	Trace             []Event
	totalRunsPerState []float64
	countPerState     []int
	innings           int
}

func (sim *Simulation) RunInning() error {
	// #nosec G404
	rnd := rand.New(rand.NewSource(time.Now().UnixMicro()))
	state := StartState
	step := 0
	runsAtState := make([]*float64, 24)
	zero := 0.0
	runsAtState[state] = &zero
	var inningRuns float64
	for state != EndState {
		next, event, runs := sim.Model.NextState(rnd, "", state)
		if sim.Trace != nil {
			sim.Trace = append(sim.Trace, Event{
				State: state,
				Next:  next,
				Event: event,
				Runs:  runs,
			})
		}
		inningRuns += runs
		if next != EndState && runsAtState[next] == nil {
			t := inningRuns
			runsAtState[next] = &t
		}
		state = next
		if step > 100 {
			return fmt.Errorf("ran %d events without terminating", step)
		}
		step++
	}
	sim.Runs += inningRuns
	if sim.totalRunsPerState == nil {
		sim.totalRunsPerState = make([]float64, 24)
		sim.countPerState = make([]int, 24)
	}
	for state, runs := range runsAtState {
		if runs != nil {
			sim.totalRunsPerState[state] += inningRuns - *runs
			sim.countPerState[state]++
		}
	}
	sim.innings++
	return nil
}

func (sim *Simulation) GetExpectedRuns() []float64 {
	re := make([]float64, 24)
	state := StartState
	for state != EndState {
		count := sim.countPerState[state]
		if count > 0 {
			re[state] = sim.totalRunsPerState[state] / float64(count)
		}
		state++
	}
	return re
}
