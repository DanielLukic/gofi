package gui

import (
	"os/exec"

	"github.com/richardwilkes/unison"
	"github.com/sahilm/fuzzy"

	"gofi/pkg/client"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

type SimpleWindowSelector struct {
	window       *unison.Window
	searchField  *unison.Field
	windows      []shared.Window
	filteredList []string
	allList      []string
	panel        *unison.Panel
}

func NewSimpleWindowSelector(windows []shared.Window) *SimpleWindowSelector {
	sws := &SimpleWindowSelector{
		windows: windows,
	}

	formattedLines := client.FormatWindows(windows, nil, nil)
	sws.allList = formattedLines
	sws.filteredList = make([]string, len(formattedLines))
	copy(sws.filteredList, formattedLines)

	return sws
}

func (sws *SimpleWindowSelector) createWindow() {
	var err error
	sws.window, err = unison.NewWindow("gofi2 - Window Selector")
	if err != nil {
		log.Error("Failed to create window: %s", err)
		return
	}

	sws.window.SetFrameRect(unison.Rect{
		Point: unison.Point{X: 100, Y: 100},
		Size:  unison.Size{Width: 800, Height: 600},
	})

	content := sws.window.Content()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: 10,
		VSpacing: 10,
	})

	sws.createSearchField(content)
	sws.createWindowList(content)

	sws.window.Pack()
}

func (sws *SimpleWindowSelector) createSearchField(parent *unison.Panel) {
	sws.searchField = unison.NewField()
	sws.searchField.SetText("")

	label := unison.NewLabel()
	label.SetTitle("Search:")

	parent.AddChild(label)
	parent.AddChild(sws.searchField)
}

func (sws *SimpleWindowSelector) createWindowList(parent *unison.Panel) {
	sws.panel = unison.NewPanel()
	sws.panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: 5,
		VSpacing: 2,
	})

	sws.updateWindowList()

	scrollArea := unison.NewScrollPanel()
	scrollArea.SetContent(sws.panel, 0, 0)

	parent.AddChild(scrollArea)
}

func (sws *SimpleWindowSelector) updateWindowList() {
	sws.panel.RemoveAllChildren()

	for _, windowText := range sws.filteredList {
		windowIndex := sws.findWindowIndex(windowText)
		if windowIndex >= 0 {
			button := unison.NewButton()
			button.SetTitle(windowText)

			window := sws.windows[windowIndex]
			button.ClickCallback = func() {
				sws.activateWindow(window)
				sws.Close()
			}

			sws.panel.AddChild(button)
		}
	}

	sws.panel.MarkForLayoutAndRedraw()
}

func (sws *SimpleWindowSelector) findWindowIndex(windowText string) int {
	for i, text := range sws.allList {
		if text == windowText {
			return i
		}
	}
	return -1
}

func (sws *SimpleWindowSelector) filterWindows(query string) {
	if query == "" {
		sws.filteredList = make([]string, len(sws.allList))
		copy(sws.filteredList, sws.allList)
	} else {
		matches := fuzzy.Find(query, sws.allList)
		sws.filteredList = make([]string, len(matches))
		for i, match := range matches {
			sws.filteredList[i] = sws.allList[match.Index]
		}
	}
	sws.updateWindowList()
}

func (sws *SimpleWindowSelector) activateWindow(window shared.Window) {
	log.Debug("Activating window: %s (%s)", window.Title, window.HexID())

	cmd := exec.Command("wmctrl", "-i", "-a", window.HexID())
	if err := cmd.Run(); err != nil {
		log.Error("Failed to activate window %s: %s", window.HexID(), err)
	}
}

func (sws *SimpleWindowSelector) Show() {
	if sws.window != nil {
		sws.window.ToFront()
	}
}

func (sws *SimpleWindowSelector) Close() {
	if sws.window != nil {
		sws.window.AttemptClose()
	}
}

func (sws *SimpleWindowSelector) Run() {
	sws.createWindow()
	sws.Show()
}
