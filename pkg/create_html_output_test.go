package pkg

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDescription(t *testing.T) {
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
			name:     "mixed content with paragraphs and lists",
			input:    "Configuration options:\n\nSupported features:\n- Authentication\n- Authorization\n- Monitoring\n\nAdditional notes about usage.",
			expected: "<p>Configuration options:</p>\n<p>Supported features:</p>\n<ul>\n<li>Authentication</li>\n<li>Authorization</li>\n<li>Monitoring</li>\n</ul>\n<p>Additional notes about usage.</p>\n",
		},
		{
			name:     "paragraph with list in between",
			input:    "Start paragraph.\n\nList items:\n- Item 1\n- Item 2\n\nEnd paragraph.",
			expected: "<p>Start paragraph.</p>\n<p>List items:</p>\n<ul>\n<li>Item 1</li>\n<li>Item 2</li>\n</ul>\n<p>End paragraph.</p>\n",
		},
		{
			name:     "indented bullet points",
			input:    "  - Indented item one\n  - Indented item two",
			expected: "<ul>\n<li>Indented item one</li>\n<li>Indented item two</li>\n</ul>\n",
		},
		{
			name:     "multiline paragraph followed by list",
			input:    "This is a long description\nthat spans multiple lines\nbefore the list.\n\n- First item\n- Second item",
			expected: "<p>This is a long description that spans multiple lines before the list.</p>\n<ul>\n<li>First item</li>\n<li>Second item</li>\n</ul>\n",
		},
		{
			name:     "special characters in content",
			input:    "HTML characters: <script>alert('test')</script>\n\n- Item with & ampersand\n- Item with \"quotes\"",
			expected: "<p>HTML characters: &lt;script&gt;alert(&#39;test&#39;)&lt;/script&gt;</p>\n<ul>\n<li>Item with &amp; ampersand</li>\n<li>Item with &#34;quotes&#34;</li>\n</ul>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDescription(tt.input)
			assert.Equal(t, template.HTML(tt.expected), result)
		})
	}
}