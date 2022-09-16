package main

import (
	"fmt"
	"log"
	"os"

	"cryptoquip/qp"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Find matches in dictionary by word shape\n")
		fmt.Fprintf(os.Stderr, "usage: %s cleartext.dictionary word [word...]\n", os.Args[0])
		return
	}

	dict, err := qp.NewShapeDict(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	allLetters := qp.NewRunesDict(dict)

	for _, str := range os.Args[2:] {
		config := qp.StringConfiguration(str)
		fmt.Printf("%s\n%s\n\n", str, config)

		configMatches := dict[config]
		for _, m := range configMatches {
			fmt.Printf("\t%s\n", m)
		}
		fmt.Printf("%d matches\n", len(configMatches))

		if entry, ok := allLetters[config]; ok {
			fmt.Printf("Found letters for configuration %s\n", config)
			for i := 0; i < entry.Length; i++ {
				fmt.Printf("Letters at %d: ", i)
				for r, _ := range entry.Runes[i] {
					fmt.Printf("%c ", r)
				}
				fmt.Println()
			}
		} else {
			fmt.Printf("Did not find letters for configuration %s\n", config)
		}
	}
}
