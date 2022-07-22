package webdata

import (
	"regexp"
	"sort"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/slshen/sb/pkg/tournament"
)

type WebData struct {
	Tournaments []*Tournament

	us string
	re stats.RunExpectancy
}

func NewWebData(us string, re stats.RunExpectancy, games []*game.Game) (*WebData, error) {
	wdat := &WebData{
		us: us,
		re: re,
	}
	groups := tournament.GroupByTournament(games)
	sort.Slice(groups, func(i, j int) bool {
		return groups[j].Date.Before(groups[i].Date)
	})
	for _, group := range groups {
		tourney, err := newTournament(wdat, group)
		if err != nil {
			return nil, err
		}
		wdat.Tournaments = append(wdat.Tournaments, tourney)
	}
	return wdat, nil
}

var nameIDRe = regexp.MustCompile(`[-/\\ ]`)

func ToID(s string) string {
	return nameIDRe.ReplaceAllLiteralString(s, "-")
}

func (wdat *WebData) Write(dir string) error {
	for i, tourney := range wdat.Tournaments {
		p := tourney.GetPage()
		p.Set("weight", i+1)
		if err := p.WriteFiles(dir); err != nil {
			return err
		}
	}
	return nil
}
