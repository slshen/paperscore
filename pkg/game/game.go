package game

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v3"
)

type Game struct {
	ID            string
	Home, Visitor string
	Date          string
	Number        int `yaml:"game"`
	Start         string
	TimeLimit     time.Duration
	Final         *struct {
		Visitor, Home int
	} `yaml:",omitempty"`
	VisitorPlays          []string `yaml:"visitorplays"`
	HomePlays             []string `yaml:"homeplays"`
	HomeID                string   `yaml:"homeid"`
	VisitorID             string   `yaml:"visitorid"`
	HomeTeam, VisitorTeam *Team
	Comments              []string
	Venue                 string

	states []*State
}

func ReadGameFiles(paths []string) (games []*Game, errs error) {
	sort.Strings(paths)
	for _, path := range paths {
		g, err := ReadGameFile(path)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		games = append(games, g)
	}
	return
}

func ReadGameFile(path string) (*Game, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadGame(path, f)
}

func ReadGame(path string, in io.Reader) (*Game, error) {
	g := &Game{}
	dec := yaml.NewDecoder(in)
	dec.KnownFields(true)
	if err := dec.Decode(g); err != nil {
		return nil, err
	}
	var errs error
	if path != "" {
		var err error
		dir := filepath.Dir(path)
		if g.HomeID != "" {
			g.HomeTeam, err = ReadTeamFile(g.Home, filepath.Join(dir, fmt.Sprintf("%s.yaml", g.HomeID)))
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
		if g.VisitorID != "" {
			g.VisitorTeam, err = ReadTeamFile(g.Visitor, filepath.Join(dir, fmt.Sprintf("%s.yaml", g.VisitorID)))
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	if g.HomeTeam == nil {
		g.HomeTeam = NewTeam(g.Home)
	}
	if g.VisitorTeam == nil {
		g.VisitorTeam = NewTeam(g.Visitor)
	}
	if g.ID == "" {
		g.ID = filepath.Base(path)
	}
	if err := g.generateStates(); err != nil {
		errs = multierror.Append(errs, err)
	}
	return g, errs
}

func (g *Game) GetStates() []*State {
	return g.states
}

func (g *Game) generateStates() error {
	m := &gameMachine{game: g}
	return m.run()
}
