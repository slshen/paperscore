package game

import (
	"fmt"
	"strconv"
	"strings"
)

type gameMachine struct {
	game                        *Game
	visitorPlay                 int
	homePlay                    int
	visitorPA                   int
	homePA                      int
	homePitcher, visitorPitcher PlayerID
	state                       *State
	lastState                   *State // last state, same side or nil if start of half-inning
	playCode                    string
	playFields                  []string
	basePutOuts                 map[string]bool
	eventCodeParser
}

func (m *gameMachine) run() error {
	m.visitorPA = 1
	m.homePA = 1
	m.state = &State{
		InningNumber: 1,
		Half:         Top,
	}
	m.state.init()
	var visitorDone, homeDone bool
	for {
		if len(m.game.states) > 0 {
			last := m.game.states[len(m.game.states)-1]
			if last.Outs == 3 {
				m.flipHalfInning()
				m.lastState = nil
			} else {
				m.lastState = last
			}
		} else {
			m.lastState = nil
		}
	next_play:
		if visitorDone && homeDone {
			break
		}
		if m.state.Top() {
			if visitorDone {
				// no more plays, just move onto the bottom
				m.flipHalfInning()
				goto next_play
			}
			m.playCode = m.game.VisitorPlays[m.visitorPlay]
			m.visitorPlay++
			visitorDone = m.visitorPlay == len(m.game.VisitorPlays)
		} else {
			if homeDone {
				m.flipHalfInning()
				goto next_play
			}
			m.playCode = m.game.HomePlays[m.homePlay]
			m.homePlay++
			homeDone = m.homePlay == len(m.game.HomePlays)
		}
		m.playFields = strings.Split(m.playCode, ",")
		if !IsPlayerID(m.getPlayField(0)) {
			if err := m.handleSpecial(); err != nil {
				return err
			}
			goto next_play
		}
		if m.state.Top() {
			m.state.Number = m.visitorPA
			m.state.Pitcher = m.homePitcher
		} else {
			m.state.Number = m.homePA
			m.state.Pitcher = m.visitorPitcher
		}
		m.state.Batter = PlayerID(m.getPlayField(0))
		if m.state.Batter == "" {
			return fmt.Errorf("no batter for %s", m.playCode)
		}
		if m.state.Pitcher == "" {
			return fmt.Errorf("no pitcher for %s", m.playCode)
		}
		m.state.Pitches = Pitches(m.getPlayField(1))
		m.state.EventCode = m.getPlayField(2)
		if m.state.EventCode == "" {
			return fmt.Errorf("empty event code in %s", m.playCode)
		}
		m.state.Comment = m.getPlayField(3)
		m.basePutOuts = nil
		m.parseEvent(m.state.EventCode)
		var err error
		var runners []PlayerID
		if m.lastState != nil {
			runners = m.lastState.Runners
		}
		m.state.Advances, err = parseAdvances(m.advancesCode, m.state.Batter, runners)
		if err != nil {
			return err
		}
		if err := m.handleEvent(); err != nil {
			return err
		}
		if err := m.moveRunners(); err != nil {
			return err
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
			m.state.Modifiers = Modifiers(m.modifiers)
			if m.state.Top() {
				m.visitorPA++
			} else {
				m.homePA++
			}
		}
		m.game.states = append(m.game.states, m.state)
		m.state = &State{
			InningNumber: m.state.InningNumber,
			Outs:         m.state.Outs,
			Half:         m.state.Half,
			Score:        m.state.Score,
		}
		m.state.init()
	}
	if m.game.Final != nil && (m.game.Final.Home != m.state.Score.Home ||
		m.game.Final.Visitor != m.state.Score.Visitor) {
		return fmt.Errorf("final score was %d-%d not %d-%d", m.state.Score.Visitor,
			m.state.Score.Home, m.game.Final.Visitor, m.game.Final.Home)
	}
	return nil
}

func (m *gameMachine) flipHalfInning() {
	m.state.Outs = 0
	m.state.Runners = make([]PlayerID, 3)
	if m.state.Top() {
		m.state.Half = Bottom
		m.state.Pitcher = m.homePitcher
	} else {
		m.state.InningNumber++
		m.state.Half = Top
		m.state.Pitcher = m.visitorPitcher
	}
}

