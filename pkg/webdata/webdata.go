package webdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/slshen/sb/pkg/tournament"
)

type Slug string

type Tournament struct {
	ID                 Slug
	Name               string    `json:"name"`
	Date               time.Time `json:"date"`
	Wins, Losses, Ties int
	Batting            *dataframe.Data `json:"batting"`
	BestAndWorstPlays  *dataframe.Data `json:"best_and_worst_plays"`
	Games              []Slug          `json:"games"`
}

type Score struct {
	Home    int `json:"home"`
	Visitor int `json:"visitor"`
}

type Game struct {
	ID       Slug
	Home     string          `json:"home"`
	Visitor  string          `json:"visitor"`
	Score    Score           `json:"score"`
	Batting  *dataframe.Data `json:"batting"`
	Pitching *dataframe.Data `json:"pitching"`
	Plays    *dataframe.Data `json:"plays"`
	Alt      *dataframe.Data `json:"alt"`
}

type WebData struct {
	Tournaments []*Tournament
	Games       []*Game

	us string
	re stats.RunExpectancy
}

func NewWebData(us string, re stats.RunExpectancy, games []*game.Game) (*WebData, error) {
	wdat := &WebData{
		us: us,
		re: re,
	}
	for _, group := range tournament.GroupByTournament(games) {
		tourney, err := wdat.newTournament(group)
		if err != nil {
			return nil, err
		}
		wdat.Tournaments = append(wdat.Tournaments, tourney)
		for _, g := range group.Games {
			game, err := wdat.newGame(g)
			if err != nil {
				return nil, err
			}
			tourney.Games = append(tourney.Games, game.ID)
			wdat.Games = append(wdat.Games, game)
		}
	}
	return wdat, nil
}

var nameIDRe = regexp.MustCompile(`[-/\\ ]`)

func (wdat *WebData) newTournament(group *tournament.Group) (*Tournament, error) {
	year, month, day := group.Date.Date()
	tourney := &Tournament{
		ID:   Slug(fmt.Sprintf("t/%d/%02d/%02d/%s", year, month, day, nameIDRe.ReplaceAllString(group.Name, "-"))),
		Name: group.Name,
		Date: group.Date,
	}
	report, err := tournament.NewReport(wdat.us, wdat.re, group)
	if err != nil {
		return nil, err
	}
	tourney.Batting = report.GetBattingData()
	tourney.BestAndWorstPlays = report.GetBestAndWorstRE24(20)
	// tourney.Wins = ?
	return tourney, nil
}

func (wdat *WebData) newGame(g *game.Game) (*Game, error) {
	report, err := tournament.NewReport(wdat.us, wdat.re, &tournament.Group{
		Date:  g.GetDate(),
		Name:  fmt.Sprintf("%s at %s", g.Visitor, g.Home),
		Games: []*game.Game{g},
	})
	if err != nil {
		return nil, err
	}
	year, month, day := g.GetDate().Date()
	game := &Game{
		ID:      Slug(fmt.Sprintf("g/%d/%02d/%02d/%s", year, month, day, g.Number)),
		Home:    g.Home,
		Visitor: g.Visitor,
		Score: Score{
			Home:    g.Final.Home,
			Visitor: g.Final.Visitor,
		},
		Batting:  report.GetBattingData(),
		Pitching: report.GetPitchingData(),
		Plays:    report.GetRE24Data(),
		Alt:      report.GetAltData(),
	}
	return game, nil
}

func (wdat *WebData) Write(dir string) error {
	for _, tourney := range wdat.Tournaments {
		if err := writeJSON(dir, tourney.ID, tourney); err != nil {
			return err
		}
	}
	for _, game := range wdat.Games {
		if err := writeJSON(dir, game.ID, game); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(dir string, id Slug, dat interface{}) error {
	path := fmt.Sprintf("%s.json", filepath.Join(dir, filepath.FromSlash(string(id))))
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(dat)
}
