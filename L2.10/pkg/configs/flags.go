package config

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type SortConfig struct {
	ColumnIndex   *int
	NumericSort   *bool
	ReverseOrder  *bool
	UniqueOnly    *bool
	MonthSort     *bool
	IgnoreSpaces  *bool
	CheckSorted   *bool
	HumanReadable *bool
}

func ParseCommandLine() (*SortConfig, []string) {
	cfg := SortConfig{}

	cfg.ColumnIndex = flag.IntP("column", "k", 1, "сортировка по указанному столбцу")
	cfg.NumericSort = flag.BoolP("numeric", "n", false, "числовая сортировка")
	cfg.ReverseOrder = flag.BoolP("reverse", "r", false, "обратный порядок")
	cfg.UniqueOnly = flag.BoolP("unique", "u", false, "только уникальные строки")
	cfg.MonthSort = flag.BoolP("month", "M", false, "сортировка по месяцам")
	cfg.IgnoreSpaces = flag.BoolP("blanks", "b", false, "игнорировать пробелы")
	cfg.CheckSorted = flag.BoolP("check", "c", false, "проверить сортировку")
	cfg.HumanReadable = flag.BoolP("human", "h", false, "человекочитаемые размеры")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: %s [опции] [файл]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Опции:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	return &cfg, flag.Args()
}
