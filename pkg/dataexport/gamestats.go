package dataexport

import (
	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/stats"
)

type GameStats struct {
	re         stats.RunExpectancy
	battingDat *dataframe.Data
}

func newGameStats(re stats.RunExpectancy) *GameStats {
	return &GameStats{
		re: re,
	}
}

func (gs *GameStats) read(g *game.Game, tournamentID string) error {
	s := stats.NewGameStats(gs.re)
	if err := s.Read(g); err != nil {
		return err
	}
	gs.appendBattingData(g, tournamentID, s.GetAllBattingData())
	return nil
}

func (gs *GameStats) appendBattingData(g *game.Game, tournamentID string, dat *dataframe.Data) {
	gameID := getGameID(g)
	gameDate := toDate(g.GetDate())
	gIDs := make([]string, dat.RowCount())
	gDates := make([]string, dat.RowCount())
	tIDs := make([]string, dat.RowCount())
	for i := range gIDs {
		gIDs[i] = gameID
		gDates[i] = gameDate
		tIDs[i] = tournamentID
	}
	dat.Columns = append(dat.Columns, &dataframe.Column{
		Name:   "GameID",
		Values: gIDs,
	}, &dataframe.Column{
		Name:   "GameDate",
		Values: gDates,
	}, &dataframe.Column{
		Name:   "TournamentID",
		Values: tIDs,
	})
	if gs.battingDat == nil {
		gs.battingDat = dat
	} else {
		gs.battingDat.Append(dat)
	}
}
