package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
)

//nolint:staticcheck
func TestNewWfpScanner(t *testing.T) {
	cfg := &config.ScanConfig{
		ThreadNum: "16",
		LogLevel:  "info",
	}

	scanner := NewWfpScanner(cfg)
	if scanner == nil {
		t.Error("NewWfpScanner should not return nil")
	}

	if scanner.config != cfg {
		t.Error("Scanner should store the provided config")
	}
}

func TestWfpScanner_GenerateWfpFile(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()

	// Create some test files
	testFiles := map[string]string{
		"main.go":        "package main\n\nfunc main() {\n\tprintln(\"Hello World\")\n}",
		"utils.go":       "package main\n\nfunc utility() string {\n\treturn \"utility\"\n}",
		"README.md":      "# Test Project\n\nThis is a test project.",
		"subdir/test.go": "package subdir\n\nfunc test() bool {\n\treturn true\n}",
	}

	for fileName, content := range testFiles {
		fullPath := filepath.Join(tempDir, fileName)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", fileName, err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}

	cfg := &config.ScanConfig{
		ThreadNum: "4",
		ToPath:    tempDir,
	}

	scanner := NewWfpScanner(cfg)
	wfpFile, err := scanner.GenerateWfpFile(tempDir)
	if err != nil {
		t.Fatalf("GenerateWfpFile failed: %v", err)
	}

	// Check if the WFP file was created
	if _, err := os.Stat(wfpFile); os.IsNotExist(err) {
		t.Errorf("WFP file was not created: %s", wfpFile)
	}

	// Read the WFP file and check its content
	content, err := os.ReadFile(wfpFile)
	if err != nil {
		t.Fatalf("Failed to read WFP file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("WFP file should not be empty")
	}

	// Check if the WFP file contains fingerprints for our test files
	// The format should be: file=<path>,<line_count>,<hash_list>
	lines := strings.Split(contentStr, "\n")
	foundGoFiles := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "file=") && strings.Contains(line, ".go") {
			foundGoFiles++
		}
	}

	if foundGoFiles < 3 { // We have 3 .go files
		t.Errorf("Expected at least 3 Go files in WFP, found %d", foundGoFiles)
	}

	// Clean up
	_ = os.Remove(wfpFile)
}

func TestWfpScanner_GenerateWfpFile_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.ScanConfig{
		ThreadNum: "4",
		ToPath:    tempDir,
	}

	scanner := NewWfpScanner(cfg)
	wfpFile, err := scanner.GenerateWfpFile(tempDir)
	if err != nil {
		t.Fatalf("GenerateWfpFile failed: %v", err)
	}

	// Check if the WFP file was created (should be empty or minimal)
	if _, err := os.Stat(wfpFile); os.IsNotExist(err) {
		t.Errorf("WFP file was not created: %s", wfpFile)
	}

	// Clean up
	_ = os.Remove(wfpFile)
}

func TestWfpScanner_GenerateWfpFile_NonExistentDirectory(t *testing.T) {
	cfg := &config.ScanConfig{
		ThreadNum: "4",
		ToPath:    "/tmp",
	}

	scanner := NewWfpScanner(cfg)
	_, err := scanner.GenerateWfpFile("/non/existent/directory")
	if err == nil {
		t.Error("GenerateWfpFile should return error for non-existent directory")
	}
}

func TestWfpScanner_shouldIncludeFile(t *testing.T) {
	scanner := &WfpScanner{}

	tests := []struct {
		fileName string
		expected bool
	}{
		// Source code files - should be included
		{"main.go", true},
		{"test.java", true},
		{"script.py", true},
		{"app.js", true},
		{"style.css", true},
		{"component.tsx", true},
		{"module.rs", true},
		{"main.cpp", true},
		{"header.h", true},
		{"source.c", true},

		// Binary files - should be excluded
		{"app.exe", false},
		{"library.dll", false},
		{"program.bin", false},
		{"archive.zip", false},
		{"image.jpg", false},
		{"photo.png", false},
		{"document.pdf", false},
		{"video.mp4", false},

		// Build/dependency files - should be excluded
		{"target/classes/App.class", false},
		{"node_modules/lib/index.js", false},
		{".git/config", false},
		{"build/output.jar", false},

		// Configuration files - should be included
		{"config.json", true},
		{"settings.xml", true},
		{"package.json", true},
		{"Dockerfile", true},
		{"Makefile", true},
		{"pom.xml", true},
		{"requirements.txt", true},
		{"go.mod", true},

		// Documentation - should be included
		{"README.md", true},
		{"LICENSE", true},
		{"CHANGELOG.txt", true},

		// Hidden files - should be excluded
		{".hidden", false},
		{".DS_Store", false},
		{".gitignore", true}, // Exception for important dotfiles
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			result := scanner.shouldIncludeFile(tt.fileName)
			if result != tt.expected {
				t.Errorf("shouldIncludeFile(%s) = %v, want %v", tt.fileName, result, tt.expected)
			}
		})
	}
}

func TestWfpScanner_calculateFileHash(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is test content\nwith multiple lines\nfor hash calculation"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scanner := &WfpScanner{}
	hash, err := scanner.calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("calculateFileHash failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	// Hash should be consistent
	hash2, err := scanner.calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("Second calculateFileHash failed: %v", err)
	}

	if hash != hash2 {
		t.Error("Hash should be consistent across calls")
	}

	// Test with different content
	differentFile := filepath.Join(tempDir, "different.txt")
	err = os.WriteFile(differentFile, []byte("Different content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create different file: %v", err)
	}

	differentHash, err := scanner.calculateFileHash(differentFile)
	if err != nil {
		t.Fatalf("calculateFileHash for different file failed: %v", err)
	}

	if hash == differentHash {
		t.Error("Different files should produce different hashes")
	}
}

func TestWfpScanner_countLines(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"Empty file", "", 0},
		{"Single line", "Hello", 1},
		{"Multiple lines", "Line 1\nLine 2\nLine 3", 3},
		{"Lines with empty lines", "Line 1\n\nLine 3\n", 4},
		{"Trailing newline", "Line 1\nLine 2\n", 2},
	}

	scanner := &WfpScanner{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			count, err := scanner.countLines(testFile)
			if err != nil {
				t.Fatalf("countLines failed: %v", err)
			}

			if count != tt.expected {
				t.Errorf("countLines() = %d, want %d", count, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkWfpScanner_calculateFileHash(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create a 1KB test file
	content := strings.Repeat("This is a line of text for benchmarking.\n", 25)
	_ = os.WriteFile(testFile, []byte(content), 0644)

	scanner := &WfpScanner{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.calculateFileHash(testFile)
	}
}

func BenchmarkWfpScanner_shouldIncludeFile(b *testing.B) {
	scanner := &WfpScanner{}
	testFiles := []string{
		"main.go", "test.java", "script.py", "app.js",
		"library.dll", "archive.zip", "image.jpg",
		"config.json", "README.md", ".hidden",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, file := range testFiles {
			scanner.shouldIncludeFile(file)
		}
	}
}
