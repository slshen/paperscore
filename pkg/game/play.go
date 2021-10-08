package game

import (
	"log"
	"strings"
)

type Play string

func (p Play) Single() bool {
	return p[0] == 'S'
}
func (p Play) Double() bool {
	return p[0] == 'D'
}
func (p Play) Triple() bool {
	return p[0] == 'T'
}
func (p Play) HomeRun() bool {
	return p[0] == 'H'
}

func (p Play) StrikeOut() bool {
	return p[0] == 'K'
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

func (p Play) StolenBase() (second bool, third bool, home bool) {
	if len(p) >= 3 && p[0] == 'S' && p[1] == 'B' {
		for _, sb := range strings.Split(string(p), ";") {
			switch sb {
			case "SB2":
				second = true
			case "SB3":
				third = true
			case "SBH":
				home = true
			}
		}
	}
	return
}

func (p Play) CatcherInterference() bool {
	return strings.HasPrefix(string(p), "C/")
}

func (p Play) ReachedOnError() *FieldingError {
	if p[0] != 'E' {
		return nil
	}
	fe, err := parseFieldingError(string(p))
	if err != nil {
		log.Fatal(err)
	}
	return fe
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
