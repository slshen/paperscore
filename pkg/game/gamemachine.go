package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/slshen/sb/pkg/gamefile"
)

type gameMachine struct {
	battingTeam  *Team
	fieldingTeam *Team
	pitcher      PlayerID
	PA           int
	basePutOuts  map[string]bool
	final        bool
	modifiers    Modifiers
}

func newGameMachine(half Half, battingTeam, fieldingTeam *Team) *gameMachine {
	m := &gameMachine{
		battingTeam:  battingTeam,
		fieldingTeam: fieldingTeam,
	}
	return m
}

func (m *gameMachine) newState(lastState *State) *State {
	state := &State{
		InningNumber: lastState.InningNumber,
		Outs:         lastState.Outs,
		Half:         lastState.Half,
		Score:        lastState.Score,
		Pitcher:      m.pitcher,
		Runners:      make([]PlayerID, 3),
		LastState:    lastState,
	}
	if state.Outs == 3 {
		state.Outs = 0
		state.LastState = nil
		state.InningNumber++
	}
	return state
}

func (m *gameMachine) handleAlternative(alt *gamefile.Alternative, lastState *State) (*State, error) {
	realLastState := lastState.LastState
	if realLastState == nil {
		// this is an alternative play at the top of the inning
		// we don't have the real real last state (which the last play
		// of the last inning), so we'll just fake one here
		realLastState = &State{
			InningNumber: lastState.InningNumber - 1,
			Outs:         3,
			Half:         lastState.Half,
			Pitcher:      lastState.Pitcher,
			Runners:      make([]PlayerID, 3),
		}
		realLastState.Batter = lastState.Batter
	}
	state := m.newState(realLastState)
	state.Batter = realLastState.Batter
	state.Pitches = realLastState.Pitches
	state.AlternativeFor = lastState
	err := m.handlePlay(alt, state)
	return state, err
}

func (m *gameMachine) handleActualPlay(play *gamefile.ActualPlay, lastState *State) (*State, error) {
	state := m.newState(lastState)
	if play.ContinuedPlateAppearance {
		if state.LastState == nil {
			return nil, fmt.Errorf("%s: ... can only be used to continue a plate appearance", play.GetPos())
		}
		state.Pitches = state.LastState.Pitches + Pitches(play.PitchSequence)
		state.Batter = state.LastState.Batter
	} else {
		state.Batter = m.battingTeam.parsePlayerID(play.Batter.String())
		state.Pitches = Pitches(play.PitchSequence)
	}
	if state.Batter == "" {
		return nil, fmt.Errorf("%s: no batter for %s", play.GetPos(), play.GetCode())
	}
	err := m.handlePlay(play, state)
	return state, err
}

func (m *gameMachine) handlePlay(play gamefile.Play, state *State) error {
	state.PlayCode = play.GetCode()
	state.AdvancesCodes = play.GetAdvances()
	if state.PlayCode == "" {
		return fmt.Errorf("%s: empty event code in %s", play.GetPos(), play.GetCode())
	}
	state.Comment = play.GetComment()
	m.basePutOuts = nil
	if err := m.parseAdvances(play, state); err != nil {
		return err
	}
	if err := m.handlePlayCode(play, state); err != nil {
		return err
	}
	if err := m.moveRunners(play, state); err != nil {
		return err
	}
	if state.Outs == 3 && !state.Complete {
		// inning ended
		state.Incomplete = true
	}
	if state.Complete {
		if !strings.HasSuffix(string(state.Pitches), "X") &&
			state.Play.BallInPlay() {
			// fix up pitches
			state.Pitches += "X"
		}
		state.Modifiers = m.modifiers
		m.PA++
	}
	state.PlateAppearance.Number = m.PA
	return nil
}

func (m *gameMachine) parseAdvances(play gamefile.Play, state *State) error {
	var runners []PlayerID
	if state.LastState != nil {
		runners = state.LastState.Runners
	}
	var err error
	state.Advances, err = parseAdvances(play, state.Batter, runners)
	return err
}

