package stats

import (
	"fmt"
	"os"
	"regexp"

	"github.com/slshen/sb/pkg/game"
)

var debugRE4 = os.Getenv("DEBUG_RE24") != ""

func getBattingRE24Change(re RunExpectancy, state, lastState *game.State) float64 {
	runsBefore := GetExpectedRuns(re, lastState)
	var runsAfter float64
	if state.Outs < 3 {
		runsAfter = GetExpectedRuns(re, state)
	}
	runsScored := float64(len(state.ScoringRunners))
	change := runsAfter - runsBefore + runsScored
	if debugRE4 {
		var outs int
		if lastState != nil {
			outs = lastState.Outs
		}
		if m, _ := regexp.MatchString(`^[a-z]`, string(state.Batter)); m {
			fmt.Printf("%4s %d %s %30s : %5.3f - %5.3f + %.0f = % 5.3f\n",
				state.Batter, outs, GetRunners(lastState), state.EventCode, runsAfter, runsBefore, runsScored, change)
		}
	}
	return change
}
