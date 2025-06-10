package gofi2

import (
	"context"
	"fmt"
	"sync"

	"github.com/richardwilkes/unison"

	"gofi/pkg/daemon"
	"gofi/pkg/gui"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

type GUIApp struct {
	api         *daemon.API
	watcher     *daemon.WindowWatcher
	isVisible   bool
	mutex       sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	exitChannel chan bool
	initialized bool
}

func NewGUIApp() *GUIApp {
	ctx, cancel := context.WithCancel(context.Background())

	api := daemon.NewAPI()
	watcher := daemon.NewWindowWatcher(nil, api)

	return &GUIApp{
		api:         api,
		watcher:     watcher,
		isVisible:   false,
		ctx:         ctx,
		cancel:      cancel,
		exitChannel: make(chan bool),
		initialized: false,
	}
}

func (app *GUIApp) Start() error {
	if !app.watcher.Start() {
		return fmt.Errorf("failed to start window watcher")
	}
	log.Debug("Window watcher started")

	return nil
}

func (app *GUIApp) Show() {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	if app.isVisible {
		return
	}

	app.isVisible = true
}

func (app *GUIApp) Hide() {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.isVisible = false
}

func (app *GUIApp) Run() {
	unison.Start(unison.StartupFinishedCallback(func() {
		windows := app.getWindowList()
		if len(windows) == 0 {
			log.Info("No windows to select from")
			return
		}

		selector := gui.NewSimpleWindowSelector(windows)
		selector.Run()
	}))
}

func (app *GUIApp) Exit() {
	app.exitChannel <- true
}

func (app *GUIApp) Cleanup() {
	app.cancel()
	if app.watcher != nil {
		app.watcher.Stop()
	}
}

func (app *GUIApp) showGUIWindowSelector() {
	windows := app.getWindowList()
	if len(windows) == 0 {
		log.Info("No windows to select from")
		return
	}

	selector := gui.NewSimpleWindowSelector(windows)
	selector.Run()
	app.Hide()
}

func (app *GUIApp) getWindowList() []shared.Window {
	windowPtrs := app.api.ClientList()
	windows := make([]shared.Window, len(windowPtrs))
	for i, w := range windowPtrs {
		windows[i] = *w
	}
	return windows
}
