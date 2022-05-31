package game

type MisplayType int

type MisplayGoodplay struct {
	MisplayType
	GoodPlayType
	Fielder int
}

const (
	NoMisplay MisplayType = iota
	FailToCover
	NotCallingCatch
	FellAsleep
)

type GoodPlayType int

const (
	NoGoodPlay GoodPlayType = iota
	DivingCatch
	LongRunningCatch
)
