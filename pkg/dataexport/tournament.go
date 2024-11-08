package dataexport

import (
	"time"

	"github.com/slshen/paperscore/pkg/tournament"
)

type Tournament struct {
	Name         string
	TournamentID string
	Date         string
	Wins         int
	Losses       int
	Ties         int

	group *tournament.Group
}

func newTournament(group *tournament.Group) *Tournament {
	return &Tournament{
		TournamentID: ToID(group.Name),
		Name:         group.Games[0].Tournament,
		Date:         group.Date.Format(time.RFC3339),
		group:        group,
	}
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
