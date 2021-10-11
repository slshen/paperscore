package game

import "fmt"

var pitchCodes = map[rune]struct {
	narrative string
	ball      bool
	strike    bool
	foul      bool
}{
	'B': {"Ball", true, false, false},
	'C': {"Called Strike", false, true, false},
	'F': {"Foul", false, false, true},
	'I': {"Intentional Ball", true, false, false},
	'K': {"Strike", false, true, false},
	'L': {"Foul Bunt", false, true, false},
	'M': {"Missed Bunt Attempt", false, true, false},
	'N': {"No Pitch", false, false, false},
	'O': {"Foul Tip on Bunt", false, true, false},
	'P': {"Pitchout", false, false, false},
	'Q': {"Swinging on Pitchout", false, true, false},
	'R': {"Foul Ball on Pitchout", false, false, true},
	'S': {"Swinging Strike", false, true, false},
	'T': {"Foul Tip", false, true, false},
	'U': {"Unknown", false, false, false},
	'X': {"Ball In Play", false, false, false},
	'Y': {"Ball In Play (Pitchout)", false, false, false},
}

func (ps Pitches) Balls() (count int) {
	for _, p := range ps {
		code := pitchCodes[p]
		if code.ball {
			count++
		}
	}
	return
}

func (ps Pitches) Strikes() (count int) {
	for _, p := range ps {
		code := pitchCodes[p]
		if code.strike {
			count++
		}
	}
	return
}

// Return the number of swings, not including the ball in play
func (ps Pitches) Swings() (count int) {
	for _, p := range ps {
		code := pitchCodes[p]
		if code.foul || (code.strike && p != 'C') {
			count++
		}
	}
	return
}

func (ps Pitches) Misses() (count int) {
	for _, p := range ps {
		if p == 'S' || p == 'T' {
			count++
		}
	}
	return
}

func (ps Pitches) Last() string {
	if l := len(ps); l > 0 {
		return string(ps[l-1])
	}
	return ""
}

func (ps Pitches) Count() string {
	return fmt.Sprintf("%d%d", ps.Balls(), ps.Strikes())
}
