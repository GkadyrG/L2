package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/GkadyrG/L2/L2.13/core"
	"github.com/GkadyrG/L2/L2.13/parser"
	"github.com/GkadyrG/L2/L2.13/utils"
)

func main() {
	// Парсим аргументы командной строки
	config, err := parser.ParseCommandLine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга аргументов: %v\n", err)
		os.Exit(1)
	}

	// Парсим спецификацию полей
	fieldSpec, err := parser.ParseFieldSpec(config.FieldSpec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга полей: %v\n", err)
		os.Exit(1)
	}

	// Проверяем разделитель
	if len(config.Delimiter) != 1 {
		fmt.Fprintf(os.Stderr, "Разделитель должен быть одним символом\n")
		os.Exit(1)
	}
	delimiter := rune(config.Delimiter[0])

	// Создаем экстрактор полей
	extractor := core.NewFieldExtractor(fieldSpec.Numbers, delimiter, config.SeparatedOnly)

	// Создаем процессор файлов
	fileProcessor := &utils.FileProcessor{}

	// Определяем функцию обработки
	processFunc := func(input io.Reader, output io.Writer) error {
		return extractor.ProcessStream(input, output)
	}

	// Обрабатываем входные данные
	if len(config.Files) == 0 {
		// Обрабатываем stdin
		err = fileProcessor.ProcessStdin(processFunc)
		if err != nil {
			log.Fatalf("Ошибка обработки stdin: %v", err)
		}
	} else {
		// Обрабатываем файлы
		fileProcessor.ProcessFiles(config.Files, processFunc)
	}
}
