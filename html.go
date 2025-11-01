package jqx

import (
	"errors"
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func htmlQuerySelector(htmlString, cssSelector string) ([]string, error) {
	sel, err := cascadia.Parse(cssSelector)
	if err != nil {
		return nil, err
	}

	// Lenient parser, only relays errors from io.Reader
	doc := must(html.Parse(strings.NewReader(htmlString)))

	nodes := cascadia.QueryAll(doc, sel)

	var rt []string
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

func htmlReplaceSelector(htmlString, cssSelector, replacement string) (string, error) {
	sel, err := cascadia.Parse(cssSelector)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	nodes := cascadia.QueryAll(doc, sel)
	if len(nodes) == 0 {
		return htmlString, nil
	}

	replacementNodes, err := html.ParseFragment(strings.NewReader(replacement), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return "", err
	}

	for _, node := range nodes {
		for _, replacementNode := range replacementNodes {
			// Clone the node to avoid issues with multiple replacements
			clone := cloneNode(replacementNode)
			node.Parent.InsertBefore(clone, node)
		}
		node.Parent.RemoveChild(node)
	}

	var sb strings.Builder
	if err := html.Render(&sb, doc); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// cloneNode creates a deep copy of an html.Node
func cloneNode(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	clone := &html.Node{
		Type:      n.Type,
		DataAtom:  n.DataAtom,
		Data:      n.Data,
		Namespace: n.Namespace,
		Attr:      make([]html.Attribute, len(n.Attr)),
	}
	copy(clone.Attr, n.Attr)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		clone.AppendChild(cloneNode(c))
	}
	return clone
}
