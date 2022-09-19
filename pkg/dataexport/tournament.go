package dataexport

import (
	"fmt"
	"time"

	"github.com/slshen/sb/pkg/dataframe/pkg"
	"github.com/slshen/sb/pkg/tournament"
)

type Tournament struct {
	Name   string
	ID     string
	Date   string
	Wins   int
	Losses int
	Ties   int

	group *tournament.Group
}

func newTournament(group *tournament.Group) *Tournament {
	return &Tournament{
		ID:    ToID(group.Name),
		Name:  group.Games[0].Tournament,
		Date:  group.Date.Format(time.RFC3339),
		group: group,
	}
}

func (t *Tournament) getResources(exp *DataExport) ([]pkg.Resource, error) {
	report, err := tournament.NewReport(exp.us, exp.re, t.group)
	if err != nil {
		return nil, err
	}
	res := []pkg.Resource{
		&pkg.DataResource{
			Description: fmt.Sprintf("Batting stats for %s", t.Name),
			Path:        fmt.Sprintf("%s/batting.csv", t.ID),
			Data:        report.GetBattingData(),
		},
		&pkg.DataResource{
			Description: fmt.Sprintf("Plays by RE24 for %s", t.Name),
			Path:        fmt.Sprintf("%s/plays.csv", t.ID),
			Data:        report.GetRE24Data(),
		},
		&pkg.DataResource{
			Description: fmt.Sprintf("Alternate plays for %s", t.Name),
			Path:        fmt.Sprintf("%s/alt.csv", t.ID),
			Data:        report.GetAltData(),
		},
	}
	return res, nil
}

/*
func (t *Tournament) readGames(exp *DataExport, group *tournament.Group) error {
	var idx *dataframe.Index
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			{Name: "Win", Values: dataframe.EmptyInts},
			{Name: "Loss", Values: dataframe.EmptyInts},
			{Name: "Ties", Values: dataframe.EmptyInts},
		},
	}
	for _, g := range group.Games {
		us, _ := g.GetUsAndThem(exp.us)
		var win, loss, tie int
		if us == g.Home {
			switch {
			case g.Final.Home > g.Final.Visitor:
				win = 1
			case g.Final.Home == g.Final.Visitor:
				tie = 1
			default:
				loss = 1
			}
		} else {
			switch {
			case g.Final.Visitor > g.Final.Home:
				win = 1
			case g.Final.Visitor == g.Final.Home:
				tie = 1
			default:
				loss = 1
			}
		}
		dat.Columns[0].AppendInt(win)
		dat.Columns[1].AppendInt(loss)
		dat.Columns[2].AppendInt(tie)
		if win == 1 {
			t.Wins++
		}
		if loss == 1 {
			t.Losses++
		}
		if tie == 1 {
			t.Ties++
		}
		game, err := newGame(exp, t.ID, g)
		if err != nil {
			return err
		}
		idx = dat.MustAppendStruct(idx, game)
	}
	exp.AddResource(&dataframe.DataResource{
		Description: fmt.Sprintf("Games for %s", t.Name),
		Path:        fmt.Sprintf("%s/games.csv", t.ID),
		Data:        dat,
	})
	return nil
}
*/
