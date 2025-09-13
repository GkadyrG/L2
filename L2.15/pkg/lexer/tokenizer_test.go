package lexer

import (
	"os"
	"reflect"
	"testing"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "simple command",
			input: "ls -l",
			want:  []string{"ls", "-l"},
		},
		{
			name:  "command with pipe",
			input: "ps aux | grep go",
			want:  []string{"ps", "aux", "|", "grep", "go"},
		},
		{
			name:  "command with redirect",
			input: "echo hello > file.txt",
			want:  []string{"echo", "hello", ">", "file.txt"},
		},
		{
			name:  "command with quotes",
			input: `echo "hello world"`,
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "command with single quotes",
			input: `echo 'hello $USER'`,
			want:  []string{"echo", "hello $USER"},
		},
		{
			name:  "command with AND",
			input: "make && ./app",
			want:  []string{"make", "&&", "./app"},
		},
		{
			name:  "command with OR", 
			input: "test -f file || touch file",
			want:  []string{"test", "-f", "file", "||", "touch", "file"},
		},
		{
			name:  "complex pipeline",
			input: "cat file | grep error | wc -l",
			want:  []string{"cat", "file", "|", "grep", "error", "|", "wc", "-l"},
		},
		{
			name:  "escaped space",
			input: `echo hello\ world`,
			want:  []string{"echo", "hello world"},
		},
		{
			name:    "unclosed quotes",
			input:   `echo "hello`,
			wantErr: true,
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "multiple spaces",
			input: "  ls   -l   ",
			want:  []string{"ls", "-l"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandVars(t *testing.T) {
	os.Setenv("TEST_VAR", "testvalue")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no variables",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "with variable",
			input: "hello $TEST_VAR",
			want:  "hello testvalue",
		},
		{
			name:  "variable in braces",
			input: "hello ${TEST_VAR}",
			want:  "hello testvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandVars(tt.input)
			if got != tt.want {
				t.Errorf("expandVars() = %v, want %v", got, tt.want)
			}
		})
	}
}