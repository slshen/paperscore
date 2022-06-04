package dataframe

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert := assert.New(t)
	dat := &Data{
		Columns: []*Column{
			{Name: "Name", Values: []string{
				"George", "Thomas", "Henry",
			}},
			{Name: "Age", Summary: Average, Values: []int{
				52, 48, 57,
			}},
		},
	}
	fmt.Println(dat)
	s := &strings.Builder{}
	assert.NoError(dat.RenderCSV(s))
	fmt.Println(s.String())
	s.Reset()
	assert.NoError(dat.RenderMarkdown(s))
	fmt.Println(s.String())
	// t.Fail()
}

func TestSort(t *testing.T) {
	assert := assert.New(t)
	values := make([]float64, 5)
	ivalues := make([]int, len(values))
	for i := range values {
		// #nosec G404
		values[i] = 1000.0 * rand.Float64()
		ivalues[i] = i
	}
	dat := &Data{
		Columns: []*Column{
			{
				Name:   "X",
				Values: values,
				Format: "%5.1f",
			},
			{
				Name:   "Index",
				Values: ivalues,
			},
		},
	}
	idx := dat.GetIndex()
	col := idx.GetColumn("X")
	assert.NotNil(col)
	assert.Equal(5, col.Len())
	assert.Equal(Float, col.GetType())
	fmt.Println(dat)
	dat = dat.RSort(func(r1, r2 int) bool {
		f1 := col.GetFloat(r1)
		f2 := col.GetFloat(r2)
		fmt.Printf("[%d] %f < [%d] %f = %v\n", r1, f1, r2, f2, f1 < f2)
		return f1 < f2
	})
	fmt.Println(dat)
	assert.True(sort.Float64sAreSorted(dat.Columns[0].Values.([]float64)))
	// t.Fail()
}

func TestStructs(t *testing.T) {
	assert := assert.New(t)
	type S struct {
		I int
		F float64
		S string
		B bool
	}
	ss := []S{
		{1, 1.1, "one", true},
		{2, 2.2, "two", false},
	}
	dat, err := FromStructs("", ss)
	assert.NoError(err)
	assert.NotNil(dat)
	fmt.Println(dat)
	assert.NotNil(dat.GetIndex().GetColumn("I"))
	assert.NotNil(dat.GetIndex().GetColumn("B"))
	dat.Arrange("S")
	assert.Equal("S", dat.Columns[0].Name)
}

func TestFilterAndSelect(t *testing.T) {
	assert := assert.New(t)
	dat := &Data{
		Columns: []*Column{
			NewColumn("Rank", "%d", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
	}
	assert.Equal(10, dat.RowCount())
	dat = dat.RFilter(func(row int) bool { return row < 5 })
	assert.Equal(5, dat.RowCount())
	dat = dat.Select(
		Col("Rank"),
		DeriveInts("Rank2", func(idx *Index, i int) int {
			return idx.GetInt(i, "Rank") * 2
		}),
		DeriveFloats("RankSqf", func(idx *Index, i int) float64 {
			r := float64(idx.GetInt(i, "Rank"))
			return r * r
		}),
	)
	assert.Equal([]int{2, 4, 6, 8, 10}, dat.Columns[1].GetInts())
	assert.Equal([]float64{1, 4, 9, 16, 25}, dat.Columns[2].GetFloats())
	dat.Add(DeriveInts("Rank0", func(idx *Index, i int) int { return i }))
	assert.Equal([]int{0, 1, 2, 3, 4}, dat.Columns[3].GetInts())
	dat = dat.Select(
		Col("Rank"),
		Rename("Rank2", "R2"),
	)
	assert.Equal(2, len(dat.Columns))
	assert.Equal("R2", dat.Columns[1].Name)
}
