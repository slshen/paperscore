package stats

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/slshen/sb/pkg/text"
)

type Data struct {
	Name            string
	Columns         []string
	Rows            []Row
	Width           map[string]int
	RestrictColumns []string
}

type Row []interface{}

func (d *Data) RenderTable(w io.Writer) {
	tab := text.Table{}
	for _, col := range d.Columns {
		if !d.includeColmumn(col) {
			continue
		}
		var w int
		if d.Width != nil {
			w = d.Width[col]
		}
		if w == 0 {
			w = len(col)
			if w < 3 {
				w = 3
			}
		}
		tab.Columns = append(tab.Columns,
			text.Column{
				Header: fmt.Sprintf("%*s", len(col), col),
				Width:  w,
			})
	}
	fmt.Fprint(w, tab.Header())
	for i := range d.Rows {
		row := d.getRow(i)
		fmt.Fprintf(w, tab.Format(), row...)
	}
}

func (d *Data) includeColmumn(col string) bool {
	if len(d.RestrictColumns) == 0 {
		return true
	}
	col = strings.ToLower(col)
	for _, r := range d.RestrictColumns {
		if strings.HasPrefix(col, strings.ToLower(r)) {
			return true
		}
	}
	return false
}

func (d *Data) getRow(i int) Row {
	if len(d.RestrictColumns) == 0 {
		return d.Rows[i]
	}
	var row Row
	for j := range d.Rows[i] {
		if d.includeColmumn(d.Columns[j]) {
			row = append(row, d.Rows[i][j])
		}
	}
	return row
}

func (d *Data) RenderCSV(w io.Writer) error {
	csv := csv.NewWriter(w)
	columns := d.Columns
	if len(d.RestrictColumns) > 0 {
		columns = nil
		for _, col := range d.Columns {
			if d.includeColmumn(col) {
				columns = append(columns, col)
			}
		}
	}
	if err := csv.Write(columns); err != nil {
		return err
	}
	record := make([]string, len(columns))
	for _, row := range d.Rows {
		i := 0
		for j := range d.Columns {
			if d.includeColmumn(d.Columns[j]) {
				record[i] = fmt.Sprintf("%v", row[j])
				i++
			}
		}
		if err := csv.Write(record); err != nil {
			return err
		}
	}
	csv.Flush()
	return nil
}
