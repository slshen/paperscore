package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCenter(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("  ello", Center("ello", 8))
	assert.Equal(" Hello", Center("Hello", 8))
}
