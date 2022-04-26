package tui

import (
	"bufio"
	"fmt"
	"os"
	"sync"

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
			fmt.Println("recovered writer", r)
		}
	}()

	tab.Content.Rows = append(tab.Content.Rows, "loaded")

	for tab.Scanner.Scan() {
		tab.Mut.Lock()
		tab.Content.Rows = append(tab.Content.Rows, tab.Scanner.Text())
		tab.Mut.Unlock()

		if tab.AutoScroll {
			tab.Content.SelectedRow = len(tab.Content.Rows) - 1
		}
	}
	if err := tab.Scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
