package stats

import "github.com/slshen/sb/pkg/dataframe"

func OnBase(idx *dataframe.Index, row int) float64 {
	hbp := idx.GetInt(row, "HitByPitch")
	h := idx.GetInt(row, "Hits")
	bb := idx.GetInt(row, "Walks")
	ab := idx.GetInt(row, "AB")
	sf := idx.GetInt(row, "SacrificeFlys")
	return float64(h+bb+hbp) / float64(ab+bb+hbp+sf)
}

func Slugging(idx *dataframe.Index, i int) float64 {
	ab := idx.GetInt(i, "AB")
	if ab == 0 {
		return 0
	}
	s := idx.GetInt(i, "Singles")
	d := idx.GetInt(i, "Doubles")
	t := idx.GetInt(i, "Triples")
	h := idx.GetInt(i, "HRs")
	return float64(s+2*d+3*t+4*h) / float64(ab)
}

func OPS(idx *dataframe.Index, i int) float64 {
	return OnBase(idx, i) + Slugging(idx, i)
}

func Thousands(fn func(idx *dataframe.Index, i int) float64) func(idx *dataframe.Index, i int) int {
	return func(idx *dataframe.Index, i int) int {
		f := fn(idx, i)
		return int(1000 * f)
	}
}

func AVG(idx *dataframe.Index, i int) float64 {
	h := idx.GetInt(i, "Hits")
	ab := idx.GetInt(i, "AB")
	if ab > 0 {
		return float64(h) / float64(ab)
	}
	return 0
}

func LAVG(idx *dataframe.Index, i int) float64 {
	h := idx.GetInt(i, "Hits")
	lo := idx.GetInt(i, "LineDriveOuts")
	ab := idx.GetInt(i, "AB")
	if ab > 0 {
		return float64(h+lo) / float64(ab)
	}
	return 0
}
