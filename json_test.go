package jqx

import (
	"strings"
	"testing"
)

func TestPrettyPrintJSONLines(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30, "isStudent": false, "courses": ["Math", "Science"]}`
	got := must(prettyPrintJSONLines(jsonStr, "."))
	want := `{
."name": "John",
."age": 30,
."isStudent": false,
."courses": [
.."Math",
.."Science"
.]
}`
	assertEqual(t, strings.Join(got, "\n"), want)
}

func TestPrettyPrintJSONLines_MalformedJSON(t *testing.T) {
	jsonStr := `{"name":,}`
	_, err := prettyPrintJSONLines(jsonStr, ".")
	assertEqual(t, err != nil, true)
}
