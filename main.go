package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bosley/xla/xlist"
)

// main is the entry point of the program.
// It handles command-line arguments, file loading, and initiates the parsing process.
// The parsed result is then printed to the console using the Element's String() method.
func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Error: Exactly one file path argument is required")
		os.Exit(1)
	}

	filePath := flag.Arg(0)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File '%s' does not exist\n", filePath)
		os.Exit(1)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		os.Exit(1)
	}

	runes := []rune(string(content))

	result := xlist.Collect(runes)

	collapsed := xlist.Collapse(result)

	if collapsed.IsError() {
		fmt.Printf("Error: %s\n", collapsed.Data)
		os.Exit(1)
	}

	// Use the new String() method to print the collapsed element
	fmt.Println(collapsed.String())
}
