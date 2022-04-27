package tui

import (
	"bufio"
	"fmt"
	"sync"
	"time"

	l "github.com/Shanduur/spawner/logger"
	"github.com/gizak/termui/v3/widgets"
)

type Tab struct {
	Title        string
	Content      *widgets.List
	AutoScroll   bool
	Mut          sync.RWMutex
	HistoryLimit int
	Scanner      *bufio.Scanner
}

type TabOpts struct {
	Title        string
	HistoryLimit int
	Scanner      *bufio.Scanner
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

func (tab *Tab) Resize(w, h int) {
	tab.Content.SetRect(0, 4, w-1, h-4)
}

func (tab *Tab) Writer() {
	defer func() {
		if r := recover(); r != nil {
			l.Log().WithField(l.From, tab.Title).Errorf("recovered writer", r)
		}
	}()

	tab.Content.Rows = append(tab.Content.Rows, "waiting...")

	ticker := time.NewTicker(time.Second).C
	start := time.Now().Second()
	for {
		for tab.Scanner.Scan() {
			tab.Mut.Lock()
			tab.Content.Rows = append(tab.Content.Rows, tab.Scanner.Text())
			tab.Mut.Unlock()

			if tab.AutoScroll {
				tab.Content.SelectedRow = len(tab.Content.Rows) - 1
			}
		}
		if err := tab.Scanner.Err(); err != nil {
			l.Log().WithField(l.From, tab.Title).Errorf("reading standard input: %s", err.Error())
			time.Sleep(time.Second)
			select {
			case <-ticker:
				tab.Content.Rows = append(tab.Content.Rows, fmt.Sprintf("waiting for %d...", time.Now().Second()-start))
			}
		}
	}
}
