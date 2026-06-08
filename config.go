package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	defaultColorPalette = "Default"
	defaultBorderStyle  = "Classic"
	defaultVolume       = 80
	defaultBackground   = false
)

// Config holds user-overridden preferences. Only fields that differ from
// defaults are written to disk, so the file stays minimal and readable.
type Config struct {
	ColorPalette    string `yaml:"color_palette,omitempty"`
	BorderStyle     string `yaml:"border_style,omitempty"`
	Volume          *int   `yaml:"volume,omitempty"`
	Background      *bool  `yaml:"background,omitempty"`
	EqualizerPreset string `yaml:"equalizer_preset,omitempty"`
}

func configFilePath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "pulse", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pulse", "config.yaml")
}

// loadConfig reads ~/.config/pulse/config.yaml and returns a Config with
// whatever overrides are present. Missing file or parse errors yield an empty
// Config (all defaults apply).
func loadConfig() *Config {
	cfg := &Config{}
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, cfg)
	return cfg
}

// saveConfig writes only the settings that differ from defaults to
// ~/.config/pulse/config.yaml. If every setting is at its default the file
// will contain an empty YAML mapping.
func (a *app) saveConfig() {
	out := &Config{}
	if a.colorPaletteName != defaultColorPalette {
		out.ColorPalette = a.colorPaletteName
	}
	if a.borderStyleName != defaultBorderStyle {
		out.BorderStyle = a.borderStyleName
	}
	if a.volume != defaultVolume {
		v := a.volume
		out.Volume = &v
	}
	if a.backgroundEnabled {
		v := true
		out.Background = &v
	}
	if a.equalizerPreset != defaultEQPreset {
		out.EqualizerPreset = a.equalizerPreset
	}

	path := configFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	data, err := yaml.Marshal(out)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}
