package jqx

import (
	"bytes"
	"slices"
	"strings"
	"testing"
	"testing/quick"

	"github.com/myaaaaaaaaa/go-jqx/proptest"
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

func TestHTMLExtract(t *testing.T) {
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

func stringGen(b []byte, tokens ...string) string {
	var sb strings.Builder

	for _, c := range b {
		token := tokens[int(c)%len(tokens)]
		sb.WriteString(token)
	}

	return sb.String()
}

// A string that is a valid HTML snippet.
type HTMLString []byte

func (b HTMLString) s() string {
	return stringGen(b,
		`hello   world`,
		`<!-- a  comment -->`,
		`<p></p>`,
		`</div>`,
		`<span>`,
		`<a href="m.htm">`,
		`<img src="i.jpg"/>`,
		`<br/>`,
		`<hr/>`,
	)
}

// A string that is a valid set of arguments for htmlExtract.
type TokenFilterString []byte

func (b *TokenFilterString) trim(n int) {
	if len(*b) > n {
		*b = (*b)[:int((*b)[0])%n+1]
	}
}
func (b TokenFilterString) s() string {
	return stringGen(b,
		`  TEXT  `,
		`  COMMENT  `,
		`  p  `,
		`  div  `,
		`  span  `,
		`  a  `,
		`  img  `,
		`  br  `,
		`  hr  `,
	)
}

func TestHTMLExtractProperties(t *testing.T) {
	fuzz := func(name string, f any) {
		t.Run(name, func(t *testing.T) {
			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})
	}

	fuzz("concatenation", func(html1, html2 HTMLString, tokenFilter TokenFilterString) bool {
		tokenFilter.trim(3)
		want := htmlExtract(html1.s(), tokenFilter.s()) + htmlExtract(html2.s(), tokenFilter.s())
		got := htmlExtract(html1.s()+html2.s(), tokenFilter.s())
		return want == got
	})
	fuzz("inequality", func(html HTMLString, tokenFilter TokenFilterString) bool {
		for i := range tokenFilter {
			tokenFilter := tokenFilter[i : i+1]
			a := htmlExtract(html.s(), tokenFilter.s())
			tokenFilter[0]++
			b := htmlExtract(html.s(), tokenFilter.s())

			if a != "" && a == b {
				return false
			}
			if a != "" && strings.Contains(b, a) {
				return false
			}
			if b != "" && strings.Contains(a, b) {
				return false
			}
		}
		return true
	})
	fuzz("sets", func(html HTMLString, tokenFilter1, tokenFilter2 TokenFilterString) bool {
		tokenFilter1.trim(5)
		tokenFilter2.trim(5)
		output1 := htmlExtract(html.s(), tokenFilter1.s())
		output2 := htmlExtract(html.s(), tokenFilter2.s())
		outputIntersect := htmlExtract(output1, tokenFilter2.s())
		outputUnion := htmlExtract(html.s(), tokenFilter1.s()+tokenFilter2.s())
		return htmlExtract(output2, tokenFilter1.s()) == outputIntersect &&
			htmlExtract(output1, tokenFilter2.s()) == outputIntersect &&
			htmlExtract(outputUnion, tokenFilter1.s()) == output1 &&
			htmlExtract(outputUnion, tokenFilter2.s()) == output2 &&
			htmlExtract(outputIntersect, tokenFilter1.s()) == outputIntersect &&
			htmlExtract(outputIntersect, tokenFilter2.s()) == outputIntersect &&
			htmlExtract(outputIntersect, tokenFilter1.s()+tokenFilter2.s()) == outputIntersect
	})

	fuzz("idempotency", func(html HTMLString, tokenFilter TokenFilterString) bool {
		tokenFilter.trim(3)
		want := htmlExtract(html.s(), tokenFilter.s())
		got := htmlExtract(want, tokenFilter.s())
		return want == got
	})

	fuzz("filtering", func(html HTMLString, tokenFilter1, tokenFilter2, tokenFilter3 TokenFilterString) bool {
		tokenFilter1.trim(2)
		tokenFilter2.trim(3)
		tokenFilter3.trim(2)
		outputA := htmlExtract(html.s(), tokenFilter2.s())
		outputB := htmlExtract(html.s(), tokenFilter1.s()+tokenFilter2.s()+tokenFilter3.s())
		return isSubsequence(outputA, outputB) &&
			isSubsequence(outputA, html.s()) &&
			isSubsequence(outputB, html.s())
	})

	fuzz("catchall", func(html HTMLString, tokenFilter TokenFilterString) bool {
		if len(tokenFilter) == 0 {
			return true
		}

		tokenFilter.trim(4)
		for i := range html {
			html[i] = tokenFilter[int(html[i])%len(tokenFilter)]
		}
		return htmlExtract(html.s(), tokenFilter.s()) == html.s()
	})

	fuzz("misc_basic", func(html HTMLString, tokenFilter TokenFilterString) bool {
		tokenFilter.trim(3)
		return len(htmlExtract(html.s(), tokenFilter.s())) <= len(html.s()) &&
			htmlExtract(html.s(), " ") == "" &&
			htmlExtract("", tokenFilter.s()) == ""
	})
}
func TestHTMLExtractSequence(t *testing.T) {
	assert := func(html, tokenFilter, want []byte) {
		t.Helper()
		assertEqual(t,
			htmlExtract(
				HTMLString(html).s(),
				TokenFilterString(tokenFilter).s(),
			),
			HTMLString(want).s(),
		)
	}
	byteSeq := []byte{0, 1, 2, 3, 4, 5, 6}
	for i := range byteSeq {
		for rep := range 4 {
			rep := rep + 1
			assert(bytes.Repeat(byteSeq, rep), byteSeq[:i], bytes.Repeat(byteSeq[:i], rep))
			assert(bytes.Repeat(byteSeq, rep), byteSeq[i:], bytes.Repeat(byteSeq[i:], rep))
			assert(bytes.Repeat(byteSeq, rep), byteSeq[i:i+1], bytes.Repeat(byteSeq[i:i+1], rep))
		}
	}
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

func TestHTMLExtractDifferential(t *testing.T) {
	charExtract := func(s, charFilter string) string {
		charSet := [128]bool{}
		for _, c := range charFilter {
			charSet[c] = true
		}
		return strings.Map(func(r rune) rune {
			if charSet[r] {
				return r
			} else {
				return -1
			}
		}, s)
	}

	htmlReplacer := strings.NewReplacer(
		"1", `hello   world`,
		"2", `<!-- a  comment -->`,
		"3", `<p></p>`,
		"4", `</div>`,
		"5", `<span>`,
		"6", `<a href="m.htm">`,
		"7", `<img src="i.jpg"/>`,
		"8", `<br/>`,
		"9", `<hr/>`,
	)
	filterReplacer := strings.NewReplacer(
		"1", `  TEXT  `,
		"2", `  COMMENT  `,
		"3", `  p  `,
		"4", `  div  `,
		"5", `  span  `,
		"6", `  a  `,
		"7", `  img  `,
		"8", `  br  `,
		"9", `  hr  `,
	)

	r := proptest.Rand(150)
	for range 200 {
		s := r.Chars("123456789")
		charFilter := r.Chars("123456789")

		want := htmlReplacer.Replace(charExtract(strings.Repeat(s, 4), charFilter))
		got := strings.Repeat(htmlExtract(
			htmlReplacer.Replace(s),
			filterReplacer.Replace(charFilter),
		), 4)

		assertEqual(t, got, want)
		if t.Failed() {
			break
		}
	}
}
