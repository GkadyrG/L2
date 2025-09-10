package parser

import (
	"os"
	"testing"
)

func TestParseFieldSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		expected    []int
		expectError bool
	}{
		{
			name:        "отдельные числа",
			spec:        "1,3,5",
			expected:    []int{1, 3, 5},
			expectError: false,
		},
		{
			name:        "диапазон",
			spec:        "2-4",
			expected:    []int{2, 3, 4},
			expectError: false,
		},
		{
			name:        "смешанный формат",
			spec:        "1,3-5,7",
			expected:    []int{1, 3, 4, 5, 7},
			expectError: false,
		},
		{
			name:        "дублирующиеся числа",
			spec:        "1,2,1,3",
			expected:    []int{1, 2, 3},
			expectError: false,
		},
		{
			name:        "пустая спецификация",
			spec:        "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "неверный диапазон",
			spec:        "5-2",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "неверное число",
			spec:        "abc",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "отрицательное число",
			spec:        "-1",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "ноль",
			spec:        "0",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFieldSpec(tt.spec)

			if tt.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но её не было")
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if len(result.Numbers) != len(tt.expected) {
				t.Errorf("ожидалось %d чисел, получено %d", len(tt.expected), len(result.Numbers))
				return
			}

			for i, num := range result.Numbers {
				if num != tt.expected[i] {
					t.Errorf("на позиции %d ожидалось %d, получено %d", i, tt.expected[i], num)
				}
			}
		})
	}
}

func TestParseCommandLine(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "валидные аргументы с файлом",
			args:        []string{"program", "-f", "1,3", "file.txt"},
			expectError: false,
		},
		{
			name:        "валидные аргументы без файла",
			args:        []string{"program", "-f", "1-3", "-d", ","},
			expectError: false,
		},
		{
			name:        "отсутствует флаг -f",
			args:        []string{"program", "file.txt"},
			expectError: true,
		},
		{
			name:        "флаг -f без значения",
			args:        []string{"program", "-f"},
			expectError: true,
		},
		{
			name:        "неизвестный флаг",
			args:        []string{"program", "-x", "value"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем тестовые аргументы
			os.Args = tt.args

			config, err := ParseCommandLine()

			if tt.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но её не было")
				}
			} else {
				if err != nil {
					t.Errorf("неожиданная ошибка: %v", err)
				}
				if config == nil {
					t.Errorf("конфигурация не должна быть nil")
				}
			}
		})
	}

	// Восстанавливаем оригинальные аргументы
	os.Args = originalArgs
}
