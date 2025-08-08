package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeCSS(t *testing.T) {
	tests := []struct {
		name        string
		cssContent  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_css",
			cssContent: `.my-class {
				color: red;
				background: blue;
				border-radius: 5px;
			}`,
			expectError: false,
		},
		{
			name:        "javascript_url",
			cssContent:  `.malicious { background: url(javascript:alert('xss')); }`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "css_expression",
			cssContent:  `.malicious { width: expression(alert('xss')); }`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "external_import",
			cssContent:  `@import url("http://evil.com/malicious.css");`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "protocol_relative_import",
			cssContent:  `@import url("//evil.com/malicious.css");`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "html_tags",
			cssContent:  `.test { color: red; } <script>alert('xss')</script>`,
			expectError: false, // HTML should be stripped, not cause error
		},
		{
			name:        "unbalanced_braces",
			cssContent:  `.test { color: red;`,
			expectError: true,
			errorMsg:    "unbalanced braces",
		},
		{
			name:        "behavior_property",
			cssContent:  `.test { behavior: url(something.htc); }`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "moz_binding",
			cssContent:  `.test { -moz-binding: url(something.xml); }`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "vbscript_url",
			cssContent:  `.test { background: url(vbscript:msgbox('xss')); }`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "data_html_url",
			cssContent:  `.test { background: url(data:text/html,<script>alert('xss')</script>); }`,
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name: "valid_css_with_comments",
			cssContent: `/* This is a comment */
			.test {
				color: red; /* inline comment */
				background: blue;
			}`,
			expectError: false,
		},
		{
			name: "empty_css",
			cssContent: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary CSS file
			tmpDir := t.TempDir()
			cssFile := filepath.Join(tmpDir, "test.css")
			
			err := os.WriteFile(cssFile, []byte(tt.cssContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test CSS file: %v", err)
			}

			result, err := SanitizeCSS(cssFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == "" && tt.cssContent != "" {
					t.Errorf("Expected non-empty result for valid CSS")
				}
			}
		})
	}
}

func TestSanitizeCSSContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "remove_html_tags",
			input:    `.test { color: red; } <script>alert('xss')</script>`,
			expected: `.test { color: red; } alert('xss')`,
			hasError: false,
		},
		{
			name:     "remove_html_entities",
			input:    `.test { content: "&lt;hello&gt;"; }`,
			expected: `.test { content: "hello"; }`,
			hasError: false,
		},
		{
			name:     "preserve_valid_css",
			input:    `.test { color: #ff0000; background: rgba(0,0,0,0.5); }`,
			expected: `.test { color: #ff0000; background: rgba(0,0,0,0.5); }`,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizeCSSContent(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSanitizeCSSFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	cssFile := filepath.Join(tmpDir, "large.css")

	// Create a CSS file that's too large
	largeContent := strings.Repeat(".test { color: red; }\n", 100000) // Should exceed 1MB
	err := os.WriteFile(cssFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large CSS file: %v", err)
	}

	_, err = SanitizeCSS(cssFile)
	if err == nil {
		t.Errorf("Expected error for oversized CSS file")
	} else if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' error, got: %v", err)
	}
}

func TestSanitizeCSSEmptyFile(t *testing.T) {
	result, err := SanitizeCSS("")
	if err != nil {
		t.Errorf("Unexpected error for empty filename: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result for empty filename, got: %s", result)
	}
}

func TestSanitizeCSSNonexistentFile(t *testing.T) {
	_, err := SanitizeCSS("/nonexistent/path/to/file.css")
	if err == nil {
		t.Errorf("Expected error for nonexistent file")
	}
}

func TestValidateCSSStructure(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		hasError bool
		errorMsg string
	}{
		{
			name:     "balanced_braces",
			css:      `.test { color: red; } .other { background: blue; }`,
			hasError: false,
		},
		{
			name:     "unbalanced_open",
			css:      `.test { color: red; .other { background: blue; }`,
			hasError: true,
			errorMsg: "unbalanced braces",
		},
		{
			name:     "unbalanced_close",
			css:      `.test { color: red; } } .other { background: blue; }`,
			hasError: true,
			errorMsg: "unbalanced braces",
		},
		{
			name:     "very_long_line",
			css:      `.test { ` + strings.Repeat("a", 15000) + `: red; }`,
			hasError: true,
			errorMsg: "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCSSStructure(tt.css)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}