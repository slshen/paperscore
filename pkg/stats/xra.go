package stats

import (
	"fmt"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type ExcessRunsAllowed struct {
	XRA float64

	re RunExpectancy
}

type Counterfactual struct {
	Outs int
	OccupiedBases
	RunsScored            int
	ResponsibleFielders   []int
	AssumeNoAdvanceErrors bool
}

func (stats *ExcessRunsAllowed) record(state, lastState *game.State) {
	if stats.re == nil {
		return
	}
	cf := stats.getCounterfactualState(state, lastState)
	if cf == nil {
		return
	}
	if cf.AssumeNoAdvanceErrors {
		for _, adv := range state.Advances {
			if adv.FieldingError != nil {
				panic(fmt.Sprintf("can't handle errors in advances for %s", state.EventCode))
			}
		}
	}
	actualRuns := float64(state.GetRunsScored()) - GetExpectedRuns(stats.re, lastState)
	if state.Outs < 3 {
		actualRuns += GetExpectedRuns(stats.re, state)
	}
	cfRuns := float64(cf.RunsScored) - GetExpectedRuns(stats.re, lastState)
	if cf.Outs < 3 {
		cfRuns += stats.re.GetExpectedRuns(cf.Outs, cf.OccupiedBases)
	}
	if actualRuns < cfRuns {
		// so it's possible that a clown show happens and it results in a better state
		// e.g. K+PB then thrown out at home.  So ignore this kinda of thing.
		return
	}
	stats.XRA += actualRuns - cfRuns
}

func (stats *ExcessRunsAllowed) getCounterfactualState(state, lastState *game.State) *Counterfactual {
	switch state.Play.Type {
	case game.StrikeOutPassedBall:
		fallthrough
	case game.StrikeOutWildPitch:
		// For both K+WP and K+PB we assume the batter would have struck out and
		// no runners advanced
		fallthrough
	case game.FoulFlyError:
		return &Counterfactual{
			OccupiedBases:         GetOccupiedBases(lastState),
			Outs:                  lastState.Outs + 1,
			ResponsibleFielders:   state.Fielders,
			RunsScored:            0,
			AssumeNoAdvanceErrors: true,
		}
	case game.ReachedOnError:
		cf := &Counterfactual{
			Outs:                lastState.Outs + 1,
			ResponsibleFielders: state.Fielders,
			OccupiedBases:       GetOccupiedBases(lastState),
		}
		if lastState.Outs < 2 {
			// With less than 2 outs, the runner may have advanced
			switch cf.OccupiedBases {
			case BasesEmpty:
				// no
			case RunnerOnFirst:
				fallthrough
			case RunnerOnSecond:
				if state.Modifiers.Trajectory() == game.Bunt ||
					state.Play.FieldingError.Fielder == 3 ||
					state.Play.FieldingError.Fielder == 4 {
					cf.OccupiedBases = RunnerOnSecond
				}
			default:
				panic("can't handle " + state.EventCode)
			}
		}
		return cf
	case game.CatcherInterference:
		// Assume the pitch would have been fouled off or otherwise
		// did not end the at-bat
		return &Counterfactual{
			OccupiedBases:         GetOccupiedBases(lastState),
			Outs:                  lastState.Outs,
			ResponsibleFielders:   state.Fielders,
			RunsScored:            0,
			AssumeNoAdvanceErrors: true,
		}
	case game.PassedBall:
		return &Counterfactual{
			OccupiedBases:         GetOccupiedBases(lastState),
			Outs:                  lastState.Outs,
			RunsScored:            0,
			ResponsibleFielders:   []int{2},
			AssumeNoAdvanceErrors: true,
		}
	case game.WildPitch:
		// WP is counted under pitching stats, so we'll not count it as
		// excess runs scored.  But we don't handle errors on a WP, so
		// assert that.
		return &Counterfactual{
			OccupiedBases:         GetOccupiedBases(state),
			Outs:                  state.Outs,
			RunsScored:            state.GetRunsScored(),
			AssumeNoAdvanceErrors: true,
		}
	}
	return nil
}

func (stats *ExcessRunsAllowed) GetXRAData() *dataframe.Data {
	return &dataframe.Data{
		Columns: []*dataframe.Column{
			{Name: "XRA", Values: dataframe.EmptyFloats},
		},
	}
}

func minint(i, j int) int {
	if i < j {
		return i
	}
	return j
}
