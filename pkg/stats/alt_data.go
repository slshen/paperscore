package stats

import (
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type AltData struct {
	re                                                        RunExpectancy
	game, bat, o, rnr, play, alt, ore24, are24, cost, comment *dataframe.Column
}

func NewAltData(re RunExpectancy) *AltData {
	alt := &AltData{
		re:      re,
		game:    dataframe.NewColumn("Game", "%10s", dataframe.EmptyStrings),
		bat:     dataframe.NewColumn("Bat", "%4s", dataframe.EmptyStrings),
		o:       dataframe.NewColumn("O", "%1d", dataframe.EmptyInts),
		rnr:     dataframe.NewColumn("Rnr", "%3s", dataframe.EmptyStrings),
		play:    dataframe.NewColumn("Play", "%30s", dataframe.EmptyStrings),
		alt:     dataframe.NewColumn("Alt", "%30s", dataframe.EmptyStrings),
		ore24:   dataframe.NewColumn("ORE24", "% 6.2f", dataframe.EmptyFloats),
		are24:   dataframe.NewColumn("ARE24", "% 6.2f", dataframe.EmptyFloats),
		cost:    dataframe.NewColumn("COST", "%6.2f", dataframe.EmptyFloats),
		comment: dataframe.NewColumn("COM", "%-15s", dataframe.EmptyStrings),
	}
	alt.cost.Summary = dataframe.Sum
	return alt
}

func (alt *AltData) GetData() *dataframe.Data {
	dat := &dataframe.Data{
		Name: "ALTRE24",
		Columns: []*dataframe.Column{
			alt.game, alt.bat, alt.o, alt.rnr, alt.play, alt.ore24,
			alt.alt, alt.are24, alt.cost, alt.comment,
		},
	}
	return dat.RSort(dataframe.Less(dataframe.CompareFloat(alt.cost)))
}

func (alt *AltData) Record(gameID string, state *game.State) float64 {
	if alt.re == nil || state.AlternativeFor == nil {
		return 0
	}
	_, _, _, change := alt.getREChange(state)
	var outs int
	if state.LastState != nil {
		outs = state.LastState.Outs
	}
	_, _, _, originalChange := alt.getREChange(state.AlternativeFor)
	alt.game.AppendString(gameID)
	alt.bat.AppendString(string(state.Batter))
	alt.o.AppendInts(outs)
	alt.rnr.AppendString(string(GetOccupiedBases(state.LastState)))
	alt.play.AppendString(state.AlternativeFor.GetPlayAdvancesCode())
	alt.alt.AppendString(state.GetPlayAdvancesCode())
	alt.ore24.AppendFloats(originalChange)
	alt.are24.AppendFloats(change)
	cost := change - originalChange
	if state.Batter.IsUs() {
		cost = -cost
	}
	alt.cost.AppendFloats(cost)
	alt.comment.AppendString(state.Comment)
	return change
}

func (alt *AltData) getREChange(state *game.State) (runsBefore float64, runsAfter float64, runsScored int, change float64) {
	runsBefore = GetExpectedRuns(alt.re, state.LastState)
	if state.Outs < 3 {
		runsAfter = GetExpectedRuns(alt.re, state)
	}
	runsScored = len(state.ScoringRunners)
	change = runsAfter - runsBefore + float64(runsScored)
	return
}
