package gamefile

import "github.com/alecthomas/participle/v2/lexer"

var gameFileDef = lexer.MustStateful(
	lexer.Rules{
		"Root": {
			rule("Ident", `[A-Za-z][-_A-Za-z0-9]*`, nil),
			rule("whitespace", `[ \t]+`, nil),
			rule("textStart", `:[ \t]*`, lexer.Push("Text")),
			rule("dashes", `---[\n\r]`, lexer.Push("Plays")),
			rule("NL", `[\n\r]`, nil),
			rule("comment", `//.*[\n\r]`, nil),
		},
		"Plays": {
			rule("Numbers", `[0-9]+[ \t]`, nil),
			rule("Keyword", `[a-z][-a-z0-9]*`, nil),
			rule("Dots", `\.\.\.`, nil),
			rule("Code", `[.0-9A-Z][^ \n\t]*`, nil),
			rule("NL", `[\n\r]`, nil),
			rule("whitespace", `[ \t]+`, nil),
			rule("textStart", `:[ \t]*`, lexer.Push("Text")),
			rule("comment", `//.*[\n\r]`, nil),
		},
		"Text": {
			rule("Text", "[^\n\r]+", nil),
			rule("NL", `[\n\r]`, lexer.Pop()),
			rule("comment", `//.*[\n\r]`, nil),
		},
	},
)

func rule(name, pattern string, action lexer.Action) lexer.Rule {
	return lexer.Rule{
		Name:    name,
		Pattern: pattern,
		Action:  action,
	}
}
