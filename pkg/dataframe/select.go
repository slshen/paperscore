package dataframe

import "fmt"

type Selection func(idx *Index) *Column

func (dat *Data) Select(sels ...Selection) *Data {
	idx := dat.GetIndex()
	res := &Data{
		Name:    dat.Name,
		Columns: make([]*Column, len(sels)),
	}
	for i, sel := range sels {
		col := sel(idx)
		if col == nil {
			panic("cannot select a column")
		}
		res.Columns[i] = col
	}
	return res
}

func (dat *Data) Add(sels ...Selection) {
	idx := dat.GetIndex()
	for _, sel := range sels {
		col := sel(idx)
		if col == nil {
			panic("cannot add nil column")
		}
		if idx.GetColumn(col.Name) != nil {
			panic(fmt.Sprintf("cannot add duplicate column %s", col.Name))
		}
		dat.Columns = append(dat.Columns, col)
	}
}

func Col(name string) Selection {
	return func(i *Index) *Column {
		return i.GetColumn(name)
	}
}

func Rename(name, newname string) Selection {
	return func(i *Index) *Column {
		col := i.GetColumn(name)
		return &Column{
			Name:          newname,
			Format:        col.Format,
			Summary:       col.Summary,
			SummaryFormat: col.SummaryFormat,
			Values:        col.Values,
		}
	}
}

func (sel Selection) WithSummary(st SummaryType) Selection {
	return func(i *Index) *Column {
		col := sel(i)
		col.Summary = st
		return col
	}
}

func (sel Selection) WithSummaryFormat(f string) Selection {
	return func(i *Index) *Column {
		col := sel(i)
		col.SummaryFormat = f
		return col
	}
}

func (sel Selection) WithFormat(f string) Selection {
	return func(i *Index) *Column {
		col := sel(i)
		col.Format = f
		return col
	}
}

func (sel Selection) WithPCT() Selection {
	return func(idx *Index) *Column {
		col := sel(idx)
		col.Format = "%4d"
		col.Summary = Average
		col.SummaryFormat = "%4.0f"
		return col
	}
}

func DeriveInts(name string, f func(idx *Index, i int) int) Selection {
	return func(idx *Index) *Column {
		values := make([]int, idx.data.RowCount())
		for i := range values {
			values[i] = f(idx, i)
		}
		return &Column{
			Name:   name,
			Values: values,
		}
	}
}

func DeriveFloats(name string, f func(idx *Index, i int) float64) Selection {
	return func(idx *Index) *Column {
		values := make([]float64, idx.data.RowCount())
		for i := range values {
			values[i] = f(idx, i)
		}
		return &Column{
			Name:   name,
			Values: values,
		}
	}
}

func DeriveStrings(name string, f func(idx *Index, i int) string) Selection {
	return func(idx *Index) *Column {
		values := make([]string, idx.data.RowCount())
		for i := range values {
			values[i] = f(idx, i)
		}
		return &Column{
			Name:   name,
			Values: values,
		}
	}
}
