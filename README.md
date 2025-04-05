# gofi

A window switcher for Linux, inspired by [rofi](https://github.com/davatorium/rofi).

A vibe-code experiment. Built with Cursor and Windsurf IDEs.

## Features

- Fast window switching with Alt-Tab style window ordering
- Alternative terminal UI mode with fuzzy finder
- Shows window desktop, title, name, command, and class

## Usage

```
./gofi [options]
```

### Command-line Options

- `-tui`: Use terminal UI mode instead of graphical mode
- `-daemon`: Run as a daemon (background process)
- `-restart`: Force restart the daemon
- `-kill`: Kill the running daemon

### Notes

- Starting the daemon is handled automatically by the application.
- The `-restart` flag can be used for testing.
- To stop the daemon, use the `-kill` flag.

## Window Properties

- `Desktop`: Which desktop number the window is on
- `Command`: Process name that started the window
- `Name`: Full window class with instance (e.g. Mail.thunderbird)
- `Class`: Short window class (e.g. thunderbird)
- `Title`: Window title

## Window Ordering

- Previously active window is always shown first
- Windows are ordered by activation history

## Dependencies

The following tools are required:
- `wmctrl`: Used for activating windows and closing active gofi gui
- `fzf`: Used for fuzzy finding
- `st`: Used for terminal emulation
- `xkill`: Used for killing windows via <ctrl-x> in gui mode

## Building

```
./build.sh
```
