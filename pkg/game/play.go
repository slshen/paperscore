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
	NoPlay
)

type Play struct {
	Type          PlayType
	Runners       []PlayerID
	Base          string
	FieldingError *FieldingError
	Fielders      []int
	StolenBases   []string
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
