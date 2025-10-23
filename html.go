package jqx

import (
	"errors"
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func htmlQuerySelector(htmlString, cssSelector string) ([]any, error) {
	sel, err := cascadia.Parse(cssSelector)
	if err != nil {
		return nil, err
	}

	// Lenient parser, only relays errors from io.Reader
	doc := must(html.Parse(strings.NewReader(htmlString)))

	nodes := cascadia.QueryAll(doc, sel)

	var rt []any
	for _, node := range nodes {
		var sb strings.Builder
		must(0, html.Render(&sb, node))
		rt = append(rt, sb.String())
	}
	return rt, nil
}

func htmlExtract(htmlString, tokenFilter string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlString))
	var sb strings.Builder

	callbacks := [16]func(){}
	tagMatchers := map[string]bool{}
	for k := range strings.FieldsSeq(tokenFilter) {
		switch k {
		case "COMMENT":
			callbacks[html.CommentToken] = func() {
				sb.WriteString(tokenizer.Token().String())
			}
		case "TEXT":
			callbacks[html.TextToken] = func() {
				sb.Write(tokenizer.Text())
			}

		default:
			tagMatchers[k] = true

			callbacks[html.StartTagToken] = func() {
				t := tokenizer.Token()
				if !tagMatchers[t.Data] {
					return
				}
				sb.WriteString(t.String())
			}
		}
	}

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		switch tokenType {
		case html.SelfClosingTagToken, html.EndTagToken:
			tokenType = html.StartTagToken
		}

		f := callbacks[tokenType]
		if f != nil {
			f()
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
