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

var mpvBaseArgs = []string{"--no-terminal", "--no-video"}

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
	a.nowPlaying = nil
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

// playFileFrom starts playback from the given offset (in whole seconds).
func (a *app) playFileFrom(f *AudioFile, startSeconds int) {
	playerName, playerArgs := a.playerBinary, a.playerBaseArgs
	if playerName == "" {
		a.setStatus("[red]No player found — install mpv")
		return
	}

	// Kill the current track before starting a new one.
	if a.currentPlay != nil {
		a.currentPlay.Process.Kill()
		a.currentPlay = nil
	}
	a.nowPlayingRadio = nil

	args := a.playerCommandArgs(playerName, playerArgs, f.Path, startSeconds)
	cmd := exec.Command(playerName, args...)
	// Detach all stdio so the player doesn't interfere with the TUI.
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		a.setStatus(fmt.Sprintf("[red]Failed to start %s: %v", playerName, err))
		return
	}

	a.currentPlay = cmd
	a.nowPlaying = f
	a.playerName = playerName
	a.playStart = time.Now().Add(-time.Duration(startSeconds) * time.Second)
	a.startPositionTicker()

	// Wait in a goroutine so we can advance to the next track when this one ends naturally.
	go func() {
		err := cmd.Wait()
		a.tv.QueueUpdateDraw(func() {
			// Guard against a race where stopPlayback already replaced currentPlay.
			if a.currentPlay == cmd {
				a.currentPlay = nil
				finished := a.nowPlaying
				a.nowPlaying = nil
				a.playerName = ""
				a.stopPositionTicker()
				// Only auto-advance when playback ended cleanly.
				if err == nil {
					a.advanceToNext(finished)
				}
			}
		})
	}()
}

// playerCommandArgs adds common playback modifiers for supported players.
func (a *app) playerCommandArgs(playerName string, playerArgs []string, path string, startSeconds int) []string {
	args := append([]string{}, playerArgs...)
	if startSeconds > 0 {
		args = append(args, fmt.Sprintf("--start=%d", startSeconds))
	}
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
	args = append(args, path)
	return args
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
	// This app controls external CLI players (not embedded audio APIs), so volume
	// changes are applied by restarting playback at the current elapsed position.
	resumeSeconds := int(time.Since(a.playStart).Seconds())
	playing := a.nowPlaying
	a.playFileFrom(playing, resumeSeconds)
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
		// Apply restored volume immediately by restarting from current playback time.
		resumeSeconds := int(time.Since(a.playStart).Seconds())
		playing := a.nowPlaying
		a.playFileFrom(playing, resumeSeconds)
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
	resumeSeconds := int(time.Since(a.playStart).Seconds())
	playing := a.nowPlaying
	a.playFileFrom(playing, resumeSeconds)
	a.setStatusTemporary("[yellow]Muted", 3*time.Second)
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

// stopPlayback kills the currently playing process, if any.
func (a *app) stopPlayback() {
	if a.currentPlay == nil {
		a.setStatus("[grey]Nothing is playing")
		return
	}
	a.currentPlay.Process.Kill()
	a.currentPlay.Wait()
	a.currentPlay = nil
	a.nowPlaying = nil
	a.nowPlayingRadio = nil
	a.playerName = ""
	a.stopPositionTicker()
	a.stopRadioTrackPoller()
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
