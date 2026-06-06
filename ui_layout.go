package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// newTUIApp creates and wires up the app struct but does not start the event loop.
func newTUIApp(dir string, files []*AudioFile) *app {
	playerBinary, playerBaseArgs := findPlayer()

	a := &app{
		tv:             tview.NewApplication(),
		files:          files,
		allFiles:       files,
		dir:            dir,
		playerBinary:   playerBinary,
		playerBaseArgs: playerBaseArgs,
		volume:         80,
		themeName:      "Default",
	}
	a.build()
	return a
}

// build constructs all widgets and assembles the final layout.
func (a *app) build() {
	a.buildTable()
	a.buildDetails()
	a.buildActions()
	a.buildLayout()
}

// buildDetails creates the read-only text view that shows metadata for the selected file.
func (a *app) buildDetails() {
	a.detailsView = tview.NewTextView().
		SetDynamicColors(true).
		SetText("[grey]Select a file to see details")
}

// buildTable creates the file list table with fixed header row and row-selection callbacks.
func (a *app) buildTable() {
	a.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

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
	a.applyTableHeaderTheme(a.themeName)

	a.populateTable()

	// Keep detail pane synchronized with the row cursor.
	a.table.SetSelectionChangedFunc(func(row, _ int) {
		if row >= 1 && row <= len(a.files) {
			a.selectedFile = a.files[row-1]
			a.probeAndShowDetails(a.selectedFile)
		}
	})
}

// applyTableHeaderTheme updates the table header text/background to match the active theme.
func (a *app) applyTableHeaderTheme(name string) {
	if a.table == nil {
		return
	}
	theme := themeByName(name)

	for col := 0; col < 4; col++ {
		cell := a.table.GetCell(0, col)
		if cell == nil {
			continue
		}
		cell.SetTextColor(theme.TableHeaderTextColor).
			SetBackgroundColor(theme.TableHeaderBgColor).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
	}
}

// populateTable clears all data rows (keeping the header) and re-fills the table from a.files.
func (a *app) populateTable() {
	for a.table.GetRowCount() > 1 {
		a.table.RemoveRow(1)
	}

	for row, f := range a.files {
		bitrateStr := "N/A"
		if f.Bitrate > 0 {
			bitrateStr = fmt.Sprintf("%d kbps", f.Bitrate)
		}

		a.table.SetCell(row+1, 0, tview.NewTableCell(" "+f.RelPath))
		a.table.SetCell(row+1, 1, tview.NewTableCell(" "+fmtSize(f.Size)))
		a.table.SetCell(row+1, 2, tview.NewTableCell(" "+f.Format))
		a.table.SetCell(row+1, 3, tview.NewTableCell(" "+bitrateStr))
	}

	if len(a.files) > 0 {
		a.table.Select(1, 0)
		a.selectedFile = a.files[0]
	} else {
		a.selectedFile = nil
		a.detailsView.SetText("[grey]No matches")
	}
}

// buildActions creates the status bar, search bar, and the action list shown in the bottom panel.
func (a *app) buildActions() {
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("")

	a.playingBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	hotkeyLine := func(key, desc string) string {
		return fmt.Sprintf("[lightyellow]%-6s[white] [gray]- %s", key, desc)
	}

	a.hotkeysView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(strings.Join([]string{
			hotkeyLine("Tab", "Switch panel"),
			hotkeyLine("C", "Open Configuration"),
			hotkeyLine("↑↓", "Navigate"),
			hotkeyLine("Enter", "Play"),
			hotkeyLine("Z", "Shuffle list"),
			hotkeyLine("S", "Stop"),
			hotkeyLine("M", "Mute/unmute"),
			hotkeyLine("+/-", "Volume up/down"),
			hotkeyLine("Ctrl+F", "Filter"),
			hotkeyLine("R", "Refresh"),
			hotkeyLine("Esc", "Quit"),
		}, "\n"))

	a.searchBar = tview.NewInputField().
		SetLabel("Filter (regexp): ").
		SetLabelColor(tcell.ColorYellow).
		SetFieldBackgroundColor(tcell.ColorDarkBlue).
		SetChangedFunc(func(text string) {
			a.scheduleFilter(text)
		}).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				a.confirmFilter()
			}
		})

	a.actionList = tview.NewList().
		AddItem("  [1]  Convert to 192 kbps", "New file: <name>_192k.<ext> (same dir)", '1', nil).
		AddItem("  [2]  Convert to OGG format", "New file: <name>.ogg (same dir)", '2', nil).
		AddItem("  [z]  Shuffle current list", "Randomize the order shown in the file table", 'z', nil).
		AddItem("  [r]  Refresh file list", "Re-scan the directory", 'r', nil).
		AddItem("  [q]  Quit", "Exit mp3lister", 'q', func() { a.tv.Stop() })

	a.actionList.SetSelectedFunc(func(idx int, _, _ string, _ rune) {
		a.runAction(idx)
	})
}

