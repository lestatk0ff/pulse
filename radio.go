package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RadioStation represents an online radio stream.
type RadioStation struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Genre   string `json:"genre"`
	Country string `json:"country"`
	Bitrate string `json:"bitrate"`
	Custom  bool   `json:"custom,omitempty"`
}

var builtinStations = []*RadioStation{
	{Name: "Groove Salad", URL: "https://ice1.somafm.com/groovesalad-256-mp3", Genre: "Ambient", Country: "US", Bitrate: "256k"},
	{Name: "Secret Agent", URL: "https://ice1.somafm.com/secretagent-128-mp3", Genre: "Lounge", Country: "US", Bitrate: "128k"},
	{Name: "Drone Zone", URL: "https://ice1.somafm.com/dronezone-256-mp3", Genre: "Ambient", Country: "US", Bitrate: "256k"},
	{Name: "Lush", URL: "https://ice1.somafm.com/lush-128-mp3", Genre: "Vocal Lounge", Country: "US", Bitrate: "128k"},
	{Name: "Indie Pop Rocks", URL: "https://ice1.somafm.com/indiepop-128-mp3", Genre: "Indie Pop", Country: "US", Bitrate: "128k"},
	{Name: "Metal Detector", URL: "https://ice1.somafm.com/metal-128-mp3", Genre: "Metal", Country: "US", Bitrate: "128k"},
	{Name: "Deep Space One", URL: "https://ice1.somafm.com/deepspaceone-256-mp3", Genre: "Ambient", Country: "US", Bitrate: "256k"},
	{Name: "Folk Forward", URL: "https://ice1.somafm.com/folkfwd-128-mp3", Genre: "Folk", Country: "US", Bitrate: "128k"},
	{Name: "Reggae Cafe", URL: "https://ice1.somafm.com/reggae-128-mp3", Genre: "Reggae", Country: "US", Bitrate: "128k"},
	{Name: "Seven Inch Soul", URL: "https://ice1.somafm.com/7soul-128-mp3", Genre: "Soul/R&B", Country: "US", Bitrate: "128k"},
	{Name: "Beat Blender", URL: "https://ice1.somafm.com/beatblender-128-mp3", Genre: "Electronic", Country: "US", Bitrate: "128k"},
	{Name: "Underground 80s", URL: "https://ice1.somafm.com/u80s-256-mp3", Genre: "80s", Country: "US", Bitrate: "256k"},
	{Name: "BBC Radio 1", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_radio_one", Genre: "Pop/Dance", Country: "UK", Bitrate: "128k"},
	{Name: "BBC Radio 2", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_radio_two", Genre: "Mixed", Country: "UK", Bitrate: "128k"},
	{Name: "BBC Radio 4", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_radio_fourfm", Genre: "Talk/Drama", Country: "UK", Bitrate: "128k"},
	{Name: "BBC World Service", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_world_service", Genre: "News/Talk", Country: "UK", Bitrate: "96k"},
	{Name: "KEXP 90.3", URL: "https://kexp-mp3-128.streamguys1.com/kexp128.mp3", Genre: "Indie/Alt", Country: "US", Bitrate: "128k"},
}

func radioConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "pulse", "radio.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pulse", "radio.json")
}

func loadCustomStations() []*RadioStation {
	data, err := os.ReadFile(radioConfigPath())
	if err != nil {
		return nil
	}
	var stations []*RadioStation
	if err := json.Unmarshal(data, &stations); err != nil {
		return nil
	}
	for _, s := range stations {
		s.Custom = true
	}
	return stations
}

func saveCustomStations(stations []*RadioStation) error {
	var custom []*RadioStation
	for _, s := range stations {
		if s.Custom {
			custom = append(custom, s)
		}
	}
	data, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		return err
	}
	path := radioConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func buildStationList() []*RadioStation {
	stations := make([]*RadioStation, len(builtinStations))
	copy(stations, builtinStations)
	return append(stations, loadCustomStations()...)
}
