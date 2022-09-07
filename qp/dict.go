package qp

import (
	"bufio"
	"log"
	"os"
)

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

		config := StringConfiguration(line)
		d[config] = append(d[config], line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("problem line %d: %v", lineCounter, err)
	}

	return d, nil
}
