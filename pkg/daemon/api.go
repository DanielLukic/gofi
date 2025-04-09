package daemon

import (
	"sync"

	"gofi/pkg/desktop"
	"gofi/pkg/shared"
)

type API struct {
	windows    *WindowList
	autoCloser *GofiAutoCloser
	mutex      sync.RWMutex
}

func NewAPI() *API {
	wm := desktop.Instance()
	autoCloser := NewGofiAutoCloser(wm)
	windows := NewWindowList(wm, NewHistory())

	return &API{
		windows:    windows,
		autoCloser: autoCloser,
	}
}

func (api *API) ClientList() []*shared.Window {
	api.mutex.RLock()
	defer api.mutex.RUnlock()
	return api.windows.ClientList()
}

func (api *API) InitializeWindowList() {
	api.mutex.Lock()
	defer api.mutex.Unlock()
	api.windows.Initialize()
}

func (api *API) UpdateWindowList() {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	api.windows.UpdateWindowList()
	api.autoCloser.CheckFocusAndClose()
}
