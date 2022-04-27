package spawner

import (
	"os"

	l "github.com/Shanduur/spawner/logger"
)

type Tee struct {
	Stdout         bool `yaml:"stdout"`
	Stderr         bool `yaml:"stderr"`
	stdoutFileName string
	stderrFileName string
	StderrFile     *os.File
	StdoutFile     *os.File
}

func (t *Tee) Open(name string) error {
	if t.Stdout {
		t.stdoutFileName = name + ".log"
		f, err := os.OpenFile(t.stdoutFileName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		t.StdoutFile = f
	}

	if t.Stderr {
		t.stderrFileName = name + ".err"
		f, err := os.OpenFile(t.stderrFileName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		t.StderrFile = f
	}

	return nil
}

func (t *Tee) Close() {
	l.Log().Infof("close called for %s | %s", t.stderrFileName, t.stdoutFileName)
	if t.Stdout {
		if err := t.StdoutFile.Close(); err != nil {
			l.Log().Errorf("error closing %s: %s", t.stdoutFileName, err)
		}
	}
	if t.Stderr {
		if err := t.StderrFile.Close(); err != nil {
			l.Log().Errorf("error closing %s: %s", t.stderrFileName, err)
		}
	}
}