func (m *gameMachine) handleSpecial() error {
	switch m.getPlayField(0) {
	case "pitcher":
		if m.state.Top() {
			m.homePitcher = PlayerID(m.getPlayField(1))
		} else {
			m.visitorPitcher = PlayerID(m.getPlayField(1))
		}
	case "inn":
		inning, err := strconv.Atoi(m.getPlayField(1))
		if err != nil || inning != m.state.InningNumber {
			return fmt.Errorf("inning %d is not %s", m.state.InningNumber, m.getPlayField(1))
		}
		runs := m.state.Score.Home
		if m.state.Top() {
			runs = m.state.Score.Visitor
		}
		score, err := strconv.Atoi(m.getPlayField(2))
		if err != nil || runs != score {
			return fmt.Errorf("at inning %d # runs is %d not %s", m.state.InningNumber, runs, m.getPlayField(2))
		}
		var outs int
		if len(m.playFields) > 3 {
			outs, err = strconv.Atoi(m.playFields[3])
		}
		if err != nil || outs != m.state.Outs {
			return fmt.Errorf("at inning %d # outs is %d not %d", m.state.InningNumber, m.state.Outs, outs)
		}
	default:
		return fmt.Errorf("unknown special play %s", m.playCode)
	}
	return nil
}

func (m *gameMachine) getPlayField(i int) string {
	if i < len(m.playFields) {
		return m.playFields[i]
	}
	return ""
}

