package daemon

import (
	"context"
	"fmt"
	"sync"

	"gofi/pkg/desktop"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

type WindowWatcher struct {
	stopEvent   *sync.WaitGroup
	eventThread *sync.WaitGroup
	wm          desktop.WindowManager
	api         *API
	isStopping  bool
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.Mutex
}

// NewWindowWatcher creates a new WindowWatcher instance
// Args:
//
//	wm: Optional window manager
//	api: Optional API instance
//
// Returns:
//
//	*WindowWatcher: New window watcher instance
func NewWindowWatcher(
	wm desktop.WindowManager,
	api *API,
) *WindowWatcher {
	if wm == nil {
		wm = desktop.Instance()
	}
	if api == nil {
		panic("API cannot be nil")
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return &WindowWatcher{
		stopEvent:   &sync.WaitGroup{},
		eventThread: &sync.WaitGroup{},
		wm:          wm,
		api:         api,
		isStopping:  false,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// ClientList gets the client list from API
// Returns:
//
//	[]*shared.Window: List of windows
func (ww *WindowWatcher) ClientList() []*shared.Window {
	return ww.api.ClientList()
}

// Start starts the window watcher
// Returns:
//
//	bool: True if started successfully
func (ww *WindowWatcher) Start() bool {
	ww.api.InitializeWindowList()

	if !ww.wm.InitEvents() {
		log.Error("Failed to initialize events")
		return false
	}

	ww.startWatcherThread()
	return true
}

// Stop stops the window watcher
// Returns:
//
//	bool: True if stopped successfully
func (ww *WindowWatcher) Stop() bool {
	ww.mutex.Lock()
	if ww.isStopping {
		ww.mutex.Unlock()
		return true
	}
	ww.isStopping = true
	ww.mutex.Unlock()

	// Signal stop
	ww.cancel()

	return true
}

// startWatcherThread starts the watcher thread
func (ww *WindowWatcher) startWatcherThread() {
	ww.eventThread.Add(1)
	go func() {
		defer ww.eventThread.Done()
		ww.windowEventThread()
	}()
}

// windowEventThread runs the window event loop
func (ww *WindowWatcher) windowEventThread() {
	for {
		select {
		case <-ww.ctx.Done():
			return
		default:
			eventName := ww.wm.AwaitEvent(ww.ctx)
			// AwaitEvent returns "" on error, cancellation, or EOF
			if eventName == "" {
				// Check if the context was cancelled, which is an expected way to stop
				if ww.ctx.Err() != nil {
					log.Debug("Window event thread stopping due to context cancellation.")
				} else {
					// If context wasn't cancelled, it's likely an X server error/disconnect
					ww.logError("AwaitEvent returned empty string, likely X connection issue or other error.")
				}
				return // Stop the thread in either case
			}

			// Filter events and update list only for relevant ones
			switch eventName {
			case "PropertyNotifyEvent",
				"MapNotifyEvent",
				"DestroyNotifyEvent",
				"CreateNotifyEvent":
				// We received a valid event name, log it
				// log.Debug("Received window event: %s", eventName)
				ww.api.UpdateWindowList()
			case "UnmapNotifyEvent", "ConfigureNotifyEvent":
				// Ignore these events for list updates
				// log.Debug("Ignoring window event: %s", eventName)
			default:
				// Log unhandled event types if necessary
				log.Warn("Unhandled window event type: %s", eventName)
			}
		}
	}
}

// logError logs an error if not stopping
// Args:
//
//	format: Format string
//	args: Arguments
func (ww *WindowWatcher) logError(format string, args ...interface{}) {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	if !ww.isStopping {
		log.Error(fmt.Sprintf(format, args...))
	}
}

// Cleanup cleans up resources
func (ww *WindowWatcher) Cleanup() {
	ww.Stop()
}
