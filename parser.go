package chunkx

import (
	"context"
	"fmt"

	"github.com/gomantics/chunkx/languages"
	sitter "github.com/smacker/go-tree-sitter"
)

// Parser provides language-agnostic parsing capabilities using tree-sitter.
type Parser struct {
	parser *sitter.Parser
}

// NewParser creates a new parser instance.
func NewParser() *Parser {
	return &Parser{
		parser: sitter.NewParser(),
	}
}

// ParseResult contains the parsed AST and metadata.
type ParseResult struct {
	Tree     *sitter.Tree
	Language languages.LanguageName
	Source   []byte
}

// Parse parses the given code using the specified language.
func (p *Parser) Parse(code string, language languages.LanguageName) (*ParseResult, error) {
	lang, ok := languages.GetLanguageConfig(language)
	if !ok {
		return nil, &LanguageError{
			Language: language,
			Err:      ErrUnsupportedLanguage,
		}
	}

	// Generic language doesn't support tree-sitter parsing
	if lang.GetParser == nil {
		return nil, &LanguageError{
			Language: language,
			Err:      ErrNoASTSupport,
		}
	}

	p.parser.SetLanguage(lang.GetParser())

	sourceCode := []byte(code)
	tree, err := p.parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, &LanguageError{
			Language: language,
			Err:      fmt.Errorf("%w: %w", ErrParseFailed, err),
		}
	}

	return &ParseResult{
		Tree:     tree,
		Language: lang.Name,
		Source:   sourceCode,
	}, nil
}

// ParseFile parses code from a file, auto-detecting the language.
func (p *Parser) ParseFile(filepath string, code string) (*ParseResult, error) {
	lang, ok := languages.DetectLanguage(filepath)
	if !ok {
		return nil, fmt.Errorf("cannot detect language for file: %s", filepath)
	}

	return p.Parse(code, lang.Name)
}

// GetNodeText returns the text content of a node.
func GetNodeText(node *sitter.Node, source []byte) string {
	return string(source[node.StartByte():node.EndByte()])
}

// GetNodeSize calculates the size of a node using the provided token counter.
func GetNodeSize(node *sitter.Node, source []byte, counter TokenCounter) (int, error) {
	text := GetNodeText(node, source)
	return counter.CountTokens(text)
}

// GetLineNumbers returns the start and end line numbers for a node (1-based).
func GetLineNumbers(node *sitter.Node) (int, int) {
	return int(node.StartPoint().Row) + 1, int(node.EndPoint().Row) + 1
}
