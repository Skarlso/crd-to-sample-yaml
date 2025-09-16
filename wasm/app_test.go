package main

import (
	"testing"
)

func TestParseDescription_WASM(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty description",
			input:    "",
			expected: "",
		},
		{
			name:     "simple paragraph",
			input:    "This is a simple description.",
			expected: "<p>This is a simple description.</p>\n",
		},
		{
			name:     "multiple paragraphs",
			input:    "First paragraph.\n\nSecond paragraph.",
			expected: "<p>First paragraph.</p>\n<p>Second paragraph.</p>\n",
		},
		{
			name:     "simple bullet list",
			input:    "Features:\n- Feature one\n- Feature two\n- Feature three",
			expected: "<p>Features:</p>\n<ul>\n<li>Feature one</li>\n<li>Feature two</li>\n<li>Feature three</li>\n</ul>\n",
		},
		{
			name:     "bullet list with different markers",
			input:    "Options:\n* Option A\n+ Option B\n• Option C",
			expected: "<p>Options:</p>\n<ul>\n<li>Option A</li>\n<li>Option B</li>\n<li>Option C</li>\n</ul>\n",
		},
		{
			name:     "mixed content",
			input:    "Configuration options:\n\nSupported features:\n- Authentication\n- Authorization\n- Monitoring\n\nAdditional notes.",
			expected: "<p>Configuration options:</p>\n<p>Supported features:</p>\n<ul>\n<li>Authentication</li>\n<li>Authorization</li>\n<li>Monitoring</li>\n</ul>\n<p>Additional notes.</p>\n",
		},
		{
			name:     "HTML escaping",
			input:    "HTML: <script>alert('test')</script>\n\n- Item with & ampersand\n- Item with \"quotes\"",
			expected: "<p>HTML: &lt;script&gt;alert(&#39;test&#39;)&lt;/script&gt;</p>\n<ul>\n<li>Item with &amp; ampersand</li>\n<li>Item with &quot;quotes&quot;</li>\n</ul>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDescription(tt.input)
			if result != tt.expected {
				t.Errorf("parseDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseDescriptionElements(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		description   string
	}{
		{
			name:          "empty description",
			input:         "",
			expectedCount: 0,
			description:   "should return no elements",
		},
		{
			name:          "simple paragraph",
			input:         "This is a simple description.",
			expectedCount: 1,
			description:   "should return one paragraph element",
		},
		{
			name:          "multiple paragraphs",
			input:         "First paragraph.\n\nSecond paragraph.",
			expectedCount: 2,
			description:   "should return two paragraph elements",
		},
		{
			name:          "simple bullet list",
			input:         "Features:\n- Feature one\n- Feature two",
			expectedCount: 2,
			description:   "should return paragraph and list elements",
		},
		{
			name:          "mixed content",
			input:         "Configuration:\n\nFeatures:\n- Auth\n- Monitor\n\nNotes.",
			expectedCount: 4,
			description:   "should return paragraph, paragraph, list, paragraph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDescriptionElements(tt.input)
			if len(result) != tt.expectedCount {
				t.Errorf("parseDescriptionElements() returned %d elements, want %d (%s)", len(result), tt.expectedCount, tt.description)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "HTML tags",
			input:    "<script>alert('test')</script>",
			expected: "&lt;script&gt;alert(&#39;test&#39;)&lt;/script&gt;",
		},
		{
			name:     "ampersand",
			input:    "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
		{
			name:     "quotes",
			input:    `He said "Hello" and she said 'Hi'`,
			expected: "He said &quot;Hello&quot; and she said &#39;Hi&#39;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}