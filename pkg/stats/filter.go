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
	teamName := g.Home
	if state.Top() {
		teamName = g.Visitor
	}
	teamName = strings.ToLower(teamName)
	if f.Team != "" {
		if !strings.HasPrefix(teamName, strings.ToLower(f.Team)) {
			return true
		}
	}
	if f.NotTeam != "" {
		if strings.HasPrefix(teamName, strings.ToLower(f.NotTeam)) {
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
