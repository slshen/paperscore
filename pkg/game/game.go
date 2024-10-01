package game

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/slshen/sb/pkg/gamefile"
)

type Score struct {
	Home    int `json:"home"`
	Visitor int `json:"visitor"`
}

type Game struct {
	File          *gamefile.File `yaml:"-"`
	ID            string
	Home, Visitor *Team
	Final         Score
	League        string
	Tournament    string
	Season        string
	Date          string
	Number        string

	visitorStates []*State
	homeStates    []*State
	states        []*State
	altStates     altStatesMap
	date          time.Time
}

type altStatesMap map[*State]*State

var gameFileRegexp = regexp.MustCompile(`\d\d\d\d\d\d\d\d-\d.(yaml|gm)`)

func globExpand(paths []string) ([]string, error) {
	var res []string
	for _, path := range paths {
		if strings.Contains(path, "*") {
			g, err := filepath.Glob(paths[0])
			if err != nil {
				return nil, err
			}
			res = append(res, g...)
		} else {
			res = append(res, path)
		}
	}
	return res, nil
}

func ReadGames(fileOrDirs []string) ([]*Game, error) {
	var games []*Game
	fileOrDirs, err := globExpand(fileOrDirs)
	if err != nil {
		return nil, err
	}
	sort.Strings(fileOrDirs)
	for _, fileOrDir := range fileOrDirs {
		stat, err := os.Stat(fileOrDir)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			dirGames, err := ReadGamesDir(fileOrDir)
			if err != nil {
				return nil, err
			}
			games = append(games, dirGames...)
			ents, err := os.ReadDir(fileOrDir)
			if err != nil {
				return nil, err
			}
			for _, ent := range ents {
				if ent.IsDir() {
					moreGames, err := ReadGames([]string{filepath.Join(fileOrDir, ent.Name())})
					if err != nil {
						return nil, err
					}
					games = append(games, moreGames...)
				}
			}
			continue
		}
		fileGames, err := ReadGameFiles([]string{fileOrDir})
		if err != nil {
			return nil, err
		}
		games = append(games, fileGames...)
	}
	return games, nil
}

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
	paths, err := globExpand(paths)
	if err != nil {
		return nil, err
	}
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
	return NewGame(gf)
}

func NewGame(gf *gamefile.File) (*Game, error) {
	g := &Game{
		File:       gf,
		Tournament: gf.Properties["tournament"],
		Season:     gf.Properties["season"],
		League:     gf.Properties["league"],
		Number:     gf.Properties["game"],
		Date:       gf.Properties["date"],
		altStates:  make(altStatesMap),
	}
	var errs error
	var err error
	dir := filepath.Dir(gf.Path)
	g.Home, err = GetTeam(dir, gf.Properties["home"], gf.Properties["homeid"])
	if err != nil {
		return nil, err
	}
	g.Visitor, err = GetTeam(dir, gf.Properties["visitor"], gf.Properties["visitorid"])
	if err != nil {
		return nil, err
	}
	if g.ID == "" {
		id := filepath.Base(gf.Path)
		dot := strings.LastIndex(id, ".")
		if dot > 0 {
			id = id[0:dot]
		}
		g.ID = id
	}
	g.date, err = g.File.GetGameDate()
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

func (g *Game) GetHomeStates() []*State {
	return g.homeStates
}

func (g *Game) GetVisitorStates() []*State {
	return g.visitorStates
}

func (g *Game) generateStates() (errs error) {
	var err error
	g.visitorStates, err = g.runPlays(g.Visitor, g.Home, Top,
		g.File.VisitorEvents)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	if len(g.visitorStates) > 0 {
		g.Final.Visitor = g.visitorStates[len(g.visitorStates)-1].Score
	}
	g.homeStates, err = g.runPlays(g.Home, g.Visitor, Bottom,
		g.File.HomeEvents)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	if len(g.homeStates) > 0 {
		g.Final.Home = g.homeStates[len(g.homeStates)-1].Score
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

func (g *Game) runPlays(battingTeam, fieldingTeam *Team, half Half, events []*gamefile.Event) (states []*State, errs error) {
	if events == nil {
		return
	}
	m := newGameMachine(battingTeam, fieldingTeam)
	lastState := &State{
		InningNumber: 1,
		Half:         half,
	}
	for _, event := range events {
		if event.Empty {
			continue
		}
		if m.final {
			errs = multierror.Append(errs,
				NewError("cannot have more plays after final score", event.Pos))
			break
		}
		switch {
		case event.Play != nil:
			state, err := m.handleActualPlay(event.Play, lastState)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if state != nil {
				for _, after := range event.Afters {
					if after.CourtesyRunner != nil {
						// assume courtesy runner is for batter
						cr := m.battingTeam.parsePlayerID(*after.CourtesyRunner)
						for i := range state.Runners {
							if state.Runners[i] == state.Batter {
								state.Runners[i] = cr
								break
							}
						}
					}
				}
				state.Comment = event.Comment
				states = append(states, state)
				lastState = state
			}
		case event.Alternative != nil:
			state, err := m.handleAlternative(event.Alternative, lastState)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
			if state != nil {
				state.Comment = event.Alternative.Comment
				if g.altStates[lastState] != nil {
					errs = multierror.Append(errs,
						NewError("only a single alternate state is allowed", event.Pos))
				} else {
					g.altStates[lastState] = state
				}
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

func (g *Game) GetAlternativeState(state *State) *State {
	return g.altStates[state]
}

func (g *Game) GetDate() time.Time {
	return g.date
}

func (g *Game) GetUsAndThem(us string) (*Team, *Team) {
	if g.Home.IsUs(us) {
		return g.Visitor, g.Home
	}
	return g.Visitor, g.Home
}

func (g *Game) GetTournament() string {
	if g.Tournament != "" {
		return g.Tournament
	}
	if g.League != "" {
		return g.League
	}
	return "Other"
}

func (g *Game) GetSeason(us string) string {
	if g.Season != "" {
		return g.Season
	}
	return ""
}
