package daemon

import (
	"gofi/pkg/desktop"
	"gofi/pkg/log"
)

// GofiAutoCloser handles the logic for automatically closing the gofi window
// when it loses focus. Uses the WindowManager directly.
type GofiAutoCloser struct {
	wm               desktop.WindowManager
	lastActiveWindow int  // Stores the ID of the window previously considered active
	wasGofiActive    bool // Flag indicating if the lastActiveWindow matched our criteria
}

// NewGofiAutoCloser creates a new GofiAutoCloser instance.
// Assumes wm is non-nil.
func NewGofiAutoCloser(wm desktop.WindowManager) *GofiAutoCloser {
	return &GofiAutoCloser{
		wm:               wm,
		lastActiveWindow: 0, // Initialize with no known active window
		wasGofiActive:    false,
	}
}

// CheckFocusAndClose checks the current active window against the previously
// known active window. If the previously active window matched the "gofi" criteria
// and the current one is different, it attempts to close the previous one.
func (gac *GofiAutoCloser) CheckFocusAndClose() {
	currentActiveID := gac.wm.ActiveWindowID()
	if currentActiveID == gac.lastActiveWindow {
		return
	}

	log.Debug("Current active window ID: %d", currentActiveID)

	// If the active window hasn't changed, or there's no current active window, do nothing.
	if currentActiveID == gac.lastActiveWindow || currentActiveID == 0 {
		// Update state in case the *content* of lastActiveWindow changed,
		// but don't trigger close logic if the ID is the same.
		gac.updateState(currentActiveID)
		return
	}

	// The active window ID has changed. Check if the *previous* one was gofi.
	if gac.wasGofiActive {
		log.Debug("Gofi window (%d) lost focus to window %d", gac.lastActiveWindow, currentActiveID)
		err := gac.wm.CloseWindow(gac.lastActiveWindow)
		if err != nil {
			log.Error("Failed to send close request to gofi window %d: %v", gac.lastActiveWindow, err)
		}
	}

	// Update state for the *next* check
	gac.updateState(currentActiveID)
}

// updateState updates the internal tracking of the last active window and
// whether it matched the gofi criteria.
func (gac *GofiAutoCloser) updateState(activeID int) {
	gac.lastActiveWindow = activeID
	if activeID == 0 {
		gac.wasGofiActive = false // No active window can't be gofi
		return
	}

	// Check if the *new* active window matches the criteria
	title := gac.wm.WindowTitle(activeID)
	_, className := gac.wm.WindowClass(activeID) // We only need class name

	isGofiTitle := title == "gofi"
	isStClass := className == "st" || className == "st-256color"

	gac.wasGofiActive = isGofiTitle && isStClass
	if gac.wasGofiActive {
		log.Debug("Gofi window (%d) became active.", activeID)
	}
}
