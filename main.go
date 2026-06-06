package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Require exactly one argument: the directory to scan.
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: mp3lister <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %q is not a valid directory\n", dir)
		os.Exit(1)
	}

	fmt.Printf("Scanning %s …\n", dir)
	// Walk the directory tree and collect all recognised audio files.
	files, err := scanDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("No audio files found.")
		os.Exit(0)
	}
	fmt.Printf("Found %d file(s). Launching UI…\n", len(files))

	// Build the TUI and hand over control to the tview event loop.
	app := newTUIApp(dir, files)

	// Setup signal handler for clean shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		app.stopPlayback()
		app.tv.Stop()
	}()

	if err := app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}
