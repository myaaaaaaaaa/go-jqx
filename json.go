package jqx

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func jsonTokenize(s string) ([]string, error) {
	decoder := json.NewDecoder(strings.NewReader(s))
	var tokens []string
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
			continue
		case string:
			tokenStr = token
		default:
			tokenStr = fmt.Sprint(token)
		}
		tokens = append(tokens, tokenStr)
	}
}
