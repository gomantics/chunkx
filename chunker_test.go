package chunkx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/gomantics/chunkx/languages"
)

func TestChunker_Chunk(t *testing.T) {
	chunker := NewChunker()

	tests := []struct {
		name          string
		code          string
		language      languages.LanguageName
		maxSize       int
		wantMinChunks int
		wantMaxChunks int
		wantErr       bool
	}{
		{
			name: "small Go function fits in one chunk",
			code: `func hello() {
	fmt.Println("Hello, World!")
}`,
			language:      languages.Go,
			maxSize:       100,
			wantMinChunks: 1,
			wantMaxChunks: 1,
			wantErr:       false,
		},
		{
			name: "larger Go code splits into multiple chunks",
			code: `package main

import "fmt"

func main() {
	fmt.Println("Line 1")
	fmt.Println("Line 2")
	fmt.Println("Line 3")
}

func helper1() {
	fmt.Println("Helper 1")
}

func helper2() {
	fmt.Println("Helper 2")
}`,
			language:      languages.Go,
			maxSize:       10,
			wantMinChunks: 2,
			wantMaxChunks: 10,
			wantErr:       false,
		},
		{
			name: "Python class splits appropriately",
			code: `class Example:
	def __init__(self):
		self.value = 0
	
	def method1(self):
		return self.value * 2
	
	def method2(self):
		return self.value * 3`,
			language:      languages.Python,
			maxSize:       15,
			wantMinChunks: 2,
			wantMaxChunks: 5,
			wantErr:       false,
		},
		{
			name:          "missing language",
			code:          `func test() {}`,
			language:      "",
			maxSize:       100,
			wantMinChunks: 0,
			wantMaxChunks: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Option{WithLanguage(tt.language)}
			if tt.maxSize > 0 {
				opts = append(opts, WithMaxSize(tt.maxSize))
			}

			chunks, err := chunker.Chunk(tt.code, opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(chunks) < tt.wantMinChunks || len(chunks) > tt.wantMaxChunks {
					t.Errorf("Chunk() returned %d chunks, want between %d and %d",
						len(chunks), tt.wantMinChunks, tt.wantMaxChunks)
				}

				// Verify chunks are non-empty and have proper metadata
				for i, chunk := range chunks {
					if chunk.Content == "" {
						t.Errorf("chunk %d has empty content", i)
					}
					if chunk.Language != tt.language {
						t.Errorf("chunk %d language = %q, want %q", i, chunk.Language, tt.language)
					}
					if len(chunk.NodeTypes) == 0 {
						t.Errorf("chunk %d has no node types", i)
					}
				}
			}
		})
	}
}

