package stats

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/game"
)

type RE24 struct {
	States  []Outcome
	Team    string
	NotTeam string
}

type Outcome struct {
	Count int
	Runs  int
}

func NewRE24() *RE24 {
	return &RE24{States: make([]Outcome, 24)}
}

func (re *RE24) Read(g *game.Game) error {
	states, err := g.GetStates()
	if err != nil {
		return err
	}
	var pending []*Outcome
	pending = append(pending, &re.States[0])
	re.States[0].Count++
	for _, state := range states {
		if re.filterOut(g, state) {
			continue
		}
		if state.Outs == 3 {
			pending = append(pending, &re.States[0])
			re.States[0].Count++
			continue
		}
		for _, outcome := range pending {
			outcome.Runs += len(state.ScoringRunners)
		}
		outcome := re.GetOutcome(state.Runners, state.Outs)
		pending = append(pending, outcome)
		outcome.Count++
	}
	return nil
}

func (re *RE24) filterOut(g *game.Game, state *game.State) bool {
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

func (re *RE24) GetOutcome(runners []game.PlayerID, outs int) *Outcome {
	index := outs * 8
	for i := range runners {
		if runners[i] != "" {
			index += 1 << i
		}
	}
	return &re.States[index]
}

func (re *RE24) GetData() *Data {
	data := &Data{
		Columns: []string{"Outs", "Runners", "Count", "Runs", "Average"},
	}
	for i := range re.States {
		runners := fmt.Sprintf("%d%d%d", (i>>2)&1, (i>>1)&1, i&1)
		outcome := re.States[i]
		var rate float64
		if outcome.Count > 0 {
			rate = float64(outcome.Runs) / float64(outcome.Count)
		}
		data.Rows = append(data.Rows, Row{
			i / 8, runners, outcome.Count, outcome.Runs, fmt.Sprintf("%0.2f", rate),
		})
	}
	return data
}
