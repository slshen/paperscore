package stats

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
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
	switch state.Play.Type {
	case game.WildPitch:
		p.WP++
	case game.StolenBase:
		p.StolenBases += len(state.Play.Runners)
	case game.StrikeOut:
		p.StolenBases += len(state.StolenBases)
	}
	if state.Complete || state.Outs == 3 {
		p.Pitches += len(state.Pitches)
		p.Strikes += state.Pitches.CountUp('S', 'C', 'F')
		p.Balls += state.Pitches.CountUp('B')
		p.Swings += state.Pitches.CountUp('S', 'F')
		p.Misses += state.Pitches.CountUp('S')
		if state.Pitches.Last() == "X" {
			if state.Play.Type == game.HitByPitch {
				p.Balls++
			} else {
				p.Strikes++
				p.Swings++
			}
		}
		if state.Play.IsHit() {
			p.Hits++
		}
		switch state.Play.Type {
		case game.StrikeOut:
			fallthrough
		case game.StrikeOutPassedBall:
			fallthrough
		case game.StrikeOutWildPitch:
			fallthrough
		case game.StrikeOutPickedOff:
			p.StrikeOuts++
			if state.Pitches.Last() == "C" {
				p.StrikeOutsLooking++
			}
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
	}
}
