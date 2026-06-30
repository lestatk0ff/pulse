package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

// Keep a larger decode/audio buffer to avoid long-session underruns (stutter/crackle).
var mpvBaseArgs = []string{
	"--no-terminal",
	"--no-video",
	"--cache=yes",
	"--cache-secs=20",
	"--demuxer-readahead-secs=20",
	"--audio-buffer=5.0",
	"--cache-pause=no",
}

// findPlayer returns "mpv" and its default args if mpv is on PATH, otherwise empty strings.
func findPlayer() (string, []string) {
	if _, err := exec.LookPath("mpv"); err == nil {
		return "mpv", mpvBaseArgs
	}
	return "", nil
}

// playFile starts an external audio player for f in the background.
// Any previously playing file is killed first so only one track plays at a time.
func (a *app) playFile(f *AudioFile) {
	a.playFileFrom(f, 0)
}

// playRadio starts streaming a radio station URL. Any current playback is killed first.
func (a *app) playRadio(s *RadioStation) {
	playerName, playerArgs := a.playerBinary, a.playerBaseArgs
	if playerName == "" {
		a.setStatus("[red]No player found — install mpv")
		return
	}
	if a.currentPlay != nil {
		a.currentPlay.Process.Kill()
		a.currentPlay = nil
	}
	if a.nowPlaying != nil {
		a.sendToFileMpv("stop") //nolint:errcheck
		a.nowPlaying = nil
		a.stopPositionTicker()
	}
	a.paused = false
	a.refreshPlayingRowHighlight()
	a.updatePlayingBar()
	a.stopRadioTrackPoller()

	args := append([]string{}, playerArgs...)
	if a.volume < 0 {
		a.volume = 0
	}
	if a.volume > 100 {
		a.volume = 100
	}
	args = append(args, fmt.Sprintf("--volume=%d", a.volume))
	if eqArg := findEQPreset(a.equalizerPreset).mpvArg(); eqArg != "" {
		args = append(args, eqArg)
	}

	// mpv exposes a JSON IPC socket we can query for current ICY track title.
	var socketPath string
	if playerName == "mpv" {
		socketPath = fmt.Sprintf("/tmp/pulse-mpv-%d.sock", os.Getpid())
		os.Remove(socketPath)
		a.mpvSocketPath = socketPath
		args = append(args, "--input-ipc-server="+socketPath)
	}

	args = append(args, s.URL)

	cmd := exec.Command(playerName, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		a.setStatus(fmt.Sprintf("[red]Failed to start %s: %v", playerName, err))
		return
	}

	a.currentPlay = cmd
	a.nowPlayingRadio = s
	a.playerName = playerName
	a.playStart = time.Now()
	a.updatePlayingBar()
	a.startPositionTicker()

	if socketPath != "" {
		a.startRadioTrackPoller(socketPath, cmd)
	}

	go func() {
		err := cmd.Wait()
		a.tv.QueueUpdateDraw(func() {
			if a.currentPlay == cmd {
				a.currentPlay = nil
				a.nowPlayingRadio = nil
				a.playerName = ""
				a.stopPositionTicker()
				a.stopRadioTrackPoller()
				if err != nil {
					a.setStatus("[yellow]Stream disconnected")
				} else {
					a.setStatus("[grey]Stream ended")
				}
			}
		})
	}()
}

