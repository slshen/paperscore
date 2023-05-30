package markov

import "math/rand"

type ExplicitModel struct {
	P [24][25]float64
	R [24][24]float64
}

func (m *ExplicitModel) NextState(rnd rand.Rand, batter string, current BaseOutState) (next BaseOutState, runs float64) {
	r := rnd.Float64()
	line := m.P[current]
	i := 0
	for {
		r -= line[i]
		if r < 0 {
			break
		}
		i++
	}
	next = BaseOutState(i)
	if next.Outs() < 3 {
		runs = m.R[current][i]
	}
	return
}
