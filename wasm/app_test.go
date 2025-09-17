package main

import (
	"testing"
)

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
