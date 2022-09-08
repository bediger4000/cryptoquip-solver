package qp

import (
	"bufio"
	"log"
	"os"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Entry represents the letters of the words having a particular configuration
// Enciphered word "zoyzkojvx" has configuration "012031456", shared with
// 15 words in /usr/share/dict/words.
type Entry struct {
	Length int
	Runes  map[int]map[rune]bool
	// Runes[0] are all the dictionary words' first letters.
	// the first letters are keys in the map-that-is-the-value
}

func NewDict(fileName string) (map[string][]string, error) {
	fin, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	d := make(map[string][]string)

	scanner := bufio.NewScanner(fin)

	lineCounter := 0

	for scanner.Scan() {
		lineCounter++
		line := scanner.Text()
		line = strings.ToLower(line)
		line = norm.NFC.String(line)

		config := StringConfiguration(line)
		d[config] = append(d[config], line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("problem line %d: %v", lineCounter, err)
	}

	return d, nil
}

func NewRunesDict(wordDict map[string][]string) map[string]*Entry {
	d := make(map[string]*Entry)

	for configuration, words := range wordDict {
		if entry, ok := d[configuration]; ok {
			// encountered this configuration before
			for _, word := range words {
				for idx, r := range word {
					if r > 'z' || r < 'a' {
						continue
					}
					if found := entry.Runes[idx][r]; !found {
						// haven't seen this letter at this position
						entry.Runes[idx][r] = true
					}
				}
			}
		} else {
			// new to us configuration
			var e Entry
			e.Length = len(configuration)
			e.Runes = make(map[int]map[rune]bool)
			for _, word := range words {
				for idx, r := range word {
					if r > 'z' || r < 'a' {
						continue
					}
					if _, ok := e.Runes[idx]; !ok {
						e.Runes[idx] = make(map[rune]bool)
					}
					e.Runes[idx][r] = true
				}
			}
			d[configuration] = &e
		}
	}

	return d
}
