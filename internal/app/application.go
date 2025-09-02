package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
	"github.com/craftslab/cleansource-sca-cli/internal/scanner"
	"github.com/craftslab/cleansource-sca-cli/internal/utils"
	"github.com/craftslab/cleansource-sca-cli/pkg/buildtools"
	"github.com/craftslab/cleansource-sca-cli/pkg/client"
)

// BuildScanApplication represents the main application
type BuildScanApplication struct {
	config *config.ScanConfig
	client *client.RemotingClient
	log    *logrus.Logger
}

// NewBuildScanApplication creates a new application instance
func NewBuildScanApplication(cfg *config.ScanConfig) *BuildScanApplication {
	return &BuildScanApplication{
		config: cfg,
		client: client.NewRemotingClient(cfg.ServerURL),
		log:    logger.GetLogger(),
	}
}

// Run executes the main application logic
func (app *BuildScanApplication) Run() error {
	// Validate configuration
	if err := app.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set output path
	app.config.SetToPath(app.config.TaskDir)

	switch app.config.ScanType {
	case "source":
		return app.runSourceScan()
	case "docker":
		return app.runDockerScan()
	case "binary":
		return app.runBinaryScan()
	default:
		return fmt.Errorf("unsupported scan type: %s", app.config.ScanType)
	}
}

// runSourceScan handles source code scanning
func (app *BuildScanApplication) runSourceScan() error {
	// Verify authentication
	if err := app.verifyAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Check scan directory
	taskDir := app.config.TaskDir
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return fmt.Errorf("scan directory does not exist: %s", taskDir)
	}

	if utils.IsDirEmpty(taskDir) {
		app.log.Warn("Scan directory is empty, scan end!")
		return nil
	}

	// Calculate directory size
	dirSize, err := app.calculateDirSize(taskDir)
	if err != nil {
		app.log.Warnf("Failed to calculate directory size: %v", err)
		dirSize = 0
	}
	app.log.Infof("Scan directory: %s, size: %d bytes", taskDir, dirSize)

	// Create scannable environment
	env := buildtools.NewScannableEnvironment(taskDir, "")

	// Generate fingerprint file
	app.log.Info("Generating fingerprint file...")
	wfpFile, err := app.generateWfpFile(env)
	if err != nil {
		return fmt.Errorf("failed to generate fingerprint file: %w", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(wfpFile) // Clean up

	// Build dependency information if enabled
	var buildFile string
	if app.config.BuildDepend {
		app.log.Info("Building dependency information...")
		buildFile, err = app.buildDependencyInfo(env)
		if err != nil {
			app.log.Warnf("Failed to build dependency information: %v", err)
		}
		if buildFile != "" {
			defer func(name string) {
				_ = os.Remove(name)
			}(buildFile) // Clean up
		}
	}

	// Create archive if needed
	var archiveFile string
	if app.config.DefaultParam.IsSaveSourceFile == 1 {
		app.log.Info("Creating source archive...")
		archiveFile, err = utils.CreateZipArchive(taskDir, app.config.ToPath)
		if err != nil {
			app.log.Warnf("Failed to create archive: %v", err)
		}
		if archiveFile != "" {
			defer func(name string) {
				_ = os.Remove(name)
			}(archiveFile) // Clean up
		}
	}

	// Upload data to server
	app.log.Info("Uploading scan data...")
	uploadData := &model.UploadData{
		WfpFile:     wfpFile,
		BuildFile:   buildFile,
		ArchiveFile: archiveFile,
		Config:      app.config,
		DirSize:     dirSize,
	}

	success, err := app.client.UploadData(uploadData)
	if err != nil {
		return fmt.Errorf("failed to upload data: %w", err)
	}

	if !success {
		return fmt.Errorf("upload was not successful")
	}

	app.log.Info("Scan completed successfully")
	return nil
}

// runDockerScan handles Docker image scanning
func (app *BuildScanApplication) runDockerScan() error {
	app.log.Info("Starting Docker scan...")
	// Implementation would go here
	return fmt.Errorf("docker scan not yet implemented")
}

// runBinaryScan handles binary file scanning
func (app *BuildScanApplication) runBinaryScan() error {
	app.log.Info("Starting binary scan...")
	// Implementation would go here
	return fmt.Errorf("binary scan not yet implemented")
}

// verifyAuth verifies authentication with the server
func (app *BuildScanApplication) verifyAuth() error {
	if app.config.Token != "" {
		app.log.Info("Verifying token...")
		app.config.AuthType = config.AuthTypeToken
		return app.client.VerifyToken(app.config.Token)
	} else {
		app.log.Info("Logging in with username/password...")
		app.config.AuthType = config.AuthTypeCookie
		return app.client.Login(app.config.Username, app.config.Password)
	}
}

// generateWfpFile generates a fingerprint file for the source code
func (app *BuildScanApplication) generateWfpFile(env *buildtools.ScannableEnvironment) (string, error) {
	wfpScanner := scanner.NewWfpScanner(app.config)
	return wfpScanner.GenerateWfpFile(env.GetDirectory())
}

// buildDependencyInfo builds dependency information
func (app *BuildScanApplication) buildDependencyInfo(env *buildtools.ScannableEnvironment) (string, error) {
	// Detect build tools and create appropriate scanner
	buildScanner := buildtools.NewBuildScanner(env, app.config)
	dependencies, err := buildScanner.ScanDependencies()
	if err != nil {
		return "", err
	}

	// Convert to JSON and write to file
	jsonData, err := json.MarshalIndent(dependencies, "", "  ")
	if err != nil {
		return "", err
	}

	buildFile := filepath.Join(app.config.ToPath, "dependencies.json")
	err = os.WriteFile(buildFile, jsonData, 0644)
	if err != nil {
		return "", err
	}

	return buildFile, nil
}

// calculateDirSize calculates the total size of a directory using concurrent processing
func (app *BuildScanApplication) calculateDirSize(rootDir string) (int64, error) {
	// Check if directory exists first
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return 0, fmt.Errorf("directory does not exist: %s", rootDir)
	}

	var totalSize int64
	var wg sync.WaitGroup
	sizeChan := make(chan int64, 100)

	// Limit concurrent goroutines
	semaphore := make(chan struct{}, runtime.NumCPU()*2)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there's an error with individual files
		}

		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		go func(size int64) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			sizeChan <- size
		}(info.Size())

		return nil
	})

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(sizeChan)
	}()

	// Sum up all sizes
	for size := range sizeChan {
		atomic.AddInt64(&totalSize, size)
	}

	return totalSize, err
}

// CalculateDirSize is a public wrapper for testing
func (app *BuildScanApplication) CalculateDirSize(rootDir string) (int64, error) {
	return app.calculateDirSize(rootDir)
}
