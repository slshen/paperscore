package dataexport

import (
	"regexp"
	"sort"
	"time"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/dataframe/pkg"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/stats"
	"github.com/slshen/paperscore/pkg/tournament"
)

type DataExport struct {
	re stats.RunExpectancy
}

func NewDataExport(re stats.RunExpectancy) *DataExport {
	exp := &DataExport{
		re: re,
	}
	return exp
}

func (exp *DataExport) Read(games []*game.Game) (*pkg.DataPackage, error) {
	dp := &pkg.DataPackage{}
	tournaments := &pkg.DataResource{
		Path:        "tournaments.csv",
		Description: "Softball tournaments",
		Data:        &dataframe.Data{},
	}
	tournamentIDS := map[*game.Game]string{}
	groups := tournament.GroupByTournament(games)
	sort.Slice(groups, func(i, j int) bool {
		return groups[j].Date.Before(groups[i].Date)
	})
	var idx *dataframe.Index
	for _, group := range groups {
		tourney := newTournament(group)
		idx = tournaments.AppendStruct(idx, tourney)
		for _, g := range group.Games {
			tournamentIDS[g] = tourney.TournamentID
		}
	}
	tournaments.Arrange("TournamentID", "Name", "Date", "Wins", "Losses", "Ties")
	dp.AddResource(tournaments)
	var events, alts Events
	var gms Games
	gs := newGameStats(exp.re)
	for _, g := range games {
		evs, as := GetEvents(exp.re, g, tournamentIDS[g])
		events = append(events, evs...)
		alts = append(alts, as...)
		gm, err := newGame(g, tournamentIDS[g])
		if err != nil {
			return nil, err
		}
		gms = append(gms, gm)
		if err := gs.read(g, tournamentIDS[g]); err != nil {
			return nil, err
		}
	}
	dp.AddResource(&pkg.DataResource{
		Path:        "games.csv",
		Description: "Game summary",
		Data:        gms.GetData(),
	})
	dp.AddResource(&pkg.DataResource{
		Path:        "batting.csv",
		Description: "Game by game batting stats",
		Data:        gs.battingDat,
	})
	dp.AddResource(&pkg.DataResource{
		Path:        "events.csv",
		Description: "All events",
		Data:        events.GetData(),
	})
	dp.AddResource(&pkg.DataResource{
		Path:        "alt_events.csv",
		Description: "Alternative events",
		Data:        alts.GetData(),
	})
	dp.AddResource(&pkg.DataResource{
		Path:        "advances.csv",
		Description: "Advances",
		Data:        events.GetAdvancesData(),
	})
	dp.AddResource(&pkg.DataResource{
		Description: "Run expectancy data used",
		Path:        "run-expectancy.csv",
		Data:        stats.GetRunExpectancyData(exp.re),
	})
	ore := &stats.ObservedRunExpectancy{}
	for _, g := range games {
		if err := ore.Read(g); err != nil {
			return nil, err
		}
	}
	dp.AddResource(&pkg.DataResource{
		Description: "Observed run expectancy",
		Path:        "observed-re.csv",
		Data:        ore.GetRunData(),
	})
	dp.AddResource(&pkg.DataResource{
		Description: "Observed run frequency",
		Path:        "observed-rf.csv",
		Data:        &ore.GetRunExpectancyFrequency().Data,
	})
	return dp, nil
}

var nameIDRe = regexp.MustCompile(`[-/\\ ]`)

func ToID(s string) string {
	return nameIDRe.ReplaceAllLiteralString(s, "-")
}

func toDate(d time.Time) string {
	return d.Format("2006-01-02")
}
