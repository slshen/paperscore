package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumns(t *testing.T) {
	assert := assert.New(t)
	dat := newData("BAT")
	idx := dat.GetIndex()
	col := idx.GetColumn("RE24")
	assert.NotNil(col)
	assert.Equal("% 6.2f", col.Format)
}
