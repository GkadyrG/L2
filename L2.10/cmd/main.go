package main

import (
	"log"
	"os"

	"github.com/GkadyrG/L2/L2.10/internal/processor"
	config "github.com/GkadyrG/L2/L2.10/pkg/configs"
)

func main() {
	settings, inputFiles := config.ParseCommandLine()

	switch len(inputFiles) {
	case 1:
		// Обработка одного файла
		if err := processor.ProcessFileToConsole(inputFiles[0], *settings); err != nil {
			log.Fatal(err)
		}
	case 0:
		// Определяем источник данных
		fileInfo, _ := os.Stdin.Stat()
		if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
			// Данные из pipe
			if err := processor.ProcessPipeInput(*settings); err != nil {
				log.Fatal(err)
			}
		} else {
			// Интерактивный ввод
			processor.ProcessInteractiveMode(*settings)
		}
	default:
		log.Fatal("Слишком много входных файлов")
	}
}
