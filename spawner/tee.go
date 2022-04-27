package spawner

import (
	"os"

	l "github.com/Shanduur/spawner/logger"
)

type Tee struct {
	Stdout     bool `yaml:"stdout"`
	Stderr     bool `yaml:"stderr"`
	StderrFile *os.File
	StdoutFile *os.File
}

func (t *Tee) Open(name string) error {
	if t.Stdout {
		f, err := os.OpenFile(name+".log", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		t.StdoutFile = f
	}

	if t.Stderr {
		f, err := os.OpenFile(name+".err", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		t.StderrFile = f
	}

	return nil
}

func (t *Tee) Close() {
	if t.Stdout {
		if err := t.StdoutFile.Close(); err != nil {
			l.Log().Errorf("error closing stdout file: %s", err)
		}
	}
	if t.Stderr {
		if err := t.StderrFile.Close(); err != nil {
			l.Log().Errorf("error closing stderr file: %s", err)
		}
	}
}
