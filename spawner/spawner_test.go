package spawner_test

import (
	"testing"

	"github.com/Shanduur/spawner/spawner"
)

func TestUnmarshall(t *testing.T) {
	spr, err := spawner.Unmarshal("./test/test_component.yaml")
	if err != nil {
		t.Errorf("unable to parse: %s", err.Error())
	}

	if len(spr.Components) != 1 {
		t.Errorf("wrong number of components, expected: %d, got: %d", 1, len(spr.Components))
	}
}
