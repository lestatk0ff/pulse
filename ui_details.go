package main

import (
	"fmt"
	"strings"
	"time"
)

const probeDebounceDelay = 140 * time.Millisecond

// probeAndShowDetails lazily runs ffprobe for the selected file and then renders details.
func (a *app) probeAndShowDetails(f *AudioFile) {
	if f == nil {
		return
	}

	requestID := a.probeRequestID.Add(1)
	a.stopProbeDebounce()

	if f.Probed {
		a.renderDetails(f)
		return
	}
	a.detailsView.SetText("[yellow]Loading…")

	// Delay external probing briefly so fast cursor navigation does not spawn one ffprobe per row.
	a.probeDebounce = time.AfterFunc(probeDebounceDelay, func() {
		format, bitrate, duration, tags := probeFile(f.Path)
		a.tv.QueueUpdateDraw(func() {
			// Drop results from an older selection/probe request.
			if a.probeRequestID.Load() != requestID {
				return
			}

			if format != "" {
				f.Format = format
			}
			f.Bitrate = bitrate
			f.Duration = duration
			f.Tags = tags
			f.Probed = true

			// Rebuild the table row so metadata updates keep the playing marker/color in sync.
			for row := 1; row <= len(a.files); row++ {
				if a.files[row-1] == f {
					a.setFileRow(row, f)
					break
				}
			}
			if a.selectedFile == f {
				a.renderDetails(f)
			}
		})
	})
}

// stopProbeDebounce cancels any delayed ffprobe launch that has not started yet.
func (a *app) stopProbeDebounce() {
	if a.probeDebounce != nil {
		a.probeDebounce.Stop()
		a.probeDebounce = nil
	}
}

// renderDetails writes the formatted metadata for f into the details pane.
func (a *app) renderDetails(f *AudioFile) {
	var sb strings.Builder

	field := func(label, value string) {
		if value != "" {
			fmt.Fprintf(&sb, "[yellow]%s:[white] %s\n", label, value)
		}
	}

	field("Format", f.Format)
	if f.Bitrate > 0 {
		field("Bitrate", fmt.Sprintf("%d kbps", f.Bitrate))
	}
	field("Duration", f.Duration)
	field("Size", fmtSize(f.Size))

	tagFields := []struct{ key, label string }{
		{"title", "Title"},
		{"artist", "Artist"},
		{"album_artist", "Album Artist"},
		{"album", "Album"},
		{"date", "Year"},
		{"track", "Track"},
		{"disc", "Disc"},
		{"genre", "Genre"},
		{"composer", "Composer"},
		{"comment", "Comment"},
	}
	if len(f.Tags) > 0 {
		sb.WriteString("\n")
		for _, t := range tagFields {
			field(t.label, f.Tags[t.key])
		}
	}

	a.detailsView.SetText(sb.String())
}

// updatePlayingBar refreshes the now-playing bar with track or station info and elapsed time.
func (a *app) updatePlayingBar() {
	if a.nowPlayingRadio != nil {
		elapsed := time.Since(a.playStart)
		elapsedStr := fmt.Sprintf("%d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		vol := fmt.Sprintf("%d%%", a.volume)
		if a.muted {
			vol = "MUTED"
		}
		radioIcon := "[green]▶ Radio:"
		if a.paused {
			radioIcon = "[yellow]⏸ Radio:"
		}
		stationPart := fmt.Sprintf("%s[white] %s", radioIcon, a.nowPlayingRadio.Name)
		if a.radioTrack != "" {
			stationPart += fmt.Sprintf("  [grey]♪ %s", a.radioTrack)
		}
		a.playingBar.SetText(fmt.Sprintf(
			"[yellow]%s[white]  %s  [gray]%s  [aqua]Vol:[white] %s  ",
			a.playerName, stationPart, elapsedStr, vol,
		))
		return
	}
	if a.nowPlaying == nil {
		a.playingBar.SetText("")
		return
	}
	track := a.nowPlaying.Name
	artist := a.nowPlaying.Tags["artist"]
	title := a.nowPlaying.Tags["title"]
	if artist != "" && title != "" {
		track = artist + " - " + title
	} else if title != "" {
		track = title
	}
	elapsed := time.Since(a.playStart)
	elapsedStr := fmt.Sprintf("%d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	vol := fmt.Sprintf("%d%%", a.volume)
	if a.muted {
		vol = "MUTED"
	}
	playIcon := "[green]▶ Playing:"
	if a.paused {
		playIcon = "[yellow]⏸ Paused:"
	}
	a.playingBar.SetText(fmt.Sprintf(
		"[yellow]%s[white]  %s[white] %s  [gray]%s / %s  [aqua]Vol:[white] %s  ",
		a.playerName, playIcon, track, elapsedStr, a.nowPlaying.Duration, vol,
	))
}
