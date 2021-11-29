package stats

import "github.com/slshen/sb/pkg/dataframe"

type REAnalysis struct {
	RE RunExpectancy

	beforeRunrs, beforeOuts *dataframe.Column
	afterRunrs, afterOuts   *dataframe.Column
	change                  *dataframe.Column
	narrative               *dataframe.Column
}

func NewREAnalysis(re RunExpectancy) *REAnalysis {
	return &REAnalysis{
		RE: re,
		beforeRunrs: &dataframe.Column{
			Name:   "BR",
			Format: "%3s",
		},
		beforeOuts: &dataframe.Column{
			Name:   "BO",
			Format: "%2d",
		},
		afterRunrs: &dataframe.Column{
			Name:   "AR",
			Format: "%3s",
		},
		afterOuts: &dataframe.Column{
			Name:   "AO",
			Format: "%2d",
		},
		change: &dataframe.Column{
			Name:   "Change",
			Format: "% 6.3f",
		},
		narrative: &dataframe.Column{
			Name:   "Narrative",
			Format: "%-20s",
		},
	}
}

func (rea *REAnalysis) Run() *dataframe.Data {
	for _, c := range []struct {
		br, ar, n string
	}{
		{"___", "__1", "B reaches first"},
		{"___", "_2_", "B reaches 2nd"},
		{"___", "3__", "B reaches 3rd"},
		{"__1", "_2_", "R1 steals 2nd"},
		{"_2_", "3__", "R2 steals 3rd"},
		{"_2_", "_21", "R2, B walks"},
		{"3__", "3_1", "R3, B walks"},
		{"3_1", "32_", "R3, R1 steals 2"},
		{"32_", "321", "R2, R3, B walks"},
		{"__1", "_21", "R1, B singles"},
		{"_21", "321", "R1, R2, B singles"},
	} {
		for outs := 0; outs < 3; outs++ {
			rea.add(c.br, outs, c.ar, outs, c.n)
		}
	}
	rea.add("__1", 0, "_2_", 1, "Sac bunt")
	rea.add("__1", 0, "__1", 1, "B out")
	rea.add("__1", 1, "_2_", 2, "Sac bunt")
	rea.add("__1", 1, "__1", 2, "B out")
	rea.add("___", 0, "___", 1, "B out")
	rea.add("__1", 0, "__1", 1, "R1, B out")
	rea.add("__1", 1, "__1", 2, "R1, B out")
	rea.add("___", 2, "___", 3, "B last out")
	rea.add("__1", 2, "__1", 3, "B last out")
	rea.add("_2_", 2, "_2_", 3, "B last out")
	return &dataframe.Data{
		Columns: []*dataframe.Column{
			rea.beforeOuts, rea.beforeRunrs, rea.afterOuts, rea.afterRunrs, rea.change,
			rea.narrative,
		},
	}
}

func (rea *REAnalysis) add(br string, bo int, ar string, ao int, n string) {
	before := rea.RE.GetExpectedRuns(bo, Runners(br))
	rea.beforeRunrs.AppendString(br)
	rea.beforeOuts.AppendInts(bo)
	var after float64
	if ao != 3 {
		after = rea.RE.GetExpectedRuns(ao, Runners(ar))
	}
	rea.afterRunrs.AppendString(ar)
	rea.afterOuts.AppendInts(ao)
	rea.change.AppendFloats(after - before)
	rea.narrative.AppendString(n)
}
