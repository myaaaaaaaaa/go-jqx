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

	// Test case 1: Select by class
	selector := ".content"
	expected := `<p class="content">Second paragraph.</p>`
	actual, err := selectHTML(html, selector)
	if err != nil {
		t.Errorf("Unexpected error for selector '%s': %v", selector, err)
	}
	if actual != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, actual)
	}

	// Test case 2: Select by tag
	selector = "h1"
	expected = `<h1 class="header">Title</h1>`
	actual, err = selectHTML(html, selector)
	if err != nil {
		t.Errorf("Unexpected error for selector '%s': %v", selector, err)
	}
	if actual != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, actual)
	}

	// Test case 3: Select multiple elements
	selector = "p"
	expected = `<p>First paragraph.</p><p class="content">Second paragraph.</p>`
	actual, err = selectHTML(html, selector)
	if err != nil {
		t.Errorf("Unexpected error for selector '%s': %v", selector, err)
	}
	if actual != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, actual)
	}

	// Test case 4: Invalid selector
	selector = "invalid["
	_, err = selectHTML(html, selector)
	if err == nil {
		t.Errorf("Expected an error for invalid selector '%s', but got nil", selector)
	}
}