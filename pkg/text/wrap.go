package text

import (
	krtext "github.com/kr/text"
	"github.com/mitchellh/go-wordwrap"
)

func Wrap(s string, width int) string {
	return wordwrap.WrapString(s, uint(width))
}

func WrapIndent(s string, width int, indent string) string {
	return krtext.Indent(Wrap(s, width), indent)
}