func (m *gameMachine) impliedAdvance(code string) {
	advance, err := parseAdvance(code)
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

func (m *gameMachine) handleEvent() error {
	switch {
	case m.eventIs("$"):
		m.state.Play = &Play{
			Type:     FlyOut,
			Fielders: []int{fielderNumber[m.eventMatches[0]]},
		}
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("$$"):
		fallthrough
	case m.eventIs("$$$"):
		m.state.Play = &Play{
			Type: GroundOut,
		}
		for _, fielder := range m.eventMatches {
			m.state.Play.Fielders = append(m.state.Play.Fielders, fielderNumber[fielder])
		}
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("K"):
		m.state.Play = &Play{
			Type: StrikeOut,
		}
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("W"):
		m.state.Play = &Play{
			Type: Walk,
		}
		m.impliedAdvance("B-1")
		m.state.Complete = true
	case m.eventIs("SB%;SB%;SB%") || m.eventIs("SB%;SB%") || m.eventIs("SB%"):
		m.state.Play = &Play{
			Type:    StolenBase,
			Runners: make([]PlayerID, len(m.eventMatches)),
		}
		for i := range m.eventMatches {
			base := m.eventMatches[i]
			var runner PlayerID
			switch base {
			case "2":
				m.impliedAdvance("1-2")
				runner = m.lastState.Runners[0]
			case "3":
				m.impliedAdvance("2-3")
				runner = m.lastState.Runners[1]
			case "H":
				m.impliedAdvance("3-H")
				runner = m.lastState.Runners[2]
			default:
				return fmt.Errorf("unknown stolen base code %s", m.state.EventCode)
			}
			m.state.Play.StolenBases = append(m.state.Play.StolenBases, base)
			if runner == "" {
				return fmt.Errorf("no runner can steal %s in %s", base, m.state.EventCode)
			}
			m.state.Play.Runners[i] = runner
		}
	case m.eventIs("K2$"):
		m.state.Play = &Play{
			Type:     StrikeOut,
			Fielders: []int{2, fielderNumber[m.eventMatches[0]]},
		}
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("K+PB"):
		m.state.Play = &Play{
			Type: StrikeOutPassedBall,
		}
		m.state.Complete = true
		m.impliedAdvance("B-1")
	case m.eventIs("K+WP"):
		m.state.Play = &Play{
			Type: StrikeOutWildPitch,
		}
		m.state.Complete = true
		m.impliedAdvance("B-1")
	case m.eventIs("S$"):
		m.state.Play = &Play{
			Type:     Single,
			Fielders: []int{fielderNumber[m.eventMatches[0]]},
		}
		m.impliedAdvance("B-1")
		m.state.Complete = true
	case m.eventIs("D$"):
		m.state.Play = &Play{
			Type:     Double,
			Fielders: []int{fielderNumber[m.eventMatches[0]]},
		}
		m.impliedAdvance("B-2")
		m.state.Complete = true
	case m.eventIs("T$"):
		m.state.Play = &Play{
			Type:     Triple,
			Fielders: []int{fielderNumber[m.eventMatches[0]]},
		}
		m.impliedAdvance("B-3")
		m.state.Complete = true
	case m.eventIs("H"):
		m.state.Play = &Play{
			Type: HomeRun,
		}
		m.impliedAdvance("B-H")
		m.state.Complete = true
	case m.eventIs("PB"):
		m.state.Play = &Play{
			Type: PassedBall,
		}
		// movement in advances
	case m.eventIs("WP"):
		m.state.Play = &Play{
			Type: WildPitch,
		}
		// movement in advances
	case m.eventIs("HP"):
		m.state.Play = &Play{
			Type: HitByPitch,
		}
		m.impliedAdvance("B-1")
		m.state.Complete = true
	case m.eventIs("E$"):
		fe, err := parseFieldingError(m.eventCode)
		if err != nil {
			return fmt.Errorf("cannot parse fielding error in %s - %w", m.eventCode, err)
		}
		m.state.Play = &Play{
			Type:          ReachedOnError,
			FieldingError: fe,
		}
		m.impliedAdvance("B-1")
		m.state.Complete = true
	case m.eventIs("C/E$"):
		m.state.Play = &Play{
			Type: CatcherInterference,
			FieldingError: &FieldingError{
				Fielder: fielderNumber[m.eventMatches[0]],
			},
		}
		m.impliedAdvance("B-1")
		m.state.Complete = true
	case m.eventIs("PO%($$)"):
		from := m.eventMatches[1]
		if !(from == "1" || from == "2" || from == "3") {
			return fmt.Errorf("illegal picked off base in %s", m.playCode)
		}
		runner, err := m.getBaseRunner(from)
		if err != nil {
			return fmt.Errorf("cannot pick off in %s - %w", m.playCode, err)
		}
		m.state.Play = &Play{
			Type:    PickedOff,
			Runners: []PlayerID{runner},
		}
		advance := m.state.Advances[from]
		if advance == nil {
			m.state.recordOut()
			m.putOut(from)
		} else {
			m.state.NotOutOnPlay = advance.FieldingError != nil
		}
	case m.eventIs("FC$"):
		// outs are in the advance, if any
		m.state.Play = &Play{
			Type: FieldersChoice,
		}
		m.impliedAdvance("B-1")
		m.state.Complete = true
	case m.eventIs("$$(%)$"):
		if !m.modifiers.Contains("GDP") {
			return fmt.Errorf("play should contain GDP modifier in %s", m.playCode)
		}
		base := m.eventMatches[2]
		runner, err := m.getBaseRunner(base)
		if err != nil {
			return fmt.Errorf("no runner in double play %s - %w", m.playCode, err)
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
	case m.eventIs("$(B)$(%)"):
		if !m.modifiers.Contains("LDP") {
			return fmt.Errorf("play should contain LDP modifier in %s (%v)", m.playCode, m.state.Modifiers)
		}
		base := m.eventMatches[2]
		runner, err := m.getBaseRunner(base)
		if err != nil {
			return fmt.Errorf("no runner in lineout double play %s - %w", m.playCode, err)
		}
		m.state.Play = &Play{
			Type:    DoublePlay,
			Runners: []PlayerID{runner},
		}
		m.state.recordOut()
		m.state.recordOut()
		m.putOut(base)
		m.state.Complete = true
	case m.eventIs("CS%($$)"):
		to := m.eventMatches[0]
		if !(to == "2" || to == "3" || to == "H") {
			return fmt.Errorf("illegal caught stealing base code %s", m.playCode)
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
			return fmt.Errorf("cannot catch stealing runner in %s - %w", m.playCode, err)
		}
		m.state.Play = &Play{
			Type:    CaughtStealing,
			Runners: []PlayerID{runner},
			Base:    to,
		}
	case m.eventIs("NP"):
		m.state.Play = &Play{
			Type: NoPlay,
		}
		// no play
	default:
		return fmt.Errorf("unknown event %s in %s", m.eventCode, m.playCode)
	}
	return nil
}

func (m *gameMachine) getBaseRunner(base string) (runner PlayerID, err error) {
	if base == "H" {
		err = fmt.Errorf("a runner cannot be at H")
		return
	}
	if m.lastState == nil {
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

func (m *gameMachine) moveRunners() error {
	for _, base := range []string{"3", "2", "1", "B"} {
		advance := m.state.Advances[base]
		if advance == nil {
			if m.basePutOuts[base] {
				// runner was put out
			} else if base != "B" && m.lastState != nil {
				// runner did not move
				number := BaseNumber[base]
				m.state.Runners[number] = m.lastState.Runners[number]
			}
			continue
		}
		from := BaseNumber[advance.From]
		to := BaseNumber[advance.To]
		switch {
		case m.lastState == nil && advance.From != "B":
			return fmt.Errorf("cannot advance a runner from %s to %s in %s at start of half-inning",
				advance.From, advance.To, m.playCode)
		case advance.From != "B" && m.lastState.Runners[from] == "":
			return fmt.Errorf("cannot advance non-existent runner from %s in %s PA %d",
				advance.From, m.playCode, m.state.PlateAppearance.Number)
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
			m.state.Runners[to] = m.state.Batter
		default:
			m.state.Runners[to] = m.lastState.Runners[from]
		}
	}
	return nil
}

func (m *gameMachine) scoreRun(runner PlayerID) {
	if m.state.Top() {
		m.state.Score.Visitor++
	} else {
		m.state.Score.Home++
	}
	m.state.ScoringRunners = append(m.state.ScoringRunners, runner)
}
