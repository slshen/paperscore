package stats

import (
	"maps"
	"slices"
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type AltData struct {
	re RunExpectancy
	game, half, inn, bat, o, rnr, play, alt,
	cost, comment, credit *dataframe.Column
}

func NewAltData(re RunExpectancy) *AltData {
	alt := &AltData{
		re:      re,
		game:    dataframe.NewColumn("Game", "%10s", dataframe.EmptyStrings),
		half:    dataframe.NewColumn("H", "%3s", dataframe.EmptyStrings),
		inn:     dataframe.NewColumn("I", "%1d", dataframe.EmptyInts),
		bat:     dataframe.NewColumn("Bat", "%4s", dataframe.EmptyStrings),
		o:       dataframe.NewColumn("O", "%1d", dataframe.EmptyInts),
		rnr:     dataframe.NewColumn("Rnr", "%3s", dataframe.EmptyStrings),
		play:    dataframe.NewColumn("Reality", "%30s", dataframe.EmptyStrings),
		alt:     dataframe.NewColumn("Alternate", "%30s", dataframe.EmptyStrings),
		cost:    dataframe.NewColumn("RCost", "%6.2f", dataframe.EmptyFloats),
		comment: dataframe.NewColumn("Comment", "%-20s", dataframe.EmptyStrings),
		credit:  dataframe.NewColumn("Credit", "%s", dataframe.EmptyStrings),
	}
	alt.cost.Summary = dataframe.Sum
	return alt
}

func (alt *AltData) GetData() *dataframe.Data {
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			alt.game, alt.inn, alt.half, alt.bat, alt.o, alt.rnr, alt.play,
			alt.alt, alt.cost, alt.comment, alt.credit,
		},
	}
	return dat.RSort(dataframe.Less(dataframe.Descending(dataframe.CompareFloat(alt.cost))))
}

func (alt *AltData) GetPerPlayerData() *dataframe.Data {
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			dataframe.NewColumn("Player", "%-6s", dataframe.EmptyStrings),
			dataframe.NewColumn("Plays", "%-20s", dataframe.EmptyStrings),
			dataframe.NewColumn("RCost", "%.2f", dataframe.EmptyFloats),
		},
	}
	for i, val := range alt.credit.GetStrings() {
		if val != "" {
			players := strings.Fields(val)
			share := alt.cost.GetFloat(i) / float64(len(players))
			for _, player := range players {
				dat.Columns[0].AppendString(player)
				dat.Columns[1].AppendString(alt.play.GetString(i))
				dat.Columns[2].AppendFloat(share)
			}
		}
	}
	res := dat.GroupBy("Player").Aggregate(
		dataframe.ASum("Cost", dat.Columns[2]).WithFormat("%4.2f").WithSummary(dataframe.Sum),
		dataframe.AFunc("Plays", dataframe.String, func(acol *dataframe.Column, group *dataframe.Group) {
			plays := &strings.Builder{}
			for _, row := range group.Rows {
				play := alt.play.GetString(row)
				if plays.Len() > 0 {
					plays.WriteString(", ")
				}
				plays.WriteString(play)
			}
			acol.AppendString(plays.String())
		}),
	)
	res = res.RSort(dataframe.Less(dataframe.Descending(dataframe.CompareFloat(res.Columns[1]))))
	res.Arrange("Cost", "Player", "Plays")
	return res
}

func (alt *AltData) Record(gameID string, state *game.State) float64 {
	if alt.re == nil || state.AlternativeFor == nil {
		return 0
	}
	if state.Half == game.Top {
		alt.half.AppendString("TOP")
	} else {
		alt.half.AppendString("BOT")
	}
	alt.inn.AppendInt(state.InningNumber)
	_, _, _, change := GetExpectedRunsChange(alt.re, state)
	var outs int
	if state.LastState != nil {
		outs = state.LastState.Outs
	}
	_, _, _, originalChange := GetExpectedRunsChange(alt.re, state.AlternativeFor)
	alt.game.AppendString(gameID)
	alt.bat.AppendString(string(state.Batter))
	alt.o.AppendInt(outs)
	alt.rnr.AppendString(string(GetOccupiedBases(state.LastState)))
	alt.play.AppendString(state.AlternativeFor.GetPlayAdvancesCode())
	alt.alt.AppendString(state.GetPlayAdvancesCode())
	price := originalChange - change
	if state.BattingTeam.Us {
		price = -price
	}
	alt.cost.AppendFloat(price)
	alt.comment.AppendString(state.Comment)
	credit := &strings.Builder{}
	for _, p := range getAltCredit(state) {
		if credit.Len() > 0 {
			credit.WriteRune(' ')
		}
		credit.WriteString(string(p))
	}
	alt.credit.AppendString(credit.String())
	return change
}

func getAltCredit(alt *game.State) []game.PlayerID {
	credits := map[game.PlayerID]bool{}
	for _, p := range alt.AlternativeCredits {
		credits[p.PlayerID] = true
	}
	state := alt.AlternativeFor
	if state.FieldingError.IsFieldingError() {
		player := state.Defense[state.FieldingError.Fielder-1]
		if player != "" {
			credits[player] = true
		}
	}
	if state.Play.Is(game.PassedBall) {
		catcher := state.Defense[1]
		if catcher != "" {
			credits[catcher] = true
		}
	}
	if state.Play.Is(game.WildPitch) {
		pitcher := state.Defense[0]
		if pitcher != "" {
			credits[pitcher] = true
		}
	}
	for _, adv := range state.Advances {
		if adv.IsFieldingError() {
			fielder := state.Defense[adv.FieldingError.Fielder-1]
			if fielder != "" {
				credits[fielder] = true
			}
		}
	}
	return slices.Collect(maps.Keys(credits))
}
