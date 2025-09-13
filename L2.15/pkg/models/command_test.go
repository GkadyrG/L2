package models

import (
	"testing"
)

func TestCmdEmpty(t *testing.T) {
	tests := []struct {
		name string
		cmd  *Cmd
		want bool
	}{
		{
			name: "completely empty",
			cmd:  &Cmd{},
			want: true,
		},
		{
			name: "with binary",
			cmd:  &Cmd{Binary: "ls"},
			want: false,
		},
		{
			name: "with arguments",
			cmd:  &Cmd{Arguments: []string{"-l"}},
			want: false,
		},
		{
			name: "with file operations",
			cmd: &Cmd{
				FileOps: []FileOperation{
					{Operation: ">", Filename: "file.txt"},
				},
			},
			want: false,
		},
		{
			name: "with next pipe",
			cmd: &Cmd{
				NextPipe: &Cmd{Binary: "grep"},
			},
			want: false,
		},
		{
			name: "with next and",
			cmd: &Cmd{
				NextAnd: &Cmd{Binary: "ls"},
			},
			want: false,
		},
		{
			name: "with next or",
			cmd: &Cmd{
				NextOr: &Cmd{Binary: "echo"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.Empty(); got != tt.want {
				t.Errorf("Cmd.Empty() = %v, want %v", got, tt.want)
			}
		})
	}
}
