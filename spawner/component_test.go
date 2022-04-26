package spawner_test

import (
	"context"
	"testing"

	"github.com/Shanduur/spawner/spawner"
)

func TestExec(t *testing.T) {
	spr, err := spawner.Unmarshal("./test/test_component.yaml")
	if err != nil {
		t.Errorf("unable to parse: %s", err.Error())
	}

	ctx := context.Background()
	for i, cmd := range spr.Components {
		err = cmd.Exec(ctx)
		if err != nil {
			t.Errorf("unable to execute component %d: %s", i, err.Error())
		}
	}
}

func TestPopulate(t *testing.T) {
	spr, err := spawner.Unmarshal("./test/test_component.yaml")
	if err != nil {
		t.Errorf("unable to parse: %s", err.Error())
	}

	if err := spr.Populate(); err != nil {
		t.Errorf("unable to parse: %s", err.Error())
	}
}
