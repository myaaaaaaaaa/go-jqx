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
			</body>
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

	// Test: Invalid HTML is parsed leniently
	invalidHTML := `<html><body><p>Invalid HTML`
	selector = "p"
	expected = `<p>Invalid HTML</p>`
	assertEqual(t, must(htmlQuerySelector(invalidHTML, selector)), expected)

	// Test: Invalid selector
	selector = "invalid["
	_, err := htmlQuerySelector(html, selector)
	if err == nil {
		t.Errorf("Expected an error for invalid selector '%s', but got nil", selector)
	}
}

func TestHtmlExtractText(t *testing.T) {
	// Test case 1: Simple HTML
	html1 := `<html><body><p>Hello, world!</p></body></html>`
	expected1 := "Hello, world!"
	assertEqual(t, must(htmlExtractText(html1)), expected1)

	// Test case 2: HTML with multiple text nodes
	html2 := `<h1>Title</h1><p>First paragraph.</p><p>Second paragraph.</p>`
	expected2 := "TitleFirst paragraph.Second paragraph."
	assertEqual(t, must(htmlExtractText(html2)), expected2)

	// Test case 3: HTML with no text
	html3 := `<div><img src="image.jpg"/></div>`
	expected3 := ""
	assertEqual(t, must(htmlExtractText(html3)), expected3)

	// Test case 4: Empty string
	html4 := ""
	expected4 := ""
	assertEqual(t, must(htmlExtractText(html4)), expected4)

	// Test case 5: HTML with comments
	html5 := `<!-- This is a comment --><body><p>Some text</p></body>`
	expected5 := "Some text"
	assertEqual(t, must(htmlExtractText(html5)), expected5)
}
