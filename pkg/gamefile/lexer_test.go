package gamefile

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	assert := assert.New(t)
	lex, err := gameFileDef.Lex("test", strings.NewReader(`team: pride-2022
date: 5/30/22
comment: Game started late
empty:
no-nl: foo
---
plays us 1-2
1 00 S6/G6/B B-1 1-3(E3/TH) 2X3(635) : bunt single
... 2XH(82)
2 1 bbbb w`))
	assert.NoError(err)
	if !assert.NotNil(lex) {
		return
	}
	toks, err := lexer.ConsumeAll(lex)
	if !assert.NoError(err) {
		return
	}
	symbolsByType := map[lexer.TokenType]string{}
	for n, i := range gameFileDef.Symbols() {
		symbolsByType[i] = n
	}
	for i, expTok := range []struct{ name, value string }{
		{"Key", "team"},
		{"Value", "pride-2022"},
		{"NL", ""},
		{"Key", "date"},
		{"Value", "5/30/22"},
		{"NL", ""},
		{"Key", "comment"},
		{"Value", "Game started late"},
		{"NL", ""},
		{"Key", "empty"},
		{"NL", ""},
		{"Key", "no-nl"},
		{"Value", "foo"},
		{"NL", ""},
		{"Keyword", "plays"},
		{"Token", "us"},
		{"Token", "1-2"},
		{"NL", ""},
		{"PA", "1"},
		{"Token", "00"},
		{"Token", "S6/G6/B"},
		{"Advance", "B-1"},
		{"Advance", "1-3(E3/TH)"},
		{"Advance", "2X3(635)"},
		{"Comment", "bunt single"},
		{"NL", ""},
		{"PA", "..."},
		{"Advance", "2XH(82)"},
		{"NL", ""},
		{"PA", "2"},
		{"Token", "1"},
		{"Token", "bbbb"},
		{"Token", "w"},
		{"EOF", ""},
	} {
		tok := toks[i]
		if expTok.name != "" {
			tokt := gameFileDef.Symbols()[expTok.name]
			actualToken := symbolsByType[tok.Type]
			assert.Less(tokt, 0, "%s not token at %v for token type %s", tok.Value, tok.Pos, expTok.name)
			assert.Equal(tokt, tok.Type, "token type is %s (%s) not %s at %v", actualToken, tok.Value, expTok.name, tok.Pos)
		}
		if expTok.value != "" {
			assert.Equal(expTok.value, tok.Value, "token value not expected at %v", tok.Pos)
		}
	}
}
