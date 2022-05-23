package export

import (
	"testing"

	"github.com/slshen/sb/pkg/game"
	"github.com/stretchr/testify/assert"
)

const testSpreadsheet = "1-2TCHib_hZ41GkAuJFtXF7Ejec5v2qBwVr-PwRKI5u0"

func TestExport(t *testing.T) {
	assert := assert.New(t)
	config, err := NewConfig()
	if err != nil {
		t.Skip(err)
	}
	config.SpreadsheetID = testSpreadsheet
	sheets, err := NewSheetExport(config)
	if !assert.NoError(err) {
		return
	}
	export, err := NewExport(sheets, nil)
	if !assert.NoError(err) {
		return
	}
	export.Us = "pride"
	files := []string{"../../data/2021/20211017-1.yaml"}
	games, err := game.ReadGameFiles(files)
	if !assert.NoError(err) {
		return
	}
	assert.NoError(export.Export(games))
}
