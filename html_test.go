package jqx

import (
	"slices"
	"strings"
	"testing"
)

func TestHTMLError(t *testing.T) {
	const html = `<html><body><div><p>hello</p></div></body></html>`
	var err error

	must(htmlQuerySelector(html, `valid`))
	must(htmlReplaceSelector(html, `valid`, ""))

	_, err = htmlQuerySelector(html, `invalid[`)
	assertEqual(t, err != nil, true)
	_, err = htmlReplaceSelector(html, `invalid[`, "")
	assertEqual(t, err != nil, true)
}

func TestHTMLQuerySelector(t *testing.T) {
	const html = `
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

		got := slices.Collect(query(selector))
		assertEqual(t, len(got), 1)
		assertString(t, got[0], want)
	}

	assert(`.content`, `<p class="content">Second paragraph.</p>`)
	assert(`h1`, `<h1 class="header">Title</h1>`)
	assert(`p`, `<p>First paragraph.</p>  ;  <p class="content">Second paragraph.</p>`)
}

func checkHTMLReplaceSelector(htmlString, cssSelector, replacement string) string {
	rt := must(htmlReplaceSelector(htmlString, cssSelector, replacement))

	cleaned := must(htmlReplaceSelector(htmlString, "nonmatching", replacement))
	{
		cleaned2 := must(htmlReplaceSelector(htmlString, "nonmatching", ""))
		if cleaned != cleaned2 {
			panic(cleaned2)
		}
		cleaned2 = must(htmlReplaceSelector(htmlString, cssSelector, "<>"))
		if cleaned != cleaned2 {
			panic(cleaned2)
		}
		cleaned2 = must(htmlReplaceSelector(cleaned, cssSelector, "<>"))
		if cleaned != cleaned2 {
			panic(cleaned2)
		}
	}

	rt2 := must(htmlReplaceSelector(cleaned, cssSelector, replacement))
	if rt != rt2 {
		panic(rt2)
	}

	return rt
}
func checkHTMLDeleteSelector(htmlString, cssSelector string) string {
	rt := must(htmlReplaceSelector(htmlString, cssSelector, ""))
	empty := must(htmlQuerySelector(rt, cssSelector))
	if len(empty) != 0 {
		panic(strings.Join(empty, " "))
	}
	rt2 := must(htmlReplaceSelector(rt, cssSelector, ""))
	if rt != rt2 {
		panic(rt2)
	}
	return rt
}

func TestHtmlReplaceSelector(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		selector    string
		replacement string
		expected    string
	}{
		{
			name:        "simple replace",
			html:        `<html><body><div><p>hello</p></div></body></html>`,
			selector:    "div",
			replacement: "<span>world</span>",
			expected:    `<html><head></head><body><span>world</span></body></html>`,
		},
		{
			name:        "multiple replace",
			html:        `<html><body><ul><li>1</li><li>2</li></ul></body></html>`,
			selector:    "li",
			replacement: "<p>replaced</p>",
			expected:    `<html><head></head><body><ul><p>replaced</p><p>replaced</p></ul></body></html>`,
		},
		{
			name:        "template",
			html:        `<html><body><ul><li>1</li><li>2</li></ul></body></html>`,
			selector:    "li",
			replacement: "  <>  ",
			expected:    `<html><head></head><body><ul>  <li>1</li>    <li>2</li>  </ul></body></html>`,
		},
		{
			name:        "no match",
			html:        `<html><body><div><p>hello</p></div></body></html>`,
			selector:    "span",
			replacement: "<span>world</span>",
			expected:    `<html><head></head><body><div><p>hello</p></div></body></html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := checkHTMLReplaceSelector(tt.html, tt.selector, tt.replacement)
			assertEqual(t, actual, tt.expected)
		})
	}
}
func TestHtmlDeleteSelector(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		expected string
	}{
		{
			name:     "simple delete",
			html:     `<html><body><div><p>hello</p></div></body></html>`,
			selector: "p",
			expected: `<html><head></head><body><div></div></body></html>`,
		},
		{
			name:     "multiple delete",
			html:     `<html><body><ul><li>1</li><li>2</li></ul></body></html>`,
			selector: "li",
			expected: `<html><head></head><body><ul></ul></body></html>`,
		},
		{
			name:     "no match",
			html:     `<html><body><div><p>hello</p></div></body></html>`,
			selector: "span",
			expected: `<html><head></head><body><div><p>hello</p></div></body></html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := checkHTMLDeleteSelector(tt.html, tt.selector)
			assertEqual(t, actual, tt.expected)
		})
	}
}
