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
	FoulFlyError
	NoPlay
)

//go:generate stringer -type=PlayType

func (p PlayType) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

type Play struct {
	Type          PlayType
	Runners       []PlayerID     `yaml:",omitempty,flow"`
	Base          string         `yaml:",omitempty"`
	FieldingError *FieldingError `yaml:",omitempty"`
	Fielders      []int          `yaml:",omitempty,flow"`
	StolenBases   []string       `yaml:",omitempty,flow"`
}

func (p *Play) Is(ts ...PlayType) bool {
	for _, t := range ts {
		if p.Type == t {
			return true
		}
	}
	return false
}

func (p *Play) Hit() bool {
	return p.Type == Single || p.Type == Double || p.Type == HomeRun
}

func (p *Play) BallInPlay() bool {
	return p.Hit() || p.Is(ReachedOnError, FieldersChoice, GroundOut, FlyOut, DoublePlay, TriplePlay)
}
