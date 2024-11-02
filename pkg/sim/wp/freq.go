package wp

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/slshen/paperscore/pkg/stats"
)

type RunFrequency struct {
	rand  *rand.Rand
	probs map[stats.OccupiedBases][][]float64
}

func LoadRunFrequency(path string) (*RunFrequency, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}
	rf := &RunFrequency{
		// #nosec:G404
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
		probs: make(map[stats.OccupiedBases][][]float64),
	}
	for _, rec := range recs[1:] {
		runrs := stats.OccupiedBases(rec[0])
		outs := rf.probs[runrs]
		if outs == nil {
			outs = make([][]float64, 3)
			rf.probs[runrs] = outs
		}
		out, _ := strconv.Atoi(rec[1])
		probs := outs[out]
		if probs == nil {
			probs = make([]float64, 10)
			outs[out] = probs
		}
		for runs := 0; runs < 10; runs++ {
			probs[runs], _ = strconv.ParseFloat(rec[2+runs], 64)
		}
	}
	for _, rnrs := range stats.OccupedBasesValues {
		outs := rf.probs[rnrs]
		if outs == nil {
			return nil, fmt.Errorf("run frequency data is missing for %s", rnrs)
		}
		for out := 0; out < 3; out++ {
			probs := outs[out]
			if len(probs) != 10 {
				return nil, fmt.Errorf("run frequency data is missing for %s/%d", rnrs, out)
			}
			if !sort.Float64sAreSorted(probs) {
				return nil, fmt.Errorf("run frequency probability must be sorted for %s/%d", rnrs, out)
			}
		}
	}
	return rf, nil
}

func (rf *RunFrequency) GetRuns(outs int, rnrs stats.OccupiedBases) int {
	probs := rf.probs[rnrs][outs]
	p := rf.rand.Float64()
	return sort.SearchFloat64s(probs, p)
}
