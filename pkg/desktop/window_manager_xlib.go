package desktop

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"

	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// XLibWindowManager provides X11 window management adhering to clean code principles.
type XLibWindowManager struct {
	display *xgb.Conn
	// Cache atoms for efficiency
	atomCache map[string]xproto.Atom
	atomMutex sync.RWMutex
}

var (
	// Singleton instance
	instance *XLibWindowManager
	// Mutex for singleton
	mutex sync.Mutex
)

// NewXLibWindowManagerClean establishes a connection to the X server and creates a new manager.
// Returns a new manager instance or an error if the connection fails.
func NewXLibWindowManager() (*XLibWindowManager, error) {
	display, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %w", err)
	}
	return &XLibWindowManager{
		display:   display,
		atomCache: make(map[string]xproto.Atom),
	}, nil
}

// Instance returns the singleton instance of the window manager
// Returns:
//
//	*XLibWindowManager: The singleton instance
func Instance() *XLibWindowManager {
	mutex.Lock()
	defer mutex.Unlock()
	if instance == nil {
		var err error
		instance, err = NewXLibWindowManager()
		if err != nil {
			log.Error("Failed to create window manager: %v", err)
			return nil
		}
	}
	return instance
}

// InitEvents subscribes to necessary X server events on the root window.
// It requests notifications for property changes and substructure modifications.
// Returns true on success, false on failure.
func (wm *XLibWindowManager) InitEvents() bool {
	root := xproto.Setup(wm.display).DefaultScreen(wm.display).Root
	mask := xproto.EventMaskPropertyChange | xproto.EventMaskSubstructureNotify
	cookie := xproto.ChangeWindowAttributesChecked(wm.display, root, xproto.CwEventMask, []uint32{uint32(mask)})

	if err := cookie.Check(); err != nil {
		log.Error("Failed to set event mask on root window: %v", err)
		return false
	}
	log.Debug("Successfully set event mask on root window")
	return true
}

// AwaitEvent waits for the next relevant X event or context cancellation.
// It returns a string identifier for the event type (e.g., "MapNotify", "PropertyNotify:_NET_ACTIVE_WINDOW")
// or an empty string if the context is cancelled or an error occurs.
func (wm *XLibWindowManager) AwaitEvent(ctx context.Context) string {
	eventChan := make(chan interface{}, 1)
	go wm.waitForXEvent(eventChan) // Use method receiver

	select {
	case result := <-eventChan:
		return wm.processXEventResult(result)
	case <-ctx.Done():
		return wm.handleAwaitCancellation(ctx)
	}
}

// waitForXEvent listens for the next event from the X server.
// It sends the event or an error through the provided channel.
func (wm *XLibWindowManager) waitForXEvent(eventChan chan<- interface{}) {
	event, err := wm.display.WaitForEvent()
	if err != nil {
		eventChan <- err // Send error (including potential EOF)
		return
	}
	if event != nil {
		eventChan <- event // Send valid event
	}
	// If event is nil but err is nil, it's an unusual case, maybe ignore?
	// For now, sending nothing effectively stalls the receiver, which might be okay
	// if it only happens on graceful shutdown scenarios not producing EOF.
}

// processXEventResult analyzes the result from waitForXEvent.
// It returns the corresponding event string identifier or an empty string on error/EOF.
func (wm *XLibWindowManager) processXEventResult(result interface{}) string {
	switch event := result.(type) {
	case error:
		if event != io.EOF {
			log.Error("Error received from X server connection: %v", event)
		} else {
			log.Debug("EOF received from X server connection (closed cleanly)")
		}
		return "" // Error or clean close
	default:
		return wm.formatEventString(event) // Format valid event
	}
}

// formatEventString creates a structured string identifier for a given X event.
func (wm *XLibWindowManager) formatEventString(event interface{}) string {
	var eventName string
	switch ev := event.(type) {
	case *xproto.PropertyNotifyEvent:
		atomName := wm.getAtomNameCached(ev.Atom) // Use cached atom name lookup
		if atomName == "_NET_ACTIVE_WINDOW" {
			eventName = "PropertyNotify:_NET_ACTIVE_WINDOW"
		} else if atomName != "" {
			eventName = fmt.Sprintf("PropertyNotify:%s", atomName) // More specific
		} else {
			eventName = "PropertyNotify:Other" // Fallback
		}
	default:
		eventName = wm.getSimplifiedTypeName(ev) // Generic fallback
	}
	// log.Debug("Formatted X event: %s", eventName) // Optional: Debug log here
	return eventName
}

