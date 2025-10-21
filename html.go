package jqx

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func htmlQuerySelector(htmlText, cssSelector string) (string, error) {
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

func htmlExtractText(htmlString string) (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlString))
	var sb strings.Builder

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			goto end
		case html.TextToken:
			sb.WriteString(tokenizer.Token().Data)
		}
	}

end:
	err := tokenizer.Err()
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return sb.String(), err
}
