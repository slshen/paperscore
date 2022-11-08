package dataframe

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/slshen/sb/pkg/text"
)

type Update interface {
	Update()
}

type StructDecoder func(val interface{}) map[string]interface{}

type Data struct {
	Name    string
	Columns []*Column
}

func FromStructs(name string, values interface{}) (*Data, error) {
	dat := &Data{
		Name: name,
	}
	v := reflect.ValueOf(values)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("slice required, not %v", v.Kind())
	}
	var idx *Index
	for i := 0; i < v.Len(); i++ {
		val := v.Index(i).Interface()
		var err error
		idx = dat.AppendStruct(idx, val)
		if err != nil {
			return nil, err
		}
	}
	return dat, nil
}

func (dat *Data) AppendStruct(idx *Index, s interface{}) *Index {
	if u, ok := s.(Update); ok {
		u.Update()
	}
	var m map[string]interface{}
	err := mapstructure.Decode(s, &m)
	if err != nil {
		panic(err)
	}
	return dat.AppendMap(idx, m)
}

// Add values from a struct to the data.  Does not support time columns.
func (dat *Data) AppendMap(idx *Index, m map[string]interface{}) *Index {
	if idx == nil {
		idx = dat.GetIndex()
		for k, v := range m {
			if idx.GetColumn(k) == nil {
				switch v.(type) {
				case int:
				case float64:
				case string:
				case bool:
				default:
					continue
				}
				dat.Columns = append(dat.Columns, &Column{
					Name: k,
				})
			}
		}
		idx = dat.GetIndex()
	}
	for k, v := range m {
		col := idx.GetColumn(k)
		if col != nil {
			switch x := v.(type) {
			case bool:
				if x {
					col.AppendInt(1)
				} else {
					col.AppendInt(0)
				}
			case string:
				col.AppendString(x)
			case float64:
				col.AppendFloat(x)
			case int:
				col.AppendInt(x)
			default:
			}
		}
	}
	return idx
}

func (dat *Data) Arrange(names ...string) {
	idx := dat.GetIndex()
	cols := make([]*Column, len(dat.Columns))
	added := map[string]bool{}
	for i, name := range names {
		cols[i] = idx.GetColumn(name)
		added[name] = true
	}
	i := len(names)
	for _, col := range dat.Columns {
		if !added[col.Name] {
			added[col.Name] = true
			cols[i] = col
			i++
		}
	}
	dat.Columns = cols
}

func (dat *Data) RemoveColumn(name string) {
	cols := make([]*Column, 0, len(dat.Columns)-1)
	for _, col := range dat.Columns {
		if col.Name != name {
			cols = append(cols, col)
		}
	}
	dat.Columns = cols
}

func (dat *Data) GetIndex() *Index {
	idx := &Index{data: dat}
	idx.Update()
	return idx
}

func (dat *Data) RApply(f func(row int)) {
	rc := dat.RowCount()
	for r := 0; r < rc; r++ {
		f(r)
	}
}

func (dat *Data) GetRow(r int) []interface{} {
	row := make([]interface{}, len(dat.Columns))
	for i, col := range dat.Columns {
		switch col.GetType() {
		case Int:
			row[i] = col.GetInt(r)
		case Float:
			row[i] = col.GetFloat(r)
		case String:
			row[i] = col.GetString(r)
		}
	}
	return row
}

func (dat *Data) RFilter(f func(row int) bool) *Data {
	res := &Data{
		Name:    dat.Name,
		Columns: make([]*Column, len(dat.Columns)),
	}
	for i := range dat.Columns {
		res.Columns[i] = dat.Columns[i].EmptyCopy()
	}
	dat.RApply(func(row int) {
		if f(row) {
			for i := range dat.Columns {
				col := dat.Columns[i]
				rcol := res.Columns[i]
				switch col.GetType() {
				case Int:
					rcol.AppendInt(col.GetInt(row))
				case Float:
					rcol.AppendFloat(col.GetFloat(row))
				case String:
					rcol.AppendString(col.GetString(row))
				}
			}
		}
	})
	return res
}

func (dat *Data) RowCount() int {
	m := 0
	for _, col := range dat.Columns {
		if n := col.Len(); n > m {
			m = n
		}
	}
	return m
}

func (dat *Data) RSort(less func(r1 int, r2 int) bool) *Data {
	rc := dat.RowCount()
	rowNumbers := make([]int, rc)
	for i := 0; i < rc; i++ {
		rowNumbers[i] = i
	}
	sort.Slice(rowNumbers, func(i, j int) bool {
		ri := rowNumbers[i]
		rj := rowNumbers[j]
		return less(ri, rj)
	})
	res := &Data{
		Name:    dat.Name,
		Columns: make([]*Column, len(dat.Columns)),
	}
	for col, scol := range dat.Columns {
		rcol := &Column{
			Name:          scol.Name,
			Format:        scol.Format,
			Summary:       scol.Summary,
			SummaryFormat: scol.SummaryFormat,
		}
		res.Columns[col] = rcol
		switch scol.GetType() {
		case Int:
			values := make([]int, scol.Len())
			for row := 0; row < scol.Len(); row++ {
				values[row] = scol.GetInt(rowNumbers[row])
			}
			rcol.Values = values
		case Float:
			values := make([]float64, scol.Len())
			for row := 0; row < scol.Len(); row++ {
				values[row] = scol.GetFloat(rowNumbers[row])
			}
			rcol.Values = values
		case String:
			values := make([]string, scol.Len())
			for row := 0; row < scol.Len(); row++ {
				values[row] = scol.GetString(rowNumbers[row])
			}
			rcol.Values = values
		}
	}
	return res
}

