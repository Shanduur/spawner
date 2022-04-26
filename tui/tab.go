package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gizak/termui/v3/widgets"
)

type Tab struct {
	Title        string
	Content      *widgets.List
	AutoScroll   bool
	Mut          sync.RWMutex
	HistoryLimit int
}

type TabOpts struct {
	Title        string
	HistoryLimit int
}

func (tab *Tab) LengthEnforcer() {
	for {
		if len(tab.Content.Rows) >= tab.HistoryLimit && len(tab.Content.Rows) > 1 {
			tab.Mut.Lock()
			for {
				tab.Content.Rows = tab.Content.Rows[1:]
				if len(tab.Content.Rows) <= tab.HistoryLimit {
					break
				}
			}
			tab.Mut.Unlock()
		}
	}
}

func (tab *Tab) Writer() {
	for {
		time.Sleep(time.Second)

		tab.Mut.Lock()
		tab.Content.Rows = append(tab.Content.Rows, fmt.Sprintf("hello"))
		tab.Mut.Unlock()

		if tab.AutoScroll {
			tab.Content.SelectedRow = len(tab.Content.Rows) - 1
		}
	}
}

func (tab *Tab) Resize(w, h int) {
	tab.Content.SetRect(0, 4, w-1, h-4)
}
