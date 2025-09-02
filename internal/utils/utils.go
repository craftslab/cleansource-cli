package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// SanitizeFileName removes or replaces invalid characters from a filename
func SanitizeFileName(fileName string) string {
	if strings.TrimSpace(fileName) == "" {
		return "unnamed"
	}

	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[<>:"/\\|?*\s]`)
	sanitized := reg.ReplaceAllString(fileName, "_")

	// Remove multiple consecutive underscores
	reg = regexp.MustCompile(`_+`)
	sanitized = reg.ReplaceAllString(sanitized, "_")

	// Trim underscores from start and end
	sanitized = strings.Trim(sanitized, "_")

	if sanitized == "" {
		return "unnamed"
	}

	return sanitized
}

// CalculateFileHash calculates SHA256 hash of a file
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
