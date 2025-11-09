package chunkx

import (
	"testing"

	"github.com/gomantics/chunkx/languages"
)

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		code     string
		language languages.LanguageName
		wantErr  bool
	}{
		{
			name:     "simple Go function",
			code:     `func main() { fmt.Println("Hello") }`,
			language: languages.Go,
			wantErr:  false,
		},
		{
			name:     "simple Python function",
			code:     `def hello(): print("Hello")`,
			language: languages.Python,
			wantErr:  false,
		},
		{
			name:     "simple JavaScript function",
			code:     `function hello() { console.log("Hello"); }`,
			language: languages.JavaScript,
			wantErr:  false,
		},
		{
			name:     "empty code",
			code:     "",
			language: languages.Go,
			wantErr:  false,
		},
		{
			name:     "unsupported language",
			code:     `some code`,
			language: "unsupported",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.code, tt.language)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Parse() returned nil result")
			}
			if !tt.wantErr && result.Tree == nil {
				t.Error("Parse() returned nil tree")
			}
			if result != nil && result.Tree != nil {
				// Tree cleanup not needed
			}
		})
	}
}

func TestParser_ParseFile(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		filepath string
		code     string
		wantErr  bool
	}{
		{
			name:     "Go file",
			filepath: "main.go",
			code:     `package main`,
			wantErr:  false,
		},
		{
			name:     "Python file",
			filepath: "script.py",
			code:     `print("hello")`,
			wantErr:  false,
		},
		{
			name:     "JavaScript file",
			filepath: "app.js",
			code:     `console.log("hello");`,
			wantErr:  false,
		},
		{
			name:     "TypeScript file",
			filepath: "app.ts",
			code:     `const x: string = "hello";`,
			wantErr:  false,
		},
		{
			name:     "Java file",
			filepath: "Main.java",
			code:     `public class Main {}`,
			wantErr:  false,
		},
		{
			name:     "unknown extension",
			filepath: "file.xyz",
			code:     `some content`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseFile(tt.filepath, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseFile() returned nil result")
			}
			if result != nil && result.Tree != nil {
				// Tree cleanup not needed
			}
		})
	}
}

func TestGetNodeText(t *testing.T) {
	parser := NewParser()
	code := `func hello() { return "world" }`

	result, err := parser.Parse(code, languages.Go)
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}
	// Tree cleanup not needed

	root := result.Tree.RootNode()
	text := GetNodeText(root, result.Source)

	if text != code {
		t.Errorf("GetNodeText() = %q, want %q", text, code)
	}
}

func TestGetNodeSize(t *testing.T) {
	parser := NewParser()
	code := `func hello() { return "world" }`

	result, err := parser.Parse(code, languages.Go)
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}
	// Tree cleanup not needed

	root := result.Tree.RootNode()
	counter := &SimpleTokenCounter{}

	size, err := GetNodeSize(root, result.Source, counter)
	if err != nil {
		t.Fatalf("GetNodeSize() error = %v", err)
	}

	// SimpleTokenCounter counts whitespace-separated words
	// The code: func hello() { return "world" }
	// Tokens: func, hello(), {, return, "world", }
	expectedSize := 6 // SimpleTokenCounter counts whitespace-separated words
	if size != expectedSize {
		t.Errorf("GetNodeSize() = %d, want %d", size, expectedSize)
	}
}

func TestGetLineNumbers(t *testing.T) {
	parser := NewParser()
	code := `func hello() {
	return "world"
}`

	result, err := parser.Parse(code, languages.Go)
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}
	// Tree cleanup not needed

	root := result.Tree.RootNode()
	startLine, endLine := GetLineNumbers(root)

	if startLine != 1 {
		t.Errorf("GetLineNumbers() start = %d, want 1", startLine)
	}
	if endLine != 3 {
		t.Errorf("GetLineNumbers() end = %d, want 3", endLine)
	}
}
