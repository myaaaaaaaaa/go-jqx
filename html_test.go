package jqx

import (
	"slices"
	"strings"
	"testing"
)

func TestHTMLQuerySelector(t *testing.T) {
	html := `
		<html>
			<body>
				<h1 class="header">Title</h1>
				<p>First paragraph.</p>
				<p class="content">Second paragraph.</p>
				<div class="footer">Footer</div>
		</html>
	`

	query := (&State{
		Globals: map[string]any{"html": html},
	}).Compile(`. as $sel | $html | [htmlq($sel)] | join("  ;  ")`)

	assert := func(selector, want string) {
		t.Helper()

		if want == "error" {
			_, err := htmlQuerySelector(html, selector)
			assertEqual(t, err != nil, true)
		} else {
			got := slices.Collect(query(selector))
			assertEqual(t, len(got), 1)
			assertString(t, got[0], want)
		}
	}

	assert(`.content`, `<p class="content">Second paragraph.</p>`)
	assert(`h1`, `<h1 class="header">Title</h1>`)
	assert(`p`, `<p>First paragraph.</p>  ;  <p class="content">Second paragraph.</p>`)

	assert(`invalid[`, "error")
}

func TestHtmlExtract(t *testing.T) {
	assert := func(html, want string) {
		t.Helper()
		assertEqual(t, htmlExtract(html, "  TEXT  a  img  "), want)
	}

	assert(
		`<!-- This is a comment --><body><p>Some text</p></body>`,
		`Some text`,
	)
	assert(
		`<p>Hello</p><img src="image.jpg" alt="olleh"/>`,
		`Hello<img src="image.jpg" alt="olleh"/>`,
	)

	for i := range 8 {
		plainText := strings.Repeat("hello  world", i)
		htmlText := strings.Repeat("<p>hello  world</p>", i)
		pText := strings.Repeat("<p></p>", i)

		assertEqual(t, htmlExtract(plainText, "  TEXT  "), plainText)
		assertEqual(t, htmlExtract(plainText, "  TEXT  p  "), plainText)
		assertEqual(t, htmlExtract(plainText, "  p  "), "")
		assertEqual(t, htmlExtract(plainText, "   "), "")

		assertEqual(t, htmlExtract(htmlText, "  TEXT  "), plainText)
		assertEqual(t, htmlExtract(htmlText, "  TEXT  p  "), htmlText)
		assertEqual(t, htmlExtract(htmlText, "  p  "), pText)
		assertEqual(t, htmlExtract(htmlText, "   "), "")

		assertEqual(t, htmlExtract(pText, "  TEXT  "), "")
		assertEqual(t, htmlExtract(pText, "  TEXT  p  "), pText)
		assertEqual(t, htmlExtract(pText, "  p  "), pText)
		assertEqual(t, htmlExtract(pText, "   "), "")
	}

	abc := ""
	for i := range 13 {
		abc += string(rune(i + 'a'))
	}
	for i := range len(abc) {
		s := strings.Split(abc, "")
		s[i] = "<hr/>" + s[i] + "<hr/>"
		want := strings.Join(s, "")
		html := strings.Join(s, "<br/>")

		assertEqual(t, htmlExtract(html, "  TEXT  "), abc)
		assertEqual(t, htmlExtract(html, "  TEXT  hr  "), want)
		assertEqual(t, htmlExtract(html, "  TEXT  br  hr  "), html)
	}
}
