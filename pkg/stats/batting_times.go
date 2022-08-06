package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type BatterTimesSeenPitcher struct {
	Team             string
	Batter           game.PlayerID
	TimesSeenPitcher string
}

type BatterPitcherData struct {
	batters map[BatterTimesSeenPitcher]*Batting
	team    map[BatterTimesSeenPitcher]*Batting
}

func NewBatterPitcherData() *BatterPitcherData {
	return &BatterPitcherData{
		batters: make(map[BatterTimesSeenPitcher]*Batting),
		team:    make(map[BatterTimesSeenPitcher]*Batting),
	}
}

func (bp *BatterPitcherData) GetBatterData() *dataframe.Data {
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			dataframe.NewEmptyColumn("TimesSeenPitcher", dataframe.String),
		},
	}
	var idx *dataframe.Index
	for bts, batting := range bp.batters {
		idx = dat.MustAppendStruct(idx, batting)
		dat.Columns[0].AppendString(bts.TimesSeenPitcher)
	}
	dat = dat.Select(
		dataframe.Col("Team"),
		dataframe.Rename("Number", "Batter").WithFormat("%14s"),
		dataframe.Rename("TimesSeenPitcher", "Times"),
		dataframe.Col("AB").WithFormat("%4d").WithSummary(dataframe.Sum),
		dataframe.DeriveFloats("AVG", AVG).WithFormat("%7.3f"),
		dataframe.DeriveFloats("LAVG", LAVG).WithFormat("%7.3f"),
		dataframe.DeriveFloats("SLG", Slugging).WithFormat("%7.3f"),
		dataframe.DeriveFloats("OBS", OnBase).WithFormat("%7.3f"),
	)
	dat = dat.RSort(dataframe.Less(
		dataframe.CompareString(dat.Columns[0]),
		dataframe.CompareString(dat.Columns[1]),
	))
	return dat
}

func (bp *BatterPitcherData) GetUsBatterData(us string) *dataframe.Data {
	dat := bp.GetBatterData()
	idx := dat.GetIndex()
	usdat := dat.RFilter(func(row int) bool {
		return strings.HasPrefix(strings.ToLower(idx.GetString(row, "Team")), us)
	})
	usdat.RemoveColumn("Team")
	usdat = usdat.Rotate([]string{"Batter"}, "Times")
	return usdat
}

func (bp *BatterPitcherData) GetTeamData() *dataframe.Data {
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			dataframe.NewEmptyColumn("TimesSeenPitcher", dataframe.String),
		},
	}
	var idx *dataframe.Index
	for bts, batting := range bp.team {
		idx = dat.MustAppendStruct(idx, batting)
		dat.Columns[0].AppendString(bts.TimesSeenPitcher)
	}
	dat = dat.Select(
		dataframe.Col("Team"),
		dataframe.Rename("TimesSeenPitcher", "Times"),
		dataframe.Col("AB").WithFormat("%4d").WithSummary(dataframe.Sum),
		dataframe.DeriveFloats("AVG", AVG).WithFormat("%7.3f"),
		dataframe.DeriveFloats("LAVG", LAVG).WithFormat("%7.3f"),
		dataframe.DeriveFloats("SLG", Slugging).WithFormat("%7.3f"),
		dataframe.DeriveFloats("OBS", OnBase).WithFormat("%7.3f"),
	)
	dat = dat.RSort(dataframe.Less(
		dataframe.CompareString(dat.Columns[0]),
		dataframe.CompareString(dat.Columns[1]),
	))
	return dat
}

func (bp *BatterPitcherData) Record(g *game.Game) {
	bp.recordSide(g.Home, g.GetHomeStates())
	bp.recordSide(g.Visitor, g.GetVisitorStates())
}

func (bp *BatterPitcherData) recordSide(team *game.Team, states []*game.State) {
	type PitcherBatter struct {
		Pitcher game.PlayerID
		Batter  game.PlayerID
	}
	timesSeen := map[PitcherBatter]int{}
	for _, state := range states {
		if !state.Complete {
			continue
		}
		if team.GetPlayer(state.Batter).Inactive {
			continue
		}
		pb := PitcherBatter{Pitcher: state.Pitcher, Batter: state.Batter}
		var ns string
		n := timesSeen[pb]
		if n < 2 {
			n++
			ns = fmt.Sprintf("%d", n)
		} else {
			ns = "3+"
		}
		timesSeen[pb] = n
		bp.recordBatter(team, state, ns)
		bp.recordTeam(team, state, ns)
	}
}

func (bp *BatterPitcherData) recordBatter(team *game.Team, state *game.State, timesSeen string) {
	bts := BatterTimesSeenPitcher{Team: team.Name, Batter: state.Batter, TimesSeenPitcher: timesSeen}
	batting := bp.batters[bts]
	if batting == nil {
		batting = &Batting{}
		batting.Team = team.Name
		batting.Number = team.GetPlayer(state.Batter).NameOrNumber()
		bp.batters[bts] = batting
	}
	batting.Record(state)
}

func (bp *BatterPitcherData) recordTeam(team *game.Team, state *game.State, timesSeen string) {
	bts := BatterTimesSeenPitcher{Team: team.Name, TimesSeenPitcher: timesSeen}
	batting := bp.team[bts]
	if batting == nil {
		batting = &Batting{}
		batting.Team = team.Name
		bp.team[bts] = batting
	}
	batting.Record(state)
}
