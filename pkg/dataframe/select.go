package dataframe

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
