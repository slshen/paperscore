package boxscore

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/playbyplay"
	"github.com/slshen/sb/pkg/text"
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

	IncludeScoringPlays bool
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
		box.handleAdvances(state, lineup, defense)
		lineup.recordOffense(state, lastState)
		if err := defense.recordDefense(state); err != nil {
			return err
		}
		defense.recordPitching(state, lastState)
	}
	return nil
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
}

func (box *BoxScore) InningScoreTable() string {
	tab := &text.Table{
		Columns: []text.Column{
			{Header: "", Width: 20, Left: true},
		},
	}
	for i := range box.InningScore {
		tab.Columns = append(tab.Columns, text.Column{
			Header: fmt.Sprintf("%2d", i+1),
			Width:  2,
		})
	}
	tab.Columns = append(tab.Columns,
		text.Column{Header: "  ", Width: 2},
		text.Column{Header: " R", Width: 2},
		text.Column{Header: " H", Width: 2},
		text.Column{Header: " E", Width: 2},
	)
	s := &strings.Builder{}
	s.WriteString(tab.Header())
	argsV := []interface{}{firstWord(box.Game.Visitor, 20)}
	argsH := []interface{}{firstWord(box.Game.Home, 20)}
	for _, score := range box.InningScore {
		argsV = append(argsV, score.Visitor)
		argsH = append(argsH, score.Home)
	}
	argsV = append(argsV, "--", box.Score.Visitor, box.VisitorLineup.TotalHits(),
		box.VisitorLineup.Errors)
	argsH = append(argsH, "--", box.Score.Home, box.HomeLineup.TotalHits(),
		box.HomeLineup.Errors)
	fmt.Fprintf(s, tab.Format(), argsV...)
	fmt.Fprintf(s, tab.Format(), argsH...)
	return s.String()
}

func (box *BoxScore) ScoringPlays() (string, error) {
	gen := playbyplay.Generator{
		Game:        box.Game,
		ScoringOnly: true,
	}
	s := &strings.Builder{}
	err := gen.Generate(s)
	return s.String(), err
}

func (box *BoxScore) Render(w io.Writer) error {
	tmpl := &template.Template{}
	tmpl.Funcs(template.FuncMap{
		"paste":   paste,
		"execute": executeFunc(tmpl),
		"ordinal": text.Ordinal,
	})
	tmpl, err := tmpl.ParseFS(templatesFS, "*.tmpl")
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "score.tmpl", box)
}
