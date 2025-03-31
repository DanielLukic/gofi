# gofi

A window switcher for Linux, inspired by [rofi](https://github.com/davatorium/rofi).

A Cursor IDE experiment. Built with Cursor IDE.

## Features

- Fast window switching using ST + FZF
- Alternative terminal UI mode with fuzzy finder
- Shows window desktop, title, name, command, and class
- Sorts windows by desktop
- Supports both X11 and Wayland (via wmctrl)
- GUI mode and TUI mode

1. Filter Properties
   - `Desktop`: Which desktop number the window is on
   - `Command`: Process name that started the window
   - `Name`: Full window class with instance (e.g. Mail.thunderbird)
   - `Class`: Short window class (e.g. thunderbird)
   - `Title`: Window title

2. Sorting/Grouping
   - Group by desktop
   - Current desktop windows shown first
   - Other desktops follow in numerical order
   - Within each desktop group, keep sort order from wmctrl

3. Implementation Requirements
   - wmctrl: Used for getting window metadata (Command, Class, Desktop)
   - fzf: Used for fuzzy finding (default mode)
   - st: Used for terminal emulation (default)

This specification is binding and should not be modified without explicit user agreement.

## Installation

```bash
go install github.com/dl/gofi@latest
```

## Usage

```bash
gofi [flags]
```

### Flags

- `--tui`: Use terminal UI mode with fuzzy finder instead of ST+FZF

## Dependencies

- `wmctrl`: For window management
- `fzf`: For fuzzy finding (default mode)
- `st`: For terminal emulation (default)
