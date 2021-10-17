package export

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Export struct {
	Us     string
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
	gameStats := stats.NewGameStats()
	gameStatsUs := stats.NewGameStats()
	gameStatsUs.OnlyTeam = export.Us
	generators := []StatsGenerator{
		&stats.RunExpectancy{
			Name: "RE",
		},
		&stats.RunExpectancy{
			Name: "RE-Us",
			Team: export.Us,
		},
		&stats.RunExpectancy{
			Name:    "RE-Them",
			NotTeam: export.Us,
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
	for _, g := range games {
		for _, gen := range generators {
			if err := gen.Read(g); err != nil {
				return err
			}
		}
	}
	for _, gen := range generators {
		data := gen.GetData()
		if err := export.sheets.ExportData(data); err != nil {
			return err
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
