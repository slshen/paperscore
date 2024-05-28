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

func (ps Pitches) Last() string {
	if l := len(ps); l > 0 {
		return string(ps[l-1])
	}
	return ""
}

func (ps Pitches) Count() (string, int, int) {
	balls := ps.CountUp('B')
	strikes := 0
	for _, p := range ps {
		if p == 'C' || p == 'S' || p == 'T' || p == 'M' || p == 'L' {
			strikes++
		} else if strikes < 2 && p == 'F' {
			strikes++
		}
	}
	return fmt.Sprintf("%d-%d", balls, strikes), balls, strikes
}
