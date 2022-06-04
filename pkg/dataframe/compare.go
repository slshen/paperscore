package dataframe

import "strings"

type Comparison func(r1, r2 int) int

func Descending(cmp Comparison) Comparison {
	return func(r1, r2 int) int {
		return -cmp(r1, r2)
	}
}

func CompareString(col *Column) Comparison {
	return func(r1, r2 int) int {
		s1 := col.GetString(r1)
		s2 := col.GetString(r2)
		return strings.Compare(s1, s2)
	}
}

func CompareInt(col *Column) Comparison {
	return func(r1, r2 int) int {
		i1 := col.GetInt(r1)
		i2 := col.GetInt(r2)
		if i1 < i2 {
			return -1
		}
		if i1 > i2 {
			return 1
		}
		return 0
	}
}

func CompareFloat(col *Column) Comparison {
	return func(r1, r2 int) int {
		f1 := col.GetFloat(r1)
		f2 := col.GetFloat(r2)
		if f1 < f2 {
			return -1
		}
		if f1 > f2 {
			return 1
		}
		return 0
	}
}

func Less(cmps ...Comparison) func(i, j int) bool {
	return func(i, j int) bool {
		for _, cmp := range cmps {
			c := cmp(i, j)
			if c != 0 {
				return c < 0
			}
		}
		return false
	}
}
