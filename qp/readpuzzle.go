package qp

import (
	"bytes"
	"fmt"
	"os"
	"sort"
)

func ReadPuzzle(fileName string, verbose bool) ([][]byte, [][]byte, []rune, map[rune]rune, error) {
	buf, err := os.ReadFile(fileName)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "reading file %s: %v\n", fileName, err)
		}
		return nil, nil, nil, nil, err
	}

	uniquePuzzleWords := make(map[string]bool)
	var words [][]byte
	var enciphered, clear rune
	letters := make(map[rune]bool)
	hints := make(map[rune]rune)

	for _, line := range bytes.Split(buf, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			if bytes.Contains(line, []byte("Solution")) {
				break
			}
			continue
		}
		if bytes.ContainsRune(line, '=') {
			fields := bytes.Split(line, []byte{'='})
			enciphered = rune(fields[0][0])
			clear = rune(fields[len(fields)-1][0])
			hints[enciphered] = clear
			continue
		}
		for _, word := range bytes.Fields(line) {
			var wo []byte
			// Weed out some punctuation: .:,
			for i := range word {
				switch word[i] {
				case ':', '.', ',', '\'', '"':
				default:
					wo = append(wo, word[i])
					letters[rune(word[i])] = true
				}
			}
			uniquePuzzleWords[string(wo)] = true
			words = append(words, wo)
		}
	}

	var uniqueLetters []rune
	for l, _ := range letters {
		uniqueLetters = append(uniqueLetters, l)
	}

	sort.Sort(RuneSlice(uniqueLetters))

	upw := make([][]byte, 0, len(uniquePuzzleWords))
	for pw := range uniquePuzzleWords {
		upw = append(upw, []byte(pw))
	}

	return words, upw, uniqueLetters, hints, nil
}
