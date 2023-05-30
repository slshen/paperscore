package expr

import (
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/slshen/sb/pkg/markov"
)

type Position = lexer.Position

type File struct {
	Path       string
	Statements []*Statement `parser:"@@*"`
}

type Statement struct {
	EventDef    *EventDef         `parser:"@@"`
	Transitions *StateTransitions `parser:"| @@"`
}

type EventDef struct {
	Pos    Position
	Name   string      `parser:"@Ident"`
	Player string      `parser:"('[' @Ident ']')?"`
	Value  *Expression `parser:"'=' @@"`
}

type Expression struct {
	Addition *Addition `parser:"@@"`
}

type Addition struct {
	Multiplication *Multiplication `parser:"@@"`
	Op             string          `parser:"( @( '+' | '-' )"`
	Next           *Addition       `parser:" @@ )"`
}

type Multiplication struct {
	Unary *Unary `parser:"@@"`
}

type Unary struct {
	Op      string   `parser:"( @('!' | '-')"`
	Unary   *Unary   `parser:" @@ )"`
	Primary *Primary `parser:"| @@"`
}

type Primary struct {
	Number        *float64    `parser:"@Number"`
	Name          string      `parser:"| @Ident"`
	SubExpression *Expression `parser:"| '(' @@ ')'"`
}

type StateTransitions struct {
	Pos    Position
	From   string   `parser:"@BaseOutState '{'"`
	Events []*Event `parser:"@@* '}'"`

	norm float64
}

type Event struct {
	Pos  Position
	Name string  `parser:"@Ident"`
	Runs float64 `parser:"( '(' @Number ')' )?"`
	To   string  `parser:"'->' @BaseOutState"`

	to markov.BaseOutState
}

func rule(name, pattern string) lexer.SimpleRule {
	return lexer.SimpleRule{
		Name:    name,
		Pattern: pattern,
	}
}

var lexerDef = lexer.MustSimple(
	[]lexer.SimpleRule{
		rule("Ident", `[a-zA-Z_][a-zA-Z_0-9]*`),
		rule("BaseOutState", `[0-3][0123x]+`),
		rule("Number", `(([1-9][0-9]*)|0)(\.[0-9]*)?`),
		rule("whitespace", `\s+`),
		rule("comment", `#[^\n]+`),
		rule("Punct", `\[|]|[=+*(){}~]|(->)|-`),
	},
)

var parser = participle.MustBuild[File](
	participle.Lexer(lexerDef),
	participle.UseLookahead(3),
)

func ParseFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	file, err := parser.Parse(path, f)
	if err != nil {
		return nil, err
	}
	file.Path = path
	return file, nil
}
