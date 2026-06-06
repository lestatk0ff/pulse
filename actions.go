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
// (0=convertBitrate, 1=convertOGG, 2=shuffleCurrentList, 3=refresh).
func (a *app) runAction(idx int) {
	if a.selectedFile == nil {
		a.setStatus("[red]No file selected — navigate to a file in the top panel first.")
		return
	}
	switch idx {
	case 0:
		a.convertBitrate(a.selectedFile)
	case 1:
		a.convertOGG(a.selectedFile)
	case 2:
		a.shuffleCurrentList()
	case 3:
		a.refresh()
	}
}

// shuffleCurrentList randomizes the order of the currently visible list.
// If a filter is active, only filtered results are shuffled.
func (a *app) shuffleCurrentList() {
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

// convertBitrate re-encodes f to 192 kbps using ffmpeg, writing <name>_192k.<ext>
// next to the original file. The conversion runs in a goroutine to avoid blocking the UI.
func (a *app) convertBitrate(f *AudioFile) {
	ext := filepath.Ext(f.Path)
	outPath := strings.TrimSuffix(f.Path, ext) + "_192k" + ext

	a.setStatus(fmt.Sprintf("[yellow]Converting [white]%s[yellow] → 192 kbps…", f.Name))

	go func() {
		cmd := exec.Command("ffmpeg", "-i", f.Path, "-b:a", "192k", "-y", outPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			a.setStatusAsync(fmt.Sprintf("[red]Error: %s", firstLine(string(out))))
		} else {
			a.setStatusAsync(fmt.Sprintf("[green]Done:[white] %s", filepath.Base(outPath)))
		}
	}()
}

// convertOGG re-encodes f to OGG/Vorbis (quality 4) using ffmpeg.
// If the source is already .ogg the output is named <name>_converted.ogg to avoid
// overwriting the original. Runs in a goroutine to avoid blocking the UI.
func (a *app) convertOGG(f *AudioFile) {
	base := strings.TrimSuffix(f.Path, filepath.Ext(f.Path))
	outPath := base + ".ogg"
	if outPath == f.Path {
		outPath = base + "_converted.ogg"
	}

	a.setStatus(fmt.Sprintf("[yellow]Converting [white]%s[yellow] → OGG…", f.Name))

	go func() {
		cmd := exec.Command("ffmpeg", "-i", f.Path, "-c:a", "libvorbis", "-q:a", "4", "-y", outPath)
		out, err := cmd.CombinedOutput()
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
