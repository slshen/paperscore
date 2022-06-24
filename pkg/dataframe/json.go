package dataframe

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (dat *Data) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	cols := []interface{}{}
	for _, col := range dat.Columns {
		cols = append(cols, map[string]interface{}{
			"field": col.Name,
			"type":  col.GetType().String(),
		})
	}
	m["columnDefs"] = cols
	rc := dat.RowCount()
	n := rc
	if dat.HasSummary() {
		n++
	}
	rows := make([]interface{}, n)
	m["rowData"] = rows
	dat.RApply(func(row int) {
		r := map[string]interface{}{}
		for _, col := range dat.Columns {
			switch col.GetType() {
			case Float:
				// round float to format
				s := strings.Trim(fmt.Sprintf(col.GetFormat(), col.GetFloat(row)), " ")
				n, _ := strconv.ParseFloat(s, 64)
				r[col.Name] = n
			case Int:
				r[col.Name] = col.GetInt(row)
			case String:
				r[col.Name] = col.GetString(row)
			}
		}
		rows[row] = r
	})
	if dat.HasSummary() {
		r := map[string]interface{}{}
		for _, col := range dat.Columns {
			if col.Summary != None {
				r[col.Name] = col.GetSummary()
			}
		}
		rows[rc] = r
	}
	return json.Marshal(m)
}
