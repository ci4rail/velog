package main

import (
	"fmt"

	"github.com/ci4rail/velog/pkg/csvlogger"
)

func main() {
	w := csvlogger.NewWriter("/mnt", "test")

	n := 0
	for {
		err := w.Write([]string{"a", "b", "c"})
		if err != nil {
			fmt.Printf("Error writing: %v\n", err)
		}
		if n%100000 == 0 {
			fmt.Printf("Wrote %d lines\n", n)
		}
		n++
	}
}
