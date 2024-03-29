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
	cycles := flag.Int("c", 8, "number of cycles to attempt")
	encodeSelf := flag.Bool("s", false, "cipher letter can encode itself, real default true")
	flag.Parse()

	*encodeSelf = !*encodeSelf
	if *encodeSelf {
		fmt.Println("Allowing cipherletters to encode themselves")
	}

	if *puzzleName == "" {
		log.Fatal("need a puzzle file name")
	}

	puzzlewords, uniquePuzzlewords, cipherLetters, hints, err := qp.ReadPuzzle(*puzzleName, *verbose)
	if err != nil {
		log.Fatal(err)
	}

	solved := &qp.Solved{
		SolvedLetters: make(map[rune]rune),
		ClearLetters:  make(map[rune]bool),
		CipherLetters: cipherLetters,
		Verbose:       *verbose,
	}
	if len(hints) > 0 {
		for cipherHint, clearHint := range hints {
			fmt.Printf("Hint: %c = %c\n\n", cipherHint, clearHint)
			solved.SetSolved(cipherHint, clearHint)
		}
	}
	solved.SetSolved('\'', '\'')
	fmt.Printf("%d  total cipher words\n", len(puzzlewords))
	fmt.Printf("%d unique cipher words\n", len(uniquePuzzlewords))
	fmt.Printf("%d  total cipher letters\n", len(solved.CipherLetters))

	totalShapeDict, err := qp.NewShapeDict(*dictName)
	if err != nil {
		log.Fatal(err)
	}
	shapeDictCharacterization(totalShapeDict, "unfiltered clear text")

	shapeDict := limitShapeDict(totalShapeDict, uniquePuzzlewords)

	// find all the dictionary words "shapes", and match up the letters with
	// those shapes.
	// The word "goober" would have the shape "011234".
	// "goober" would add 'g' to position 0 of words with shape "011234",
	// add 'o' to position 1 of words with shape "011234",
	// add 'o' to position 2 of words with shape "011234",
	// add 'b' to position 3 of words with shape "011234",
	// etc etc
	allLetters := qp.NewRunesDict(shapeDict)

	// cycle through the steps of finding clear text letters for
	// cipher text letters
	for cycle := 0; len(solved.CipherLetters) > len(solved.SolvedLetters) && cycle < *cycles; cycle++ {

		fmt.Printf("---start cycle %d---\n\n", cycle)

		shapeDictCharacterization(shapeDict, fmt.Sprintf("cycle %d", cycle))

		// map of cipher letters to correpsonding set of clear text letters
		// that get found during this cycle.
		possibleLetters := make(map[rune]map[rune]bool)

		// look through all the puzzle words and find the intersection of
		// all the sets-of-cleartext-letters for any given cipher letter
		seenWordAlready := make(map[string]bool)
		for _, str := range uniquePuzzlewords {

			// Doesn't pay off to examine the same word several times
			if seenWordAlready[string(str)] {
				continue
			}
			seenWordAlready[string(str)] = true

			config := qp.StringConfiguration(string(str))
			fmt.Printf("\ncipher word under consideration: %s\ncipher word shape %s\n", str, config)

			configMatches := shapeDict[config]
			fmt.Printf("\t%d shape matches on %q\n", len(configMatches), config)
			if len(configMatches) < 6 {
				for i := range configMatches {
					fmt.Printf("\t%s\n", configMatches[i])
				}
			}

			if entry, ok := allLetters[config]; ok {
				for i := 0; i < entry.Length; i++ {
					// all the letters found at index i in all clear text words with this configuration
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
						// find common letters in clearLetters and entry.Runes[i]
						possibleLetters[cipherLetter] = intersectSlices(entry.Runes[i], clearLetters)
						if *verbose {
							hasN := len(possibleLetters[cipherLetter])
							fmt.Printf("cipher letter %c had %d clear letters, has %d\n", cipherLetter, hadN, hasN)
							printLetters(cipherLetter, "now associated with", possibleLetters[cipherLetter])
						}
					} else {
						possibleLetters[cipherLetter] = make(map[rune]bool)
						for newLetter := range entry.Runes[i] {
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

		printSortedPossible(cycle, possibleLetters)
		if !*encodeSelf {
			// in real Cryptoquips, Cryptoquotes and Celebrity Ciphers,
			// a cipherletter isn't itself as a clearletter
			for cipherletter, matches := range possibleLetters {
				if _, ok := matches[cipherletter]; ok {
					if *verbose {
						fmt.Printf("deleting %c from matching clearletter for %c\n", cipherletter, cipherletter)
					}
					delete(matches, cipherletter)
					possibleLetters[cipherletter] = matches
				}
			}
		}

		// if any ciper letters have a set of cleartext letters of size 1,
		// mark those cipher letters as solved.
		markSingleSolvedLettes(solved, possibleLetters)

		// Compose regular expressions for each puzzle (cipher) word based
		// on the sets of cleartext letters.
		shapeMatches := cwMustMatch(solved, uniquePuzzlewords, possibleLetters)

		// recreate a "shape dictionary" based on words that match the regular
		// expressions, and exist in the current shape dictionary.
		shapeDict = shapeDictFromRegexp(solved, shapeDict, shapeMatches)
		shapeDictCharacterization(shapeDict, "new")

		// Figure out the sets of clear text letters associated with each
		// cipher letter from the newly re-created shape dictionary.
		// Solved cleartext letters don't get removed here.
		allLetters = qp.NewRunesDict(shapeDict)

		printSolvedLetters(solved)

		fmt.Println("\nSolved Puzzle:")
		printSolvedWords(puzzlewords, solved)

		fmt.Printf("---end cycle %d---\n\n", cycle)
	}
}

// shapeDictCharacterization prints out "size" of a shape dictionary,
// a map[string][]string, where the map key is a word "shape" or "configuration",
// and the key's associated value is a slice of string words that have that shape.
func shapeDictCharacterization(shapeDict map[string][]string, phrase string) {
	wordCount := 0
	for _, words := range shapeDict {
		wordCount += len(words)
	}
	fmt.Printf("%s shape dictionary has %d shapes, %d words\n", phrase, len(shapeDict), wordCount)
	if len(shapeDict) < 11 {
		for shape, matches := range shapeDict {
			fmt.Printf("\tshape %s has %d matches\n", shape, len(matches))
		}
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

func printSortedPossible(cycle int, possibleLetters map[rune]map[rune]bool) {
	var keys []rune
	for cipherLetter := range possibleLetters {
		keys = append(keys, cipherLetter)
	}
	sort.Sort(qp.RuneSlice(keys))

	fmt.Printf("After cycle %d shape comparisons:\n", cycle)

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
	for l := range m {
		letters = append(letters, l)
	}
	sort.Sort(qp.RuneSlice(letters))
	for i := range letters {
		fmt.Printf(" %c", letters[i])
	}
	fmt.Println()
}

type lrange struct {
	begin rune
	end   rune
}

// regexpForLetter makes a regular expression that matches a single
// cleartext letter from map m, which contains all of the letters that
// a cipher letter represents.
func regexpForLetter(solved *qp.Solved, cipherLetter rune, m map[rune]bool) string {
	if len(m) == 0 {
		// should this be an error? should it get logged?
		return ""
	}
	if len(m) == 1 {
		for l := range m {
			return fmt.Sprintf("%c", l)
		}
	}

	// If this cipher letter is already solved, put clear letter in as the
	// regular expression.
	if sl, ok := solved.SolvedLetters[cipherLetter]; ok {
		return fmt.Sprintf("%c", sl)
	}

	var letters []rune
	for l := range m {
		// l is potentially the solution for cipherLetter
		if _, ok := solved.ClearLetters[l]; ok {
			// clear letter l is already known a match for some other cipher letter
			continue
		}
		letters = append(letters, l)
	}

	// This is an odd thing to have to check.
	if len(letters) == 0 {
		// loop above threw out all the entries of m because each of them
		// is a known solution for some other cipher letter
		fmt.Fprintf(os.Stderr, "cipher letter %c has no possible matches\n", cipherLetter)
		fmt.Fprintf(os.Stderr, "candidate matches for %c: ", cipherLetter)
		for l := range m {
			fmt.Fprintf(os.Stderr, " %c", l)
		}
		fmt.Fprintf(os.Stderr, "\n  They're all solved\n")
		os.Exit(1)
	}
	sort.Sort(qp.RuneSlice(letters))

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

// cwMustMatch composes regular expressions that cipherwords must match
func cwMustMatch(solved *qp.Solved, puzzlewords [][]byte, possibleLetters map[rune]map[rune]bool) []*shapeMatch {

	var smatches []*shapeMatch

	cipherLetterRegexps := make(map[rune]string)

	for _, cipherword := range puzzlewords {
		cwregexp := "^"
		for _, b := range cipherword {
			r := rune(b)
			if sl, ok := solved.SolvedLetters[r]; ok {
				cipherLetterRegexps[r] = fmt.Sprintf("%c", sl)
			} else if _, ok := cipherLetterRegexps[r]; !ok {
				cipherLetterRegexps[r] = regexpForLetter(solved, r, possibleLetters[r])
			}
			clregexp := cipherLetterRegexps[r]
			cwregexp += clregexp
		}
		cwregexp += "$"
		if solved.Verbose {
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

// shapeDictFromRegexp makes a new "shape dictionary" from the previous
// cycle's shape dictionary and the regular expressions composed from
// the clear text letters from intersecting the previous cycle's
// shape dictionary entries.
func shapeDictFromRegexp(solved *qp.Solved, shapeDict map[string][]string, shapeMatches []*shapeMatch) map[string][]string {

	newShapeDict := make(map[string][]string)

	// map keyed by cipher letter, values are slices of runes
	// that match that cipher letter
	lettersFromRgxp := make(map[rune]map[rune]bool)

	if solved.Verbose {
		fmt.Printf("creating new shape dictionary with %d shape matchers\n", len(shapeMatches))
	}

	for _, sm := range shapeMatches {
		if solved.Verbose {
			fmt.Printf("\trecreating shape dictionary for %s:%s - %s\n",
				sm.cipherWord, sm.configuration, sm.pattern,
			)
		}
		wordMatched := make(map[string]bool)
		rgxp, err := regexp.Compile(sm.pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pattern %s: %v", sm.pattern, err)
			continue
		}
		if solved.Verbose {
			fmt.Printf("\t%d shape matches for %s in current shape dictionary\n",
				len(shapeDict[sm.configuration]),
				sm.configuration,
			)
		}

		rgxpMatchedShapeMatches := 0

		for _, shapeWord := range shapeDict[sm.configuration] {
			if !rgxp.MatchString(shapeWord) {
				continue
			}
			if wordMatched[shapeWord] {
				continue
			}
			rgxpMatchedShapeMatches++
			newShapeDict[sm.configuration] = append(
				newShapeDict[sm.configuration],
				shapeWord,
			)
			wordMatched[shapeWord] = true

			for idx, sl := range shapeWord {
				// sl cleartext letter could solve sm.cipherWord[idx]
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
		if solved.Verbose {
			fmt.Printf("\tpattern %s matched %d dictionary words\n", sm.pattern, rgxpMatchedShapeMatches)
			fmt.Printf("\tcipherword %q could be %d dictionary words\n", sm.cipherWord, len(wordMatched))
			if len(wordMatched) < 11 {
				for word := range wordMatched {
					fmt.Printf("\t\t%s\n", word)
				}
			}

		}
		if len(wordMatched) == 1 {
			// we can match all the letters in sm.cipherWord
			// to the clear text letters in newShapeDict[sm.configuration],
			// setting a key/value in the map solvedLetters.
			// Unless there's already a value in solvedLetters for the cipher letter,
			// and it's not the letter in sm.cipherWord[i]
			var soleMatch string
			for soleMatch = range wordMatched {
			}
			if solved.Verbose {
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
		} else if len(wordMatched) > 1 {
			// See if some letter(s) are the same in the same position of all words
			letters := make([]map[rune]bool, 0)
			for word := range wordMatched {
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
					for c = range m {
					}
					fmt.Printf("At position %d in shape matches, cipher letter %c, only 1 clear letter: %c\n", idx, sm.cipherWord[idx], c)
					solved.SetSolved(rune(sm.cipherWord[idx]), c)
				}
			}
		}
	}

	if solved.Verbose {
		for r, ltrs := range lettersFromRgxp {
			fmt.Printf("cipher letter %c clear letters from regexps: ", r)
			sortThenPrint(ltrs)
		}
	}

	for cipherLetter, clearLetters := range lettersFromRgxp {
		if len(clearLetters) == 1 {
			for clearLetter := range clearLetters {
				solved.SetSolved(cipherLetter, clearLetter)
			}
		}
	}

	return newShapeDict
}

// printSolvedLetters prints a human-comprehensible correspondence
// of cipher- to solved-letters.
func printSolvedLetters(solved *qp.Solved) {
	fmt.Printf("\nSolved letters:\n")
	for i := range solved.CipherLetters {
		fmt.Printf("%c ", solved.CipherLetters[i])
	}
	fmt.Println()
	for i := range solved.CipherLetters {
		if clear, ok := solved.SolvedLetters[solved.CipherLetters[i]]; ok {
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
func markSingleSolvedLettes(solved *qp.Solved, possibleLetters map[rune]map[rune]bool) {
	for cipherLetter, letters := range possibleLetters {
		if len(letters) == 1 {
			for singleLetter := range letters {
				solved.SetSolved(cipherLetter, singleLetter)
			}
		}
	}
}

// intersectSlices returns a set that's the intersection of
// two sets of runes.
func intersectSlices(sl1, sl2 map[rune]bool) map[rune]bool {
	intersection := make(map[rune]bool)

	for newLetter := range sl1 {
		if sl2[newLetter] {
			intersection[newLetter] = true
		}
	}

	return intersection
}

// limitShapeDict called on the shape dictionary derived from the whole clear
// text dictionary, and the list of puzzle words. Called before the first
// cycle, so it doesn't have to deal with a shape dictionary that has shapes
// not found in the cipher letters
func limitShapeDict(totalShapeDict map[string][]string, puzzlewords [][]byte) map[string][]string {

	shapeDict := make(map[string][]string)
	seenWordAlready := make(map[string]bool)

	for _, wordBytes := range puzzlewords {
		word := string(wordBytes)
		if seenWordAlready[word] {
			continue
		}
		cfg := qp.StringConfiguration(word)
		shapeDict[cfg] = totalShapeDict[cfg]
	}

	return shapeDict
}
