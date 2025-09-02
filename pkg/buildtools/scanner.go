package buildtools

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
)

// ScannableEnvironment represents the scanning environment
type ScannableEnvironment struct {
	directory     string
	buildTreeFile string
	buildFile     string // Add missing buildFile field
	log           *logrus.Logger
}

// NewScannableEnvironment creates a new scannable environment
func NewScannableEnvironment(directory, buildFile string) *ScannableEnvironment {
	return &ScannableEnvironment{
		directory:     directory,
		buildTreeFile: "",        // Will be set later if needed
		buildFile:     buildFile, // Use the passed parameter as buildFile
		log:           logger.GetLogger(),
	}
}

// GetDirectory returns the scanning directory
func (se *ScannableEnvironment) GetDirectory() string {
	return se.directory
}

// GetBuildTreeFile returns the build tree file path
func (se *ScannableEnvironment) GetBuildTreeFile() string {
	return se.buildTreeFile
}

// GetBuildFile returns the build file path
func (se *ScannableEnvironment) GetBuildFile() string {
	return se.buildFile
}

// SetBuildFile sets the build file path
func (se *ScannableEnvironment) SetBuildFile(buildFile string) {
	se.buildFile = buildFile
}

// Scannable represents an interface for build tool scanners
type Scannable interface {
	ExeFind() error
	FileFind() error
	ScanExecute() ([]model.DependencyRoot, error)
}

// BuildScanner manages different build tool scanners
type BuildScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	scanners    []Scannable
	log         *logrus.Logger
}

// NewBuildScanner creates a new build scanner
func NewBuildScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *BuildScanner {
	scanner := &BuildScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}

	// Initialize scanners based on detected build tools
	scanner.initializeScanners()
	return scanner
}

// initializeScanners initializes the appropriate scanners based on detected build files
func (bs *BuildScanner) initializeScanners() {
	scanDir := bs.environment.GetDirectory()

	// Check for Maven
	if bs.fileExists(filepath.Join(scanDir, "pom.xml")) {
		bs.scanners = append(bs.scanners, NewMavenScanner(bs.environment, bs.config))
		bs.log.Info("Detected Maven project")
	}

	// Check for Gradle
	if bs.fileExists(filepath.Join(scanDir, "build.gradle")) ||
		bs.fileExists(filepath.Join(scanDir, "build.gradle.kts")) {
		bs.scanners = append(bs.scanners, NewGradleScanner(bs.environment, bs.config))
		bs.log.Info("Detected Gradle project")
	}

	// Check for Python pip
	if bs.fileExists(filepath.Join(scanDir, "requirements.txt")) ||
		bs.fileExists(filepath.Join(scanDir, "setup.py")) ||
		bs.fileExists(filepath.Join(scanDir, "pyproject.toml")) {
		bs.scanners = append(bs.scanners, NewPipScanner(bs.environment, bs.config))
		bs.log.Info("Detected Python pip project")
	}

	// Check for Pipenv
	if bs.fileExists(filepath.Join(scanDir, "Pipfile")) {
		bs.scanners = append(bs.scanners, NewPipenvScanner(bs.environment, bs.config))
		bs.log.Info("Detected Python Pipenv project")
	}

	// Check for Node.js
	if bs.fileExists(filepath.Join(scanDir, "package.json")) {
		bs.scanners = append(bs.scanners, NewNpmScanner(bs.environment, bs.config))
		bs.log.Info("Detected Node.js project")
	}

	// Check for Go
	if bs.fileExists(filepath.Join(scanDir, "go.mod")) {
		bs.scanners = append(bs.scanners, NewGoScanner(bs.environment, bs.config))
		bs.log.Info("Detected Go project")
	}

	if len(bs.scanners) == 0 {
		bs.log.Warn("No supported build tools detected")
	}
}

// ScanDependencies scans dependencies using all detected scanners
func (bs *BuildScanner) ScanDependencies() ([]model.DependencyRoot, error) {
	var allDependencies []model.DependencyRoot

	for _, scanner := range bs.scanners {
		// Check if executable is available
		if err := scanner.ExeFind(); err != nil {
			bs.log.Warnf("Executable not found for scanner: %v", err)
			continue
		}

		// Check if required files exist
		if err := scanner.FileFind(); err != nil {
			bs.log.Warnf("Required files not found for scanner: %v", err)
			continue
		}

		// Execute scan
		dependencies, err := scanner.ScanExecute()
		if err != nil {
			bs.log.Warnf("Scan execution failed: %v", err)
			continue
		}

		allDependencies = append(allDependencies, dependencies...)
	}

	return allDependencies, nil
}

// fileExists checks if a file exists
func (bs *BuildScanner) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DetectBuildTools detects build tools in the environment
func (bs *BuildScanner) DetectBuildTools() []string {
	var detectedTools []string
	scanDir := bs.environment.GetDirectory()

	buildFiles := map[string]string{
		"pom.xml":          "maven",
		"build.gradle":     "gradle",
		"build.gradle.kts": "gradle",
		"requirements.txt": "pip",
		"setup.py":         "pip",
		"pyproject.toml":   "pip",
		"Pipfile":          "pipenv",
		"package.json":     "npm",
		"go.mod":           "go",
		"Cargo.toml":       "cargo",
		"composer.json":    "composer",
	}

	for fileName, toolName := range buildFiles {
		if bs.fileExists(filepath.Join(scanDir, fileName)) {
			detectedTools = append(detectedTools, toolName)
		}
	}

	return detectedTools
}

// detectBuildToolFromFile detects build tool from a specific file
func detectBuildToolFromFile(filePath string) (string, bool) {
	baseName := filepath.Base(filePath)

	buildFiles := map[string]string{
		"pom.xml":          "maven",
		"build.gradle":     "gradle",
		"build.gradle.kts": "gradle",
		"requirements.txt": "pip",
		"setup.py":         "pip",
		"pyproject.toml":   "pip",
		"Pipfile":          "pipenv",
		"package.json":     "npm",
		"go.mod":           "go",
		"Cargo.toml":       "cargo",
		"composer.json":    "composer",
	}

	if tool, exists := buildFiles[baseName]; exists {
		return tool, true
	}

	return "", false
}
