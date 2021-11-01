package export

import (
	"fmt"
	"strings"

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
	re := &stats.RunExpectancy{}
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
			get: func() *stats.Data { return export.getUsPitchingData(gameStats) },
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

func (export *Export) getUsPitchingData(gs *stats.GameStats) *stats.Data {
	alldata := gs.GetPitchingData()
	data := &stats.Data{
		Name:    "PIT",
		Columns: alldata.Columns,
		Width:   alldata.Width,
	}
	for _, row := range alldata.Rows {
		name := row[0].(string)
		if strings.HasPrefix(strings.ToLower(name), export.Us) {
			slash := strings.Index(name, "/")
			row2 := make([]interface{}, len(row))
			for i := range row {
				if i == 0 {
					row2[0] = name[slash+1:]
				} else {
					row2[i] = row[i]
				}
			}
			data.Rows = append(data.Rows, row2)
		}
	}
	return data
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
