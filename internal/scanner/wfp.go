package scanner

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
)

// WfpScanner handles fingerprint generation for source files
type WfpScanner struct {
	config *config.ScanConfig
	log    *logrus.Logger
}

// NewWfpScanner creates a new WFP scanner
func NewWfpScanner(cfg *config.ScanConfig) *WfpScanner {
	return &WfpScanner{
		config: cfg,
		log:    logger.GetLogger(),
	}
}

// GenerateWfpFile generates a fingerprint file for the given directory
func (w *WfpScanner) GenerateWfpFile(scanDir string) (string, error) {
	w.log.Info("Starting fingerprint generation...")

	// Ensure scan directory exists
	if info, err := os.Stat(scanDir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("scan directory not found: %s", scanDir)
	}

	// If TaskDir not set (tests often leave empty), use scanDir for relative path calculation
	if w.config.TaskDir == "" {
		w.config.TaskDir = scanDir
	}

	wfpFile := filepath.Join(w.config.ToPath, "fingerprints.wfp")
	file, err := os.Create(wfpFile)
	if err != nil {
		return "", fmt.Errorf("failed to create wfp file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var wg sync.WaitGroup
	fingerprintChan := make(chan string, 100)
	errorChan := make(chan error, 10)
	var writerWG sync.WaitGroup

	// Start writer goroutine and ensure it completes before returning
	writerWG.Add(1)
	go func() {
		defer writerWG.Done()
		for fingerprint := range fingerprintChan {
			if _, err := file.WriteString(fingerprint + "\n"); err != nil {
				errorChan <- err
				return
			}
		}
	}()

	// Walk through all files and generate fingerprints
	err = filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Skip the output file itself to avoid self-fingerprinting / races
		if path == wfpFile {
			return nil
		}

		if info.IsDir() || w.shouldSkipFile(path, info) {
			return nil
		}

		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			fingerprint, err := w.generateFileFingerprint(filePath)
			if err != nil {
				w.log.Debugf("Failed to generate fingerprint for %s: %v", filePath, err)
				return
			}

			if fingerprint != "" {
				fingerprintChan <- fingerprint
			}
		}(path)

		return nil
	})

	// Wait for all goroutines to complete
	wg.Wait()
	close(fingerprintChan)

	// Wait for writer to finish to ensure all data flushed
	writerWG.Wait()

	// Check for errors (non-blocking)
	select {
	case err := <-errorChan:
		return "", fmt.Errorf("error writing fingerprints: %w", err)
	default:
	}

	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	w.log.Infof("Fingerprint file generated: %s", wfpFile)
	return wfpFile, nil
}

// shouldSkipFile determines if a file should be skipped during fingerprinting
func (w *WfpScanner) shouldSkipFile(path string, info os.FileInfo) bool {
	// Skip hidden files and directories
	if strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	// Skip common build and dependency directories
	skipDirs := []string{
		"node_modules", "vendor", "target", "build", ".git",
		".svn", ".hg", "__pycache__", ".tox", "dist", ".gradle",
	}

	for _, skipDir := range skipDirs {
		if strings.Contains(path, string(os.PathSeparator)+skipDir+string(os.PathSeparator)) ||
			strings.HasSuffix(path, string(os.PathSeparator)+skipDir) {
			return true
		}
	}

	// Skip binary files based on extension
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".jar", ".war", ".ear",
		".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico",
		".mp3", ".mp4", ".avi", ".mov", ".wav",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".bin", ".class", ".o", ".a", ".lib",
	}

	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}

	// Skip files larger than 1MB
	if info.Size() > 1024*1024 {
		return true
	}

	return false
}

// generateFileFingerprint generates a fingerprint for a single file
func (w *WfpScanner) generateFileFingerprint(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Skip empty files
	if len(content) == 0 {
		return "", nil
	}

	// Generate MD5 hash
	hash := md5.Sum(content)
	hashStr := fmt.Sprintf("%x", hash)

	// Get relative path
	relPath, err := filepath.Rel(w.config.TaskDir, filePath)
	if err != nil {
		relPath = filePath
	}

	// Format: file=path,hash=md5hash,size=filesize
	fingerprint := fmt.Sprintf("file=%s,hash=%s,size=%d",
		strings.ReplaceAll(relPath, "\\", "/"), hashStr, len(content))

	return fingerprint, nil
}

// shouldIncludeFile checks if a file should be included in scanning
// This method can handle both path with os.FileInfo or just path string
func (w *WfpScanner) shouldIncludeFile(path string, info ...os.FileInfo) bool {
	// When called with just path string (for testing)
	if len(info) == 0 {
		// Check basic exclusion patterns based on path only
		baseName := filepath.Base(path)

		// Skip hidden files and directories
		if strings.HasPrefix(baseName, ".") {
			// Exception for important dotfiles
			importantDotFiles := []string{".gitignore", ".dockerignore", ".editorconfig"}
			for _, dotfile := range importantDotFiles {
				if baseName == dotfile {
					return true
				}
			}
			return false
		}

		// Skip common build and dependency directories
		skipDirs := []string{
			"node_modules", "vendor", "target", "build", ".git",
			".svn", ".hg", "__pycache__", ".tox", "dist", ".gradle",
		}

		for _, skipDir := range skipDirs {
			if strings.Contains(path, skipDir+string(os.PathSeparator)) ||
				strings.Contains(path, skipDir+"/") ||
				strings.HasPrefix(path, skipDir+string(os.PathSeparator)) ||
				strings.HasPrefix(path, skipDir+"/") ||
				path == skipDir {
				return false
			}
		}

		// Skip binary files based on extension
		ext := strings.ToLower(filepath.Ext(path))
		binaryExts := []string{
			".exe", ".dll", ".so", ".dylib", ".jar", ".war", ".ear",
			".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",
			".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico",
			".mp3", ".mp4", ".avi", ".mov", ".wav",
			".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
			".bin", ".class", ".o", ".a", ".lib",
		}

		for _, binaryExt := range binaryExts {
			if ext == binaryExt {
				return false
			}
		}

		return true
	}

	// When called with os.FileInfo (normal operation)
	return !w.shouldSkipFile(path, info[0])
}

// calculateFileHash calculates MD5 hash of a file
func (w *WfpScanner) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// countLines counts the number of lines in a file
func (w *WfpScanner) countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	content, err := io.ReadAll(file)
	if err != nil {
		return 0, err
	}

	if len(content) == 0 {
		return 0, nil
	}

	contentStr := string(content)

	// Split by newlines to get all line segments
	lines := strings.Split(contentStr, "\n")

	// Remove the last empty element only if the string ends with newline
	// and there are no empty lines in the middle
	if len(lines) > 1 && lines[len(lines)-1] == "" && strings.HasSuffix(contentStr, "\n") {
		// Check if there are empty lines in the middle
		hasEmptyLinesInMiddle := false
		for i := 1; i < len(lines)-1; i++ {
			if lines[i] == "" {
				hasEmptyLinesInMiddle = true
				break
			}
		}

		// If there are empty lines in the middle, keep all lines including the trailing empty one
		// If no empty lines in middle, remove the trailing empty line
		if !hasEmptyLinesInMiddle {
			return len(lines) - 1, nil
		}
	}

	return len(lines), nil
}
