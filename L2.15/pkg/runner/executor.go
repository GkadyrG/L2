package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/GkadyrG/L2/L2.15/pkg/commands"
	"github.com/GkadyrG/L2/L2.15/pkg/environment"
	"github.com/GkadyrG/L2/L2.15/pkg/models"
)

type Executor struct {
	registry    *commands.CommandRegistry
	environment environment.Env
	input       io.Reader
	output      io.Writer
	errOutput   io.Writer

	mutex      sync.Mutex
	activeProc *os.Process
	openFiles  []io.Closer
}

func NewExecutor() *Executor {
	return &Executor{
		registry:    commands.NewRegistry(),
		environment: environment.NewSystemEnv(),
		input:       os.Stdin,
		output:      os.Stdout,
		errOutput:   os.Stderr,
		openFiles:   make([]io.Closer, 0),
	}
}

func (e *Executor) RunCommand(ctx context.Context, cmd *models.Cmd) error {
	if cmd == nil || cmd.Binary == "" {
		return fmt.Errorf("пустая команда")
	}

	// Настройка файловых операций
	if err := e.setupFiles(cmd); err != nil {
		return err
	}
	defer e.cleanup()

	// Выбор типа выполнения
	switch {
	case cmd.NextPipe != nil:
		return e.runPipeline(ctx, cmd)
	case cmd.NextAnd != nil:
		if err := e.execSingle(ctx, cmd); err != nil {
			return err
		}
		return e.RunCommand(ctx, cmd.NextAnd)

	case cmd.NextOr != nil:
		if err := e.execSingle(ctx, cmd); err != nil {
			return e.RunCommand(ctx, cmd.NextOr)
		}
		return nil
	default:
		return e.execSingle(ctx, cmd)
	}
}

func (e *Executor) setupFiles(cmd *models.Cmd) error {
	e.cleanup()
	e.input, e.output = os.Stdin, os.Stdout

	for _, fileOp := range cmd.FileOps {
		switch fileOp.Operation {
		case "<":
			f, err := os.Open(fileOp.Filename)
			if err != nil {
				return fmt.Errorf("не удалось открыть файл для чтения: %v", err)
			}
			e.openFiles = append(e.openFiles, f)
			e.input = f

		case ">":
			f, err := os.OpenFile(fileOp.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("не удалось открыть файл для записи: %v", err)
			}
			e.openFiles = append(e.openFiles, f)
			e.output = f
		}
	}

	return nil
}

func (e *Executor) runPipeline(ctx context.Context, cmd *models.Cmd) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("ошибка создания пайпа: %v", err)
	}

	errCh := make(chan error, 2)

	// Левая часть пайпа
	go func() {
		defer pw.Close()
		leftExec := &Executor{
			registry:    e.registry,
			environment: e.environment,
			input:       e.input,
			output:      pw,
			errOutput:   e.errOutput,
		}
		errCh <- leftExec.execSingle(ctx, cmd)
	}()

	// Правая часть пайпа
	go func() {
		defer pr.Close()
		rightExec := &Executor{
			registry:    e.registry,
			environment: e.environment,
			input:       pr,
			output:      e.output,
			errOutput:   e.errOutput,
		}
		errCh <- rightExec.RunCommand(ctx, cmd.NextPipe)
	}()

	// Ожидаем завершения обеих частей
	var firstErr error
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (e *Executor) execSingle(ctx context.Context, cmd *models.Cmd) error {
	if cmd.Empty() {
		return nil
	}

	// Проверяем встроенные команды
	if handler, exists := e.registry.Get(cmd.Binary); exists {
		return handler.Run(cmd.Arguments, e.environment, e.input, e.output)
	}

	// Внешняя команда
	proc := exec.CommandContext(ctx, cmd.Binary, cmd.Arguments...)
	proc.Stdin = e.input
	proc.Stdout = e.output
	proc.Stderr = e.errOutput

	e.mutex.Lock()
	if err := proc.Start(); err != nil {
		e.mutex.Unlock()
		return fmt.Errorf("не удалось запустить %s: %v", cmd.Binary, err)
	}
	e.activeProc = proc.Process
	e.mutex.Unlock()

	defer func() {
		e.mutex.Lock()
		e.activeProc = nil
		e.mutex.Unlock()
	}()

	return proc.Wait()
}

func (e *Executor) Stop() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.activeProc != nil {
		e.activeProc.Signal(os.Interrupt)
	}
}

func (e *Executor) cleanup() {
	for _, f := range e.openFiles {
		f.Close()
	}
	e.openFiles = e.openFiles[:0]
}

func (e *Executor) CurrentDirectory() (string, error) {
	return e.environment.(*environment.SystemEnv).DirName()
}
