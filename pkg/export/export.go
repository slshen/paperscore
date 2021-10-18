package export

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Export struct {
	Us     string
	League string
	sheets *SheetExport
}

type StatsGenerator interface {
	Read(*game.Game) error
	GetData() *stats.Data
}

type GameStatsGenerator struct {
	read           func(*game.Game) error
	get            func() *stats.Data
	dataNameSuffix string
}

func NewExport(sheets *SheetExport) (*Export, error) {
	return &Export{
		sheets: sheets,
	}, nil
}

func (export *Export) Export(games []*game.Game) error {
	re := &stats.RunExpectancy{Filter: stats.Filter{
		League: export.League,
		Team:   export.Us,
	}}
	if err := export.readGames(games, []StatsGenerator{re}); err != nil {
		return err
	}
	gameStats := stats.NewGameStats(re)
	gameStats.League = export.League
	gameStatsUs := stats.NewGameStats(re)
	gameStatsUs.League = export.League
	gameStatsUs.Team = export.Us
	generators := []StatsGenerator{
		&stats.RunExpectancy{
			Name:   "RE",
			Filter: stats.Filter{League: export.League},
		},
		&stats.RunExpectancy{
			Name:   "RE-Us",
			Filter: stats.Filter{League: export.League, Team: export.Us},
		},
		&stats.RunExpectancy{
			Name:   "RE-Them",
			Filter: stats.Filter{League: export.League, NotTeam: export.Us},
		},
		&GameStatsGenerator{
			read:           gameStats.Read,
			get:            gameStats.GetBattingData,
			dataNameSuffix: "-ALL",
		},
		&GameStatsGenerator{
			get:            gameStats.GetPitchingData,
			dataNameSuffix: "-ALL",
		},
		&GameStatsGenerator{
			read: gameStatsUs.Read,
			get:  gameStatsUs.GetBattingData,
		},
		&GameStatsGenerator{
			get: gameStatsUs.GetPitchingData,
		},
	}
	if err := export.readGames(games, generators); err != nil {
		return err
	}
	for _, gen := range generators {
		data := gen.GetData()
		if err := export.sheets.ExportData(data); err != nil {
			return err
		}
	}
	return nil
}

func (export *Export) readGames(games []*game.Game, generators []StatsGenerator) error {
	for _, g := range games {
		for _, gen := range generators {
			if err := gen.Read(g); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gen *GameStatsGenerator) Read(g *game.Game) error {
	if gen.read != nil {
		return gen.read(g)
	}
	return nil
}

func (gen *GameStatsGenerator) GetData() *stats.Data {
	data := gen.get()
	if gen.dataNameSuffix != "" {
		data.Name = fmt.Sprintf("%s%s", data.Name, gen.dataNameSuffix)
	}
	return data
}
