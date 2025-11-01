package jqx

import (
	"testing"
)

func TestHtmlReplaceSelector(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		selector      string
		replacement   string
		expected      string
		expectedError bool
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
			name:        "no match",
			html:        `<html><body><div><p>hello</p></div></body></html>`,
			selector:    "span",
			replacement: "<span>world</span>",
			expected:    `<html><head></head><body><div><p>hello</p></div></body></html>`,
		},
		{
			name:          "invalid selector",
			html:          `<html><body><div><p>hello</p></div></body></html>`,
			selector:      "[",
			replacement:   "<span>world</span>",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := htmlReplaceSelector(tt.html, tt.selector, tt.replacement)

			if (err != nil) != tt.expectedError {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if !tt.expectedError && actual != tt.expected {
				t.Fatalf("expected:\n%s\ngot:\n%s", tt.expected, actual)
			}
		})
	}
}
