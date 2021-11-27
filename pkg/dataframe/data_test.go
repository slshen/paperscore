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
	}
	ss := []S{
		{1, 1.1, "one"},
		{2, 2.2, "two"},
	}
	dat, err := FromStructs("", ss)
	assert.NoError(err)
	assert.NotNil(dat)
	fmt.Println(dat)
	// assert.FailNow("")
}
