package game

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Game struct {
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
	states                []*State
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
	err := dec.Decode(g)
	if err != nil {
		return nil, err
	}
	if path != "" {
		dir := filepath.Dir(path)
		if g.HomeID != "" {
			g.HomeTeam, err = ReadTeamFile(g.Home, filepath.Join(dir, fmt.Sprintf("%s.yaml", g.HomeID)))
		}
		if err == nil && g.VisitorID != "" {
			g.VisitorTeam, err = ReadTeamFile(g.Visitor, filepath.Join(dir, fmt.Sprintf("%s.yaml", g.VisitorID)))
		}
	}
	if g.HomeTeam == nil {
		g.HomeTeam = NewTeam(g.Home)
	}
	if g.VisitorTeam == nil {
		g.VisitorTeam = NewTeam(g.Visitor)
	}
	return g, err
}

func (g *Game) GetStates() ([]*State, error) {
	if g.states != nil {
		return g.states, nil
	}
	m := &gameMachine{game: g}
	err := m.run()
	return g.states, err
}
