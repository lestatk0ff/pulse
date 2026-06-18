package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *app) toggleRadioMode() {
	if a.filterActive {
		a.exitFilter()
	}
	a.radioMode = !a.radioMode
	if a.radioMode {
		a.enterRadioMode()
	} else {
		a.exitRadioMode()
	}
}

func (a *app) enterRadioMode() {
	a.allStations = buildStationList()
	a.stations = a.allStations
	a.tableFrame.SetTitle(" Radio Stations — Enter to play · A to add · D to delete custom ")
	a.updateHotkeys()
	a.populateRadioTable()
	a.tv.SetFocus(a.table)
}

func (a *app) exitRadioMode() {
	// Restore file table headers before repopulating.
	headers := []string{"File Name", "Size", "Format", "Bitrate"}
	for col, h := range headers {
		a.table.SetCell(0, col,
			tview.NewTableCell(" "+h+" ").
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false).
				SetBackgroundColor(tcell.ColorDarkBlue),
		)
	}
	a.applyTableHeaderTheme(a.colorPaletteName)
	a.tableFrame.SetTitle(fmt.Sprintf(" Audio Files — %s (↑↓ scroll) ", a.dir))
	a.updateHotkeys()
	a.populateTable()
	if a.selectedFile != nil {
		a.probeAndShowDetails(a.selectedFile)
	} else {
		a.detailsView.SetText("[grey]Select a file to see details")
	}
	a.tv.SetFocus(a.table)
}

func (a *app) updateHotkeys() {
	line := func(key, desc string) string {
		return fmt.Sprintf("[lightyellow]%-6s[white] [gray]- %s", key, desc)
	}
	var lines []string
	if a.radioMode {
		tabDesc := "Switch panel"
		if !a.actionsVisible {
			tabDesc = "Switch panel (Actions hidden)"
		}
		lines = []string{
			line("Tab", tabDesc),
			line("Ctrl+P", "Show/hide Actions panel"),
			line("C", "Configuration"),
			line("↑↓", "Navigate"),
			line("Enter", "Play station"),
			line("A", "Add station"),
			line("D", "Delete custom"),
			line("S", "Stop"),
			line("M", "Mute/unmute"),
			line("+/-", "Volume up/down"),
			line("Ctrl+F", "Filter"),
			line("R", "File browser"),
			line("Esc", "Quit"),
		}
	} else {
		tabDesc := "Switch panel"
		if !a.actionsVisible {
			tabDesc = "Switch panel (Actions hidden)"
		}
		lines = []string{
			line("Tab", tabDesc),
			line("Ctrl+P", "Show/hide Actions panel"),
			line("C", "Configuration"),
			line("↑↓", "Navigate"),
			line("Enter", "Play"),
			line("Z", "Shuffle list"),
			line("S", "Stop"),
			line("M", "Mute/unmute"),
			line("+/-", "Volume up/down"),
			line("Ctrl+F", "Filter"),
			line("R", "Radio"),
			line("F5", "Refresh"),
			line("Esc", "Quit"),
		}
	}
	a.hotkeysView.SetText(strings.Join(lines, "\n"))
}

func (a *app) populateRadioTable() {
	headers := []string{"Station Name", "Genre", "Country", "Bitrate"}
	for col, h := range headers {
		a.table.SetCell(0, col,
			tview.NewTableCell(" "+h+" ").
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false).
				SetBackgroundColor(tcell.ColorDarkBlue),
		)
	}
	a.applyTableHeaderTheme(a.colorPaletteName)

	for a.table.GetRowCount() > 1 {
		a.table.RemoveRow(1)
	}

	for row, s := range a.stations {
		nameCell := tview.NewTableCell(" " + s.Name)
		if s.Custom {
			nameCell.SetTextColor(tcell.ColorAqua)
		}
		a.table.SetCell(row+1, 0, nameCell)
		a.table.SetCell(row+1, 1, tview.NewTableCell(" "+s.Genre))
		a.table.SetCell(row+1, 2, tview.NewTableCell(" "+s.Country))
		a.table.SetCell(row+1, 3, tview.NewTableCell(" "+s.Bitrate))
	}

	if len(a.stations) > 0 {
		a.table.Select(1, 0)
		a.selectedStation = a.stations[0]
		a.showRadioDetails(a.selectedStation)
	} else {
		a.selectedStation = nil
		a.detailsView.SetText("[grey]No stations")
	}
}

