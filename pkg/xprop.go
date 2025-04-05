package pkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
)

type XUtil struct {
	X *xgbutil.XUtil
}

func NewXUtil() (*XUtil, error) {
	X, err := xgbutil.NewConn()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %v", err)
	}
	return &XUtil{X: X}, nil
}

func (x *XUtil) Close() {
	x.X.Conn().Close()
}

func ListWindows(x *XUtil) ([]Window, error) {
	var names, _ = _filtered_windows(x)
	windows := _make_windows(names, x)
	if len(names) > 1 {
		// Swap first two entries:
		windows[0], windows[1] = windows[1], windows[0]
	}
	return windows, nil
}

func _make_windows(names []string, x *XUtil) []Window {
	windows := make([]Window, len(names))
	for i, windowID := range names {
		info, _ := _win_info(x, windowID)
		windows[i] = Window{
			ID:      windowID,
			Desktop: 0,
			PID:     info["PID"],
			Command: info["Command"],
			Class:   info["Name"],
			Name:    info["Class"],
			Title:   info["Title"],
		}
	}
	return windows
}

func _filtered_windows(x *XUtil) ([]string, error) {
	windowIDs, _ := _list_windows(x)
	normalWindows := _filter_normal(x, windowIDs)
	return normalWindows, nil
}

func _list_windows(x *XUtil) ([]string, error) {
	clients, _ := ewmh.ClientListStackingGet(x.X)
	windowIDs := make([]string, len(clients))
	for i, client := range clients {
		windowIDs[len(clients)-i-1] = fmt.Sprintf("0x%x", client)
	}
	return windowIDs, nil
}

func _filter_normal(x *XUtil, windowIDs []string) []string {
	var normalWindows []string
	for _, windowID := range windowIDs {
		isNormal, _ := _is_normal(x, windowID)
		if isNormal {
			normalWindows = append(normalWindows, windowID)
		}
	}
	return normalWindows
}

func _is_normal(x *XUtil, windowID string) (bool, error) {

	windowType, _ := _xprop(x, windowID, "_NET_WM_WINDOW_TYPE")
	if strings.Contains(windowType, "_NET_WM_WINDOW_TYPE_DOCK") ||
		strings.Contains(windowType, "_NET_WM_WINDOW_TYPE_DESKTOP") {
		return false, nil
	}

	windowState, _ := _xprop(x, windowID, "_NET_WM_STATE")
	if strings.Contains(windowState, "_NET_WM_STATE_SKIP_TASKBAR") {
		return false, nil
	}

	return true, nil
}

func _win_info(x *XUtil, windowID string) (map[string]string, error) {
	title, _ := _xprop(x, windowID, "WM_NAME")
	class, _ := _xprop(x, windowID, "WM_CLASS")
	fmt.Println(class)
	desktop, _ := _xprop(x, windowID, "_NET_WM_DESKTOP")
	info := make(map[string]string)
	info["Title"] = title
	info["Class"] = class
	info["Desktop"] = desktop
	return info, nil
}

func get_active_window(x *XUtil) (string, error) {
	// Get the active window from EWMH
	active, err := ewmh.ActiveWindowGet(x.X)
	if err != nil {
		return "", fmt.Errorf("failed to get active window: %v", err)
	}

	// Convert to hex string format
	return fmt.Sprintf("0x%x", active), nil
}

// setup_active_window_events sets up the X connection to listen for active window changes
func setup_active_window_events(x *XUtil, callback func(string)) error {
	// Get the _NET_ACTIVE_WINDOW atom
	net_active_window, err := xprop.Atm(x.X, "_NET_ACTIVE_WINDOW")
	if err != nil {
		return fmt.Errorf("failed to get _NET_ACTIVE_WINDOW atom: %v", err)
	}

	// Set up event mask on root window to listen for property changes
	root := x.X.RootWin()
	xproto.ChangeWindowAttributes(x.X.Conn(), root, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})

	// Start a goroutine to listen for events
	go func() {
		for {
			ev, err := x.X.Conn().WaitForEvent()
			if err != nil {
				continue
			}

			// Check if it's a property notify event
			if propEv, ok := ev.(xproto.PropertyNotifyEvent); ok {
				// Check if the property that changed is _NET_ACTIVE_WINDOW
				if propEv.Atom == net_active_window {
					// Get the new active window
					activeWin, err := get_active_window(x)
					if err == nil && activeWin != "" {
						// Call the callback with the new active window
						callback(activeWin)
					}
				}
			}
		}
	}()

	return nil
}

