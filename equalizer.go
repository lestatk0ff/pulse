package main

import (
	"fmt"
	"math"
	"strings"
)

// EQPreset holds per-band dB gains for superequalizer (18 bands).
// Bands map to: 65 92 131 185 262 370 523 740 1047 1480 2093 2960 4186 5920 8372 11840 16744 22000 Hz
type EQPreset struct {
	Name        string
	Description string
	GainsDB     [18]float64 // dB offset per band; 0 = flat
}

var eqPresets = []EQPreset{
	{
		Name: "Flat", Description: "No equalizer — original audio output",
		GainsDB: [18]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	{
		Name: "Rock", Description: "Punchy bass and crisp highs",
		GainsDB: [18]float64{5, 4, 3, 2, 0, -1, -1, -1, 0, 1, 2, 3, 3, 4, 4, 3, 2, 1},
	},
	{
		Name: "Pop", Description: "Boosted mids for a warm, present sound",
		GainsDB: [18]float64{-1, 0, 1, 2, 3, 4, 4, 3, 2, 2, 2, 1, 0, 0, -1, -1, -1, -1},
	},
	{
		Name: "Jazz", Description: "Warm lows and smooth highs",
		GainsDB: [18]float64{4, 3, 2, 2, 1, 0, -1, -1, 0, 1, 1, 2, 2, 2, 1, 0, 0, 0},
	},
	{
		Name: "Classical", Description: "Wide soundstage with natural tonality",
		GainsDB: [18]float64{4, 3, 3, 2, 1, 0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 2, 1, 0},
	},
	{
		Name: "Bass Boost", Description: "Deep, powerful low frequencies",
		GainsDB: [18]float64{8, 7, 6, 5, 4, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	{
		Name: "Vocal", Description: "Enhanced vocals and speech clarity",
		GainsDB: [18]float64{-2, -2, -1, 0, 1, 2, 3, 4, 5, 5, 4, 3, 2, 1, 0, 0, 0, 0},
	},
}

const defaultEQPreset = "Rock"

func findEQPreset(name string) *EQPreset {
	for i := range eqPresets {
		if eqPresets[i].Name == name {
			return &eqPresets[i]
		}
	}
	return &eqPresets[0]
}

// afValue returns the bare lavfi filter string used for IPC set_property af.
// Returns empty string for the Flat preset (no filter).
func (p *EQPreset) afValue() string {
	if p.Name == "Flat" || p.Name == "" {
		return ""
	}
	parts := make([]string, 18)
	for i, db := range p.GainsDB {
		parts[i] = fmt.Sprintf("%db=%.3f", i+1, math.Pow(10, db/20.0))
	}
	return "lavfi=[superequalizer=" + strings.Join(parts, ":") + "]"
}

// mpvArg returns the --af argument for mpv using lavfi superequalizer.
// superequalizer expects linear gain values in [0, 20] where 1.0 = unity.
func (p *EQPreset) mpvArg() string {
	if p.Name == "Flat" || p.Name == "" {
		return ""
	}
	parts := make([]string, 18)
	for i, db := range p.GainsDB {
		linear := math.Pow(10, db/20.0)
		parts[i] = fmt.Sprintf("%db=%.3f", i+1, linear)
	}
	return "--af=lavfi=[superequalizer=" + strings.Join(parts, ":") + "]"
}
