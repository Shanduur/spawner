package tui

import (
	"bufio"
	"fmt"
	"time"

	l "github.com/Shanduur/spawner/logger"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"golang.org/x/term"
)

type TuiOpts struct {
	Header      string
	RefreshRate int
}

type Tui struct {
	Header    *widgets.Paragraph
	ActiveTab int
	Tabs      []*Tab
	TabPane   *widgets.TabPane
	Keymap    *widgets.Paragraph
	Error     *widgets.Paragraph

	eventPoller  <-chan ui.Event
	ticker       <-chan time.Time
	refreshRate  time.Duration
	width        int
	height       int
	displayError bool
}

func Init(opts TuiOpts) (*Tui, error) {
	header := widgets.NewParagraph()
	header.Text = opts.Header
	header.SetRect(0, 0, 1, 1)
	header.Border = false
	header.TextStyle.Bg = ui.ColorBlack
	header.TextStyle.Fg = ui.ColorWhite

	errpar := widgets.NewParagraph()
	errpar.Text = "?"
	errpar.SetRect(0, 0, 1, 1)
	errpar.Border = false
	errpar.TextStyle.Fg = ui.ColorRed

	tabpane := widgets.NewTabPane()
	tabpane.SetRect(0, 0, 1, 1)
	tabpane.Border = true

	keymap := widgets.NewParagraph()
	keymap.Text = `"q" - quit, "h" - left, "j" - down, "k" - up, "l" - right`
	keymap.SetRect(0, 0, 1, 1)
	keymap.Border = true
	keymap.TextStyle.Bg = ui.ColorBlack
	keymap.TextStyle.Fg = ui.ColorWhite

	t := &Tui{
		Header:      header,
		TabPane:     tabpane,
		Keymap:      keymap,
		Error:       errpar,
		refreshRate: time.Duration(opts.RefreshRate),
	}

	if err := t.AddTab(TabOpts{
		Title:   "main",
		Scanner: bufio.NewScanner(l.Buffer),
	}); err != nil {
		return nil, fmt.Errorf("unable to add main tab: %w", err)
	}

	t.ActiveTab = 0

	return t, nil
}

func (tui *Tui) adjustSize() {
	if !term.IsTerminal(0) {
		panic("not in terminal!")
	}

	for {
		w, h, err := term.GetSize(0)
		if err != nil {
			tui.displayError = true
			tui.Error.Text = err.Error()
		}

		if w < 60 || h < 12 {
			tui.displayError = true
			tui.Error.Text = fmt.Sprintf("wrong size of terminal\ngot (%d, %d) wanted at least (60, 12)", w, h)
			tui.Error.SetRect(0, 0, w-1, h-1)
		}

		if tui.width != w || tui.height != h {
			tui.displayError = false
			tui.width = w
			tui.height = h

			tui.Header.SetRect(0, 0, tui.width-1, 1)
			tui.TabPane.SetRect(0, 1, tui.width-1, 4)
			tui.Keymap.SetRect(0, tui.height-4, tui.width-1, tui.height-1)

			for i := 0; i < len(tui.Tabs); i++ {
				tui.Tabs[i].Resize(tui.width, tui.height)
			}

			ui.Clear()
		}
	}
}

func (tui *Tui) AddTab(opts TabOpts) error {
	lst := widgets.NewList()
	lst.Border = true
	lst.Title = opts.Title
	lst.SetRect(0, 0, 1, 1)
	lst.BorderStyle.Fg = ui.ColorWhite
	lst.SelectedRowStyle.Bg = ui.ColorWhite
	lst.SelectedRowStyle.Fg = ui.ColorBlack

	if opts.Title == "" {
		opts.Title = fmt.Sprintf("%d", len(tui.Tabs))
	}

	if opts.HistoryLimit == 0 {
		opts.HistoryLimit = 1000
	}

	tui.Tabs = append(tui.Tabs, &Tab{
		Title:        opts.Title,
		HistoryLimit: opts.HistoryLimit,
		Content:      lst,
		AutoScroll:   true,
		Scanner:      opts.Scanner,
	})

	tui.TabPane.TabNames = append(tui.TabPane.TabNames, opts.Title)

	go tui.Tabs[len(tui.Tabs)-1].LengthEnforcer()

	go tui.Tabs[len(tui.Tabs)-1].Writer()

	return nil
}

func (tui *Tui) Start() error {
	err := ui.Init()
	if err != nil {
		return fmt.Errorf("unable to init ui: %w", err)
	}

	tui.eventPoller = ui.PollEvents()
	tui.ticker = time.NewTicker(time.Second / tui.refreshRate).C

	go tui.adjustSize()

	for {
		select {
		case e := <-tui.eventPoller:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "h":
				tui.TabPane.FocusLeft()
			case "l":
				tui.TabPane.FocusRight()
			case "j":
				tui.Tabs[tui.ActiveTab].Content.ScrollUp()
				tui.Tabs[tui.ActiveTab].AutoScroll = false
			case "k":
				tui.Tabs[tui.ActiveTab].Content.ScrollDown()
				if len(tui.Tabs[tui.ActiveTab].Content.Rows)-1 == tui.Tabs[tui.ActiveTab].Content.SelectedRow {
					tui.Tabs[tui.ActiveTab].AutoScroll = true
				}
			case "<PageDown>":
				tui.Tabs[tui.ActiveTab].AutoScroll = true
			}
		case <-tui.ticker:
			if !tui.displayError {
				ui.Render(tui.Header, tui.TabPane)
				tui.renderTab()
				ui.Render(tui.Keymap)
			} else {
				ui.Render(tui.Error)
			}
		}
	}
}

func (tui *Tui) renderTab() {
	tui.ActiveTab = tui.TabPane.ActiveTabIndex
	ui.Render(tui.Tabs[tui.ActiveTab].Content)
}
