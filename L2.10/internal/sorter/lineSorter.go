package sorter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	config "github.com/GkadyrG/L2/L2.10/pkg/configs"
)

var monthMapping map[string]int

func init() {
	monthMapping = map[string]int{
		"jan": 1, "january": 1, "feb": 2, "february": 2,
		"mar": 3, "march": 3, "apr": 4, "april": 4,
		"may": 5, "jun": 6, "june": 6, "jul": 7, "july": 7,
		"aug": 8, "august": 8, "sep": 9, "september": 9,
		"oct": 10, "october": 10, "nov": 11, "november": 11,
		"dec": 12, "december": 12,
	}
}

type LineSorter struct {
	textLines []string
	config    config.SortConfig
}

func CreateLineSorter(lines []string, cfg config.SortConfig) *LineSorter {
	return &LineSorter{lines, cfg}
}

func (ls *LineSorter) GetSortedLines() []string {
	return ls.textLines
}

func (ls *LineSorter) PerformSort() {
	sort.SliceStable(ls.textLines, func(i, j int) bool {
		key1 := generateSortKey(ls.config, ls.textLines[i])
		key2 := generateSortKey(ls.config, ls.textLines[j])

		if key1 == "0" {
			return true
		}
		if key2 == "0" {
			return false
		}

		if *ls.config.ReverseOrder {
			return key1 >= key2
		}
		return key1 < key2
	})
}

func generateSortKey(cfg config.SortConfig, inputLine string) string {
	var targetPart string
	lineParts := strings.Fields(inputLine)

	if *cfg.ColumnIndex < 1 || *cfg.ColumnIndex > len(lineParts) {
		return "0"
	} else {
		targetPart = lineParts[*cfg.ColumnIndex-1]
	}

	if *cfg.IgnoreSpaces {
		targetPart = strings.TrimSpace(targetPart)
	}

	if (*cfg.HumanReadable || *cfg.NumericSort) && *cfg.MonthSort {
		panic("конфликт флагов: числовая сортировка и сортировка по месяцам несовместимы")
	}

	switch {
	case *cfg.HumanReadable:
		return parseHumanReadableSize(targetPart)
	case *cfg.NumericSort:
		return parseNumericValue(targetPart)
	case *cfg.MonthSort:
		return parseMonthValue(targetPart)
	default:
		return targetPart
	}
}

func parseNumericValue(input string) string {
	if value, err := strconv.ParseFloat(input, 64); err != nil {
		return "0"
	} else {
		return fmt.Sprintf("%020.0f", value)
	}
}

func parseHumanReadableSize(input string) string {
	if len(input) == 0 {
		return "0"
	}

	numericPart := input
	factor := 1.0
	suffix := strings.ToUpper(input[len(input)-1:])

	if strings.ContainsAny(suffix, "KMGT") {
		numericPart = strings.TrimRight(input, "KMGT")
		switch suffix {
		case "K":
			factor = 1e3
		case "M":
			factor = 1e6
		case "G":
			factor = 1e9
		case "T":
			factor = 1e12
		}
	}

	value, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return "0"
	}
	return fmt.Sprintf("%020.0f", value*factor)
}

func parseMonthValue(input string) string {
	if monthNum, exists := monthMapping[strings.ToLower(input)]; exists {
		return fmt.Sprintf("%02d", monthNum)
	} else {
		return "0"
	}
}

func isFileSorted(filePath string, cfg config.SortConfig) bool {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Ошибка открытия файла: %v", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return true // Пустой файл считается отсортированным
	}

	previousLine := scanner.Text()
	previousKey := generateSortKey(cfg, previousLine)

	for scanner.Scan() {
		currentLine := scanner.Text()
		currentKey := generateSortKey(cfg, currentLine)

		if *cfg.ReverseOrder {
			if currentKey > previousKey {
				return false
			}
		} else {
			if currentKey < previousKey {
				return false
			}
		}

		previousKey = currentKey
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Ошибка чтения файла: %v", err)
		return false
	}

	return true
}