func (m *gameMachine) handleSpecialEvent(event *gamefile.Event, state *State) (*State, error) {
	if event.Pitcher != "" {
		m.pitcher = m.fieldingTeam.parsePlayerID(event.Pitcher)
	}
	if event.Score != "" {
		if state.Outs != 3 {
			return nil, fmt.Errorf("%s: the inning with %d outs has not ended after %s",
				event.Pos, state.Outs, state.PlayCode)
		}
		score, err := strconv.Atoi(event.Score)
		if err != nil || state.Score != score {
			return nil, fmt.Errorf("%s: in inning %d # runs is %d not %s", event.Pos,
				state.InningNumber, state.Score, event.Score)
		}
	}
	if event.Final != "" {
		score, err := strconv.Atoi(event.Final)
		if err != nil || state.Score != score {
			return nil, fmt.Errorf("%s: in inning %d final score is %d not %s", event.Pos,
				state.InningNumber, state.Score, event.Score)
		}
		m.final = true
	}
	if event.RAdjRunner != "" {
		runner := m.battingTeam.parsePlayerID(event.RAdjRunner.String())
		base := event.RAdjBase
		if runner == "" || !(base == "1" || base == "2" || base == "3") {
			return nil, fmt.Errorf("%s: invalid base %s for radj", event.Pos, event.RAdjBase)
		}
		if state.Outs != 3 {
			return nil, fmt.Errorf("%s: radj must be at the inning start", event.Pos)
		}
		lastState := &State{
			InningNumber: state.InningNumber + 1,
			Half:         state.Half,
			Outs:         0,
			Score:        state.Score,
			Runners:      make([]PlayerID, 3),
		}
		lastState.Runners[BaseNumber[base]] = runner
		return lastState, nil
	}
	return nil, nil
}

func (m *gameMachine) impliedAdvance(play gamefile.Play, state *State, code string) {
	advance, err := parseAdvance(play, code)
	if err != nil {
		panic(err)
	}
	if state.Advances[advance.From] == nil {
		state.Advances[advance.From] = advance
	}
	if state.Advances[advance.From].To == advance.To {
		state.Advances[advance.From].Implied = true
	}
}

