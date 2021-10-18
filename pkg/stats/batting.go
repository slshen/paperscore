package stats

import "github.com/slshen/sb/pkg/game"

type Batting struct {
	Player                         *game.Player `yaml:"-" mapstructure:",squash"`
	PA, AB, Hits, Walks            int
	LineDrives                     int
	StrikeOuts                     int
	StrikeOutsLooking              int
	RunsScored                     int
	Singles, Doubles, Triples, HRs int
	StolenBases, CaughtStealing    int
	PickedOff                      int
	LOB                            int
	PitchesSeen, Swings, Misses    int
	GroundOuts, FlyOuts            int
	HitByPitch                     int
	OnBase                         int
	SacrificeBunts                 int
	SacrificeFlys                  int
	ReachedOnError                 int
	FieldersChoice                 int
	ReachedOnK                     int
	Games                          map[string]bool
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
		b.PA++
		if state.Play.Hit() {
			b.Hits++
		}
		switch state.Play.Type {
		case game.Single:
			b.Singles++
		case game.Double:
			b.Doubles++
		case game.Triple:
			b.Triples++
		case game.HomeRun:
			b.HRs++
		case game.StrikeOut:
			b.StrikeOuts++
			if state.Pitches.Last() == "C" {
				b.StrikeOutsLooking++
			}
		case game.Walk:
			b.Walks++
		case game.GroundOut:
			b.GroundOuts++
		case game.FlyOut:
			b.FlyOuts++
		case game.HitByPitch:
			b.HitByPitch++
		case game.ReachedOnError:
			b.ReachedOnError++
		case game.FieldersChoice:
			b.FieldersChoice++
		case game.StrikeOutPassedBall:
			fallthrough
		case game.StrikeOutWildPitch:
			b.ReachedOnK++
		}
		if !(state.Play.Is(game.Walk, game.HitByPitch, game.CatcherInterference) ||
			(state.Play.Type == game.ReachedOnError && state.Modifiers.Contains(game.Obstruction)) ||
			state.Modifiers.Contains(game.SacrificeFly, game.SacrificeHit)) {
			b.AB++
		}
		if state.Modifiers.Trajectory() == game.LineDrive {
			b.LineDrives++
		}
		if state.Play.Hit() || state.Play.Is(game.Walk, game.HitByPitch) {
			b.OnBase++
		}
		if state.Modifiers.Contains(game.SacrificeHit) {
			b.SacrificeBunts++
		}
		if state.Modifiers.Contains(game.SacrificeFly) {
			b.SacrificeFlys++
		}
	}
	b.PitchesSeen += state.Pitches.Balls() + state.Pitches.Strikes()
	b.Swings += state.Pitches.Swings()
	b.Misses += state.Pitches.Misses()
	return
}
