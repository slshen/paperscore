package text

import (
	"strings"
)

// NameShorten returns a string made from the first letter of each word (for words longer than
// 2 characters) except the last word, which is included in full, unless the last word is a single
// character.
// Example: "John Ronald Reuel Tolkien" -> "JRR Tolkien"
// Example: "Angie W" -> "Angie W"
func NameShorten(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	if len(words) == 1 {
		return words[0]
	}
	if last := words[len(words)-1]; len(last) == 1 {
		return s
	}
	var b strings.Builder
	for _, word := range words[:len(words)-1] {
		if len(word) <= 2 {
			b.WriteString(word)
		} else {
			b.WriteByte(word[0])
		}
	}
	b.WriteByte(' ')
	b.WriteString(words[len(words)-1])
	return b.String()
}

func Initialize(s string) string {
	var b strings.Builder
	for _, word := range strings.Fields(s) {
		for i, ch := range word {
			if i == 0 || (ch >= 'A' && ch <= 'Z') {
				b.WriteRune(ch)
			} else {
				break
			}
		}
	}
	return b.String()
}
