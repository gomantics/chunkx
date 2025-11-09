package chunkx

import (
	"strings"
)

// TokenCounter defines the interface for counting tokens in text.
type TokenCounter interface {
	CountTokens(text string) (int, error)
}

// SimpleTokenCounter provides a basic whitespace-based token counting implementation.
type SimpleTokenCounter struct{}

// CountTokens returns the number of whitespace-separated words in the text.
func (s *SimpleTokenCounter) CountTokens(text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	return len(strings.Fields(text)), nil
}

// ByteCounter counts bytes instead of tokens.
type ByteCounter struct{}

// CountTokens returns the number of bytes in the text.
func (b *ByteCounter) CountTokens(text string) (int, error) {
	return len(text), nil
}

// LineCounter counts lines instead of tokens.
type LineCounter struct{}

// CountTokens returns the number of lines in the text.
func (l *LineCounter) CountTokens(text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	lines := strings.Count(text, "\n")
	// If text doesn't end with newline, add 1
	if !strings.HasSuffix(text, "\n") {
		lines++
	}
	return lines, nil
}
