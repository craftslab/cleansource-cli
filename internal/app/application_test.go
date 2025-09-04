package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
)

// nolint: staticcheck
func TestNewBuildScanApplication(t *testing.T) {
	cfg := &config.ScanConfig{
		ServerURL: "https://example.com",
		TaskDir:   "/tmp/test",
	}

	app := NewBuildScanApplication(cfg)

	if app == nil {
		t.Error("NewBuildScanApplication should not return nil")
	}

	if app.config != cfg {
		t.Error("Application should store the provided config")
	}

	if app.client == nil {
		t.Error("Application should initialize the client")
	}
}

func TestBuildScanApplication_Run_ValidationError(t *testing.T) {
	// Create config with missing required fields
	cfg := &config.ScanConfig{
		TaskDir: "", // Missing task directory
	}

	app := NewBuildScanApplication(cfg)
	err := app.Run()

	if err == nil {
		t.Error("Run should return error for invalid configuration")
	}
}

func TestBuildScanApplication_Run_UnsupportedScanType(t *testing.T) {
	cfg := &config.ScanConfig{
		TaskDir:   "/tmp/test",
		ServerURL: "https://example.com",
		Username:  "testuser",
		Password:  "testpass",
		ScanType:  "unsupported",
	}

	app := NewBuildScanApplication(cfg)
	err := app.Run()

	if err == nil {
		t.Error("Run should return error for unsupported scan type")
	}

	if err.Error() != "unsupported scan type: unsupported" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

func TestBuildScanApplication_calculateDirSize(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create some test files
	testFiles := map[string]string{
		"file1.txt":             "Hello World",
		"file2.txt":             "This is a longer file with more content",
		"subdir/file3.txt":      "Nested file",
		"subdir/sub2/file4.txt": "Deeply nested file",
	}

	expectedSize := int64(0)
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
		expectedSize += int64(len(content))
	}

	cfg := &config.ScanConfig{}
	app := NewBuildScanApplication(cfg)

	calculatedSize, err := app.calculateDirSize(tempDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}

	if calculatedSize != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, calculatedSize)
	}
}

func TestBuildScanApplication_calculateDirSize_NonExistentDir(t *testing.T) {
	cfg := &config.ScanConfig{}
	app := NewBuildScanApplication(cfg)

	size, err := app.calculateDirSize("/non/existent/directory")

	// Should handle error gracefully and return 0
	if size != 0 {
		t.Errorf("Expected size 0 for non-existent directory, got %d", size)
	}

	// Error is expected for non-existent directory
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestBuildScanApplication_calculateDirSize_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.ScanConfig{}
	app := NewBuildScanApplication(cfg)

	size, err := app.calculateDirSize(tempDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected size 0 for empty directory, got %d", size)
	}
}

func TestBuildScanApplication_calculateDirSize_LargeDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple files to test concurrent processing
	numFiles := 50
	fileSize := 1000 // bytes per file

	for i := 0; i < numFiles; i++ {
		fileName := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		content := make([]byte, fileSize)
		for j := range content {
			content[j] = byte('A' + (j % 26))
		}
		err := os.WriteFile(fileName, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	cfg := &config.ScanConfig{}
	app := NewBuildScanApplication(cfg)

	size, err := app.calculateDirSize(tempDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}

	expectedSize := int64(numFiles * fileSize)
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

// Mock tests would require more complex setup, so here are basic integration tests

func TestBuildScanApplication_runSourceScan_NonExistentDir(t *testing.T) {
	cfg := &config.ScanConfig{
		TaskDir:   "/non/existent/directory",
		ServerURL: "https://example.com",
		Username:  "testuser",
		Password:  "testpass",
		ScanType:  "source",
	}

	app := NewBuildScanApplication(cfg)
	err := app.runSourceScan()

	if err == nil {
		t.Error("runSourceScan should return error for non-existent directory")
	}
}

func TestBuildScanApplication_runDockerScan_NotImplemented(t *testing.T) {
	cfg := &config.ScanConfig{
		TaskDir:   "/tmp/test",
		ServerURL: "https://example.com",
		Username:  "testuser",
		Password:  "testpass",
		ScanType:  "docker",
	}

	app := NewBuildScanApplication(cfg)
	err := app.runDockerScan()

	if err == nil {
		t.Error("runDockerScan should return error as it's not implemented")
	}

	if err.Error() != "docker scan not yet implemented" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

func TestBuildScanApplication_runBinaryScan_NotImplemented(t *testing.T) {
	cfg := &config.ScanConfig{
		TaskDir:   "/tmp/test",
		ServerURL: "https://example.com",
		Username:  "testuser",
		Password:  "testpass",
		ScanType:  "binary",
	}

	app := NewBuildScanApplication(cfg)
	err := app.runBinaryScan()

	if err == nil {
		t.Error("runBinaryScan should return error as it's not implemented")
	}

	if err.Error() != "binary scan not yet implemented" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

// Benchmark tests
func BenchmarkBuildScanApplication_calculateDirSize(b *testing.B) {
	// Create a test directory with multiple files
	tempDir := b.TempDir()

	for i := 0; i < 100; i++ {
		fileName := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		content := make([]byte, 1000)
		_ = os.WriteFile(fileName, content, 0644)
	}

	cfg := &config.ScanConfig{}
	app := NewBuildScanApplication(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = app.calculateDirSize(tempDir)
	}
}
