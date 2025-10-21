package jqx

import (
	"bytes"
	"errors"
	"fmt"
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
			sb.Write(tokenizer.Text())
		case html.StartTagToken, html.SelfClosingTagToken:
			t := tokenizer.Token()
			for _, attr := range t.Attr {
				if attr.Key == "alt" {
					sb.WriteString(fmt.Sprintf("\n(%s: %s)\n", t.Data, attr.Val))
				}
			}
		}
	}

end:
	err := tokenizer.Err()
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return sb.String(), err
}
