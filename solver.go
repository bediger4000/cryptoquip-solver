package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"unicode"

	"cryptoquip/qp"
)

func main() {
	dictName := flag.String("d", "/usr/share/dict/words", "cleartext dictionary")
	puzzleName := flag.String("p", "", "puzzle file name")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	dict, err := qp.NewDict(*dictName)
	if err != nil {
		log.Fatal(err)
	}

	if *puzzleName == "" {
		log.Fatal("need a puzzle file name")
	}

	puzzlewords, enciphered, clear, err := qp.ReadPuzzle(*puzzleName, *verbose)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Hint: %c = %c\n\n", enciphered, clear)

	allLetters := qp.NewRunesDict(dict)

	possibleLetters := make(map[rune]map[rune]bool)

	for _, str := range puzzlewords {
		config := qp.StringConfiguration(string(str))
		fmt.Printf("%s\n%s\n", str, config)

		configMatches := dict[config]
		fmt.Printf("%d matches\n", len(configMatches))

		if entry, ok := allLetters[config]; ok {
			fmt.Printf("Found letters for %s, configuration %s\n", str, config)
			for i := 0; i < entry.Length; i++ {
				// all the letters found at index i in all words with this configuration
				cipherLetter := rune(str[i])
				if unicode.IsPunct(cipherLetter) {
					continue
				}
				if clearLetters, ok := possibleLetters[cipherLetter]; ok {
					if *verbose {
						printLetters(cipherLetter, "currently associated with", clearLetters)
					}
					hadN := len(clearLetters)
					// find intersection of clearLetters and entry.Runes[i]
					intersection := make(map[rune]bool)
					for newLetter, _ := range entry.Runes[i] {
						if clearLetters[newLetter] {
							intersection[newLetter] = true
						}
					}
					possibleLetters[cipherLetter] = intersection
					hasN := len(intersection)
					fmt.Printf("cipher letter %c had %d clear letters, has %d\n", cipherLetter, hadN, hasN)
					if *verbose {
						printLetters(cipherLetter, "now associated with", possibleLetters[cipherLetter])
					}
				} else {
					possibleLetters[cipherLetter] = make(map[rune]bool)
					for newLetter, _ := range entry.Runes[i] {
						possibleLetters[cipherLetter][newLetter] = true
					}
					fmt.Printf("cipher letter %c starts with %d clear letters\n", cipherLetter, len(possibleLetters[cipherLetter]))
					if *verbose {
						printLetters(cipherLetter, "initially associated with", possibleLetters[cipherLetter])
					}
				}
			}
			fmt.Println()
		} else {
			fmt.Printf("Did not find letters for %s, configuration %s\n", str, config)
		}
	}

	// Delete all but the clue from clue's cipher letter
	fmt.Printf("setting hint cipherletter %c to %c\n", enciphered, clear)
	possibleLetters[enciphered] = make(map[rune]bool)
	possibleLetters[enciphered][clear] = true

	// Delete the clue from all other cipher letters
	for cipherLetter, letters := range possibleLetters {
		if cipherLetter == enciphered {
			continue
		}
		delete(letters, clear)
	}

	// Determine if some letter(s) only appear in one cipher letter's list
	counts := make(map[rune]int)
	for _, possibles := range possibleLetters {
		for l, _ := range possibles {
			counts[l]++
		}
	}
	for l, count := range counts {
		if count == 1 {
			// there's only a single letter with value in l.
			// Rid the cipher letter it's associated with of all other clear letters
			fmt.Printf("%c only appears once\n", l)
			for cipherletter, possibles := range possibleLetters {
				if possibles[l] {
					fmt.Printf("singleton clear text letter %c associated with cipher letter %c\n", l, cipherletter)
					single := make(map[rune]bool)
					single[l] = true
					possibleLetters[cipherletter] = single
					break // don't have to look for other cipherletters
				}
			}
			// don't have to remove it from other cipherletter's clear letters
		}
	}

	for cipherLetter, possibles := range possibleLetters {
		printLetters(cipherLetter, "", possibles)
	}

	cipherLetterRegexps := make(map[rune]string)

	// compose regular expressions for cipherwords
	for _, cipherword := range puzzlewords {
		cwregexp := "^"
		for _, b := range cipherword {
			r := rune(b)
			if _, ok := cipherLetterRegexps[r]; !ok {
				cipherLetterRegexps[r] = composeRegexp(possibleLetters[r])
			}
			clregexp := cipherLetterRegexps[r]
			cwregexp += clregexp
		}
		cwregexp += "$"
		fmt.Printf("%q must match '%s'\n", cipherword, cwregexp)
	}
}

func printLetters(cipherLetter rune, format string, m map[rune]bool) {
	ln := len(m)
	fmt.Printf("cipher letter %c %s (%d):", cipherLetter, format, ln)
	var letters []rune
	for l, _ := range m {
		letters = append(letters, l)
	}
	sort.Sort(RuneSlice(letters))
	for i := range letters {
		fmt.Printf(" %c", letters[i])
	}
	fmt.Println()
}

type RuneSlice []rune

func (rs RuneSlice) Len() int           { return len(rs) }
func (rs RuneSlice) Less(i, j int) bool { return rs[i] < rs[j] }
func (rs RuneSlice) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

type lrange struct {
	begin rune
	end   rune
}

func composeRegexp(m map[rune]bool) string {
	if len(m) == 0 {
		return ""
	}
	if len(m) == 1 {
		for l, _ := range m {
			return fmt.Sprintf("%c", l)
		}
	}
	var letters []rune
	for l, _ := range m {
		letters = append(letters, l)
	}
	sort.Sort(RuneSlice(letters))

	var ranges []*lrange
	var currRange = &lrange{
		begin: letters[0],
		end:   letters[0],
	}

	for _, l := range letters[1:] {
		if l > currRange.end+1 {
			ranges = append(ranges, currRange)
			currRange = &lrange{
				begin: l,
			}
		}
		currRange.end = l
	}
	ranges = append(ranges, currRange)

	str := ""
	for i := range ranges {
		if ranges[i].begin == ranges[i].end {
			str = fmt.Sprintf("%s%c", str, ranges[i].begin)
			continue
		}
		if ranges[i].begin+1 == ranges[i].end {
			str = fmt.Sprintf("%s%c%c", str, ranges[i].begin, ranges[i].end)
			continue
		}
		str = fmt.Sprintf("%s%c-%c", str, ranges[i].begin, ranges[i].end)
	}

	return fmt.Sprintf("[%s]", str)
}
