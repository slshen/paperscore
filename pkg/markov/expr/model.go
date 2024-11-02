package expr

import (
	"fmt"
	"math/rand"

	"github.com/slshen/paperscore/pkg/markov"
)

type ExprModel struct {
	File *File

	events      map[string]*EventDef
	pEvent      map[string]float64
	transitions map[markov.BaseOutState]*StateTransitions
	statesSeen  map[markov.BaseOutState]*Position
}

func NewModel(file *File) (*ExprModel, Diagnostics) {
	m := &ExprModel{
		File:        file,
		events:      map[string]*EventDef{},
		pEvent:      map[string]float64{},
		transitions: map[markov.BaseOutState]*StateTransitions{},
		statesSeen:  map[markov.BaseOutState]*Position{},
	}
	return m, m.initialize()
}

func (m *ExprModel) initialize() Diagnostics {
	var result Diagnostics
	for _, c := range m.File.Statements {
		if c.EventDef != nil {
			if err := m.initializeEvent(c.EventDef); err != nil {
				result = result.Append(err)
			}
		}
	}
	if err := m.initializeEventValues(); err != nil {
		result = result.Append(err)
	}
	for _, c := range m.File.Statements {
		if t := c.Transitions; t != nil {
			if err := m.initializeState(c.Transitions); err != nil {
				result = result.Append(err)
			}
		}
	}
	for state, pos := range m.statesSeen {
		if state != markov.EndState && m.transitions[state] == nil {
			result = result.Append(
				fmt.Errorf("%s:no transitions defined for %s first seen here", pos, state))
		}
	}
	if m.statesSeen[markov.EndState] == nil {
		result = result.Append(fmt.Errorf("%s: no transitions to 3 outs", m.File.Path))
	}
	return result
}

func (m *ExprModel) initializeEvent(event *EventDef) error {
	if event.Player != "" {
		return fmt.Errorf("%s:player overlays unsupported", event.Pos)
	}
	m.events[event.Name] = event
	return nil
}

func (m *ExprModel) initializeEventValues() Diagnostics {
	var result Diagnostics
	/*
		for _, event := range m.events {

			if event.Invert != nil {
				ref := m.events[*event.Invert]
				switch {
				case ref == nil:
					result = result.Append(fmt.Errorf("%s: invert event %s does not exist", event.Pos, *event.Invert))
				case ref.Value == nil:
					result = result.Append(fmt.Errorf("%s: invert event %s must not be inverted itself", event.Pos, ref.Name))
				default:
					m.pEvent[event.Name] = 1 - *ref.Value
				}
			} else {
				m.pEvent[event.Name] = *event.Value
			}
		}
	*/
	return result
}

func (m *ExprModel) initializeState(t *StateTransitions) Diagnostics {
	var result Diagnostics
	from, err := markov.ParseBaseOutState(t.From)
	if err != nil {
		result = result.Append(fmt.Errorf("%s:invalid from - %s", t.Pos, err))
	}
	if o := m.transitions[from]; o != nil {
		result = result.Append(
			fmt.Errorf("%s:transitions for %s duplicate those at %s", t.Pos, from, o.Pos))
	} else {
		m.transitions[from] = t
	}
	var p float64
	for _, e := range t.Events {
		to, err := markov.ParseBaseOutState(e.To)
		if err != nil {
			result = result.Append(fmt.Errorf("%s:invalid to - %s", e.Pos, err))
		}
		e.to = to
		if m.statesSeen[to] == nil {
			m.statesSeen[to] = &e.Pos
		}
		p += m.pEvent[e.Name]
	}
	t.norm = p
	if t.norm != 1 {
		result = result.Append(fmt.Sprintf("%s:transition probabilities from %s sum to %f not 1", t.Pos, t.From, p))
	}
	return result
}

func (m *ExprModel) NextState(rnd *rand.Rand, batter string, current markov.BaseOutState) (next markov.BaseOutState, event string, runs float64) {
	transitions := m.transitions[current]
	p := rnd.Float64() * transitions.norm
	for _, e := range transitions.Events {
		pe := m.pEvent[e.Name]
		if p < pe {
			next = e.to
			runs = e.Runs
			event = e.Name
			return
		}
		p -= pe
	}
	panic("no event")
}
