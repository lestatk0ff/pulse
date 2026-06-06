package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ColorPalette groups all color-related values used by the UI theme.
type ColorPalette struct {
	Name                 string
	DetailsFrameColor    tcell.Color
	TableFrameColor      tcell.Color
	ActionsFrameColor    tcell.Color
	HotkeysFrameColor    tcell.Color
	OverlayFrameColor    tcell.Color
	TableHeaderTextColor tcell.Color
	TableHeaderBgColor   tcell.Color
}

// BorderRunes contains the characters used to draw primitive borders.
type BorderRunes struct {
	Horizontal  rune
	Vertical    rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	LeftT       rune
	RightT      rune
	TopT        rune
	BottomT     rune
	Cross       rune
}

// BorderStyle groups border rendering attributes and glyphs used by the UI.
type BorderStyle struct {
	Name              string
	FrameAttributes   tcell.AttrMask
	OverlayAttributes tcell.AttrMask
	NormalRunes       BorderRunes
	FocusRunes        BorderRunes
}

var colorPaletteOrder = []string{"Default", "Dark blue", "Blue", "Space", "Sunset", "Mono"}

var colorPaletteRegistry = map[string]ColorPalette{
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

var borderStyleOrder = []string{"Classic", "Rounded", "Double", "Heavy"}

var borderStyleRegistry = map[string]BorderStyle{
	"Classic": {
		Name:              "Classic",
		FrameAttributes:   tcell.AttrNone,
		OverlayAttributes: tcell.AttrNone,
		NormalRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsLightHorizontal,
			Vertical:    tview.BoxDrawingsLightVertical,
			TopLeft:     tview.BoxDrawingsLightDownAndRight,
			TopRight:    tview.BoxDrawingsLightDownAndLeft,
			BottomLeft:  tview.BoxDrawingsLightUpAndRight,
			BottomRight: tview.BoxDrawingsLightUpAndLeft,
			LeftT:       tview.BoxDrawingsLightVerticalAndRight,
			RightT:      tview.BoxDrawingsLightVerticalAndLeft,
			TopT:        tview.BoxDrawingsLightDownAndHorizontal,
			BottomT:     tview.BoxDrawingsLightUpAndHorizontal,
			Cross:       tview.BoxDrawingsLightVerticalAndHorizontal,
		},
		FocusRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsDoubleHorizontal,
			Vertical:    tview.BoxDrawingsDoubleVertical,
			TopLeft:     tview.BoxDrawingsDoubleDownAndRight,
			TopRight:    tview.BoxDrawingsDoubleDownAndLeft,
			BottomLeft:  tview.BoxDrawingsDoubleUpAndRight,
			BottomRight: tview.BoxDrawingsDoubleUpAndLeft,
			LeftT:       tview.BoxDrawingsDoubleVerticalAndRight,
			RightT:      tview.BoxDrawingsDoubleVerticalAndLeft,
			TopT:        tview.BoxDrawingsDoubleDownAndHorizontal,
			BottomT:     tview.BoxDrawingsDoubleUpAndHorizontal,
			Cross:       tview.BoxDrawingsDoubleVerticalAndHorizontal,
		},
	},
	"Rounded": {
		Name:              "Rounded",
		FrameAttributes:   tcell.AttrNone,
		OverlayAttributes: tcell.AttrBold,
		NormalRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsLightHorizontal,
			Vertical:    tview.BoxDrawingsLightVertical,
			TopLeft:     tview.BoxDrawingsLightArcDownAndRight,
			TopRight:    tview.BoxDrawingsLightArcDownAndLeft,
			BottomLeft:  tview.BoxDrawingsLightArcUpAndRight,
			BottomRight: tview.BoxDrawingsLightArcUpAndLeft,
			LeftT:       tview.BoxDrawingsLightVerticalAndRight,
			RightT:      tview.BoxDrawingsLightVerticalAndLeft,
			TopT:        tview.BoxDrawingsLightDownAndHorizontal,
			BottomT:     tview.BoxDrawingsLightUpAndHorizontal,
			Cross:       tview.BoxDrawingsLightVerticalAndHorizontal,
		},
		FocusRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsDoubleHorizontal,
			Vertical:    tview.BoxDrawingsDoubleVertical,
			TopLeft:     tview.BoxDrawingsDoubleDownAndRight,
			TopRight:    tview.BoxDrawingsDoubleDownAndLeft,
			BottomLeft:  tview.BoxDrawingsDoubleUpAndRight,
			BottomRight: tview.BoxDrawingsDoubleUpAndLeft,
			LeftT:       tview.BoxDrawingsDoubleVerticalAndRight,
			RightT:      tview.BoxDrawingsDoubleVerticalAndLeft,
			TopT:        tview.BoxDrawingsDoubleDownAndHorizontal,
			BottomT:     tview.BoxDrawingsDoubleUpAndHorizontal,
			Cross:       tview.BoxDrawingsDoubleVerticalAndHorizontal,
		},
	},
	"Double": {
		Name:              "Double",
		FrameAttributes:   tcell.AttrNone,
		OverlayAttributes: tcell.AttrNone,
		NormalRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsDoubleHorizontal,
			Vertical:    tview.BoxDrawingsDoubleVertical,
			TopLeft:     tview.BoxDrawingsDoubleDownAndRight,
			TopRight:    tview.BoxDrawingsDoubleDownAndLeft,
			BottomLeft:  tview.BoxDrawingsDoubleUpAndRight,
			BottomRight: tview.BoxDrawingsDoubleUpAndLeft,
			LeftT:       tview.BoxDrawingsDoubleVerticalAndRight,
			RightT:      tview.BoxDrawingsDoubleVerticalAndLeft,
			TopT:        tview.BoxDrawingsDoubleDownAndHorizontal,
			BottomT:     tview.BoxDrawingsDoubleUpAndHorizontal,
			Cross:       tview.BoxDrawingsDoubleVerticalAndHorizontal,
		},
		FocusRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsHeavyHorizontal,
			Vertical:    tview.BoxDrawingsHeavyVertical,
			TopLeft:     tview.BoxDrawingsHeavyDownAndRight,
			TopRight:    tview.BoxDrawingsHeavyDownAndLeft,
			BottomLeft:  tview.BoxDrawingsHeavyUpAndRight,
			BottomRight: tview.BoxDrawingsHeavyUpAndLeft,
			LeftT:       tview.BoxDrawingsHeavyVerticalAndRight,
			RightT:      tview.BoxDrawingsHeavyVerticalAndLeft,
			TopT:        tview.BoxDrawingsHeavyDownAndHorizontal,
			BottomT:     tview.BoxDrawingsHeavyUpAndHorizontal,
			Cross:       tview.BoxDrawingsHeavyVerticalAndHorizontal,
		},
	},
	"Heavy": {
		Name:              "Heavy",
		FrameAttributes:   tcell.AttrBold,
		OverlayAttributes: tcell.AttrBold,
		NormalRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsHeavyHorizontal,
			Vertical:    tview.BoxDrawingsHeavyVertical,
			TopLeft:     tview.BoxDrawingsHeavyDownAndRight,
			TopRight:    tview.BoxDrawingsHeavyDownAndLeft,
			BottomLeft:  tview.BoxDrawingsHeavyUpAndRight,
			BottomRight: tview.BoxDrawingsHeavyUpAndLeft,
			LeftT:       tview.BoxDrawingsHeavyVerticalAndRight,
			RightT:      tview.BoxDrawingsHeavyVerticalAndLeft,
			TopT:        tview.BoxDrawingsHeavyDownAndHorizontal,
			BottomT:     tview.BoxDrawingsHeavyUpAndHorizontal,
			Cross:       tview.BoxDrawingsHeavyVerticalAndHorizontal,
		},
		FocusRunes: BorderRunes{
			Horizontal:  tview.BoxDrawingsDoubleHorizontal,
			Vertical:    tview.BoxDrawingsDoubleVertical,
			TopLeft:     tview.BoxDrawingsDoubleDownAndRight,
			TopRight:    tview.BoxDrawingsDoubleDownAndLeft,
			BottomLeft:  tview.BoxDrawingsDoubleUpAndRight,
			BottomRight: tview.BoxDrawingsDoubleUpAndLeft,
			LeftT:       tview.BoxDrawingsDoubleVerticalAndRight,
			RightT:      tview.BoxDrawingsDoubleVerticalAndLeft,
			TopT:        tview.BoxDrawingsDoubleDownAndHorizontal,
			BottomT:     tview.BoxDrawingsDoubleUpAndHorizontal,
			Cross:       tview.BoxDrawingsDoubleVerticalAndHorizontal,
		},
	},
}

