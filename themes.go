package main

import "github.com/gdamore/tcell/v2"

// Theme defines colors used by frame borders/titles and table headers.
type Theme struct {
	Name                 string
	DetailsFrameColor    tcell.Color
	TableFrameColor      tcell.Color
	ActionsFrameColor    tcell.Color
	HotkeysFrameColor    tcell.Color
	OverlayFrameColor    tcell.Color
	TableHeaderTextColor tcell.Color
	TableHeaderBgColor   tcell.Color
}

var themeOrder = []string{"Default", "Dark blue", "Blue", "Space", "Sunset", "Mono"}

var themeRegistry = map[string]Theme{
	"Default": {
		Name:                 "Default",
		DetailsFrameColor:    tcell.ColorYellow,
		TableFrameColor:      tcell.ColorGreen,
		ActionsFrameColor:    tcell.ColorAqua,
		HotkeysFrameColor:    tcell.ColorAqua,
		OverlayFrameColor:    tcell.ColorYellow,
		TableHeaderTextColor: tcell.ColorYellow,
		TableHeaderBgColor:   tcell.ColorDarkBlue,
	},
	"Dark blue": {
		Name:                 "Dark blue",
		DetailsFrameColor:    tcell.ColorDarkBlue,
		TableFrameColor:      tcell.ColorDarkBlue,
		ActionsFrameColor:    tcell.ColorDarkBlue,
		HotkeysFrameColor:    tcell.ColorDarkBlue,
		OverlayFrameColor:    tcell.ColorDarkBlue,
		TableHeaderTextColor: tcell.ColorWhite,
		TableHeaderBgColor:   tcell.ColorDarkBlue,
	},
	"Blue": {
		Name:                 "Blue",
		DetailsFrameColor:    tcell.ColorBlue,
		TableFrameColor:      tcell.ColorBlue,
		ActionsFrameColor:    tcell.ColorBlue,
		HotkeysFrameColor:    tcell.ColorBlue,
		OverlayFrameColor:    tcell.ColorBlue,
		TableHeaderTextColor: tcell.ColorWhite,
		TableHeaderBgColor:   tcell.ColorBlue,
	},
	"Space": {
		Name:                 "Space",
		DetailsFrameColor:    tcell.NewRGBColor(120, 170, 255),
		TableFrameColor:      tcell.NewRGBColor(120, 170, 255),
		ActionsFrameColor:    tcell.NewRGBColor(120, 170, 255),
		HotkeysFrameColor:    tcell.NewRGBColor(120, 170, 255),
		OverlayFrameColor:    tcell.NewRGBColor(190, 210, 255),
		TableHeaderTextColor: tcell.NewRGBColor(230, 240, 255),
		TableHeaderBgColor:   tcell.NewRGBColor(18, 26, 46),
	},
	"Sunset": {
		Name:                 "Sunset",
		DetailsFrameColor:    tcell.NewRGBColor(255, 170, 80),
		TableFrameColor:      tcell.NewRGBColor(255, 170, 80),
		ActionsFrameColor:    tcell.NewRGBColor(255, 170, 80),
		HotkeysFrameColor:    tcell.NewRGBColor(255, 170, 80),
		OverlayFrameColor:    tcell.NewRGBColor(255, 200, 130),
		TableHeaderTextColor: tcell.NewRGBColor(255, 245, 225),
		TableHeaderBgColor:   tcell.NewRGBColor(88, 36, 20),
	},
	"Mono": {
		Name:                 "Mono",
		DetailsFrameColor:    tcell.NewRGBColor(180, 180, 180),
		TableFrameColor:      tcell.NewRGBColor(180, 180, 180),
		ActionsFrameColor:    tcell.NewRGBColor(180, 180, 180),
		HotkeysFrameColor:    tcell.NewRGBColor(180, 180, 180),
		OverlayFrameColor:    tcell.NewRGBColor(210, 210, 210),
		TableHeaderTextColor: tcell.NewRGBColor(240, 240, 240),
		TableHeaderBgColor:   tcell.NewRGBColor(45, 45, 45),
	},
}

func themeByName(name string) Theme {
	if t, ok := themeRegistry[name]; ok {
		return t
	}
	return themeRegistry["Default"]
}
