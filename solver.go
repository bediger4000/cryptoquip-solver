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
	cycles := flag.Int("c", 4, "number of cycles to attempt")
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

	var solved qp.Solved

	// Keep track of solved letters in a map.
	// Keys are cipher letters, values are the corresponding cleartext letters
	solved.SolvedLetters = make(map[rune]rune)
	solved.ClearLetters = make(map[rune]bool)
	solved.SolvedLetters[enciphered] = clear
	solved.CipherLetters = sortOutCiperLetters(puzzlewords)
	solved.Verbose = *verbose
	fmt.Printf("%d total cipher letters\n", len(solved.CipherLetters))

	// find all the dictionary words "shapes", and match up the letters with
	// those shapes.
	// The word "goober" would have the shape "011234".
	// "goober" would add
	allLetters := qp.NewRunesDict(shapeDict)

	for cycle := 0; len(solved.CipherLetters) > len(solved.SolvedLetters) && cycle < *cycles; cycle++ {

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
					if sl, ok := solved.SolvedLetters[cipherLetter]; ok {
						// This cipher letter has a clear text letter
						if *verbose {
							fmt.Printf("cipher letter %c already has a solved clear text letter %c\n", cipherLetter, sl)
						}
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
						// leave already solved cipher-letter-solutions out of possibleLetters
						for cl, sl := range solved.SolvedLetters {
							if cl == cipherLetter {
								continue
							}
							delete(possibleLetters[cipherLetter], sl)
						}
						printLetters(cipherLetter, "begins cycle with", possibleLetters[cipherLetter])
					}
				}
				fmt.Println()
			} else {
				fmt.Printf("Did not find letters for %s, configuration %s\n", str, config)
			}
		}

		printSortedPossible(possibleLetters)
		markSingleSolvedLettes(&solved, possibleLetters, *verbose)

		shapeMatches := cwMustMatch(&solved, puzzlewords, possibleLetters, *verbose)
		shapeDict = weedShapeDict(&solved, shapeDict, shapeMatches, *verbose)
		allLetters = qp.NewRunesDict(shapeDict)

		printSolvedLetters(solved.CipherLetters, solved.SolvedLetters)

		fmt.Println("\nSolved Puzzle:")
		printSolvedWords(puzzlewords, &solved)

		fmt.Printf("---end cycle %d---\n\n", cycle)
	}
}

