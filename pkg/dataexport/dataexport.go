package dataexport

import (
	"regexp"
	"sort"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/dataframe/pkg"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/slshen/sb/pkg/tournament"
)

type DataExport struct {
	us string
	re stats.RunExpectancy
}

func NewDataExport(us string, re stats.RunExpectancy) *DataExport {
	exp := &DataExport{
		us: us,
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
	var events Events
	for _, g := range games {
		evs, err := GetEvents(g)
		if err != nil {
			return nil, err
		}
		events = append(events, evs...)
	}
	dp.AddResource(&pkg.DataResource{
		Path:        "events.csv",
		Description: "All events",
		Data:        events.GetData(),
	})
	dp.AddResource(&pkg.DataResource{
		Path:        "advances.csv",
		Description: "Advances",
		Data:        events.GetAdvancesData(),
	})
	groups := tournament.GroupByTournament(games)
	sort.Slice(groups, func(i, j int) bool {
		return groups[j].Date.Before(groups[i].Date)
	})
	var idx *dataframe.Index
	for _, group := range groups {
		tourney := newTournament(group)
		idx = tournaments.AppendStruct(idx, tourney)
		res, err := tourney.getResources(exp)
		if err != nil {
			return nil, err
		}
		dp.AddResource(res...)
	}
	tournaments.Arrange("ID", "Name", "Date", "Wins", "Losses", "Ties")
	dp.AddResource(tournaments)
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