func (m *gameMachine) handlePlayCode(play gamefile.Play, state *State) error {
	pp := playCodeParser{}
	pp.parsePlayCode(play.GetCode())
	m.modifiers = pp.modifiers
	switch {
	case pp.playIs("$"):
		state.Play = &Play{
			Type:     FlyOut,
			Fielders: pp.getFielders(0),
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("$$"):
		fallthrough
	case pp.playIs("$$$"):
		state.Play = &Play{
			Type: GroundOut,
		}
		for i := range pp.playMatches {
			state.Play.Fielders = append(state.Play.Fielders, pp.getFielder(i))
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("K"):
		state.Play = &Play{
			Type: StrikeOut,
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("K+SB%"):
		state.Play = &Play{
			Type:    StrikeOut,
			Runners: make([]PlayerID, 1),
		}
		state.recordOut()
		state.Complete = true
		if err := m.handleStolenBase(play, state, pp.playMatches); err != nil {
			return err
		}
	case pp.playIs("K+PO%($$)"):
		from := pp.playMatches[0]
		if !(from == "1" || from == "2" || from == "3") {
			return fmt.Errorf("%s: illegal picked off base in %s", play.GetPos(), pp.playCode)
		}
		runner, err := state.GetBaseRunner(from)
		if err != nil {
			return fmt.Errorf("%s: cannot pick off in %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = &Play{
			Type:     StrikeOutPickedOff,
			Fielders: pp.getFielders(1, 2),
			Runners:  []PlayerID{runner},
		}
		state.Complete = true
		state.recordOut()
		advance := state.Advances[from]
		if advance == nil {
			state.recordOut()
			m.putOut(from)
		} else {
			return fmt.Errorf("%s: picked off runner on %s cannot advance", play.GetPos(), from)
		}
	case pp.playIs("W+WP"):
		state.Play = &Play{
			Type: WalkWildPitch,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("W+PB"):
		state.Play = &Play{
			Type: WalkPassedBall,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("W"):
		state.Play = &Play{
			Type: Walk,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("SB%;SB%;SB%") || pp.playIs("SB%;SB%") || pp.playIs("SB%"):
		state.Play = &Play{
			Type:    StolenBase,
			Runners: make([]PlayerID, len(pp.playMatches)),
		}
		if err := m.handleStolenBase(play, state, pp.playMatches); err != nil {
			return err
		}
	case pp.playIs("K2$"):
		state.Play = &Play{
			Type:     StrikeOut,
			Fielders: []int{2, pp.getFielder(0)},
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("K+PB"):
		state.Play = &Play{
			Type:     StrikeOutPassedBall,
			Fielders: []int{2},
		}
		state.Complete = true
		m.impliedAdvance(play, state, "B-1")
	case pp.playIs("K+WP"):
		state.Play = &Play{
			Type:     StrikeOutWildPitch,
			Fielders: []int{1},
		}
		state.Complete = true
		m.impliedAdvance(play, state, "B-1")
	case pp.playIs("S$"):
		state.Play = &Play{
			Type: Single,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("D$"):
		state.Play = &Play{
			Type: Double,
		}
		m.impliedAdvance(play, state, "B-2")
		state.Complete = true
	case pp.playIs("DGR"):
		state.Play = &Play{
			Type: Double,
		}
		m.impliedAdvance(play, state, "B-2")
		state.Complete = true
	case pp.playIs("T$"):
		state.Play = &Play{
			Type: Triple,
		}
		m.impliedAdvance(play, state, "B-3")
		state.Complete = true
	case pp.playIs("H"):
		state.Play = &Play{
			Type: HomeRun,
		}
		m.impliedAdvance(play, state, "B-H")
		state.Complete = true
	case pp.playIs("PB"):
		state.Play = &Play{
			Type: PassedBall,
		}
		// movement in advances
	case pp.playIs("WP"):
		state.Play = &Play{
			Type: WildPitch,
		}
		// movement in advances
	case pp.playIs("HP"):
		state.Play = &Play{
			Type: HitByPitch,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("E$"):
		fe, err := parseFieldingError(play, pp.playCode)
		if err != nil {
			return fmt.Errorf("%s: cannot parse fielding error in %s - %w", play.GetPos(),
				pp.playCode, err)
		}
		state.Play = &Play{
			Type:          ReachedOnError,
			Fielders:      pp.getFielders(0),
			FieldingError: fe,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("C/E$"):
		state.Play = &Play{
			Type: CatcherInterference,
			FieldingError: &FieldingError{
				Fielder: pp.getFielder(0),
			},
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("PO%(E$)"):
		state.Play = &Play{
			Type:     PickedOff,
			Fielders: pp.getFielders(1),
			FieldingError: &FieldingError{
				Fielder: pp.getFielder(1),
			},
		}
		state.NotOutOnPlay = true
	case pp.playIs("PO%($$)"):
		from := pp.playMatches[0]
		if !(from == "1" || from == "2" || from == "3") {
			return fmt.Errorf("%s: illegal picked off base in %s", play.GetPos(), pp.playCode)
		}
		runner, err := state.GetBaseRunner(from)
		if err != nil {
			return fmt.Errorf("%s: cannot pick off in %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = &Play{
			Type:     PickedOff,
			Runners:  []PlayerID{runner},
			Fielders: pp.getFielders(1, 2),
		}
		advance := state.Advances[from]
		if advance == nil {
			state.recordOut()
			m.putOut(from)
		} else {
			state.NotOutOnPlay = advance.FieldingError != nil
		}
	case pp.playIs("FC$"):
		// outs are in the advance, if any
		state.Play = &Play{
			Type:     FieldersChoice,
			Fielders: pp.getFielders(0),
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("$$(%)$") || pp.playIs("$(%)$$") || pp.playIs("$$(%)$$"):
		if !m.modifiers.Contains("GDP") {
			return fmt.Errorf("%s: play should contain GDP modifier in %s", play.GetPos(), pp.playCode)
		}
		base := pp.playMatches[2]
		runner, err := state.GetBaseRunner(base)
		if err != nil {
			return fmt.Errorf("%s: no runner in double play %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = &Play{
			Type:    DoublePlay,
			Runners: []PlayerID{runner},
		}
		// should pass fielders to record out to do assists
		state.recordOut()
		state.recordOut()
		m.putOut(base)
		state.Complete = true
	case pp.playIs("$(B)$(%)"):
		fallthrough
	case pp.playIs("$(B)$$(%)"):
		fallthrough
	case pp.playIs("$(B)$$$(%)"):
		if !m.modifiers.Contains("LDP", "FDP") {
			return fmt.Errorf("%s: play should contain LDP or FDP modifier in %s (%v)",
				play.GetPos(), pp.playCode, state.Modifiers)
		}
		base := pp.playMatches[len(pp.playMatches)-1]
		runner, err := state.GetBaseRunner(base)
		if err != nil {
			return fmt.Errorf("no runner in lineout double play %s - %w", pp.playCode, err)
		}
		state.Play = &Play{
			Type:    DoublePlay,
			Runners: []PlayerID{runner},
		}
		state.recordOut()
		state.recordOut()
		m.putOut(base)
		state.Complete = true
	case pp.playIs("CS%($$)"):
		fallthrough
	case pp.playIs("CS%($$$)"):
		fallthrough
	case pp.playIs("CS%($$$$)"):
		to := pp.playMatches[0]
		if !(to == "2" || to == "3" || to == "H") {
			return fmt.Errorf("illegal caught stealing base code %s", pp.playCode)
		}
		from := PreviousBase[to]
		advance := state.Advances[from]
		if advance == nil {
			state.recordOut()
			m.putOut(from)
		} else {
			state.NotOutOnPlay = advance.FieldingError != nil
		}
		runner, err := state.GetBaseRunner(from)
		if err != nil {
			return fmt.Errorf("%s: cannot catch stealing runner in %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = &Play{
			Type:    CaughtStealing,
			Runners: []PlayerID{runner},
			Base:    to,
			//Fielders: ,
		}
	case pp.playIs("FLE$"):
		state.Play = &Play{
			Type:     FoulFlyError,
			Fielders: pp.getFielders(0),
			FieldingError: &FieldingError{
				Fielder: pp.getFielder(0),
			},
		}
	case pp.playIs("NP"):
		state.Play = &Play{
			Type: NoPlay,
		}
		// no play
	default:
		return fmt.Errorf("%s: unknown play %s", play.GetPos(), play.GetCode())
	}
	return nil
}

func (m *gameMachine) handleStolenBase(play gamefile.Play, state *State, eventMatches []string) error {
	if state.LastState == nil {
		return fmt.Errorf("%s: cannot steal bases at the start of a half-inning", play.GetPos())
	}
	for i := range eventMatches {
		base := eventMatches[i]
		var runner PlayerID
		switch base {
		case "2":
			m.impliedAdvance(play, state, "1-2")
			runner = state.LastState.Runners[0]
		case "3":
			m.impliedAdvance(play, state, "2-3")
			runner = state.LastState.Runners[1]
		case "H":
			m.impliedAdvance(play, state, "3-H")
			runner = state.LastState.Runners[2]
		default:
			return fmt.Errorf("%s: unknown stolen base code", play.GetPos())
		}
		state.Play.StolenBases = append(state.Play.StolenBases, base)
		if runner == "" {
			return fmt.Errorf("%s: no runner can steal %s", play.GetPos(), base)
		}
		state.Play.Runners[i] = runner
	}
	return nil
}

func (m *gameMachine) putOut(base string) {
	if m.basePutOuts == nil {
		m.basePutOuts = make(map[string]bool)
	}
	m.basePutOuts[base] = true
}

func (m *gameMachine) moveRunners(play gamefile.Play, state *State) error {
	for _, base := range []string{"3", "2", "1", "B"} {
		advance := state.Advances[base]
		if advance == nil {
			if m.basePutOuts[base] {
				// runner was put out
			} else if base != "B" && state.LastState != nil {
				// runner did not move
				number := BaseNumber[base]
				state.Runners[number] = state.LastState.Runners[number]
			}
			continue
		}
		from := BaseNumber[advance.From]
		to := BaseNumber[advance.To]
		switch {
		case state.LastState == nil && advance.From != "B":
			return fmt.Errorf("%s: cannot advance a runner from %s to %s at start of half-inning",
				play.GetPos(), advance.From, advance.To)
		case advance.From != "B" && state.LastState != nil && state.LastState.Runners[from] == "":
			return fmt.Errorf("%s: cannot advance non-existent runner from %s",
				play.GetPos(), advance.From)
		case advance.Out:
			state.recordOut()
			if advance.From != "B" {
				state.Runners[from] = ""
			}
		case advance.To == "H":
			if advance.From == "B" {
				m.scoreRun(state, state.Batter)
			} else {
				m.scoreRun(state, state.LastState.Runners[from])
			}
		case advance.From == "B":
			if state.Runners[to] != "" {
				return fmt.Errorf("%s: cannot advance runner %s to %d because it's already occupied by %s",
					play.GetPos(), state.Batter, to+1, state.Runners[to])
			}
			state.Runners[to] = state.Batter
		default:
			if state.Runners[to] != "" {
				return fmt.Errorf("%s: cannot advance runner %s to %d because it's already occupied by %s",
					play.GetPos(), state.LastState.Runners[from], to+1, state.Runners[to])
			}
			state.Runners[to] = state.LastState.Runners[from]
		}
	}
	return nil
}

func (m *gameMachine) scoreRun(state *State, runner PlayerID) {
	state.Score++
	state.ScoringRunners = append(state.ScoringRunners, runner)
}
