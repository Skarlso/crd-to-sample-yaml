package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// MaxCSSFileSize limits CSS file size to 1MB to prevent abuse.
	MaxCSSFileSize = 1024 * 1024
)

var (
	// Dangerous CSS patterns that should be blocked.
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)javascript:`),                    // javascript: URLs
		regexp.MustCompile(`(?i)expression\s*\(`),                // CSS expression() function (IE)
		regexp.MustCompile(`(?i)@import\s+url\s*\(\s*['"]*http`), // External @import URLs
		regexp.MustCompile(`(?i)@import\s+url\s*\(\s*['"]*//`),   // Protocol-relative @import URLs
		regexp.MustCompile(`(?i)behavior\s*:`),                   // IE behavior property
		regexp.MustCompile(`(?i)-moz-binding\s*:`),               // Firefox binding property
		regexp.MustCompile(`(?i)vbscript:`),                      // VBScript URLs
		regexp.MustCompile(`(?i)data:\s*text/html`),              // HTML data URLs
		regexp.MustCompile(`(?i)mocha:`),                         // Mocha protocol
		regexp.MustCompile(`(?i)livescript:`),                    // LiveScript protocol
	}

	// HTML-like patterns that should be removed.
	htmlPatterns = []*regexp.Regexp{
		regexp.MustCompile(`<[^>]*>`),         // HTML tags
		regexp.MustCompile(`&[a-zA-Z0-9#]+;`), // HTML entities
	}

	// cssCommentPattern = regexp.MustCompile(`/\*[\s\S]*?\*/`).
)

// SanitizeCSS reads and sanitizes a CSS file, removing dangerous content.
func SanitizeCSS(cssFilePath string) (string, error) {
	if cssFilePath == "" {
		return "", nil
	}

	fileInfo, err := os.Stat(cssFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read CSS file info: %w", err)
	}

	if fileInfo.Size() > MaxCSSFileSize {
		return "", fmt.Errorf("CSS file is too large (max %d bytes)", MaxCSSFileSize)
	}

	cssContent, err := os.ReadFile(filepath.Clean(cssFilePath))
	if err != nil {
		return "", fmt.Errorf("failed to read CSS file: %w", err)
	}

	return sanitizeCSSContent(string(cssContent))
}

// sanitizeCSSContent sanitizes CSS content by removing dangerous patterns.
func sanitizeCSSContent(css string) (string, error) {
	for _, pattern := range htmlPatterns {
		css = pattern.ReplaceAllString(css, "")
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(css) {
			return "", fmt.Errorf("CSS contains dangerous pattern: %s", pattern.String())
		}
	}

	if err := validateCSSStructure(css); err != nil {
		return "", fmt.Errorf("invalid CSS structure: %w", err)
	}

	css = strings.TrimSpace(css)

	return css, nil
}

// validateCSSStructure performs basic CSS structure validation.
func validateCSSStructure(css string) error {
	openBraces := strings.Count(css, "{")
	closeBraces := strings.Count(css, "}")

	if openBraces != closeBraces {
		return fmt.Errorf("unbalanced braces in CSS (open: %d, close: %d)", openBraces, closeBraces)
	}

	lines := strings.Split(css, "\n")
	for i, line := range lines {
		const maxSize = 10000
		if len(line) > maxSize { // 10KB per line max.
			return fmt.Errorf("line %d is too long (%d characters)", i+1, len(line))
		}
	}

	return nil
}
