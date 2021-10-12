package text

import "fmt"

var ordinalDictionary = map[int]string{
	0: "th",
	1: "st",
	2: "nd",
	3: "rd",
	4: "th",
	5: "th",
	6: "th",
	7: "th",
	8: "th",
	9: "th",
}

func Ordinal(i int) string {
	return fmt.Sprintf("%d%s", i, ordinalDictionary[i%10])
}
