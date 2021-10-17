package stats

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/slshen/sb/pkg/text"
)

type Data struct {
	Name    string
	Columns []string
	Rows    []Row
	Width   map[string]int
}

type Row []interface{}

func (d *Data) RenderTable(w io.Writer) {
	tab := text.Table{}
	for _, col := range d.Columns {
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
	for _, row := range d.Rows {
		fmt.Fprintf(w, tab.Format(), row...)
	}
}

func (d *Data) RenderCSV(w io.Writer) error {
	csv := csv.NewWriter(w)
	if err := csv.Write(d.Columns); err != nil {
		return err
	}
	record := make([]string, len(d.Columns))
	for _, row := range d.Rows {
		for i := range row {
			record[i] = fmt.Sprintf("%v", row[i])
		}
		if err := csv.Write(record); err != nil {
			return err
		}
	}
	csv.Flush()
	return nil
}
