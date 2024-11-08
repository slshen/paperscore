package stats

import (
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type REData struct {
	re                                                       RunExpectancy
	game, bat, o, rnr, play, after, before, r, re24, runners *dataframe.Column
}

func NewREData(re RunExpectancy) *REData {
	return &REData{
		re:      re,
		game:    dataframe.NewColumn("Game", "%10s", dataframe.EmptyStrings),
		bat:     dataframe.NewColumn("Bat", "%4s", dataframe.EmptyStrings),
		o:       dataframe.NewColumn("O", "%1d", dataframe.EmptyInts),
		rnr:     dataframe.NewColumn("Rnr", "%3s", dataframe.EmptyStrings),
		play:    dataframe.NewColumn("Play", "%30s", dataframe.EmptyStrings),
		after:   dataframe.NewColumn("After", "%5.2f", dataframe.EmptyFloats),
		before:  dataframe.NewColumn("Bfore", "%5.2f", dataframe.EmptyFloats),
		r:       dataframe.NewColumn("R", "%1d", dataframe.EmptyInts),
		re24:    dataframe.NewColumn("RE24", "% 6.2f", dataframe.EmptyFloats),
		runners: dataframe.NewColumn("Runners", "%-20s", dataframe.EmptyStrings),
	}
}

func (red *REData) GetData() *dataframe.Data {
	return &dataframe.Data{
		Name: "RE24",
		Columns: []*dataframe.Column{
			red.game, red.bat, red.o, red.rnr, red.play, red.after, red.before,
			red.r, red.re24, red.runners,
		},
	}
}

func (red *REData) Record(gameID string, state *game.State) float64 {
	if red.re == nil {
		return 0
	}
	runsBefore, runsAfter, runsScored, change := GetExpectedRunsChange(red.re, state)
	var outs int
	if state.LastState != nil {
		outs = state.LastState.Outs
	}
	red.game.AppendString(gameID)
	red.bat.AppendString(string(state.Batter))
	red.o.AppendInt(outs)
	red.rnr.AppendString(string(GetOccupiedBases(state.LastState)))
	red.play.AppendString(state.GetPlayAdvancesCode())
	red.after.AppendFloat(runsAfter)
	red.before.AppendFloat(runsBefore)
	red.r.AppendInt(runsScored)
	red.re24.AppendFloat(change)
	var runnersStrings []string
	for _, from := range []string{"1", "2", "3"} {
		adv := state.Advances.From(from)
		if adv != nil {
			runnersStrings = append(runnersStrings, string(adv.Runner))
		}
	}
	red.runners.AppendString(strings.Join(runnersStrings, " "))
	return change
}
