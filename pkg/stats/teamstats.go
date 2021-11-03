package stats

import (
	"github.com/slshen/sb/pkg/game"
)

type TeamStats struct {
	Batting  map[game.PlayerID]*Batting
	Pitching map[game.PlayerID]*Pitching

	players PlayerLookup
}

type PlayerLookup interface {
	GetPlayer(game.PlayerID) *game.Player
}

func NewStats(players PlayerLookup) *TeamStats {
	return &TeamStats{
		players:  players,
		Batting:  make(map[game.PlayerID]*Batting),
		Pitching: make(map[game.PlayerID]*Pitching),
	}
}

func (stats *TeamStats) RecordBatting(g *game.Game, state, lastState *game.State, re RunExpectancy) {
	batting := stats.GetBatting(state.Batter)
	batting.Record(state)
	if re != nil {
		batting.RecordRE24(state, lastState, re)
	}
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
	if lastState != nil {
		// look for a lead runnerID on first
		var runnerID game.PlayerID
		if lastState.Runners[2] == "" && lastState.Runners[1] == "" {
			runnerID = lastState.Runners[0]
		}
		if runnerID != "" {
			// count SB2 and stolen base opportunties
			runner := stats.GetBatting(runnerID)
			if state.Play.Type == game.StolenBase {
				runner.SB2++
			}
			i := 0
			if !(lastState.Complete || lastState.Incomplete) {
				i = len(lastState.Pitches)
			}
			one := false
			for j, pitch := range state.Pitches[i:] {
				if pitch == 'S' || pitch == 'C' || pitch == 'B' {
					lastPitch := j == len(state.Pitches[i:])-1
					if lastPitch {
						if pitch == 'B' && state.Play.Type == game.Walk {
							continue
						}
						if (pitch == 'S' || pitch == 'C') && state.Play.Type == game.StrikeOut && state.Outs == 3 {
							continue
						}
					}
					if !one {
						runner.SB2Opp++
						one = true
					}
					runner.SB2PitchOpp++
				}
			}
		}
	}

	for _, runnerID := range state.ScoringRunners {
		runner := stats.GetBatting(runnerID)
		runner.RunsScored++
	}
	if re != nil {
		reChange := GetExpectedRuns(re, state) - GetExpectedRuns(re, lastState) + float64(len(state.ScoringRunners))
		if state.Play.Is(game.StolenBase, game.CaughtStealing, game.PickedOff) {
			perRunner := reChange / float64(len(state.Play.Runners))
			for _, runnerID := range state.Play.Runners {
				runner := stats.GetBatting(runnerID)
				runner.RE24 += perRunner
			}
		} else if state.Play.Is(game.WildPitch, game.PassedBall) {
			// this will give apportioned credit to all runnners, even
			// if some of them were out
			perRunner := reChange / float64(len(state.Advances))
			for _, advance := range state.Advances {
				runner := stats.GetBatting(advance.Runner)
				runner.RE24 += perRunner
			}
		}
	}
}

func (stats *TeamStats) GetBatting(batter game.PlayerID) *Batting {
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

func (stats *TeamStats) GetPitching(pitcher game.PlayerID) *Pitching {
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

func (stats *TeamStats) RecordPitching(g *game.Game, state, lastState *game.State) {
	pitching := stats.GetPitching(state.Pitcher)
	pitching.Record(state, lastState)
	pitching.Games[g.ID] = true
}
