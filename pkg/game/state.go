package game

import "regexp"

type Pitches string
type PlayerID string

type Half string

const (
	Top    = Half("Top")
	Bottom = Half("Bottom")
)

type State struct {
	InningNumber int
	Half
	Outs  int
	Score struct {
		Home, Visitor int
	}
	OutsOnPlay int `yaml:",omitempty"`
	Pitcher    PlayerID
	PlateAppearance
	Advances       []Advance  `yaml:",omitempty"`
	ScoringRunners []PlayerID `yaml:",flow,omitempty"`
	Runners        []PlayerID `yaml:",flow"`
	Comment        string     `yaml:",omitempty"`
}

type PlateAppearance struct {
	Number    int
	EventCode string
	Play
	Batter PlayerID
	Pitches
	Complete   bool `yaml:",omitempty"`
	Incomplete bool `yaml:",omitempty"`
	Modifiers  `yaml:",omitempty,flow"`
}

func (state *State) copy() *State {
	n := *state
	n.Complete = false
	n.Incomplete = false
	n.Comment = ""
	n.ScoringRunners = nil
	n.Advances = nil
	n.OutsOnPlay = 0
	if len(state.Runners) > 0 {
		n.Runners = make([]PlayerID, 3)
		for i := range state.Runners {
			n.Runners[i] = state.Runners[i]
		}
	}
	return &n
}

func (state *State) Top() bool {
	return state.Half == Top
}

func (state *State) recordOut() {
	state.Outs++
	state.OutsOnPlay++
}

var playerIDRegexp = regexp.MustCompile(`^[a-z]*\d+$`)

func IsPlayerID(s string) bool {
	return playerIDRegexp.MatchString(s)
}
