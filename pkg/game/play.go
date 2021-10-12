package game

import (
	"sort"
	"strings"
)

type Play string

func (p Play) Single() bool {
	return len(p) == 2 && p[0] == 'S' && isFielder(p[1])
}
func (p Play) Double() bool {
	return len(p) == 2 && p[0] == 'D' && isFielder(p[1])
}
func (p Play) Triple() bool {
	return p[0] == 'T'
}
func (p Play) HomeRun() bool {
	return p == "H"
}

func (p Play) StrikeOut() bool {
	return p[0] == 'K'
}

func (p Play) CaughtStealing() bool {
	return strings.HasPrefix(string(p), "CS")
}

func (p Play) HitByPitch() bool {
	return p == "HP"
}

func (p Play) Walk() bool {
	return p == "W"
}

func (p Play) Hit() bool {
	return p.Single() || p.Double() || p.Triple() || p.HomeRun()
}

func (p Play) StolenBase() bool {
	return len(p) >= 3 && p[0] == 'S' && p[1] == 'B'
}

func (p Play) StolenBases() (bases []string) {
	for _, sb := range strings.Split(string(p), ";") {
		switch sb {
		case "SB2":
			bases = append(bases, "2")
		case "SB3":
			bases = append(bases, "3")
		case "SBH":
			bases = append(bases, "H")
		}
	}
	sort.Slice(bases, func(i, j int) bool {
		return BaseNumber[bases[j]] < BaseNumber[bases[i]]
	})
	return
}

func (p Play) CatcherInterference() bool {
	return strings.HasPrefix(string(p), "C/")
}

func (p Play) ReachedOnError() bool {
	return p[0] == 'E'
}

func (p Play) FieldingError() (*FieldingError, error) {
	return parseFieldingError(string(p))
}

func (p Play) FieldersChoice() bool {
	return strings.HasPrefix(string(p), "FC")
}

func (p Play) BallInPlay() bool {
	return p.Hit() || p.FieldersChoice() || (p[0] >= '1' && p[0] <= '9')
}

func (p Play) WildPitch() bool {
	return p == "WP" || p == "K+WP"
}

func (p Play) PassedBall() bool {
	return p == "PB" || p == "K+PB"
}

func isFielder(ch byte) bool {
	return ch >= '1' && ch <= '9'
}

func (p Play) GroundOut() bool {
	if len(p) < 2 {
		return false
	}
	for _, f := range p {
		if !isFielder(byte(f)) {
			return false
		}
	}
	return true
}

func (p Play) FlyOut() bool {
	return len(p) == 1 && isFielder(p[0])
}