// startRadioTrackPoller waits for the mpv IPC socket then polls media-title every 5 seconds.
func (a *app) startRadioTrackPoller(socketPath string, cmd *exec.Cmd) {
	ch := make(chan struct{})
	a.stopRadioPoller = ch

	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(socketPath); err == nil {
				break
			}
			select {
			case <-ch:
				return
			case <-time.After(100 * time.Millisecond):
			}
		}

		for {
			title := queryMpvTitle(socketPath)
			if title != "" {
				a.tv.QueueUpdateDraw(func() {
					if a.currentPlay == cmd && title != a.radioTrack {
						a.radioTrack = title
						a.updatePlayingBar()
					}
				})
			}
			select {
			case <-ch:
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()
}

// stopRadioTrackPoller stops the poller goroutine and clears track state.
func (a *app) stopRadioTrackPoller() {
	if a.stopRadioPoller != nil {
		close(a.stopRadioPoller)
		a.stopRadioPoller = nil
	}
	a.radioTrack = ""
	if a.mpvSocketPath != "" {
		os.Remove(a.mpvSocketPath)
		a.mpvSocketPath = ""
	}
}

// queryMpvTitle connects to the mpv IPC socket and returns the current media-title.
func queryMpvTitle(socketPath string) string {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return ""
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second)) //nolint:errcheck

	fmt.Fprintf(conn, `{"command":["get_property","media-title"],"request_id":1}`+"\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var resp struct {
			Data      interface{} `json:"data"`
			Error     string      `json:"error"`
			RequestID int         `json:"request_id"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue
		}
		if resp.RequestID == 1 && resp.Error == "success" {
			if title, ok := resp.Data.(string); ok {
				return title
			}
		}
	}
	return ""
}

// playFileFrom loads a file into the persistent file-mode mpv via IPC.
func (a *app) playFileFrom(f *AudioFile, startSeconds int) {
	if a.playerBinary == "" {
		a.setStatus("[red]No player found — install mpv")
		return
	}

	a.stopRadioTrackPoller()
	if a.currentPlay != nil {
		a.currentPlay.Process.Kill()
		a.currentPlay = nil
	}
	a.nowPlayingRadio = nil
	a.nowPlaying = nil
	a.paused = false
	a.refreshPlayingRowHighlight()
	a.updatePlayingBar()

	if err := a.ensureFilePlayer(); err != nil {
		a.setStatus(fmt.Sprintf("[red]Failed to start mpv: %v", err))
		return
	}

	opts := ""
	if startSeconds > 0 {
		opts = fmt.Sprintf("start=%d", startSeconds)
	}
	a.sendToFileMpv("loadfile", f.Path, "replace", opts) //nolint:errcheck

	if a.volume < 0 {
		a.volume = 0
	}
	if a.volume > 100 {
		a.volume = 100
	}
	a.sendToFileMpv("set_property", "volume", a.volume)                             //nolint:errcheck
	a.sendToFileMpv("set_property", "af", findEQPreset(a.equalizerPreset).afValue()) //nolint:errcheck

	a.nowPlaying = f
	a.playerName = a.playerBinary
	a.playStart = time.Now().Add(-time.Duration(startSeconds) * time.Second)
	a.refreshPlayingRowHighlight()
	a.updatePlayingBar()
	a.startPositionTicker()
}

// startFilePlayer starts a persistent mpv process in idle mode and begins listening for events.
func (a *app) startFilePlayer() error {
	socketPath := fmt.Sprintf("/tmp/pulse-file-%d.sock", os.Getpid())
	os.Remove(socketPath)
	a.filePlayerSocket = socketPath

	args := append([]string{}, a.playerBaseArgs...)
	args = append(args, "--idle=yes", "--input-ipc-server="+socketPath)
	cmd := exec.Command(a.playerBinary, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	if err := cmd.Start(); err != nil {
		return err
	}
	a.filePlayer = cmd
	a.startFileEventListener()
	return nil
}

// ensureFilePlayer starts the file-mode mpv if it is not already running.
func (a *app) ensureFilePlayer() error {
	if a.isFilePlayerRunning() {
		return nil
	}
	a.stopFileEventListener()
	if a.filePlayer != nil && a.filePlayer.Process != nil {
		a.filePlayer.Process.Kill()
		a.filePlayer.Wait() //nolint:errcheck
		a.filePlayer = nil
	}
	if err := a.startFilePlayer(); err != nil {
		return err
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if a.isFilePlayerRunning() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("mpv IPC socket not ready")
}

// isFilePlayerRunning returns true when the file-mode mpv IPC socket is reachable.
func (a *app) isFilePlayerRunning() bool {
	if a.filePlayerSocket == "" {
		return false
	}
	conn, err := net.Dial("unix", a.filePlayerSocket)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// sendToMpvSocket sends a single JSON IPC command to an mpv process via its Unix socket.
func sendToMpvSocket(socketPath string, args ...interface{}) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second)) //nolint:errcheck
	data, _ := json.Marshal(map[string]interface{}{"command": args})
	_, err = fmt.Fprintf(conn, "%s\n", data)
	return err
}

// sendToFileMpv sends a single JSON IPC command to the file-mode mpv process.
func (a *app) sendToFileMpv(args ...interface{}) error {
	return sendToMpvSocket(a.filePlayerSocket, args...)
}

// startFileEventListener starts a goroutine that reads MPV events and triggers auto-advance on eof.
func (a *app) startFileEventListener() {
	ch := make(chan struct{})
	a.stopFileEvents = ch
	socketPath := a.filePlayerSocket

	go func() {
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(socketPath); err == nil {
				break
			}
			select {
			case <-ch:
				return
			case <-time.After(50 * time.Millisecond):
			}
		}

		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			return
		}
		defer conn.Close()

		go func() {
			<-ch
			conn.Close()
		}()

		type mpvEvent struct {
			Event  string `json:"event"`
			Reason string `json:"reason"`
		}
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var ev mpvEvent
			if json.Unmarshal(scanner.Bytes(), &ev) != nil {
				continue
			}
			if ev.Event == "end-file" && ev.Reason == "eof" {
				a.tv.QueueUpdateDraw(func() {
					if a.nowPlaying != nil {
						finished := a.nowPlaying
						a.nowPlaying = nil
						a.playerName = ""
						a.stopPositionTicker()
						a.refreshPlayingRowHighlight()
						a.updatePlayingBar()
						a.advanceToNext(finished)
					}
				})
			}
		}
	}()
}

// stopFileEventListener stops the file event listener goroutine.
func (a *app) stopFileEventListener() {
	if a.stopFileEvents != nil {
		close(a.stopFileEvents)
		a.stopFileEvents = nil
	}
}

// shutdownFilePlayer stops the event listener and kills the persistent mpv process.
func (a *app) shutdownFilePlayer() {
	a.stopFileEventListener()
	if a.filePlayer != nil && a.filePlayer.Process != nil {
		a.filePlayer.Process.Kill()
		a.filePlayer.Wait() //nolint:errcheck
		a.filePlayer = nil
	}
	if a.filePlayerSocket != "" {
		os.Remove(a.filePlayerSocket)
		a.filePlayerSocket = ""
	}
}

func (a *app) adjustVolume(delta int) {
	// If user changes volume while muted, treat that as an explicit unmute action.
	// We keep `volume` as the effective output volume used for launching the player.
	if a.muted {
		a.muted = false
		a.volumeBeforeMute = a.volume
	}
	old := a.volume
	newVolume := old + delta
	if newVolume < 0 {
		newVolume = 0
	}
	if newVolume > 100 {
		newVolume = 100
	}
	if newVolume == old {
		a.setStatusTemporary(fmt.Sprintf("[grey]Volume already at %d%%", old), 3*time.Second)
		return
	}
	a.volume = newVolume
	a.saveConfig()
	if a.nowPlayingRadio != nil {
		station := a.nowPlayingRadio
		a.playRadio(station)
		a.setStatusTemporary(fmt.Sprintf("[green]Volume:[white] %d%%", a.volume), 3*time.Second)
		return
	}
	if a.nowPlaying == nil {
		a.setStatusTemporary(fmt.Sprintf("[green]Volume set:[white] %d%%", a.volume), 3*time.Second)
		return
	}
	a.sendToFileMpv("set_property", "volume", a.volume) //nolint:errcheck
	a.setStatusTemporary(fmt.Sprintf("[green]Volume:[white] %d%%", a.volume), 3*time.Second)
}

func (a *app) toggleMute() {
	if a.muted {
		// Unmute restores the exact snapshot captured when mute was enabled.
		a.muted = false
		restore := a.volumeBeforeMute
		if restore < 0 {
			restore = 0
		}
		if restore > 100 {
			restore = 100
		}
		a.volume = restore
		if a.nowPlayingRadio != nil {
			station := a.nowPlayingRadio
			a.playRadio(station)
			a.setStatusTemporary(fmt.Sprintf("[green]Unmuted:[white] %d%%", a.volume), 3*time.Second)
			return
		}
		if a.nowPlaying == nil {
			a.setStatusTemporary(fmt.Sprintf("[green]Unmuted:[white] %d%%", a.volume), 3*time.Second)
			return
		}
		a.sendToFileMpv("set_property", "volume", a.volume) //nolint:errcheck
		a.setStatusTemporary(fmt.Sprintf("[green]Unmuted:[white] %d%%", a.volume), 3*time.Second)
		return
	}

	// Mute keeps previous volume in memory so a second M restores it.
	a.muted = true
	a.volumeBeforeMute = a.volume
	a.volume = 0
	if a.nowPlayingRadio != nil {
		station := a.nowPlayingRadio
		a.playRadio(station)
		a.setStatusTemporary("[yellow]Muted", 3*time.Second)
		return
	}
	if a.nowPlaying == nil {
		a.setStatusTemporary("[yellow]Muted", 3*time.Second)
		return
	}
	a.sendToFileMpv("set_property", "volume", 0) //nolint:errcheck
	a.setStatusTemporary("[yellow]Muted", 3*time.Second)
}

// togglePause pauses or resumes the current playback via mpv's cycle pause command.
func (a *app) togglePause() {
	if a.nowPlaying == nil && a.nowPlayingRadio == nil {
		a.setStatus("[grey]Nothing is playing")
		return
	}
	if a.paused {
		a.playStart = a.playStart.Add(time.Since(a.pausedAt))
		a.paused = false
		if a.nowPlaying != nil {
			a.sendToFileMpv("cycle", "pause") //nolint:errcheck
		} else {
			sendToMpvSocket(a.mpvSocketPath, "cycle", "pause") //nolint:errcheck
		}
		a.startPositionTicker()
	} else {
		a.pausedAt = time.Now()
		a.paused = true
		if a.nowPlaying != nil {
			a.sendToFileMpv("cycle", "pause") //nolint:errcheck
		} else {
			sendToMpvSocket(a.mpvSocketPath, "cycle", "pause") //nolint:errcheck
		}
		a.stopPositionTicker()
	}
	a.updatePlayingBar()
}

// advanceToNext plays the file after finished in the current list, or clears state if at the end.
func (a *app) advanceToNext(finished *AudioFile) {
	for i, f := range a.files {
		if f == finished {
			if i+1 < len(a.files) {
				next := a.files[i+1]
				// Keep table selection in sync with auto-play progression.
				a.table.Select(i+2, 0) // row 0 is the header; file i+1 is at row i+2
				a.selectedFile = next
				a.probeAndShowDetails(next)
				a.playFile(next)
			}
			return
		}
	}
}

// stopPlayback stops the currently playing file or radio stream.
func (a *app) stopPlayback() {
	if a.nowPlaying != nil {
		a.sendToFileMpv("stop") //nolint:errcheck
		a.nowPlaying = nil
		a.playerName = ""
		a.paused = false
		a.stopPositionTicker()
		a.refreshPlayingRowHighlight()
		a.updatePlayingBar()
		return
	}
	if a.currentPlay != nil {
		a.currentPlay.Process.Kill()
		a.currentPlay.Wait() //nolint:errcheck
		a.currentPlay = nil
		a.nowPlayingRadio = nil
		a.playerName = ""
		a.paused = false
		a.stopPositionTicker()
		a.stopRadioTrackPoller()
		a.refreshPlayingRowHighlight()
		a.updatePlayingBar()
		return
	}
	a.setStatus("[grey]Nothing is playing")
}

// startPositionTicker stops any running ticker and starts a new one that updates
// the playing bar once per second.
func (a *app) startPositionTicker() {
	a.stopPositionTicker()
	ch := make(chan struct{})
	a.stopTicker = ch
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ch:
				return
			case <-ticker.C:
				a.tv.QueueUpdateDraw(func() {
					a.updatePlayingBar()
				})
			}
		}
	}()
}

// stopPositionTicker stops the position ticker goroutine if one is running.
func (a *app) stopPositionTicker() {
	if a.stopTicker != nil {
		close(a.stopTicker)
		a.stopTicker = nil
	}
}
