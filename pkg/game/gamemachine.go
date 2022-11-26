package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/slshen/sb/pkg/gamefile"
)

type gameMachine struct {
	battingTeam  *Team
	fieldingTeam *Team
	pitcher      PlayerID
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

func (m *gameMachine) newState(pos lexer.Position, lastState *State) *State {
	state := &State{
		Pos:          FileLocation{Filename: pos.Filename, Line: pos.Line},
		InningNumber: lastState.InningNumber,
		Outs:         lastState.Outs,
		Half:         lastState.Half,
		Score:        lastState.Score,
		Pitcher:      m.pitcher,
		Runners:      [3]PlayerID{},
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
			// Runners:      make([]PlayerID, 3),
		}
		realLastState.Batter = lastState.Batter
	}
	state := m.newState(alt.Pos, realLastState)
	state.Batter = lastState.Batter
	state.Pitches = lastState.Pitches
	state.AlternativeFor = lastState
	err := m.handlePlay(alt, state)
	return state, err
}

func (m *gameMachine) handleActualPlay(play *gamefile.ActualPlay, lastState *State) (*State, error) {
	state := m.newState(play.Pos, lastState)
	state.PlateAppearance.Number = play.PlateAppearance.Int()
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
		// verify that we've struck out with 3 strikes, or walked with 4 balls
		// or that we put the ball in play without walking or striking out
		_, balls, strikes := state.Pitches.Count()
		if state.Play.IsBallInPlay() {
			if strikes > 2 {
				return fmt.Errorf("%s: cannot put ball in play with %d strikes (%s)", state.Pos, strikes, play.GetCode())
			}
			if balls > 3 {
				return fmt.Errorf("%s: cannot put ball in play with %d balls (%s)", state.Pos, balls, play.GetCode())
			}
		}
		if state.Play.IsStrikeOut() {
			if state.Pitches[len(state.Pitches)-1] == 'X' {
				return fmt.Errorf("%s: strike out pitch sequence should not end in X", state.Pos)
			}
			if strikes != 3 {
				return fmt.Errorf("%s: must strike out with 3 strikes", state.Pos)
			}
			if balls > 3 {
				return fmt.Errorf("%s: cannot strike out with more than 3 balls", state.Pos)
			}
		}
		if state.Play.IsWalk() {
			if state.Pitches[len(state.Pitches)-1] == 'X' {
				return fmt.Errorf("%s: walk pitch sequence should not end in X", state.Pos)
			}
			if strikes > 2 {
				return fmt.Errorf("%s: cannot walk with more than 2 strikes", state.Pos)
			}
			if balls != 4 {
				return fmt.Errorf("%s: must walk with 4 balls", state.Pos)
			}
		}
		if !strings.HasSuffix(string(state.Pitches), "X") &&
			state.Play.IsBallInPlay() {
			// fix up pitches
			state.Pitches += "X"
		}
		state.Modifiers = m.modifiers
	}
	return nil
}

func (m *gameMachine) parseAdvances(play gamefile.Play, state *State) error {
	var runners [3]PlayerID
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
			// Runners:      make([]PlayerID, 3),
		}
		// lastState.Runners = make([]PlayerID, 3)
		lastState.Runners[BaseNumber[base]] = runner
		return lastState, nil
	}
	return nil, nil
}

func (m *gameMachine) impliedAdvance(play gamefile.Play, state *State, code string) *Advance {
	impliedAdvance, err := parseAdvance(play, code)
	if err != nil {
		panic(err)
	}
	from := impliedAdvance.From
	advance := state.Advances.From(from)
	if advance == nil {
		advance = impliedAdvance
		state.Advances = append(state.Advances, impliedAdvance)
	}
	if advance.To == impliedAdvance.To {
		advance.Implied = true
	}
	return advance
}

