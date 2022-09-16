package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Need filename on command line\n")
		return
	}

	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	txp := makeTranspose()
	clearLetters := make(map[rune]bool)

	for _, b := range buf {
		c := rune(b)
		if 'a' <= c && c <= 'z' {
			clearLetters[c] = true
			c = txp[c]
		} else if 'A' <= c && c <= 'Z' {
			clearLetters[c] = true
			c = txp[unicode.ToLower(c)]
		}
		fmt.Printf("%c", c)
	}

	clears := make([]int, len(clearLetters))
	cnt := 0
	for cl, _ := range clearLetters {
		clears[cnt] = int(cl)
		cnt++
	}

	sort.Ints(clears)

	clearText := "# clear  "
	cipherText := "# cipher "
	for i := range clears {
		clear := rune(clears[i])
		cipher := txp[clear]
		fmt.Printf("#%c=%c\n", cipher, clear)
		clearText = fmt.Sprintf("%s %c", clearText, clear)
		cipherText = fmt.Sprintf("%s %c", cipherText, cipher)
	}
	fmt.Println(clearText)
	fmt.Println(cipherText)
}

func makeTranspose() map[rune]rune {
	rand.Seed(time.Now().UnixNano() + int64(os.Getpid()))
	assoc := make(map[rune]rune)
	for r := 'a'; r <= 'z'; r++ {
	OUT:
		for {
			offset := rand.Intn(int('z'-'a') + 1)
			x := rune('a' + offset)
			if _, ok := assoc[x]; !ok {
				assoc[x] = r
				break OUT
			}
		}
	}
	return assoc
}
