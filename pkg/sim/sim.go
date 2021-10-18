package sim

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Batter struct {
	ID    game.PlayerID
	probs []prPlayType
}

type prPlayType struct {
	playType game.PlayType
	pr       float64
}

func NewBatter(b *stats.Batting) *Batter {
	pa := float64(b.PA)
	var prSB2 float64
	if b.SB2Opportunities > 0 {
		prSB2 = float64(b.SB2) / float64(b.SB2Opportunities)
	}
	return &Batter{
		ID: b.Player.PlayerID,
		probs: []prPlayType{
			{game.Single, float64(b.Singles) / pa},
			{game.Double, float64(b.Doubles) / pa},
			{game.Triple, float64(b.Triples) / pa},
			{game.HomeRun, float64(b.HRs) / pa},
			{game.StrikeOut, float64(b.StrikeOuts) / pa},
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

func GenerateGame(team *game.Team, stats *stats.TeamStats) *game.Game {
	g, err := game.ReadGame("sim", strings.NewReader("date: simulation\n"))
	if err != nil {
		panic(err)
	}
	g.Visitor = team.Name
	g.VisitorTeam = team
	for player := range team.Players {
		_, err = g.AddVisitorPlay(fmt.Sprintf("%s,BBBB,W.B-1", player))
		if err != nil {
			panic(err)
		}
		fmt.Println(player)
		break
	}
	return g
}
