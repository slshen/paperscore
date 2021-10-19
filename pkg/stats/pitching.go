package stats

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
)

type Pitching struct {
	Player                             *game.Player `yaml:"-" mapstructure:",squash"`
	Pitches, Strikes, Balls            int
	Swings, Misses                     int
	Hits, Doubles, Triples, HRs, Walks int
	StrikeOuts, StrikeOutsLooking      int
	Outs, GroundOuts, FlyOuts          int
	WP, HP                             int
	BattersFaced                       int
	StolenBases                        int
	Games                              map[string]bool
}

func (p *Pitching) Whiff() string {
	// misses/swings
	if p.Swings > 0 {
		return fmt.Sprintf("%.03f", float64(p.Misses)/float64(p.Swings))[1:]
	}
	return ""
}

func (p *Pitching) SwStr() string {
	// % pitches swung & miss
	if p.Pitches > 0 {
		return fmt.Sprintf("%.03f", float64(p.Misses)/float64(p.Pitches))[1:]
	}
	return ""
}

func (p *Pitching) Record(state *game.State, lastState *game.State) {
	p.Outs += state.OutsOnPlay
	if lastState == nil || lastState.Batter != state.Batter || lastState.Pitcher != state.Pitcher {
		p.BattersFaced++
	}
	switch state.Play.Type {
	case game.WildPitch:
		p.WP++
	case game.StolenBase:
		p.StolenBases += len(state.Play.Runners)
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
		if state.Play.Hit() {
			p.Hits++
		}
		switch state.Play.Type {
		case game.StrikeOut:
			p.StrikeOuts++
			if state.Pitches.Last() == "C" {
				p.StrikeOutsLooking++
			}
		case game.Walk:
			p.Walks++
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
