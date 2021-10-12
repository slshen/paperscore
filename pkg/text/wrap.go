package text

import krtext "github.com/kr/text"

var Wrap = krtext.Wrap

func WrapIndent(s string, width int, indent string) string {
	return krtext.Indent(krtext.Wrap(s, width), indent)
}
