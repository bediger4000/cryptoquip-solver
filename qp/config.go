package qp

import "unicode"

func StringConfiguration(line string) string {

	if len(line) == 0 {
		return ""
	}

	key := make([]rune, len(line))
	idx := 0

	var scorecard [255]rune

	i := 0

	for _, r := range line {
		l := unicode.ToLower(r)
		if unicode.IsLetter(l) {
			if previous := scorecard[int(l)]; previous != 0 {
				key[idx] = previous
			} else {
				// single-quote appears before '0' in Unicode,
				// so only single-quotes that appear in input words
				// appear in "shape" of those words
				scorecard[l] = '0' + rune(i)
				i++
				key[idx] = scorecard[l]
			}
			idx++
			continue
		}
		if l == '\'' {
			key[idx] = l
			idx++
			continue
		}
	}

	return string(key)
}
