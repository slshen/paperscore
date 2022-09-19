package sheetexport

import (
	"testing"

	"github.com/slshen/sb/pkg/config"
	"github.com/slshen/sb/pkg/dataframe"
	"github.com/stretchr/testify/assert"
)

func TestSheets(t *testing.T) {
	assert := assert.New(t)
	config, err := config.NewConfig()
	if err != nil {
		t.Skip(err)
	}
	config.SpreadsheetID = testSpreadsheet
	s, err := NewSheetExport(config)
	if !assert.NoError(err) {
		return
	}
	assert.NotNil(s)
	data := &dataframe.Data{
		Name: "test",
		Columns: []*dataframe.Column{
			{Name: "one", Values: []int{1, 10}},
			{Name: "two", Values: []string{"2", "20"}},
			{Name: "three", Values: []float64{3.0, 33.0}},
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
