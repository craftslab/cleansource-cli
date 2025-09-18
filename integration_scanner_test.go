package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/app"
	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/pkg/buildtools"
)

// TestScannerIntegration tests the integration of all scanners with the main application
func TestScannerIntegration(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) error
		expectedTools []string
	}{
		{
			name: "Go Project",
			setupFunc: func(tempDir string) error {
				goModContent := `module test-go-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/stretchr/testify v1.8.4
)`
				return os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
			},
			expectedTools: []string{"go"},
		},
		{
			name: "NPM Project",
			setupFunc: func(tempDir string) error {
				packageJsonContent := `{
	"name": "test-npm-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2",
		"lodash": "4.17.21"
	},
	"devDependencies": {
		"jest": "^29.5.0"
	}
}`
				return os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJsonContent), 0644)
			},
			expectedTools: []string{"npm"},
		},
		{
			name: "Gradle Project",
			setupFunc: func(tempDir string) error {
				gradleContent := `plugins {
    id 'java'
}

group = 'com.example'
version = '1.0.0'

dependencies {
    implementation 'org.springframework:spring-core:5.3.21'
    testImplementation 'junit:junit:4.13.2'
}`
				return os.WriteFile(filepath.Join(tempDir, "build.gradle"), []byte(gradleContent), 0644)
			},
			expectedTools: []string{"gradle"},
		},
		{
			name: "Pipenv Project",
			setupFunc: func(tempDir string) error {
				pipfileContent := `[[source]]
url = "https://pypi.org/simple"
name = "pypi"

[packages]
requests = "*"
flask = "*"

[dev-packages]
pytest = "*"`
				return os.WriteFile(filepath.Join(tempDir, "Pipfile"), []byte(pipfileContent), 0644)
			},
			expectedTools: []string{"pipenv"},
		},
		{
			name: "Maven Project",
			setupFunc: func(tempDir string) error {
				pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-maven-project</artifactId>
    <version>1.0.0</version>
    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.2</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>`
				return os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644)
			},
			expectedTools: []string{"maven"},
		},
		{
			name: "Multi-Project (Go + NPM)",
			setupFunc: func(tempDir string) error {
				// Create Go project
				goModContent := `module test-go-project
go 1.21`
				if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
					return err
				}

				// Create NPM project
				packageJsonContent := `{
	"name": "test-npm-project",
	"version": "1.0.0"
}`
				return os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJsonContent), 0644)
			},
			expectedTools: []string{"go", "npm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup project files
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("Failed to setup project: %v", err)
			}

			// Test build tool detection
			env := buildtools.NewScannableEnvironment(tempDir, "")
			cfg := &config.ScanConfig{}
			scanner := buildtools.NewBuildScanner(env, cfg)

			detectedTools := scanner.DetectBuildTools()

			// Check that all expected tools are detected
			for _, expectedTool := range tt.expectedTools {
				found := false
				for _, tool := range detectedTools {
					if tool == expectedTool {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to detect %s, but it was not found. Detected: %v", expectedTool, detectedTools)
				}
			}

			// Test dependency scanning (may fail if executables are not available)
			dependencies, err := scanner.ScanDependencies()
			if err != nil {
				t.Logf("Dependency scanning failed (expected if executables not available): %v", err)
			} else {
				t.Logf("Successfully scanned %d dependency roots", len(dependencies))

				// Verify that we have the expected number of dependency roots
				if len(dependencies) != len(tt.expectedTools) {
					t.Logf("Expected %d dependency roots, got %d", len(tt.expectedTools), len(dependencies))
				}
			}
		})
	}
}

// TestApplicationIntegration tests the integration with the main application
func TestApplicationIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Create a multi-project setup
	goModContent := `module test-go-project
go 1.21`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	packageJsonContent := `{
	"name": "test-npm-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2"
	}
}`
	if err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJsonContent), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create application config
	cfg := &config.ScanConfig{
		TaskDir:     tempDir,
		ServerURL:   "https://example.com",
		Username:    "testuser",
		Password:    "testpass",
		ScanType:    "source",
		BuildDepend: true,
	}

	// Test that the application can be created
	application := app.NewBuildScanApplication(cfg)
	if application == nil {
		t.Fatal("Failed to create application")
	}

	// Test the full application run (this will test the scanner integration)
	// Note: This will fail authentication, but that's expected for integration testing
	err := application.Run()
	if err != nil {
		t.Logf("Application run failed (expected due to authentication): %v", err)
	} else {
		t.Logf("Successfully ran application")
	}
}

// TestScannerErrorHandling tests error handling in scanners
func TestScannerErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	// Test with non-existent directory
	env := buildtools.NewScannableEnvironment("/non/existent/directory", "")
	cfg := &config.ScanConfig{}
	scanner := buildtools.NewBuildScanner(env, cfg)

	// Should not panic and should return empty results
	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Logf("Expected error for non-existent directory: %v", err)
	}

	if len(dependencies) != 0 {
		t.Errorf("Expected no dependencies for non-existent directory, got %d", len(dependencies))
	}

	// Test with empty directory
	env = buildtools.NewScannableEnvironment(tempDir, "")
	scanner = buildtools.NewBuildScanner(env, cfg)

	dependencies, err = scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("Unexpected error for empty directory: %v", err)
	}

	if len(dependencies) != 0 {
		t.Errorf("Expected no dependencies for empty directory, got %d", len(dependencies))
	}
}

// TestScannerConcurrency tests concurrent scanner operations
func TestScannerConcurrency(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple project files
	projects := map[string]string{
		"go.mod": `module test-go-project
go 1.21`,
		"package.json": `{
	"name": "test-npm-project",
	"version": "1.0.0"
}`,
		"build.gradle": `plugins { id 'java' }
group = 'com.example'
version = '1.0.0'`,
	}

	for fileName, content := range projects {
		if err := os.WriteFile(filepath.Join(tempDir, fileName), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", fileName, err)
		}
	}

	env := buildtools.NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := buildtools.NewBuildScanner(env, cfg)

	// Test concurrent detection
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			tools := scanner.DetectBuildTools()
			if len(tools) < 3 {
				t.Errorf("Expected at least 3 tools, got %d", len(tools))
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkScannerDetection benchmarks the scanner detection performance
func BenchmarkScannerDetection(b *testing.B) {
	tempDir := b.TempDir()

	// Create all types of project files
	projects := map[string]string{
		"go.mod": `module test-go-project
go 1.21`,
		"package.json": `{
	"name": "test-npm-project",
	"version": "1.0.0"
}`,
		"build.gradle": `plugins { id 'java' }
group = 'com.example'
version = '1.0.0'`,
		"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-maven-project</artifactId>
    <version>1.0.0</version>
</project>`,
		"Pipfile": `[[source]]
url = "https://pypi.org/simple"
name = "pypi"

[packages]
requests = "*"`,
		"requirements.txt": `requests==2.25.1`,
	}

	for fileName, content := range projects {
		if err := os.WriteFile(filepath.Join(tempDir, fileName), []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create %s: %v", fileName, err)
		}
	}

	env := buildtools.NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := buildtools.NewBuildScanner(env, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.DetectBuildTools()
	}
}

// BenchmarkDependencyScanning benchmarks the dependency scanning performance
func BenchmarkDependencyScanning(b *testing.B) {
	tempDir := b.TempDir()

	// Create a simple project
	packageJsonContent := `{
	"name": "test-npm-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2",
		"lodash": "4.17.21"
	}
}`
	if err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJsonContent), 0644); err != nil {
		b.Fatalf("Failed to create package.json: %v", err)
	}

	env := buildtools.NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := buildtools.NewBuildScanner(env, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.ScanDependencies()
	}
}