// getSimplifiedTypeName extracts the base type name from a potentially prefixed type.
func (wm *XLibWindowManager) getSimplifiedTypeName(v interface{}) string {
	typeName := fmt.Sprintf("%T", v)
	if parts := strings.Split(typeName, "."); len(parts) > 1 {
		return parts[len(parts)-1] // Strip package prefix
	}
	return typeName
}

// handleAwaitCancellation handles the cancellation of AwaitEvent via context.
// It logs the cancellation and attempts to close the display connection.
// Returns an empty string.
func (wm *XLibWindowManager) handleAwaitCancellation(ctx context.Context) string {
	log.Debug("AwaitEvent aborted by context: %v", ctx.Err())
	// Closing the display connection is crucial to unblock waitForXEvent.
	// Assuming this manager instance owns the connection.
	if wm.display != nil {
		wm.display.Close()
		wm.display = nil // Prevent double close
	}
	return ""
}

// ActiveWindowID queries the X server for the ID of the currently active window.
// It uses the _NET_ACTIVE_WINDOW property on the root window.
// Returns the window ID, or 0 if none is found or an error occurs.
func (wm *XLibWindowManager) ActiveWindowID() int {
	root := wm.getRootWindow()
	if root == 0 {
		return 0
	}

	activeWinAtom := wm.getAtomCached("_NET_ACTIVE_WINDOW")
	if activeWinAtom == 0 {
		return 0
	}

	prop, err := xproto.GetProperty(
		wm.display, false, root, activeWinAtom,
		xproto.GetPropertyTypeAny, 0, 1, // Request 1 item (4 bytes for window ID)
	).Reply()

	if err != nil {
		log.Error("Failed to get _NET_ACTIVE_WINDOW property: %v", err)
		return 0
	}

	if !wm.isValidPropertyReply(prop, 32, 1) { // Expect 1 item of 32-bit format
		log.Warn("Invalid reply format for _NET_ACTIVE_WINDOW")
		return 0
	}

	windowID := int(binary.LittleEndian.Uint32(prop.Value))
	if !wm.isValidWindow(xproto.Window(windowID)) {
		// log.Debug("Active window ID %d reported but is not valid", windowID)
		return 0 // Don't return invalid IDs
	}
	return windowID
}

// StackingList retrieves the list of client windows managed by the X server.
// It uses the _NET_CLIENT_LIST property on the root window.
// Returns a slice of Window pointers, or nil on error.
func (wm *XLibWindowManager) StackingList() []*shared.Window {
	root := wm.getRootWindow()
	if root == 0 {
		return nil
	}

	clientListAtom := wm.getAtomCached("_NET_CLIENT_LIST")
	if clientListAtom == 0 {
		return nil
	}

	prop, err := xproto.GetProperty(
		wm.display, false, root, clientListAtom,
		xproto.AtomWindow, 0, 1024, // Request up to 1024 window IDs
	).Reply()

	if err != nil {
		log.Error("Failed to get _NET_CLIENT_LIST property: %v", err)
		return nil
	}

	// Reply.Value contains a list of 32-bit (4-byte) window IDs
	numWindows := int(prop.ValueLen)
	windows := make([]*shared.Window, 0, numWindows)
	valueBytes := prop.Value

	for i := 0; i < numWindows; i++ {
		start := i * 4
		if start+4 > len(valueBytes) {
			log.Warn("Insufficient data in _NET_CLIENT_LIST reply")
			break // Avoid panic
		}
		windowID := xproto.Window(binary.LittleEndian.Uint32(valueBytes[start : start+4]))
		if windowID == 0 {
			continue // Skip null window IDs
		}

		windowInfo := wm.createWindowInfo(windowID)
		if windowInfo != nil {
			windows = append(windows, windowInfo)
		}
	}
	return windows
}

