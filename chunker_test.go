package chunkx

import (
	"strings"
	"testing"

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

func TestIntegration_RealGoCode(t *testing.T) {
	chunker := NewChunker()

	// Real Go code example
	code := `package main

import (
	"fmt"
	"net/http"
	"log"
)

type Server struct {
	port string
	handler http.Handler
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Start() error {
	log.Printf("Starting server on port %s", s.port)
	return http.ListenAndServe(":"+s.port, s.handler)
}

func (s *Server) SetHandler(h http.Handler) {
	s.handler = h
}

func main() {
	server := NewServer("8080")
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	
	server.SetHandler(handler)
	
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}`

	chunks, err := chunker.Chunk(code, WithLanguage(languages.Go), WithMaxSize(50))
	if err != nil {
		t.Fatalf("failed to chunk Go code: %v", err)
	}

	// Verify chunking properties
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks for large code, got %d", len(chunks))
	}

	// Verify that chunks don't break in the middle of functions
	for i, chunk := range chunks {
		// Simple heuristic: if chunk contains "func", it should contain the closing brace
		if strings.Contains(chunk.Content, "func") && strings.Contains(chunk.Content, "{") {
			openBraces := strings.Count(chunk.Content, "{")
			closeBraces := strings.Count(chunk.Content, "}")
			if openBraces > closeBraces {
				t.Logf("Chunk %d may have split a function:\n%s", i, chunk.Content)
			}
		}
	}

	// Reconstruct the code and verify nothing is lost
	var reconstructed string
	lastEndByte := 0
	for _, chunk := range chunks {
		// Handle potential gaps between chunks
		if chunk.StartByte > lastEndByte {
			reconstructed += string([]byte(code)[lastEndByte:chunk.StartByte])
		}
		reconstructed += chunk.Content
		lastEndByte = chunk.EndByte
	}
	// Add any remaining content
	if lastEndByte < len(code) {
		reconstructed += string([]byte(code)[lastEndByte:])
	}

	// The reconstructed code should match the original
	if reconstructed != code && len(chunks) > 0 && chunks[0].StartByte == 0 {
		t.Errorf("reconstructed code doesn't match original\nOriginal length: %d\nReconstructed length: %d",
			len(code), len(reconstructed))
	}
}

func TestIntegration_RealPythonCode(t *testing.T) {
	chunker := NewChunker()

	// Real Python code example
	code := `import asyncio
import aiohttp
from typing import List, Dict, Any

class AsyncWebScraper:
	def __init__(self, max_concurrent: int = 10):
		self.max_concurrent = max_concurrent
		self.session = None
		self.results = []
	
	async def __aenter__(self):
		self.session = aiohttp.ClientSession()
		return self
	
	async def __aexit__(self, exc_type, exc_val, exc_tb):
		await self.session.close()
	
	async def fetch_url(self, url: str) -> Dict[str, Any]:
		try:
			async with self.session.get(url) as response:
				return {
					'url': url,
					'status': response.status,
					'content': await response.text()
				}
		except Exception as e:
			return {
				'url': url,
				'error': str(e)
			}
	
	async def scrape_urls(self, urls: List[str]) -> List[Dict[str, Any]]:
		semaphore = asyncio.Semaphore(self.max_concurrent)
		
		async def fetch_with_semaphore(url):
			async with semaphore:
				return await self.fetch_url(url)
		
		tasks = [fetch_with_semaphore(url) for url in urls]
		self.results = await asyncio.gather(*tasks)
		return self.results

async def main():
	urls = [
		'https://example.com',
		'https://example.org',
		'https://example.net'
	]
	
	async with AsyncWebScraper(max_concurrent=5) as scraper:
		results = await scraper.scrape_urls(urls)
		for result in results:
			if 'error' in result:
				print(f"Error fetching {result['url']}: {result['error']}")
			else:
				print(f"Successfully fetched {result['url']} with status {result['status']}")

if __name__ == "__main__":
	asyncio.run(main())`

	chunks, err := chunker.Chunk(code, WithLanguage(languages.Python), WithMaxSize(40))
	if err != nil {
		t.Fatalf("failed to chunk Python code: %v", err)
	}

	// Verify we got multiple chunks
	if len(chunks) < 3 {
		t.Errorf("expected at least 3 chunks for large Python code, got %d", len(chunks))
	}

	// Verify chunks maintain indentation structure
	for i, chunk := range chunks {
		lines := strings.Split(chunk.Content, "\n")
		for j, line := range lines {
			if line != "" && len(line) > 0 && line[0] != '\t' && line[0] != ' ' {
				// Non-indented lines should typically be at the start of logical blocks
				if j > 0 && strings.TrimSpace(lines[j-1]) != "" {
					// Check if this is a valid Python construct that should start at column 0
					validStarts := []string{"import ", "from ", "class ", "def ", "async ", "if __name__"}
					isValid := false
					for _, start := range validStarts {
						if strings.HasPrefix(line, start) {
							isValid = true
							break
						}
					}
					if !isValid && !strings.HasPrefix(line, "@") {
						t.Logf("Chunk %d may have broken indentation at line %d: %s", i, j, line)
					}
				}
			}
		}
	}
}

func TestIntegration_MultipleLanguages(t *testing.T) {
	chunker := NewChunker()

	langs := []struct {
		name languages.LanguageName
		code string
	}{
		{
			name: languages.JavaScript,
			code: `const express = require('express');

class UserController {
	constructor(userService) {
		this.userService = userService;
	}
	
	async getUsers(req, res) {
		try {
			const users = await this.userService.findAll();
			res.json(users);
		} catch (error) {
			res.status(500).json({ error: error.message });
		}
	}
	
	async createUser(req, res) {
		try {
			const user = await this.userService.create(req.body);
			res.status(201).json(user);
		} catch (error) {
			res.status(400).json({ error: error.message });
		}
	}
}

module.exports = UserController;`,
		},
		{
			name: languages.Java,
			code: `package com.example.demo;

import java.util.List;
import java.util.ArrayList;
import java.util.stream.Collectors;

public class DataProcessor {
	private List<String> data;
	
	public DataProcessor() {
		this.data = new ArrayList<>();
	}
	
	public void addData(String item) {
		if (item != null && !item.isEmpty()) {
			data.add(item);
		}
	}
	
	public List<String> processData() {
		return data.stream()
			.filter(item -> item.length() > 3)
			.map(String::toUpperCase)
			.sorted()
			.collect(Collectors.toList());
	}
	
	public int getDataSize() {
		return data.size();
	}
}`,
		},
	}

	for _, lang := range langs {
		t.Run(string(lang.name), func(t *testing.T) {
			chunks, err := chunker.Chunk(lang.code,
				WithLanguage(lang.name),
				WithMaxSize(25))
			if err != nil {
				t.Fatalf("failed to chunk %s code: %v", lang.name, err)
			}

			if len(chunks) == 0 {
				t.Errorf("no chunks produced for %s code", lang.name)
			}

			// Verify all chunks have the correct language
			for i, chunk := range chunks {
				if chunk.Language != lang.name {
					t.Errorf("chunk %d has wrong language: got %s, want %s",
						i, chunk.Language, lang.name)
				}
			}
		})
	}
}
