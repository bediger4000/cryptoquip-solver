package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"unicode"

	"cryptoquip/qp"
)

func main() {
	dictName := flag.String("d", "/usr/share/dict/words", "cleartext dictionary")
	puzzleName := flag.String("p", "", "puzzle file name")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	shapeDict, err := qp.NewShapeDict(*dictName)
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

	allLetters := qp.NewRunesDict(shapeDict)

	for cycle := 0; cycle < 3; cycle++ {

		fmt.Printf("---start cycle %d---\n\n", cycle)

		possibleLetters := make(map[rune]map[rune]bool)

		for _, str := range puzzlewords {
			config := qp.StringConfiguration(string(str))
			fmt.Printf("%s\n%s\n", str, config)

			configMatches := shapeDict[config]
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
						if *verbose {
							hasN := len(intersection)
							fmt.Printf("cipher letter %c had %d clear letters, has %d\n", cipherLetter, hadN, hasN)
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

		stillDeleting := true
		for stillDeleting {
			deletedSingletonCount := 0
			for cipherLetter, possibles := range possibleLetters {
				if len(possibles) == 1 {
					// use range to extract the single key from map possibles
					for clearletter, _ := range possibles {
						if *verbose {
							fmt.Printf("cipher letter %c has only 1 clear letter: %c\n", cipherLetter, clearletter)
						}
						// Remove clearletter's value from all the other possibleLetters
						for cl, poss := range possibleLetters {
							if cl == cipherLetter {
								continue
							}
							if *verbose {
								fmt.Printf("deleted clear letter %c from cipher letter %c possibles\n", clearletter, cl)
							}
							delete(poss, clearletter)
							deletedSingletonCount++
						}
					}
				}
				if deletedSingletonCount < 1 {
					stillDeleting = false
				}
			}
		}

		printSortedPossible(possibleLetters)

		shapeMatches := cwMustMatch(puzzlewords, possibleLetters, *verbose)
		shapeDict = weedShapeDict(shapeDict, shapeMatches, *verbose)
		allLetters = qp.NewRunesDict(shapeDict)

		fmt.Printf("---end cycle %d---\n\n", cycle)
	}
}

func printSortedPossible(possibleLetters map[rune]map[rune]bool) {
	var keys []rune
	for cipherLetter, _ := range possibleLetters {
		keys = append(keys, cipherLetter)
	}
	sort.Sort(RuneSlice(keys))

	for i := range keys {
		printLetters(keys[i], "", possibleLetters[keys[i]])
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

type shapeMatch struct {
	cipherWord    string
	configuration string
	pattern       string
	rgxp          *regexp.Regexp
}

// compose regular expressions that cipherwords must match
func cwMustMatch(puzzlewords [][]byte, possibleLetters map[rune]map[rune]bool, verbose bool) []*shapeMatch {

	var smatches []*shapeMatch

	cipherLetterRegexps := make(map[rune]string)

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
		if verbose {
			fmt.Printf("%q must match '%s'\n", cipherword, cwregexp)
		}
		str := string(cipherword)
		smatches = append(smatches,
			&shapeMatch{
				cipherWord:    str,
				configuration: qp.StringConfiguration(str),
				pattern:       cwregexp,
			},
		)
	}
	return smatches
}

func weedShapeDict(shapeDict map[string][]string, shapeMatches []*shapeMatch, verbose bool) map[string][]string {
	newShapeDict := make(map[string][]string)
	for _, sm := range shapeMatches {
		rgxp, err := regexp.Compile(sm.pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pattern %s: %v", sm.pattern, err)
			continue
		}
		countMatches := 0
		for _, shapeWord := range shapeDict[sm.configuration] {
			if rgxp.MatchString(shapeWord) {
				newShapeDict[sm.configuration] = append(newShapeDict[sm.configuration], shapeWord)
				countMatches++
			}
		}
		if verbose {
			fmt.Printf("cipherword %q could be %d dictionary words\n", sm.cipherWord, countMatches)
			if countMatches < 6 {
				for _, word := range newShapeDict[sm.configuration] {
					fmt.Printf("\t%s\n", word)
				}

			}
		}
	}
	return newShapeDict
}
