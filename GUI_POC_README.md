# GoFi2 GUI PoC ✅ WORKING

This is a Proof of Concept (PoC) for adding GUI support to gofi2 using [richardwilkes/unison](https://github.com/richardwilkes/unison) instead of the terminal-based interface.

## ✅ Status: WORKING

The PoC is now fully functional! The GLFW initialization issue has been resolved.

## What was built

1. **New `-gui` flag**: Added support for a GUI mode alongside the existing terminal mode
2. **Unison integration**: Uses the Unison toolkit for cross-platform GUI development
3. **Go fuzzy search**: Replaced `fzf` with [sahilm/fuzzy](https://github.com/sahilm/fuzzy) for Go-native fuzzy searching
4. **Simple window selector**: A basic GUI that shows windows as clickable buttons with search functionality

## New Dependencies

- `github.com/richardwilkes/unison v0.80.4` - Cross-platform GUI toolkit
- `github.com/sahilm/fuzzy v0.1.1` - Go fuzzy finder library

## How to test

### Prerequisites
- X11 development libraries (for Unison)
- `wmctrl` (for window activation)

### Build
```bash
go build -v ./cmd/gofi2
```

### Kill any existing instances first
```bash
./gofi2 -kill
```

### Run GUI version
```bash
./gofi2 -gui --log debug
```

### Run traditional terminal version
```bash
./gofi2
```

## Architecture

### Traditional Mode
- Uses `st` terminal + `fzf` for selection
- Spawns external processes for UI

### GUI Mode  
- Pure Go GUI using Unison
- In-process fuzzy searching
- Native window management
- GLFW-based rendering

## Key Files

- `cmd/gofi2/main.go` - Updated to support `-gui` flag
- `pkg/gofi2/gui_app.go` - GUI version of the app
- `pkg/gui/simple_selector.go` - Unison-based window selector
- `pkg/gofi2/interfaces.go` - Common interface for both app types

## Fixed Issues

- ✅ **GLFW Initialization**: Fixed "The GLFW library is not initialized" panic by ensuring `unison.Start()` is called before creating windows
- ✅ **Window Creation Order**: Deferred window creation until after GLFW is ready
- ✅ **Instance Management**: Properly handle existing instance detection

## Current Limitations

This is a basic PoC with the following limitations:

1. **Simple UI**: Basic button list instead of a sophisticated table/list view
2. **No live search**: Search field doesn't automatically filter as you type
3. **Basic styling**: Uses default Unison styling
4. **No keyboard shortcuts**: GUI relies on mouse interaction

## Future Improvements

- Add live search filtering as you type
- Better keyboard navigation (arrow keys, Enter, Escape)
- Improved styling and layout
- Better error handling
- Window management improvements
- More sophisticated list/table view

## Why Unison?

- **Pure Go**: No C dependencies beyond what GLFW requires
- **Cross-platform**: Works on macOS, Windows, and Linux
- **Active development**: Regularly updated and maintained
- **Game project friendly**: Aligns with the "fail fast" philosophy
- **Successfully integrated**: Proven to work with existing gofi2 architecture 