package stats

import (
	"github.com/slshen/sb/pkg/game"
)

type Batting struct {
	PlayerData                     `mapstructure:",squash"`
	PA, AB, Hits, Walks            int
	LineDrives                     int
	StrikeOuts                     int
	StrikeOutsLooking              int
	RunsScored                     int
	Singles, Doubles, Triples, HRs int
	StolenBases, CaughtStealing    int
	SB2, SB2PitchOpp, SB2Opp       int
	PickedOff                      int
	LOB                            int
	PitchesSeen, Swings, Misses    int
	Strikes, CalledStrikes         int
	GroundOuts, FlyOuts            int
	HitByPitch                     int
	OnBase                         int
	SacrificeBunts                 int
	SacrificeFlys                  int
	ReachedOnError                 int
	FieldersChoice                 int
	ReachedOnK                     int
	RE24                           float64
	LineDriveOuts                  int
	LOPH                           int
}

func (b *Batting) Update() {
	b.PlayerData.Update()
	b.LOPH = b.LineDriveOuts + b.Hits
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
		if state.Play.IsHit() {
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
			fallthrough
		case game.StrikeOutPickedOff:
			b.StrikeOuts++
			if state.Pitches.Last() == "C" {
				b.StrikeOutsLooking++
			}
		case game.WalkPickedOff:
			fallthrough
		case game.Walk:
			fallthrough
		case game.WalkWildPitch:
			fallthrough
		case game.WalkPassedBall:
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
		if state.IsAB() {
			b.AB++
		}
		if state.Modifiers.Trajectory() == game.LineDrive {
			b.LineDrives++
			if !state.Play.IsHit() {
				b.LineDriveOuts++
			}
		}
		if state.Play.IsHit() || state.Play.Is(game.Walk, game.WalkPickedOff, game.WalkWildPitch, game.WalkPassedBall, game.HitByPitch) {
			b.OnBase++
		}
		if state.Modifiers.Contains(game.SacrificeHit) {
			b.SacrificeBunts++
		}
		if state.Modifiers.Contains(game.SacrificeFly) {
			b.SacrificeFlys++
		}
	}
	if state.Complete || state.Incomplete {
		b.PitchesSeen += len(state.Pitches)
		b.Strikes += state.Pitches.CountUp('C', 'S', 'F')
		b.Swings += state.Pitches.CountUp('S', 'F')
		b.Misses += state.Pitches.CountUp('S')
		b.CalledStrikes += state.Pitches.CountUp('C')
		if state.Pitches.Last() == "X" {
			if state.Play.Type != game.HitByPitch {
				b.Strikes++
				b.Swings++
			}
		}
	}
	return
}
