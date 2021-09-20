package game

import (
	"fmt"
	"regexp"
	"sort"
)

type Advance struct {
	Code               string
	From, To           string
	Out                bool  `yaml:",omitempty"`
	Fielders           []int `yaml:",omitempty,flow"`
	RunnerInterference bool  `yaml:",omitempty"`
	*FieldingError     `yaml:",omitempty"`
}

var advanceRegexp = regexp.MustCompile(`^([B123])([X-])([123H])(?:\(([^)]+)\))?$`)

func parseAdvance(s string) (*Advance, error) {
	m := advanceRegexp.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("illegal advance code %s", s)
	}
	a := &Advance{
		Code: s,
		From: m[1],
		To:   m[3],
		Out:  m[2] == "X",
	}
	if a.Out {
		if m[4] == "RINT" {
			a.RunnerInterference = true
		} else {
			for _, f := range m[4] {
				if f >= '1' && f <= '9' {
					a.Fielders = append(a.Fielders, int(f-'1')+1)
				} else {
					return nil, fmt.Errorf("illegal fielder %c for put out in advance code %s", f, s)
				}
			}
			if len(a.Fielders) == 0 {
				return nil, fmt.Errorf("no fielders for put out in advancde code %s", s)
			}
		}
	} else if m[4] != "" {
		var err error
		a.FieldingError, err = parseFieldingError(m[4])
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

func baseNumber(b string) int {
	switch b {
	case "B":
		return -1
	case "1":
		return 0
	case "2":
		return 1
	case "3":
		return 2
	case "H":
		return 3
	}
	return -2
}

func sortAdvances(advances []Advance) {
	sort.Slice(advances, func(i, j int) bool {
		bi := baseNumber(advances[i].From)
		bj := baseNumber(advances[j].To)
		return bi < bj
	})
}
