package stats

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
)

//go:embed *_columns.txt
var columnsFS embed.FS

func newData(name string) *dataframe.Data {
	dat := &dataframe.Data{
		Name: name,
	}
	cols, err := columnsFS.ReadFile(fmt.Sprintf("%s_columns.txt", strings.ToLower(name)))
	if err != nil {
		panic(err)
	}
	for scan := bufio.NewScanner(bytes.NewReader(cols)); scan.Scan(); {
		parts := strings.Split(scan.Text(), " ")
		col := &dataframe.Column{
			Name: parts[0],
		}
		if len(parts) == 2 {
			col.Format = parts[1]
		}
		dat.Columns = append(dat.Columns, col)
	}
	return dat
}
