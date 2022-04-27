package spawner

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	l "github.com/Shanduur/spawner/logger"
)

type Component struct {
	Entrypoint  []string    `yaml:"entrypoint"`
	Cmd         []string    `yaml:"cmd"`
	KillCmd     []string    `yaml:"kill-cmd"`
	Depends     string      `yaml:"depends"`
	WorkDir     string      `yaml:"workdir"`
	After       []Component `yaml:"after"`
	Before      []Component `yaml:"before"`
	Tee         Tee         `yaml:"tee"`
	SkipPrefix  bool        `yaml:"skip-prefix"`
	PreventKill bool        `yaml:"prevent-kill"`
	ExecCmd     *exec.Cmd
	ContextDir  string
	LogDir      string

	populated bool
	prefix    string

	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

func (cmd Component) String() string {
	var cmdArray []string
	cmdArray = append(cmdArray, cmd.Entrypoint...)
	cmdArray = append(cmdArray, cmd.Cmd...)

	str := strings.Join(cmdArray, "-")
	if len(str) > 16 {
		str = str[:16]
	}

	return str
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

	cmd.prefix = prefix
	if !cmd.SkipPrefix {
		cmd.WorkDir = path.Join(prefix, cmd.WorkDir)
	}

	return nil
}

func (cmd *Component) Kill() {
	defer cmd.Tee.Close()

	for i := 0; i < len(cmd.Before); i++ {
		cmd.Before[i].Kill()
	}

	if len(cmd.KillCmd) > 0 {
		kcmd := Component{
			Entrypoint:  cmd.KillCmd,
			PreventKill: true,
		}

		if err := kcmd.Populate(); err != nil {
			l.Log().Errorf("failed to populate error command: %s", err.Error())
		}

		if err := kcmd.Exec(context.TODO()); err != nil {
			l.Log().Errorf("failed to populate error command: %s", err.Error())
		}
	} else if cmd.ExecCmd.Process != nil && !cmd.PreventKill {
		if err := cmd.ExecCmd.Process.Kill(); err != nil {
			l.Log().Warn(err)
		}
	}

	for i := 0; i < len(cmd.After); i++ {
		cmd.After[i].Kill()
	}
}

func (cmd *Component) Populate() error {
	var err error
	cmd.ContextDir, err = os.Getwd()
	if err != nil {
		return err
	}

	cmd.LogDir = path.Join(cmd.ContextDir, fmt.Sprintf("%s-logs", cmd.prefix))

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

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	var buf bytes.Buffer
	tpl, err := template.New(cmd.String() + "workidr").Parse(cmd.WorkDir)
	if err != nil {
		return err
	}

	if err := tpl.Execute(&buf, cmd); err == nil {
		cmd.WorkDir = buf.String()
	}

	wd, err := filepath.Abs(cmd.WorkDir)
	if err != nil {
		return err
	}
	if len(wd) > len(cmd.WorkDir) {
		cmd.WorkDir = wd
	}
	// l.Log().Info(cmd.WorkDir)

	cd, err := filepath.Abs(cmd.ContextDir)
	if err != nil {
		return err
	}
	if len(cd) > len(cmd.ContextDir) {
		cmd.ContextDir = cd
	}

	ld, err := filepath.Abs(cmd.LogDir)
	if err != nil {
		return err
	}
	if len(cd) > len(cmd.ContextDir) {
		cmd.ContextDir = ld
	}

	if err = os.MkdirAll(cmd.WorkDir, 0777); err != nil {
		return err
	}

	if err = os.MkdirAll(cmd.LogDir, 0777); err != nil {
		return err
	}

	if err = cmd.Tee.Open(path.Join(
		cmd.LogDir,
		strings.ReplaceAll(cmd.String(), " ", "_"),
	)); err != nil {
		return err
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

	cmd.ExecCmd = exec.Command(name, args...)
	cmd.Stdout, err = cmd.ExecCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to create stdout pipe for %s", cmd.String())
	}

	cmd.Stderr, err = cmd.ExecCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to create stderr pipe for %s", cmd.String())
	}

	cmd.ExecCmd.Dir = cmd.WorkDir

	cmd.populated = true

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

	go func() {
		var w io.Writer
		if cmd.Tee.Stdout {
			w = io.MultiWriter(cmd.Tee.StdoutFile)
			_, err := io.Copy(w, cmd.Stdout)
			if err != nil {
				l.Log().Printf("stdout error %s: %s", cmd.String(), err.Error())
			}
		}
	}()

	go func() {
		var w io.Writer
		if cmd.Tee.Stderr {
			w = io.MultiWriter(cmd.Tee.StderrFile)
			_, err := io.Copy(w, cmd.Stderr)
			if err != nil {
				l.Log().Printf("stderr error %s: %s", cmd.String(), err.Error())
			}
		}
	}()

	err = cmd.ExecCmd.Start()
	if err != nil {
		return NewErrExecutionError(cmd.String(), err)
	}

	err = cmd.ExecCmd.Wait()
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
