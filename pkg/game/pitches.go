package game

import "fmt"

func (ps Pitches) CountUp(codes ...rune) (count int) {
	for _, p := range ps {
		for _, code := range codes {
			if p == code {
				count++
			}
		}
	}
	return
}

func (ps Pitches) Last() rune {
	if ps.IsUnknown() {
		return '?'
	}
	if l := len(ps); l > 0 {
		return rune(ps[l-1])
	}
	return 0
}

func (ps Pitches) IsUnknown() bool {
	if len(ps) == 0 {
		return false
	}
	for _, p := range ps {
		if p != '?' && p != '.' {
			return false
		}
	}
	return true
}

func (ps Pitches) Count() (bool, string, int, int) {
	if ps.IsUnknown() {
		// unrecorded
		return false, "?", 0, 0
	}
	balls := ps.CountUp('B')
	strikes := 0
	for _, p := range ps {
		if p == 'C' || p == 'S' || p == 'T' || p == 'M' || p == 'L' {
			strikes++
		} else if strikes < 2 && p == 'F' {
			strikes++
		}
	}
	return true, fmt.Sprintf("%d-%d", balls, strikes), balls, strikes
}
