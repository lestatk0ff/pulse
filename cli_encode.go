package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runEncode(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	dirFlag := fs.String("d", "", "directory to scan for FLAC files")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pulse encode -d DIR")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Recursively converts all FLAC files in DIR to MP3.")
		fmt.Fprintln(os.Stderr, "  Encoder:  LAME VBR V0 (~245 kbps avg, best quality/size ratio)")
		fmt.Fprintln(os.Stderr, "  Output:   <name>.mp3 written next to each source .flac file")
		fmt.Fprintln(os.Stderr, "  Metadata: ID3v2.3 tags + embedded cover art preserved")
		fmt.Fprintln(os.Stderr, "  Existing .mp3 files are overwritten without prompting.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}
	fs.Parse(args) //nolint:errcheck // ExitOnError handles it

	if *dirFlag == "" {
		fs.Usage()
		os.Exit(1)
	}

	info, err := os.Stat(*dirFlag)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %q is not a valid directory\n", *dirFlag)
		os.Exit(1)
	}

	files, err := scanDir(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	var flacs []*AudioFile
	for _, f := range files {
		if strings.ToLower(filepath.Ext(f.Path)) == ".flac" {
			flacs = append(flacs, f)
		}
	}

	if len(flacs) == 0 {
		fmt.Println("No FLAC files found.")
		return
	}
	fmt.Printf("Found %d FLAC file(s) in %s\n\n", len(flacs), *dirFlag)

	errCount := 0
	for i, f := range flacs {
		outPath := strings.TrimSuffix(f.Path, filepath.Ext(f.Path)) + ".mp3"
		fmt.Printf("[%d/%d] %s\n", i+1, len(flacs), f.RelPath)

		out, err := flacToMP3Cmd(f.Path, outPath).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "        error: %s\n", firstLine(string(out)))
			errCount++
		}
	}

	fmt.Printf("\nDone: %d converted", len(flacs)-errCount)
	if errCount > 0 {
		fmt.Printf(", %d failed", errCount)
	}
	fmt.Println()
}
