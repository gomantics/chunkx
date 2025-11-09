package chunkx

import (
	"errors"
	"fmt"

	"github.com/gomantics/chunkx/languages"
)

// Sentinel errors that can be checked with errors.Is().
var (
	// ErrLanguageNotSpecified is returned when no language is specified for chunking.
	ErrLanguageNotSpecified = errors.New("language must be specified")

	// ErrUnsupportedLanguage is returned when the specified language is not supported.
	ErrUnsupportedLanguage = errors.New("unsupported language")

	// ErrNoASTSupport is returned when a language doesn't support AST parsing.
	ErrNoASTSupport = errors.New("language does not support AST parsing")

	// ErrParseFailed is returned when parsing fails.
	ErrParseFailed = errors.New("failed to parse code")

	// ErrNodeSize is returned when node size calculation fails.
	ErrNodeSize = errors.New("failed to calculate node size")
)

// LanguageError wraps language-specific errors with the language name.
type LanguageError struct {
	Language languages.LanguageName
	Err      error
}

func (e *LanguageError) Error() string {
	return fmt.Sprintf("%s: %v", e.Language, e.Err)
}

func (e *LanguageError) Unwrap() error {
	return e.Err
}