func printSolvedWords(puzzlewords [][]byte, solved *qp.Solved) {
	lineLength := 0
	cipherLine := ""
	clearLine := ""
	spacer := ""
	for _, word := range puzzlewords {
		cipherLine = fmt.Sprintf("%s%s%s", cipherLine, spacer, string(word))

		clearWord := ""
		for _, b := range word {
			x := '?'
			if c, ok := solved.SolvedLetters[rune(b)]; ok {
				x = c
			}
			clearWord = fmt.Sprintf("%s%c", clearWord, x)
		}
		clearLine = fmt.Sprintf("%s%s%s", clearLine, spacer, clearWord)

		spacer = " "
		lineLength = len(cipherLine)
		if lineLength > 72 {
			fmt.Println(cipherLine)
			fmt.Println(clearLine)
			fmt.Println()
			cipherLine = ""
			clearLine = ""
			spacer = ""
		}
	}
	lineLength = len(cipherLine)
	if lineLength > 0 {
		fmt.Println(cipherLine)
		fmt.Println(clearLine)
		fmt.Println()
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
	sortThenPrint(m)
}

func sortThenPrint(m map[rune]bool) {

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

// composeRegexpForLetter makes a regular expression that matches a single
// cleartext letter from map m, which contains all of the letters that
// a cipher letter represents.
func composeRegexpForLetter(solved *qp.Solved, cipherLetter rune, m map[rune]bool) string {
	if len(m) == 0 {
		return ""
	}
	if len(m) == 1 {
		for l, _ := range m {
			return fmt.Sprintf("%c", l)
		}
	}

	// If this cipher letter is already solved, put clear letter in as the
	// regular expression.
	if sl, ok := solved.SolvedLetters[cipherLetter]; ok {
		return fmt.Sprintf("%c", sl)
	}

	var letters []rune
	for l, _ := range m {
		// l is potentially the solution for cipherLetter
		if _, ok := solved.ClearLetters[l]; ok {
			// clear letter l is already known a match for some other cipher letter
			continue
		}
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
}

// compose regular expressions that cipherwords must match
func cwMustMatch(solved *qp.Solved, puzzlewords [][]byte, possibleLetters map[rune]map[rune]bool, verbose bool) []*shapeMatch {

	var smatches []*shapeMatch

	cipherLetterRegexps := make(map[rune]string)

	for _, cipherword := range puzzlewords {
		cwregexp := "^"
		for _, b := range cipherword {
			r := rune(b)
			if sl, ok := solved.SolvedLetters[r]; ok {
				cipherLetterRegexps[r] = fmt.Sprintf("%c", sl)
			} else if _, ok := cipherLetterRegexps[r]; !ok {
				cipherLetterRegexps[r] = composeRegexpForLetter(solved, r, possibleLetters[r])
			}
			clregexp := cipherLetterRegexps[r]
			cwregexp += clregexp
		}
		cwregexp += "$"
		if verbose {
			fmt.Printf("cipher word %q must match regexp '%s'\n", cipherword, cwregexp)
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

func weedShapeDict(solved *qp.Solved, shapeDict map[string][]string, shapeMatches []*shapeMatch, verbose bool) map[string][]string {

	newShapeDict := make(map[string][]string)

	// map keyed by cipher letter, values are slices of runes
	// that match that cipher letter
	lettersFromRgxp := make(map[rune]map[rune]bool)

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

				for idx, sl := range shapeWord {
					// sl cleartext letter could solve sm.cipherWord[idx]
					// make(map[rune]map[rune]bool)
					if ltrs, ok := lettersFromRgxp[rune(sm.cipherWord[idx])]; ok {
						// seen this cipher letter before
						ltrs[sl] = true
					} else {
						ltrs = make(map[rune]bool)
						ltrs[sl] = true
						lettersFromRgxp[rune(sm.cipherWord[idx])] = ltrs
					}
				}
			}
		}
		if verbose {
			fmt.Printf("cipherword %q could be %d dictionary words\n", sm.cipherWord, len(shapeMatches))
			if len(shapeMatches) < 11 {
				for word, _ := range shapeMatches {
					fmt.Printf("\t%s\n", word)
				}
			}

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
				if sl1, ok := solved.SolvedLetters[cl]; ok {
					// sl2 and sl1 should be identical, otherwise there's a problem
					if sl1 != sl2 {
						fmt.Printf("PROBLEM: %c != %c at position %d in %q and %q\n",
							sl1, sl2,
							idx,
							soleMatch, sm.cipherWord,
						)
					}
				} else {
					solved.SetSolved(cl, sl2)
				}
			}
		} else if len(shapeMatches) > 1 {
			// See if some letter(s) are the same in the same position of all words
			letters := make([]map[rune]bool, 0)
			for word, _ := range shapeMatches {
				for idx, r := range word {
					if idx >= len(letters) {
						letters = append(letters, make(map[rune]bool))
					}
					letters[idx][r] = true
				}
			}
			for idx, m := range letters {
				if len(m) == 1 {
					// There is only one cleartext letter at position idx
					// in all of the matching-shape-words.
					var c rune
					for c, _ = range m {
					}
					fmt.Printf("At position %d in shape matches, cipher letter %c, only 1 clear letter: %c\n", idx, sm.cipherWord[idx], c)
					solved.SetSolved(rune(sm.cipherWord[idx]), c)
				}
			}
		}
	}

	if verbose {
		for r, ltrs := range lettersFromRgxp {
			fmt.Printf("cipher letter %c clear letters from regexps: ", r)
			sortThenPrint(ltrs)
		}
		wordCount := 0
		for _, words := range shapeDict {
			wordCount += len(words)
		}
		fmt.Printf("old shape dictionary has %d shapes, %d words\n", len(shapeDict), wordCount)
		wordCount = 0
		for _, words := range newShapeDict {
			wordCount += len(words)
		}
		fmt.Printf("new shape dictionary has %d shapes, %d words\n", len(newShapeDict), wordCount)
	}

	for cipherLetter, clearLetters := range lettersFromRgxp {
		if len(clearLetters) == 1 {
			for clearLetter, _ := range clearLetters {
				solved.SetSolved(cipherLetter, clearLetter)
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

// markSingleSolvedLettes trys to mark as solved any cipher letters that
// have a single possible letter left. Var possibleLetters contains the
// clear text letters left after intersecting the possible letters from
// the shape-keyed dictionary.
func markSingleSolvedLettes(solved *qp.Solved, possibleLetters map[rune]map[rune]bool, verbose bool) {
	for cipherLetter, letters := range possibleLetters {
		if len(letters) == 1 {
			for singleLetter, _ := range letters {
				solved.SetSolved(cipherLetter, singleLetter)
			}
		}
	}
}
