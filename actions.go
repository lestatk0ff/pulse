package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// runAction dispatches the action list item at idx to the appropriate handler.
// idx matches the order items were added in buildActions
// (0=convertFLACtoMP3, 1=shuffleCurrentList, 2=refresh).
func (a *app) runAction(idx int) {
	if a.radioMode {
		return
	}
	if a.selectedFile == nil {
		a.setStatus("[red]No file selected — navigate to a file in the top panel first.")
		return
	}
	switch idx {
	case 0:
		a.convertFLACtoMP3(a.selectedFile)
	case 1:
		a.shuffleCurrentList()
	case 2:
		a.refresh()
	}
}

// shuffleCurrentList randomizes the order of the currently visible list.
// If a filter is active, only filtered results are shuffled.
func (a *app) shuffleCurrentList() {
	if a.radioMode {
		return
	}
	if len(a.files) < 2 {
		a.setStatusTemporary("[grey]Need at least 2 songs to shuffle", 3*time.Second)
		return
	}

	selected := a.selectedFile
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(a.files), func(i, j int) {
		a.files[i], a.files[j] = a.files[j], a.files[i]
	})

	a.populateTable()

	if selected != nil {
		for i, f := range a.files {
			if f == selected {
				a.table.Select(i+1, 0)
				a.selectedFile = f
				a.probeAndShowDetails(f)
				break
			}
		}
	}

	a.setStatusTemporary(fmt.Sprintf("[green]Shuffled:[white] %d song(s)", len(a.files)), 3*time.Second)
}

// convertFLACtoMP3 re-encodes a FLAC file to MP3 320 kbps using the LAME encoder,
// writing <name>.mp3 next to the original. The conversion runs in a goroutine to avoid
// blocking the UI.
func (a *app) convertFLACtoMP3(f *AudioFile) {
	if strings.ToLower(filepath.Ext(f.Path)) != ".flac" {
		a.setStatus("[red]Action requires a .flac file — select a FLAC track first.")
		return
	}
	outPath := strings.TrimSuffix(f.Path, filepath.Ext(f.Path)) + ".mp3"

	a.setStatus(fmt.Sprintf("[yellow]Converting [white]%s[yellow] → MP3 VBR V0…", f.Name))

	go func() {
		out, err := flacToMP3Cmd(f.Path, outPath).CombinedOutput()
		if err != nil {
			a.setStatusAsync(fmt.Sprintf("[red]Error: %s", firstLine(string(out))))
		} else {
			a.setStatusAsync(fmt.Sprintf("[green]Done:[white] %s", filepath.Base(outPath)))
		}
	}()
}

// refresh re-scans a.dir in a goroutine and reloads the table with the new file list.
func (a *app) refresh() {
	a.setStatus("[yellow]Scanning…")
	go func() {
		files, err := scanDir(a.dir)
		// QueueUpdateDraw ensures the UI update runs on the main tview goroutine.
		a.tv.QueueUpdateDraw(func() {
			if err != nil {
				a.setStatus(fmt.Sprintf("[red]Scan error: %v", err))
				return
			}
			a.allFiles = files
			a.selectedFile = nil
			if a.filterActive {
				a.applyFilter(a.searchBar.GetText())
			} else {
				a.files = files
				a.populateTable()
			}
			a.setStatus(fmt.Sprintf("[green]Refreshed:[white] %d file(s) found", len(files)))
		})
	}()
}

// flacToMP3Cmd builds the ffmpeg command for FLAC → MP3 VBR V0 conversion.
// Shared by the TUI action and the CLI encode subcommand.
func flacToMP3Cmd(inPath, outPath string) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-i", inPath,
		"-map", "0:a",
		"-map", "0:v?",            // cover art — skipped if absent
		"-c:a", "libmp3lame",
		"-q:a", "0",               // VBR V0 ~245 kbps avg
		"-compression_level", "0", // LAME -q 0: best psychoacoustic model
		"-c:v", "copy",
		"-id3v2_version", "3",
		"-y", outPath,
	)
}
