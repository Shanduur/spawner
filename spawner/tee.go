package spawner

import "os"

type Tee struct {
	Stdout     bool `yaml:"stdout"`
	Stderr     bool `yaml:"stderr"`
	Combined   bool `yaml:"combined"`
	StderrFile *os.File
	StdoutFile *os.File
}

func (t *Tee) Open(name string) error {
	if t.Combined {
		f, err := os.OpenFile(name+".log", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		t.StdoutFile = f
		t.StderrFile = f
		return nil
	}

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
		t.StdoutFile.Close()
	}
	if t.Stderr {
		t.StderrFile.Close()
	}
	if t.Combined {
		t.StdoutFile.Close()
	}
}