// buildLayout assembles all panels and registers global key handling.
func (a *app) buildLayout() {
	a.detailsFrame = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.detailsView, 0, 1, false)
	a.detailsFrame.SetBorder(true).
		SetTitle(" Details ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)

	a.tableFrame = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.table, 0, 1, true)
	a.tableFrame.SetBorder(true).
		SetTitle(fmt.Sprintf(" Audio Files — %s (↑↓ scroll) ", a.dir)).
		SetTitleColor(tcell.ColorGreen).
		SetBorderColor(tcell.ColorGreen)

	topPanel := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.detailsFrame, 32, 0, false).
		AddItem(a.tableFrame, 0, 1, true)

	a.hotkeysFrame = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.hotkeysView, 0, 1, false)
	a.hotkeysFrame.SetBorder(true).
		SetTitle(" Hotkeys ").
		SetTitleColor(tcell.ColorAqua).
		SetBorderColor(tcell.ColorAqua)

	a.actionsFrame = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.actionList, 0, 1, false)
	a.actionsFrame.SetBorder(true).
		SetTitle(" Actions (Tab to focus, Enter to run) ").
		SetTitleColor(tcell.ColorAqua).
		SetBorderColor(tcell.ColorAqua)

	bottomRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.actionsFrame, 0, 1, false).
		AddItem(a.hotkeysFrame, 32, 0, false)

	statusRow := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.playingBar, 1, 0, false).
		AddItem(a.statusBar, 1, 0, false)

	a.statusPages = tview.NewPages()
	a.statusPages.AddPage("normal", statusRow, true, true)
	a.statusPages.AddPage("search", a.searchBar, true, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topPanel, 0, 1, true).
		AddItem(bottomRow, 0, 1, false).
		AddItem(a.statusPages, 2, 0, false)

	a.rootPages = tview.NewPages()
	a.rootPages.AddPage("main", root, true, true)
	a.applyFrameTheme(a.themeName)

	a.tv.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyCtrlF:
			if !a.filterActive {
				a.filterActive = true
				a.statusPages.SwitchToPage("search")
				a.tv.SetFocus(a.searchBar)
			}
			return nil
		case tcell.KeyTab:
			if a.filterActive {
				return ev
			}
			if a.table.HasFocus() {
				a.tv.SetFocus(a.actionList)
			} else {
				a.tv.SetFocus(a.table)
			}
			return nil
		case tcell.KeyEscape:
			if a.overlayOpen {
				a.handleOverlayEscape()
				return nil
			}
			if a.filterActive {
				a.exitFilter()
				return nil
			}
			a.tv.Stop()
			return nil
		case tcell.KeyEnter:
			if a.table.HasFocus() && a.selectedFile != nil {
				a.playFile(a.selectedFile)
				return nil
			}
		case tcell.KeyRune:
			if a.filterActive {
				return ev
			}
			switch ev.Rune() {
			case 'c', 'C':
				a.openConfigurationOverlay()
				return nil
			case '+':
				a.adjustVolume(5)
				return nil
			case '-':
				a.adjustVolume(-5)
				return nil
			case 'm', 'M':
				a.toggleMute()
				return nil
			case '1':
				a.runAction(0)
				return nil
			case '2':
				a.runAction(1)
				return nil
			case 's', 'S':
				a.stopPlayback()
				return nil
			case 'z', 'Z':
				a.shuffleCurrentList()
				return nil
			case 'r', 'R':
				a.refresh()
				return nil
			case 'q', 'Q':
				a.tv.Stop()
				return nil
			}
		}
		return ev
	})

	a.tv.SetRoot(a.rootPages, true).SetFocus(a.table)
}

// applyFrameTheme updates frame border/title colors using one of the named presets.
func (a *app) applyFrameTheme(name string) {
	theme := themeByName(name)
	a.themeName = theme.Name
	a.applyTableHeaderTheme(theme.Name)

	a.detailsFrame.SetTitleColor(theme.DetailsFrameColor).SetBorderColor(theme.DetailsFrameColor)
	a.tableFrame.SetTitleColor(theme.TableFrameColor).SetBorderColor(theme.TableFrameColor)
	a.actionsFrame.SetTitleColor(theme.ActionsFrameColor).SetBorderColor(theme.ActionsFrameColor)
	a.hotkeysFrame.SetTitleColor(theme.HotkeysFrameColor).SetBorderColor(theme.HotkeysFrameColor)
	if a.configList != nil {
		a.configList.SetTitleColor(theme.OverlayFrameColor).SetBorderColor(theme.OverlayFrameColor)
	}
	if a.themesList != nil {
		a.themesList.SetTitleColor(theme.OverlayFrameColor).SetBorderColor(theme.OverlayFrameColor)
	}
}

// run starts the tview event loop and blocks until the app exits.
func (a *app) run() error {
	defer a.stopPlayback()
	return a.tv.Run()
}
