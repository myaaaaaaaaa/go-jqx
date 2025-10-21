package jqx

import (
	"testing"
)

func TestSelectHTML(t *testing.T) {
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
	assertEqual(t, must(selectHTML(html, selector)), expected)

	// Test: Select by tag
	selector = "h1"
	expected = `<h1 class="header">Title</h1>`
	assertEqual(t, must(selectHTML(html, selector)), expected)

	// Test: Select multiple elements
	selector = "p"
	expected = `<p>First paragraph.</p><p class="content">Second paragraph.</p>`
	assertEqual(t, must(selectHTML(html, selector)), expected)

	// Test: Invalid HTML is parsed leniently
	invalidHTML := `<html><body><p>Invalid HTML`
	selector = "p"
	expected = `<p>Invalid HTML</p>`
	assertEqual(t, must(selectHTML(invalidHTML, selector)), expected)

	// Test: Invalid selector
	selector = "invalid["
	_, err := selectHTML(html, selector)
	if err == nil {
		t.Errorf("Expected an error for invalid selector '%s', but got nil", selector)
	}
}
