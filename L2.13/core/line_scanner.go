package core

import (
	"bufio"
	"io"
)

// LineScanner предоставляет интерфейс для сканирования строк
type LineScanner struct {
	scanner *bufio.Scanner
	err     error
	hasNext bool
	next    string
}

// NewLineScanner создает новый сканер строк
func NewLineScanner(reader io.Reader) *LineScanner {
	scanner := bufio.NewScanner(reader)
	ls := &LineScanner{
		scanner: scanner,
		err:     nil,
		hasNext: false,
		next:    "",
	}
	ls.advance()
	return ls
}

// HasNext проверяет, есть ли следующая строка
func (ls *LineScanner) HasNext() bool {
	return ls.hasNext
}

// Next возвращает следующую строку
func (ls *LineScanner) Next() string {
	if !ls.hasNext {
		return ""
	}
	result := ls.next
	ls.advance()
	return result
}

// Err возвращает ошибку сканирования
func (ls *LineScanner) Err() error {
	return ls.err
}

// advance переходит к следующей строке
func (ls *LineScanner) advance() {
	if ls.scanner.Scan() {
		ls.hasNext = true
		ls.next = ls.scanner.Text()
	} else {
		ls.hasNext = false
		ls.next = ""
		ls.err = ls.scanner.Err()
	}
}