func (a *app) showRadioDetails(s *RadioStation) {
	if s == nil {
		a.detailsView.SetText("")
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "[yellow]Name:[white] %s\n", s.Name)
	fmt.Fprintf(&sb, "[yellow]Genre:[white] %s\n", s.Genre)
	fmt.Fprintf(&sb, "[yellow]Country:[white] %s\n", s.Country)
	fmt.Fprintf(&sb, "[yellow]Bitrate:[white] %s\n", s.Bitrate)
	fmt.Fprintf(&sb, "\n[yellow]URL:[white]\n%s\n", s.URL)
	if s.Custom {
		fmt.Fprintf(&sb, "\n[aqua]Custom station[white]")
	}
	a.detailsView.SetText(sb.String())
}

func (a *app) applyRadioFilter(pattern string) {
	if pattern == "" {
		a.stations = a.allStations
		a.populateRadioTable()
		return
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		a.stations = nil
		a.populateRadioTable()
		a.detailsView.SetText("[red]Invalid regexp")
		return
	}
	filtered := a.radioFilterBuf[:0]
	for _, s := range a.allStations {
		if re.MatchString(s.Name) || re.MatchString(s.Genre) || re.MatchString(s.Country) {
			filtered = append(filtered, s)
		}
	}
	a.radioFilterBuf = filtered
	a.stations = filtered
	a.populateRadioTable()
}

// openAddStationOverlay shows a form for entering a new custom station.
func (a *app) openAddStationOverlay() {
	if a.overlayOpen {
		return
	}

	form := tview.NewForm()
	form.AddInputField("Name    ", "", 30, nil, nil)
	form.AddInputField("URL     ", "", 46, nil, nil)
	form.AddInputField("Genre   ", "", 20, nil, nil)
	form.AddInputField("Country ", "", 20, nil, nil)

	form.AddButton("Add", func() {
		name := strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText())
		url := strings.TrimSpace(form.GetFormItem(1).(*tview.InputField).GetText())
		genre := strings.TrimSpace(form.GetFormItem(2).(*tview.InputField).GetText())
		country := strings.TrimSpace(form.GetFormItem(3).(*tview.InputField).GetText())

		if name == "" || url == "" {
			a.setStatusTemporary("[red]Name and URL are required", 3*time.Second)
			return
		}

		station := &RadioStation{
			Name: name, URL: url, Genre: genre, Country: country, Bitrate: "?", Custom: true,
		}
		a.allStations = append(a.allStations, station)
		if err := saveCustomStations(a.allStations); err != nil {
			a.setStatusTemporary(fmt.Sprintf("[red]Save error: %v", err), 3*time.Second)
		}
		a.stations = a.allStations
		a.populateRadioTable()
		a.closeAddStationOverlay()
		a.setStatusTemporary(fmt.Sprintf("[green]Added:[white] %s", name), 3*time.Second)
	})

	form.AddButton("Cancel", func() {
		a.closeAddStationOverlay()
	})

	form.SetBorder(true).
		SetTitle(" Add Radio Station ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)

	centered := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(form, 64, 0, true).
				AddItem(nil, 0, 1, false),
			14, 0, true,
		).
		AddItem(nil, 0, 1, false)

	a.previousFocus = a.tv.GetFocus()
	a.rootPages.AddPage("radio-add", centered, true, true)
	a.overlayOpen = true
	a.activeOverlay = "radio-add"
	a.tv.SetFocus(form)
	a.applyTheme()
}

func (a *app) closeAddStationOverlay() {
	if !a.overlayOpen || a.activeOverlay != "radio-add" {
		return
	}
	a.rootPages.RemovePage("radio-add")
	a.overlayOpen = false
	a.activeOverlay = ""
	a.applyTheme()
	if a.previousFocus != nil {
		a.tv.SetFocus(a.previousFocus)
		return
	}
	a.tv.SetFocus(a.table)
}

func (a *app) deleteSelectedStation() {
	if a.selectedStation == nil {
		return
	}
	if !a.selectedStation.Custom {
		a.setStatusTemporary("[red]Built-in stations cannot be deleted", 3*time.Second)
		return
	}
	name := a.selectedStation.Name
	newAll := a.allStations[:0]
	for _, s := range a.allStations {
		if s != a.selectedStation {
			newAll = append(newAll, s)
		}
	}
	a.allStations = newAll
	if err := saveCustomStations(a.allStations); err != nil {
		a.setStatusTemporary(fmt.Sprintf("[red]Save error: %v", err), 3*time.Second)
		return
	}
	a.stations = a.allStations
	a.populateRadioTable()
	a.setStatusTemporary(fmt.Sprintf("[green]Deleted:[white] %s", name), 3*time.Second)
}
