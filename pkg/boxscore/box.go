package boxscore

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/table"
)

//go:embed "*.tmpl"
var templatesFS embed.FS

type Score struct {
	Home, Visitor int
}

type Comment struct {
	Half         game.Half
	Inning, Outs int
	Batter       string
	Text         string
}

type BoxScore struct {
	Game          *game.Game
	Score         Score
	InningScore   []Score
	HomeLineup    *Lineup
	VisitorLineup *Lineup
	Comments      []Comment
}

func NewBoxScore(g *game.Game) (*BoxScore, error) {
	boxscore := &BoxScore{
		Game:          g,
		HomeLineup:    newLineup(g.Home, g.HomeTeam),
		VisitorLineup: newLineup(g.Visitor, g.VisitorTeam),
	}
	if err := boxscore.run(); err != nil {
		return nil, err
	}
	return boxscore, nil
}

func (box *BoxScore) run() error {
	states, err := box.Game.GetStates()
	if err != nil {
		return err
	}
	for i, state := range states {
		var lineup, defense *Lineup
		if state.Top() {
			lineup = box.VisitorLineup
			defense = box.HomeLineup
		} else {
			lineup = box.HomeLineup
			defense = box.VisitorLineup
		}
		if state.Comment != "" {
			box.Comments = append(box.Comments, Comment{
				Half:   state.Half,
				Inning: state.InningNumber,
				Outs:   state.Outs,
				Batter: lineup.Team.GetPlayer(state.Batter).NameOrNumber(),
				Text:   state.Comment,
			})
		}
		box.Score.Home = state.Score.Home
		box.Score.Visitor = state.Score.Visitor
		for len(box.InningScore) < state.InningNumber {
			box.InningScore = append(box.InningScore, Score{})
		}
		var lastState *game.State
		if i > 0 {
			lastState = states[i-1]
		}
		lineup.insertBatter(state.Batter)
		defense.insertPitcher(state.Pitcher)
		if lastState != nil {
			box.handleSB(state, lastState, lineup)
		}
		box.handleAdvances(state, lineup, defense)
		lineup.recordOffensePA(state)
		defense.recordDefensePA(state)
		defense.recordPitching(state, lastState)
		if state.Outs == 3 {
			lob := 0
			for _, runner := range state.Runners {
				if runner != "" {
					lob++
				}
			}
			if state.Complete {
				// Check - credit LOB only if the batter makes the out
				data := lineup.getBatterData(state.Batter)
				data.LOB += lob
			}
			lineup.Total.LOB += lob
		}
	}
	return nil
}

func (box *BoxScore) handleSB(state, lastState *game.State, lineup *Lineup) {
	if second, third, home := state.Play.StolenBase(); second || third || home {
		if second {
			lineup.recordSteal(lastState.Runners[0])
		}
		if third {
			lineup.recordSteal(lastState.Runners[1])
		}
		if home {
			lineup.recordSteal(lastState.Runners[2])
		}
	}
}

func (box *BoxScore) handleAdvances(state *game.State, lineup, defense *Lineup) {
	for _, adv := range state.Advances {
		if adv.FieldingError != nil {
			defense.recordError(adv.FieldingError)
		}
		if adv.To == "H" && !adv.Out {
			score := &box.InningScore[state.InningNumber-1]
			if state.Top() {
				score.Visitor++
			} else {
				score.Home++
			}
		}
	}
	for _, runner := range state.ScoringRunners {
		lineup.recordRunScored(runner)
	}
}

func (box *BoxScore) InningScoreTable() string {
	tab := &table.Table{
		Columns: []table.Column{
			{Header: "", Width: 20, Left: true},
		},
	}
	for i := range box.InningScore {
		tab.Columns = append(tab.Columns, table.Column{
			Header: fmt.Sprintf("%2d", i+1),
			Width:  2,
		})
	}
	tab.Columns = append(tab.Columns,
		table.Column{Header: "  ", Width: 2},
		table.Column{Header: " R", Width: 2},
		table.Column{Header: " H", Width: 2},
		table.Column{Header: " E", Width: 2},
	)
	s := &strings.Builder{}
	s.WriteString(tab.Header())
	argsV := []interface{}{firstWord(box.Game.Visitor, 20)}
	argsH := []interface{}{firstWord(box.Game.Home, 20)}
	for _, score := range box.InningScore {
		argsV = append(argsV, score.Visitor)
		argsH = append(argsH, score.Home)
	}
	argsV = append(argsV, "--", box.Score.Visitor, box.VisitorLineup.Total.Hits,
		box.VisitorLineup.Total.Errors)
	argsH = append(argsH, "--", box.Score.Home, box.HomeLineup.Total.Hits,
		box.HomeLineup.Total.Errors)
	fmt.Fprintf(s, tab.Format(), argsV...)
	fmt.Fprintf(s, tab.Format(), argsH...)
	return s.String()
}

func (box *BoxScore) Render(w io.Writer) error {
	tmpl := &template.Template{}
	tmpl.Funcs(template.FuncMap{
		"paste":   paste,
		"execute": executeFunc(tmpl),
		"ordinal": ordinal,
	})
	tmpl, err := tmpl.ParseFS(templatesFS, "*.tmpl")
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "score.tmpl", box)
}
