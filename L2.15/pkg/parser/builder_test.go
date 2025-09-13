package parser

import (
	"reflect"
	"testing"

	"github.com/GkadyrG/L2/L2.15/pkg/models"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []string
		want    *models.Cmd
		wantErr bool
	}{
		{
			name:   "simple command",
			tokens: []string{"ls", "-l"},
			want: &models.Cmd{
				Binary:    "ls",
				Arguments: []string{"-l"},
			},
		},
		{
			name:   "command with redirect",
			tokens: []string{"echo", "hello", ">", "file.txt"},
			want: &models.Cmd{
				Binary:    "echo",
				Arguments: []string{"hello"},
				FileOps: []models.FileOperation{
					{Operation: ">", Filename: "file.txt"},
				},
			},
		},
		{
			name:   "command with pipe",
			tokens: []string{"ls", "|", "grep", "go"},
			want: &models.Cmd{
				Binary: "ls",
				NextPipe: &models.Cmd{
					Binary:    "grep",
					Arguments: []string{"go"},
				},
			},
		},
		{
			name:   "command with AND",
			tokens: []string{"make", "&&", "ls"},
			want: &models.Cmd{
				Binary: "make",
				NextAnd: &models.Cmd{
					Binary: "ls",
				},
			},
		},
		{
			name:   "command with OR",
			tokens: []string{"test", "-f", "file", "||", "touch", "file"},
			want: &models.Cmd{
				Binary:    "test",
				Arguments: []string{"-f", "file"},
				NextOr: &models.Cmd{
					Binary:    "touch",
					Arguments: []string{"file"},
				},
			},
		},
		{
			name:   "complex pipeline",
			tokens: []string{"cat", "file", "|", "grep", "error", "|", "wc", "-l"},
			want: &models.Cmd{
				Binary:    "cat",
				Arguments: []string{"file"},
				NextPipe: &models.Cmd{
					Binary:    "grep",
					Arguments: []string{"error"},
					NextPipe: &models.Cmd{
						Binary:    "wc",
						Arguments: []string{"-l"},
					},
				},
			},
		},
		{
			name:    "empty tokens",
			tokens:  []string{},
			wantErr: true,
		},
		{
			name:    "pipe without command",
			tokens:  []string{"|", "grep"},
			wantErr: true,
		},
		{
			name:    "redirect without file",
			tokens:  []string{"echo", ">"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildCommand(tt.tokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !commandsEqual(got, tt.want) {
				t.Errorf("BuildCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func commandsEqual(a, b *models.Cmd) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Binary != b.Binary {
		return false
	}
	if !reflect.DeepEqual(a.Arguments, b.Arguments) {
		return false
	}
	if !reflect.DeepEqual(a.FileOps, b.FileOps) {
		return false
	}
	return commandsEqual(a.NextPipe, b.NextPipe) &&
		commandsEqual(a.NextAnd, b.NextAnd) &&
		commandsEqual(a.NextOr, b.NextOr)
}
