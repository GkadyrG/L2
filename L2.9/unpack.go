package unpack

import (
	"fmt"
	"strconv"
	"strings"
)

// Unpack распаковывает строку вида "a4bc2d5e" -> "aaaabccddddde".
// Поддерживает экранирование через \.
// Возвращает ошибку, если строка некорректна.
func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var sb strings.Builder
	var prev rune
	escaped := false

	for i, r := range input {
		if escaped {
			sb.WriteRune(r)
			prev = r
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if r >= '0' && r <= '9' {
			if i == 0 {
				return "", fmt.Errorf("string starts with a number")
			}
			count, err := strconv.Atoi(string(r))
			if err != nil {
				return "", fmt.Errorf("failed to convert string to int: %w", err)
			}
			for j := 1; j < count; j++ { // prev уже есть
				sb.WriteRune(prev)
			}
			continue
		}

		// обычный символ
		sb.WriteRune(r)
		prev = r
	}

	if escaped {
		return "", fmt.Errorf("unfinished escape sequence")
	}

	return sb.String(), nil
}
