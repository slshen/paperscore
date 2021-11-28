package stats

import (
	"fmt"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type TeamStats struct {
	Batting          map[game.PlayerID]*Batting
	Pitching         map[game.PlayerID]*Pitching
	LOB              int
	Batters          []game.PlayerID
	Pitchers         []game.PlayerID
	Errors           int
	ErrorsByPosition []int `yaml:",flow"`
	Team             *game.Team

	re RunExpectancy
}

func NewStats(team *game.Team, re RunExpectancy) *TeamStats {
	return &TeamStats{
		Team:     team,
		re:       re,
		Batting:  make(map[game.PlayerID]*Batting),
		Pitching: make(map[game.PlayerID]*Pitching),
	}
}

func (stats *TeamStats) GetPitchingData() *dataframe.Data {
	dat := newData("PIT")
	var idx *dataframe.Index
	for _, player := range stats.Pitchers {
		pitching := stats.Pitching[player]
		var err error
		idx, err = dat.AppendStruct(idx, pitching)
		if err != nil {
			panic(err)
		}
	}
	return dat
}

func (stats *TeamStats) recordError(e *game.FieldingError) {
	for len(stats.ErrorsByPosition) < e.Fielder {
		stats.ErrorsByPosition = append(stats.ErrorsByPosition, 0)
	}
	stats.ErrorsByPosition[e.Fielder-1]++
	stats.Errors++
}

func (stats *TeamStats) GetBattingData() *dataframe.Data {
	dat := newData("BAT")
	var idx *dataframe.Index
	for _, player := range stats.Batters {
		batting := stats.Batting[player]
		var err error
		idx, err = dat.AppendStruct(idx, batting)
		if err != nil {
			panic(err)
		}
	}
	return dat
}

func (stats *TeamStats) RecordBatting(g *game.Game, state, lastState *game.State) {
	batting := stats.GetBatting(state.Batter)
	stats.LOB += batting.Record(state)
	if stats.re != nil {
		batting.RecordRE24(state, lastState, stats.re)
	}
	batting.GameAppearances[g.ID] = true
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
	if stats.re != nil {
		var afterRuns float64
		if state.Outs != 3 {
			afterRuns = GetExpectedRuns(stats.re, state)
		}
		reChange := afterRuns - GetExpectedRuns(stats.re, lastState) + float64(len(state.ScoringRunners))
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
			PlayerData: NewPlayerData(stats.Team.Name, stats.Team.GetPlayer(batter)),
		}
		stats.Batting[batter] = b
		stats.Batters = append(stats.Batters, batter)
		if len(stats.Team.Players) > 0 && stats.Team.Players[batter] == nil {
			fmt.Printf("batter %s does not have a team entry\n", batter)
		}
	}
	return b
}

func (stats *TeamStats) GetPitching(pitcher game.PlayerID) *Pitching {
	p := stats.Pitching[pitcher]
	if p == nil {
		p = &Pitching{
			PlayerData: NewPlayerData(stats.Team.Name, stats.Team.GetPlayer(pitcher)),
		}
		stats.Pitching[pitcher] = p
		stats.Pitchers = append(stats.Pitchers, pitcher)
	}
	return p
}

func (stats *TeamStats) RecordFielding(g *game.Game, state, lastState *game.State) {
	pitching := stats.GetPitching(state.Pitcher)
	pitching.Record(state, lastState)
	pitching.GameAppearances[g.ID] = true
	switch state.Play.Type {
	case game.ReachedOnError:
		stats.recordError(state.Play.FieldingError)
	case game.PickedOff:
		fallthrough
	case game.CaughtStealing:
		if state.NotOutOnPlay && state.Play.FieldingError != nil {
			stats.recordError(state.Play.FieldingError)
		}
	case game.FoulFlyError:
		stats.recordError(state.Play.FieldingError)
	}
	for _, adv := range state.Advances {
		if adv.FieldingError != nil {
			stats.recordError(adv.FieldingError)
		}
	}
}
