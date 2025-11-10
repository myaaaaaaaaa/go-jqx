package jqx

import (
	"reflect"
	"testing"
)

func TestJsonTokenize(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30, "isStudent": false, "courses": ["Math", "Science"]}`
	expected := []string{"name", "John", "age", "30", "isStudent", "false", "courses", "Math", "Science"}
	actual, err := jsonTokenize(jsonStr)
	if err != nil {
		t.Errorf("jsonTokenize() error = %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("jsonTokenize() = %v, want %v", actual, expected)
	}
}

func TestJsonTokenize_MalformedJSON(t *testing.T) {
	jsonStr := `{"name":,}`
	_, err := jsonTokenize(jsonStr)
	if err == nil {
		t.Errorf("jsonTokenize() error = %v, wantErr %v", err, true)
	}
}
