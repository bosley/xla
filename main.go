package main

import (
	"flag"
	"fmt"
	"os"
)

// main is the entry point of the program.
// It handles command-line arguments, file loading, and initiates the parsing process.
// The parsed result is then printed to the console.
// Users can modify this function to integrate the parser into their own applications or to customize output formatting.
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

	// Collect takes in any slice of ruins and expect a full top level statement,
	// meaning that comments or lists must be present, all input not matching this
	// other than white space result in an error
	result := Collect(runes)

	// Print the result
	fmt.Printf("Parsed result: %+v\n", result)

	// If you want to print the original content as well:
	fmt.Printf("Original content: %s\n", string(runes))
}
