package game

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/slshen/sb/pkg/gamefile"
)

type Game struct {
	File          *gamefile.File `yaml:"-"`
	ID            string
	Home, Visitor string
	Final         *struct {
		Visitor, Home int
	} `yaml:",omitempty"`
	HomeID                string `yaml:"homeid"`
	VisitorID             string `yaml:"visitorid"`
	HomeTeam, VisitorTeam *Team
	League                string
	Tournament            string
	Date                  string
	Number                string

	visitorStates []*State
	homeStates    []*State
	states        []*State
	altStates     altStatesMap
	date          time.Time
}

type altStatesMap map[*State][]*State

var gameFileRegexp = regexp.MustCompile(`\d\d\d\d\d\d\d\d-\d.yaml`)

func ReadGamesDir(dir string) ([]*Game, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	gmfiles, err := filepath.Glob(filepath.Join(dir, "*.gm"))
	if err != nil {
		return nil, err
	}
	files = append(files, gmfiles...)
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
	var (
		gf  *gamefile.File
		err error
	)
	if strings.HasSuffix(path, ".yaml") {
		gf, err = gamefile.ParseYAMLFile(path)
	} else {
		gf, err = gamefile.ParseFile(path)
	}
	if err != nil {
		return nil, err
	}
	return newGame(gf)
}

func parseGameDate(d string) (time.Time, error) {
	t, err := time.Parse("1/2/06", d)
	if err != nil {
		t, err = time.Parse("1/2/2006", d)
	}
	return t, err
}

func newGame(gf *gamefile.File) (*Game, error) {
	g := &Game{
		File:       gf,
		Home:       gf.Properties["home"],
		HomeID:     gf.Properties["homeid"],
		Visitor:    gf.Properties["visitor"],
		VisitorID:  gf.Properties["visitorid"],
		Tournament: gf.Properties["tournament"],
		League:     gf.Properties["league"],
		Number:     gf.Properties["game"],
		Date:       gf.Properties["date"],
		altStates:  make(altStatesMap),
	}
	var errs error
	if gf.Path != "" {
		var err error
		dir := filepath.Dir(gf.Path)
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
	if g.Home == "" {
		g.Home = g.HomeTeam.Name
	}
	if g.Visitor == "" {
		g.Visitor = g.VisitorTeam.Name
	}
	if g.ID == "" {
		id := filepath.Base(gf.Path)
		dot := strings.LastIndex(id, ".")
		if dot > 0 {
			id = id[0:dot]
		}
		g.ID = id
	}
	var err error
	g.date, err = parseGameDate(g.File.Properties["date"])
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := g.generateStates(); err != nil {
		errs = multierror.Append(errs, err)
	}
	return g, errs
}

func (g *Game) GetStates() []*State {
	return g.states
}

func (g *Game) generateStates() (errs error) {
	var err error
	g.visitorStates, err = g.runPlays(g.VisitorTeam, g.HomeTeam, Top,
		g.File.GetVisitorEvents())
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	g.homeStates, err = g.runPlays(g.HomeTeam, g.VisitorTeam, Bottom,
		g.File.GetHomeEvents())
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	g.states = make([]*State, 0, len(g.visitorStates)+len(g.homeStates))
	half := Top
	for i, j := 0, 0; i < len(g.visitorStates) || j < len(g.homeStates); {
		var state *State
		if half == Top {
			if i < len(g.visitorStates) {
				state = g.visitorStates[i]
				i++
			}
		} else {
			if j < len(g.homeStates) {
				state = g.homeStates[j]
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

func (g *Game) runPlays(battingTeam, fieldingTeam *Team, half Half, events *gamefile.TeamEvents) (states []*State, errs error) {
	if events == nil {
		return
	}
	m := newGameMachine(half, battingTeam, fieldingTeam)
	lastState := &State{
		InningNumber: 1,
		Half:         half,
		Runners:      make([]PlayerID, 3),
	}
	for _, event := range events.Events {
		if event.Empty {
			continue
		}
		if m.final {
			errs = multierror.Append(errs,
				fmt.Errorf("%s: cannot have more plays after final score", event.Pos))
			break
		}
		switch {
		case event.Play != nil:
			state, err := m.handleActualPlay(event.Play, lastState)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if state != nil {
				states = append(states, state)
				lastState = state
			}
		case event.Alternative != nil:
			state, err := m.handleAlternative(event.Alternative, lastState)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if state != nil {
				g.altStates[lastState] = append(g.altStates[lastState], state)
			}
		default:
			s, err := m.handleSpecialEvent(event, lastState)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if s != nil {
				lastState = s
			}
		}
	}
	return
}

func (g *Game) GetAlternativeStates(state *State) []*State {
	return g.altStates[state]
}

func (g *Game) GetDate() time.Time {
	return g.date
}
