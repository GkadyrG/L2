package parser

import (
	"fmt"
	"os"
	"strings"
)

// CommandLineConfig содержит конфигурацию командной строки
type CommandLineConfig struct {
	FieldSpec     string
	Delimiter     string
	SeparatedOnly bool
	Files         []string
}

// ParseCommandLine парсит аргументы командной строки
func ParseCommandLine() (*CommandLineConfig, error) {
	config := &CommandLineConfig{}

	// Простой парсер аргументов без внешних библиотек
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-f", "--fields":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("флаг -f требует значение")
			}
			config.FieldSpec = args[i+1]
			i++ // Пропускаем следующий аргумент

		case "-d", "--delimiter":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("флаг -d требует значение")
			}
			config.Delimiter = args[i+1]
			i++ // Пропускаем следующий аргумент

		case "-s", "--separated":
			config.SeparatedOnly = true

		case "--help", "-h":
			printUsage()
			os.Exit(0)

		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("неизвестный флаг: %s", arg)
			}
			// Это файл
			config.Files = append(config.Files, arg)
		}
	}

	// Устанавливаем значения по умолчанию
	if config.Delimiter == "" {
		config.Delimiter = "\t"
	}

	// Проверяем обязательные параметры
	if config.FieldSpec == "" {
		return nil, fmt.Errorf("обязательно указать номера полей с помощью флага -f")
	}

	return config, nil
}

// printUsage выводит справку по использованию
func printUsage() {
	fmt.Fprintf(os.Stderr, "Использование: %s [ОПЦИИ] [ФАЙЛЫ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nОпции:\n")
	fmt.Fprintf(os.Stderr, "  -f, --fields СПЕЦИФИКАЦИЯ    Номера полей для вывода (например: 1,3-5)\n")
	fmt.Fprintf(os.Stderr, "  -d, --delimiter РАЗДЕЛИТЕЛЬ  Разделитель полей (по умолчанию: табуляция)\n")
	fmt.Fprintf(os.Stderr, "  -s, --separated              Выводить только строки с разделителем\n")
	fmt.Fprintf(os.Stderr, "  -h, --help                   Показать эту справку\n")
	fmt.Fprintf(os.Stderr, "\nПримеры:\n")
	fmt.Fprintf(os.Stderr, "  %s -f 1,3-5 -d ',' file.csv\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  cat file.txt | %s -f 2\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -f 1-3 -s data.txt\n", os.Args[0])
}
