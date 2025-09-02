package config

import (
	"os"
	"path/filepath"
)

// ScanConfig represents the main configuration for the build scanner
type ScanConfig struct {
	// Authentication
	ServerURL string
	Username  string
	Password  string
	Token     string
	AuthType  AuthType

	// Project information
	CustomProject string
	CustomProduct string
	CustomVersion string

	// Scan parameters
	TaskDir     string
	ScanType    string
	TaskType    string
	ToPath      string
	BuildDepend bool
	LicenseName string
	ThreadNum   string
	LogLevel    string

	// Notification
	NotificationEmail string

	// Build tool paths
	MavenPath           string
	MavenBuildCommand   string
	PipPath             string
	PipRequirementsPath string

	// Default parameters
	DefaultParam *DefaultParamInfo
}

// DefaultParamInfo represents default scanning parameters
type DefaultParamInfo struct {
	ScanWay                  int      `json:"scanWay"`
	IsSaveSourceFile         int      `json:"isSaveSourceFile"`
	MixedBinaryScanFlag      int      `json:"mixedBinaryScanFlag"`
	MixedBinaryScanFilePaths []string `json:"mixedBinaryScanFilePaths"`
}

// AuthType represents authentication type
type AuthType int

const (
	AuthTypeCookie AuthType = iota
	AuthTypeToken
)

// NewScanConfig creates a new scan configuration with default values
func NewScanConfig() *ScanConfig {
	return &ScanConfig{
		ScanType:    "source",
		TaskType:    "scan",
		BuildDepend: true,
		ThreadNum:   "30",
		LogLevel:    "info",
		DefaultParam: &DefaultParamInfo{
			ScanWay:             1, // Full scan
			IsSaveSourceFile:    0,
			MixedBinaryScanFlag: 0,
		},
	}
}

// SetToPath sets the output path, using parent of scan directory if not specified
func (c *ScanConfig) SetToPath(scanPath string) {
	if c.ToPath == "" {
		if scanPath != "" {
			c.ToPath = filepath.Dir(scanPath)
		} else {
			c.ToPath, _ = os.Getwd()
		}
	} else {
		absPath, err := filepath.Abs(c.ToPath)
		if err == nil {
			c.ToPath = absPath
		}
	}
}

// Validate validates the configuration
func (c *ScanConfig) Validate() error {
	if c.TaskDir == "" {
		return ErrMissingTaskDir
	}
	if c.ServerURL == "" {
		return ErrMissingServerURL
	}
	if c.Username == "" && c.Token == "" {
		return ErrMissingAuth
	}
	return nil
}
