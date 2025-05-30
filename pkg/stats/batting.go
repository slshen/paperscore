package stats

import (
	"github.com/slshen/paperscore/pkg/game"
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
	GroundOuts, FlyOuts, PopOuts   int
	GIDP                           int
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
	FoulBunts                      int
	MissedBunts                    int
	PopupBunts                     int
	BuntHits                       int
	BuntSacrifices                 int
	BuntOuts                       int
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
			fallthrough
		case game.GroundRuleDouble:
			b.Doubles++
		case game.Triple:
			b.Triples++
		case game.HomeRun:
			b.HRs++
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
			if state.Trajectory() == game.PopUp {
				b.PopOuts++
			} else {
				b.FlyOuts++
			}
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
		if state.IsStrikeOut() {
			b.StrikeOuts++
			if state.Pitches.Last() == 'C' {
				b.StrikeOutsLooking++
			}
		}
		if state.IsAB() {
			b.AB++
		}
		trajectory := state.Modifiers.Trajectory()
		if trajectory == game.LineDrive {
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
		if state.Modifiers.Contains(game.GroundedIntoDoublePlay) {
			b.GIDP++
		}
		if trajectory == game.Bunt || trajectory == game.BuntGrounder || trajectory == game.BuntPopup {
			switch {
			case state.Play.IsHit():
				b.BuntHits++
			case state.Modifiers.Contains(game.SacrificeHit):
				b.BuntSacrifices++
			default:
				b.BuntOuts++
			}
			if trajectory == game.BuntPopup {
				b.PopupBunts++
			}
		}
		lastPitch := state.Pitches.Last()
		if state.Play.Type == game.StrikeOut && (lastPitch == 'L' || lastPitch == 'M') {
			b.BuntOuts++
		}
		if lastPitch == 'X' {
			b.Strikes++
			b.Swings++
		}
	}
	if state.Complete || state.Incomplete {
		known, _, balls, strikes := state.Pitches.Count()
		if known {
			b.PitchesSeen += balls + strikes
			b.Strikes += strikes
			b.Swings += state.Pitches.CountUp('S', 'F', 'M', 'T')
			b.Misses += state.Pitches.CountUp('S', 'M', 'T')
			b.CalledStrikes += state.Pitches.CountUp('C')
			b.MissedBunts += state.Pitches.CountUp('M')
			b.FoulBunts += state.Pitches.CountUp('L')
			if state.Pitches.Last() == 'X' {
				b.PitchesSeen++
				b.Swings++
			}
		}
	}
	return
}
