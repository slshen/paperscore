package dataframe

type Index struct {
	idx  map[string]int
	data *Data
}

func (idx *Index) GetData() *Data {
	return idx.data
}

func (idx *Index) Update() {
	idx.idx = map[string]int{}
	for i, col := range idx.data.Columns {
		idx.idx[col.Name] = i
	}
}

func (idx *Index) GetColumn(name string) *Column {
	i, ok := idx.idx[name]
	if ok {
		return idx.data.Columns[i]
	}
	return nil
}

func (idx *Index) GetValue(row int, name string) interface{} {
	return idx.GetColumn(name).GetValue(row)
}

func (idx *Index) GetInt(row int, name string) int {
	return idx.GetColumn(name).GetInt(row)
}

func (idx *Index) GetFloat(row int, name string) float64 {
	return idx.GetColumn(name).GetFloat(row)
}

func (idx *Index) GetString(row int, name string) string {
	return idx.GetColumn(name).GetString(row)
}