func colorPaletteByName(name string) ColorPalette {
	if palette, ok := colorPaletteRegistry[name]; ok {
		return palette
	}
	return colorPaletteRegistry["Default"]
}

func borderStyleByName(name string) BorderStyle {
	if style, ok := borderStyleRegistry[name]; ok {
		return style
	}
	return borderStyleRegistry["Classic"]
}

func applyGlobalBorderStyle(style BorderStyle) {
	tview.Borders.Horizontal = style.NormalRunes.Horizontal
	tview.Borders.Vertical = style.NormalRunes.Vertical
	tview.Borders.TopLeft = style.NormalRunes.TopLeft
	tview.Borders.TopRight = style.NormalRunes.TopRight
	tview.Borders.BottomLeft = style.NormalRunes.BottomLeft
	tview.Borders.BottomRight = style.NormalRunes.BottomRight
	tview.Borders.LeftT = style.NormalRunes.LeftT
	tview.Borders.RightT = style.NormalRunes.RightT
	tview.Borders.TopT = style.NormalRunes.TopT
	tview.Borders.BottomT = style.NormalRunes.BottomT
	tview.Borders.Cross = style.NormalRunes.Cross

	tview.Borders.HorizontalFocus = style.FocusRunes.Horizontal
	tview.Borders.VerticalFocus = style.FocusRunes.Vertical
	tview.Borders.TopLeftFocus = style.FocusRunes.TopLeft
	tview.Borders.TopRightFocus = style.FocusRunes.TopRight
	tview.Borders.BottomLeftFocus = style.FocusRunes.BottomLeft
	tview.Borders.BottomRightFocus = style.FocusRunes.BottomRight
}
