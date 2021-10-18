package stats

import (
	"strings"

	"github.com/slshen/sb/pkg/game"
)

type Filter struct {
	Team, NotTeam     string
	League, NotLeague string
}

func (f *Filter) filterOut(g *game.Game, state *game.State) bool {
	team := g.Home
	if state.Top() {
		team = g.Visitor
	}
	team = strings.ToLower(team)
	if f.Team != "" {
		if !strings.HasPrefix(team, strings.ToLower(f.Team)) {
			return true
		}
	}
	if f.NotTeam != "" {
		if strings.HasPrefix(team, strings.ToLower(f.NotTeam)) {
			return true
		}
	}
	league := strings.ToLower(g.League)
	if f.League != "" {
		if !strings.HasPrefix(league, strings.ToLower(f.League)) {
			return true
		}
	}
	if f.NotLeague != "" {
		if strings.HasPrefix(league, strings.ToLower(f.NotLeague)) {
			return true
		}
	}
	return false
}
