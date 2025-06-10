package gofi2

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"gofi/pkg/client"
	"gofi/pkg/daemon"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

type App struct {
	api         *daemon.API
	watcher     *daemon.WindowWatcher
	isVisible   bool
	mutex       sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	exitChannel chan bool
}

func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())

	api := daemon.NewAPI()
	watcher := daemon.NewWindowWatcher(nil, api)

	return &App{
		api:         api,
		watcher:     watcher,
		isVisible:   false,
		ctx:         ctx,
		cancel:      cancel,
		exitChannel: make(chan bool),
	}
}

func (app *App) Start() error {
	if !app.watcher.Start() {
		return fmt.Errorf("failed to start window watcher")
	}
	log.Debug("Window watcher started")
	return nil
}

func (app *App) Show() {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	if app.isVisible {
		return
	}

	app.isVisible = true
	go app.showWindowSelector()
}

func (app *App) Hide() {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.isVisible = false
}

func (app *App) Run() {
	<-app.exitChannel
}

func (app *App) Exit() {
	app.exitChannel <- true
}

func (app *App) Cleanup() {
	app.cancel()
	if app.watcher != nil {
		app.watcher.Stop()
	}
}

func (app *App) showWindowSelector() {
	defer app.Hide()

	windows := app.getWindowList()
	if len(windows) == 0 {
		log.Info("No windows to select from")
		return
	}

	formattedLines := client.FormatWindows(windows, nil, nil)
	tempFiles := createTempFiles()
	if tempFiles == nil {
		return
	}
	defer cleanupTempFiles(tempFiles)

	writeWindowList(formattedLines, tempFiles["list"])
	createFzfScript(tempFiles)
	runFzf(tempFiles["exec"])
}

func (app *App) getWindowList() []shared.Window {
	windowPtrs := app.api.ClientList()
	windows := make([]shared.Window, len(windowPtrs))
	for i, w := range windowPtrs {
		windows[i] = *w
	}
	return windows
}

func createTempFiles() map[string]string {
	tempFiles := make(map[string]string)
	for _, name := range []string{"list", "exec", "result"} {
		file, err := os.CreateTemp("", fmt.Sprintf("gofi2-%s-*", name))
		if err != nil {
			log.Error("Failed to create temp file: %s", err)
			return nil
		}
		tempFiles[name] = file.Name()
		file.Close()
	}
	return tempFiles
}

func writeWindowList(formattedLines []string, listFile string) {
	if err := os.WriteFile(listFile, []byte(strings.Join(formattedLines, "\n")), 0644); err != nil {
		log.Error("Failed to write window list: %s", err)
	}
}

func createFzfScript(tempFiles map[string]string) {
	script := fmt.Sprintf(`#!/bin/bash

get_win_id() {
    sed "s/.*0x/0x/;s/}//"
}
export -f get_win_id

kill_window() {
    xargs xkill -id >> /tmp/gofi2.log 2>&1
}
export -f kill_window

export FZF_DEFAULT_OPTS="
  --color=bg+:#313244,bg:#1e1e2e,spinner:#f5e0dc,hl:#f38ba8
  --color=fg:#cdd6f4,header:#f38ba8,info:#cba6f7,pointer:#f5e0dc
  --color=marker:#f5e0dc,fg+:#cdd6f4,prompt:#cba6f7,hl+:#f38ba8
  --bind='alt-x:execute(echo {{+}} | get_win_id | kill_window >> /tmp/gofi2.log 2>&1)+abort'
"

gofi2=$(xdotool search --name '^gofi2$')
if [ -n "$gofi2" ]; then
    wmctrl -i -r $gofi2 -b add,skip_taskbar,above
fi

selected=$(cat %s | fzf | sed 's/.*0x/0x/g')
if [ -n "$selected" ]; then
    echo "$selected" > %s
    wmctrl -i -a $selected
fi
`, tempFiles["list"], tempFiles["result"])

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

func runFzf(scriptPath string) {
	cmd := exec.Command("st",
		"-g", "124x30+1200+800",
		"-f", "Monospace:size=12",
		"-t", "gofi2",
		"--", scriptPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Error("Failed to run terminal: %s", err)
	}
}

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
