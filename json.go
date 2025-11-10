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
		if err == io.EOF {
			break
		}
		if err != nil {
			return tokens, err
		}
		switch token.(type) {
		case json.Delim:
			continue
		default:
			tokens = append(tokens, fmt.Sprint(token))
		}
	}
	return tokens, nil
}
