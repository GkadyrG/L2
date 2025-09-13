package runner

import (
	"context"
	"io"

	"os"
	"strings"
	"testing"
	"time"

	"github.com/GkadyrG/L2/L2.15/pkg/models"
)

func captureStdout(f func()) string {
	// сохраняем оригинальный stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// запускаем функцию
	f()

	// восстанавливаем stdout
	w.Close()
	os.Stdout = oldStdout

	// читаем результат
	var buf strings.Builder
	_, _ = io.Copy(&buf, r) // копируем содержимое пайпа в буфер
	r.Close()

	return buf.String()
}

func TestExecutorBuiltins(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     *models.Cmd
		wantErr bool
	}{
		{
			name: "echo command",
			cmd: &models.Cmd{
				Binary:    "echo",
				Arguments: []string{"hello", "test"},
			},
		},
		{
			name: "pwd command",
			cmd: &models.Cmd{
				Binary: "pwd",
			},
		},
		{
			name: "cd to temp dir",
			cmd: &models.Cmd{
				Binary:    "cd",
				Arguments: []string{os.TempDir()},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.RunCommand(ctx, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutorExternalCommands(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     *models.Cmd
		wantErr bool
	}{
		{
			name: "ls command",
			cmd: &models.Cmd{
				Binary:    "ls",
				Arguments: []string{"-la"},
			},
		},
		{
			name: "date command",
			cmd: &models.Cmd{
				Binary: "date",
			},
		},
		{
			name: "nonexistent command",
			cmd: &models.Cmd{
				Binary: "nonexistent_command_xyz",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.RunCommand(ctx, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutorPipeline(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	// echo hello | tr a-z A-Z
	cmd := &models.Cmd{
		Binary:    "echo",
		Arguments: []string{"hello"},
		NextPipe: &models.Cmd{
			Binary:    "tr",
			Arguments: []string{"a-z", "A-Z"},
		},
	}

	err := executor.RunCommand(ctx, cmd)
	if err != nil {
		t.Errorf("Pipeline execution failed: %v", err)
	}
}

func TestExecutorConditional(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	// true && echo success
	cmdAnd := &models.Cmd{
		Binary: "true",
		NextAnd: &models.Cmd{
			Binary:    "echo",
			Arguments: []string{"success"},
		},
	}

	err := executor.RunCommand(ctx, cmdAnd)
	if err != nil {
		t.Errorf("AND execution failed: %v", err)
	}

	// false || echo fallback
	cmdOr := &models.Cmd{
		Binary: "false",
		NextOr: &models.Cmd{
			Binary:    "echo",
			Arguments: []string{"fallback"},
		},
	}

	err = executor.RunCommand(ctx, cmdOr)
	if err != nil {
		t.Errorf("OR execution failed: %v", err)
	}
}

func TestExecutorTimeout(t *testing.T) {
	executor := NewExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// sleep 5 - должно прерваться по таймауту
	cmd := &models.Cmd{
		Binary:    "sleep",
		Arguments: []string{"5"},
	}

	start := time.Now()
	err := executor.RunCommand(ctx, cmd)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if duration >= 5*time.Second {
		t.Error("Command should have been interrupted by timeout")
	}
}

func TestExecutorFileOperations(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	tempFile := "test_output.txt"
	defer os.Remove(tempFile)

	// echo hello > test_output.txt
	cmd := &models.Cmd{
		Binary:    "echo",
		Arguments: []string{"hello", "world"},
		FileOps: []models.FileOperation{
			{Operation: ">", Filename: tempFile},
		},
	}

	err := executor.RunCommand(ctx, cmd)
	if err != nil {
		t.Errorf("File redirect failed: %v", err)
	}

	// Проверяем содержимое файла
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	expected := "hello world\n"
	if string(content) != expected {
		t.Errorf("File content = %q, want %q", string(content), expected)
	}
}
