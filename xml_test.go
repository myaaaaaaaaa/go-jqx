package jqx

import (
	"testing"
)

func TestXmlQueryPath(t *testing.T) {
	xml := `<root><book><title>The Go Programming Language</title></book></root>`
	xpath := "/root/book/title"

	result, err := xmlQueryPath(xml, xpath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `<title>The Go Programming Language</title>`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestXmlQueryPath_NoResult(t *testing.T) {
	xml := `<root><book><title>The Go Programming Language</title></book></root>`
	xpath := "/root/book/author"

	result, err := xmlQueryPath(xml, xpath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestXmlQueryPath_InvalidXPath(t *testing.T) {
	xml := `<root><book><title>The Go Programming Language</title></book></root>`
	xpath := "/root/book/title["

	_, err := xmlQueryPath(xml, xpath)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestXmlQueryPath_MultipleMatches(t *testing.T) {
	xml := `<root><book><title>Title 1</title><title>Title 2</title></book></root>`
	xpath := "/root/book/title"

	result, err := xmlQueryPath(xml, xpath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `<title>Title 1</title><title>Title 2</title>`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}