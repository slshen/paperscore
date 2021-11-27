package dataframe

import (
	"fmt"
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
	// assert.FailNow("")
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
