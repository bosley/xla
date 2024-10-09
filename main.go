package main

import (
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

	// Check if a filename is provided as an argument
	if len(os.Args) < 2 {
		slog.Error("Usage: go run main.go <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]

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

	// Create a new runtime
	result := xvm.New(nodes).Run()
	if result.Type == xlist.NodeTypeError {
		slog.Error("Execution failure", "error", result.Data.(error))
		os.Exit(1)
	}
	fmt.Println(result.ToStringDeep())
}
