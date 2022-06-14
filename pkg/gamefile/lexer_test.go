package gamefile

import (
	"fmt"
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
...`))
	assert.NoError(err)
	if !assert.NotNil(lex) {
		return
	}
	toks, err := lexer.ConsumeAll(lex)
	if !assert.NoError(err) {
		return
	}
	for n, i := range gameFileDef.Symbols() {
		fmt.Println(n, i)
	}
	for i, expTok := range []struct{ name, value string }{
		{"Ident", "team"},
		{"Text", "pride-2022"},
		{"NL", ""},
		{"Ident", "date"},
		{"Text", "5/30/22"},
		{"NL", ""},
		{"Ident", "comment"},
		{"Text", "Game started late"},
		{"NL", ""},
		{"Ident", "empty"},
		{"NL", ""},
		{"Ident", "no-nl"},
		{"Text", "foo"},
		{"NL", ""},
		{"Keyword", "plays"},
		{"Keyword", "us"},
		{"Code", "1-2"},
		{"NL", ""},
		{"Numbers", "1 "},
		{"Numbers", "00 "},
		{"Code", "S6/G6/B"},
		{"Code", "B-1"},
		{"Code", "1-3(E3/TH)"},
		{"Code", "2X3(635)"},
		{"Text", "bunt single"},
		{"NL", ""},
		{"Dots", "..."},
		{"EOF", ""},
	} {
		tok := toks[i]
		if expTok.name != "" {
			tokt := gameFileDef.Symbols()[expTok.name]
			assert.Less(tokt, 0, "%s not token at %v", tok.Value, tok.Pos)
			assert.Equal(tokt, tok.Type, "token not %s at %v", expTok.name, tok.Pos)
		}
		if expTok.value != "" {
			assert.Equal(expTok.value, tok.Value)
		}
	}
}
