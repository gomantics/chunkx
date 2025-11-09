package languages

// LanguageName represents a supported programming language.
type LanguageName string

// Supported language constants.
const (
	Bash       LanguageName = "bash"
	C          LanguageName = "c"
	CPP        LanguageName = "cpp"
	CSharp     LanguageName = "csharp"
	CSS        LanguageName = "css"
	Cue        LanguageName = "cue"
	Dockerfile LanguageName = "dockerfile"
	Elixir     LanguageName = "elixir"
	Elm        LanguageName = "elm"
	Go         LanguageName = "go"
	Groovy     LanguageName = "groovy"
	HCL        LanguageName = "hcl"
	HTML       LanguageName = "html"
	Java       LanguageName = "java"
	JavaScript LanguageName = "javascript"
	Kotlin     LanguageName = "kotlin"
	Lua        LanguageName = "lua"
	Markdown   LanguageName = "markdown"
	OCaml      LanguageName = "ocaml"
	PHP        LanguageName = "php"
	Protobuf   LanguageName = "protobuf"
	Python     LanguageName = "python"
	Ruby       LanguageName = "ruby"
	Rust       LanguageName = "rust"
	Scala      LanguageName = "scala"
	SQL        LanguageName = "sql"
	Svelte     LanguageName = "svelte"
	Swift      LanguageName = "swift"
	TOML       LanguageName = "toml"
	TypeScript LanguageName = "typescript"
	YAML       LanguageName = "yaml"
	Generic    LanguageName = "generic" // Fallback for unsupported languages
)

// String returns the string representation of the language.
func (l LanguageName) String() string {
	return string(l)
}
