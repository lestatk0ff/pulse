package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const filterDebounceDelay = 140 * time.Millisecond

// setStatus updates the status bar from the UI goroutine.
func (a *app) setStatus(msg string) {
	a.statusNonce.Add(1)
	a.statusBar.SetText(msg)
}

// setStatusAsync updates the status bar from background goroutines safely.
func (a *app) setStatusAsync(msg string) {
	a.tv.QueueUpdateDraw(func() {
		a.statusNonce.Add(1)
		a.statusBar.SetText(msg)
	})
}

// setStatusTemporary shows a message and auto-clears it unless newer status text was set.
func (a *app) setStatusTemporary(msg string, d time.Duration) {
	a.setStatus(msg)
	expected := a.statusNonce.Load()
	go func() {
		time.Sleep(d)
		a.tv.QueueUpdateDraw(func() {
			if a.statusNonce.Load() != expected {
				return
			}
			if strings.HasPrefix(a.statusBar.GetText(false), "[yellow]Filter:[white]") {
				return
			}
			a.statusBar.SetText("")
		})
	}()
}

// applyFilter filters the current list (files or radio stations) and refreshes the table.
func (a *app) applyFilter(pattern string) {
	if a.radioMode {
		a.applyRadioFilter(pattern)
		return
	}
	if pattern == "" {
		a.files = a.allFiles
		a.populateTable()
		return
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		a.files = nil
		a.populateTable()
		a.detailsView.SetText("[red]Invalid regexp")
		return
	}
	// Reuse a dedicated backing slice so live filtering allocates less.
	filtered := a.filteredBuf[:0]
	for _, f := range a.allFiles {
		if re.MatchString(f.RelPath) {
			filtered = append(filtered, f)
		}
	}
	a.filteredBuf = filtered
	a.files = filtered
	a.populateTable()
}

// scheduleFilter debounces high-frequency typing events before applying a regexp filter.
func (a *app) scheduleFilter(pattern string) {
	a.stopFilterDebounce()
	a.filterDebounce = time.AfterFunc(filterDebounceDelay, func() {
		a.tv.QueueUpdateDraw(func() {
			// Ignore stale timer callbacks that no longer match the current input.
			if a.searchBar.GetText() != pattern {
				return
			}
			a.applyFilter(pattern)
		})
	})
}

// stopFilterDebounce cancels any pending delayed filter apply.
func (a *app) stopFilterDebounce() {
	if a.filterDebounce != nil {
		a.filterDebounce.Stop()
		a.filterDebounce = nil
	}
}

// confirmFilter hides the input while keeping the currently filtered result set.
func (a *app) confirmFilter() {
	a.stopFilterDebounce()
	a.applyFilter(a.searchBar.GetText())
	a.filterActive = false
	a.statusPages.SwitchToPage("normal")
	a.tv.SetFocus(a.table)
	if text := a.searchBar.GetText(); text != "" {
		total, shown := len(a.allFiles), len(a.files)
		if a.radioMode {
			total, shown = len(a.allStations), len(a.stations)
		}
		a.statusBar.SetText(fmt.Sprintf(
			"[yellow]Filter:[white] %s  [grey](%d/%d)  Ctrl+F to edit, Esc to clear",
			text, shown, total,
		))
	} else {
		a.clearFilterStatusIfShown()
	}
}

// exitFilter clears filter input/results and restores the full list view.
func (a *app) exitFilter() {
	a.stopFilterDebounce()
	a.filterActive = false
	a.searchBar.SetText("")
	if a.radioMode {
		a.stations = a.allStations
		a.populateRadioTable()
	} else {
		a.files = a.allFiles
		a.populateTable()
	}
	a.clearFilterStatusIfShown()
	a.statusPages.SwitchToPage("normal")
	a.tv.SetFocus(a.table)
}

// clearFilterStatusIfShown removes the filter hint while preserving other status messages.
func (a *app) clearFilterStatusIfShown() {
	if strings.HasPrefix(a.statusBar.GetText(false), "[yellow]Filter:[white]") {
		a.statusBar.SetText("")
	}
}
