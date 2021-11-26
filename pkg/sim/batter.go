package sim

import (
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Batter struct {
	*stats.Batting
	probs []prPlayType
}

type prPlayType struct {
	playType game.PlayType
	pr       float64
}

func newBatter(b *stats.Batting) *Batter {
	pa := float64(b.PA)
	var prSB2 float64
	if b.SB2Opp > 0 {
		prSB2 = float64(b.SB2) / float64(b.SB2Opp)
	}
	return &Batter{
		Batting: b,
		probs: []prPlayType{
			{game.Single, float64(b.Singles) / pa},
			{game.Double, float64(b.Doubles) / pa},
			{game.Triple, float64(b.Triples) / pa},
			{game.HomeRun, float64(b.HRs) / pa},
			{game.StrikeOut, float64(b.StrikeOuts) / pa},
			// should be Walk + WalkWildPitch
			{game.Walk, float64(b.Walks) / pa},
			{game.ReachedOnError, float64(b.ReachedOnError) / pa},
			{game.StolenBase, prSB2},
		},
	}
}

func (b *Batter) GeneratePlay(p float64) game.PlayType {
	for _, pr := range b.probs {
		if p < pr.pr {
			return pr.playType
		}
		p -= pr.pr
	}
	return game.GroundOut
}
