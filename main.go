package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/bosley/xla/xlist"
	"github.com/bosley/xla/xrt"
)

// CustomHandler is a custom slog.Handler that colorizes the log level
type CustomHandler struct {
	slog.Handler
	w io.Writer
}

// Handle implements the slog.Handler interface
func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	var coloredLevel string
	switch r.Level {
	case slog.LevelDebug:
		coloredLevel = fmt.Sprintf("\x1b[36m%s\x1b[0m", level) // Cyan
	case slog.LevelInfo:
		coloredLevel = fmt.Sprintf("\x1b[32m%s\x1b[0m", level) // Green
	case slog.LevelWarn:
		coloredLevel = fmt.Sprintf("\x1b[33m%s\x1b[0m", level) // Yellow
	case slog.LevelError:
		coloredLevel = fmt.Sprintf("\x1b[31m%s\x1b[0m", level) // Red
	default:
		coloredLevel = level
	}

	msg := fmt.Sprintf("time=%s level=%s msg:\t%s", r.Time.Format("2006-01-02T15:04:05.000Z"), coloredLevel, r.Message)

	if r.NumAttrs() > 0 {
		attrs := make([]string, 0, r.NumAttrs())
		r.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
			return true
		})
		msg += " " + strings.Join(attrs, " ")
	}

	_, err := fmt.Fprintln(h.w, msg)
	return err
}

// setupLogger initializes slog with Debug level and custom formatting
func setupLogger() {
	handler := &CustomHandler{
		Handler: slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}),
		w:       os.Stderr,
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// main is the entry point of the program.
// It handles command-line arguments, file loading, and initiates the parsing process.
// The parsed result is then printed to the console using the Element's String() method.
// Finally, it creates a runtime with the collapsed list and runs it.
func main() {
	setupLogger()
	slog.Debug("Starting application")

	flag.Parse()

	if flag.NArg() != 1 {
		slog.Error("Exactly one file path argument is required")
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

	// Create a runtime with the collapsed list
	runtime := xrt.NewRuntime(collapsed)

	// Run the runtime and get the result
	runtimeResult := runtime.Run()

	// Check if the runtime result is an error
	if runtimeResult.IsError() {
		fmt.Printf("Runtime Error: %s\n", runtimeResult.Data)
		os.Exit(1)
	}

	// Print the runtime result
	fmt.Println("Runtime Result:")
	fmt.Println(runtimeResult.String())
}
