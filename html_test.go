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

// HTMLString is a string that is a valid HTML snippet.
type HTMLString string

// Generate generates a random HTML snippet.
func (HTMLString) Generate(r *rand.Rand, size int) reflect.Value {
	var sb strings.Builder
	tags := []string{"p", "div", "span", "a", "img", "h1", "h2", "h3"}
	for i := 0; i < size; i++ {
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

// ArgsString is a string that is a valid set of arguments for htmlExtract.
type ArgsString string

// Generate generates a random set of arguments for htmlExtract.
func (ArgsString) Generate(r *rand.Rand, size int) reflect.Value {
	args := []string{"p", "div", "span", "a", "img", "h1", "h2", "h3", "TEXT", "COMMENT"}
	rand.Shuffle(len(args), func(i, j int) { args[i], args[j] = args[j], args[i] })
	return reflect.ValueOf(ArgsString(strings.Join(args[:r.Intn(len(args)+1)], " ")))
}

func TestHtmlExtractProperties(t *testing.T) {
	t.Run("concatenation", func(t *testing.T) {
		f := func(html1, html2 HTMLString, args ArgsString) bool {
			want := htmlExtract(string(html1), string(args)) + htmlExtract(string(html2), string(args))
			got := htmlExtract(string(html1)+string(html2), string(args))
			return want == got
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("idempotency", func(t *testing.T) {
		f := func(html HTMLString, args ArgsString) bool {
			want := htmlExtract(string(html), string(args))
			got := htmlExtract(want, string(args))
			return want == got
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("empty_arguments", func(t *testing.T) {
		f := func(html HTMLString) bool {
			return htmlExtract(string(html), " ") == ""
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

}