// createWindowInfo gathers details (ID, Title, Type, Class) for a given window ID.
// Returns a Window struct pointer, or nil if essential info is missing or window is invalid.
func (wm *XLibWindowManager) createWindowInfo(windowID xproto.Window) *shared.Window {
	if !wm.isValidWindow(windowID) {
		// log.Debug("Skipping invalid window ID %d during info creation", windowID)
		return nil
	}

	title := wm.getWindowName(windowID)
	// We might decide later that windows without titles are not interesting
	// if title == "" {
	// 	log.Debug("Window %d has no usable title, skipping", windowID)
	// 	return nil
	// }

	windowType := wm.getWindowType(windowID)
	instance, class := wm.getWindowClass(windowID)
	desktop := wm.getWindowDesktop(windowID) // Get desktop number
	pid := wm.getWindowPID(windowID)         // Get process ID

	return &shared.Window{
		ID:        int(windowID),
		Title:     title,
		Type:      windowType, // "Normal" or "Special" (or others based on EWMH)
		Instance:  instance,
		ClassName: class,
		Desktop:   desktop, // Assign the fetched desktop number
		PID:       pid,     // Assign the fetched process ID
	}
}

// getWindowName attempts to retrieve the window title (_NET_WM_NAME or WM_NAME).
// Returns the title string or an empty string if not found or invalid UTF-8.
func (wm *XLibWindowManager) getWindowName(windowID xproto.Window) string {
	// Prefer _NET_WM_NAME (UTF8)
	utf8Atom := wm.getAtomCached("UTF8_STRING")
	if utf8Atom == 0 {
		log.Warn("Could not get UTF8_STRING atom, cannot fetch _NET_WM_NAME")
		// Fallback below might still work if WM_NAME exists
	} else {
		nameBytes := wm.getWindowPropertyBytes(windowID, "_NET_WM_NAME", utf8Atom)
		if nameBytes != nil {
			// Should already be UTF8, but check validity? Go string handles it.
			return string(nameBytes)
		}
	}

	// Fallback to WM_NAME (STRING encoding - often Latin-1 or system locale)
	nameBytes := wm.getWindowPropertyBytes(windowID, "WM_NAME", xproto.AtomString)
	if nameBytes != nil {
		// This might not be valid UTF-8, but Go strings can hold arbitrary bytes.
		// Conversion might be needed if strict UTF-8 is required downstream.
		return string(nameBytes)
	}

	// log.Debug("No name property found for window %d", windowID)
	return ""
}

// getWindowType determines if a window is "Normal" or "Special" based on _NET_WM_WINDOW_TYPE.
// Returns "Normal" or "Special". Defaults to "Normal" if type property is absent.
func (wm *XLibWindowManager) getWindowType(windowID xproto.Window) string {
	typeAtom := wm.getAtomCached("_NET_WM_WINDOW_TYPE")
	normalAtom := wm.getAtomCached("_NET_WM_WINDOW_TYPE_NORMAL")

	// Default to Normal if property is missing
	propTypeDefault := "Normal"

	if typeAtom == 0 {
		return propTypeDefault // Cannot check type atom
	}

	prop, err := xproto.GetProperty(
		wm.display, false, windowID, typeAtom,
		xproto.AtomAtom, 0, 64, // Request multiple atoms (usually just one or a few)
	).Reply()

	if err != nil || prop == nil || prop.Format != 32 || prop.ValueLen == 0 {
		// Error, or property not set, or wrong format
		return propTypeDefault
	}

	// Property exists, default assumption changes to "Special" unless proven "Normal"
	windowType := "Special"
	containsOnlyNormal := true // Assume true until proven otherwise
	numAtoms := int(prop.ValueLen)
	atomBytes := prop.Value

	for i := 0; i < numAtoms; i++ {
		start := i * 4
		if start+4 > len(atomBytes) {
			log.Warn("Window %d: Malformed _NET_WM_WINDOW_TYPE data", windowID)
			return windowType // Return current assumption ("Special") on malformed data
		}
		currentAtom := xproto.Atom(binary.LittleEndian.Uint32(atomBytes[start : start+4]))
		if normalAtom == 0 || currentAtom != normalAtom {
			containsOnlyNormal = false // Found a non-normal atom, or cannot verify normal atom
		}
		// Potentially check for other known types like DOCK, DIALOG etc. here if needed.
	}

	// If it contained *only* the normal atom (and we could get the normal atom)
	if containsOnlyNormal && numAtoms > 0 && normalAtom != 0 {
		windowType = "Normal"
	} else if numAtoms == 0 {
		// Edge case: Property exists but is empty list? Treat as default.
		windowType = propTypeDefault
	}

	// log.Debug("Window %d classified as type: %s", windowID, windowType)
	return windowType
}

