package qp

import (
	"bytes"
	"fmt"
	"os"
)

func ReadPuzzle(fileName string, verbose bool) ([][]byte, rune, rune, error) {
	buf, err := os.ReadFile(fileName)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "reading file %s: %v\n", fileName, err)
		}
		return nil, ' ', ' ', err
	}

	var words [][]byte
	var enciphered, clear rune

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
			continue
		}
		for _, word := range bytes.Fields(line) {
			var wo []byte
			// Weed out some punctuation: .:,
			for i := range word {
				switch word[i] {
				case ':', '.', ',':
				default:
					wo = append(wo, word[i])
				}
			}
			words = append(words, wo)
		}
	}

	return words, enciphered, clear, nil
}
