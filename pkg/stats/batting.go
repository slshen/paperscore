package stats

import "github.com/slshen/sb/pkg/game"

type Batting struct {
	Player                         *game.Player `yaml:"-"`
	AB, Runs, Hits /*RBI,*/, Walks int
	StrikeOuts                     int
	StrikeOutsLooking              int
	RunsScored                     int
	Singles, Doubles, Triples, HRs int
	StolenBases, CaughtStealing    int
	LOB                            int
	PitchesSeen, Swings, Misses    int
}

func (b *Batting) Record(state *game.State) (teamLOB int) {
	if state.Outs == 3 {
		lob := 0
		for _, runner := range state.Runners {
			if runner != "" {
				lob++
			}
		}
		if state.Complete {
			b.LOB += lob
		} else {
			teamLOB = lob
		}
	}
	if state.Complete {
		if state.Play.Hit() {
			b.Hits++
			if state.Play.Single() {
				b.Singles++
			}
			if state.Play.Double() {
				b.Doubles++
			}
			if state.Play.Triple() {
				b.Triples++
			}
			if state.Play.HomeRun() {
				b.HRs++
			}
		}
		if state.Play.StrikeOut() {
			b.StrikeOuts++
		}
		if state.Play.Walk() {
			b.Walks++
		}
		if !(state.Play.Walk() || state.Play.HitByPitch() ||
			state.Play.CatcherInterference() ||
			(state.Play.ReachedOnError() && state.Modifiers.Contains(game.Obstruction)) ||
			state.Modifiers.Contains(game.SacrificeFly, game.SacrificeHit)) {
			b.AB++
		}
		b.PitchesSeen = state.Pitches.Balls() + state.Pitches.Strikes()
		b.Swings = state.Pitches.Swings()
		b.Misses = state.Pitches.Misses()
	}
	return
}
