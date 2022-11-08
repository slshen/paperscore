package dataexport

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
)

type Event struct {
	EventID              string
	File                 string
	Line                 int
	GameID               string
	TournamentID         string
	Home                 string
	Visitor              string
	Tournament           string
	Date                 string
	GameNumber           int
	InningNumber         int
	BattingTeam          string
	Half                 string
	Outs                 int
	Score                int
	OutsOnPlay           int
	Pitcher              string
	PitcherNumber        string
	PANumber             int
	PlayCode             string
	AdvancesCodes        string
	Type                 string
	R1                   string
	R2                   string
	R3                   string
	PrevR1               string
	PrevR2               string
	PrevR3               string
	CaughtStealingBase   string
	CaughtStealingRunner string
	PickedOffRunner      string
	FieldingError        string
	Fielders             string
	StolenBases          int
	Batter               string
	BatterNumber         string
	Pitches              string
	NotOutOnPlay         bool
	Complete             bool
	Incomplete           bool
	Modifiers            string
	RunsOnPlay           int
	Comment              string
	AlternativeFor       string
	REChange             float64

	state *game.State
}

type Events []*Event

func GetEvents(re stats.RunExpectancy, g *game.Game, tournamentID string) (events Events, alts Events) {
	for _, state := range g.GetStates() {
		event := getEvent(g, re, state, tournamentID)
		events = append(events, event)
		if alt := g.GetAlternativeState(state); alt != nil {
			altEvent := getEvent(g, re, alt, tournamentID)
			alts = append(alts, altEvent)
		}
	}
	return
}

func getEvent(g *game.Game, re stats.RunExpectancy, state *game.State, tournamentID string) *Event {
	var battingTeam string
	if state.Top() {
		battingTeam = g.Visitor.Name
	} else {
		battingTeam = g.Home.Name
	}
	batter := getBatterPlayer(g, state)
	pitcher := getPitcherPlayer(g, state)
	event := &Event{
		EventID:              getEventID(g, state),
		File:                 state.Pos.Filename,
		Line:                 state.Pos.Line,
		GameID:               getGameID(g),
		TournamentID:         tournamentID,
		Home:                 g.Home.Name,
		Visitor:              g.Visitor.Name,
		Tournament:           g.Tournament,
		Date:                 g.GetDate().Format("2006-01-02"),
		GameNumber:           parseInt(g.Number),
		InningNumber:         state.InningNumber,
		Half:                 "B",
		BattingTeam:          battingTeam,
		Outs:                 state.Outs,
		Score:                state.Score,
		OutsOnPlay:           state.OutsOnPlay,
		Pitcher:              pitcher.NameOrNumber(),
		PitcherNumber:        pitcher.Number,
		PANumber:             state.PlateAppearance.Number,
		PlayCode:             state.PlayCode,
		AdvancesCodes:        strings.Join(state.AdvancesCodes, " "),
		R1:                   string(state.Runners[0]),
		R2:                   string(state.Runners[1]),
		R3:                   string(state.Runners[2]),
		Fielders:             getFielders(state),
		CaughtStealingBase:   state.CaughtStealingBase,
		CaughtStealingRunner: string(state.CaughtStealingRunner),
		PickedOffRunner:      string(state.PickedOffRunner),
		Batter:               batter.NameOrNumber(),
		BatterNumber:         batter.Number,
		Pitches:              string(state.Pitches),
		NotOutOnPlay:         state.NotOutOnPlay,
		FieldingError:        state.FieldingError.String(),
		Modifiers:            strings.Join(state.Modifiers, "/"),
		StolenBases:          len(state.StolenBases),
		RunsOnPlay:           len(state.ScoringRunners),
		Complete:             state.Complete,
		Incomplete:           state.Incomplete,
		Comment:              state.Comment,
		state:                state,
	}
	if last := state.LastState; last != nil {
		event.PrevR1 = string(last.Runners[0])
		event.PrevR2 = string(last.Runners[1])
		event.PrevR3 = string(last.Runners[2])
	}
	if state.Top() {
		event.Half = "T"
	}
	if re != nil {
		_, _, _, event.REChange = stats.GetExpectedRunsChange(re, state)
	}
	if state.AlternativeFor != nil {
		event.AlternativeFor = getEventID(g, state.AlternativeFor)
	}
	return event
}

func getBatterPlayer(g *game.Game, state *game.State) *game.Player {
	team := g.Home
	if state.Top() {
		team = g.Visitor
	}
	return team.GetPlayer(state.Batter)
}

func getPitcherPlayer(g *game.Game, state *game.State) *game.Player {
	team := g.Visitor
	if state.Top() {
		team = g.Home
	}
	return team.GetPlayer(state.Pitcher)
}

func getFielders(state *game.State) string {
	var s strings.Builder
	for _, f := range state.Fielders {
		fmt.Fprintf(&s, "%d", f)
	}
	return s.String()
}

func getGameID(g *game.Game) string {
	d := g.GetDate()
	return fmt.Sprintf("%s-%s", d.Format("20060102"), g.Number)
}

func getEventID(g *game.Game, state *game.State) string {
	return fmt.Sprintf("%s-%d", getGameID(g), state.Pos.Line)
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func (evs Events) GetData() *dataframe.Data {
	dat := &dataframe.Data{}
	var idx *dataframe.Index
	for _, ev := range evs {
		idx = dat.AppendStruct(idx, ev)
	}
	return dat
}

func (evs Events) GetAdvancesData() *dataframe.Data {
	dat := &dataframe.Data{
		Columns: []*dataframe.Column{
			{Name: "EventID", Values: dataframe.EmptyStrings},
		},
	}
	var idx *dataframe.Index
	for _, ev := range evs {
		for _, adv := range ev.state.Advances {
			dat.Columns[0].AppendString(ev.EventID)
			var m map[string]interface{}
			if err := mapstructure.Decode(adv, &m); err != nil {
				panic(err)
			}
			m["Fielders"] = join(adv.Fielders, "%d", " ")
			idx = dat.AppendMap(idx, m)
		}
	}
	return dat
}

func join[S ~[]E, E any](s S, format, sep string) string {
	var r strings.Builder
	for i, e := range s {
		if i > 0 {
			r.WriteString(sep)
		}
		fmt.Fprintf(&r, format, e)
	}
	return r.String()
}
