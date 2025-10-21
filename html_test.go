package jqx

import (
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

	assert := func(selector, want string) {
		t.Helper()

		got, err := htmlQuerySelector(html, selector)
		if err != nil {
			got = "error"
		}
		assertEqual(t, got, want)
	}

	assert(`.content`, `<p class="content">Second paragraph.</p>`)
	assert(`h1`, `<h1 class="header">Title</h1>`)
	assert(`p`, `<p>First paragraph.</p><p class="content">Second paragraph.</p>`)

	assert(`invalid[`, "error")
}

func TestHtmlExtractText(t *testing.T) {
	assert := func(html, want string) {
		t.Helper()
		assertEqual(t, htmlExtractText(html), want)
	}

	assert(
		`<html><body><p>Hello, world!</p></body></html>`,
		"Hello, world!",
	)
	assert(
		`<h1>Title</h1><p>First paragraph.</p><p>Second paragraph.</p>`,
		"TitleFirst paragraph.Second paragraph.",
	)
	assert(
		`<div><img src="image.jpg"/></div>`,
		"",
	)
	assert(
		"",
		"",
	)
	assert(
		`<!-- This is a comment --><body><p>Some text</p></body>`,
		"Some text",
	)
	assert(
		"Hello",
		"Hello",
	)
	assert(
		`<p>Hello</p><img src="image.jpg" alt="olleh"/>`,
		"Hello\n(img: olleh)\n",
	)
}
