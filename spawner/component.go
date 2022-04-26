package spawner

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Component struct {
	Entrypoint []string    `yaml:"entrypoint"`
	Cmd        []string    `yaml:"cmd"`
	Depends    string      `yaml:"depends"`
	WorkDir    string      `yaml:"workdir"`
	After      []Component `yaml:"after"`
	Before     []Component `yaml:"before"`
	Tee        Tee         `yaml:"tee"`

	populated bool
	prefix    string

	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func (cmd Component) String() string {
	var cmdArray []string
	cmdArray = append(cmdArray, cmd.Entrypoint...)
	cmdArray = append(cmdArray, cmd.Cmd...)

	return strings.Join(cmdArray, " ")
}

func (cmd *Component) AddPrefix(prefix string) error {
	for i := 0; i < len(cmd.Before); i++ {
		if err := cmd.Before[i].AddPrefix(prefix); err != nil {
			return err
		}
	}

	for i := 0; i < len(cmd.After); i++ {
		if err := cmd.After[i].AddPrefix(prefix); err != nil {
			return err
		}
	}

	cmd.WorkDir = path.Join(prefix, cmd.WorkDir)
	cmd.prefix = prefix

	return nil
}

func (cmd *Component) Populate() error {
	for i := 0; i < len(cmd.Before); i++ {
		if err := cmd.Before[i].Populate(); err != nil {
			return err
		}
	}

	for i := 0; i < len(cmd.After); i++ {
		if err := cmd.After[i].Populate(); err != nil {
			return err
		}
	}

	cmd.populated = true

	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	wd, err := filepath.Abs(cmd.WorkDir)
	if err != nil {
		return err
	}
	if len(wd) > len(cmd.WorkDir) {
		cmd.WorkDir = wd
	}

	if err = os.MkdirAll(cmd.WorkDir, 0777); err != nil {
		return err
	}

	if err = cmd.Tee.Open(path.Join(
		cmd.prefix,
		strings.ReplaceAll(cmd.String(), " ", "_"),
	)); err != nil {
		return err
	}

	return nil
}

func NewComponent(entrypoint string) *Component {
	return &Component{
		Entrypoint: []string{entrypoint},
	}
}

type ErrExecutionError struct {
	argEntrypoint string
	reason        error
}

func NewErrExecutionError(argEntrypoint string, err error) ErrExecutionError {
	return ErrExecutionError{
		argEntrypoint: argEntrypoint,
		reason:        err,
	}
}

func (err ErrExecutionError) Error() string {
	return fmt.Sprintf("execution of %s failed, reason: %s", err.argEntrypoint, err.reason)
}

func (cmd Component) ArrayExpand(array []string) ([]string, error) {
	for i := 0; i < len(array); i++ {
		var buf bytes.Buffer
		tpl, err := template.New(cmd.String()).Parse(array[i])
		if err != nil {
			return []string{}, err
		}

		if err := tpl.Execute(&buf, cmd); err == nil {
			array[i] = buf.String()
		}
	}

	return array, nil
}

func (cmd *Component) Exec(ctx context.Context) error {
	var err error

	if !cmd.populated {
		err := cmd.Populate()
		if err != nil {
			return fmt.Errorf("error during populating component %s: %w", cmd.String(), err)
		}
	}

	for _, beforeCmd := range cmd.Before {
		err := beforeCmd.Exec(ctx)
		if err != nil {
			return err
		}
	}

	if len(cmd.WorkDir) > 0 {
		err = os.MkdirAll(cmd.WorkDir, 0777)
		if err != nil {
			return fmt.Errorf("problem with WorkDir: %w", err)
		}
	}

	var componentArray []string
	componentArray = append(componentArray, cmd.Entrypoint...)
	componentArray = append(componentArray, cmd.Cmd...)

	if len(componentArray) <= 0 {
		return fmt.Errorf("neither entrypoint nor component provided")
	}

	componentArray, err = cmd.ArrayExpand(componentArray)
	if err != nil {
		return fmt.Errorf("unable to expand array: %w", err)
	}

	name := componentArray[0]
	var args []string
	if len(componentArray) > 1 {
		args = append(args, componentArray[1:]...)
	}

	ex := exec.Command(name, args...)
	stdout, err := ex.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to create stdout pipe for %s", cmd.String())
	}

	go func() {
		var w io.Writer
		if cmd.Tee.Combined || cmd.Tee.Stdout {
			w = io.MultiWriter(cmd.Tee.StdoutFile, os.Stdout)
		} else {
			w = os.Stdout
		}

		_, err := io.Copy(w, stdout)
		if err != nil {
			log.Printf("stdout error %s: %s", cmd.String(), err.Error())
		}
	}()

	stderr, err := ex.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to create stderr pipe for %s", cmd.String())
	}

	go func() {
		var w io.Writer
		if cmd.Tee.Combined || cmd.Tee.Stderr {
			w = io.MultiWriter(cmd.Tee.StderrFile, os.Stderr)
		} else {
			w = os.Stdout
		}
		_, err := io.Copy(w, stderr)
		if err != nil {
			log.Printf("stderr error %s: %s", cmd.String(), err.Error())
		}
	}()

	ex.Dir = cmd.WorkDir

	err = ex.Start()
	if err != nil {
		return NewErrExecutionError(cmd.String(), err)
	}

	err = ex.Wait()
	if err != nil {
		return NewErrExecutionError(cmd.String(), err)
	}

	for _, afterCmd := range cmd.After {
		err := afterCmd.Exec(ctx)
		if err != nil {
			return err
		}
	}

	return err
}

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
