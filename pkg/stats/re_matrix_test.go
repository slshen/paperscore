package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestREMatrix(t *testing.T) {
	assert := assert.New(t)
	re, err := ReadREMatrix("../../data/mlb_re_2010-2015.csv")
	assert.NoError(err)
	assert.NotNil(re)
}
