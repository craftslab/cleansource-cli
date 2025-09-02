package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/app"
	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
)

func TestMain(m *testing.M) {
	// Setup
	logger.InitLogger("error") // Reduce log noise during tests

	// Run tests
	code := m.Run()

	// Teardown (if needed)

	os.Exit(code)
}

func TestIntegration_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary project structure
	tempDir := t.TempDir()

	// Create a sample Maven project
	projectDir := filepath.Join(tempDir, "test-project")
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>integration-test</artifactId>
    <version>1.0.0</version>
    <name>Integration Test Project</name>

    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.2</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>`

	pomFile := filepath.Join(projectDir, "pom.xml")
	err = os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	// Create some source files
	srcDir := filepath.Join(projectDir, "src", "main", "java", "com", "example")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	mainJava := `package com.example;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello World!");
    }

    public String getMessage() {
        return "Hello from Main";
    }
}`

	err = os.WriteFile(filepath.Join(srcDir, "Main.java"), []byte(mainJava), 0644)
	if err != nil {
		t.Fatalf("Failed to create Main.java: %v", err)
	}

	utilJava := `package com.example;

import java.util.List;
import java.util.ArrayList;

public class Util {
    public static List<String> getItems() {
        List<String> items = new ArrayList<>();
        items.add("item1");
        items.add("item2");
        return items;
    }
}`

	err = os.WriteFile(filepath.Join(srcDir, "Util.java"), []byte(utilJava), 0644)
	if err != nil {
		t.Fatalf("Failed to create Util.java: %v", err)
	}

	// Create test configuration (without actual server connection)
	cfg := &config.ScanConfig{
		TaskDir:     projectDir,
		ScanType:    "source",
		TaskType:    "scan",
		ToPath:      tempDir,
		BuildDepend: true,
		ThreadNum:   "4",
		LogLevel:    "error",
		// Note: No server config for this test
	}

	// Test configuration validation
	err = cfg.Validate()
	if err == nil {
		t.Error("Expected validation error due to missing server config")
	}

	// Add minimal server config
	cfg.ServerURL = "https://example.com"
	cfg.Username = "testuser"
	cfg.Password = "testpass"

	err = cfg.Validate()
	if err != nil {
		t.Fatalf("Configuration should be valid now: %v", err)
	}

	// Test path setting
	cfg.SetToPath(projectDir)
	expectedToPath := tempDir
	if cfg.ToPath != expectedToPath {
		t.Errorf("Expected ToPath to be %s, got %s", expectedToPath, cfg.ToPath)
	}

	// Create application (but don't run due to no real server)
	application := app.NewBuildScanApplication(cfg)
	if application == nil {
		t.Error("Application should be created successfully")
	}

	// Test directory size calculation
	size, err := application.CalculateDirSize(projectDir)
	if err != nil {
		t.Fatalf("Failed to calculate directory size: %v", err)
	}

	if size == 0 {
		t.Error("Directory size should be greater than 0")
	}

	t.Logf("Project directory size: %d bytes", size)
}

func TestIntegration_ConfigurationScenarios(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *config.ScanConfig
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Valid configuration with token",
			setupFunc: func() *config.ScanConfig {
				cfg := config.NewScanConfig()
				cfg.TaskDir = t.TempDir()
				cfg.ServerURL = "https://example.com"
				cfg.Token = "test-token"
				return cfg
			},
			wantErr: false,
		},
		{
			name: "Valid configuration with username/password",
			setupFunc: func() *config.ScanConfig {
				cfg := config.NewScanConfig()
				cfg.TaskDir = t.TempDir()
				cfg.ServerURL = "https://example.com"
				cfg.Username = "testuser"
				cfg.Password = "testpass"
				return cfg
			},
			wantErr: false,
		},
		{
			name: "Missing task directory",
			setupFunc: func() *config.ScanConfig {
				cfg := config.NewScanConfig()
				cfg.ServerURL = "https://example.com"
				cfg.Token = "test-token"
				return cfg
			},
			wantErr: true,
			errMsg:  "task directory is required",
		},
		{
			name: "Missing server URL",
			setupFunc: func() *config.ScanConfig {
				cfg := config.NewScanConfig()
				cfg.TaskDir = t.TempDir()
				cfg.Token = "test-token"
				return cfg
			},
			wantErr: true,
			errMsg:  "server URL is required",
		},
		{
			name: "Missing authentication",
			setupFunc: func() *config.ScanConfig {
				cfg := config.NewScanConfig()
				cfg.TaskDir = t.TempDir()
				cfg.ServerURL = "https://example.com"
				return cfg
			},
			wantErr: true,
			errMsg:  "username/password or token is required for authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupFunc()
			err := cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestIntegration_MultiLanguageProject(t *testing.T) {
	// Create a multi-language project
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "multi-lang-project")
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Add Maven project
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>multi-lang</artifactId>
    <version>1.0.0</version>
</project>`
	err = os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	// Add Node.js project
	packageContent := `{
  "name": "multi-lang-project",
  "version": "1.0.0",
  "dependencies": {
    "lodash": "^4.17.21",
    "express": "^4.18.0"
  }
}`
	err = os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(packageContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Add Python project
	requirementsContent := `requests==2.25.1
numpy>=1.20.0
pandas==1.3.0`
	err = os.WriteFile(filepath.Join(projectDir, "requirements.txt"), []byte(requirementsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	// Add Go project
	goModContent := `module github.com/example/multi-lang

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/sirupsen/logrus v1.9.3
)`
	err = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test build tool detection (this would be part of the full workflow)
	// For now, just verify the files exist
	expectedFiles := []string{"pom.xml", "package.json", "requirements.txt", "go.mod"}
	for _, file := range expectedFiles {
		fullPath := filepath.Join(projectDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	t.Logf("Multi-language project created successfully with %d build files", len(expectedFiles))
}

// Benchmark for the full integration workflow
func BenchmarkIntegration_ProjectScan(b *testing.B) {
	// Setup a test project
	tempDir := b.TempDir()
	projectDir := filepath.Join(tempDir, "benchmark-project")
	_ = os.MkdirAll(projectDir, 0755)

	// Create multiple source files
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf(`package com.example;

public class Class%d {
    private String field%d;

    public String getField%d() {
        return field%d;
    }

    public void setField%d(String value) {
        this.field%d = value;
    }
}`, i, i, i, i, i, i)

		fileName := filepath.Join(projectDir, fmt.Sprintf("Class%d.java", i))
		_ = os.WriteFile(fileName, []byte(content), 0644)
	}

	cfg := &config.ScanConfig{
		TaskDir:  projectDir,
		ToPath:   tempDir,
		LogLevel: "error",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.SetToPath(projectDir)
		application := app.NewBuildScanApplication(cfg)
		_, _ = application.CalculateDirSize(projectDir)
	}
}
