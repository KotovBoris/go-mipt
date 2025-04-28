//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	stats := make(map[string]uint)

	paths := os.Args[1:]
	for i := range paths {
		file, _ := os.Open(paths[i])
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			stats[line]++
		}

	}

	for key, val := range stats {
		if val <= 1 {
			continue
		}
		fmt.Printf("%d\t%s\n", val, key)
	}
}
