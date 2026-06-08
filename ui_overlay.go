package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *app) handleOverlayEscape() {
	switch a.activeOverlay {
	case "radio-add":
		a.closeAddStationOverlay()
		return
	case "theme-colors":
		a.closeThemeColorsOverlay()
		return
	case "theme-borders":
		a.closeBorderStylesOverlay()
		return
	case "themes":
		a.closeThemesOverlay()
		return
	default:
		a.closeConfigurationOverlay()
	}
}

func (a *app) colorOptionLabel(name string) string {
	if a.colorPaletteName == name {
		return name + " (current)"
	}
	return name
}

func (a *app) borderOptionLabel(name string) string {
	if a.borderStyleName == name {
		return name + " (current)"
	}
	return name
}

func borderOptionDetails(name string) string {
	switch name {
	case "Classic":
		return "┌──┐ light lines"
	case "Rounded":
		return "╭──╮ rounded corners"
	case "Double":
		return "╔══╗ double-line frame"
	case "Heavy":
		return "┏━━┓ heavy line emphasis"
	default:
		return ""
	}
}

func (a *app) populateThemesList() {
	if a.themesList == nil {
		return
	}
	a.themesList.Clear()
	a.themesList.AddItem("Colors", "Choose the active color palette", 0, func() { a.openThemeColorsOverlay() })
	a.themesList.AddItem("Border", "Choose border glyph style", 0, func() { a.openBorderStylesOverlay() })
}

func (a *app) populateThemeColorsList() {
	if a.themeColorsList == nil {
		return
	}
	a.themeColorsList.Clear()
	for _, paletteName := range colorPaletteOrder {
		nameCopy := paletteName
		a.themeColorsList.AddItem(a.colorOptionLabel(nameCopy), "", 0, func() {
			a.colorPaletteName = nameCopy
			a.applyTheme()
			a.populateThemeColorsList()
			a.saveConfig()
			a.setStatusTemporary(fmt.Sprintf("[green]Color palette: %s", nameCopy), 2*time.Second)
		})
	}
}

func (a *app) populateBorderStylesList() {
	if a.borderStylesList == nil {
		return
	}
	a.borderStylesList.Clear()
	for _, styleName := range borderStyleOrder {
		nameCopy := styleName
		a.borderStylesList.AddItem(a.borderOptionLabel(nameCopy), borderOptionDetails(nameCopy), 0, func() {
			a.borderStyleName = nameCopy
			a.applyTheme()
			a.populateBorderStylesList()
			a.saveConfig()
			a.setStatusTemporary(fmt.Sprintf("[green]Border style: %s", nameCopy), 2*time.Second)
		})
	}
}

// openConfigurationOverlay displays a centered configuration frame above the main layout.
func (a *app) openConfigurationOverlay() {
	if a.overlayOpen {
		return
	}

	a.previousFocus = a.tv.GetFocus()
	a.configList = tview.NewList().
		ShowSecondaryText(false).
		AddItem("Themes", "", 0, nil)
	a.configList.SetBorder(true).
		SetTitle(" Configuration (Esc to exit) ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)
	a.configList.SetSelectedFunc(func(idx int, _, _ string, _ rune) {
		if idx == 0 {
			a.openThemesOverlay()
		}
	})
	a.configList.SetDoneFunc(func() {
		a.closeConfigurationOverlay()
	})

	// Build a centered modal by surrounding the list with spacer rows and columns.
	centered := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(a.configList, 44, 0, true).
				AddItem(nil, 0, 1, false),
			10, 0, true,
		).
		AddItem(nil, 0, 1, false)

	a.rootPages.AddPage("configuration", centered, true, true)
	a.overlayOpen = true
	a.activeOverlay = "configuration"
	a.tv.SetFocus(a.configList)
	a.applyTheme()
}