func TestChunker_WithOverlap(t *testing.T) {
	chunker := NewChunker()

	code := `func function1() {
	fmt.Println("Function 1")
}

func function2() {
	fmt.Println("Function 2")
}

func function3() {
	fmt.Println("Function 3")
}`

	// Chunk with 20% overlap
	chunks, err := chunker.Chunk(code,
		WithLanguage(languages.Go),
		WithMaxSize(10),
		WithOverlap(20))
	if err != nil {
		t.Fatalf("failed to chunk with overlap: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	// Verify that consecutive chunks share overlapping content
	for i := 0; i < len(chunks)-1; i++ {
		currentChunk := chunks[i]
		nextChunk := chunks[i+1]

		// The current chunk should contain some content from the next chunk
		// Get the last portion of the current chunk
		currentLines := strings.Split(currentChunk.Content, "\n")
		nextLines := strings.Split(nextChunk.Content, "\n")

		// Check if there's any overlap by looking for shared content
		hasOverlap := false
		for _, currentLine := range currentLines {
			trimmed := strings.TrimSpace(currentLine)
			if trimmed == "" {
				continue
			}
			for _, nextLine := range nextLines {
				if strings.TrimSpace(nextLine) == trimmed {
					hasOverlap = true
					break
				}
			}
			if hasOverlap {
				break
			}
		}

		if !hasOverlap {
			t.Errorf("chunk %d should overlap with chunk %d, but no shared content found\nChunk %d:\n%s\n\nChunk %d:\n%s",
				i, i+1, i, currentChunk.Content, i+1, nextChunk.Content)
		}
	}

	// Verify that total content with overlap is greater than individual chunks
	// (because of the redundant overlapping sections)
	totalSize := 0
	for _, chunk := range chunks {
		totalSize += len(chunk.Content)
	}

	// The total size should be greater than the original code due to overlaps
	if totalSize < len(code) {
		t.Errorf("total overlapped content (%d bytes) should be >= original code (%d bytes)",
			totalSize, len(code))
	}
}

func TestChunker_CustomTokenCounter(t *testing.T) {
	chunker := NewChunker()

	// Custom counter that counts semicolons as tokens
	customCounter := &semicolonCounter{}

	code := `fmt.Println("1"); fmt.Println("2"); fmt.Println("3");`

	chunks, err := chunker.Chunk(code,
		WithLanguage(languages.Go),
		WithMaxSize(2), // Max 2 semicolons per chunk
		WithTokenCounter(customCounter))
	if err != nil {
		t.Fatalf("Chunk() with custom counter failed: %v", err)
	}

	// Should split into at least 2 chunks (3 semicolons total, max 2 per chunk)
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks, got %d", len(chunks))
	}
}

func TestChunker_PreservesSemanticBoundaries(t *testing.T) {
	chunker := NewChunker()

	// Code with clear function boundaries
	code := `package main

func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}

func multiply(a, b int) int {
	return a * b
}`

	chunks, err := chunker.Chunk(code,
		WithLanguage(languages.Go),
		WithMaxSize(15)) // Small enough to force splitting
	if err != nil {
		t.Fatalf("Chunk() failed: %v", err)
	}

	// Each chunk should contain complete functions (not split mid-function)
	for i, chunk := range chunks {
		// Count opening and closing braces
		openBraces := strings.Count(chunk.Content, "{")
		closeBraces := strings.Count(chunk.Content, "}")

		// In properly chunked code, braces should be balanced
		if openBraces != closeBraces {
			t.Errorf("chunk %d has unbalanced braces: %d open, %d close\nContent: %s",
				i, openBraces, closeBraces, chunk.Content)
		}
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"single line", "hello", 1},
		{"two lines", "hello\nworld", 2},
		{"trailing newline", "hello\n", 2},
		{"multiple newlines", "a\n\n\nb", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countLines(tt.input)
			if got != tt.expected {
				t.Errorf("countLines(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestChunker_Generic(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		maxSize  int
		expected int // expected number of chunks
	}{
		{
			name: "simple generic text",
			code: `This is a simple text file.
It has multiple lines.
Each line should be grouped together.
Until the max size is reached.`,
			maxSize:  20,
			expected: 2,
		},
		{
			name: "small generic text fits in one chunk",
			code: `Short text
That fits`,
			maxSize:  100,
			expected: 1,
		},
		{
			name: "generic with multiple chunks",
			code: `Line 1
Line 2
Line 3
Line 4
Line 5`,
			maxSize:  3,
			expected: 5, // Each line has 2 tokens, so maxSize 3 fits 1 line per chunk
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewChunker()
			chunks, err := chunker.Chunk(tt.code,
				WithLanguage(languages.Generic),
				WithMaxSize(tt.maxSize))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(chunks) != tt.expected {
				t.Errorf("expected %d chunks, got %d", tt.expected, len(chunks))
			}

			// Verify all chunks have generic language
			for i, chunk := range chunks {
				if chunk.Language != languages.Generic {
					t.Errorf("chunk %d has wrong language: %s", i, chunk.Language)
				}
				if len(chunk.NodeTypes) != 1 || chunk.NodeTypes[0] != "generic" {
					t.Errorf("chunk %d has wrong node types: %v", i, chunk.NodeTypes)
				}
			}
		})
	}
}

// Helper token counter for testing
type semicolonCounter struct{}

func (s *semicolonCounter) CountTokens(text string) (int, error) {
	return strings.Count(text, ";"), nil
}

// ChunkingResult represents the complete output for a chunked file
type ChunkingResult struct {
	File     string  `json:"file"`
	Language string  `json:"language"`
	Chunks   []Chunk `json:"chunks"`
}

// TestChunkingExamples tests chunking of real code examples and creates
// human-readable approval snapshots in JSON format
func TestChunkingExamples(t *testing.T) {
	sourcesDir := "testdata/sources"

	// Read all source files
	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		t.Fatalf("Failed to read sources directory: %v", err)
	}

	chunker := NewChunker()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		filepath := filepath.Join(sourcesDir, filename)

		t.Run(filename, func(t *testing.T) {
			// Chunk the file using ChunkFile for automatic language detection
			chunks, err := chunker.ChunkFile(filepath, WithMaxSize(50))
			if err != nil {
				t.Fatalf("Failed to chunk %s: %v", filename, err)
			}

			// Determine language from chunks (all should have same language)
			language := ""
			if len(chunks) > 0 {
				language = string(chunks[0].Language)
			}

			// Create result structure
			result := ChunkingResult{
				File:     filename,
				Language: language,
				Chunks:   chunks,
			}

			// Use approvals to verify the JSON structure
			// This will create filename.approved.json on first run
			approvals.VerifyJSONStruct(t, result)
		})
	}
}

// TestChunkingExamplesWithOverlap tests chunking with overlap enabled
func TestChunkingExamplesWithOverlap(t *testing.T) {
	approvals.UseFolder("testdata")

	// Test just the Go example with overlap to show the difference
	filepath := "testdata/sources/example.go"

	chunker := NewChunker()
	chunks, err := chunker.ChunkFile(filepath, WithMaxSize(50), WithOverlap(20))
	if err != nil {
		t.Fatalf("Failed to chunk with overlap: %v", err)
	}

	result := ChunkingResult{
		File:     "example.go",
		Language: string(chunks[0].Language),
		Chunks:   chunks,
	}

	approvals.VerifyJSONStruct(t, result)
}
