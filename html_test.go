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

	// Test: Select by class
	selector := ".content"
	expected := `<p class="content">Second paragraph.</p>`
	assertEqual(t, must(htmlQuerySelector(html, selector)), expected)

	// Test: Select by tag
	selector = "h1"
	expected = `<h1 class="header">Title</h1>`
	assertEqual(t, must(htmlQuerySelector(html, selector)), expected)

	// Test: Select multiple elements
	selector = "p"
	expected = `<p>First paragraph.</p><p class="content">Second paragraph.</p>`
	assertEqual(t, must(htmlQuerySelector(html, selector)), expected)

	// Test: Invalid selector
	selector = "invalid["
	_, err := htmlQuerySelector(html, selector)
	if err == nil {
		t.Errorf("Expected an error for invalid selector '%s', but got nil", selector)
	}
}

func TestHtmlExtractText(t *testing.T) {
	// Test: Simple HTML
	html1 := `<html><body><p>Hello, world!</p></body></html>`
	expected1 := "Hello, world!"
	assertEqual(t, must(htmlExtractText(html1)), expected1)

	// Test: HTML with multiple text nodes
	html2 := `<h1>Title</h1><p>First paragraph.</p><p>Second paragraph.</p>`
	expected2 := "TitleFirst paragraph.Second paragraph."
	assertEqual(t, must(htmlExtractText(html2)), expected2)

	// Test: HTML with no text
	html3 := `<div><img src="image.jpg"/></div>`
	expected3 := ""
	assertEqual(t, must(htmlExtractText(html3)), expected3)

	// Test: Empty string
	html4 := ""
	expected4 := ""
	assertEqual(t, must(htmlExtractText(html4)), expected4)

	// Test: HTML with comments
	html5 := `<!-- This is a comment --><body><p>Some text</p></body>`
	expected5 := "Some text"
	assertEqual(t, must(htmlExtractText(html5)), expected5)

	// Test: Plain text
	html6 := "Hello"
	expected6 := "Hello"
	assertEqual(t, must(htmlExtractText(html6)), expected6)

	// Test: Alt text
	html7 := `<p>Hello</p><img src="image.jpg" alt="olleh"/>`
	expected7 := "Hello\n(img: olleh)\n"
	assertEqual(t, must(htmlExtractText(html7)), expected7)
}
