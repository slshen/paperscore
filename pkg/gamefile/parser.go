package gamefile

import (
	"os"

	"github.com/alecthomas/participle/v2"
)

var parser = participle.MustBuild(&File{}, participle.Lexer(gameFileDef))

func ParseFile(path string) (*File, error) {
	file := &File{Path: path}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := parser.Parse(path, f, file); err != nil {
		return nil, err
	}
	return file, file.validate()
}
