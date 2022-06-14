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
	state        *State
	lastState    *State
	PA           int
	pitcher      PlayerID
	basePutOuts  map[string]bool
	final        bool
	modifiers    Modifiers
}

func newGameMachine(half Half, lastState *State, battingTeam, fieldingTeam *Team) *gameMachine {
	if lastState == nil {
		lastState = &State{
			InningNumber: 1,
			Half:         half,
			Runners:      make([]PlayerID, 3),
		}
	}
	m := &gameMachine{
		battingTeam:  battingTeam,
		fieldingTeam: fieldingTeam,
		lastState:    lastState,
	}
	return m
}

func (m *gameMachine) handlePlay(play *gamefile.Play) (*State, error) {
	m.state = &State{
		InningNumber: m.lastState.InningNumber,
		Outs:         m.lastState.Outs,
		Half:         m.lastState.Half,
		Score:        m.lastState.Score,
		Pitcher:      m.pitcher,
		Runners:      make([]PlayerID, 3),
	}
	if m.state.Outs == 3 {
		m.state.Outs = 0
		m.state.InningNumber++
	}
	if play.ContinuedPlateAppearance {
		if m.lastState == nil {
			return nil, fmt.Errorf("%s: ... can only be used to continue a plate appearance", play.Pos)
		}
		m.state.Pitches = m.lastState.Pitches + Pitches(play.PitchSequence)
		m.state.Batter = m.lastState.Batter
	} else {
		m.state.Batter = m.battingTeam.parsePlayerID(play.Batter.String())
		m.state.Pitches = Pitches(play.PitchSequence)
	}
	if m.state.Batter == "" {
		return nil, fmt.Errorf("%s: no batter for %s", play.Pos, play.Code)
	}
	m.state.EventCode = play.Code
	if m.state.EventCode == "" {
		return nil, fmt.Errorf("empty event code in %s", play.Code)
	}
	m.state.Comment = play.Comment
	m.basePutOuts = nil
	if err := m.parseAdvances(play); err != nil {
		return nil, err
	}
	if err := m.handlePlayCode(play); err != nil {
		return nil, err
	}
	if err := m.moveRunners(play); err != nil {
		return nil, err
	}
	if m.state.Outs == 3 && !m.state.Complete {
		// inning ended
		m.state.Incomplete = true
	}
	if m.state.Complete {
		if !strings.HasSuffix(string(m.state.Pitches), "X") &&
			m.state.Play.BallInPlay() {
			m.state.Pitches += "X"
		}
		m.PA++
		m.state.Modifiers = m.modifiers
	}
	m.state.PlateAppearance.Number = m.PA
	m.lastState = m.state
	return m.state, nil
}

func (m *gameMachine) parseAdvances(play *gamefile.Play) error {
	var runners []PlayerID
	if m.lastState.InningNumber == m.state.InningNumber {
		runners = m.lastState.Runners
	}
	var err error
	m.state.Advances, err = parseAdvances(play, m.state.Batter, runners)
	return err
}

func (m *gameMachine) handleSpecialEvent(event *gamefile.Event) error {
	if event.Pitcher != "" {
		m.pitcher = m.fieldingTeam.parsePlayerID(event.Pitcher)
	}
	if event.Score != "" {
		if m.lastState.Outs != 3 {
			return fmt.Errorf("%s: the inning with %d outs has not ended after %s",
				event.Pos, m.lastState.Outs, m.lastState.EventCode)
		}
		score, err := strconv.Atoi(event.Score)
		if err != nil || m.state.Score != score {
			return fmt.Errorf("%s: in inning %d # runs is %d not %s", event.Pos,
				m.state.InningNumber, m.state.Score, event.Score)
		}
	}
	if event.Final != "" {
		score, err := strconv.Atoi(event.Final)
		if err != nil || m.state.Score != score {
			return fmt.Errorf("%s: in inning %d final score is %d not %s", event.Pos,
				m.state.InningNumber, m.state.Score, event.Score)
		}
		m.final = true
	}
	if event.RAdjRunner != "" {
		runner := m.battingTeam.parsePlayerID(event.RAdjRunner)
		base := event.RAdjBase
		if runner == "" || !(base == "1" || base == "2" || base == "3") {
			return fmt.Errorf("%s: invalid base %s for radj", event.Pos, event.RAdjBase)
		}
		if m.state.Outs != 3 {
			return fmt.Errorf("%s: radj must be at the inning start", event.Pos)
		}
		m.lastState = &State{
			InningNumber: m.state.InningNumber + 1,
			Half:         m.state.Half,
			Outs:         0,
			Score:        m.state.Score,
			Runners:      make([]PlayerID, 3),
		}
		m.lastState.Runners[BaseNumber[base]] = runner
	}
	return nil
}

