package dataexport

import (
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type Game struct {
	GameID       string
	GameNumber   string
	GameDate     string
	Tournament   string
	TournamentID string
	Home         string
	Visitor      string
	HomeScore    int
	VisitorScore int
}

type Games []*Game

func newGame(g *game.Game, tournamentID string) *Game {
	return &Game{
		GameID:       getGameID(g),
		GameNumber:   g.Number,
		GameDate:     g.GetDate().Format("2006-01-02"),
		Tournament:   g.Tournament,
		TournamentID: tournamentID,
		Home:         g.Home.Name,
		Visitor:      g.Visitor.Name,
		HomeScore:    g.Final.Home,
		VisitorScore: g.Final.Visitor,
	}
}

func (games Games) GetData() *dataframe.Data {
	dat := &dataframe.Data{}
	var idx *dataframe.Index
	for _, g := range games {
		idx = dat.AppendStruct(idx, g)
	}
	return dat
}
