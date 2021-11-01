package sim

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"gopkg.in/yaml.v3"
)

type Simulation struct {
	Innings  int
	TeamName string `yaml:"team"`
	TeamID   string `yaml:"teamid"`
	Players  []game.PlayerID
	Games    string
	Seed     int64

	Probabilities struct {
		WildPitch     float64 `yaml:"wild_pitch"`
		CatchStealing float64 `yaml:"catch_stealing"`
		ReachOnError  float64 `yaml:"reach_on_error"`
	}

	team    *game.Team
	stats   *stats.TeamStats
	batters []*Batter
	rand    *rand.Rand
}

func NewSimulation(path string) (*Simulation, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sim := &Simulation{}
	if err := yaml.Unmarshal(dat, sim); err != nil {
		return nil, err
	}
	if sim.TeamID == "" {
		return nil, fmt.Errorf("teamid is required")
	}
	dir := filepath.Dir(path)
	sim.team, err = game.ReadTeamFile(sim.TeamName, filepath.Join(dir, fmt.Sprintf("%s.yaml", sim.TeamID)))
	if err != nil {
		return nil, err
	}
	if sim.Games == "" {
		return nil, fmt.Errorf("games is required")
	}
	files, err := filepath.Glob(filepath.Join(dir, sim.Games))
	if err != nil {
		return nil, err
	}
	gs := stats.NewGameStats(nil)
	for _, f := range files {
		g, err := game.ReadGameFile(f)
		if err != nil {
			return nil, err
		}
		if err := gs.Read(g); err != nil {
			return nil, err
		}
	}
	sim.stats = gs.GetStats(sim.team)
	for _, b := range sim.Players {
		sim.batters = append(sim.batters, newBatter(sim.stats.GetBatting(b)))
	}
	var src rand.Source
	if sim.Seed != 0 {
		src = rand.NewSource(sim.Seed)
	} else {
		src = rand.NewSource(time.Now().UnixNano())
	}
	// #nosec: G404
	sim.rand = rand.New(src)
	return sim, nil
}

func (sim *Simulation) Run() (*game.Game, error) {
	g, err := game.ReadGame("sim", strings.NewReader("date: simulation\n"))
	if err != nil {
		return nil, err
	}
	g.Visitor = sim.team.Name
	g.VisitorTeam = sim.team
loop:
	for {
		for _, batter := range sim.batters {
			state, err := sim.generatePlay(g, batter)
			if err != nil {
				return nil, err
			}
			if state.InningNumber == sim.Innings {
				break loop
			}
		}
	}
	return g, nil
}

func (sim *Simulation) generatePlay(g *game.Game, batter *Batter) (*game.State, error) {
	var lastState *game.State
	if states := g.GetStates(); len(states) > 0 {
		lastState = states[len(states)-1]
	}
	play := &strings.Builder{}
	if lastState != nil {
		fmt.Println(lastState)
	}
	p := sim.rand.Float64()

	playType := batter.GeneratePlay(p)

	switch playType {
	case game.Single:
		// use 0 as the fielder, since we don't care here
		fmt.Fprintf(play, "S0")
	case game.Double:
		fmt.Fprintf(play, "D0")
	case game.Triple:
		fmt.Fprintf(play, "T0")
	case game.HomeRun:
		fmt.Fprintf(play, "H")
	}
	return g.AddVisitorPlay(play.String())
}
