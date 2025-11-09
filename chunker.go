// Package chunkx provides AST-based code chunking using the CAST algorithm.
//
// ChunkX implements the CAST (Chunking via Abstract Syntax Trees) method for
// semantically-aware code chunking. Unlike line-based chunking, CAST respects
// code structure by parsing source into an AST and creating chunks that align
// with syntactic boundaries (functions, classes, methods).
//
// Basic usage:
//
//	chunker := chunkx.NewChunker()
//	chunks, err := chunker.Chunk(code, chunkx.WithLanguage(languages.Go))
//
// Supports 30+ languages including Bash, C, C++, C#, CSS, Cue, Dockerfile, Elixir,
// Elm, Go, Groovy, HCL, HTML, Java, JavaScript, Kotlin, Lua, Markdown, OCaml, PHP,
// Protobuf, Python, Ruby, Rust, Scala, SQL, Svelte, Swift, TOML, TypeScript, and YAML.
//
// For unsupported file types, the chunker automatically falls back to a generic
// line-based chunking algorithm.
package chunkx

import (
	"fmt"
	"os"
	"strings"

	"github.com/gomantics/chunkx/languages"
	sitter "github.com/smacker/go-tree-sitter"
)

// Default configuration values.
const (
	// DefaultMaxSize is the default maximum chunk size in tokens.
	DefaultMaxSize = 1500

	// DefaultOverlap is the default overlap percentage between chunks.
	DefaultOverlap = 0

	// MaxOverlap is the maximum allowed overlap percentage.
	MaxOverlap = 50
)

// Chunker provides AST-based code chunking capabilities.
type Chunker interface {
	Chunk(code string, opts ...Option) ([]Chunk, error)
	ChunkFile(path string, opts ...Option) ([]Chunk, error)
}

// castChunker implements the CAST algorithm for code chunking.
type castChunker struct {
	parser *Parser
}

// NewChunker creates a new CAST chunker instance.
func NewChunker() Chunker {
	return &castChunker{
		parser: NewParser(),
	}
}

// config holds the configuration for chunking operations.
type config struct {
	maxSize      int
	overlap      float64
	language     languages.LanguageName
	tokenCounter TokenCounter
}

// Option configures the chunker.
type Option func(*config)

// WithMaxSize sets the maximum chunk size in tokens.
func WithMaxSize(tokens int) Option {
	return func(c *config) {
		c.maxSize = tokens
	}
}

// WithOverlap sets the overlap percentage (0-MaxOverlap).
func WithOverlap(percent float64) Option {
	return func(c *config) {
		if percent < 0 {
			percent = 0
		} else if percent > MaxOverlap {
			percent = MaxOverlap
		}
		c.overlap = percent
	}
}

// WithLanguage sets the language for parsing.
// Use the exported constants: languages.Go, languages.Python, etc.
func WithLanguage(lang languages.LanguageName) Option {
	return func(c *config) {
		c.language = lang
	}
}

// WithTokenCounter sets a custom token counter.
func WithTokenCounter(counter TokenCounter) Option {
	return func(c *config) {
		c.tokenCounter = counter
	}
}

// newDefaultConfig creates a new config with default values.
func newDefaultConfig() *config {
	return &config{
		maxSize:      DefaultMaxSize,
		overlap:      DefaultOverlap,
		tokenCounter: &SimpleTokenCounter{},
	}
}

// Chunk splits the code into semantically coherent chunks.
func (c *castChunker) Chunk(code string, opts ...Option) ([]Chunk, error) {
	cfg := newDefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.language == "" {
		return nil, ErrLanguageNotSpecified
	}

	// Use generic chunking for the generic language
	if cfg.language == languages.Generic {
		return c.chunkGeneric(code, cfg)
	}

	parseResult, err := c.parser.Parse(code, cfg.language)
	if err != nil {
		// Fallback to generic chunking if parsing fails
		return c.chunkGeneric(code, cfg)
	}

	root := parseResult.Tree.RootNode()
	chunks, err := c.chunkCode(root, parseResult.Source, cfg)
	if err != nil {
		return nil, err
	}

	// Apply overlap if configured
	if cfg.overlap > 0 {
		chunks = c.applyOverlap(chunks, cfg.overlap)
	}

	return chunks, nil
}

