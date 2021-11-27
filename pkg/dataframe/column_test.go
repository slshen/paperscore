package dataframe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWidth(t *testing.T) {
	assert := assert.New(t)
	col := &Column{Format: "%-10s"}
	assert.Equal(10, col.GetWidth())
}

func TestAppend(t *testing.T) {
	assert := assert.New(t)
	col := &Column{}
	col.AppendInts(1, 2, 3, 4, 5)
	assert.Equal([]int{1, 2, 3, 4, 5}, col.GetInts())
}