// getWindowClass retrieves the WM_CLASS property (instance and class name).
// Returns the instance and class strings. Both are empty if the property is missing or invalid.
func (wm *XLibWindowManager) getWindowClass(windowID xproto.Window) (string, string) {
	// WM_CLASS is type STRING, contains two null-terminated strings.
	classBytes := wm.getWindowPropertyBytes(windowID, "WM_CLASS", xproto.AtomString)
	if classBytes == nil {
		return "", ""
	}

	// Find the first null terminator
	nullPos := bytes.IndexByte(classBytes, 0)
	if nullPos == -1 || nullPos+1 >= len(classBytes) {
		// Malformed: No null terminator or nothing after it
		// Sometimes the instance is first, sometimes the class.
		// Let's return the whole thing as instance if malformed.
		log.Warn("Window %d: Malformed WM_CLASS property: %q", windowID, string(classBytes))
		return string(classBytes), "" // Or maybe both empty? ""/"" seems safer.
		// return "", ""
	}

	instance := string(classBytes[:nullPos])
	// The second part might also have a null terminator, trim it.
	class := strings.TrimRight(string(classBytes[nullPos+1:]), "\x00")

	return instance, class
}

// getWindowPropertyBytes retrieves a window property as raw bytes.
// It requests the property by name and expected type atom.
// Returns the byte slice value, or nil if not found or an error occurs.
func (wm *XLibWindowManager) getWindowPropertyBytes(windowID xproto.Window, propName string, propType xproto.Atom) []byte {
	atom := wm.getAtomCached(propName)
	if atom == 0 {
		// log.Debug("Could not get atom for property '%s'", propName)
		return nil
	}

	prop, err := xproto.GetProperty(
		wm.display, false, windowID, atom,
		propType, 0, 1024*1024, // Request up to 1MB, should be enough for most props
	).Reply()

	if err != nil {
		// Don't log error here, property might just not exist
		// log.Error("Failed to get property '%s' for window %d: %v", propName, windowID, err)
		return nil
	}

	if prop == nil || prop.ValueLen == 0 {
		// log.Debug("Property '%s' not found or empty for window %d", propName, windowID)
		return nil
	}

	// We could check prop.Format and prop.Type here for stricter validation if needed.
	// For example, check if prop.Type == propType (if propType wasn't Any).

	return prop.Value
}

// getRootWindow retrieves the root window ID for the default screen.
// Returns the root window ID, or 0 on failure.
func (wm *XLibWindowManager) getRootWindow() xproto.Window {
	setup := xproto.Setup(wm.display)
	if setup == nil || len(setup.Roots) == 0 {
		log.Error("Failed to get X server setup information or no screens found")
		return 0
	}
	// Assuming default screen
	return setup.Roots[wm.display.DefaultScreen].Root
}

// isValidWindow checks if a window ID corresponds to an existing window on the server.
// Returns true if the window exists, false otherwise.
func (wm *XLibWindowManager) isValidWindow(windowID xproto.Window) bool {
	// Getting attributes is a common way to check validity. An error means it's likely invalid.
	_, err := xproto.GetWindowAttributes(wm.display, windowID).Reply()
	return err == nil
}