func _xprop(x *XUtil, windowID, property string) (string, error) {
	var wid uint64
	if strings.HasPrefix(windowID, "0x") {
		wid, _ = strconv.ParseUint(windowID[2:], 16, 64)
	} else {
		wid, _ = strconv.ParseUint(windowID, 10, 64)
	}

	win := xproto.Window(wid)

	switch property {
	case "WM_NAME":
		name, err := ewmh.WmNameGet(x.X, win)
		if err != nil {
			// Try the old-style name as fallback
			atoms, err := xprop.GetProperty(x.X, win, "WM_NAME")
			if err != nil {
				return "", fmt.Errorf("failed to get window name: %v", err)
			}
			if len(atoms.Value) > 0 {
				name = string(atoms.Value)
			}
		}
		return name, nil

	case "WM_CLASS":
		atoms, err := xprop.GetProperty(x.X, win, "WM_CLASS")
		if err != nil {
			return "", fmt.Errorf("failed to get window class: %v", err)
		}
		if len(atoms.Value) > 0 {
			classes := strings.Split(string(atoms.Value), "\x00")
			if len(classes) >= 2 {
				return fmt.Sprintf("%s %s", classes[0], classes[1]), nil
			}
		}
		return "", fmt.Errorf("invalid WM_CLASS format")

	case "_NET_WM_DESKTOP":
		desktop, err := ewmh.WmDesktopGet(x.X, win)
		if err != nil {
			return "", fmt.Errorf("failed to get window desktop: %v", err)
		}
		return fmt.Sprintf("%d", desktop), nil

	case "_NET_WM_WINDOW_TYPE":
		// Get the property directly to avoid type issues
		atoms, err := xprop.GetProperty(x.X, win, "_NET_WM_WINDOW_TYPE")
		if err != nil {
			return "", fmt.Errorf("failed to get window type: %v", err)
		}

		// Convert atom values to names
		typeStrings := make([]string, 0)
		for i := 0; i < len(atoms.Value)/4; i++ {
			// Extract 32-bit atom value
			atomVal := xproto.Atom(uint32(atoms.Value[i*4]) |
				uint32(atoms.Value[i*4+1])<<8 |
				uint32(atoms.Value[i*4+2])<<16 |
				uint32(atoms.Value[i*4+3])<<24)

			name, err := xprop.AtomName(x.X, atomVal)
			if err != nil {
				typeStrings = append(typeStrings, fmt.Sprintf("Unknown(%d)", atomVal))
			} else {
				typeStrings = append(typeStrings, name)
			}
		}
		return strings.Join(typeStrings, ", "), nil

	case "_NET_WM_STATE":
		// Get the property directly to avoid type issues
		atoms, err := xprop.GetProperty(x.X, win, "_NET_WM_STATE")
		if err != nil {
			return "", fmt.Errorf("failed to get window state: %v", err)
		}

		// Convert atom values to names
		stateStrings := make([]string, 0)
		for i := 0; i < len(atoms.Value)/4; i++ {
			// Extract 32-bit atom value
			atomVal := xproto.Atom(uint32(atoms.Value[i*4]) |
				uint32(atoms.Value[i*4+1])<<8 |
				uint32(atoms.Value[i*4+2])<<16 |
				uint32(atoms.Value[i*4+3])<<24)

			name, err := xprop.AtomName(x.X, atomVal)
			if err != nil {
				stateStrings = append(stateStrings, fmt.Sprintf("Unknown(%d)", atomVal))
			} else {
				stateStrings = append(stateStrings, name)
			}
		}
		return strings.Join(stateStrings, ", "), nil
	}

	return "", fmt.Errorf("unsupported property: %s", property)
}
