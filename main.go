package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const usage = `Pulse — terminal music player

Usage:
  pulse [DIR]           Launch the TUI player.
                        DIR is optional; if omitted the player starts without
                        a file list and you can navigate later.

  pulse encode -d DIR   Batch-convert every FLAC file found recursively in DIR
                        to MP3 (VBR V0, ~245 kbps, LAME encoder).
                        Each output file is written next to its source:
                          /music/album/track.flac → /music/album/track.mp3
                        Existing .mp3 files are overwritten without prompting.

  pulse -h, --help      Show this message.

Runtime dependencies (must be in PATH): mpv, ffmpeg, ffprobe
`

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "encode":
			runEncode(os.Args[2:])
			return
		case "-h", "--help":
			fmt.Print(usage)
			return
		}
	}

	var dir string
	var files []*AudioFile

	if len(os.Args) >= 2 {
		dir = os.Args[1]
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "error: %q is not a valid directory\n", dir)
			fmt.Fprint(os.Stderr, usage)
			os.Exit(1)
		}

		fmt.Printf("Scanning %s …\n", dir)
		files, err = scanDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d file(s). Launching UI…\n", len(files))
	}

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
