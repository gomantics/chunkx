package chunkx

import "github.com/gomantics/chunkx/languages"

// Chunk represents a semantically coherent unit of code extracted via AST-based chunking.
type Chunk struct {
	Content   string                 // The actual code content
	StartLine int                    // Starting line number (1-based)
	EndLine   int                    // Ending line number (1-based)
	StartByte int                    // Starting byte offset
	EndByte   int                    // Ending byte offset
	NodeTypes []string               // AST node types included in this chunk
	Language  languages.LanguageName // Programming language of the chunk
}
