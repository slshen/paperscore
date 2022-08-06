package dataframe

import (
	"fmt"
	"sort"
	"strings"
)

// Rotate is a limited form of a pivot table.  The columns of the rotated table
// will be the fixed columns, followed by columns for each value in the pivot
// column * the remaining columns.
func (dat *Data) Rotate(fixed []string, pivot string) *Data {
	rot := &Data{}
	idx := dat.GetIndex()
	fixedColumns := make([]*Column, len(fixed))
	for i, fix := range fixed {
		col := idx.GetColumn(fix)
		fixedColumns[i] = col
		rot.Columns = append(rot.Columns, col.EmptyCopy())
	}
	pivotColumn := idx.GetColumn(pivot)
	pivotValues := dat.getPivotValues(pivotColumn)
	pivotValueIndex := map[interface{}]int{}
	for i, val := range pivotValues {
		pivotValueIndex[val] = i
	}
	ridx := rot.GetIndex()
	var rotatedColumns []*Column
	for i, val := range pivotValues {
		for _, col := range dat.Columns {
			if col.Name == pivot || ridx.GetColumn(col.Name) != nil {
				continue
			}
			if i == 0 {
				rotatedColumns = append(rotatedColumns, col)
			}
			rcol := col.EmptyCopy()
			rcol.Name = fmt.Sprintf("%s-%s", val, col.Name)
			rot.Columns = append(rot.Columns, rcol)
		}
	}
	fixedValues := make([]interface{}, len(fixed))
	dat.RApply(func(row int) {
		newRow := false
		for i, fixedCol := range fixedColumns {
			value := fixedCol.GetValue(row)
			if !newRow && fixedValues[i] != value {
				newRow = true
			}
			fixedValues[i] = value
		}
		if newRow {
			for i, fixedCol := range fixedColumns {
				val := fixedCol.GetValue(row)
				rot.Columns[i].AppendValue(val)
			}
		}
		pivotVal := pivotColumn.GetValue(row)
		index := pivotValueIndex[pivotVal]
		for i, rotColumn := range rotatedColumns {
			val := rotColumn.GetValue(row)
			rot.Columns[len(fixed)+index*len(rotatedColumns)+i].AppendValue(val)
		}
	})
	return rot
}

func (dat *Data) getPivotValues(pivotCol *Column) []string {
	var pivotValues []string
	pivotValuesSeen := map[string]bool{}
	for i := 0; i < pivotCol.Len(); i++ {
		val := strings.TrimSpace(fmt.Sprintf(pivotCol.GetFormat(), pivotCol.GetValue(i)))
		if !pivotValuesSeen[val] {
			pivotValuesSeen[val] = true
			pivotValues = append(pivotValues, val)
		}
	}
	sort.Strings(pivotValues)
	return pivotValues
}
