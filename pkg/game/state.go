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
	Outs       int
	Score      int
	OutsOnPlay int `yaml:",omitempty"`
	Pitcher    PlayerID
	PlateAppearance
	Fielders       []int      `yaml:",flow,omitempty"`
	Advances       Advances   `yaml:",omitempty"`
	ScoringRunners []PlayerID `yaml:",flow,omitempty"`
	Runners        []PlayerID `yaml:",flow"`
	Comment        string     `yaml:",omitempty"`
}

type PlateAppearance struct {
	Number    int
	EventCode string
	*Play
	Batter PlayerID
	Pitches
	NotOutOnPlay bool `yaml:",omitempty"` // not out on CS, POCS due to error
	Complete     bool `yaml:",omitempty"` // PA completed
	Incomplete   bool `yaml:",omitempty"` // inning ended w/batter still up
	Modifiers    `yaml:",omitempty,flow"`
}

func (state *State) Top() bool {
	return state.Half == Top
}

func (state *State) recordOut() {
	state.Outs++
	state.OutsOnPlay++
}

func (state *State) GetRunsScored() int {
	return len(state.ScoringRunners)
}

var playerIDRegexp = regexp.MustCompile(`^[a-z]*\d+$`)

func IsPlayerID(s string) bool {
	return playerIDRegexp.MatchString(s)
}
