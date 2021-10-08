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
	playCode                    string
	playFields                  []string
	eventCodeParser
	impliedAdvances []string
}

func (m *gameMachine) run() error {
	m.visitorPA = 1
	m.homePA = 1
	m.state = &State{
		InningNumber: 1,
		Half:         Top,
	}
	var visitorDone, homeDone bool
	for {
		if last := m.lastState(); last != nil && last.Outs == 3 {
			m.flipHalfInning()
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
			m.state.Pitcher = m.visitorPitcher
		} else {
			m.state.Number = m.homePA
			m.state.Pitcher = m.homePitcher
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
		m.state.Play = m.parseEvent(m.state.EventCode)
		if err := m.handleEvent(); err != nil {
			return err
		}
		var err error
		m.state.Advances, err = m.parseAdvances(m.impliedAdvances)
		if err != nil {
			return err
		}
		if err := m.moveRunners(m.state.Advances); err != nil {
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
		m.state = m.state.copy()
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
		m.state.Pitcher = m.visitorPitcher
	} else {
		m.state.InningNumber++
		m.state.Half = Top
		m.state.Pitcher = m.homePitcher
	}
}

func (m *gameMachine) handleSpecial() error {
	switch m.getPlayField(0) {
	case "pitcher":
		if m.state.Top() {
			m.visitorPitcher = PlayerID(m.getPlayField(1))
		} else {
			m.homePitcher = PlayerID(m.getPlayField(1))
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
			return fmt.Errorf("# of at inning %d is %d not %s", m.state.InningNumber, runs, m.getPlayField(2))
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

func (m *gameMachine) handleEvent() error {
	m.impliedAdvances = nil
	switch {
	case m.eventIs("$"):
		fallthrough
	case m.eventIs("$$"):
		fallthrough
	case m.eventIs("$$$"):
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("K"):
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("W"):
		m.impliedAdvances = []string{"B-1"}
		m.state.Complete = true
	case m.eventIs("SB%;SB%;SB%") || m.eventIs("SB%;SB%") || m.eventIs("SB%"):
		for i := range m.eventMatches {
			base := m.eventMatches[i]
			switch base {
			case "2":
				m.impliedAdvances = append(m.impliedAdvances, "1-2")
			case "3":
				m.impliedAdvances = append(m.impliedAdvances, "2-3")
			case "H":
				m.impliedAdvances = append(m.impliedAdvances, "3-H")
			default:
				return fmt.Errorf("unknown stolen base code %s", m.state.EventCode)
			}
		}
	case m.eventIs("K2$"):
		m.state.recordOut()
		m.state.Complete = true
	case m.eventIs("K+PB"):
		fallthrough
	case m.eventIs("K+WP"):
		m.state.Complete = true
		m.impliedAdvances = append(m.impliedAdvances, "B-1")
	case m.eventIs("S$"):
		m.impliedAdvances = append(m.impliedAdvances, "B-1")
		m.state.Complete = true
	case m.eventIs("D$"):
		m.impliedAdvances = append(m.impliedAdvances, "B-2")
		m.state.Complete = true
	case m.eventIs("T$"):
		m.impliedAdvances = append(m.impliedAdvances, "B-3")
		m.state.Complete = true
	case m.eventIs("H"):
		m.impliedAdvances = append(m.impliedAdvances, "B-H")
		for i := range m.state.Runners {
			if m.state.Runners[i] != "" {
				m.impliedAdvances = append(m.impliedAdvances, fmt.Sprintf("%d-H", i+1))
			}
		}
		m.state.Complete = true
	case m.eventIs("PB"):
	case m.eventIs("WP"):
	case m.eventIs("HP"):
		m.impliedAdvances = append(m.impliedAdvances, "B-1")
	case m.eventIs("E$"):
		m.impliedAdvances = append(m.impliedAdvances, "B-1")
		m.state.Complete = true
	case m.eventIs("C/E$"):
		m.impliedAdvances = append(m.impliedAdvances, "B-1")
		m.state.Complete = true
	case m.eventIs("PO%($$)"):
		m.state.recordOut()
		return m.eraseRunner(m.eventMatches[1])
	case m.eventIs("FC$"):
		// outs are in the advance, if any
		m.impliedAdvances = append(m.impliedAdvances, "B-1")
	case m.eventIs("$$(%)$"):
		if !m.modifiers.Contains("GDP") {
			return fmt.Errorf("play should contain GDP modifier in %s", m.playCode)
		}
		// should pass fielders to record out to do assists
		m.state.recordOut()
		m.state.recordOut()
		return m.eraseRunner(m.eventMatches[2])
	case m.eventIs("$(B)$(%)"):
		if !m.state.Modifiers.Contains("LDP") {
			return fmt.Errorf("play should contain LDP modifier in %s", m.playCode)
		}
		m.state.recordOut()
		m.state.recordOut()
		return m.eraseRunner(m.eventMatches[2])
	case m.eventIs("CS%($$)"):
		m.state.recordOut()
		return m.eraseAdvancingRunner(m.eventMatches[1])
	case m.eventIs("NP"):
		// no play
	default:
		return fmt.Errorf("unknown event %s in %s", m.eventCode, m.playCode)
	}
	return nil
}

func (m *gameMachine) eraseAdvancingRunner(base string) error {
	var index int
	switch base {
	case "2":
		index = 0
	case "3":
		index = 1
	case "H":
		index = 2
	default:
		return fmt.Errorf("unknown base %s in %s", base, m.playCode)
	}
	if m.state.Runners[index] == "" {
		return fmt.Errorf("no runner on %s to erase in %s", base, m.playCode)
	}
	m.state.Runners[index] = ""
	return nil
}

func (m *gameMachine) eraseRunner(base string) error {
	var index int
	switch base {
	case "1":
		index = 0
	case "2":
		index = 1
	case "3":
		index = 2
	default:
		return fmt.Errorf("unknown base %s in %s", base, m.playCode)
	}
	if m.state.Runners[index] == "" {
		return fmt.Errorf("no runner on %s to erase in %s", base, m.playCode)
	}
	m.state.Runners[index] = ""
	return nil
}

func (m *gameMachine) moveRunners(advances []Advance) error {
	for i := len(advances) - 1; i >= 0; i-- {
		advance := advances[i]
		if len(m.state.Runners) == 0 {
			m.state.Runners = make([]PlayerID, 3)
		}
		from := baseNumber(advance.From)
		to := baseNumber(advance.To)
		if from < -1 || from > 2 || to < 0 || to > 3 {
			return fmt.Errorf("invalid advance %s", advance.Code)
		}
		if advance.From != "B" && m.lastState() != nil && m.lastState().Runners[from] == "" {
			return fmt.Errorf("cannot advance non-existent runner from %s in %s PA %d",
				advance.From, m.playCode, m.state.PlateAppearance.Number)
		}
		switch {
		case advance.Out:
			m.state.recordOut()
			if advance.From != "B" {
				m.state.Runners[from] = ""
			}
		case advance.To == "H":
			if advance.From == "B" {
				m.scoreRun(m.state.Batter)
			} else {
				m.scoreRun(m.state.Runners[from])
			}
			if advance.From != "B" {
				m.state.Runners[from] = ""
			}
		case advance.From == "B":
			m.state.Runners[to] = m.state.Batter
		default:
			m.state.Runners[to] = m.lastState().Runners[from]
			m.state.Runners[from] = ""
		}
	}
	return nil
}

func (m *gameMachine) lastState() *State {
	if len(m.game.states) > 0 {
		return m.game.states[len(m.game.states)-1]
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
