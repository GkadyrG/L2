package unpack

import "testing"

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		// Простые примеры
		{"a4bc2d5e", "aaaabccddddde", false},
		{"abcd", "abcd", false},
		{"", "", false},

		// Ошибочные случаи
		{"45", "", true},    // только цифры
		{"3abc", "", true},  // начинается с цифры
		{"abc\\", "", true}, // незавершённый escape

		// Экранирование
		{"qwe\\4\\5", "qwe45", false}, // цифры экранированы
		{"qwe\\45", "qwe44444", false},
		{"a\\2b3", "a2bbb", false}, // смешанный кейс
	}

	for _, tt := range tests {
		got, err := Unpack(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("Unpack(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("Unpack(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
