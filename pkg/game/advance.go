package game

import (
	"fmt"
	"regexp"

	"github.com/slshen/sb/pkg/gamefile"
)

type Advance struct {
	Code               string
	From, To           string
	Out                bool  `yaml:",omitempty"`
	Fielders           []int `yaml:",omitempty,flow"`
	RunnerInterference bool  `yaml:",omitempty"`
	Implied            bool  `yaml:",omitempty"`
	Runner             PlayerID
	WildPitch          bool `yaml:",omitempty"`
	PassedBall         bool `yaml:",omitempty"`
	Steal              bool `yaml:",omitempty"`
	FieldingError      `yaml:",omitempty"`
}

type Advances []*Advance

var advanceRegexp = regexp.MustCompile(`^([B123])([X-])([123H])(?:\(([^)]+)\))?$`)
var BaseNumber = map[string]int{
	"1": 0,
	"2": 1,
	"3": 2,
	"H": 3,
}
var PreviousBase = map[string]string{
	"H": "3",
	"3": "2",
	"2": "1",
}
var NextBase = map[string]string{
	"1": "2",
	"2": "3",
	"3": "H",
}
var runnerNumber = map[string]int{
	"1": 0,
	"2": 1,
	"3": 2,
}

func (a *Advance) GoString() string {
	return a.Code
}

func parseAdvance(play gamefile.Play, s string) (*Advance, error) {
	m := advanceRegexp.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("%s: illegal advance code %s", play.GetPos(), s)
	}
	a := &Advance{
		Code: s,
		From: m[1],
		To:   m[3],
		Out:  m[2] == "X",
	}
	switch {
	case a.Out:
		if m[4] == "RINT" {
			a.RunnerInterference = true
		} else {
			for _, f := range m[4] {
				if f >= '1' && f <= '9' {
					a.Fielders = append(a.Fielders, int(f-'1')+1)
				} else {
					return nil, fmt.Errorf("%s: illegal fielder %c for put out in advance code %s",
						play.GetPos(), f, s)
				}
			}
			if len(a.Fielders) == 0 {
				return nil, fmt.Errorf("%s: no fielders for put out in advancde code %s",
					play.GetPos(), s)
			}
		}
	case m[4] == "WP":
		a.WildPitch = true
	case m[4] == "PB":
		a.PassedBall = true
	case m[4] != "":
		var err error
		a.FieldingError, err = parseFieldingError(play, m[4])
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (advs Advances) From(base string) *Advance {
	for _, adv := range advs {
		if adv.From == base {
			return adv
		}
	}
	return nil
}

func parseAdvances(play gamefile.Play, batter PlayerID, runners [3]PlayerID) (advances Advances, err error) {
	for _, as := range play.GetAdvances() {
		var advance *Advance
		advance, err = parseAdvance(play, as)
		if err != nil {
			return
		}
		if advances.From(advance.From) != nil {
			err = fmt.Errorf("%s: cannot advance %s twice in %s", play.GetPos(), advance.From, as)
			return
		}
		if advance.From == "B" {
			advance.Runner = batter
		} else {
			/*if runners == nil {
				err = fmt.Errorf("%s: no runner to advance from %s at the start of a half-inning",
					play.GetPos(), advance.From)
				return
			}*/
			advance.Runner = runners[runnerNumber[advance.From]]
			if advance.Runner == "" {
				err = fmt.Errorf("%s: no runner to advance from %s in %s", play.GetPos(),
					advance.From, as)
				return
			}
		}
		advances = append(advances, advance)
	}
	return
}
