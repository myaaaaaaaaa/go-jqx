package jqx

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/quick"
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

	for i := range 5 {
		fullText := strings.Repeat("<!-- comment -->text", i)
		commentText := strings.Repeat("<!-- comment -->", i)
		plainText := strings.Repeat("text", i)

		assertEqual(t, htmlExtract(fullText, "  TEXT  "), plainText)
		assertEqual(t, htmlExtract(fullText, "  TEXT  COMMENT  "), fullText)
		assertEqual(t, htmlExtract(fullText, "  COMMENT  "), commentText)
		assertEqual(t, htmlExtract(fullText, "    "), "")

		assertEqual(t, htmlExtract(commentText, "  TEXT  COMMENT  "), commentText)
		assertEqual(t, htmlExtract(commentText, "  TEXT  "), "")
		assertEqual(t, htmlExtract(commentText, "  COMMENT  "), commentText)
		assertEqual(t, htmlExtract(commentText, "    "), "")

		assertEqual(t, htmlExtract(plainText, "  TEXT  COMMENT  "), plainText)
		assertEqual(t, htmlExtract(plainText, "  TEXT  "), plainText)
		assertEqual(t, htmlExtract(plainText, "  COMMENT  "), "")
		assertEqual(t, htmlExtract(plainText, "    "), "")
	}

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

// A string that is a valid HTML snippet.
type HTMLString []byte

// Generate generates a random HTML snippet.
func (HTMLString) Generate(r *rand.Rand, size int) reflect.Value {
	var sb strings.Builder
	tags := []string{"p", "div", "span", "a", "img", "h1", "h2", "h3"}
	for range size {
		switch r.Intn(3) {
		case 0: // Text
			sb.WriteString("some text ")
		case 1: // Comment
			sb.WriteString("<!-- some comment -->")
		case 2: // Tag
			tag := tags[r.Intn(len(tags))]
			sb.WriteString("<" + tag + ">")
			if r.Intn(2) == 0 {
				sb.WriteString("some content")
			}
			sb.WriteString("</" + tag + ">")
		}
	}
	return reflect.ValueOf(HTMLString(sb.String()))
}

// A string that is a valid set of arguments for htmlExtract.
type TokenFilterString []byte

// Generate generates a random set of arguments for htmlExtract.
func (TokenFilterString) Generate(r *rand.Rand, size int) reflect.Value {
	tokenFilter := []string{"p", "div", "span", "a", "img", "h1", "h2", "h3", "TEXT", "COMMENT"}
	rand.Shuffle(len(tokenFilter), func(i, j int) { tokenFilter[i], tokenFilter[j] = tokenFilter[j], tokenFilter[i] })
	return reflect.ValueOf(TokenFilterString(strings.Join(tokenFilter[:r.Intn(len(tokenFilter)+1)], " ")))
}

func (b HTMLString) s() string        { return string(b) }
func (b TokenFilterString) s() string { return string(b) }

func TestHtmlExtractProperties(t *testing.T) {
	fuzz := func(name string, f any) {
		t.Run(name, func(t *testing.T) {
			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})
	}

	fuzz("concatenation", func(html1, html2 HTMLString, tokenFilter TokenFilterString) bool {
		want := htmlExtract(html1.s(), tokenFilter.s()) + htmlExtract(html2.s(), tokenFilter.s())
		got := htmlExtract(html1.s()+html2.s(), tokenFilter.s())
		return want == got
	})

	fuzz("idempotency", func(html HTMLString, tokenFilter TokenFilterString) bool {
		want := htmlExtract(html.s(), tokenFilter.s())
		got := htmlExtract(want, tokenFilter.s())
		return want == got
	})

	fuzz("filtering", func(html HTMLString, tokenFilter1, tokenFilter2, tokenFilter3 TokenFilterString) bool {
		outputA := htmlExtract(html.s(), tokenFilter2.s())
		outputB := htmlExtract(html.s(), tokenFilter1.s()+" "+tokenFilter2.s()+" "+tokenFilter3.s())
		return isSubsequence(outputA, outputB) &&
			isSubsequence(outputA, html.s()) &&
			isSubsequence(outputB, html.s())
	})

	fuzz("empty_arguments", func(html HTMLString) bool {
		return htmlExtract(html.s(), " ") == ""
	})
}

func isSubsequence(sub, super string) bool {
	i, j := 0, 0
	for i < len(sub) && j < len(super) {
		if sub[i] == super[j] {
			i++
		}
		j++
	}
	return i == len(sub)
}
