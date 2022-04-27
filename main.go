package main

import (
	"bufio"
	"context"
	"os"
	"sync"

	l "github.com/Shanduur/spawner/logger"
	"github.com/Shanduur/spawner/spawner"
	"github.com/Shanduur/spawner/tui"
)

func main() {
	var wg sync.WaitGroup

	spr, err := spawner.Unmarshal(".spawnfile.yml")
	if err != nil {
		l.Log().Fatalf("unable to parse file: %s", err.Error())
	}
	defer os.RemoveAll(spr.Prefix)

	ctx := context.Background()

	x, err := tui.Init(tui.TuiOpts{
		Header:      "spawner",
		RefreshRate: 30,
	})
	if err != nil {
		l.Log().Fatal(err.Error())
	}
	defer tui.Close()

	if err := spr.Populate(); err != nil {
		l.Log().Fatal(err.Error())
	}

	for i := 0; i < len(spr.Components); i++ {
		if err := x.AddTab(tui.TabOpts{
			Title:        spr.Components[i].String() + ".err",
			HistoryLimit: 1000,
			Scanner:      bufio.NewScanner(spr.Components[i].Stderr),
		}); err != nil {
			l.Log().Fatal(err.Error())
		}

		if err := x.AddTab(tui.TabOpts{
			Title:        spr.Components[i].String() + ".out",
			HistoryLimit: 1000,
			Scanner:      bufio.NewScanner(spr.Components[i].Stdout),
		}); err != nil {
			l.Log().Fatal(err.Error())
		}
	}

	if err := spr.SpawnAll(&wg, ctx); err != nil {
		l.Log().Fatalf("unable to spawn component: %s", err.Error())
	}

	if err := x.Start(); err != nil {
		l.Log().Error("error during execution: %s", err.Error())
	}

	if err := spr.KillAll(); err != nil {
		l.Log().Error("error during execution: %s", err.Error())
	}
}
