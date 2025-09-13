package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/GkadyrG/L2/L2.15/pkg/environment"
)

// Команда cd
type ChangeDirectory struct{}

func (cd *ChangeDirectory) Name() string { return "cd" }

func (cd *ChangeDirectory) Run(args []string, env environment.Env, _ io.Reader, _ io.Writer) error {
	switch len(args) {
	case 0:
		home, err := env.HomeDir()
		if err != nil {
			return err
		}
		return env.ChangeDir(home)
	case 1:
		return env.ChangeDir(args[0])
	default:
		return fmt.Errorf("cd: слишком много аргументов")
	}
}

// Команда pwd
type PrintWorkingDirectory struct{}

func (pwd *PrintWorkingDirectory) Name() string { return "pwd" }

func (pwd *PrintWorkingDirectory) Run(_ []string, env environment.Env, _ io.Reader, out io.Writer) error {
	dir, err := env.CurrentDir()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, dir)
	return nil
}

// Команда echo
type Echo struct{}

func (e *Echo) Name() string { return "echo" }

func (e *Echo) Run(args []string, _ environment.Env, _ io.Reader, out io.Writer) error {
	fmt.Fprintln(out, strings.Join(args, " "))
	return nil
}

// Команда ps
type ProcessList struct{}

func (ps *ProcessList) Name() string { return "ps" }

func (ps *ProcessList) Run(args []string, _ environment.Env, _ io.Reader, out io.Writer) error {
	cmd := exec.Command("ps", args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Команда kill
type KillProcess struct{}

func (k *KillProcess) Name() string { return "kill" }

func (k *KillProcess) Run(args []string, _ environment.Env, _ io.Reader, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("использование: kill PID")
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("неверный PID: %s", args[0])
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("процесс %d не найден", pid)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("ошибка завершения процесса %d: %v", pid, err)
	}

	fmt.Fprintf(out, "Процесс %d завершен\n", pid)
	return nil
}
