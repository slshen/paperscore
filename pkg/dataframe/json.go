package dataframe

import (
	"encoding/json"
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
	rows := make([]interface{}, dat.RowCount())
	m["rowData"] = rows
	dat.RApply(func(row int) {
		r := map[string]interface{}{}
		for _, col := range dat.Columns {
			switch col.GetType() {
			case Float:
				// round float to format
				r[col.Name] = RoundToFormat(col.GetFormat(), col.GetFloat(row))
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
				sum := col.GetSummary()
				switch val := sum.(type) {
				case float64:
					r[col.Name] = RoundToFormat(col.GetSummaryFormat(), val)
				default:
					r[col.Name] = val
				}
			}
		}
		m["summaryRow"] = r
	}
	return json.Marshal(m)
}
