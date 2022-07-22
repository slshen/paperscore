package boxscore

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/playbyplay"
	"github.com/slshen/sb/pkg/stats"
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
	IncludePlays        bool
}

func NewBoxScore(g *game.Game, re stats.RunExpectancy) (*BoxScore, error) {
	gs := stats.NewGameStats(re)
	if err := gs.Read(g); err != nil {
		return nil, err
	}
	boxscore := &BoxScore{
		Game:          g,
		HomeLineup:    &Lineup{gs.TeamStats[g.Home.Name]},
		VisitorLineup: &Lineup{gs.TeamStats[g.Visitor.Name]},
	}
	if err := boxscore.run(); err != nil {
		return nil, err
	}
	return boxscore, nil
}

func (box *BoxScore) run() error {
	states := box.Game.GetStates()
	for _, state := range states {
		if state.Comment != "" {
			box.Comments = append(box.Comments, Comment{
				Half:   state.Half,
				Inning: state.InningNumber,
				Outs:   state.Outs,
				Text:   state.Comment,
			})
		}
		if state.Top() {
			box.Score.Visitor = state.Score
		} else {
			box.Score.Home = state.Score
		}
		for len(box.InningScore) < state.InningNumber {
			box.InningScore = append(box.InningScore, Score{})
		}
		for _, adv := range state.Advances {
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
	return nil
}

func (box *BoxScore) InningScoreTable() *dataframe.Data {
	tab := &dataframe.Data{
		Columns: []*dataframe.Column{
			{
				Name:   fmt.Sprintf("%s #%s", box.Game.Date, box.Game.Number),
				Format: "%-20s",
				Values: []string{
					firstWord(box.Game.Visitor.Name, 20),
					firstWord(box.Game.Home.Name, 20),
				},
			},
		},
	}
	for i, score := range box.InningScore {
		tab.Columns = append(tab.Columns, &dataframe.Column{
			Name:   fmt.Sprintf("%2d", i+1),
			Format: "%2d",
			Values: []int{score.Visitor, score.Home},
		})
	}
	tab.Columns = append(tab.Columns,
		&dataframe.Column{
			Name:   " -",
			Format: "%2s",
			Values: []string{"", ""},
		},
		&dataframe.Column{
			Name: " R", Format: "%2d",
			Values: []int{box.Score.Visitor, box.Score.Home},
		},
		&dataframe.Column{
			Name: " H", Format: "%2d",
			Values: []int{box.VisitorLineup.TotalHits(), box.HomeLineup.TotalHits()},
		},
		&dataframe.Column{
			Name: " E", Format: "%2d",
			Values: []int{box.VisitorLineup.Errors, box.HomeLineup.Errors},
		},
	)
	return tab
}

func (box *BoxScore) ScoringPlays() (string, error) {
	gen := playbyplay.Generator{
		Game:        box.Game,
		ScoringOnly: box.IncludeScoringPlays,
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