func (dat *Data) Append(sdat *Data) {
	idx := sdat.GetIndex()
	for _, col := range dat.Columns {
		scol := idx.GetColumn(col.Name)
		if scol != nil {
			switch col.GetType() {
			case Int:
				col.Values = append(col.GetInts(), scol.GetInts()...)
			case Float:
				col.Values = append(col.GetFloats(), scol.GetFloats()...)
			case String:
				col.Values = append(col.GetStrings(), scol.GetStrings()...)
			}
		}
	}
}

func (dat *Data) HasSummary() bool {
	for _, col := range dat.Columns {
		if col.Summary != None {
			return true
		}
	}
	return false
}

func (dat *Data) String() string {
	s := &strings.Builder{}
	if dat.Name != "" {
		w := len(dat.Columns)
		for _, col := range dat.Columns {
			w += col.GetWidth()
		}
		fmt.Fprintln(s, text.Center(dat.Name, w))
	}
	f := &strings.Builder{}
	hasSummaryRow := dat.HasSummary()
	for i, col := range dat.Columns {
		if i > 0 {
			s.WriteRune(' ')
			f.WriteRune(' ')
		}
		fmt.Fprintf(s, "%s", text.Center(col.Name, col.GetWidth()))
		f.WriteString(col.GetFormat())
	}
	s.WriteRune('\n')
	f.WriteRune('\n')
	dat.RApply(func(row int) {
		r := dat.GetRow(row)
		fmt.Fprintf(s, f.String(), r...)
	})
	if hasSummaryRow && dat.RowCount() > 0 {
		for i, col := range dat.Columns {
			if i > 0 {
				s.WriteRune(' ')
			}
			val := ""
			if col.Summary != None {
				val = strings.Repeat("-", col.GetWidth())
			}
			fmt.Fprintf(s, "%*s", col.GetWidth(), val)
		}
		s.WriteRune('\n')
		for i, col := range dat.Columns {
			if i > 0 {
				s.WriteRune(' ')
			}
			if col.Summary != None {
				fmt.Fprintf(s, col.GetSummaryFormat(), col.GetSummary())
			} else {
				fmt.Fprintf(s, "%*s", col.GetWidth(), " ")
			}
		}
		s.WriteRune('\n')
	}
	return s.String()
}

func (dat *Data) RenderCSV(w io.Writer, withHeader bool) error {
	cw := csv.NewWriter(w)
	record := make([]string, len(dat.Columns))
	if withHeader {
		for i, col := range dat.Columns {
			record[i] = col.Name
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	var err error
	dat.RApply(func(row int) {
		if err != nil {
			return
		}
		for i, col := range dat.Columns {
			record[i] = strings.TrimSpace(fmt.Sprintf(col.GetFormat(), col.GetValue(row)))
		}
		err = cw.Write(record)
	})
	if err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

func (dat *Data) RenderMarkdown(w io.Writer) error {
	if dat.Name != "" {
		fmt.Fprintf(w, "# %s\n", dat.Name)
	}
	var hasSummary bool
	for _, col := range dat.Columns {
		hasSummary = hasSummary || col.Summary != None
		fmt.Fprintf(w, "| %*s ", col.GetWidth(), col.Name)
	}
	fmt.Fprintln(w, "|")
	for _, col := range dat.Columns {
		wd := col.GetWidth()
		if wd < 3 {
			wd = 3
		}
		fmt.Fprintf(w, "| %s ", strings.Repeat("-", wd))
	}
	fmt.Fprintln(w, "|")
	dat.RApply(func(row int) {
		for _, col := range dat.Columns {
			fmt.Fprintf(w, "| ")
			fmt.Fprintf(w, col.GetFormat(), col.GetValue(row))
			fmt.Fprintf(w, " ")
		}
		fmt.Fprintln(w, "|")
	})
	if hasSummary {
		for _, col := range dat.Columns {
			fmt.Fprintf(w, "| ")
			if col.Summary != None {
				fmt.Fprintf(w, col.GetSummaryFormat(), col.GetSummary())
			} else {
				fmt.Fprintf(w, "%*s", col.GetWidth(), "")
			}
			fmt.Fprintf(w, " ")
		}
		fmt.Fprintln(w, "|")
	}
	return nil
}

func (dat *Data) RenderMarkdownString() string {
	s := &strings.Builder{}
	_ = dat.RenderMarkdown(s)
	return s.String()
}
