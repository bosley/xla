package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]
	nodes, err := Parse(filename)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Parsed nodes:")
	for i, node := range nodes {
		fmt.Printf("%d: %s\n", i, node.ToStringDeep())
	}
}
