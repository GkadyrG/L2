package processor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	config "github.com/GkadyrG/L2/L2.10/pkg/configs"
	"github.com/GkadyrG/L2/L2.10/internal/sorter"
)

func ProcessPipeInput(cfg config.SortConfig) error {
	tempInput, err := os.CreateTemp("", "input_data_*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tempInput.Name())
	defer tempInput.Close()

	if _, err = io.Copy(tempInput, os.Stdin); err != nil {
		return err
	}
	tempInput.Close()

	tempOutput, err := os.CreateTemp("", "output_data_*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tempOutput.Name())
	defer tempOutput.Close()

	if err = sorter.ExecuteExternalSort(tempInput.Name(), tempOutput.Name(), cfg); err != nil {
		return err
	}

	if _, err = tempOutput.Seek(0, 0); err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, tempOutput)
	return err
}

func ProcessFileToConsole(inputPath string, cfg config.SortConfig) error {
	tempOutput, err := os.CreateTemp("", "file_output_*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tempOutput.Name())
	defer tempOutput.Close()

	if err = sorter.ExecuteExternalSort(inputPath, tempOutput.Name(), cfg); err != nil {
		return err
	}

	if _, err = tempOutput.Seek(0, 0); err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, tempOutput)
	return err
}

func ProcessInteractiveMode(cfg config.SortConfig) {
	reader := bufio.NewReader(os.Stdin)
	var textLines []string

	fmt.Fprintln(os.Stderr, "Введите текст для сортировки (Ctrl+D для завершения):")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		textLines = append(textLines, strings.TrimSuffix(line, "\n"))
	}

	lineProcessor := sorter.CreateLineSorter(textLines, cfg)
	lineProcessor.PerformSort()

	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()

	for _, line := range lineProcessor.GetSortedLines() {
		if _, err := output.WriteString(line + "\n"); err != nil {
			fmt.Printf("Ошибка записи: %v", err)
		}
	}
}
