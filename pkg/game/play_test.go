package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlay(t *testing.T) {
	assert := assert.New(t)
	single := Play("S6")
	assert.True(single.Hit())
	assert.True(single.Single())
	sb2, sb3, sbh := Play("SB2").StolenBase()
	assert.True(sb2)
	assert.False(sb3 || sbh)
}
