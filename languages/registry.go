package languages

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/cue"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/groovy"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/lua"
	tree_sitter_markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/protobuf"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

// LanguageConfig represents a language configuration.
type LanguageConfig struct {
	Name       LanguageName
	Extensions []string
	GetParser  func() *sitter.Language
}

var registry = map[string]LanguageConfig{
	"bash": {
		Name:       Bash,
		Extensions: []string{".sh", ".bash"},
		GetParser:  bash.GetLanguage,
	},
	"c": {
		Name:       C,
		Extensions: []string{".c", ".h"},
		GetParser:  c.GetLanguage,
	},
	"cpp": {
		Name:       CPP,
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".h", ".hh", ".hxx"},
		GetParser:  cpp.GetLanguage,
	},
	"csharp": {
		Name:       CSharp,
		Extensions: []string{".cs"},
		GetParser:  csharp.GetLanguage,
	},
	"css": {
		Name:       CSS,
		Extensions: []string{".css"},
		GetParser:  css.GetLanguage,
	},
	"cue": {
		Name:       Cue,
		Extensions: []string{".cue"},
		GetParser:  cue.GetLanguage,
	},
	"dockerfile": {
		Name:       Dockerfile,
		Extensions: []string{"Dockerfile", ".dockerfile"},
		GetParser:  dockerfile.GetLanguage,
	},
	"elixir": {
		Name:       Elixir,
		Extensions: []string{".ex", ".exs"},
		GetParser:  elixir.GetLanguage,
	},
	"elm": {
		Name:       Elm,
		Extensions: []string{".elm"},
		GetParser:  elm.GetLanguage,
	},
	"go": {
		Name:       Go,
		Extensions: []string{".go"},
		GetParser:  golang.GetLanguage,
	},
	"groovy": {
		Name:       Groovy,
		Extensions: []string{".groovy", ".gradle"},
		GetParser:  groovy.GetLanguage,
	},
	"hcl": {
		Name:       HCL,
		Extensions: []string{".hcl", ".tf"},
		GetParser:  hcl.GetLanguage,
	},
	"html": {
		Name:       HTML,
		Extensions: []string{".html", ".htm"},
		GetParser:  html.GetLanguage,
	},
	"java": {
		Name:       Java,
		Extensions: []string{".java"},
		GetParser:  java.GetLanguage,
	},
	"javascript": {
		Name:       JavaScript,
		Extensions: []string{".js", ".jsx", ".mjs", ".cjs"},
		GetParser:  javascript.GetLanguage,
	},
	"kotlin": {
		Name:       Kotlin,
		Extensions: []string{".kt", ".kts"},
		GetParser:  kotlin.GetLanguage,
	},
	"lua": {
		Name:       Lua,
		Extensions: []string{".lua"},
		GetParser:  lua.GetLanguage,
	},
	"markdown": {
		Name:       Markdown,
		Extensions: []string{".md", ".markdown"},
		GetParser:  tree_sitter_markdown.GetLanguage,
	},
	"ocaml": {
		Name:       OCaml,
		Extensions: []string{".ml", ".mli"},
		GetParser:  ocaml.GetLanguage,
	},
	"php": {
		Name:       PHP,
		Extensions: []string{".php", ".phtml"},
		GetParser:  php.GetLanguage,
	},
	"protobuf": {
		Name:       Protobuf,
		Extensions: []string{".proto"},
		GetParser:  protobuf.GetLanguage,
	},
	"python": {
		Name:       Python,
		Extensions: []string{".py", ".pyi", ".pyw"},
		GetParser:  python.GetLanguage,
	},
	"ruby": {
		Name:       Ruby,
		Extensions: []string{".rb", ".rake", ".gemspec"},
		GetParser:  ruby.GetLanguage,
	},
	"rust": {
		Name:       Rust,
		Extensions: []string{".rs"},
		GetParser:  rust.GetLanguage,
	},
	"scala": {
		Name:       Scala,
		Extensions: []string{".scala", ".sc"},
		GetParser:  scala.GetLanguage,
	},
	"sql": {
		Name:       SQL,
		Extensions: []string{".sql"},
		GetParser:  sql.GetLanguage,
	},
	"svelte": {
		Name:       Svelte,
		Extensions: []string{".svelte"},
		GetParser:  svelte.GetLanguage,
	},
	"swift": {
		Name:       Swift,
		Extensions: []string{".swift"},
		GetParser:  swift.GetLanguage,
	},
	"toml": {
		Name:       TOML,
		Extensions: []string{".toml"},
		GetParser:  toml.GetLanguage,
	},
	"typescript": {
		Name:       TypeScript,
		Extensions: []string{".ts", ".tsx"},
		GetParser:  typescript.GetLanguage,
	},
	"yaml": {
		Name:       YAML,
		Extensions: []string{".yaml", ".yml"},
		GetParser:  yaml.GetLanguage,
	},
	"generic": {
		Name:       Generic,
		Extensions: []string{}, // Generic doesn't have specific extensions
		GetParser:  nil,        // Generic doesn't use tree-sitter
	},
}

// GetLanguageConfig returns the language configuration for the given name.
func GetLanguageConfig(name LanguageName) (LanguageConfig, bool) {
	lang, ok := registry[strings.ToLower(string(name))]
	return lang, ok
}

// DetectLanguage attempts to detect the language from a file path.
// Returns the generic language as a fallback if detection fails.
func DetectLanguage(filepath string) (LanguageConfig, bool) {
	ext := strings.ToLower(filepath)
	if !strings.HasPrefix(ext, ".") {
		if idx := strings.LastIndex(filepath, "."); idx >= 0 {
			ext = strings.ToLower(filepath[idx:])
		}
	}

	// Check for exact filename matches (e.g., "Dockerfile")
	filename := filepath
	if idx := strings.LastIndex(filepath, "/"); idx >= 0 {
		filename = filepath[idx+1:]
	}

	for _, lang := range registry {
		for _, langExt := range lang.Extensions {
			if ext == langExt || filename == langExt {
				return lang, true
			}
		}
	}

	// Fallback to generic language
	return registry["generic"], true
}
