package gamefile

import (
	"os"

	"github.com/alecthomas/participle/v2"
)

var parser = participle.MustBuild[File](participle.Lexer(gameFileDef))

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
	return file, file.Validate()
}
