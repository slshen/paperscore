package dataexport

import (
	"github.com/slshen/sb/pkg/boxscore"
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type Game struct {
	GameID        string
	GameNumber    string
	GameDate      string
	Tournament    string
	TournamentID  string
	Home          string
	Visitor       string
	HomeScore     int
	HomeHits      int
	HomeErrors    int
	VisitorScore  int
	VisitorHits   int
	VisitorErrors int
}

type Games []*Game

func newGame(g *game.Game, tournamentID string) (*Game, error) {
	box, err := boxscore.NewBoxScore(g, nil)
	if err != nil {
		return nil, err
	}
	return &Game{
		GameID:        getGameID(g),
		GameNumber:    g.Number,
		GameDate:      g.GetDate().Format("2006-01-02"),
		Tournament:    g.Tournament,
		TournamentID:  tournamentID,
		Home:          g.Home.Name,
		HomeScore:     g.Final.Home,
		HomeHits:      box.HomeLineup.TotalHits(),
		HomeErrors:    box.HomeLineup.Errors,
		Visitor:       g.Visitor.Name,
		VisitorScore:  g.Final.Visitor,
		VisitorHits:   box.VisitorLineup.TotalHits(),
		VisitorErrors: box.VisitorLineup.Errors,
	}, nil
}

func (games Games) GetData() *dataframe.Data {
	dat := &dataframe.Data{}
	var idx *dataframe.Index
	for _, g := range games {
		idx = dat.AppendStruct(idx, g)
	}
	return dat
}
