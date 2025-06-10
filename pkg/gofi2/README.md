# gofi2

gofi2 is the next generation window selector that combines the daemon and client functionality from the original gofi into a single persistent application.

## Architecture

- **Single persistent application**: Combines window monitoring and selection into one process
- **Instance management**: Only one instance runs at a time, subsequent launches activate the existing instance
- **Always running**: After window selection or escape, the app hides but continues monitoring windows
- **Live window list**: Maintains up-to-date window list via X11 event monitoring

## Key Components

- `App`: Main application that manages window monitoring and GUI display
- `InstanceManager`: Handles single-instance behavior via file locks and IPC
- Reuses existing `daemon.API` and `daemon.WindowWatcher` for X11 monitoring
- Uses same fzf-based window selection as original gofi

## Usage

```bash
# Start gofi2 (or activate if already running)
./gofi2

# With debug logging
./gofi2 --log debug
```

## Behavior

1. First launch: Starts persistent app, shows window selector
2. Subsequent launches: Activates existing instance window selector
3. After selection/escape: Hides UI but keeps running and monitoring windows
4. Maintains live window list through X11 event monitoring

## Files

- `instance.go`: Single-instance management with file locks and IPC
- `app.go`: Main application combining daemon and client functionality 