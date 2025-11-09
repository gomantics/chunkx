package chunkx

import (
	"testing"
)

func TestSimpleTokenCounter(t *testing.T) {
	counter := &SimpleTokenCounter{}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"single word", "hello", 1},
		{"multiple words", "hello world foo bar", 4},
		{"with newlines", "hello\nworld", 2},
		{"with tabs", "hello\tworld", 2},
		{"multiple spaces", "hello   world", 2},
		{"code snippet", "func main() { fmt.Println(\"hello\") }", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := counter.CountTokens(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tt.expected {
				t.Errorf("expected %d tokens, got %d", tt.expected, count)
			}
		})
	}
}

func TestByteCounter(t *testing.T) {
	counter := &ByteCounter{}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"single char", "a", 1},
		{"hello world", "hello world", 11},
		{"unicode", "hello 世界", 12}, // "hello " = 6 bytes, "世界" = 6 bytes (3 bytes per char)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := counter.CountTokens(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tt.expected {
				t.Errorf("expected %d bytes, got %d", tt.expected, count)
			}
		})
	}
}

func TestLineCounter(t *testing.T) {
	counter := &LineCounter{}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"single line no newline", "hello world", 1},
		{"single line with newline", "hello world\n", 1},
		{"multiple lines", "line1\nline2\nline3", 3},
		{"multiple lines with trailing newline", "line1\nline2\nline3\n", 3},
		{"empty lines", "line1\n\nline3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := counter.CountTokens(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tt.expected {
				t.Errorf("expected %d lines, got %d", tt.expected, count)
			}
		})
	}
}