func (m *gameMachine) impliedAdvance(play *gamefile.Play, code string) {
	advance, err := parseAdvance(play, code)
	if err != nil {
		panic(err)
	}
	if m.state.Advances[advance.From] == nil {
		m.state.Advances[advance.From] = advance
	}
	if m.state.Advances[advance.From].To == advance.To {
		m.state.Advances[advance.From].Implied = true
	}
}

func (m *gameMachine) handlePlayCode(play *gamefile.Play) error {
	pp := playCodeParser{}
	pp.parsePlayCode(play.Code)
	m.modifiers = pp.modifiers
	switch {
	case pp.playIs("$"):
		m.state.Play = &Play{
			Type:     FlyOut,
			Fielders: pp.getFielders(0),
		}
		m.state.recordOut()
		m.state.Complete = true
	case pp.playIs("$$"):
		fallthrough
	case pp.playIs("$$$"):
		m.state.Play = &Play{
			Type: GroundOut,
		}
		for i := range pp.playMatches {
			m.state.Play.Fielders = append(m.state.Play.Fielders, pp.getFielder(i))
		}
		m.state.recordOut()
		m.state.Complete = true
	case pp.playIs("K"):
		m.state.Play = &Play{
			Type: StrikeOut,
		}
		m.state.recordOut()
		m.state.Complete = true
	case pp.playIs("K+SB%"):
		m.state.Play = &Play{
			Type:    StrikeOut,
			Runners: make([]PlayerID, 1),
		}
		m.state.recordOut()
		m.state.Complete = true
		if err := m.handleStolenBase(play, pp.playMatches); err != nil {
			return err
		}
	case pp.playIs("W+WP"):
		m.state.Play = &Play{
			Type: WalkWildPitch,
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("W+PB"):
		m.state.Play = &Play{
			Type: WalkPassedBall,
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("W"):
		m.state.Play = &Play{
			Type: Walk,
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("SB%;SB%;SB%") || pp.playIs("SB%;SB%") || pp.playIs("SB%"):
		m.state.Play = &Play{
			Type:    StolenBase,
			Runners: make([]PlayerID, len(pp.playMatches)),
		}
		if err := m.handleStolenBase(play, pp.playMatches); err != nil {
			return err
		}
	case pp.playIs("K2$"):
		m.state.Play = &Play{
			Type:     StrikeOut,
			Fielders: []int{2, pp.getFielder(0)},
		}
		m.state.recordOut()
		m.state.Complete = true
	case pp.playIs("K+PB"):
		m.state.Play = &Play{
			Type:     StrikeOutPassedBall,
			Fielders: []int{2},
		}
		m.state.Complete = true
		m.impliedAdvance(play, "B-1")
	case pp.playIs("K+WP"):
		m.state.Play = &Play{
			Type:     StrikeOutWildPitch,
			Fielders: []int{1},
		}
		m.state.Complete = true
		m.impliedAdvance(play, "B-1")
	case pp.playIs("S$"):
		m.state.Play = &Play{
			Type: Single,
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("D$"):
		m.state.Play = &Play{
			Type: Double,
		}
		m.impliedAdvance(play, "B-2")
		m.state.Complete = true
	case pp.playIs("DGR"):
		m.state.Play = &Play{
			Type: Double,
		}
		m.impliedAdvance(play, "B-2")
		m.state.Complete = true
	case pp.playIs("T$"):
		m.state.Play = &Play{
			Type: Triple,
		}
		m.impliedAdvance(play, "B-3")
		m.state.Complete = true
	case pp.playIs("H"):
		m.state.Play = &Play{
			Type: HomeRun,
		}
		m.impliedAdvance(play, "B-H")
		m.state.Complete = true
	case pp.playIs("PB"):
		m.state.Play = &Play{
			Type: PassedBall,
		}
		// movement in advances
	case pp.playIs("WP"):
		m.state.Play = &Play{
			Type: WildPitch,
		}
		// movement in advances
	case pp.playIs("HP"):
		m.state.Play = &Play{
			Type: HitByPitch,
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("E$"):
		fe, err := parseFieldingError(play, pp.playCode)
		if err != nil {
			return fmt.Errorf("cannot parse fielding error in %s - %w", pp.playCode, err)
		}
		m.state.Play = &Play{
			Type:          ReachedOnError,
			Fielders:      pp.getFielders(0),
			FieldingError: fe,
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("C/E$"):
		m.state.Play = &Play{
			Type: CatcherInterference,
			FieldingError: &FieldingError{
				Fielder: pp.getFielder(0),
			},
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("PO%(E$)"):
		m.state.Play = &Play{
			Type:     PickedOff,
			Fielders: pp.getFielders(1),
			FieldingError: &FieldingError{
				Fielder: pp.getFielder(1),
			},
		}
		m.state.NotOutOnPlay = true
	case pp.playIs("PO%($$)"):
		from := pp.playMatches[0]
		if !(from == "1" || from == "2" || from == "3") {
			return fmt.Errorf("illegal picked off base in %s", pp.playCode)
		}
		runner, err := m.getBaseRunner(from)
		if err != nil {
			return fmt.Errorf("cannot pick off in %s - %w", pp.playCode, err)
		}
		m.state.Play = &Play{
			Type:     PickedOff,
			Runners:  []PlayerID{runner},
			Fielders: pp.getFielders(1, 2),
		}
		advance := m.state.Advances[from]
		if advance == nil {
			m.state.recordOut()
			m.putOut(from)
		} else {
			m.state.NotOutOnPlay = advance.FieldingError != nil
		}
	case pp.playIs("FC$"):
		// outs are in the advance, if any
		m.state.Play = &Play{
			Type:     FieldersChoice,
			Fielders: pp.getFielders(0),
		}
		m.impliedAdvance(play, "B-1")
		m.state.Complete = true
	case pp.playIs("$$(%)$") || pp.playIs("$(%)$$"):
		if !m.modifiers.Contains("GDP") {
			return fmt.Errorf("play should contain GDP modifier in %s", pp.playCode)
		}
		base := pp.playMatches[2]
		runner, err := m.getBaseRunner(base)
		if err != nil {
			return fmt.Errorf("no runner in double play %s - %w", pp.playCode, err)
		}
		m.state.Play = &Play{
			Type:    DoublePlay,
			Runners: []PlayerID{runner},
		}
		// should pass fielders to record out to do assists
		m.state.recordOut()
		m.state.recordOut()
		m.putOut(base)
		m.state.Complete = true
	case pp.playIs("$(B)$(%)"):
		fallthrough
	case pp.playIs("$(B)$$(%)"):
		fallthrough
	case pp.playIs("$(B)$$$(%)"):
		if !m.modifiers.Contains("LDP", "FDP") {
			return fmt.Errorf("play should contain LDP or FDP modifier in %s (%v)", pp.playCode, m.state.Modifiers)
		}
		base := pp.playMatches[len(pp.playMatches)-1]
		runner, err := m.getBaseRunner(base)
		if err != nil {
			return fmt.Errorf("no runner in lineout double play %s - %w", pp.playCode, err)
		}
		m.state.Play = &Play{
			Type:    DoublePlay,
			Runners: []PlayerID{runner},
		}
		m.state.recordOut()
		m.state.recordOut()
		m.putOut(base)
		m.state.Complete = true
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
		advance := m.state.Advances[from]
		if advance == nil {
			m.state.recordOut()
			m.putOut(from)
		} else {
			m.state.NotOutOnPlay = advance.FieldingError != nil
		}
		runner, err := m.getBaseRunner(from)
		if err != nil {
			return fmt.Errorf("%s: cannot catch stealing runner in %s - %w", play.Pos, pp.playCode, err)
		}
		m.state.Play = &Play{
			Type:    CaughtStealing,
			Runners: []PlayerID{runner},
			Base:    to,
			//Fielders: ,
		}
	case pp.playIs("FLE$"):
		m.state.Play = &Play{
			Type:     FoulFlyError,
			Fielders: pp.getFielders(0),
			FieldingError: &FieldingError{
				Fielder: pp.getFielder(0),
			},
		}
	case pp.playIs("NP"):
		m.state.Play = &Play{
			Type: NoPlay,
		}
		// no play
	default:
		return fmt.Errorf("%s: unknown play %s", play.Pos, play.Code)
	}
	return nil
}

func (m *gameMachine) handleStolenBase(play *gamefile.Play, eventMatches []string) error {
	for i := range eventMatches {
		base := eventMatches[i]
		var runner PlayerID
		switch base {
		case "2":
			m.impliedAdvance(play, "1-2")
			runner = m.lastState.Runners[0]
		case "3":
			m.impliedAdvance(play, "2-3")
			runner = m.lastState.Runners[1]
		case "H":
			m.impliedAdvance(play, "3-H")
			runner = m.lastState.Runners[2]
		default:
			return fmt.Errorf("%s: unknown stolen base code", play.Pos)
		}
		m.state.Play.StolenBases = append(m.state.Play.StolenBases, base)
		if runner == "" {
			return fmt.Errorf("%s: no runner can steal %s", play.Pos, base)
		}
		m.state.Play.Runners[i] = runner
	}
	return nil
}

func (m *gameMachine) getBaseRunner(base string) (runner PlayerID, err error) {
	if base == "H" {
		err = fmt.Errorf("a runner cannot be at H")
		return
	}
	if m.lastState.InningNumber != m.state.InningNumber {
		err = fmt.Errorf("no runners are on base at the start of a half-inning")
		return
	}
	runner = m.lastState.Runners[runnerNumber[base]]
	if runner == "" {
		err = fmt.Errorf("no runner on %s", base)
	}
	return
}

func (m *gameMachine) putOut(base string) {
	if m.basePutOuts == nil {
		m.basePutOuts = make(map[string]bool)
	}
	m.basePutOuts[base] = true
}

func (m *gameMachine) moveRunners(play *gamefile.Play) error {
	for _, base := range []string{"3", "2", "1", "B"} {
		advance := m.state.Advances[base]
		if advance == nil {
			if m.basePutOuts[base] {
				// runner was put out
			} else if base != "B" && m.lastState.InningNumber == m.state.InningNumber {
				// runner did not move
				number := BaseNumber[base]
				m.state.Runners[number] = m.lastState.Runners[number]
			}
			continue
		}
		from := BaseNumber[advance.From]
		to := BaseNumber[advance.To]
		switch {
		case m.lastState.InningNumber != m.state.InningNumber && advance.From != "B":
			return fmt.Errorf("%s: cannot advance a runner from %s to %s at start of half-inning",
				play.Pos, advance.From, advance.To)
		case advance.From != "B" && m.lastState.Runners[from] == "":
			return fmt.Errorf("%s: cannot advance non-existent runner from %s",
				play.Pos, advance.From)
		case advance.Out:
			m.state.recordOut()
			if advance.From != "B" {
				m.state.Runners[from] = ""
			}
		case advance.To == "H":
			if advance.From == "B" {
				m.scoreRun(m.state.Batter)
			} else {
				m.scoreRun(m.lastState.Runners[from])
			}
		case advance.From == "B":
			if m.state.Runners[to] != "" {
				return fmt.Errorf("%s: cannot advance runner %s to %d because it's already occupied by %s",
					play.Pos, m.state.Batter, to+1, m.state.Runners[to])
			}
			m.state.Runners[to] = m.state.Batter
		default:
			if m.state.Runners[to] != "" {
				return fmt.Errorf("%s: cannot advance runner %s to %d because it's already occupied by %s",
					play.Pos, m.lastState.Runners[from], to+1, m.state.Runners[to])
			}
			m.state.Runners[to] = m.lastState.Runners[from]
		}
	}
	return nil
}

func (m *gameMachine) scoreRun(runner PlayerID) {
	m.state.Score++
	m.state.ScoringRunners = append(m.state.ScoringRunners, runner)
}
