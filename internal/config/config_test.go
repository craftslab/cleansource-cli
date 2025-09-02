package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewScanConfig(t *testing.T) {
	cfg := NewScanConfig()

	// Test default values
	if cfg.ScanType != "source" {
		t.Errorf("Expected default ScanType to be 'source', got %s", cfg.ScanType)
	}
	if cfg.TaskType != "scan" {
		t.Errorf("Expected default TaskType to be 'scan', got %s", cfg.TaskType)
	}
	if !cfg.BuildDepend {
		t.Error("Expected default BuildDepend to be true")
	}
	if cfg.ThreadNum != "30" {
		t.Errorf("Expected default ThreadNum to be '30', got %s", cfg.ThreadNum)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel to be 'info', got %s", cfg.LogLevel)
	}

	// Test default param
	if cfg.DefaultParam == nil {
		t.Error("Expected DefaultParam to be initialized")
	} else {
		if cfg.DefaultParam.ScanWay != 1 {
			t.Errorf("Expected default ScanWay to be 1, got %d", cfg.DefaultParam.ScanWay)
		}
		if cfg.DefaultParam.IsSaveSourceFile != 0 {
			t.Errorf("Expected default IsSaveSourceFile to be 0, got %d", cfg.DefaultParam.IsSaveSourceFile)
		}
	}
}

func TestScanConfig_SetToPath(t *testing.T) {
	tests := []struct {
		name        string
		initialPath string
		scanPath    string
		expected    string
	}{
		{
			name:        "Empty toPath with valid scanPath",
			initialPath: "",
			scanPath:    "/home/user/project/src",
			expected:    "/home/user/project",
		},
		{
			name:        "Empty toPath with empty scanPath",
			initialPath: "",
			scanPath:    "",
			expected:    "", // Will be current directory
		},
		{
			name:        "Existing toPath should be converted to absolute",
			initialPath: "relative/path",
			scanPath:    "/some/scan/path",
			expected:    "", // Will be absolute path of relative/path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewScanConfig()
			cfg.ToPath = tt.initialPath
			cfg.SetToPath(tt.scanPath)

			switch tt.name {
			case "Empty toPath with valid scanPath":
				expected := filepath.Dir(tt.scanPath)
				if cfg.ToPath != expected {
					t.Errorf("Expected ToPath to be %s, got %s", expected, cfg.ToPath)
				}
			case "Empty toPath with empty scanPath":
				// Should be current working directory
				if cfg.ToPath == "" {
					t.Error("Expected ToPath to be set to current directory")
				}
			case "Existing toPath should be converted to absolute":
				// Should be absolute path
				if !filepath.IsAbs(cfg.ToPath) {
					t.Error("Expected ToPath to be converted to absolute path")
				}
			}
		})
	}
}

func TestScanConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *ScanConfig
		wantErr   error
	}{
		{
			name: "Valid configuration with username/password",
			setupFunc: func() *ScanConfig {
				cfg := NewScanConfig()
				cfg.TaskDir = "/tmp/test"
				cfg.ServerURL = "https://example.com"
				cfg.Username = "testuser"
				cfg.Password = "testpass"
				return cfg
			},
			wantErr: nil,
		},
		{
			name: "Valid configuration with token",
			setupFunc: func() *ScanConfig {
				cfg := NewScanConfig()
				cfg.TaskDir = "/tmp/test"
				cfg.ServerURL = "https://example.com"
				cfg.Token = "test-token"
				return cfg
			},
			wantErr: nil,
		},
		{
			name: "Missing task directory",
			setupFunc: func() *ScanConfig {
				cfg := NewScanConfig()
				cfg.ServerURL = "https://example.com"
				cfg.Username = "testuser"
				cfg.Password = "testpass"
				return cfg
			},
			wantErr: ErrMissingTaskDir,
		},
		{
			name: "Missing server URL",
			setupFunc: func() *ScanConfig {
				cfg := NewScanConfig()
				cfg.TaskDir = "/tmp/test"
				cfg.Username = "testuser"
				cfg.Password = "testpass"
				return cfg
			},
			wantErr: ErrMissingServerURL,
		},
		{
			name: "Missing authentication",
			setupFunc: func() *ScanConfig {
				cfg := NewScanConfig()
				cfg.TaskDir = "/tmp/test"
				cfg.ServerURL = "https://example.com"
				return cfg
			},
			wantErr: ErrMissingAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupFunc()
			err := cfg.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestAuthType(t *testing.T) {
	if AuthTypeCookie != 0 {
		t.Errorf("Expected AuthTypeCookie to be 0, got %d", AuthTypeCookie)
	}
	if AuthTypeToken != 1 {
		t.Errorf("Expected AuthTypeToken to be 1, got %d", AuthTypeToken)
	}
}

// Helper function to create a temporary directory for testing
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "test-scan-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func TestScanConfig_SetToPath_Integration(t *testing.T) {
	// Create a temporary directory structure
	tempDir := createTempDir(t)
	scanDir := filepath.Join(tempDir, "project", "src")
	err := os.MkdirAll(scanDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create scan directory: %v", err)
	}

	cfg := NewScanConfig()
	cfg.SetToPath(scanDir)

	expectedParent := filepath.Join(tempDir, "project")
	if cfg.ToPath != expectedParent {
		t.Errorf("Expected ToPath to be %s, got %s", expectedParent, cfg.ToPath)
	}
}
