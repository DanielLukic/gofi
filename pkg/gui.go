package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GuiSelectWindow() error {
	x, err := NewXUtil()
	if err != nil {
		return fmt.Errorf("failed to connect to X server: %v", err)
	}
	defer x.Close()

	windows, _ := ListWindows(x)

	_close_existing_instance(windows)

	list := _prepare_fzf_list(windows)

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

	cmd := exec.Command("st", "-g", "124x30+1200+800", "-f", "Monospace:size=12", "-t", "gofi", "--", execFile)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error launching terminal: %v", err)
	}

	return nil
}

func _close_existing_instance(windows []Window) {
	for _, w := range windows {
		if strings.Contains(w.Title, "gofi") && w.Command == "st" {
			exec.Command("wmctrl", "-i", "-c", w.ID).Run()
		}
	}
}

func _prepare_fzf_list(windows []Window) string {
	maxCmdLen := 15
	maxTitleLen := 55
	maxClassLen := 30

	var lines []string
	for _, window := range windows {
		cmd := _truncate(window.Command, maxCmdLen)
		title := _truncate(window.Title, maxTitleLen)
		class := _truncate(window.Name, maxClassLen) // Use full class name

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
