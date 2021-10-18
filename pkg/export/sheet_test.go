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

func TestColumnLetters(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("A", columnLetters(0))
	assert.Equal("B", columnLetters(1))
	assert.Equal("Z", columnLetters(25))
	assert.Equal("AA", columnLetters(26))
	assert.Equal("AB", columnLetters(27))
	assert.Equal("AZ", columnLetters(51))
	assert.Equal("BA", columnLetters(52))
}
