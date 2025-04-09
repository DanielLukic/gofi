# gofi

Gofi is a tool for listing and selecting active windows on your desktop. It uses a
client-daemon architecture for efficient window monitoring and selection.

## How it Works

Gofi employs a client-daemon model:

*   **Daemon:** Runs in the background, monitors X11 window events (creation,
    deletion, focus changes). It maintains an up-to-date list of active windows.

*   **Client:** Communicates with the daemon via IPC to retrieve the window list.
    It then uses the `st` terminal to display this list, leveraging `fzf` for
    interactive fuzzy searching and selection. Once a window is selected, the
    client uses `wmctrl` to bring it into focus. `wmctrl` and `xkill` are also
    used for managing existing `gofi` windows.

## Usage

Two primary scenarios:

1. Reconfigure your "Alt-Tab" (or equivalent) to call `gofi`. It will show a terminal
and run gofi inside it.

2. Call `gofi -tui` inside a terminal to run gofi without spawning a new terminal.

A daemon will be auto-spawned in both cases, if not already running.

Use "Enter" to select the window to activate aka jump to. Type a few letters to find
the window you want to select.

Note: "Alt-x" will `xkill` the selected window.

## Dependencies

Gofi requires the following external programs to be installed and available in your
`$PATH`:

*   `st` (Simple Terminal)
*   `fzf` (Command-line fuzzy finder)
*   `wmctrl` (Utility to interact with EWMH/NetWM compatible X Window Managers)
*   `xkill` (Tool to kill an X client by its X resource)

*   X11 Libraries (Development libraries might be required for building, e.g.,
    `libx11-dev` on Debian/Ubuntu).

## Usage

To list and select windows (default behavior):
```bash
gofi
```

To run the daemon in the background:
```bash
gofi --daemon
```

To not spawn a new terminal:
```bash
gofi -tui
```

To stop a running daemon:
```bash
gofi --kill
```

To change the log level (e.g., to debug):
```bash
gofi --log debug
```

## License

This project is released into the public domain under The Unlicense - see the
[LICENSE](LICENSE) file for details.