package game

import (
	"fmt"
	"regexp"
	"strings"
)

type Pitches string
type PlayerID string
type FileLocation struct {
	Filename string
	Line     int
}

type Half string

const (
	Top    = Half("Top")
	Bottom = Half("Bottom")
)

type State struct {
	Pos          FileLocation `yaml:",flow"`
	InningNumber int
	Half
	Outs    int
	Score   int
	Pitcher PlayerID
	PlateAppearance
	// Fielders       []int      `yaml:",flow,omitempty"`

	Runners [3]PlayerID `yaml:",omitempty,flow"`
	// Runners        []PlayerID `yaml:",flow"`
	Comment        string `yaml:",omitempty"`
	LastState      *State `yaml:"-"`
	AlternativeFor *State `yaml:"-"`
}

type PlateAppearance struct {
	Number        int
	PlayCode      string
	AdvancesCodes []string
	Advances      Advances `yaml:",omitempty"`
	Play
	Batter PlayerID
	Pitches
	Complete   bool `yaml:",omitempty"` // PA completed
	Incomplete bool `yaml:",omitempty"` // inning ended w/batter still up
	Modifiers  `yaml:",omitempty,flow"`
}

func (pos FileLocation) String() string {
	return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
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

func (state *State) GetBaseRunner(base string) (runner PlayerID, err error) {
	if base == "H" {
		err = fmt.Errorf("a runner cannot be at H")
		return
	}
	if state.LastState == nil || (state.LastState.InningNumber != state.InningNumber) {
		err = fmt.Errorf("no runners are on base at the start of a half-inning")
		return
	}
	runner = state.LastState.Runners[runnerNumber[base]]
	if runner == "" {
		err = fmt.Errorf("no runner on %s", base)
	}
	return
}

func (state *State) IsAB() bool {
	return state.Complete &&
		!(state.Play.Is(Walk, WalkPickedOff, HitByPitch, WalkWildPitch, WalkPassedBall, CatcherInterference) ||
			(state.Play.Type == ReachedOnError && state.Modifiers.Contains(Obstruction)) ||
			state.Modifiers.Contains(SacrificeFly, SacrificeHit))
}

func (pa *PlateAppearance) GetPlayAdvancesCode() string {
	s := &strings.Builder{}
	fmt.Fprintf(s, "%s", pa.PlayCode)
	for _, adv := range pa.AdvancesCodes {
		fmt.Fprintf(s, " %s", adv)
	}
	return s.String()
}

var playerIDRegexp = regexp.MustCompile(`^[a-z]*\d+$`)

func IsPlayerID(s string) bool {
	return playerIDRegexp.MatchString(s)
}

func (id PlayerID) IsUs() bool {
	return len(id) > 0 && !(id[0] >= '0' && id[0] <= '9')
}
