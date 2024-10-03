package game

import (
	"regexp"
	"strconv"
	"strings"
)

type Modifiers []string

const (
	Throwing               = "TH"
	SacrificeHit           = "SH"
	SacrificeFly           = "SF"
	Obstruction            = "OBS"
	GroundedIntoDoublePlay = "GDP"
)

const (
	Bunt         = Trajectory("B")
	BuntGrounder = Trajectory("BG")
	BuntPopup    = Trajectory("BP")
	FlyBall      = Trajectory("F")
	PopUp        = Trajectory("P")
	GroundBall   = Trajectory("G")
	LineDrive    = Trajectory("L")
)

type Trajectory string
type Location struct {
	Fielder int
	Length  string
}

var locationRe = regexp.MustCompile(`[A-Z]+([1-9])(S|D)?`)

func (mods Modifiers) Trajectory() Trajectory {
	for _, m := range mods {
		if m == "B" {
			// have been using 23/G2/B to indicate a bunt
			return Bunt
		}
	}
	for _, m := range mods {
		switch {
		case m == "":
			return GroundBall
		case strings.HasPrefix(m, "BG"):
			// more properly should be using 23/BG2 or 2/BP2
			return BuntGrounder
		case strings.HasPrefix(m, "BP"):
			return BuntPopup
		case m[0] == 'F':
			return FlyBall
		case m[0] == 'G':
			return GroundBall
		case m[0] == 'P':
			return PopUp
		case m[0] == 'L':
			return LineDrive
		}
	}
	return ""
}

func (mods Modifiers) Location() *Location {
	for _, m := range mods {
		if m[0] != 'E' {
			rm := locationRe.FindStringSubmatch(m)
			if rm != nil {
				// F8S = short center, P6D deep shortstop
				fielder, _ := strconv.Atoi(rm[1])
				var length string
				if rm[2] == "D" {
					length = "deep"
				} else if rm[2] == "S" {
					length = "short"
				}
				return &Location{
					Fielder: fielder,
					Length:  length,
				}
			}
		}
	}
	return nil
}

func (mods Modifiers) Contains(codes ...string) bool {
	for _, m := range mods {
		for _, code := range codes {
			if m == code {
				return true
			}
		}
	}
	return false
}
