package gamefile

import "github.com/alecthomas/participle/v2/lexer"

var gameFileDef = lexer.MustStateful(
	lexer.Rules{
		"Root": {
			rule("Key", `[A-Za-z][-_A-Za-z0-9]*`, nil),
			rule("valueStart", `:[ \t]*`, lexer.Push("PropertyValue")),
			rule("whitespace", `[ \t]+`, nil),
			rule("dashes", `---[\n\r]*`, lexer.Push("Events")),
			rule("NL", `[\n\r]`, nil),
			rule("comment", `//.*[\n\r]`, nil),
		},
		"PropertyValue": {
			rule("Value", `[^\n\r]+`, nil),
			rule("NL", `[\n\r]`, lexer.Pop()),
		},
		"Events": {
			rule("PA", `[1-9][0-9]*|alt|\.\.\.`, lexer.Push("PA")),
			rule("Keyword", `[^ \t\n\r]+`, lexer.Push("Command")),
			rule("NL", `[\n\r]`, nil),
			rule("whitespace", `[ \t]+`, nil),
			rule("comment", `//.*[\n\r]`, nil),
		},
		"PA": {
			rule("Advance", `[Bb123][-Xx][123Hh]([^ \t\n\r]*)`, nil),
			rule("colon", `(:|--)[ \t]*`, lexer.Push("PAComment")),
			rule("NL", `[\n\r]`, lexer.Pop()),
			rule("Token", `[^ \t\n\r]+`, nil),
			rule("whitespace", `[ \t]+`, nil),
			rule("comment", `//.*[\n\r]`, nil),
		},
		"Command": {
			rule("Token", `[^ \t\n\r]+`, nil),
			rule("NL", `[\n\r]`, lexer.Pop()),
			rule("whitespace", `[ \t]+`, nil),
			rule("comment", `//.*[\n\r]`, nil),
		},
		"PAComment": {
			rule("Comment", "[^\n\r]+", nil),
			rule("comment", `//.*[\n\r]`, nil),
			lexer.Return(),
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
