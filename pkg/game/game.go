package game

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
	League                string
	Tournament            string

	states []*State
	date   time.Time
}

var gameFileRegexp = regexp.MustCompile(`\d\d\d\d\d\d\d\d-\d.yaml`)

func ReadGamesDir(dir string) ([]*Game, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var gameFiles []string
	for _, f := range files {
		if gameFileRegexp.MatchString(f) {
			gameFiles = append(gameFiles, f)
		}
	}
	return ReadGameFiles(gameFiles)
}

func ReadGameFiles(paths []string) (games []*Game, errs error) {
	if len(paths) == 1 && strings.Contains(paths[0], "*") {
		g, err := filepath.Glob(paths[0])
		if err != nil {
			return nil, err
		}
		paths = g
	}
	sort.Strings(paths)
	for _, path := range paths {
		g, err := ReadGameFile(path)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("in game %s - %w", path, err))
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
	var err error
	g.date, err = parseGameDate(g.Date)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := g.generateStates(); err != nil {
		errs = multierror.Append(errs, err)
	}
	return g, errs
}

func parseGameDate(d string) (time.Time, error) {
	t, err := time.Parse("1/2/06", d)
	if err != nil {
		t, err = time.Parse("1/2/2006", d)
	}
	return t, err
}

func (g *Game) GetStates() []*State {
	return g.states
}

func (g *Game) AddVisitorPlay(playCode string) (*State, error) {
	g.VisitorPlays = append(g.VisitorPlays, playCode)
	var lastState *State
	if len(g.states) > 0 {
		lastState = g.states[len(g.states)-1]
	}
	m := newGameMachine(Top, lastState)
	state, err := m.runOne(playCode)
	if state != nil {
		g.states = append(g.states, state)
	}
	return state, err
}

func (g *Game) generateStates() (errs error) {
	visitorStates, err := g.runPlays(Top, g.VisitorPlays)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	homeStates, err := g.runPlays(Bottom, g.HomePlays)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	g.states = make([]*State, 0, len(visitorStates)+len(homeStates))
	half := Top
	for i, j := 0, 0; i < len(visitorStates) || j < len(homeStates); {
		var state *State
		if half == Top {
			if i < len(visitorStates) {
				state = visitorStates[i]
				i++
			}
		} else {
			if j < len(homeStates) {
				state = homeStates[j]
				j++
			}
		}
		if state != nil {
			g.states = append(g.states, state)
		}
		if state != nil && state.Outs == 3 || state == nil {
			if half == Top {
				half = Bottom
			} else {
				half = Top
			}
		}
	}
	return
}

func (g *Game) runPlays(half Half, playCodes []string) (states []*State, errs error) {
	m := newGameMachine(half, nil)
	for _, play := range playCodes {
		state, err := m.runOne(play)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		if state != nil {
			states = append(states, state)
		}
	}
	return
}

func (g *Game) GetDate() time.Time {
	return g.date
}
