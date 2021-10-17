package export

import (
	"testing"

	"github.com/slshen/sb/pkg/stats"
	"github.com/stretchr/testify/assert"
)

func TestSheets(t *testing.T) {
	assert := assert.New(t)
	config, err := NewConfig()
	if err != nil {
		t.Skip(err)
	}
	config.SpreadsheetID = testSpreadsheet
	s, err := NewSheetExport(config)
	if !assert.NoError(err) {
		return
	}
	assert.NotNil(s)
	data := &stats.Data{
		Name:    "test",
		Columns: []string{"one", "two", "three"},
		Rows: []stats.Row{
			{1, 2, 3},
			{"four", "five", "six"},
		},
	}
	assert.NoError(s.ExportData(data))
}
