package qp

import "fmt"

type Solved struct {
	CipherLetters []rune        // alphabetized slice of cipherletters
	SolvedLetters map[rune]rune // cipherletter key to clear text letter value
	ClearLetters  map[rune]bool // all the clear letters so far
	Verbose       bool
}

func (s *Solved) SetSolved(cipherLetter, clearLetter rune) {
	if prevClear, ok := s.SolvedLetters[cipherLetter]; ok {
		if clearLetter == prevClear {
			// Already had this as a solved letter pair
			return
		}
		fmt.Printf("PROBLEM: setting cipher letter %c to clear letter %c, already had a clear letter %c\n",
			cipherLetter, clearLetter, prevClear,
		)
		return
	}
	s.SolvedLetters[cipherLetter] = clearLetter
	s.ClearLetters[clearLetter] = true
	if s.Verbose {
		fmt.Printf("\tcipher letter %c solved as %c\n", cipherLetter, clearLetter)
	}
}
