package game

type PlayEvent struct {
	PlayCode      string
	Pitches       string
	AdvancesCodes []string
	Comment       string
}

type PitchingChangeEvent struct {
	Pitcher PlayerID
}
