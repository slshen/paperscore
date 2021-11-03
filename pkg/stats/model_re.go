package stats

func getExpectedRuns(re RunExpectancy, outs int, runners string) float64 {
	return re.GetExpectedRuns(outs, runners[2] != '_', runners[1] != '_', runners[0] != '_')
}

func NewModeledRunExpectancy(ref RunExpectancy, re RunExpectancy) RunExpectancy {
	m := make(reMatrix)
	re0 := getExpectedRuns(re, 0, "___")
	re1 := getExpectedRuns(re, 1, "___")
	re2 := getExpectedRuns(re, 2, "___")
	m["___"] = []float64{re0, re1, re2}
	m["__1"] = []float64{
		getExpectedRuns(m, 0, "___") * getExpectedRuns(ref, 0, "__1") / getExpectedRuns(ref, 0, "___"),
		re1 * getExpectedRuns(ref, 1, "__1") / getExpectedRuns(ref, 1, "___"),
		re2 * getExpectedRuns(ref, 2, "__1") / getExpectedRuns(ref, 2, "___"),
	}
	m["_2_"] = []float64{
		re0 * getExpectedRuns(ref, 0, "__1") / getExpectedRuns(ref, 0, "___"),
	}
	return m
}
