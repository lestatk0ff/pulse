# Pulse

Terminal UI app (Go + tview) to scan a directory of audio files, inspect metadata, play tracks, filter the list, and run quick conversion actions.

## Main Features

- Recursive scan of a directory for audio files (`.mp3`, `.ogg`, `.flac`, `.wav`, `.aac`, `.m4a`)
- Table view with file name, size, format, and bitrate
- Metadata/details pane with tags (title, artist, album, year, genre, etc.)
- Playback controls from the UI (play selected track, stop, auto-advance)
- Regexp filter (`Ctrl+F`) with live results
- Conversion actions:
  - Convert selected file to `192 kbps`
  - Convert selected file to `OGG/Vorbis`
- Refresh action to re-scan files without restarting the app

## MVP Requirements

Minimum tools needed for the app to work end-to-end:

- Go `1.21+`
- FFmpeg tools (must include both `ffmpeg` and `ffprobe` in `PATH`)
- At least one supported CLI audio player in `PATH`:
  - `mpv` (preferred)
  - `ffplay`
  - `mplayer`

## Installation Instructions (Official Pages)

- Go install docs: https://go.dev/doc/install
- FFmpeg downloads/install docs: https://ffmpeg.org/download.html
- mpv install docs: https://mpv.io/installation/
- MPlayer project page: https://mplayerhq.hu/

## Build And Run

From this folder:

```bash
go mod tidy
go run . /path/to/music
```

Or build a binary:

```bash
go build -o pulse .
./pulse /path/to/music
```

## Code Layout

- `main.go`: CLI argument validation and app bootstrap.
- `types.go`: shared state structs (`app`, `AudioFile`, ffprobe JSON types).
- `ui_layout.go`: main layout and keyboard wiring.
- `ui_overlay.go`: configuration and theme overlay flow.
- `ui_details.go`: lazy probe rendering and details/now-playing view updates.
- `ui_filter_status.go`: filter lifecycle and status message helpers.
- `themes.go`: built-in frame/table-header theme definitions.
- `scanner.go`: recursive audio file discovery.
- `media_probe.go`: `ffprobe` metadata extraction for a selected file.
- `player.go`: CLI player integration (`mpv`/`ffplay`/`mplayer`).
- `actions.go`: conversion/shuffle/refresh action handlers.
- `helpers.go`: small formatting/string helpers.

## Controls

- `Tab` - switch panel
- `Up/Down` - navigate files
- `Enter` - play selected file (in main table) or run selected menu item (in Configuration/Themes)
- `C` - open Configuration overlay
- `Z` - shuffle currently visible list
- `S` - stop playback
- `M` - mute/unmute playback (restores previous volume)
- `+` - increase playback volume
- `-` - decrease playback volume
- `Ctrl+F` - open filter input
- `R` - refresh file list
- `Esc` - close current overlay (if open), clear filter (if active), or quit app

## Notes

- Metadata is loaded lazily with `ffprobe` when a file is selected.
- Conversions create new files in the same directory as the source file.
