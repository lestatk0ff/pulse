package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *app) handleOverlayEscape() {
	if a.activeOverlay == "themes" {
		a.closeThemesOverlay()
		return
	}
	a.closeConfigurationOverlay()
}

func (a *app) themeOptionLabel(name string) string {
	if a.themeName == name {
		return name + " (current)"
	}
	return name
}

func (a *app) populateThemesList() {
	if a.themesList == nil {
		return
	}
	a.themesList.Clear()
	for _, theme := range themeOrder {
		themeCopy := theme
		a.themesList.AddItem(a.themeOptionLabel(themeCopy), "", 0, func() {
			a.applyFrameTheme(themeCopy)
			a.populateThemesList()
			a.setStatusTemporary(fmt.Sprintf("[green]Frame theme: %s", themeCopy), 2*time.Second)
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
		AddItem("Themes", "", 0, nil).
		AddItem("MOCK", "", 0, nil)
	a.configList.SetBorder(true).
		SetTitle(" Configuration (Esc to exit) ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)
	a.configList.SetSelectedFunc(func(idx int, _, _ string, _ rune) {
		switch idx {
		case 0:
			a.openThemesOverlay()
		case 1:
			a.setStatusTemporary("[yellow]MOCK has no action yet", 2*time.Second)
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
	a.applyFrameTheme(a.themeName)
}

func (a *app) openThemesOverlay() {
	if !a.overlayOpen {
		return
	}

	a.themesList = tview.NewList().
		ShowSecondaryText(false)
	a.populateThemesList()
	a.themesList.SetBorder(true).
		SetTitle(" Themes - Choose frame color (Esc to back) ").
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
	a.applyFrameTheme(a.themeName)
}

func (a *app) closeThemesOverlay() {
	a.rootPages.RemovePage("themes")
	a.activeOverlay = "configuration"
	if a.configList != nil {
		a.tv.SetFocus(a.configList)
	}
}

// closeConfigurationOverlay removes overlay pages and restores the prior focus target.
func (a *app) closeConfigurationOverlay() {
	if !a.overlayOpen && a.activeOverlay == "" {
		return
	}
	a.rootPages.RemovePage("themes")
	a.rootPages.RemovePage("configuration")
	a.overlayOpen = false
	a.activeOverlay = ""
	a.configList = nil
	a.themesList = nil
	if a.previousFocus != nil {
		a.tv.SetFocus(a.previousFocus)
		return
	}
	a.tv.SetFocus(a.table)
}
