# chunkx

A Go library for AST-based code chunking implementing the CAST (Chunking via Abstract Syntax Trees) algorithm from the paper ["cAST: Enhancing Code Retrieval-Augmented Generation with Structural Chunking via Abstract Syntax Tree"](https://arxiv.org/pdf/2506.15655).

## Features

- **Syntax-aware chunking**: Respects code structure (functions, classes, methods) instead of arbitrarily splitting at line boundaries
- **Multi-language support**: Works with 30+ languages via tree-sitter parsers
- **Generic fallback**: Automatically falls back to line-based chunking for unsupported file types
- **Configurable chunk sizes**: Set maximum chunk size in tokens, bytes, or lines
- **Custom token counters**: Pluggable interface for custom tokenization strategies
- **Overlap support**: Optional chunk overlapping for better context preservation

## Installation

```bash
go get github.com/gomantics/chunkx
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/gomantics/chunkx"
    "github.com/gomantics/chunkx/languages"
)

func main() {
    chunker := chunkx.NewChunker()

    code := `package main

func hello() {
    fmt.Println("Hello, World!")
}

func goodbye() {
    fmt.Println("Goodbye!")
}`

    chunks, err := chunker.Chunk(code,
        chunkx.WithLanguage(languages.Go),
        chunkx.WithMaxSize(50))
    if err != nil {
        panic(err)
    }

    for i, chunk := range chunks {
        fmt.Printf("Chunk %d (lines %d-%d):\n%s\n\n",
            i+1, chunk.StartLine, chunk.EndLine, chunk.Content)
    }
}
```

## Usage

### Basic Chunking

```go
chunker := chunkx.NewChunker()

// Chunk code with language specified
chunks, err := chunker.Chunk(code, chunkx.WithLanguage(languages.Python))
```

### File-based Chunking

```go
// Auto-detects language from file extension
chunks, err := chunker.ChunkFile("main.go")
```

### Custom Configuration

```go
chunks, err := chunker.Chunk(code,
    chunkx.WithLanguage(languages.Go),
    chunkx.WithMaxSize(1500),      // Max 1500 tokens per chunk
    chunkx.WithOverlap(10),         // 10% overlap between chunks
)
```

### Custom Token Counter

```go
type MyTokenCounter struct{}

func (m *MyTokenCounter) CountTokens(text string) (int, error) {
    // Your custom tokenization logic
    return len(strings.Fields(text)), nil
}

chunks, err := chunker.Chunk(code,
    chunkx.WithLanguage(languages.Go),
    chunkx.WithTokenCounter(&MyTokenCounter{}))
```

### OpenAI-Compatible Token Counting

For production use with OpenAI models, you can integrate [tiktoken-go](https://github.com/pkoukk/tiktoken-go) for accurate token counting:

```go
import (
    "github.com/pkoukk/tiktoken-go"

    "github.com/gomantics/chunkx"
    "github.com/gomantics/chunkx/languages"
)

// TikTokenCounter uses OpenAI's tiktoken for accurate token counting
type TikTokenCounter struct {
    encoding *tiktoken.Tiktoken
}

// NewTikTokenCounter creates a counter for a specific OpenAI model
func NewTikTokenCounter(model string) (*TikTokenCounter, error) {
    encoding, err := tiktoken.EncodingForModel(model)
    if err != nil {
        return nil, err
    }
    return &TikTokenCounter{encoding: encoding}, nil
}

func (t *TikTokenCounter) CountTokens(text string) (int, error) {
    tokens := t.encoding.Encode(text, nil, nil)
    return len(tokens), nil
}

// Usage example
func main() {
    tokenCounter, err := NewTikTokenCounter("gpt-4")
    if err != nil {
        panic(err)
    }

    chunker := chunkx.NewChunker()
    chunks, err := chunker.Chunk(code,
        chunkx.WithLanguage(languages.Python),
        chunkx.WithMaxSize(8000),
        chunkx.WithTokenCounter(tokenCounter))
    if err != nil {
        panic(err)
    }

    // Your chunks are now sized according to GPT-4's tokenization
    for _, chunk := range chunks {
        // Process chunks...
    }
}
```

This ensures your chunks respect the exact token limits of OpenAI models like GPT-3.5, GPT-4, and GPT-4o.

### Built-in Token Counters

- `SimpleTokenCounter`: Whitespace-based word counting (default)
- `ByteCounter`: Counts bytes instead of tokens
- `LineCounter`: Counts lines instead of tokens

```go
// Use byte-based chunking
chunks, err := chunker.Chunk(code,
    chunkx.WithLanguage(languages.Python),
    chunkx.WithMaxSize(4096),
    chunkx.WithTokenCounter(&chunkx.ByteCounter{}))
```

## Supported Languages

ChunkX supports 30+ programming languages via tree-sitter. Use the exported language constants from the `languages` package (e.g., `languages.Go`, `languages.Python`, `languages.JavaScript`, etc.). See the [languages package](languages/registry.go) for the complete list of supported languages and file extensions.

For files with unrecognized extensions or explicitly using `languages.Generic`, ChunkX automatically falls back to a line-based chunking algorithm that maintains the chunking semantics without requiring AST parsing.

## How It Works

ChunkX implements the CAST algorithm which:

1. Parses source code into an Abstract Syntax Tree (AST)
2. Recursively traverses the AST to identify semantic units
3. Groups nodes while respecting the maximum chunk size
4. Splits large nodes that exceed the size limit
5. Merges smaller sibling nodes to maximize chunk density

This approach ensures that chunks:

- Preserve syntactic integrity (no mid-function splits)
- Maintain semantic coherence
- Are self-contained and meaningful
- Respect language-specific structures

## Chunk Structure

```go
type Chunk struct {
    Content    string                // The actual code content
    StartLine  int                   // Starting line number (1-based)
    EndLine    int                   // Ending line number (1-based)
    StartByte  int                   // Starting byte offset
    EndByte    int                   // Ending byte offset
    NodeTypes  []string              // AST node types included
    Language   languages.LanguageName // Programming language
}
```

## Performance

Benchmarks on Apple M4 Max (3s run):

```
BenchmarkASTChunking-14                              41301     85932 ns/op   19520 B/op     170 allocs/op
BenchmarkLineBasedChunking-14                      4392780       831.6 ns/op   1904 B/op      10 allocs/op
BenchmarkASTChunkingLarge-14                         4681    769800 ns/op  110464 B/op     794 allocs/op
BenchmarkLineBasedChunkingLarge-14                 437184      8273 ns/op   16880 B/op      27 allocs/op
BenchmarkASTChunkingMultipleLanguages-14            22951    156257 ns/op   42336 B/op     336 allocs/op
BenchmarkTokenCounters/SimpleTokenCounter-14        51332     70434 ns/op    4760 B/op      20 allocs/op
BenchmarkTokenCounters/ByteCounter-14               40485     88952 ns/op   21504 B/op     227 allocs/op
BenchmarkTokenCounters/LineCounter-14               51607     70349 ns/op    3224 B/op      19 allocs/op
BenchmarkOverlapChunking/Overlap0-14                42333     85163 ns/op   19544 B/op     172 allocs/op
BenchmarkOverlapChunking/Overlap10-14               41676     85761 ns/op   21832 B/op     187 allocs/op
BenchmarkOverlapChunking/Overlap25-14               42122     85715 ns/op   22032 B/op     187 allocs/op
BenchmarkOverlapChunking/Overlap50-14               41696     85976 ns/op   22360 B/op     187 allocs/op
```

AST-based chunking is ~100x slower than naive line-based chunking but produces semantically superior chunks that improve RAG performance. The SimpleTokenCounter and LineCounter provide the best performance, while ByteCounter has slightly higher overhead due to more allocations. Chunk overlap has minimal performance impact (~0.5% overhead).

## Examples

The `testdata/` directory contains real-world code examples in multiple languages, along with their chunked outputs in JSON format. These examples serve as both documentation and regression tests:

- **`testdata/sources/`**: Example source files in Go, Python, JavaScript, TypeScript, Java, Rust, and C++
- **`testdata/*.approved.json`**: Snapshot test outputs showing how each example is chunked

To see how chunkx handles different languages and chunk sizes, browse the approved JSON files. They show:

- Complete chunk content
- Line and byte ranges
- AST node types included in each chunk
- How semantic boundaries are preserved

The snapshots are automatically verified using [go-approval-tests](https://github.com/approvals/go-approval-tests) to ensure chunking behavior remains consistent across changes.

## Testing

```bash
# Run tests
go test ./...

# Run benchmarks
go test -bench=. -benchtime=10s

# Run with coverage
go test -cover ./...

# Run approval tests (regenerate snapshots on first failure)
go test -run TestChunkingExamples
```

## Use Cases

- **RAG Systems**: Improve retrieval quality by providing semantically coherent code chunks
- **Code Search**: Index code at meaningful boundaries
- **Documentation**: Generate documentation from logical code units
- **Code Analysis**: Process code in structured segments
- **LLM Context Windows**: Fit code into token limits while preserving structure

## Design Principles

1. **Minimalist**: Clean, focused codebase with no unnecessary abstractions
2. **Well-tested**: Comprehensive unit, integration, and benchmark tests
3. **Pluggable**: Interface-based design for extensibility
4. **Language-agnostic**: Works consistently across programming languages

## References

- [cAST Paper (arXiv:2506.15655)](https://arxiv.org/pdf/2506.15655)
- [Tree-sitter](https://tree-sitter.github.io/tree-sitter/)
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter)

## License

[MIT](./LICENSE)
