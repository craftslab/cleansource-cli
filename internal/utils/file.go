package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsDirEmpty checks if a directory is empty
func IsDirEmpty(dirname string) bool {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return true
	}
	return len(entries) == 0
}

// CreateZipArchive creates a ZIP archive of the specified directory
func CreateZipArchive(sourceDir, outputDir string) (string, error) {
	// Create output file path
	baseName := filepath.Base(sourceDir)
	zipPath := filepath.Join(outputDir, baseName+".zip")

	// Create ZIP file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func(zipFile *os.File) {
		_ = zipFile.Close()
	}(zipFile)

	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		_ = zipWriter.Close()
	}(zipWriter)

	// Walk through source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and certain files
		if info.IsDir() || shouldSkipForArchive(path) {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Normalize path separators for ZIP
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Create file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		// Create writer for this file
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		_ = os.Remove(zipPath) // Clean up on error
		return "", fmt.Errorf("failed to create archive: %w", err)
	}

	return zipPath, nil
}

// shouldSkipForArchive determines if a file should be skipped when creating archives
func shouldSkipForArchive(path string) bool {
	// Skip hidden files and directories
	if strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	// Skip common build and dependency directories
	skipPatterns := []string{
		"node_modules", "vendor", "target", "build", ".git",
		".svn", ".hg", "__pycache__", ".tox", "dist", ".gradle",
		".idea", ".vscode", "*.tmp", "*.log",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// IsTextFile determines if a file is likely a text file based on its extension
func IsTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	textExts := []string{
		".txt", ".md", ".go", ".java", ".py", ".js", ".ts", ".html", ".css",
		".xml", ".json", ".yaml", ".yml", ".toml", ".ini", ".conf", ".cfg",
		".sh", ".bat", ".cmd", ".ps1", ".sql", ".log", ".properties",
		".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rb", ".pl", ".r",
		".scala", ".kotlin", ".swift", ".dart", ".rust", ".rs",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	return false
}

// NormalizePath normalizes a file path for cross-platform compatibility
func NormalizePath(path string) string {
	return filepath.ToSlash(path)
}
