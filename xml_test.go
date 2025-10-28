package jqx

import (
	"strings"
	"testing"
)

func checkXmlQueryPath(xmlString, xpath string) string {
	rt := strings.Join(must(xmlQueryPath(xmlString, xpath)), "")
	r2 := strings.Join(must(xmlQueryPath(xmlString+xmlString, xpath)), "")

	if rt+rt != r2 {
		panic(strings.Join([]string{
			xmlString,
			rt,
			r2,
		}, "\n"))
	}

	return rt
}

func TestXmlQueryPath(t *testing.T) {
	xml := `<root><book><title>The Go Programming Language</title></book></root>`
	xpath := "/root/book/title"

	got := checkXmlQueryPath(xml, xpath)
	want := `<title>The Go Programming Language</title>`
	assertEqual(t, got, want)
}

func TestXmlQueryPath_NoResult(t *testing.T) {
	xml := `<root><book><title>The Go Programming Language</title></book></root>`
	xpath := "/root/book/author"

	got := checkXmlQueryPath(xml, xpath)
	assertEqual(t, got, "")
}

func TestXmlQueryPath_InvalidXPath(t *testing.T) {
	xml := `<root><book><title>The Go Programming Language</title></book></root>`
	xpath := "/root/book/title["

	_, err := xmlQueryPath(xml, xpath)
	assertEqual(t, err != nil, true)
}

func TestXmlQueryPath_MultipleMatches(t *testing.T) {
	xml := `<root><book><title>Title 1</title><title>Title 2</title></book></root>`
	xpath := "/root/book/title"

	got := checkXmlQueryPath(xml, xpath)
	want := `<title>Title 1</title><title>Title 2</title>`
	assertEqual(t, got, want)
}
