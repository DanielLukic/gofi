# gofi

Gofi is a tool for listing and selecting active windows on your desktop. 

## How it Works

Gofi runs as a single process that:

*   Monitors X11 window events (creation, deletion, focus changes) in real-time
*   Maintains an up-to-date list of active windows
*   Uses the `st` terminal to display this list, leveraging `fzf` for interactive fuzzy searching and selection
*   Uses `wmctrl` to bring the selected window into focus
*   Uses `wmctrl` and `xkill` to manage existing `gofi` windows

## Usage

Reconfigure your "Alt-Tab" (or equivalent) to call `gofi`. It will show a terminal and run gofi inside it.

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

## Command Line Options

To list and select windows (default behavior):
```bash
gofi
```

To kill a running gofi instance:
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