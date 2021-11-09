package stats

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"strings"
)

//go:embed *_columns.txt
var columnsFS embed.FS

type dataMaker struct {
	columnIndexes map[string]int
	data          *Data
}

func newDataMaker(name string) *dataMaker {
	dm := &dataMaker{
		columnIndexes: make(map[string]int),
		data: &Data{
			Name: name,
			Width: map[string]int{
				"Name": 12,
				"RE24": 6,
			},
		},
	}
	cols, err := columnsFS.ReadFile(fmt.Sprintf("%s_columns.txt", strings.ToLower(name)))
	if err != nil {
		panic(err)
	}
	for scan := bufio.NewScanner(bytes.NewReader(cols)); scan.Scan(); {
		col := scan.Text()
		dm.columnIndexes[col] = len(dm.columnIndexes)
		dm.data.Columns = append(dm.data.Columns, col)
	}
	return dm
}

func (dm *dataMaker) addRow(m map[string]interface{}) {
	row := make([]interface{}, len(dm.columnIndexes))
	for k, v := range m {
		if _, ok := dm.columnIndexes[k]; !ok {
			panic(fmt.Sprintf("no column named %s defined in %s", k, dm.data.Name))
		}
		row[dm.columnIndexes[k]] = v
	}
	dm.data.Rows = append(dm.data.Rows, row)
}
