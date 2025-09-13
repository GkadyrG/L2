package commands

import (
	"bytes"
	"os"

	"testing"

	"github.com/GkadyrG/L2/L2.15/pkg/environment"
)

func TestChangeDirectory(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "test_cd")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	env := environment.NewSystemEnv()
	cd := &ChangeDirectory{}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "change to temp dir",
			args: []string{tempDir},
		},
		{
			name: "change to home",
			args: []string{},
		},
		{
			name:    "too many args",
			args:    []string{"dir1", "dir2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cd.Run(tt.args, env, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeDirectory.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEcho(t *testing.T) {
	echo := &Echo{}
	env := environment.NewSystemEnv()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple echo",
			args:     []string{"hello", "world"},
			expected: "hello world\n",
		},
		{
			name:     "empty echo",
			args:     []string{},
			expected: "\n",
		},
		{
			name:     "single arg",
			args:     []string{"test"},
			expected: "test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := echo.Run(tt.args, env, nil, &buf)
			if err != nil {
				t.Errorf("Echo.Run() error = %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("Echo.Run() output = %q, want %q", buf.String(), tt.expected)
			}
		})
	}
}

func TestPrintWorkingDirectory(t *testing.T) {
	pwd := &PrintWorkingDirectory{}
	env := environment.NewSystemEnv()

	var buf bytes.Buffer
	err := pwd.Run([]string{}, env, nil, &buf)
	if err != nil {
		t.Errorf("PrintWorkingDirectory.Run() error = %v", err)
	}

	expectedDir, _ := os.Getwd()
	expectedDir += "\n"

	if buf.String() != expectedDir {
		t.Errorf("PrintWorkingDirectory.Run() output = %q, want %q", buf.String(), expectedDir)
	}
}

func TestCommandRegistry(t *testing.T) {
	registry := NewRegistry()

	// Проверяем, что встроенные команды зарегистрированы
	builtins := []string{"cd", "pwd", "echo", "ps", "kill"}
	for _, name := range builtins {
		if _, exists := registry.Get(name); !exists {
			t.Errorf("Builtin command %s not registered", name)
		}
	}

	// Проверяем несуществующую команду
	if _, exists := registry.Get("nonexistent"); exists {
		t.Error("Nonexistent command should not be found")
	}
}
