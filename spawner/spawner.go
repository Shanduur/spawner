package spawner

import (
	"context"
	"os"
	"sync"

	l "github.com/Shanduur/spawner/logger"
	"gopkg.in/yaml.v2"
)

type Spawner struct {
	Prefix     string      `yaml:"spawndir"`
	Components []Component `yaml:"components"`
}

func (spr Spawner) Spawn(cmd Component, wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cmd.Tee.Close()
		if err := cmd.Exec(ctx); err != nil {
			l.Log().Printf("component failed: %s", err.Error())
		}
	}()
}

func (spr *Spawner) SpawnAll(wg *sync.WaitGroup, ctx context.Context) error {
	for _, cmd := range spr.Components {
		spr.Spawn(cmd, wg, ctx)
	}

	return nil
}

func (spr *Spawner) KillAll() error {
	for _, cmd := range spr.Components {
		cmd.Kill()
	}

	return nil
}

func (spr *Spawner) Populate() error {
	for i := 0; i < len(spr.Components); i++ {
		if err := spr.Components[i].AddPrefix(spr.Prefix); err != nil {
			return err
		}
	}

	for i := 0; i < len(spr.Components); i++ {
		if err := spr.Components[i].Populate(); err != nil {
			return err
		}
	}

	return nil
}

func Unmarshal(file string) (Spawner, error) {
	b, err := os.ReadFile(file)

	var spr Spawner
	if err = yaml.Unmarshal(b, &spr); err != nil {
		return spr, err
	}

	os.RemoveAll(spr.Prefix)

	return spr, nil
}
