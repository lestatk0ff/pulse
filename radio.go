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
