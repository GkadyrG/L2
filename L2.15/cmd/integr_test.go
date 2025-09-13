package main

import (
	"context"

	"testing"
	"time"

	"github.com/GkadyrG/L2/L2.15/pkg/lexer"
	"github.com/GkadyrG/L2/L2.15/pkg/parser"
	"github.com/GkadyrG/L2/L2.15/pkg/runner"
)

func TestFullIntegration(t *testing.T) {
	executor := runner.NewExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "simple echo",
			input: "echo hello world",
		},
		{
			name:  "pwd command",
			input: "pwd",
		},
		{
			name:  "simple pipeline",
			input: "echo hello | tr a-z A-Z",
		},
		{
			name:  "conditional success",
			input: "true && echo success",
		},
		{
			name:  "conditional fallback",
			input: "false || echo fallback",
		},
		{
			name:    "invalid command",
			input:   "nonexistent_xyz_command",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Парсим входную строку
			tokens, err := lexer.ParseInput(tc.input)
			if err != nil {
				t.Fatalf("Tokenization failed: %v", err)
			}

			// Строим команду
			cmd, err := parser.BuildCommand(tokens)
			if err != nil {
				t.Fatalf("Command building failed: %v", err)
			}

			// Выполняем
			err = executor.RunCommand(ctx, cmd)
			if (err != nil) != tc.wantErr {
				t.Errorf("Execution error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
