package stats

import (
	"fmt"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type TeamStats struct {
	Batting  map[game.PlayerID]*Batting
	Pitching map[game.PlayerID]*Pitching
	*FieldingStats
	LOB      int
	Batters  []game.PlayerID
	Pitchers []game.PlayerID
	Team     *game.Team
}

func NewStats(team *game.Team, re RunExpectancy) *TeamStats {
	return &TeamStats{
		Team:          team,
		FieldingStats: newFieldingStats(),
		Batting:       make(map[game.PlayerID]*Batting),
		Pitching:      make(map[game.PlayerID]*Pitching),
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
	dat.Add(
		dataframe.DeriveFloats("Slugging", Slugging),
		dataframe.DeriveFloats("OnBasePct", OnBase),
		dataframe.DeriveFloats("OPS", OPS),
	)
	return dat
}

func (stats *TeamStats) RecordBatting(g *game.Game, state *game.State, reChange float64) {
	batting := stats.GetBatting(state.Batter)
	stats.LOB += batting.Record(state)
	if state.Complete {
		batting.RE24 += reChange
	}
	batting.GameAppearances[g.ID] = true
	switch state.Play.Type {
	case game.CaughtStealing:
		if !state.NotOutOnPlay {
			runner := stats.GetBatting(state.Play.Runners[0])
			runner.CaughtStealing++
		}
	case game.StolenBase:
		// TODO account for K+SB
		for _, runnerID := range state.Play.Runners {
			runner := stats.GetBatting(runnerID)
			runner.StolenBases++
		}
	}
	if state.LastState != nil {
		// look for a lead runnerID on first
		var runnerID game.PlayerID
		if state.LastState.Runners[2] == "" && state.LastState.Runners[1] == "" {
			runnerID = state.LastState.Runners[0]
		}
		if runnerID != "" {
			// count SB2 and stolen base opportunties
			runner := stats.GetBatting(runnerID)
			if state.Play.Type == game.StolenBase {
				runner.SB2++
			}
			i := 0
			if !(state.LastState.Complete || state.LastState.Incomplete) {
				i = len(state.LastState.Pitches)
			}
			one := false
			for j, pitch := range state.Pitches[i:] {
				if pitch == 'S' || pitch == 'C' || pitch == 'B' {
					lastPitch := j == len(state.Pitches[i:])-1
					if lastPitch {
						if pitch == 'B' && state.Play.Type == game.Walk {
							continue
						}
						if (pitch == 'S' || pitch == 'C') &&
							(state.Play.Type == game.StrikeOut || state.Play.Type == game.StrikeOutPickedOff) &&
							state.Outs == 3 {
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
	if reChange != 0 {
		var runners []game.PlayerID
		if state.Play.Is(game.StolenBase, game.CaughtStealing, game.PickedOff, game.WalkPickedOff) {
			runners = state.Play.Runners
		} else if state.Play.Is(game.WildPitch, game.PassedBall) {
			// this will give apportioned credit to all runnners, even
			// if some of them were out
			for _, advance := range state.Advances {
				runners = append(runners, advance.Runner)
			}
		}
		if len(runners) > 0 {
			perRunner := reChange / float64(len(runners))
			for _, runnerID := range runners {
				runner := stats.GetBatting(runnerID)
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

func (stats *TeamStats) RecordFielding(g *game.Game, state *game.State) {
	pitching := stats.GetPitching(state.Pitcher)
	pitching.Record(state)
	pitching.GameAppearances[g.ID] = true
	switch state.Play.Type {
	case game.ReachedOnError:
		stats.recordError(state.Play.FieldingError)
	case game.WalkPickedOff:
		fallthrough
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
