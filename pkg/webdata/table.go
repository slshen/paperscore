package webdata

import (
	"encoding/json"
	"fmt"

	"github.com/slshen/sb/pkg/dataframe"
)

type Table struct {
	Data *dataframe.Data
}

func newTable(dat *dataframe.Data) *Table {
	return &Table{
		Data: dat,
	}
}

func (t *Table) GetPage() *Page {
	p := &Page{
		ID: ToID(t.Data.Name),
		Resources: map[string]ResourceContent{
			"data.json": func() ([]byte, error) {
				return json.Marshal(t.Data)
			},
		},
		Content: t.GetContent,
	}
	return p.Set("title", t.Data.Name)
}

func (t *Table) GetContent() string {
	return fmt.Sprintf("{{%% table %%}}\n")
}
