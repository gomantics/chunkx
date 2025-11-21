# chunkx

[![Go Reference](https://pkg.go.dev/badge/github.com/gomantics/chunkx.svg)](https://pkg.go.dev/github.com/gomantics/chunkx)
[![CI](https://github.com/gomantics/chunkx/actions/workflows/ci.yml/badge.svg)](https://github.com/gomantics/chunkx/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomantics/chunkx)](https://goreportcard.com/report/github.com/gomantics/chunkx)

A Go library for AST-based code chunking implementing the CAST (Chunking via Abstract Syntax Trees) algorithm.

## Installation

```bash
go get github.com/gomantics/chunkx
```

## Documentation

For complete documentation, usage examples, and API reference, visit:

**https://gomantics.dev/chunkx**

## Features

- Syntax-aware chunking that respects code structure
- Support for 30+ programming languages via tree-sitter
- Configurable chunk sizes (tokens, bytes, or lines)
- Custom token counters (including OpenAI tiktoken)
- Optional chunk overlapping for better context

## Quick Example

```go
package main

import (
    "github.com/gomantics/chunkx"
    "github.com/gomantics/chunkx/languages"
)

func main() {
    chunker := chunkx.NewChunker()
    chunks, _ := chunker.Chunk(code,
        chunkx.WithLanguage(languages.Go),
        chunkx.WithMaxSize(1500))
}
```

## License

[MIT](./LICENSE)
