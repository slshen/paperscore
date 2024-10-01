package gamefile

import (
	"os"

	"github.com/alecthomas/participle/v2"
)

var Parser = participle.MustBuild[File](participle.Lexer(gameFileDef))

func ParseString(path string, text string) (*File, error) {
	file, err := Parser.ParseString(path, text)
	if err == nil {
		file.Path = path
		err = file.Validate()
	}
	return file, err
}

func ParseFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	file, err := Parser.Parse(path, f)
	if err != nil {
		return nil, err
	}
	file.Path = path
	return file, file.Validate()
}
