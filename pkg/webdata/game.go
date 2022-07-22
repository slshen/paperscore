package webdata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/slshen/sb/pkg/boxscore"
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/tournament"
)

type Game struct {
	Them  string
	Title string

	Batting  *dataframe.Data
	Pitching *dataframe.Data
	Plays    *dataframe.Data
	Alt      *dataframe.Data

	box  *boxscore.BoxScore
	game *game.Game
}

func newGame(wdat *WebData, g *game.Game) (*Game, error) {
	report, err := tournament.NewReport(wdat.us, wdat.re, &tournament.Group{
		Date:  g.GetDate(),
		Name:  fmt.Sprintf("%s at %s", g.Visitor.Name, g.Home.Name),
		Games: []*game.Game{g},
	})
	if err != nil {
		return nil, err
	}
	box, err := boxscore.NewBoxScore(g, wdat.re)
	if err != nil {
		return nil, err
	}

	_, them := g.GetUsAndThem(wdat.us)
	game := &Game{
		Them:     them.Name,
		Title:    fmt.Sprintf("%s #%s %s", g.Date, g.Number, them.Name),
		Batting:  report.GetBattingData(),
		Pitching: report.GetPitchingData(),
		Plays: report.GetRE24Data().Select(
			dataframe.DeriveInts("#", func(idx *dataframe.Index, i int) int {
				return i + 1
			}),
			dataframe.Col("Bat"),
			dataframe.Rename("O", "Outs"),
			dataframe.Col("Rnr"),
			dataframe.Col("RE24"),
			dataframe.Col("R"),
			dataframe.Col("Play"),
		),
		Alt: report.GetAltData().Select(
			dataframe.Col("RCost"),
			dataframe.Col("Bat"),
			dataframe.Rename("O", "Outs"),
			dataframe.Col("Rnr"),
			dataframe.Col("Reality"),
			dataframe.Col("Alternate"),
			dataframe.Col("Comment"),
		),
		box:  box,
		game: g,
	}
	game.Pitching.Name = "Pitching"
	game.Plays.Name = "Plays"
	game.Batting.Name = "Batting"
	game.Alt.Name = "Alt Plays"
	return game, nil
}

func (g *Game) GetPage() *Page {
	p := &Page{
		ID: ToID(fmt.Sprintf("%s-%s %s", g.game.Date, g.game.Number, g.Them)),
		Pages: []HasPage{
			newTable(g.Batting),
			newTable(g.Pitching),
			newTable(g.Plays),
			newTable(g.Alt),
		},
		Resources: map[string]ResourceContent{},
		Content:   g.GetContent,
	}
	p.Resources[filepath.Base(g.game.File.Path)] = func() ([]byte, error) {
		return os.ReadFile(g.game.File.Path)
	}
	p.Set("title", g.Title)
	return p
}

func (g *Game) GetContent() string {
	s := &strings.Builder{}
	fmt.Fprintf(s, "```\n")
	_ = g.box.Render(s)
	fmt.Fprintf(s, "```\n")
	return s.String()
}
