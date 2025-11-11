package jqx

import (
	"bytes"
	"encoding/json"
	"strings"
)

func prettyPrintJSONLines(s string, indent string) ([]string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(s), "", indent)
	if err != nil {
		return nil, err
	}
	return strings.Split(out.String(), "\n"), nil
}
