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

func (stats *Stats) GetBatting(batter game.PlayerID) *Batting {
	b := stats.Batting[batter]
	if b == nil {
		b = &Batting{
			Player: stats.players.GetPlayer(batter),
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
		}
		stats.Pitching[pitcher] = p
	}
	return p
}
