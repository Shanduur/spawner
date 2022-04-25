package spawner_test

import (
	"context"
	"os"
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

	for _, cmd := range spr.Components {
		if cmd.Stderr != os.Stderr {
			t.Errorf("cmd.Stderr and os.Stderr not equal for %s, got: %v, wanted: %v", cmd.String(), cmd.Stderr, os.Stderr)
		}

		if cmd.Stdout != os.Stdout {
			t.Errorf("cmd.Stdout and os.Stdout not equal for %s, got: %v, wanted: %v", cmd.String(), cmd.Stderr, os.Stderr)
		}

		if cmd.Stdin != os.Stdin {
			t.Errorf("cmd.Stdin and os.Stdin not equal for %s, got: %v, wanted: %v", cmd.String(), cmd.Stderr, os.Stderr)
		}
	}
}