// ChunkFile chunks code from a file.
func (c *castChunker) ChunkFile(path string, opts ...Option) ([]Chunk, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	cfg := newDefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	// Auto-detect language if not specified
	if cfg.language == "" {
		detectedLang, _ := languages.DetectLanguage(path)
		cfg.language = detectedLang.Name

		// Use generic chunking if the language doesn't support AST parsing
		if detectedLang.GetParser == nil {
			return c.chunkGeneric(string(content), cfg)
		}

		parseResult, err := c.parser.ParseFile(path, string(content))
		if err != nil {
			// Fallback to generic chunking if parsing fails
			return c.chunkGeneric(string(content), cfg)
		}

		root := parseResult.Tree.RootNode()
		chunks, err := c.chunkCode(root, parseResult.Source, cfg)
		if err != nil {
			return nil, err
		}

		if cfg.overlap > 0 {
			chunks = c.applyOverlap(chunks, cfg.overlap)
		}

		return chunks, nil
	}

	return c.Chunk(string(content), opts...)
}

// chunkCode implements the main CAST algorithm
func (c *castChunker) chunkCode(node *sitter.Node, source []byte, cfg *config) ([]Chunk, error) {
	size, err := GetNodeSize(node, source, cfg.tokenCounter)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNodeSize, err)
	}

	// If node fits within max size, return it as a single chunk
	if size <= cfg.maxSize {
		return []Chunk{c.nodeToChunk(node, source, cfg.language)}, nil
	}

	// Otherwise, chunk the node's children
	childCount := int(node.ChildCount())
	if childCount == 0 {
		// Leaf node that's too large - return as is (can't split further)
		return []Chunk{c.nodeToChunk(node, source, cfg.language)}, nil
	}

	children := make([]*sitter.Node, 0, childCount)
	for i := 0; i < childCount; i++ {
		if child := node.Child(i); child != nil {
			children = append(children, child)
		}
	}

	return c.chunkNodes(children, source, cfg)
}

// chunkNodes implements the node grouping logic.
func (c *castChunker) chunkNodes(nodes []*sitter.Node, source []byte, cfg *config) ([]Chunk, error) {
	var chunks []Chunk
	var currentNodes []*sitter.Node
	currentSize := 0

	for _, node := range nodes {
		nodeSize, err := GetNodeSize(node, source, cfg.tokenCounter)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrNodeSize, err)
		}

		// If adding this node would exceed max size
		if len(currentNodes) > 0 && currentSize+nodeSize > cfg.maxSize {
			// Save current chunk
			chunks = append(chunks, c.mergeNodesToChunk(currentNodes, source, cfg.language))
			currentNodes = nil
			currentSize = 0
		}

		// If single node exceeds max size, recursively chunk it
		if nodeSize > cfg.maxSize {
			if len(currentNodes) > 0 {
				chunks = append(chunks, c.mergeNodesToChunk(currentNodes, source, cfg.language))
				currentNodes = nil
				currentSize = 0
			}

			subChunks, err := c.chunkCode(node, source, cfg)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, subChunks...)
		} else {
			currentNodes = append(currentNodes, node)
			currentSize += nodeSize
		}
	}

	// Don't forget the last chunk
	if len(currentNodes) > 0 {
		chunks = append(chunks, c.mergeNodesToChunk(currentNodes, source, cfg.language))
	}

	return chunks, nil
}

// nodeToChunk converts a single node to a Chunk.
func (c *castChunker) nodeToChunk(node *sitter.Node, source []byte, language languages.LanguageName) Chunk {
	startLine, endLine := GetLineNumbers(node)

	return Chunk{
		Content:   GetNodeText(node, source),
		StartLine: startLine,
		EndLine:   endLine,
		StartByte: int(node.StartByte()),
		EndByte:   int(node.EndByte()),
		NodeTypes: []string{node.Type()},
		Language:  language,
	}
}

