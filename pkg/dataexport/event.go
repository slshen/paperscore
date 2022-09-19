package dataexport

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
)

type Event struct {
	EventID      string
	File         string
	Line         int
	Home         string
	Visitor      string
	Tournament   string
	Date         string
	GameNumber   int
	InningNumber int
	BattingTeam  string
	game.Half
	Outs                 int
	Score                int
	OutsOnPlay           int
	Pitcher              game.PlayerID
	PANumber             int
	PlayCode             string
	AdvancesCodes        []string
	Type                 game.PlayType
	Runner1              game.PlayerID
	Runner2              game.PlayerID
	Runner3              game.PlayerID
	CaughtStealingBase   string
	CaughtStealingRunner game.PlayerID
	PickedOffRunner      game.PlayerID
	FieldingError        string
	Fielder1             int
	Fielder2             int
	Fielder3             int
	Fielder4             int
	Fielder5             int
	Fielder6             int
	StolenBases          int
	Batter               game.PlayerID
	Pitches              game.Pitches
	NotOutOnPlay         bool
	Complete             bool
	Incomplete           bool
	Modifiers            string
	RunsOnPlay           int
	Comment              string
	AlternativeFor       string

	state *game.State
}

type Events []*Event

func GetEvents(g *game.Game) (Events, error) {
	var events Events
	for _, state := range g.GetStates() {
		var battingTeam string
		if state.Top() {
			battingTeam = g.Visitor.Name
		} else {
			battingTeam = g.Home.Name
		}
		event := &Event{
			EventID:              getEventID(g, state),
			File:                 state.Pos.Filename,
			Line:                 state.Pos.Line,
			Home:                 g.Home.Name,
			Visitor:              g.Visitor.Name,
			Tournament:           g.Tournament,
			Date:                 g.GetDate().Format("2006-01-02"),
			GameNumber:           parseInt(g.Number),
			InningNumber:         state.InningNumber,
			BattingTeam:          battingTeam,
			Half:                 state.Half,
			Outs:                 state.Outs,
			Score:                state.Score,
			OutsOnPlay:           state.OutsOnPlay,
			Pitcher:              state.Pitcher,
			PANumber:             state.PlateAppearance.Number,
			PlayCode:             state.PlayCode,
			AdvancesCodes:        state.AdvancesCodes,
			Runner1:              state.Runners[0],
			Runner2:              state.Runners[1],
			Runner3:              state.Runners[2],
			Fielder1:             getFielder(state, 0),
			Fielder2:             getFielder(state, 1),
			Fielder3:             getFielder(state, 2),
			Fielder4:             getFielder(state, 3),
			Fielder5:             getFielder(state, 4),
			Fielder6:             getFielder(state, 5),
			CaughtStealingBase:   state.CaughtStealingBase,
			CaughtStealingRunner: state.CaughtStealingRunner,
			PickedOffRunner:      state.PickedOffRunner,
			Batter:               state.Batter,
			Pitches:              state.Pitches,
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
		if state.AlternativeFor != nil {
			event.AlternativeFor = getEventID(g, state.AlternativeFor)
		}
		events = append(events, event)
	}
	return events, nil
}

func getFielder(state *game.State, n int) int {
	if n < len(state.Fielders) {
		return state.Fielders[n]
	}
	return 0
}

func getEventID(g *game.Game, state *game.State) string {
	d := g.GetDate()
	return fmt.Sprintf("%s-%s-%d", d.Format("20060102"), g.Number, state.Pos.Line)
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
