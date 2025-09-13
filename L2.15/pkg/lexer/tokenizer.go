package lexer

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

func ParseInput(input string) ([]string, error) {
	var tokens []string
	var current strings.Builder

	inQuotes := false
	quoteChar := rune(0)
	escape := false
	singleQuoted := false // <--- новый флаг, запоминает, что токен из одинарных кавычек

	runes := []rune(strings.TrimSpace(input))

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case escape:
			current.WriteRune(r)
			escape = false

		case r == '\\' && (!inQuotes || quoteChar == '"'):
			escape = true

		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
				if r == '\'' {
					singleQuoted = true
				}
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}

		case isOperator(runes, i) && !inQuotes:
			if current.Len() > 0 {
				tok := current.String()
				if !singleQuoted {
					tok = expandVars(tok)
				}
				tokens = append(tokens, tok)
				current.Reset()
				singleQuoted = false
			}

			if op, size := matchOperator(runes, i); size > 0 {
				tokens = append(tokens, op)
				i += size - 1
			}

		case unicode.IsSpace(r) && !inQuotes:
			if current.Len() > 0 {
				tok := current.String()
				if !singleQuoted {
					tok = expandVars(tok)
				}
				tokens = append(tokens, tok)
				current.Reset()
				singleQuoted = false
			}

		default:
			current.WriteRune(r)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("незакрытые кавычки")
	}

	if current.Len() > 0 {
		tok := current.String()
		if !singleQuoted {
			tok = expandVars(tok)
		}
		tokens = append(tokens, tok)
	}

	if tokens == nil {
		tokens = []string{}
	}

	return tokens, nil
}

func isOperator(runes []rune, pos int) bool {
	switch runes[pos] {
	case '|', '>', '<', '&':
		return true
	}
	return false
}

func matchOperator(runes []rune, pos int) (string, int) {
	if pos+1 < len(runes) {
		curr, next := runes[pos], runes[pos+1]
		if curr == '&' && next == '&' {
			return "&&", 2
		}
		if curr == '|' && next == '|' {
			return "||", 2
		}
	}
	return string(runes[pos]), 1
}

func expandVars(s string) string {
	if !strings.Contains(s, "$") {
		return s
	}
	return os.ExpandEnv(s)
}
