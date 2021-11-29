package stats

import "github.com/slshen/sb/pkg/dataframe"

func GetBiggestRE24(dat *dataframe.Data, n int) *dataframe.Data {
	idx := dat.GetIndex()
	dat = dat.RSort(func(r1, r2 int) bool {
		return idx.GetFloat(r1, "RE24") > idx.GetFloat(r2, "RE24")
	})
	rc := dat.RowCount()
	if n > 0 && rc > (2*n)+1 {
		// take top n, bottom n
		dat = dat.RFilter(func(row int) bool {
			return row < n || row > (rc-n)
		})
	}
	return dat
}
