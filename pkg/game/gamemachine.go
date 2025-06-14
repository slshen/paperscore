package game

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/slshen/paperscore/pkg/gamefile"
)

type gameMachine struct {
	battingTeam  *Team
	fieldingTeam *Team
	pitcher      PlayerID
	basePutOuts  map[string]bool
	final        bool
	modifiers    Modifiers
}

func newGameMachine(battingTeam, fieldingTeam *Team) *gameMachine {
	m := &gameMachine{
		battingTeam:  battingTeam,
		fieldingTeam: fieldingTeam,
	}
	return m
}

func (m *gameMachine) newState(pos lexer.Position, lastState *State) *State {
	state := &State{
		Pos:          FileLocation{Filename: pos.Filename, Line: pos.Line},
		BattingTeam:  m.battingTeam,
		FieldingTeam: m.fieldingTeam,
		InningNumber: lastState.InningNumber,
		Outs:         lastState.Outs,
		Half:         lastState.Half,
		Score:        lastState.Score,
		Pitcher:      m.pitcher,
		Runners:      [3]PlayerID{},
		Defense:      lastState.Defense,
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
			BattingTeam:  m.battingTeam,
			FieldingTeam: m.fieldingTeam,
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
	for _, p := range alt.Credit {
		player := m.battingTeam.Players[m.battingTeam.parsePlayerID(p)]
		if player == nil {
			player = m.fieldingTeam.Players[m.fieldingTeam.parsePlayerID(p)]
		}
		if player == nil {
			return nil, NewError("no player %s for alt credit on either team", alt.Pos, p)
		}
		state.AlternativeCredits = append(state.AlternativeCredits, player)
	}
	err := m.handlePlay(alt, state)
	return state, err
}

func (m *gameMachine) handleActualPlay(play *gamefile.ActualPlay, lastState *State) (*State, error) {
	state := m.newState(play.Pos, lastState)
	state.PlateAppearance.Number = play.PlateAppearance.Int()
	if play.ContinuedPlateAppearance {
		if state.LastState == nil {
			return nil, NewError("... can only be used to continue a plate appearance", play.GetPos())
		}
		state.Pitches = state.LastState.Pitches + Pitches(play.PitchSequence)
		state.Batter = state.LastState.Batter
	} else {
		state.Batter = m.battingTeam.parsePlayerID(play.Batter)
		state.Pitches = Pitches(play.PitchSequence)
	}
	if state.Batter == "" {
		return nil, NewError("no batter for %s", play.GetPos(), play.GetCode())
	}
	err := m.handlePlay(play, state)
	for _, after := range play.Afters {
		// handle subs for runners on base
		var runnerEnter, runnerExit PlayerID
		if after.CourtesyRunner != nil {
			runnerEnter = m.battingTeam.parsePlayerID(*after.CourtesyRunner)
			if after.CourtesyRunnerFor == nil {
				runnerExit = state.Batter
			} else {
				runnerExit = m.battingTeam.parsePlayerID(*after.CourtesyRunnerFor)
			}
		}
		if after.Sub != nil {
			runnerEnter = m.battingTeam.parsePlayerID(after.Sub.Enter)
			runnerExit = m.battingTeam.parsePlayerID(after.Sub.Exit)
		}
		for i := range state.Runners {
			if state.Runners[i] == runnerExit {
				state.Runners[i] = runnerEnter
				break
			}
		}
	}
	return state, err
}

func (m *gameMachine) handlePlay(play gamefile.Play, state *State) error {
	state.PlayCode = play.GetCode()
	state.AdvancesCodes = play.GetAdvances()
	if state.PlayCode == "" {
		return NewError("empty event code in %s", play.GetPos(), play.GetCode())
	}
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
		known, _, balls, strikes := state.Pitches.Count()
		if known && state.Play.IsBallInPlay() {
			if strikes > 2 {
				return NewError("cannot put ball in play with %d strikes (%s)", state.Pos, strikes, play.GetCode())
			}
			if balls > 3 {
				return NewError("cannot put ball in play with %d balls (%s)", state.Pos, balls, play.GetCode())
			}
		}
		if known && state.Play.IsStrikeOut() {
			if state.Pitches.Last() == 'X' {
				return NewError("strike out pitch sequence should not end in X", state.Pos)
			}
			if strikes != 3 {
				return NewError("must strike out with 3 strikes", state.Pos)
			}
			if balls > 3 {
				return NewError("cannot strike out with more than 3 balls", state.Pos)
			}
		}
		if known && state.Play.IsWalk() {
			if state.Pitches.Last() == 'X' {
				return NewError("walk pitch sequence should not end in X", state.Pos)
			}
			if strikes > 2 {
				return NewError("cannot walk with more than 2 strikes", state.Pos)
			}
			if balls != 4 {
				return NewError("must walk with 4 balls", state.Pos)
			}
		}
		if known && state.Pitches.Last() != 'X' &&
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
		state.Defense[0] = m.pitcher
	}
	if event.PlayerName != nil {
		playerID := m.battingTeam.parsePlayerID(event.PlayerName.Player)
		player := m.battingTeam.GetPlayer(playerID)
		player.Name = event.PlayerName.GetName()
	}
	if len(event.Defense) > 0 {
		state = state.Copy()
		for _, pp := range event.Defense {
			state.Defense[pp.PositionNumber()-1] = m.fieldingTeam.parsePlayerID(pp.Player)
			if pp.PositionNumber() == 1 {
				m.pitcher = state.Defense[0]
			}
		}
		return state, nil
	}
	if event.DefenseSub != nil {
		enter := m.fieldingTeam.parsePlayerID(event.DefenseSub.Enter)
		exit := m.fieldingTeam.parsePlayerID(event.DefenseSub.Exit)
		for _, pp := range event.Defense {
			player := m.fieldingTeam.parsePlayerID(pp.Player)
			if player == exit {
				state = state.Copy()
				state.Defense[pp.PositionNumber()-1] = enter
				return state, nil
			}
		}
		return nil, NewError("cannot sub %s for %s because %s is not in the field", event.Pos, enter, exit, exit)
	}
	if event.Score != "" {
		if state.Outs != 3 {
			return nil, NewError("the inning with %d outs has not ended after %s",
				event.Pos, state.Outs, state.PlayCode)
		}
		score, err := strconv.Atoi(event.Score)
		if err != nil || state.Score != score {
			return nil, NewError("in inning %d # runs is %d not %s", event.Pos,
				state.InningNumber, state.Score, event.Score)
		}
	}
	if event.Final != "" {
		score, err := strconv.Atoi(event.Final)
		if err != nil || state.Score != score {
			return nil, NewError("in inning %d final score is %d not %s", event.Pos,
				state.InningNumber, state.Score, event.Score)
		}
		m.final = true
	}
	if event.RAdjRunner != "" {
		runner := m.battingTeam.parsePlayerID(event.RAdjRunner.String())
		base := event.RAdjBase
		if runner == "" || !(base == "1" || base == "2" || base == "3") {
			return nil, NewError("invalid base %s for radj", event.Pos, event.RAdjBase)
		}
		if state.Outs != 3 {
			return nil, NewError("radj must be at the inning start", event.Pos)
		}
		lastState := &State{
			BattingTeam:  m.battingTeam,
			FieldingTeam: m.fieldingTeam,
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
			// verify that we're only scoring a SacrificeFly if a runner scores
			ok := false
			for _, adv := range state.Advances {
				if adv.To == "H" && !adv.Out {
					ok = true
				}
			}
			if !ok {
				return NewError("cannot score SacrificeFly unless a runner scores", play.GetPos())
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
	case pp.playIs("K+CS%($$)"):
		state.Play = Play{
			Type:     StrikeOutCaughtStealing,
			Fielders: pp.getFielders(1, 2),
		}
		state.Complete = true
		state.recordOut()
		return m.handleCaughtStealing(play, state, pp, NoError)
	case pp.playIs("K+PO%($$)") || pp.playIs("K+PO%(E$)"):
		from := pp.playMatches[0]
		if !(from == "1" || from == "2" || from == "3") {
			return NewError("illegal picked off base in %s", play.GetPos(), pp.playCode)
		}
		_, err := state.GetBaseRunner(from)
		if err != nil {
			return NewError("cannot pick off in %s - %w", play.GetPos(), pp.playCode, err)
		}
		state.Play = Play{
			Type: StrikeOutPickedOff,
		}
		state.Complete = true
		state.recordOut()
		if strings.Contains(pp.playCode, "(E") {
			state.Play.NotOutOnPlay = true
			state.Play.FieldingError = FieldingError{
				Fielder: pp.getFielder(1),
			}
		} else {
			state.Play.Fielders = pp.getFielders(1, 2)
			advance := state.Advances.From(from)
			if advance == nil {
				state.recordOut()
				m.putOut(from)
			} else {
				return NewError("picked off runner on %s cannot advance", play.GetPos(), from)
			}
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
			Type: GroundRuleDouble,
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
		if actualPlay, ok := play.(*gamefile.ActualPlay); ok {
			if !strings.HasSuffix(actualPlay.PitchSequence, "H") {
				return NewError("HP pitch sequence %s should end with H", play.GetPos(),
					actualPlay.PitchSequence)
			}
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("E$"):
		fe, err := parseFieldingError(play, pp.playCode)
		if err != nil {
			return NewError("cannot parse fielding error in %s - %w", play.GetPos(),
				pp.playCode, err)
		}
		state.Play = Play{
			Type:          ReachedOnError,
			Fielders:      pp.getFielders(0),
			FieldingError: fe,
		}
		m.impliedAdvance(play, state, "B-1")
		state.Complete = true
	case pp.playIs("C"):
		var fielder int
		e := regexp.MustCompile(`E([1-9])`)
		for _, modifier := range pp.modifiers {
			m := e.FindStringSubmatch(modifier)
			if m != nil {
				fielder = fielderNumber[m[1]]
			}
		}
		if fielder == 0 {
			return NewError("no fielder in catcher's interference", play.GetPos())
		}
		state.Play = Play{
			Type: CatcherInterference,
			FieldingError: FieldingError{
				Fielder: fielder,
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
	case pp.playIs("$(%)$$"):
		return m.handleGroundBallDoublePlay(play, state, pp, pp.playMatches[1])
	case pp.playIs("$$(%)$") || pp.playIs("$$(%)$$"):
		return m.handleGroundBallDoublePlay(play, state, pp, pp.playMatches[2])
	case pp.playIs("$(B)$(%)") || pp.playIs("$(B)$$(%)") ||
		pp.playIs("$(B)$$$(%)"):
		if !m.modifiers.Contains("LDP", "FDP") {
			return NewError("play should contain LDP or FDP modifier in %s (%v)",
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
		return m.handleCaughtStealing(play, state, pp, fieldingError)
	case pp.playIs("CS%($$)") || pp.playIs("CS%($$$)") || pp.playIs("CS%($$$$)"):
		return m.handleCaughtStealing(play, state, pp, NoError)
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
		return NewError("unknown play %s", play.GetPos(), play.GetCode())
	}
	return nil
}

func (m *gameMachine) handleGroundBallDoublePlay(play gamefile.Play, state *State, pp playCodeParser, runnerBase string) error {
	if !m.modifiers.Contains("GDP") {
		return NewError("play should contain GDP modifier in %s", play.GetPos(), pp.playCode)
	}
	_, err := state.GetBaseRunner(runnerBase)
	if err != nil {
		return NewError("no runner in double play %s - %w", play.GetPos(), pp.playCode, err)
	}
	state.Play = Play{
		Type: DoublePlay,
	}
	nextBase := NextBase[runnerBase]
	if nextBase == "" {
		return NewError("double play runner cannot be at %s", play.GetPos(), runnerBase)
	}
	paren := strings.IndexRune(pp.playCode, '(')
	fielders := pp.playCode[0:paren]
	m.impliedAdvance(play, state, fmt.Sprintf("%sX%s(%s)", runnerBase, nextBase, fielders))
	// should pass fielders to record out to do assists
	state.recordOut()
	// m.putOut(runnerBase)
	state.Complete = true
	return nil
}

func (m *gameMachine) handleCaughtStealing(play gamefile.Play, state *State, pp playCodeParser, fieldingError FieldingError) error {
	to := pp.playMatches[0]
	if !(to == "2" || to == "3" || to == "H") {
		return fmt.Errorf("illegal caught stealing base code %s", pp.playCode)
	}
	from := PreviousBase[to]
	advance := state.Advances.From(from)
	runner, err := state.GetBaseRunner(from)
	if err != nil {
		return NewError("cannot catch stealing runner in %s - %w", play.GetPos(), pp.playCode, err)
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
		return NewError("illegal picked off base %s", play.GetPos(), from)
	}
	runner, err := state.GetBaseRunner(from)
	if err != nil {
		return NewError("cannot pick off runner - %w", play.GetPos(), err)
	}
	state.Play = Play{
		Type:            playType,
		Fielders:        pp.getAllFielders(1),
		PickedOffRunner: runner,
		FieldingError:   fieldingError,
	}
	advance := state.Advances.From(from)
	state.NotOutOnPlay = (advance != nil && advance.IsFieldingError()) || fieldingError.IsFieldingError()
	if !state.NotOutOnPlay {
		state.recordOut()
		m.putOut(from)
	}
	return nil
}

func (m *gameMachine) handleStolenBase(play gamefile.Play, state *State, eventMatches []string) error {
	if state.LastState == nil {
		return NewError("cannot steal bases at the start of a half-inning", play.GetPos())
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
			return NewError("unknown stolen base code", play.GetPos())
		}
		adv.Runner = runner
		adv.Steal = true
		state.Play.StolenBases = append(state.Play.StolenBases, base)
		if runner == "" {
			return NewError("no runner can steal %s", play.GetPos(), base)
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
			return NewError("cannot advance a runner from %s to %s at start of half-inning",
				play.GetPos(), advance.From, advance.To)
		case advance.From != "B" && state.LastState != nil && state.LastState.Runners[from] == "":
			return NewError("cannot advance non-existent runner from %s",
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
				return NewError("cannot advance batter-runner %s to %d because it's already occupied by %s",
					play.GetPos(), state.Batter, to+1, state.Runners[to])
			}
			state.Runners[to] = state.Batter
		default:
			if state.Runners[to] != "" && !isFieldersChoice3rdOut(state) {
				return NewError("cannot advance runner %s to %d because it's already occupied by %s",
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
