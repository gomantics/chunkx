package chunkx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gomantics/chunkx/languages"
)

// lineBasedChunker implements a simple line-based chunking algorithm for comparison
type lineBasedChunker struct {
	maxLines int
}

func (l *lineBasedChunker) chunk(code string) []string {
	lines := strings.Split(code, "\n")
	var chunks []string

	for i := 0; i < len(lines); i += l.maxLines {
		end := min(i+l.maxLines, len(lines))
		chunk := strings.Join(lines[i:end], "\n")
		chunks = append(chunks, chunk)
	}

	return chunks
}

var testCode = `package main

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
}

func helper1() {
	fmt.Println("Helper 1")
}

func helper2() {
	fmt.Println("Helper 2")
}`

// BenchmarkASTChunking measures the performance of AST-based chunking
func BenchmarkASTChunking(b *testing.B) {
	chunker := NewChunker()

	for b.Loop() {
		_, err := chunker.Chunk(testCode, WithLanguage(languages.Go), WithMaxSize(20))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLineBasedChunking measures the performance of line-based chunking
func BenchmarkLineBasedChunking(b *testing.B) {
	lineChunker := &lineBasedChunker{maxLines: 10}

	for b.Loop() {
		_ = lineChunker.chunk(testCode)
	}
}

// BenchmarkASTChunkingLarge tests with larger code
func BenchmarkASTChunkingLarge(b *testing.B) {
	// Create a large code sample by repeating the test code
	largeCode := strings.Repeat(testCode+"\n\n", 10)
	chunker := NewChunker()

	for b.Loop() {
		_, err := chunker.Chunk(largeCode, WithLanguage(languages.Go), WithMaxSize(50))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLineBasedChunkingLarge tests with larger code
func BenchmarkLineBasedChunkingLarge(b *testing.B) {
	largeCode := strings.Repeat(testCode+"\n\n", 10)
	lineChunker := &lineBasedChunker{maxLines: 25}

	for b.Loop() {
		_ = lineChunker.chunk(largeCode)
	}
}

// BenchmarkASTChunkingMultipleLanguages tests performance across languages
func BenchmarkASTChunkingMultipleLanguages(b *testing.B) {
	langs := []struct {
		name languages.LanguageName
		code string
	}{
		{
			name: languages.Go,
			code: testCode,
		},
		{
			name: languages.Python,
			code: `class Example:
	def __init__(self):
		self.value = 0
	
	def method1(self):
		for i in range(10):
			self.value += i
		return self.value * 2
	
	def method2(self):
		result = []
		for i in range(self.value):
			result.append(i * i)
		return result`,
		},
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
}

module.exports = UserController;`,
		},
	}

	chunker := NewChunker()

	for b.Loop() {
		for _, lang := range langs {
			_, err := chunker.Chunk(lang.code, WithLanguage(lang.name), WithMaxSize(20))
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkTokenCounters compares different token counting strategies
func BenchmarkTokenCounters(b *testing.B) {
	counters := []struct {
		name    string
		counter TokenCounter
	}{
		{"SimpleTokenCounter", &SimpleTokenCounter{}},
		{"ByteCounter", &ByteCounter{}},
		{"LineCounter", &LineCounter{}},
	}

	chunker := NewChunker()

	for _, tc := range counters {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_, err := chunker.Chunk(testCode,
					WithLanguage(languages.Go),
					WithMaxSize(100),
					WithTokenCounter(tc.counter))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkOverlapChunking tests the performance impact of overlap
func BenchmarkOverlapChunking(b *testing.B) {
	chunker := NewChunker()
	overlapPercentages := []float64{0, 10, 25, 50}

	for _, overlap := range overlapPercentages {
		b.Run(fmt.Sprintf("Overlap%.0f", overlap), func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_, err := chunker.Chunk(testCode,
					WithLanguage(languages.Go),
					WithMaxSize(20),
					WithOverlap(overlap))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
