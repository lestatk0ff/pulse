# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Terminal UI (TUI) music player written in Go using `tview` + `tcell`. Scans a directory of audio files, displays metadata, plays tracks via `mpv`, supports online radio streaming, and runs FFmpeg conversion actions.

**Runtime dependencies (must be in PATH):** `mpv`, `ffmpeg`, `ffprobe`

## Commands

```bash
# Run in development
go run . /path/to/music

# Build binary
go build -o pulse .

# Validate changes (run all three after edits)
go vet ./...
go build ./...
go test ./...   # no test files currently; still the sanity check
```

## Architecture

Single module, flat package (`package main`). All `.go` files live in the repo root.

**State** — `types.go` holds the central `app` struct, which owns every `tview` widget pointer and all runtime state (`paused`, `volume`, `currentFile`, `filterActive`, etc.). Pass `*app` rather than threading state through function arguments.

**UI construction** — `ui_layout.go:newTUIApp()` builds the full layout, wires global keybindings, applies the initial theme, and loads config. The layout is a `tview.Flex` composed of: file table (center), details pane (right), now-playing bar (bottom), and an optionally visible Actions panel (`actionsFrame`).

**UI files split by concern:**
- `ui_layout.go` — root layout, keybindings, theme application, Actions panel toggle
- `ui_overlay.go` — Configuration/Themes/Equalizer overlay flow
- `ui_details.go` — lazy `ffprobe` metadata rendering and now-playing bar updates
- `ui_filter_status.go` — filter lifecycle and `setStatus`/`setStatusAsync` helpers
- `ui_radio.go` — Radio mode: station table, add/delete overlays, ICY track polling

**Non-UI files:**
- `player.go` — mpv process lifecycle, pause/resume, volume, EQ args
- `actions.go` — background FFmpeg conversion and shuffle/refresh handlers
- `scanner.go` — recursive audio file discovery
- `media_probe.go` — `ffprobe` JSON extraction
- `radio.go` — `RadioStation` model, built-in list, custom station persistence (`~/.config/pulse/radio.json`)
- `config.go` — `Config` struct, `loadConfig`/`saveConfig`, XDG path resolution
- `themes.go` — all color palette and border style definitions
- `equalizer.go` — EQ preset catalog; `mpvArg()` converts dB gains to linear (1.0 = flat, range 0–20)

## Key Conventions

- **Thread safety:** slow operations (probing, conversion, playback) run in goroutines; all widget updates go through `app.QueueUpdateDraw`.
- **Status messages:** use `a.setStatus(msg)` (main goroutine) or `a.setStatusAsync(msg)` (background goroutine).
- **Config persistence:** to add a new persisted setting — add field to `Config` in `config.go`, apply it in `newTUIApp`, add UI toggle in `ui_overlay.go`, call `a.saveConfig()` at the change site.
- **Themes:** to add/remove color palettes or border styles, edit `colorPaletteOrder`/`borderStyleOrder` and their registry variables in `themes.go` — don't add color switch blocks in UI files.
- **EQ presets:** append an `EQPreset` to `eqPresets` in `equalizer.go`. `GainsDB` values are in dB; conversion to linear happens in `mpvArg()`. Do not pass raw dB integers directly to mpv.
- **Background transparency:** `tview.List` bakes background into text styles at creation time — update both box background and text styles together in `applyListBackground` (`ui_layout.go`).

## Debugging Hints

| Symptom | Start here |
|---|---|
| Playback / pause issues | `player.go` → `togglePause`; `types.go` → `paused`/`pausedAt`; `ui_details.go` → `updatePlayingBar` |
| Metadata wrong/missing | `media_probe.go`, then `ui_details.go` |
| Scan misses files | `scanner.go`, then table refresh in `actions.go` |
| UI hangs | Look for blocking calls in `ui_layout.go`, `ui_overlay.go`, `ui_details.go`; move to goroutine |
| Theme/color broken | `ui_overlay.go` overlay flow → `applyTheme`/`applyBackgroundColor` in `ui_layout.go` → `themes.go` |
| Settings not persisting | `config.go` (`loadConfig`/`saveConfig`) → call site in `newTUIApp` → save callers in `ui_overlay.go` and `player.go` |
| Radio station persistence | `radio.go` (`loadCustomStations`/`saveCustomStations`) → call sites in `ui_radio.go` |
