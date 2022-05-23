package stats

import (
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type REData struct {
	re                                                      RunExpectancy
	game, id, o, rnr, play, after, before, r, re24, runners *dataframe.Column
}

func NewREData(re RunExpectancy) *REData {
	return &REData{
		re:      re,
		game:    dataframe.NewColumn("Game", "%10s", dataframe.EmptyStrings),
		id:      dataframe.NewColumn("ID", "%4s", dataframe.EmptyStrings),
		o:       dataframe.NewColumn("O", "%1d", dataframe.EmptyInts),
		rnr:     dataframe.NewColumn("Rnr", "%3s", dataframe.EmptyStrings),
		play:    dataframe.NewColumn("Play", "%30s", dataframe.EmptyStrings),
		after:   dataframe.NewColumn("After", "%5.1f", dataframe.EmptyFloats),
		before:  dataframe.NewColumn("Bfore", "%5.1f", dataframe.EmptyFloats),
		r:       dataframe.NewColumn("R", "%1d", dataframe.EmptyInts),
		re24:    dataframe.NewColumn("RE24", "% 6.1f", dataframe.EmptyFloats),
		runners: dataframe.NewColumn("Runners", "%-20s", dataframe.EmptyStrings),
	}
}

func (red *REData) GetData() *dataframe.Data {
	return &dataframe.Data{
		Name: "RE24",
		Columns: []*dataframe.Column{
			red.game, red.id, red.o, red.rnr, red.play, red.after, red.before,
			red.r, red.re24, red.runners,
		},
	}
}

func (red *REData) Record(gameID string, state, lastState *game.State, advances game.Advances) float64 {
	if red.re == nil {
		return 0
	}
	runsBefore := GetExpectedRuns(red.re, lastState)
	var runsAfter float64
	if state.Outs < 3 {
		runsAfter = GetExpectedRuns(red.re, state)
	}
	runsScored := len(state.ScoringRunners)
	change := runsAfter - runsBefore + float64(runsScored)
	var outs int
	if lastState != nil {
		outs = lastState.Outs
	}
	red.game.AppendString(gameID)
	red.id.AppendString(string(state.Batter))
	red.o.AppendInts(outs)
	red.rnr.AppendString(string(GetOccupiedBases(lastState)))
	red.play.AppendString(state.EventCode)
	red.after.AppendFloats(runsAfter)
	red.before.AppendFloats(runsBefore)
	red.r.AppendInts(runsScored)
	red.re24.AppendFloats(change)
	var runnersStrings []string
	for _, from := range []string{"B", "1", "2", "3"} {
		adv := advances[from]
		if adv != nil {
			runnersStrings = append(runnersStrings, string(adv.Runner))
		}
	}
	red.runners.AppendString(strings.Join(runnersStrings, " "))
	return change
}
