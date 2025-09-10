package utils

import (
	"fmt"
	"io"
	"os"
)

// FileProcessor обрабатывает файлы
type FileProcessor struct{}

// ProcessFile обрабатывает один файл
func (fp *FileProcessor) ProcessFile(filename string, processor func(io.Reader, io.Writer) error) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл '%s': %w", filename, err)
	}
	defer file.Close()

	return processor(file, os.Stdout)
}

// ProcessFiles обрабатывает несколько файлов
func (fp *FileProcessor) ProcessFiles(filenames []string, processor func(io.Reader, io.Writer) error) {
	for _, filename := range filenames {
		err := fp.ProcessFile(filename, processor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка обработки файла '%s': %v\n", filename, err)
		}
	}
}

// ProcessStdin обрабатывает стандартный ввод
func (fp *FileProcessor) ProcessStdin(processor func(io.Reader, io.Writer) error) error {
	// Проверяем, что stdin не является терминалом
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("ошибка проверки stdin: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("нет входных данных (используйте файл или перенаправление)")
	}

	return processor(os.Stdin, os.Stdout)
}
