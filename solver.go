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

type Solved struct {
	cipherLetters []rune        // alphabetized slice of cipherletters
	solvedLetters map[rune]rune // cipherletter key to clear text letter value
	clearLetters  map[rune]bool // all the clear letters so far
}

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

	var solved Solved

	// Keep track of solved letters in a map.
	// Keys are cipher letters, values are the corresponding cleartext letters
	solved.solvedLetters = make(map[rune]rune)
	solved.clearLetters = make(map[rune]bool)
	solved.solvedLetters[enciphered] = clear
	solved.cipherLetters = sortOutCiperLetters(puzzlewords)
	fmt.Printf("%d total cipher letters\n", len(solved.cipherLetters))

	// find all the dictionary words "shapes", and match up the letters with
	// those shapes.
	// The word "goober" would have the shape "011234".
	// "goober" would add
	allLetters := qp.NewRunesDict(shapeDict)

	for cycle := 0; cycle < 4; cycle++ {

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
					if sl, ok := solved.solvedLetters[cipherLetter]; ok {
						// This cipher letter has a clear text letter
						possibleLetters[cipherLetter] = make(map[rune]bool)
						possibleLetters[cipherLetter][sl] = true
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

		printSortedPossible(possibleLetters)

		shapeMatches := cwMustMatch(solved.solvedLetters, puzzlewords, possibleLetters, *verbose)
		shapeDict = weedShapeDict(solved.solvedLetters, shapeDict, shapeMatches, *verbose)
		allLetters = qp.NewRunesDict(shapeDict)

		// remove solved letters from allLetters
		removeLetters(&solved, allLetters)

		printSolvedLetters(solved.cipherLetters, solved.solvedLetters)

		// TODO - print puzzle with solution so far

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
func cwMustMatch(solvedLetters map[rune]rune, puzzlewords [][]byte, possibleLetters map[rune]map[rune]bool, verbose bool) []*shapeMatch {

	var smatches []*shapeMatch

	cipherLetterRegexps := make(map[rune]string)

	for _, cipherword := range puzzlewords {
		cwregexp := "^"
		for _, b := range cipherword {
			r := rune(b)
			if sl, ok := solvedLetters[r]; ok {
				cipherLetterRegexps[r] = fmt.Sprintf("%c", sl)
			} else if _, ok := cipherLetterRegexps[r]; !ok {
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

func weedShapeDict(solvedLetters map[rune]rune, shapeDict map[string][]string, shapeMatches []*shapeMatch, verbose bool) map[string][]string {

	newShapeDict := make(map[string][]string)

	for _, sm := range shapeMatches {
		rgxp, err := regexp.Compile(sm.pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pattern %s: %v", sm.pattern, err)
			continue
		}

		shapeMatches := make(map[string]bool)
		for _, shapeWord := range shapeDict[sm.configuration] {
			if rgxp.MatchString(shapeWord) {
				newShapeDict[sm.configuration] = append(
					newShapeDict[sm.configuration],
					shapeWord,
				)
				shapeMatches[shapeWord] = true
			}
		}
		if verbose {
			fmt.Printf("cipherword %q could be %d dictionary words\n", sm.cipherWord, len(shapeMatches))
			if len(shapeMatches) < 6 {
				for word, _ := range shapeMatches {
					fmt.Printf("\t%s\n", word)
				}
			}
			// TODO see if some letter(s) are the same in the same position of all words

		}
		if len(shapeMatches) == 1 {
			// we can match all the letters in sm.cipherWord
			// to the clear text letters in newShapeDict[sm.configuration],
			// setting a key/value in the map solvedLetters.
			// Unless there's already a value in solvedLetters for the cipher letter,
			// and it's not the letter in sm.cipherWord[i]
			var soleMatch string
			for soleMatch, _ = range shapeMatches {
			}
			if verbose {
				fmt.Printf("single match of %q in word shapes dictionary %q\n",
					sm.cipherWord,
					soleMatch,
				)
			}
			soleMatchRunes := []rune(soleMatch)
			for idx, cl := range sm.cipherWord {
				sl2 := soleMatchRunes[idx]
				if sl1, ok := solvedLetters[cl]; ok {
					// sl2 and sl1 should be identical, otherwise there's a problem
					if sl1 != sl2 {
						fmt.Printf("PROBLEM: %c != %c at %d in %q and %q\n",
							sl1, sl2,
							idx,
							soleMatch, sm.cipherWord,
						)
					}
				} else {
					if verbose {
						fmt.Printf("\tcipher letter %c solved as %c\n", cl, sl2)
					}
					solvedLetters[cl] = sl2
				}
			}
		}
	}
	return newShapeDict
}

// sortOutCiperLetters creates a sorted array of all
// the runes in the puzzle.
func sortOutCiperLetters(puzzlewords [][]byte) []rune {
	m := make(map[rune]bool)
	for _, word := range puzzlewords {
		for _, r := range word {
			m[rune(r)] = true
		}
	}
	cipherLetters := make([]rune, 0, len(m))
	for r, _ := range m {
		cipherLetters = append(cipherLetters, r)
	}
	sort.Sort(RuneSlice(cipherLetters))
	return cipherLetters
}

func printSolvedLetters(cipherLetters []rune, mrr map[rune]rune) {
	fmt.Printf("\nSolved letters:\n")

	for i := range cipherLetters {
		fmt.Printf("%c ", cipherLetters[i])
	}
	fmt.Println()
	for i := range cipherLetters {
		if clear, ok := mrr[cipherLetters[i]]; ok {
			fmt.Printf("%c ", clear)
		} else {
			fmt.Printf("? ")
		}
	}
	fmt.Println()
}

// removeLetters takes out solved letters so that they don't clutter
// the possible letters
func removeLetters(solved *Solved, allLetters map[string]*qp.Entry) {
	/*
		type Entry struct {
		    Length int
		    Runes  map[int]map[rune]bool
		    // Runes[0] are all the dictionary words' first letters.
		    // the first letters are keys in the map-that-is-the-value
		    // Runes[1] are all the dictionary words with this shape 2nd letters,
		    // Runes[2] are the 3rd letters from dictionary words with this shape, etc
		}
	*/
}