// isValidPropertyReply checks common validity conditions for a GetProperty reply.
// Verifies non-nil reply, expected format, and minimum value length.
func (wm *XLibWindowManager) isValidPropertyReply(reply *xproto.GetPropertyReply, expectedFormat byte, minValueLen uint32) bool {
	if reply == nil {
		// log.Debug("Property reply is nil")
		return false
	}
	if reply.Format != expectedFormat {
		// log.Debug("Property reply format mismatch: got %d, want %d", reply.Format, expectedFormat)
		return false
	}
	if reply.ValueLen < minValueLen {
		// log.Debug("Property reply value length too short: got %d, want at least %d", reply.ValueLen, minValueLen)
		return false
	}
	// Check if byte slice length matches expected data length based on format and ValueLen
	bytesPerItem := int(reply.Format / 8)
	expectedBytes := int(reply.ValueLen) * bytesPerItem
	if len(reply.Value) < expectedBytes {
		log.Warn("Property reply byte slice length (%d) is less than expected (%d) based on ValueLen (%d) and Format (%d)", len(reply.Value), expectedBytes, reply.ValueLen, reply.Format)
		return false // Data is truncated or inconsistent
	}
	return true
}

// getAtomCached retrieves an X atom, using a cache for efficiency.
// Returns the atom ID, or 0 on failure.
func (wm *XLibWindowManager) getAtomCached(name string) xproto.Atom {
	// Check cache first (read lock)
	wm.atomMutex.RLock()
	atom, found := wm.atomCache[name]
	wm.atomMutex.RUnlock()
	if found {
		return atom
	}

	// Not found, need to query X server (write lock)
	wm.atomMutex.Lock()
	defer wm.atomMutex.Unlock()

	// Double-check cache after acquiring write lock, another goroutine might have added it
	atom, found = wm.atomCache[name]
	if found {
		return atom
	}

	// Query X server
	reply, err := xproto.InternAtom(wm.display, false, uint16(len(name)), name).Reply()
	if err != nil {
		log.Error("Failed to intern atom '%s': %v", name, err)
		return 0 // Indicate failure
	}

	// Store in cache
	wm.atomCache[name] = reply.Atom
	return reply.Atom
}

// getAtomNameCached retrieves the name for a given atom, potentially using a cache (reverse mapping).
// Note: Implementing an efficient reverse cache is more complex. This version queries directly.
// Returns the atom name string, or "" on failure.
func (wm *XLibWindowManager) getAtomNameCached(atom xproto.Atom) string {
	// TODO: Implement reverse atom cache if performance becomes an issue.
	// For now, query directly.
	reply, err := xproto.GetAtomName(wm.display, atom).Reply()
	if err != nil {
		// Don't log error loudly, atom might just not have a standard name
		// log.Warn("Failed to get name for atom %d: %v", atom, err)
		return ""
	}
	if reply == nil {
		return ""
	}
	return string(reply.Name)
}

// getWindowDesktop retrieves the desktop number (_NET_WM_DESKTOP) for a window.
// Returns the desktop number (0-indexed) or -1 if sticky (on all desktops) or error.
func (wm *XLibWindowManager) getWindowDesktop(windowID xproto.Window) int {
	desktopAtom := wm.getAtomCached("_NET_WM_DESKTOP")
	if desktopAtom == 0 {
		return -1 // Cannot determine desktop without atom
	}

	// Property type is CARDINAL (32-bit unsigned integer)
	propBytes := wm.getWindowPropertyBytes(windowID, "_NET_WM_DESKTOP", xproto.AtomCardinal)
	if propBytes == nil || len(propBytes) < 4 {
		// Property not set or invalid length
		// Some WMs might omit it for sticky windows, though spec says use 0xFFFFFFFF
		return -1 // Treat missing property as sticky/unknown
	}

	desktopID := binary.LittleEndian.Uint32(propBytes)

	if desktopID == 0xFFFFFFFF { // Standard value for sticky windows
		return -1
	}

	// Return desktop number as int (typically 0-indexed)
	return int(desktopID)
}

