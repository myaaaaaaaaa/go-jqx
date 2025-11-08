package jqx

import (
	"errors"
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func htmlQuerySelector(htmlString, cssSelector string) ([]string, error) {
	sel, err := cascadia.Parse(cssSelector)
	if err != nil {
		return nil, err
	}

	// Lenient parser, only relays errors from io.Reader
	doc := must(html.Parse(strings.NewReader(htmlString)))

	var rt []string
	for _, node := range cascadia.QueryAll(doc, sel) {
		var sb strings.Builder
		must(0, html.Render(&sb, node))
		rt = append(rt, sb.String())
	}
	return rt, nil
}
func htmlReplaceSelector(htmlString, cssSelector, replacement string) (string, error) {
	sel, err := cascadia.Parse(cssSelector)
	if err != nil {
		return "", err
	}

	doc := must(html.Parse(strings.NewReader(htmlString)))
	replaceSplit := strings.Split(replacement, "<>")

	for _, node := range cascadia.QueryAll(doc, sel) {
		replacement := replacement

		if len(replaceSplit) > 1 {
			var orig strings.Builder
			must(0, html.Render(&orig, node))
			replacement = strings.Join(replaceSplit, orig.String())
		}

		node.Type = html.RawNode
		node.Data = replacement
	}

	var sb strings.Builder
	must(0, html.Render(&sb, doc))
	return sb.String(), nil
}

func htmlTokenize(htmlString, tokenFilter string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlString))
	var sb strings.Builder

	tokenFilterMap := map[string]bool{}
	for k := range strings.FieldsSeq(tokenFilter) {
		tokenFilterMap[k] = true
	}

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}

		keepToken := false

		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			name, _ := tokenizer.TagName()
			// https://go.dev/wiki/CompilerOptimizations#map-lookup-by-byte
			keepToken = tokenFilterMap[string(name)]
		default:
			keepToken = tokenFilterMap[strings.ToUpper(tokenType.String())]
		}

		if keepToken {
			sb.Write(tokenizer.Raw())
		}
	}

	err := tokenizer.Err()
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		panic(err)
	}

	return sb.String()
}

func htmlTokenizeToMaps(htmlString string) []map[string]any {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlString))
	var maps []map[string]any

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		raw := string(tokenizer.Raw())
		token := tokenizer.Token()
		m := map[string]any{
			"VALUE": token.Data,
			"TYPE":  token.Type.String(),
			"RAW":   raw,
		}
		for _, a := range token.Attr {
			m[a.Key] = a.Val
		}
		maps = append(maps, m)
	}

	return maps
}
