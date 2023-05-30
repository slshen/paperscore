package expr

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	assert := assert.New(t)
	lex, err := lexerDef.LexString("test", `ROE = 0.05 0xxx -> 1xxx 0321 0 1 -`)
	assert.NoError(err)
	toks, err := lexer.ConsumeAll(lex)
	if !assert.NoError(err) {
		return
	}
	for i, val := range []string{"ROE", "=", "0.05", "0xxx", "->", "1xxx", "0321", "0", "1", "-"} {
		assert.Less(i, len(toks))
		assert.Equal(val, toks[i].Value)
	}
}