// getWindowPID retrieves the process ID (_NET_WM_PID) for a window.
// Returns the PID, or 0 if the property is not set or an error occurs.
func (wm *XLibWindowManager) getWindowPID(windowID xproto.Window) int {
	pidAtom := wm.getAtomCached("_NET_WM_PID")
	if pidAtom == 0 {
		return 0 // Cannot determine PID without atom
	}

	// Property type is CARDINAL (32-bit unsigned integer)
	propBytes := wm.getWindowPropertyBytes(windowID, "_NET_WM_PID", xproto.AtomCardinal)
	if propBytes == nil || len(propBytes) < 4 {
		// Property not set or invalid length
		return 0
	}

	pid := binary.LittleEndian.Uint32(propBytes)

	// Return PID as int
	return int(pid)
}

// WindowTitle gets the title of a window by ID.
// Delegates to the internal getWindowName helper.
func (wm *XLibWindowManager) WindowTitle(windowID int) string {
	// getWindowName handles invalid IDs internally
	return wm.getWindowName(xproto.Window(windowID))
}

// WindowClass gets the class and instance name of a window by ID.
// Delegates to the internal getWindowClass helper.
func (wm *XLibWindowManager) WindowClass(windowID int) (string, string) {
	// getWindowClass handles invalid IDs internally
	return wm.getWindowClass(xproto.Window(windowID))
}

// CloseWindow sends a _NET_CLOSE_WINDOW client message to request graceful closure.
// Args:
//
//	windowID: The ID of the window to close.
//
// Returns:
//
//	error: An error if the message could not be sent.
func (wm *XLibWindowManager) CloseWindow(windowID int) error {
	if !wm.isValidWindow(xproto.Window(windowID)) {
		return fmt.Errorf("cannot close invalid window ID %d", windowID)
	}

	closeAtom := wm.getAtomCached("_NET_CLOSE_WINDOW")
	if closeAtom == 0 {
		return fmt.Errorf("could not get _NET_CLOSE_WINDOW atom")
	}

	root := wm.getRootWindow()
	if root == 0 {
		return fmt.Errorf("could not get root window")
	}

	// EWMH spec for _NET_CLOSE_WINDOW:
	// window = the respective client window
	// message_type = _NET_CLOSE_WINDOW
	// format = 32
	// data.l[0] = timestamp
	// data.l[1] = source indication (1 for normal apps)
	// data.l[2-4] = 0
	// We need a timestamp. Using CurrentTime is often problematic.
	// Fetching a property might work, but let's try 0 (most WMs allow it).
	// Source indication 1 is typical for application requests.

	evData := [20]byte{} // 5 * 32-bit = 20 bytes
	// data.l[0] - timestamp (using 0 for now)
	binary.LittleEndian.PutUint32(evData[0:4], 0)
	// data.l[1] - source indication
	binary.LittleEndian.PutUint32(evData[4:8], 1)
	// data.l[2-4] are already 0

	// Convert byte data to uint32 slice for the event data union
	dataUint32 := make([]uint32, 5)
	for i := 0; i < 5; i++ {
		dataUint32[i] = binary.LittleEndian.Uint32(evData[i*4 : (i+1)*4])
	}

	cm := xproto.ClientMessageEvent{
		Format: 32,
		Window: xproto.Window(windowID),
		Type:   closeAtom,
		Data:   xproto.ClientMessageDataUnionData32New(dataUint32),
	}

	// Send the event to the root window, as window managers listen there
	// for these types of messages on behalf of client windows.
	// Masks determine which clients receive the event. SubstructureNotify
	// and SubstructureRedirect cover most WM scenarios.
	cookie := xproto.SendEventChecked(
		wm.display,
		false, // propagate = false
		root,  // destination = root window
		uint32(xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify),
		string(cm.Bytes()),
	)

	err := cookie.Check()
	if err != nil {
		log.Error("Failed to send _NET_CLOSE_WINDOW event for window %d: %v", windowID, err)
		return fmt.Errorf("failed to send close event: %w", err)
	}

	log.Debug("Sent _NET_CLOSE_WINDOW event for window %d", windowID)
	return nil
}

// Cleanup closes the connection to the X server.
func (wm *XLibWindowManager) Cleanup() {
	wm.atomMutex.Lock() // Ensure no atom operations are ongoing
	defer wm.atomMutex.Unlock()

	if wm.display != nil {
		log.Debug("Closing X server connection")
		wm.display.Close()
		wm.display = nil // Prevent further use
	}
}
