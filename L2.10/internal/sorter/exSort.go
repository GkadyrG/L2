package sorter

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strconv"

	config "github.com/GkadyrG/L2/L2.10/pkg/configs"
)

const DefaultChunkLimit = 8000 // Размер блока данных

type ExternalSorter struct {
	config         config.SortConfig
	temporaryFiles []string
	sourceFile     string
	targetFile     string
	blockSize      int
}

func NewExternalSorter(cfg config.SortConfig, source, target string, blockSize int) *ExternalSorter {
	return &ExternalSorter{cfg, make([]string, 0), source, target, blockSize}
}

func ExecuteExternalSort(inputPath, outputPath string, cfg config.SortConfig) error {
	sorter := NewExternalSorter(cfg, inputPath, outputPath, DefaultChunkLimit)

	if *sorter.config.CheckSorted {
		if isFileSorted(sorter.sourceFile, cfg) {
			fmt.Println("Файл уже отсортирован")
			return nil
		} else {
			fmt.Println("Файл не отсортирован")
			return nil
		}
	}

	if err := sorter.divideAndSortBlocks(); err != nil {
		return err
	}
	return sorter.combineBlocks()
}

func (es *ExternalSorter) divideAndSortBlocks() error {
	file, err := os.Open(es.sourceFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dataBuffer := make([]string, 0, es.blockSize)
	currentSize := 0
	blockIndex := 0

	for scanner.Scan() {
		dataBuffer = append(dataBuffer, scanner.Text())
		currentSize++

		if currentSize >= es.blockSize {
			es.processSingleBlock(blockIndex, dataBuffer)
			blockIndex++
			dataBuffer = dataBuffer[:0]
			currentSize = 0
		}
	}

	if len(dataBuffer) > 0 {
		es.processSingleBlock(blockIndex, dataBuffer)
	}

	return nil
}

func (es *ExternalSorter) processSingleBlock(blockNum int, lines []string) {
	lineSorter := CreateLineSorter(lines, es.config)
	lineSorter.PerformSort()

	blockFileName := "temp_block_" + strconv.Itoa(blockNum) + ".tmp"
	blockFile, _ := os.Create(blockFileName)
	defer blockFile.Close()

	for _, line := range lineSorter.GetSortedLines() {
		if _, err := blockFile.WriteString(line + "\n"); err != nil {
			fmt.Printf("Ошибка записи в блок: %v", err)
		}
	}

	es.temporaryFiles = append(es.temporaryFiles, blockFileName)
}

func (es *ExternalSorter) combineBlocks() error {
	outputFile, err := os.Create(es.targetFile)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(outputFile)
	defer func() {
		writer.Flush()
		outputFile.Close()
	}()

	blockFiles := make([]*os.File, len(es.temporaryFiles))
	blockScanners := make([]*bufio.Scanner, len(es.temporaryFiles))

	for i, tempFile := range es.temporaryFiles {
		f, err := os.Open(tempFile)
		if err != nil {
			return err
		}
		blockFiles[i] = f
		blockScanners[i] = bufio.NewScanner(f)
	}

	priorityQueue := &MinimalHeap{}
	heap.Init(priorityQueue)

	for i, scanner := range blockScanners {
		if scanner.Scan() {
			heap.Push(priorityQueue, QueueElement{content: scanner.Text(), sourceIndex: i})
		}
	}

	var previousLine string
	isFirstLine := true

	for priorityQueue.Len() > 0 {
		currentElement := heap.Pop(priorityQueue).(QueueElement)

		if *es.config.UniqueOnly {
			if isFirstLine {
				previousLine = currentElement.content
				if _, err := writer.WriteString(currentElement.content + "\n"); err != nil {
					fmt.Printf("Ошибка записи: %v", err)
				}
				isFirstLine = false
			} else if generateSortKey(es.config, currentElement.content) != generateSortKey(es.config, previousLine) {
				previousLine = currentElement.content
				if _, err := writer.WriteString(currentElement.content + "\n"); err != nil {
					fmt.Printf("Ошибка записи: %v", err)
				}
			}
		} else {
			if _, err := writer.WriteString(currentElement.content + "\n"); err != nil {
				fmt.Printf("Ошибка записи: %v", err)
			}
		}

		if blockScanners[currentElement.sourceIndex].Scan() {
			nextContent := blockScanners[currentElement.sourceIndex].Text()
			if !(*es.config.UniqueOnly && nextContent == previousLine) {
				heap.Push(priorityQueue, QueueElement{
					content:     nextContent,
					sourceIndex: currentElement.sourceIndex,
				})
			}
		}
	}

	for _, f := range blockFiles {
		f.Close()
		os.Remove(f.Name())
	}

	return nil
}
