package game

import "strings"

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
