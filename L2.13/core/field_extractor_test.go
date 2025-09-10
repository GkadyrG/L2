package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestFieldExtractor_ExtractFields(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		fieldNumbers   []int
		delimiter      rune
		separatedOnly  bool
		expectedResult string
	}{
		{
			name:           "базовое извлечение полей",
			line:           "a:b:c:d:e",
			fieldNumbers:   []int{2, 4},
			delimiter:      ':',
			separatedOnly:  false,
			expectedResult: "b:d",
		},
		{
			name:           "порядок полей не важен",
			line:           "1:2:3:4:5",
			fieldNumbers:   []int{5, 1},
			delimiter:      ':',
			separatedOnly:  false,
			expectedResult: "5:1",
		},
		{
			name:           "пустые поля сохраняются",
			line:           "a::c:d:",
			fieldNumbers:   []int{2, 4},
			delimiter:      ':',
			separatedOnly:  false,
			expectedResult: ":d",
		},
		{
			name:           "поля за пределами диапазона игнорируются",
			line:           "x:y:z",
			fieldNumbers:   []int{1, 5, 2},
			delimiter:      ':',
			separatedOnly:  false,
			expectedResult: "x:y",
		},
		{
			name:           "строка без разделителя при separatedOnly=false",
			line:           "простая строка",
			fieldNumbers:   []int{1, 2},
			delimiter:      ':',
			separatedOnly:  false,
			expectedResult: "простая строка",
		},
		{
			name:           "строка без разделителя при separatedOnly=true",
			line:           "простая строка",
			fieldNumbers:   []int{1, 2},
			delimiter:      ':',
			separatedOnly:  true,
			expectedResult: "",
		},
		{
			name:           "дублирующиеся номера полей",
			line:           "a:b:c:d:e",
			fieldNumbers:   []int{2, 2, 4},
			delimiter:      ':',
			separatedOnly:  false,
			expectedResult: "b:b:d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewFieldExtractor(tt.fieldNumbers, tt.delimiter, tt.separatedOnly)
			result := extractor.ExtractFields(tt.line)

			if result != tt.expectedResult {
				t.Errorf("ожидался результат '%s', получен '%s'", tt.expectedResult, result)
			}
		})
	}
}

func TestFieldExtractor_ProcessStream(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		fieldNumbers   []int
		delimiter      rune
		separatedOnly  bool
		expectedOutput string
	}{
		{
			name:           "обработка нескольких строк",
			input:          "a:b:c\n1:2:3\nx:y:z",
			fieldNumbers:   []int{1, 3},
			delimiter:      ':',
			separatedOnly:  false,
			expectedOutput: "a:c\n1:3\nx:z\n",
		},
		{
			name:           "фильтрация строк без разделителя",
			input:          "a:b:c\nпростая строка\nx:y:z",
			fieldNumbers:   []int{1, 2},
			delimiter:      ':',
			separatedOnly:  true,
			expectedOutput: "a:b\nx:y\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewFieldExtractor(tt.fieldNumbers, tt.delimiter, tt.separatedOnly)

			var output bytes.Buffer
			input := strings.NewReader(tt.input)

			err := extractor.ProcessStream(input, &output)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if output.String() != tt.expectedOutput {
				t.Errorf("ожидался вывод:\n%s\nполучен:\n%s", tt.expectedOutput, output.String())
			}
		})
	}
}
