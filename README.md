# Pulse

Terminal UI app (Go + tview) to scan a directory of audio files, inspect metadata, play tracks, filter the list, run quick conversion actions, and stream online radio.

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
- Online radio browser with built-in stations and custom station support

## MVP Requirements

Minimum tools needed for the app to work end-to-end:

- Go `1.21+`
- FFmpeg tools (must include both `ffmpeg` and `ffprobe` in `PATH`)
- `mpv` in `PATH`

## Installation Instructions (Official Pages)

- Go install docs: https://go.dev/doc/install
- FFmpeg downloads/install docs: https://ffmpeg.org/download.html
- mpv install docs: https://mpv.io/installation/

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

## Online Radio

Press `R` from the file browser to switch to the **Radio** mode. The table is replaced with a list of stations showing name, genre, country, and bitrate.

### Built-in stations

17 stations are included out of the box across multiple genres (Ambient, Lounge, Folk, Metal, Soul, Electronic, Pop, Talk/News). No configuration is needed — just press `Enter` on any station to start streaming.

### Playing a station

Use `↑`/`↓` to navigate the station list and press `Enter` to connect. The now-playing bar at the bottom shows the player name, station name, and elapsed listening time. Volume (`+`/`-`), mute (`M`), and stop (`S`) all work the same as in file mode.

Press `R` again to return to the file browser. Any stream continues playing after switching back.

### Adding a custom station

Press `A` while in radio mode to open the **Add Radio Station** form:

| Field   | Required | Description                        |
|---------|----------|------------------------------------|
| Name    | Yes      | Display name shown in the table    |
| URL     | Yes      | Direct stream URL (HTTP/HTTPS/HLS) |
| Genre   | No       | Genre label shown in the table     |
| Country | No       | Country label shown in the table   |

Fill in the fields, tab to the **Add** button, and press `Enter`. The station is saved immediately to `~/.config/pulse/radio.json` (or `$XDG_CONFIG_HOME/pulse/radio.json`) and appears at the bottom of the list highlighted in cyan.

Most stream formats that `mpv` supports work: direct MP3/AAC/OGG streams, HLS (`.m3u8`), and playlist files (`.pls`, `.m3u`).

### Removing a custom station

Select a custom station (cyan text) and press `D` or `Delete`. Built-in stations cannot be removed; only custom ones can be deleted.

### Filtering stations

`Ctrl+F` opens the same regexp filter as in file mode. It matches against station name, genre, and country simultaneously. Press `Esc` to clear the filter and restore the full list.

## Code Layout

- `main.go`: CLI argument validation and app bootstrap.
- `types.go`: shared state structs (`app`, `AudioFile`, `RadioStation`, ffprobe JSON types).
- `ui_layout.go`: main layout and keyboard wiring.
- `ui_overlay.go`: configuration and theme overlay flow.
- `ui_details.go`: lazy probe rendering and details/now-playing view updates.
- `ui_filter_status.go`: filter lifecycle and status message helpers.
- `ui_radio.go`: radio mode UI — station table, details pane, add/delete overlays, filter.
- `radio.go`: built-in station list, custom station persistence (`radio.json`).
- `themes.go`: built-in frame/table-header theme definitions.
- `scanner.go`: recursive audio file discovery.
- `media_probe.go`: `ffprobe` metadata extraction for a selected file.
- `player.go`: mpv integration for file playback and radio streaming.
- `actions.go`: conversion/shuffle/refresh action handlers.
- `helpers.go`: small formatting/string helpers.

## Controls

### File browser

| Key      | Action                                                              |
|----------|---------------------------------------------------------------------|
| `Tab`    | Switch focus between file table and action list                     |
| `↑` / `↓` | Navigate files                                                    |
| `Enter`  | Play selected file                                                  |
| `C`      | Open Configuration overlay (themes, border styles)                  |
| `Z`      | Shuffle currently visible list                                      |
| `S`      | Stop playback                                                       |
| `M`      | Mute / unmute (restores previous volume on unmute)                  |
| `+`      | Increase volume by 5%                                               |
| `-`      | Decrease volume by 5%                                               |
| `Ctrl+F` | Open regexp filter input                                            |
| `R`      | Switch to Radio mode                                                |
| `F5`     | Refresh — re-scan the directory                                     |
| `Esc`    | Close overlay / clear filter / quit                                 |

### Radio mode

| Key          | Action                                          |
|--------------|-------------------------------------------------|
| `↑` / `↓`   | Navigate stations                               |
| `Enter`      | Connect and stream the selected station         |
| `A`          | Open **Add Station** form (name + URL required) |
| `D` / `Del`  | Delete selected custom station                  |
| `S`          | Stop stream                                     |
| `M`          | Mute / unmute                                   |
| `+` / `-`    | Volume up / down                                |
| `Ctrl+F`     | Filter by name, genre, or country               |
| `R`          | Return to file browser                          |
| `Esc`        | Close overlay / clear filter / quit             |

## Notes

- Metadata is loaded lazily with `ffprobe` when a file is selected.
- Conversions create new files in the same directory as the source file.
