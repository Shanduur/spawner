package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/Shanduur/spawner/spawner"
)

func main() {
	var wg sync.WaitGroup

	spr, err := spawner.Unmarshal(".spawnfile.yml")
	if err != nil {
		log.Fatalf("unable to parse file: %s", err.Error())
	}
	defer os.RemoveAll(spr.Prefix)

	ctx := context.Background()

	if err := spr.SpawnAll(&wg, ctx); err != nil {
		log.Fatalf("unable to spawn component: %s", err.Error())
	}

	wg.Wait()
}
