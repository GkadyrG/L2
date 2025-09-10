package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// FieldSpecification определяет спецификацию полей
type FieldSpecification struct {
	Numbers []int
}

// ParseFieldSpec парсит строку спецификации полей
func ParseFieldSpec(spec string) (*FieldSpecification, error) {
	if spec == "" {
		return nil, fmt.Errorf("спецификация полей не может быть пустой")
	}

	fs := &FieldSpecification{
		Numbers: make([]int, 0),
	}

	// Разбиваем по запятым
	parts := strings.Split(spec, ",")
	seen := make(map[int]bool)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Обрабатываем диапазоны и отдельные числа
		numbers, err := parseFieldPart(part)
		if err != nil {
			return nil, fmt.Errorf("ошибка в части '%s': %w", part, err)
		}

		// Добавляем числа, избегая дубликатов
		for _, num := range numbers {
			if !seen[num] {
				fs.Numbers = append(fs.Numbers, num)
				seen[num] = true
			}
		}
	}

	if len(fs.Numbers) == 0 {
		return nil, fmt.Errorf("не найдено ни одного валидного номера поля")
	}

	return fs, nil
}

// parseFieldPart парсит часть спецификации (число или диапазон)
func parseFieldPart(part string) ([]int, error) {
	if strings.Contains(part, "-") {
		return parseRange(part)
	}
	return parseSingleNumber(part)
}

// parseRange парсит диапазон чисел (например, "3-7")
func parseRange(rangeStr string) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("неверный формат диапазона: %s", rangeStr)
	}

	start, err := parseSingleNumber(parts[0])
	if err != nil {
		return nil, fmt.Errorf("неверное начальное значение диапазона: %w", err)
	}

	end, err := parseSingleNumber(parts[1])
	if err != nil {
		return nil, fmt.Errorf("неверное конечное значение диапазона: %w", err)
	}

	if start[0] > end[0] {
		return nil, fmt.Errorf("начальное значение диапазона больше конечного: %s", rangeStr)
	}

	// Генерируем все числа в диапазоне
	var result []int
	for i := start[0]; i <= end[0]; i++ {
		result = append(result, i)
	}

	return result, nil
}

// parseSingleNumber парсит одно число
func parseSingleNumber(numStr string) ([]int, error) {
	num, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil {
		return nil, fmt.Errorf("неверное число: %s", numStr)
	}

	if num < 1 {
		return nil, fmt.Errorf("номера полей должны быть положительными числами")
	}

	return []int{num}, nil
}
