package dataframe

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWidth(t *testing.T) {
	assert := assert.New(t)
	col := &Column{Format: "%-10s"}
	assert.Equal(10, col.GetWidth())
	col.Format = "% 6.2f"
	assert.Equal(6, col.GetWidth())
	assert.Equal(" 10.48", fmt.Sprintf("% 6.2f", 10.48))
	assert.Equal(" 10.48", fmt.Sprintf("% 6.2f", 10.48))
}

func TestAppend(t *testing.T) {
	assert := assert.New(t)
	col := &Column{}
	col.AppendInts(1, 2, 3, 4, 5)
	assert.Equal([]int{1, 2, 3, 4, 5}, col.GetInts())
}
