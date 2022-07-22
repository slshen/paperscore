package dataframe

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
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
	jdat, err := dat.MarshalJSON()
	assert.NoError(err)
	fmt.Println(string(jdat))
	var m map[string]interface{}
	err = json.Unmarshal(jdat, &m)
	assert.NoError(err)
	assert.Len(m["columnDefs"], 2)
	assert.Len(m["rowData"], 3)
	assert.NotNil(m["summaryRow"])
}
