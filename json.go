package jqx

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func jsonTokenize(s, indent string) ([]string, error) {
	decoder := json.NewDecoder(strings.NewReader(s))
	var tokens []string
	indentCount := 0
	for {
		token, err := decoder.Token()

		switch err {
		case nil:
		case io.EOF:
			return tokens, nil
		default:
			return nil, err
		}

		var tokenStr string
		switch token := token.(type) {
		case json.Delim:
			switch token {
			case '{', '[':
				indentCount++
			case '}', ']':
				indentCount--
			}
			continue
		case string:
			tokenStr = token
		default:
			tokenStr = fmt.Sprint(token)
		}
		tokens = append(tokens, strings.Repeat(indent, indentCount)+tokenStr)
	}
}
