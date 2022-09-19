package game

type PlayType byte

const (
	Single PlayType = iota
	Double
	Triple
	HomeRun
	CaughtStealing
	HitByPitch
	Walk
	WalkWildPitch
	WalkPassedBall
	WalkPickedOff
	StolenBase
	PickedOff
	CatcherInterference
	ReachedOnError
	FieldersChoice
	WildPitch
	PassedBall
	GroundOut
	FlyOut
	DoublePlay
	TriplePlay
	StrikeOut
	StrikeOutPassedBall
	StrikeOutWildPitch
	StrikeOutPickedOff
	StrikeOutStolenBase
	FoulFlyError
	NoPlay
)

//go:generate stringer -type=PlayType

func (p PlayType) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

type Play struct {
	Type                 PlayType
	FieldingError        FieldingError `yaml:",omitempty"`
	Fielders             []int         `yaml:",omitempty,flow"`
	StolenBases          []string      `yaml:",omitempty,flow"`
	ScoringRunners       []PlayerID    `yaml:",flow,omitempty"`
	OutsOnPlay           int           `yaml:",omitempty"`
	PickedOffRunner      PlayerID      `yaml:",omitempty"`
	CaughtStealingRunner PlayerID      `yaml:",omitempty"`
	CaughtStealingBase   string        `yaml:",omitempty"`
	NotOutOnPlay         bool          `yaml:",omitempty"` // not out on CS, POCS due to error
}

func (p *Play) Is(ts ...PlayType) bool {
	for _, t := range ts {
		if p.Type == t {
			return true
		}
	}
	return false
}

func (p *Play) IsHit() bool {
	return p.Type == Single || p.Type == Double || p.Type == Triple || p.Type == HomeRun
}

func (p *Play) IsStrikeOut() bool {
	return p.Is(StrikeOut, StrikeOutPassedBall, StrikeOutPickedOff, StrikeOutWildPitch,
		StrikeOutStolenBase)
}

func (p *Play) IsWalk() bool {
	return p.Is(Walk, WalkPassedBall, WalkPickedOff, WalkPickedOff)
}

func (p *Play) IsBallInPlay() bool {
	return p.IsHit() || p.Is(ReachedOnError, FieldersChoice, GroundOut, FlyOut, DoublePlay, TriplePlay)
}

/*func (p *Play) GetRunner(n int) PlayerID {
	if n < len(p.Runners) {
		return p.Runners[n]
	}
	return ""
}*/
