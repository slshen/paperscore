package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/game"
)

type RunExpectancy struct {
	Name    string
	Team    string
	NotTeam string

	states []*State24
}

type State24 struct {
	Outs    int
	Runners string
	Count   int
	Runs    int
}

func (re *RunExpectancy) Read(g *game.Game) error {
	if re.states == nil {
		re.states = make([]*State24, 24)
		for i := 0; i < 24; i++ {
			re.states[i] = &State24{
				Outs:    i / 8,
				Runners: fmt.Sprintf("%d%d%d", (i>>2)&1, (i>>1)&1, i&1),
			}
		}
	}
	states := g.GetStates()
	observed := re.reset()
	for _, state := range states {
		if re.filterOut(g, state) {
			continue
		}
		if state.Outs == 3 {
			observed = re.reset()
			continue
		}
		for _, state24 := range observed {
			state24.Runs += len(state.ScoringRunners)
		}
		index := re.getIndex(state)
		state24 := re.states[index]
		observed = append(observed, state24)
		state24.Count++
	}
	return nil
}

func (re *RunExpectancy) reset() []*State24 {
	nono := re.states[0]
	nono.Count++
	return []*State24{nono}
}

func (re *RunExpectancy) filterOut(g *game.Game, state *game.State) bool {
	team := g.Home
	if state.Top() {
		team = g.Visitor
	}
	team = strings.ToLower(team)
	if re.Team != "" {
		if !strings.HasPrefix(team, strings.ToLower(re.Team)) {
			return true
		}
	}
	if re.NotTeam != "" {
		if strings.HasPrefix(team, strings.ToLower(re.NotTeam)) {
			return true
		}
	}
	return false
}

func (re *RunExpectancy) getIndex(state *game.State) int {
	index := state.Outs * 8
	for i, runner := range state.Runners {
		if runner != "" {
			index += 1 << i
		}
	}
	return index
}

func (re *RunExpectancy) GetData() *Data {
	data := &Data{
		Name:    re.Name,
		Columns: []string{"Outs", "On321", "Count", "Runs", "Average"},
	}
	for i := range re.states {
		state24 := re.states[i]
		var rate float64
		if state24.Count > 0 {
			rate = float64(state24.Runs) / float64(state24.Count)
		}
		data.Rows = append(data.Rows, Row{
			state24.Outs, state24.Runners, state24.Count, state24.Runs, fmt.Sprintf("%0.2f", rate),
		})
	}
	return data
}
