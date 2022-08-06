package dataframe

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRotate(t *testing.T) {
	dat := &Data{
		Columns: []*Column{
			{Name: "Category", Values: []string{
				"cats", "cats", "cats", "cats",
				"dogs", "dogs", "dogs", "dogs",
				"fish", "fish", "fish", "fish",
				"birds", "birds", "birds", "birds",
			}},
			{Name: "Qtr", Values: []string{
				"1st", "2nd", "3rd", "4th",
				"1st", "2nd", "3rd", "4th",
				"1st", "2nd", "3rd", "4th",
				"1st", "2nd", "3rd", "4th",
			}},
			{Name: "Count", Values: []int{
				1, 2, 3, 4,
				5, 6, 7, 8,
				9, 10, 11, 12,
				13, 14, 15, 16,
			}},
			{Name: "Weight", Values: []float64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 20, 21,
				22, 23, 24, 25,
			}},
		},
	}
	dat = dat.RSort(Less(CompareString(dat.Columns[0]),
		CompareString(dat.Columns[1])))
	fmt.Println(dat)
	rot := dat.Rotate([]string{"Category"}, "Qtr")
	fmt.Println(rot)
	assert.Equal(t, 4, rot.RowCount())
}