// mergeNodesToChunk merges multiple nodes into a single chunk.
func (c *castChunker) mergeNodesToChunk(nodes []*sitter.Node, source []byte, language languages.LanguageName) Chunk {
	if len(nodes) == 0 {
		return Chunk{Language: language}
	}

	// Find the span of all nodes
	firstNode := nodes[0]
	lastNode := nodes[len(nodes)-1]

	startByte := firstNode.StartByte()
	endByte := lastNode.EndByte()

	// Collect all node types
	nodeTypes := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeTypes = append(nodeTypes, node.Type())
	}

	startLine, _ := GetLineNumbers(firstNode)
	_, endLine := GetLineNumbers(lastNode)

	return Chunk{
		Content:   string(source[startByte:endByte]),
		StartLine: startLine,
		EndLine:   endLine,
		StartByte: int(startByte),
		EndByte:   int(endByte),
		NodeTypes: nodeTypes,
		Language:  language,
	}
}

// applyOverlap adds overlap between consecutive chunks.
func (c *castChunker) applyOverlap(chunks []Chunk, overlapPercent float64) []Chunk {
	if len(chunks) <= 1 || overlapPercent <= 0 {
		return chunks
	}

	overlappedChunks := make([]Chunk, 0, len(chunks))

	for i := range chunks {
		chunk := chunks[i]

		// Calculate overlap size
		overlapSize := int(float64(len(chunk.Content)) * (overlapPercent / 100.0))

		// Add content from previous chunk if available
		if i > 0 && overlapSize > 0 {
			prevChunk := chunks[i-1]
			prevContent := prevChunk.Content

			// Get the last N characters from previous chunk
			startIdx := max(len(prevContent)-overlapSize, 0)

			chunk.Content = prevContent[startIdx:] + "\n" + chunk.Content
			// Adjust start position to reflect the overlap
			chunk.StartByte -= (len(prevContent) - startIdx)
			chunk.StartLine = prevChunk.StartLine + countLines(prevContent[:startIdx])
		}

		// Add content from next chunk if available
		if i < len(chunks)-1 && overlapSize > 0 {
			nextChunk := chunks[i+1]
			nextContent := nextChunk.Content

			// Get the first N characters from next chunk
			endIdx := min(overlapSize, len(nextContent))

			chunk.Content = chunk.Content + "\n" + nextContent[:endIdx]
			// Adjust end position to reflect the overlap
			chunk.EndByte += endIdx
			chunk.EndLine = nextChunk.StartLine + countLines(nextContent[:endIdx]) - 1
		}

		overlappedChunks = append(overlappedChunks, chunk)
	}

	return overlappedChunks
}

// countLines counts the number of lines in a string.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	count := 1
	for _, r := range s {
		if r == '\n' {
			count++
		}
	}
	return count
}

// chunkGeneric implements a simple line-based chunking algorithm for unsupported languages.
// This is used as a fallback when tree-sitter parsing is not available.
func (c *castChunker) chunkGeneric(code string, cfg *config) ([]Chunk, error) {
	lines := strings.Split(code, "\n")
	var chunks []Chunk
	var currentLines []string
	currentSize := 0
	currentStartLine := 1

	for i, line := range lines {
		lineSize, err := cfg.tokenCounter.CountTokens(line)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrNodeSize, err)
		}

		// If adding this line would exceed max size and we have content
		if len(currentLines) > 0 && currentSize+lineSize > cfg.maxSize {
			// Save current chunk
			chunk := Chunk{
				Content:   strings.Join(currentLines, "\n"),
				StartLine: currentStartLine,
				EndLine:   currentStartLine + len(currentLines) - 1,
				StartByte: 0, // Generic chunks don't track byte positions accurately
				EndByte:   0,
				NodeTypes: []string{"generic"},
				Language:  cfg.language,
			}
			chunks = append(chunks, chunk)
			currentLines = nil
			currentSize = 0
			currentStartLine = i + 1
		}

		currentLines = append(currentLines, line)
		currentSize += lineSize
	}

	// Don't forget the last chunk
	if len(currentLines) > 0 {
		chunk := Chunk{
			Content:   strings.Join(currentLines, "\n"),
			StartLine: currentStartLine,
			EndLine:   currentStartLine + len(currentLines) - 1,
			StartByte: 0,
			EndByte:   0,
			NodeTypes: []string{"generic"},
			Language:  cfg.language,
		}
		chunks = append(chunks, chunk)
	}

	// Apply overlap if configured
	if cfg.overlap > 0 {
		chunks = c.applyOverlap(chunks, cfg.overlap)
	}

	return chunks, nil
}