func (a *app) openThemesOverlay() {
	if !a.overlayOpen {
		return
	}

	a.themesList = tview.NewList().
		ShowSecondaryText(true)
	a.populateThemesList()
	a.themesList.SetBorder(true).
		SetTitle(" Theme Configuration (Esc to back) ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)
	a.themesList.SetDoneFunc(func() {
		a.closeThemesOverlay()
	})

	centered := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(a.themesList, 44, 0, true).
				AddItem(nil, 0, 1, false),
			10, 0, true,
		).
		AddItem(nil, 0, 1, false)

	a.rootPages.AddPage("themes", centered, true, true)
	a.activeOverlay = "themes"
	a.tv.SetFocus(a.themesList)
	a.applyTheme()
}

func (a *app) closeThemesOverlay() {
	a.rootPages.RemovePage("theme-colors")
	a.rootPages.RemovePage("theme-borders")
	a.rootPages.RemovePage("themes")
	a.themeColorsList = nil
	a.borderStylesList = nil
	a.activeOverlay = "configuration"
	a.applyTheme()
	if a.configList != nil {
		a.tv.SetFocus(a.configList)
	}
}

func (a *app) openThemeColorsOverlay() {
	if a.activeOverlay != "themes" {
		return
	}

	a.themeColorsList = tview.NewList().
		ShowSecondaryText(false)
	a.populateThemeColorsList()
	a.themeColorsList.SetBorder(true).
		SetTitle(" Theme Colors (Esc to back) ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)
	a.themeColorsList.SetDoneFunc(func() {
		a.closeThemeColorsOverlay()
	})

	centered := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(a.themeColorsList, 44, 0, true).
				AddItem(nil, 0, 1, false),
			10, 0, true,
		).
		AddItem(nil, 0, 1, false)

	a.rootPages.AddPage("theme-colors", centered, true, true)
	a.activeOverlay = "theme-colors"
	a.tv.SetFocus(a.themeColorsList)
	a.applyTheme()
}

func (a *app) closeThemeColorsOverlay() {
	a.rootPages.RemovePage("theme-colors")
	a.themeColorsList = nil
	a.activeOverlay = "themes"
	a.applyTheme()
	if a.themesList != nil {
		a.tv.SetFocus(a.themesList)
	}
}

func (a *app) openBorderStylesOverlay() {
	if a.activeOverlay != "themes" {
		return
	}

	a.borderStylesList = tview.NewList().
		ShowSecondaryText(true)
	a.populateBorderStylesList()
	a.borderStylesList.SetBorder(true).
		SetTitle(" Theme Border Styles (Esc to back) ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)
	a.borderStylesList.SetDoneFunc(func() {
		a.closeBorderStylesOverlay()
	})

	centered := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(a.borderStylesList, 52, 0, true).
				AddItem(nil, 0, 1, false),
			10, 0, true,
		).
		AddItem(nil, 0, 1, false)

	a.rootPages.AddPage("theme-borders", centered, true, true)
	a.activeOverlay = "theme-borders"
	a.tv.SetFocus(a.borderStylesList)
	a.applyTheme()
}

func (a *app) closeBorderStylesOverlay() {
	a.rootPages.RemovePage("theme-borders")
	a.borderStylesList = nil
	a.activeOverlay = "themes"
	a.applyTheme()
	if a.themesList != nil {
		a.tv.SetFocus(a.themesList)
	}
}

// closeConfigurationOverlay removes overlay pages and restores the prior focus target.
func (a *app) closeConfigurationOverlay() {
	if !a.overlayOpen && a.activeOverlay == "" {
		return
	}
	a.rootPages.RemovePage("theme-colors")
	a.rootPages.RemovePage("theme-borders")
	a.rootPages.RemovePage("themes")
	a.rootPages.RemovePage("configuration")
	a.overlayOpen = false
	a.activeOverlay = ""
	a.configList = nil
	a.themesList = nil
	a.themeColorsList = nil
	a.borderStylesList = nil
	a.applyTheme()
	if a.previousFocus != nil {
		a.tv.SetFocus(a.previousFocus)
		return
	}
	a.tv.SetFocus(a.table)
}
