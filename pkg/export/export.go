package export

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/slshen/sb/pkg/tournament"
)

type Export struct {
	Us     string
	League string
	sheets *SheetExport
	re     stats.RunExpectancy
}

type StatsGenerator interface {
	Read(*game.Game) error
	GetData() *dataframe.Data
}

type REStatsGenerator struct {
	name string
	*stats.ObservedRunExpectancy
}

type GameStatsGenerator struct {
	read           func(*game.Game) error
	get            func() *dataframe.Data
	dataNameSuffix string
}

func NewExport(sheets *SheetExport, re stats.RunExpectancy) (*Export, error) {
	return &Export{
		sheets: sheets,
		re:     re,
	}, nil
}

func (export *Export) tournaments(games []*game.Game) ([]StatsGenerator, error) {
	gens := []StatsGenerator{}
	for _, gr := range tournament.GroupByTournament(games) {
		rep := &tournament.Report{
			Us:    export.Us,
			Group: gr,
		}
		if err := rep.Run(export.re); err != nil {
			return nil, err
		}
		gens = append(gens, &GameStatsGenerator{
			get: rep.GetBattingData,
		})
	}
	return gens, nil
}

func (export *Export) Export(games []*game.Game) error {
	gameStats := stats.NewGameStats(export.re)
	generators := []StatsGenerator{
		&REStatsGenerator{"RE", &stats.ObservedRunExpectancy{}},
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
			get: func() *dataframe.Data { return export.getUsBattingData(gameStats) },
		},
		&GameStatsGenerator{
			get: func() *dataframe.Data { return export.getUsPitchingData(gameStats) },
		},
	}
	if export.League == "" {
		tg, err := export.tournaments(games)
		if err != nil {
			return err
		}
		generators = append(generators, tg...)
	}
	if export.League != "" {
		var leagueGames []*game.Game
		for _, g := range games {
			if strings.HasPrefix(strings.ToLower(g.League), export.League) {
				leagueGames = append(leagueGames, g)
			}
		}
		games = leagueGames
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

func (export *Export) getUsBattingData(gs *stats.GameStats) *dataframe.Data {
	dat := gs.GetBattingData()
	idx := dat.GetIndex()
	return dat.RFilter(func(row int) bool {
		team := strings.ToLower(idx.GetValue(row, "Team").(string))
		return strings.HasPrefix(team, export.Us)
	})
}

func (export *Export) getUsPitchingData(gs *stats.GameStats) *dataframe.Data {
	dat := gs.GetPitchingData()
	idx := dat.GetIndex()
	return dat.RFilter(func(row int) bool {
		team := strings.ToLower(idx.GetValue(row, "Team").(string))
		return strings.HasPrefix(team, export.Us)
	})
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

func (gen *GameStatsGenerator) GetData() *dataframe.Data {
	data := gen.get()
	if gen.dataNameSuffix != "" {
		data.Name = fmt.Sprintf("%s%s", data.Name, gen.dataNameSuffix)
	}
	return data
}

func (gen *REStatsGenerator) GetData() *dataframe.Data {
	data := stats.GetRunExpectancyData(gen)
	data.Name = gen.name
	return data
}
