package dataframe

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Update interface {
	Update()
}

type Data struct {
	Name    string
	Columns []*Column
}

type Index struct {
	idx  map[string]int
	data *Data
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
		idx, err = dat.AppendStruct(idx, val)
		if err != nil {
			return nil, err
		}
	}
	return dat, nil
}

func (dat *Data) AppendStruct(idx *Index, s interface{}) (*Index, error) {
	var m map[string]interface{}
	if u, ok := s.(Update); ok {
		u.Update()
	}
	if err := mapstructure.Decode(s, &m); err != nil {
		return nil, err
	}
	if idx == nil {
		idx = dat.GetIndex()
		for k, v := range m {
			if idx.GetColumn(k) == nil {
				_, ok := v.(int)
				if !ok {
					_, ok = v.(float64)
					if !ok {
						_, ok = v.(string)
						if !ok {
							_, ok = v.(bool)
						}
					}
				}
				if ok {
					dat.Columns = append(dat.Columns, &Column{
						Name: k,
					})
				}
			}
		}
		idx = dat.GetIndex()
	}
	for k, v := range m {
		col := idx.GetColumn(k)
		if i, ok := v.(int); ok {
			col.AppendInts(i)
			continue
		}
		if f, ok := v.(float64); ok {
			col.AppendFloats(f)
			continue
		}
		if s, ok := v.(string); ok {
			col.AppendString(s)
			continue
		}
		if b, ok := v.(bool); ok {
			if b {
				col.AppendInts(1)
			} else {
				col.AppendInts(0)
			}
			continue
		}
	}
	return idx, nil
}

func (dat *Data) Arrange(names []string) {
	idx := dat.GetIndex()
	cols := make([]*Column, len(dat.Columns))
	for i, name := range names {
		cols[i] = idx.GetColumn(name)
	}
	dat.Columns = cols
}

func (dat *Data) GetIndex() *Index {
	idx := map[string]int{}
	for i, col := range dat.Columns {
		idx[col.Name] = i
	}
	return &Index{
		idx:  idx,
		data: dat,
	}
}

func (dat *Data) Select(names []string) *Data {
	res := &Data{
		Columns: make([]*Column, len(names)),
	}
	idx := dat.GetIndex()
	for i, name := range names {
		res.Columns[i] = idx.GetColumn(name)
	}
	return res
}

func (dat *Data) SelectFunc(f func(name string) bool) *Data {
	var names []string
	for _, col := range dat.Columns {
		if f(col.Name) {
			names = append(names, col.Name)
		}
	}
	return dat.Select(names)
}

func (dat *Data) RApply(f func(row []interface{})) {
	r := 0
	for {
		row := dat.Row(r)
		if row == nil {
			break
		}
		f(row)
		r++
	}
}

func (dat *Data) Row(r int) []interface{} {
	var row []interface{}
	for i, col := range dat.Columns {
		if r < col.Len() {
			if row == nil {
				row = make([]interface{}, len(dat.Columns))
			}
			switch col.GetType() {
			case Int:
				row[i] = col.GetInts()[r]
			case Float:
				row[i] = col.GetFloats()[r]
			case String:
				row[i] = col.GetStrings()[r]
			}
		}
	}
	return row
}

func (dat *Data) RFilter(f func(row []interface{}) bool) *Data {
	res := &Data{
		Name:    dat.Name,
		Columns: make([]*Column, len(dat.Columns)),
	}
	for i := range dat.Columns {
		res.Columns[i] = dat.Columns[i].EmptyCopy()
	}
	dat.RApply(func(row []interface{}) {
		if f(row) {
			for i := range dat.Columns {
				col := dat.Columns[i]
				rcol := res.Columns[i]
				switch col.GetType() {
				case Int:
					rcol.AppendInts(row[i].(int))
				case Float:
					rcol.AppendFloats(row[i].(float64))
				case String:
					rcol.AppendString(row[i].(string))
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

func (dat *Data) RSort(less func(r1 []interface{}, r2 []interface{}) bool) *Data {
	rows := make([][]interface{}, dat.RowCount())
	rc := dat.RowCount()
	for i := 0; i < rc; i++ {
		rows[i] = dat.Row(i)
	}
	sort.Slice(rows, func(i, j int) bool {
		return less(rows[i], rows[j])
	})
	res := &Data{
		Name:    dat.Name,
		Columns: make([]*Column, len(dat.Columns)),
	}
	for i, col := range dat.Columns {
		rcol := &Column{
			Name:   col.Name,
			Format: col.Format,
		}
		res.Columns[i] = rcol
		switch col.GetType() {
		case Int:
			values := make([]int, col.Len())
			for j := range values {
				values[j] = rows[j][i].(int)
			}
			rcol.Values = values
		case Float:
			values := make([]float64, col.Len())
			for j := range values {
				values[j] = rows[j][i].(float64)
			}
			rcol.Values = values
		case String:
			values := make([]string, col.Len())
			for j := range values {
				values[j] = rows[j][i].(string)
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

func (dat *Data) String() string {
	s := &strings.Builder{}
	f := &strings.Builder{}
	var hasSummaryRow bool
	for i, col := range dat.Columns {
		hasSummaryRow = hasSummaryRow || col.Summary != None
		if i > 0 {
			s.WriteRune(' ')
			f.WriteRune(' ')
		}
		fmt.Fprintf(s, "%*s", col.GetWidth(), col.Name)
		f.WriteString(col.GetFormat())
	}
	s.WriteRune('\n')
	f.WriteRune('\n')
	dat.RApply(func(row []interface{}) {
		fmt.Fprintf(s, f.String(), row...)
	})
	if hasSummaryRow {
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

func (dat *Data) RenderCSV(w io.Writer) error {
	cw := csv.NewWriter(w)
	record := make([]string, len(dat.Columns))
	for i, col := range dat.Columns {
		record[i] = col.Name
	}
	if err := cw.Write(record); err != nil {
		return err
	}
	var err error
	dat.RApply(func(row []interface{}) {
		if err != nil {
			return
		}
		for i := range row {
			record[i] = fmt.Sprintf("%v", row[i])
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
	var hasSummary bool
	for _, col := range dat.Columns {
		hasSummary = hasSummary || col.Summary != None
		fmt.Fprintf(w, "| %*s ", col.GetWidth(), col.Name)
	}
	fmt.Fprintln(w, "|")
	for _, col := range dat.Columns {
		fmt.Fprintf(w, "| %s ", strings.Repeat("-", col.GetWidth()))
	}
	fmt.Fprintln(w, "|")
	dat.RApply(func(row []interface{}) {
		for i, col := range dat.Columns {
			fmt.Fprintf(w, "| ")
			fmt.Fprintf(w, col.GetFormat(), row[i])
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

type ColumnRename [2]string

func (dat *Data) SelectRename(cols []ColumnRename) *Data {
	idx := dat.GetIndex()
	res := &Data{
		Name:    dat.Name,
		Columns: make([]*Column, len(cols)),
	}
	for i, cr := range cols {
		col := idx.GetColumn(cr[0])
		if col == nil {
			panic("no column " + cr[0])
		}
		res.Columns[i] = &Column{
			Name:   cr[1],
			Format: col.Format,
			Values: col.Values,
		}
	}
	return res
}

func (idx *Index) GetColumn(name string) *Column {
	i, ok := idx.idx[name]
	if ok {
		return idx.data.Columns[i]
	}
	return nil
}

func (idx *Index) GetValue(row []interface{}, name string) interface{} {
	i, ok := idx.idx[name]
	if ok {
		return row[i]
	}
	return nil
}
