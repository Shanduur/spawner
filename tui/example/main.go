package main

import (
	"log"

	"github.com/Shanduur/spawner/tui"
)

func main() {
	x, err := tui.Init(tui.TuiOpts{
		Header:      "spawner",
		RefreshRate: 30,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := x.AddTab(tui.TabOpts{
		Title:        "example",
		HistoryLimit: 5,
	}); err != nil {
		log.Fatal(err.Error())
	}

	x.Start()
}
