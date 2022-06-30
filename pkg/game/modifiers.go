package game

type Modifiers []string

const (
	Throwing     = "TH"
	SacrificeHit = "SH"
	SacrificeFly = "SF"
	Obstruction  = "OBS"
)

const (
	Bunt       = Trajectory("B")
	FlyBall    = Trajectory("F")
	PopUp      = Trajectory("P")
	GroundBall = Trajectory("G")
	LineDrive  = Trajectory("L")
)

type Trajectory string

func (mods Modifiers) Trajectory() Trajectory {
	for _, m := range mods {
		switch {
		case m == "":
			return GroundBall
		case m == "B":
			return Bunt
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
