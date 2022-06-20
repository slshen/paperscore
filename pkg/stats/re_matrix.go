package stats

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type reMatrix map[string][]float64

func ReadREMatrix(path string) (RunExpectancy, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rem := make(reMatrix)
	for {
		rec, err := r.Read()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if rec == nil {
			break
		}
		if len(rem) == 0 && rec[0] == "Runr" {
			continue
		}
		vals := make([]float64, 3)
		rem[rec[0]] = vals
		for i, field := range rec[1:4] {
			vals[i], err = strconv.ParseFloat(strings.TrimSpace(field), 64)
			if err != nil {
				return nil, err
			}
		}
	}
	return rem, nil
}

func (rem reMatrix) GetExpectedRuns(outs int, runrs OccupiedBases) float64 {
	vals := rem[string(runrs)]
	return vals[outs]
}
