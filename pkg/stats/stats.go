package stats

import "github.com/slshen/sb/pkg/game"

type Stats struct {
	Batting  map[game.PlayerID]*Batting
	Pitching map[game.PlayerID]*Pitching

	players PlayerLookup
}

type PlayerLookup interface {
	GetPlayer(game.PlayerID) *game.Player
}

func NewStats(players PlayerLookup) *Stats {
	return &Stats{
		players:  players,
		Batting:  make(map[game.PlayerID]*Batting),
		Pitching: make(map[game.PlayerID]*Pitching),
	}
}

func (stats *Stats) RecordBatting(g *game.Game, state *game.State) {
	batting := stats.GetBatting(state.Batter)
	batting.Record(state)
	batting.Games[g.ID] = true
	switch state.Play.Type {
	case game.CaughtStealing:
		if !state.NotOutOnPlay {
			runner := stats.GetBatting(state.Play.Runners[0])
			runner.CaughtStealing++
		}
	case game.StolenBase:
		for _, runnerID := range state.Play.Runners {
			runner := stats.GetBatting(runnerID)
			runner.StolenBases++
		}
	}
}

func (stats *Stats) GetBatting(batter game.PlayerID) *Batting {
	b := stats.Batting[batter]
	if b == nil {
		b = &Batting{
			Player: stats.players.GetPlayer(batter),
			Games:  map[string]bool{},
		}
		stats.Batting[batter] = b
	}
	return b
}

func (stats *Stats) GetPitching(pitcher game.PlayerID) *Pitching {
	p := stats.Pitching[pitcher]
	if p == nil {
		p = &Pitching{
			Player: stats.players.GetPlayer(pitcher),
			Games:  map[string]bool{},
		}
		stats.Pitching[pitcher] = p
	}
	return p
}

func (stats *Stats) RecordPitching(g *game.Game, state, lastState *game.State) {
	pitching := stats.GetPitching(state.Pitcher)
	pitching.Record(state, lastState)
	pitching.Games[g.ID] = true
}
