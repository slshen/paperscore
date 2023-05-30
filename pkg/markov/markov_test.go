package markov

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseOutStateString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("0xxx", BaseOutState(0).String())
	assert.Equal("13x1", BaseOutState(13).String())
	assert.Equal("1321", FromOutsAndRunners(1, true, true, true).String())
	assert.Equal(BaseOutState(0), MustParseBaseOutState("0xxx"))
	assert.Equal(BaseOutState(1), MustParseBaseOutState("0xx1"))
	assert.Equal(BaseOutState(9), MustParseBaseOutState("1xx1"))
	assert.Equal(BaseOutState(7), MustParseBaseOutState("0321"))
	assert.Equal(BaseOutState(15), MustParseBaseOutState("1321"))
}
