package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bosley/xla/xlist"
	"github.com/bosley/xla/xvm"
)

func main() {
	// Set up slog with Debug level
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Define flags
	resourcesPath := flag.String("resources", "./resources", "Path to resources for xvm.New()")
	flag.Parse()

	// Check if a filename is provided as an argument
	if flag.NArg() < 1 {
		slog.Error("Usage: go run main.go [--resources <path>] <filename>")
		os.Exit(1)
	}

	filename := flag.Arg(0)

	// Parse the file
	nodes, err := xlist.Parse(filename)
	if err != nil {
		slog.Error("Failed to parse file", "error", err)
		os.Exit(1)
	}

	// Validate that the parsed result is not an error
	for _, node := range nodes {
		if node.Type == xlist.NodeTypeError {
			slog.Error("Error in parsed content", "error", node.Data)
			os.Exit(1)
		}
	}

	// Create a new runtime with the specified resources path
	runtime, err := xvm.New(nodes, *resourcesPath)

	if err != nil {
		slog.Error("Failed to create runtime", "error", err)
		os.Exit(1)
	}

	result := runtime.Run()

	if result.Type == xlist.NodeTypeError {
		slog.Error("Execution failure", "error", result.Data.(error))
		os.Exit(1)
	}
	fmt.Println(result.ToStringDeep())
}
