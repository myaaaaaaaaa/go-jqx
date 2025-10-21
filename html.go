package jqx

import (
	"bytes"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func selectHTML(htmlText, cssSelector string) (string, error) {
	// Parse the CSS selector
	sel, err := cascadia.Parse(cssSelector)
	if err != nil {
		return "", err
	}

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return "", err
	}

	// Find matching nodes
	nodes := cascadia.QueryAll(doc, sel)

	// Render the matched nodes to a string
	var buffer bytes.Buffer
	for _, node := range nodes {
		err := html.Render(&buffer, node)
		if err != nil {
			return "", err
		}
	}

	return buffer.String(), nil
}
