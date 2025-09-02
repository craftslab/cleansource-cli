package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")

	// File doesn't exist yet
	if FileExists(tempFile) {
		t.Error("FileExists should return false for non-existent file")
	}

	// Create the file
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File should exist now
	if !FileExists(tempFile) {
		t.Error("FileExists should return true for existing file")
	}

	// Test with directory
	if !FileExists(tempDir) {
		t.Error("FileExists should return true for existing directory")
	}
}

func TestIsDirEmpty(t *testing.T) {
	tempDir := t.TempDir()

	// Empty directory
	if !IsDirEmpty(tempDir) {
		t.Error("IsDirEmpty should return true for empty directory")
	}

	// Add a file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Directory is no longer empty
	if IsDirEmpty(tempDir) {
		t.Error("IsDirEmpty should return false for non-empty directory")
	}

	// Test with non-existent directory
	if !IsDirEmpty("/non/existent/path") {
		t.Error("IsDirEmpty should return true for non-existent directory")
	}
}

func TestCreateZipArchive(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	outputDir := filepath.Join(tempDir, "output")

	// Create source directory and files
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Add some test files
	testFiles := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range testFiles {
		fullPath := filepath.Join(sourceDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create zip archive
	zipFile, err := CreateZipArchive(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("CreateZipArchive failed: %v", err)
	}

	// Check if zip file was created
	if !FileExists(zipFile) {
		t.Error("Zip file was not created")
	}

	// Check file size
	info, err := os.Stat(zipFile)
	if err != nil {
		t.Fatalf("Failed to stat zip file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Zip file is empty")
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_file.txt", "normal_file.txt"},
		{"file with spaces.txt", "file_with_spaces.txt"},
		{"file/with/slashes.txt", "file_with_slashes.txt"},
		{"file\\with\\backslashes.txt", "file_with_backslashes.txt"},
		{"file:with:colons.txt", "file_with_colons.txt"},
		{"file*with*stars.txt", "file_with_stars.txt"},
		{"file?with?questions.txt", "file_with_questions.txt"},
		{"file\"with\"quotes.txt", "file_with_quotes.txt"},
		{"file<with<brackets>.txt", "file_with_brackets_.txt"},
		{"file|with|pipes.txt", "file_with_pipes.txt"},
		{"", "unnamed"},
		{"   ", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "directory")

	// Directory doesn't exist
	if FileExists(testDir) {
		t.Error("Test directory should not exist initially")
	}

	// Create directory
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	// Directory should exist now
	if !FileExists(testDir) {
		t.Error("Directory should exist after EnsureDir")
	}

	// Should not fail if directory already exists
	err = EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir should not fail for existing directory: %v", err)
	}
}

func TestCalculateFileHash(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is test content for hashing"

	// Create test file
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate hash
	hash, err := CalculateFileHash(testFile)
	if err != nil {
		t.Fatalf("CalculateFileHash failed: %v", err)
	}

	// Hash should not be empty
	if hash == "" {
		t.Error("Hash should not be empty")
	}

	// Hash should be consistent
	hash2, err := CalculateFileHash(testFile)
	if err != nil {
		t.Fatalf("Second CalculateFileHash failed: %v", err)
	}

	if hash != hash2 {
		t.Error("Hash should be consistent across calls")
	}

	// Different content should produce different hash
	differentFile := filepath.Join(tempDir, "different.txt")
	err = os.WriteFile(differentFile, []byte("Different content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create different test file: %v", err)
	}

	differentHash, err := CalculateFileHash(differentFile)
	if err != nil {
		t.Fatalf("CalculateFileHash for different file failed: %v", err)
	}

	if hash == differentHash {
		t.Error("Different files should produce different hashes")
	}

	// Non-existent file should return error
	_, err = CalculateFileHash("/non/existent/file")
	if err == nil {
		t.Error("CalculateFileHash should return error for non-existent file")
	}
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is test content"

	// Create test file
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get file size
	size, err := GetFileSize(testFile)
	if err != nil {
		t.Fatalf("GetFileSize failed: %v", err)
	}

	expectedSize := int64(len(testContent))
	if size != expectedSize {
		t.Errorf("Expected file size %d, got %d", expectedSize, size)
	}

	// Non-existent file should return error
	_, err = GetFileSize("/non/existent/file")
	if err == nil {
		t.Error("GetFileSize should return error for non-existent file")
	}
}

// Benchmark tests
func BenchmarkFileExists(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileExists(testFile)
	}
}

func BenchmarkCalculateFileHash(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := make([]byte, 1024) // 1KB of content
	for i := range content {
		content[i] = byte(i % 256)
	}
	_ = os.WriteFile(testFile, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateFileHash(testFile)
	}
}
