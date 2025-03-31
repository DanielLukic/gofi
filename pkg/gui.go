package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FzfSelector implements window selection using FZF
type FzfSelector struct {
	manager  Manager
	filter   Filter
	terminal string
}

// NewFzfSelector creates a new FzfSelector instance
func NewFzfSelector(manager Manager, filter Filter) *FzfSelector {
	return &FzfSelector{
		manager:  manager,
		filter:   filter,
		terminal: "st", // default terminal
	}
}

// SetTerminal sets the terminal emulator to use
func (s *FzfSelector) SetTerminal(terminal string) {
	s.terminal = terminal
}

// SelectWindow opens a terminal with FZF for window selection
func (s *FzfSelector) SelectWindow() error {
	// Get windows and filter them
	windows := s.manager.GetWindows()

	// Close any existing gofi instances (fire & forget)
	s.closeExistingInstance(windows)

	var filteredWindows []Window
	for _, w := range windows {
		if s.filter.ShouldInclude(w) {
			filteredWindows = append(filteredWindows, w)
		}
	}

	// Sort windows by desktop
	currentDesktop, err := s.manager.GetCurrentDesktop()
	if err != nil {
		return err
	}
	filteredWindows = SortWindows(filteredWindows, currentDesktop)

	// Prepare the list for fzf
	list := s.prepareFzfList(filteredWindows)

	// Write list to temporary file
	tmpDir := os.TempDir()
	listFile := filepath.Join(tmpDir, "fzf_list")
	execFile := filepath.Join(tmpDir, "fzf_exec")

	if err := os.WriteFile(listFile, []byte(list), 0644); err != nil {
		return fmt.Errorf("error writing list file: %v", err)
	}

	// Create and write the execution script
	execScript := `#!/usr/bin/bash

# Catppuccin Mocha colors
export FZF_DEFAULT_OPTS="
  --color=bg+:#313244,bg:#1e1e2e,spinner:#f5e0dc,hl:#f38ba8
  --color=fg:#cdd6f4,header:#f38ba8,info:#cba6f7,pointer:#f5e0dc
  --color=marker:#f5e0dc,fg+:#cdd6f4,prompt:#cba6f7,hl+:#f38ba8
"

selected=$(cat ` + listFile + ` | fzf | sed 's/.*0x/0x/g')
wmctrl -i -a $selected`

	if err := os.WriteFile(execFile, []byte(execScript), 0744); err != nil {
		return fmt.Errorf("error writing exec file: %v", err)
	}

	// Launch terminal with fzf
	var cmd *exec.Cmd
	switch s.terminal {
	case "st":
		cmd = exec.Command("st", "-g", "124x30+1200+800", "-f", "Monospace:size=12", "-t", "gofi", "--", execFile)
	case "xterm":
		cmd = exec.Command("xterm", "-geometry", "124x30+1200+800", "-fa", "Monospace", "-fs", "12", "-T", "gofi", "-e", execFile)
	default:
		return fmt.Errorf("unsupported terminal: %s", s.terminal)
	}

	// Start the command and wait a moment to ensure it launches properly
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error launching terminal: %v", err)
	}

	return nil
}

// closeExistingInstance checks if there's already a gofi instance running and closes it
func (s *FzfSelector) closeExistingInstance(windows []Window) {
	for _, w := range windows {
		// Look for a window with "gofi" in the title and "st" as the command
		if strings.Contains(w.Title, "gofi") && w.Command == "st" {
			// Close the window using wmctrl (fire & forget)
			exec.Command("wmctrl", "-i", "-c", w.ID).Run()
		}
	}
}

func (s *FzfSelector) prepareFzfList(windows []Window) string {
	// Calculate maximum widths
	maxCmdLen := 15
	maxTitleLen := 55
	maxClassLen := 30

	var lines []string
	for _, window := range windows {
		cmd := s.truncateString(window.Command, maxCmdLen)
		title := s.truncateString(window.Title, maxTitleLen)
		class := s.truncateString(window.Name, maxClassLen) // Use full class name

		// Format according to rufi.rb order
		line := fmt.Sprintf("%d: %-*s %-*s %-*s %s",
			window.Desktop,
			maxCmdLen, cmd,
			maxTitleLen, title,
			maxClassLen, class,
			window.ID)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (s *FzfSelector) truncateString(str string, maxLen int) string {
	if len(str) > maxLen {
		return str[:maxLen]
	}
	return str
}
