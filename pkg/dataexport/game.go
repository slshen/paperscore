package dataexport

import (
	"time"
)

type Game struct {
	ID      string
	Home    string
	Visitor string
	Date    time.Time
	Number  string
	Box     string
}

/*

func newGame(exp *DataExport, tid string, g *game.Game) (*Game, error) {
	report, err := tournament.NewReport(exp.us, exp.re, &tournament.Group{
		Date:  g.GetDate(),
		Name:  fmt.Sprintf("%s at %s", g.Visitor.Name, g.Home.Name),
		Games: []*game.Game{g},
	})
	if err != nil {
		return nil, err
	}
	box, err := boxscore.NewBoxScore(g, exp.re)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s at %s %s game %s", g.Visitor.Name, g.Home.Name, g.Date, g.Number)
	game := &Game{
		ID:      ToID(fmt.Sprintf("%s at %s %s %s", g.Visitor.Name, g.Home.Name, g.Date, g.Number)),
		Home:    g.Home.Name,
		Visitor: g.Visitor.Name,
		Date:    g.GetDate(),
		Number:  g.Number,
		Box:     box.InningScoreTable().String(),
	}
	exp.AddResource(&dataframe.DataResource{
		Description: fmt.Sprintf("Batting for %s", name),
		Path:        fmt.Sprintf("%s/%s/batting.csv", tid, game.ID),
		Data:        report.GetBattingData(),
	}, &dataframe.DataResource{
		Description: fmt.Sprintf("Pitching for %s", name),
		Path:        fmt.Sprintf("%s/%s/pitching.csv", tid, game.ID),
		Data:        report.GetPitchingData(),
	})
	/*
			Title:    fmt.Sprintf("%s #%s %s", g.Date, g.Number, them.Name),
			Batting:  report.GetBattingData(),
			Pitching: report.GetPitchingData(),
			Plays: report.GetRE24Data().Select(
				dataframe.DeriveInts("#", func(idx *dataframe.Index, i int) int {
					return i + 1
				}),
				dataframe.Col("Bat"),
				dataframe.Rename("O", "Outs"),
				dataframe.Col("Rnr"),
				dataframe.Col("RE24"),
				dataframe.Col("R"),
				dataframe.Col("Play"),
			),
			Alt: report.GetAltData().Select(
				dataframe.Col("RCost"),
				dataframe.Col("Bat"),
				dataframe.Rename("O", "Outs"),
				dataframe.Col("Rnr"),
				dataframe.Col("Reality"),
				dataframe.Col("Alternate"),
				dataframe.Col("Comment"),
			),
			box:  box,
			game: g,
		}
		game.Pitching.Name = "Pitching"
		game.Plays.Name = "Plays"
		game.Batting.Name = "Batting"
		game.Alt.Name = "Alt Plays"

	return game, nil
}
*/
