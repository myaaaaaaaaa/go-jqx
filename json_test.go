package jqx

import (
	"testing"
)

func TestJsonTokenize(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30, "isStudent": false, "courses": ["Math", "Science"]}`
	got := must(jsonTokenize(jsonStr))
	assertString(t, got, `[name John age 30 isStudent false courses Math Science]`)
}

func TestJsonTokenize_MalformedJSON(t *testing.T) {
	jsonStr := `{"name":,}`
	_, err := jsonTokenize(jsonStr)
	assertEqual(t, err != nil, true)
}
