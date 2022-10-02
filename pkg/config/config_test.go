package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnviron(t *testing.T) {
	t.Setenv("SOFTBALL_SHEET_JSON_KEY", "/tmp/foo.json")
	config := GetConfig()
	assert := assert.New(t)
	assert.Equal("/tmp/foo.json", config.GetString("SheetJsonKey"))
}
