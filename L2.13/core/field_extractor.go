package core

import (
	"fmt"
	"io"
	"strings"
)

// FieldExtractor обрабатывает строки и извлекает нужные поля
type FieldExtractor struct {
	fieldNumbers  []int
	delimiter     rune
	separatedOnly bool
}

// NewFieldExtractor создает новый экстрактор полей
func NewFieldExtractor(fieldNumbers []int, delimiter rune, separatedOnly bool) *FieldExtractor {
	return &FieldExtractor{
		fieldNumbers:  fieldNumbers,
		delimiter:     delimiter,
		separatedOnly: separatedOnly,
	}
}

// ProcessStream обрабатывает поток данных
func (fe *FieldExtractor) ProcessStream(input io.Reader, output io.Writer) error {
	lines := make([]string, 0)

	// Читаем все строки
	scanner := NewLineScanner(input)
	for scanner.HasNext() {
		line := scanner.Next()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения входных данных: %w", err)
	}

	// Обрабатываем каждую строку
	for _, line := range lines {
		result := fe.ExtractFields(line)
		if result != "" {
			fmt.Fprintln(output, result)
		}
	}

	return nil
}

// ExtractFields извлекает поля из одной строки
func (fe *FieldExtractor) ExtractFields(line string) string {
	// Проверяем наличие разделителя
	if !strings.ContainsRune(line, fe.delimiter) {
		if fe.separatedOnly {
			return "" // Пропускаем строки без разделителя
		}
		return line // Возвращаем строку как есть
	}

	// Разбиваем на поля
	fields := strings.Split(line, string(fe.delimiter))

	// Собираем нужные поля
	var resultFields []string
	for _, fieldNum := range fe.fieldNumbers {
		if fieldNum > 0 && fieldNum <= len(fields) {
			resultFields = append(resultFields, fields[fieldNum-1])
		}
	}

	return strings.Join(resultFields, string(fe.delimiter))
}
