package stats

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
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
		line := scan.Text()
		space := strings.Index(line, " ")
		name := line
		format := ""
		if space > 0 {
			name = line[0:space]
			format = line[space+1:]
		}
		col := &dataframe.Column{
			Name:   name,
			Format: format,
		}
		dat.Columns = append(dat.Columns, col)
	}
	return dat
}
