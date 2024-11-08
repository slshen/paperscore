package stats

import (
	"fmt"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
)

type PitcherTimesLineup struct {
	pitchers map[PitcherTeamTimes]*Pitching
}

type PitcherTeamTimes struct {
	Pitcher   game.PlayerID
	Team      string
	TimesSeen int
}

func NewPitcherTimesLineup() *PitcherTimesLineup {
	return &PitcherTimesLineup{
		pitchers: make(map[PitcherTeamTimes]*Pitching),
	}
}

func (ptl *PitcherTimesLineup) Record(g *game.Game) {
	ptl.recordSide(g.Visitor, g.GetVisitorStates())
}

func (ptl *PitcherTimesLineup) recordSide(team *game.Team, states []*game.State) {
	type pitcherBatter struct {
		Pitcher game.PlayerID
		Batter  game.PlayerID
	}
	timesSeen := map[pitcherBatter]int{}
	for _, state := range states {
		if !state.Complete {
			continue
		}
		pb := pitcherBatter{Pitcher: state.Pitcher, Batter: state.Batter}
		t := timesSeen[pb]
		if t < 3 {
			t++
			timesSeen[pb] = t
		}
		ptt := PitcherTeamTimes{Pitcher: state.Pitcher, Team: team.Name, TimesSeen: t}
		pitching := ptl.pitchers[ptt]
		if pitching == nil {
			pitching = &Pitching{}
			pitching.Team = team.Name
			pitching.Name = team.GetPlayer(state.Pitcher).NameOrNumber()
			ptl.pitchers[ptt] = pitching
		}
		pitching.Record(state)
	}
}

func (ptl *PitcherTimesLineup) GetData() *dataframe.Data {
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			{Name: "Times", Values: dataframe.EmptyStrings},
		},
	}
	var idx *dataframe.Index
	for ptt, pitching := range ptl.pitchers {
		idx = dat.AppendStruct(idx, pitching)
		var s string
		if ptt.TimesSeen < 3 {
			s = fmt.Sprintf("%d", ptt.TimesSeen)
		} else {
			s = fmt.Sprintf("%d+", ptt.TimesSeen)
		}
		dat.Columns[0].AppendString(s)
	}
	dat = dat.Select(
		dataframe.Col("Times"),
	)
	return dat
}
