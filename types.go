package main

import (
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/rivo/tview"
)

// ffprobeStream holds audio stream metadata returned by ffprobe.
type ffprobeStream struct {
	CodecName string `json:"codec_name"` // e.g. "mp3", "vorbis"
	CodecType string `json:"codec_type"` // "audio", "video", etc.
	BitRate   string `json:"bit_rate"`   // stream-level bitrate in bits/s (string)
}

// ffprobeFormat holds container-level metadata returned by ffprobe.
type ffprobeFormat struct {
	BitRate  string            `json:"bit_rate"` // overall bitrate in bits/s (string)
	Duration string            `json:"duration"` // total duration in seconds (string)
	Tags     map[string]string `json:"tags"`     // ID3/Vorbis tags: title, artist, album, …
}

// ffprobeOutput is the top-level JSON structure returned by ffprobe -print_format json.
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

// AudioFile represents a single audio file discovered during the directory scan.
type AudioFile struct {
	Path     string            // absolute path on disk
	RelPath  string            // path relative to the scanned root dir (shown in the table)
	Name     string            // base filename
	Size     int64             // file size in bytes
	Format   string            // codec name in uppercase, e.g. "MP3", "FLAC"
	Bitrate  int               // bitrate in kbps (0 = unknown)
	Duration string            // formatted as "m:ss"
	Tags     map[string]string // metadata tags keyed by lowercase name
	Probed   bool              // true after ffprobe has been run for this file
}

// app is the root application struct that owns the TUI widgets and runtime state.
type app struct {
	tv               *tview.Application // tview event loop and screen manager
	table            *tview.Table       // main file list widget (top-right panel)
	detailsView      *tview.TextView    // file detail pane (top-left panel)
	actionList       *tview.List        // action menu widget (bottom panel)
	statusBar        *tview.TextView    // left portion of the status bar
	playingBar       *tview.TextView    // right portion: now-playing indicator
	hotkeysView      *tview.TextView    // hotkeys panel text
	searchBar        *tview.InputField  // regexp filter input shown on Ctrl+F
	statusPages      *tview.Pages       // switches between statusRow and searchBar
	rootPages        *tview.Pages       // top-level pages for overlay screens above the main layout
	detailsFrame     *tview.Flex        // bordered frame that wraps detailsView
	tableFrame       *tview.Flex        // bordered frame that wraps the file table
	actionsFrame     *tview.Flex        // bordered frame that wraps actionList
	hotkeysFrame     *tview.Flex        // bordered frame that wraps hotkeysView
	configList       *tview.List        // configuration menu shown when C is pressed
	themesList       *tview.List        // themes configuration menu with Colors/Border entries
	themeColorsList  *tview.List        // color palette submenu list
	borderStylesList *tview.List        // border style submenu list
	files            []*AudioFile       // currently displayed files (may be a filtered subset)
	allFiles         []*AudioFile       // full unfiltered list
	filteredBuf      []*AudioFile       // reusable backing slice to reduce filter allocations
	filterActive     bool               // true while the filter input bar is visible
	filterDebounce   *time.Timer        // coalesces rapid filter keystrokes into one apply
	overlayOpen      bool               // true while a modal overlay page is visible
	activeOverlay    string             // active overlay page name: "configuration", "themes", "theme-colors", "theme-borders", or ""
	colorPaletteName string             // selected color palette name
	borderStyleName  string             // selected border style name
	previousFocus    tview.Primitive    // focus owner before opening an overlay
	dir              string             // the directory path passed on the command line
	selectedFile     *AudioFile         // file currently highlighted in the table
	radioMode        bool               // true when showing the radio station browser
	stations         []*RadioStation    // currently displayed stations (may be filtered)
	allStations      []*RadioStation    // full station list (built-ins + custom)
	radioFilterBuf   []*RadioStation    // reusable backing slice for radio filter
	selectedStation  *RadioStation      // station currently highlighted in radio mode
	nowPlayingRadio  *RadioStation      // station currently streaming; nil when a file is playing
	radioTrack       string             // current ICY StreamTitle polled from mpv IPC ("Artist - Title")
	mpvSocketPath    string             // path to the mpv IPC socket for the active radio stream
	stopRadioPoller  chan struct{}       // closed to stop the radio track poller goroutine
	currentPlay      *exec.Cmd          // running player process; nil when nothing is playing
	nowPlaying       *AudioFile         // file currently being played; mirrors currentPlay
	playerName       string             // name of the player binary used for the current track
	playerBinary     string             // cached CLI player binary selected at startup
	playerBaseArgs   []string           // cached base arguments for the selected CLI player
	volume           int                // playback volume percent (0-100)
	muted            bool               // true when output is muted via M toggle
	volumeBeforeMute int                // volume snapshot to restore when unmuting
	playStart        time.Time          // when the current track started
	stopTicker       chan struct{}      // closed to stop the per-second position ticker
	probeDebounce    *time.Timer        // delays ffprobe launch while cursor is moving quickly
	probeRequestID   atomic.Int64       // monotonically increasing token to ignore stale probe results
	statusNonce      atomic.Int64       // bumps on each status update; used to cancel delayed clears
}
