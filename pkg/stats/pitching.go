package stats

import (
	"fmt"

	"github.com/slshen/paperscore/pkg/game"
)

type Pitching struct {
	PlayerData                         `mapstructure:",squash"`
	Pitches, Strikes, Balls            int
	Swings, Misses                     int
	Hits, Doubles, Triples, HRs, Walks int
	StrikeOuts, StrikeOutsLooking      int
	Outs, GroundOuts, FlyOuts          int
	WP, HP                             int
	BattersFaced                       int
	StolenBases                        int
	Whiff                              int
	SwStr                              int
	IP                                 string
}

func (p *Pitching) Update() {
	p.PlayerData.Update()
	p.Whiff = 0
	if p.Swings > 0 {
		p.Whiff = int(1000.0 * float64(p.Misses) / float64(p.Swings))
	}
	p.SwStr = 0
	if p.Pitches > 0 {
		p.SwStr = int(1000.0 * float64(p.Misses) / float64(p.Pitches))
	}
	p.IP = fmt.Sprintf("%d.%d", p.Outs/3, p.Outs%3)
}

func (p *Pitching) Record(state *game.State) {
	p.Outs += state.OutsOnPlay
	if state.LastState == nil || state.LastState.Batter != state.Batter || state.LastState.Pitcher != state.Pitcher {
		p.BattersFaced++
	}
	for _, adv := range state.Advances {
		if adv.WildPitch {
			p.WP++
		}
	}
	if state.Play.Is(game.WildPitch, game.WalkWildPitch, game.StrikeOutWildPitch) {
		p.WP++
	}
	if state.Play.Is(game.StolenBase, game.StrikeOutStolenBase) || len(state.StolenBases) > 0 {
		p.StolenBases++
	}
	if state.Complete || state.Outs == 3 {
		known, _, balls, strikes := state.Pitches.Count()
		if known {
			p.Pitches += balls + strikes
			p.Strikes += strikes
			p.Balls += balls
			p.Swings += state.Pitches.CountUp('S', 'F', 'M')
			p.Misses += state.Pitches.CountUp('S', 'M')
			last := state.Pitches.Last()
			if last == 'H' {
				p.Balls++
			}
			if last == 'X' {
				p.Strikes++
				p.Swings++
				p.Pitches++
			}
		}
		if state.Play.IsHit() {
			p.Hits++
		}
		switch state.Play.Type {
		case game.WalkPickedOff:
			fallthrough
		case game.Walk:
			fallthrough
		case game.WalkPassedBall:
			p.Walks++
		case game.WalkWildPitch:
			p.Walks++
			p.WP++
		case game.HitByPitch:
			p.HP++
		case game.Double:
			fallthrough
		case game.GroundRuleDouble:
			p.Doubles++
		case game.Triple:
			p.Triples++
		case game.HomeRun:
			p.HRs++
		case game.GroundOut:
			p.GroundOuts++
		case game.FlyOut:
			p.FlyOuts++
		}
		if state.IsStrikeOut() {
			p.StrikeOuts++
			if state.Pitches.Last() == 'C' {
				p.StrikeOutsLooking++
			}
		}
	}
}
