package client

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// FuzzyFinder is the command used for fuzzy finding. Can be replaced for testing.
var FuzzyFinder = "fzf"

// SelectWindow shows GUI for window selection using fzf in st terminal
// Args:
//
//	windows: List of windows to select from
//
// Returns:
//
//	*shared.Window: Selected window or nil
func SelectWindow(windows []shared.Window) {
	formattedLines := FormatWindows(windows, nil, nil)
	tempFiles := createTempFiles()
	defer cleanupTempFiles(tempFiles)

	writeWindowList(formattedLines, tempFiles["list"])
	createFzfScript(tempFiles)
	runTerminalWithFzf(tempFiles["exec"])
}

// createTempFiles creates temporary files for fzf script
// Returns:
//
//	map[string]string: Map of file paths
func createTempFiles() map[string]string {
	tempFiles := make(map[string]string)
	for _, name := range []string{"list", "exec", "result"} {
		file, err := os.CreateTemp("", fmt.Sprintf("gofi-%s-*", name))
		if err != nil {
			log.Error("Failed to create temp file: %s", err)
			return nil
		}
		tempFiles[name] = file.Name()
		file.Close()
	}
	return tempFiles
}

// writeWindowList writes window list to temporary file
// Args:
//
//	formattedLines: Formatted window lines
//	listFile: Path to list file
func writeWindowList(formattedLines []string, listFile string) {
	if err := os.WriteFile(listFile, []byte(strings.Join(formattedLines, "\n")), 0644); err != nil {
		log.Error("Failed to write window list: %s", err)
	}
}

// createFzfScript creates executable script for fzf
// Args:
//
//	tempFiles: Map of temporary files
func createFzfScript(tempFiles map[string]string) {
	script := fmt.Sprintf(`#!/bin/sh
# Catppuccin Mocha colors
export FZF_DEFAULT_OPTS="
  --color=bg+:#313244,bg:#1e1e2e,spinner:#f5e0dc,hl:#f38ba8
  --color=fg:#cdd6f4,header:#f38ba8,info:#cba6f7,pointer:#f5e0dc
  --color=marker:#f5e0dc,fg+:#cdd6f4,prompt:#cba6f7,hl+:#f38ba8
  --bind='ctrl-x:execute(echo {{+}} | sed \"s/.*\\(0x[0-9a-f]*\\)\\$/\\1/\" | xargs xkill -id)+abort'
"
selected=$(cat %s | %s | sed 's/.*0x/0x/g')
if [ -n "$selected" ]; then
    echo "$selected" > %s
    wmctrl -i -a $selected
fi
`, tempFiles["list"], FuzzyFinder, tempFiles["result"])

	file, err := os.Create(tempFiles["exec"])
	if err != nil {
		log.Error("Failed to create exec file: %s", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(script); err != nil {
		log.Error("Failed to write to exec file: %s", err)
		return
	}

	if err := os.Chmod(tempFiles["exec"], 0755); err != nil {
		log.Error("Failed to make script executable: %s", err)
	}
}

// runTerminalWithFzf runs st terminal with the fzf script
// Args:
//
//	scriptPath: Path to script file
func runTerminalWithFzf(scriptPath string) {
	cmd := exec.Command("st",
		"-g", "124x30+1200+800", // geometry: widthxheight+x+y
		"-f", "Monospace:size=12",
		"-t", "gofi",
		"--", scriptPath,
	)

	if err := cmd.Run(); err != nil {
		log.Error("Failed to run terminal: %s", err)
	}
}

// cleanupTempFiles cleans up temporary files
// Args:
//
//	tempFiles: Map of temporary files
func cleanupTempFiles(tempFiles map[string]string) {
	if tempFiles == nil {
		return
	}

	for _, file := range tempFiles {
		if err := os.Remove(file); err != nil {
			log.Error("Failed to remove temp file %s: %s", file, err)
		}
	}
}