func (m *gameMachine) handlePlayCode(play gamefile.Play, state *State) error {
	pp := playCodeParser{}
	pp.parsePlayCode(play.GetCode())
	m.modifiers = pp.modifiers
	switch {
	case pp.playIs("$"):
		state.Play = Play{
			Type:     FlyOut,
			Fielders: pp.getFielders(0),
		}
		if m.modifiers.Contains(SacrificeFly) {
			// verify that we're only scoring a SacraficeFly if a runner scores
			ok := false
			for _, adv := range state.Advances {
				if adv.To == "H" && !adv.Out {
					ok = true
				}
			}
			if !ok {
				return fmt.Errorf("%s: cannot score SacrificeFly unless a runner scores", play.GetPos())
			}
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("$$"):
		fallthrough
	case pp.playIs("$$$"):
		state.Play = Play{
			Type: GroundOut,
		}
		for i := range pp.playMatches {
			state.Play.Fielders = append(state.Play.Fielders, pp.getFielder(i))
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("K"):
		state.Play = Play{
			Type: StrikeOut,
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("K+SB%"):
		state.Play = Play{
			Type: StrikeOutStolenBase,
			// Runners: make([]PlayerID, 1),
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
		_, err := state.GetBaseRunner(from)
		if err != nil {
			return fmt.Errorf("%s: cannot pick off in %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = Play{
			Type:     StrikeOutPickedOff,
			Fielders: pp.getFielders(1, 2),
		}
		state.Complete = true
		state.recordOut()
		advance := state.Advances.From(from)
		if advance == nil {
			state.recordOut()
			m.putOut(from)
		} else {
			return fmt.Errorf("%s: picked off runner on %s cannot advance", play.GetPos(), from)
		}
	case pp.playIs("W+WP"):
		state.Play = Play{
			Type: WalkWildPitch,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("W+PB"):
		state.Play = Play{
			Type: WalkPassedBall,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("W+PO%($$)"):
		if err := m.handlePickedoff(play, state, pp, WalkPickedOff, NoError); err != nil {
			return err
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("W"):
		state.Play = Play{
			Type: Walk,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("W+SB%"):
		state.Play = Play{
			Type: Walk,
		}
		m.impliedAdvance(play, state, "B-1")
		if err := m.handleStolenBase(play, state, pp.playMatches); err != nil {
			return err
		}
		state.Complete = true
	case pp.playIs("SB%;SB%;SB%") || pp.playIs("SB%;SB%") || pp.playIs("SB%"):
		state.Play = Play{
			Type: StolenBase,
			// Runners: make([]PlayerID, len(pp.playMatches)),
		}
		if err := m.handleStolenBase(play, state, pp.playMatches); err != nil {
			return err
		}
	case pp.playIs("K2$") || pp.playIs("K2"):
		state.Play = Play{
			Type:     StrikeOut,
			Fielders: pp.getAllFielders(0),
		}
		state.recordOut()
		state.Complete = true
	case pp.playIs("K+PB"):
		state.Play = Play{
			Type:     StrikeOutPassedBall,
			Fielders: []int{2},
		}
		state.Complete = true
		m.impliedAdvance(play, state, "B-1")
	case pp.playIs("K+WP"):
		state.Play = Play{
			Type:     StrikeOutWildPitch,
			Fielders: []int{1},
		}
		state.Complete = true
		m.impliedAdvance(play, state, "B-1")
	case pp.playIs("S$"):
		state.Play = Play{
			Type: Single,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("D$"):
		state.Play = Play{
			Type: Double,
		}
		m.impliedAdvance(play, state, "B-2")
		state.Complete = true
	case pp.playIs("DGR"):
		state.Play = Play{
			Type: Double,
		}
		m.impliedAdvance(play, state, "B-2")
		state.Complete = true
	case pp.playIs("T$"):
		state.Play = Play{
			Type: Triple,
		}
		m.impliedAdvance(play, state, "B-3")
		state.Complete = true
	case pp.playIs("H$") || pp.playIs("H"):
		state.Play = Play{
			Type:     HomeRun,
			Fielders: pp.getAllFielders(0),
		}
		m.impliedAdvance(play, state, "B-H")
		state.Complete = true
	case pp.playIs("PB"):
		state.Play = Play{
			Type: PassedBall,
		}
		// movement in advances
	case pp.playIs("WP"):
		state.Play = Play{
			Type: WildPitch,
		}
		// movement in advances
	case pp.playIs("HP"):
		state.Play = Play{
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
		state.Play = Play{
			Type:          ReachedOnError,
			Fielders:      pp.getFielders(0),
			FieldingError: fe,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("C/E$"):
		state.Play = Play{
			Type: CatcherInterference,
			FieldingError: FieldingError{
				Fielder: pp.getFielder(0),
			},
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("PO%(E$)"):
		fieldingError := FieldingError{
			Fielder: pp.getFielder(1),
		}
		if err := m.handlePickedoff(play, state, pp, PickedOff, fieldingError); err != nil {
			return err
		}
	case pp.playIs("PO%($$)") || pp.playIs("PO%($$$)") || pp.playIs("PO%($$$$)"):
		if err := m.handlePickedoff(play, state, pp, PickedOff, NoError); err != nil {
			return err
		}
	case pp.playIs("FC$"):
		// outs are in the advance, if any
		state.Play = Play{
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
		_, err := state.GetBaseRunner(base)
		if err != nil {
			return fmt.Errorf("%s: no runner in double play %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = Play{
			Type: DoublePlay,
		}
		// should pass fielders to record out to do assists
		state.recordOut()
		state.recordOut()
		m.putOut(base)
		state.Complete = true
	case pp.playIs("$(B)$(%)") || pp.playIs("$(B)$$(%)") ||
		pp.playIs("$(B)$$$(%)"):
		if !m.modifiers.Contains("LDP", "FDP") {
			return fmt.Errorf("%s: play should contain LDP or FDP modifier in %s (%v)",
				play.GetPos(), pp.playCode, state.Modifiers)
		}
		base := pp.playMatches[len(pp.playMatches)-1]
		_, err := state.GetBaseRunner(base)
		if err != nil {
			return fmt.Errorf("no runner in lineout double play %s - %w", pp.playCode, err)
		}
		state.Play = Play{
			Type: DoublePlay,
			// Runners: []PlayerID{runner},
		}
		state.recordOut()
		state.recordOut()
		m.putOut(base)
		state.Complete = true
	case pp.playIs("CS%(E$)"):
		fieldingError := FieldingError{
			Fielder: pp.getFielder(1),
		}
		return m.handleCaughtStealing(play, state, pp, PickedOff, fieldingError)
	case pp.playIs("CS%($$)") || pp.playIs("CS%($$$)") || pp.playIs("CS%($$$$)"):
		return m.handleCaughtStealing(play, state, pp, CaughtStealing, NoError)
	case pp.playIs("FLE$"):
		state.Play = Play{
			Type:     FoulFlyError,
			Fielders: pp.getFielders(0),
			FieldingError: FieldingError{
				Fielder: pp.getFielder(0),
			},
		}
	case pp.playIs("NP"):
		state.Play = Play{
			Type: NoPlay,
		}
		// no play
	default:
		return fmt.Errorf("%s: unknown play %s", play.GetPos(), play.GetCode())
	}
	return nil
}

func (m *gameMachine) handleCaughtStealing(play gamefile.Play, state *State, pp playCodeParser, playType PlayType, fieldingError FieldingError) error {
	to := pp.playMatches[0]
	if !(to == "2" || to == "3" || to == "H") {
		return fmt.Errorf("illegal caught stealing base code %s", pp.playCode)
	}
	from := PreviousBase[to]
	advance := state.Advances.From(from)
	runner, err := state.GetBaseRunner(from)
	if err != nil {
		return fmt.Errorf("%s: cannot catch stealing runner in %s - %w", play.GetPos(), pp.playCode, err)
	}
	state.Play = Play{
		Type:                 CaughtStealing,
		CaughtStealingRunner: runner,
		CaughtStealingBase:   to,
		FieldingError:        fieldingError,
	}
	if advance == nil {
		state.recordOut()
		m.putOut(from)
	} else {
		state.NotOutOnPlay = advance.IsFieldingError() || fieldingError.IsFieldingError()
	}
	return nil
}

func (m *gameMachine) handlePickedoff(play gamefile.Play, state *State, pp playCodeParser, playType PlayType, fieldingError FieldingError) error {
	from := pp.playMatches[0]
	if !(from == "1" || from == "2" || from == "3") {
		return fmt.Errorf("%s: illegal picked off base %s", play.GetPos(), from)
	}
	runner, err := state.GetBaseRunner(from)
	if err != nil {
		return fmt.Errorf("%s: cannot pick off runner - %w", play.GetPos(), err)
	}
	state.Play = Play{
		Type:            playType,
		Fielders:        pp.getAllFielders(1),
		PickedOffRunner: runner,
		FieldingError:   fieldingError,
	}
	advance := state.Advances.From(from)
	if advance == nil {
		state.recordOut()
		m.putOut(from)
	} else {
		state.NotOutOnPlay = advance.IsFieldingError() || fieldingError.IsFieldingError()
	}
	return nil
}

func (m *gameMachine) handleStolenBase(play gamefile.Play, state *State, eventMatches []string) error {
	if state.LastState == nil {
		return fmt.Errorf("%s: cannot steal bases at the start of a half-inning", play.GetPos())
	}
	for i := range eventMatches {
		base := eventMatches[i]
		var (
			runner PlayerID
			adv    *Advance
		)
		switch base {
		case "2":
			adv = m.impliedAdvance(play, state, "1-2")
			runner = state.LastState.Runners[0]
		case "3":
			adv = m.impliedAdvance(play, state, "2-3")
			runner = state.LastState.Runners[1]
		case "H":
			adv = m.impliedAdvance(play, state, "3-H")
			runner = state.LastState.Runners[2]
		default:
			return fmt.Errorf("%s: unknown stolen base code", play.GetPos())
		}
		adv.Runner = runner
		adv.Steal = true
		state.Play.StolenBases = append(state.Play.StolenBases, base)
		if runner == "" {
			return fmt.Errorf("%s: no runner can steal %s", play.GetPos(), base)
		}
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
		advance := state.Advances.From(base)
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
			if state.Runners[to] != "" && !isFieldersChoice3rdOut(state) {
				return fmt.Errorf("%s: cannot advance batter-runner %s to %d because it's already occupied by %s",
					play.GetPos(), state.Batter, to+1, state.Runners[to])
			}
			state.Runners[to] = state.Batter
		default:
			if state.Runners[to] != "" && !isFieldersChoice3rdOut(state) {
				return fmt.Errorf("%s: cannot advance runner %s to %d because it's already occupied by %s",
					play.GetPos(), state.LastState.Runners[from], to+1, state.Runners[to])
			}
			state.Runners[to] = state.LastState.Runners[from]
		}
	}
	return nil
}

func isFieldersChoice3rdOut(state *State) bool {
	return state.Play.Type == FieldersChoice && state.Outs == 3
}

func (m *gameMachine) scoreRun(state *State, runner PlayerID) {
	state.Score++
	state.ScoringRunners = append(state.ScoringRunners, runner)
}
