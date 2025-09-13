package terminal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GkadyrG/L2/L2.15/pkg/lexer"
	"github.com/GkadyrG/L2/L2.15/pkg/parser"
	"github.com/GkadyrG/L2/L2.15/pkg/runner"
)

type Shell struct {
	executor *runner.Executor
	active   bool
}

func New() *Shell {
	return &Shell{
		executor: runner.NewExecutor(),
		active:   true,
	}
}

func (s *Shell) Start() {
	// Обработка сигналов
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	reader := bufio.NewReader(os.Stdin)

	for s.active {
		select {
		case sig := <-signals:
			s.handleSignal(sig)
		default:
			if !s.processInput(reader) {
				return
			}
		}
	}
}

func (s *Shell) handleSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGINT:
		fmt.Fprintln(os.Stderr, "\n^C")
		s.executor.Stop()
	case syscall.SIGTERM:
		fmt.Println("\nВыход...")
		s.active = false
	}
}

func (s *Shell) processInput(reader *bufio.Reader) bool {
	// Приглашение
	dirName, err := s.executor.CurrentDirectory()
	if err != nil {
		dirName = "unknown"
	}
	fmt.Printf("\033[32m%s\033[0m$ ", dirName)

	// Чтение команды
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nДо свидания!")
			return false
		}
		fmt.Fprintf(os.Stderr, "Ошибка чтения: %v\n", err)
		return true
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return true
	}

	// Парсинг
	tokens, err := lexer.ParseInput(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка разбора: %v\n", err)
		return true
	}

	cmd, err := parser.BuildCommand(tokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга: %v\n", err)
		return true
	}

	// Выполнение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.executor.RunCommand(ctx, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка выполнения: %v\n", err)
	}

	return true
}
