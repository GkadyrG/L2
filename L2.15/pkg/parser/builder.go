package parser

import (
	"fmt"

	"github.com/GkadyrG/L2/L2.15/pkg/models"
)

func BuildCommand(tokens []string) (*models.Cmd, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("пустая команда")
	}

	return parseTokenList(tokens, 0)
}

func parseTokenList(tokens []string, start int) (*models.Cmd, error) {
	if start >= len(tokens) {
		return nil, fmt.Errorf("неожиданный конец")
	}

	cmd := &models.Cmd{}
	i := start

	// Читаем команду и аргументы до оператора
	for i < len(tokens) {
		token := tokens[i]

		switch token {
		case "|":
			if cmd.Binary == "" {
				return nil, fmt.Errorf("пайп без команды")
			}
			nextCmd, err := parseTokenList(tokens, i+1)
			if err != nil {
				return nil, err
			}
			cmd.NextPipe = nextCmd
			return cmd, nil

		case "&&":
			if cmd.Binary == "" {
				return nil, fmt.Errorf("&& без команды")
			}
			nextCmd, err := parseTokenList(tokens, i+1)
			if err != nil {
				return nil, err
			}
			cmd.NextAnd = nextCmd
			return cmd, nil

		case "||":
			if cmd.Binary == "" {
				return nil, fmt.Errorf("|| без команды")
			}
			nextCmd, err := parseTokenList(tokens, i+1)
			if err != nil {
				return nil, err
			}
			cmd.NextOr = nextCmd
			return cmd, nil

		case ">", "<":
			if i+1 >= len(tokens) {
				return nil, fmt.Errorf("отсутствует файл для %s", token)
			}
			fileOp := models.FileOperation{
				Operation: token,
				Filename:  tokens[i+1],
			}
			cmd.FileOps = append(cmd.FileOps, fileOp)
			i += 2
			continue

		default:
			if cmd.Binary == "" {
				cmd.Binary = token
			} else {
				cmd.Arguments = append(cmd.Arguments, token)
			}
		}
		i++
	}

	if cmd.Binary == "" {
		return nil, fmt.Errorf("команда не указана")
	}

	return cmd, nil
}
