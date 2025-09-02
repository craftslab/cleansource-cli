package buildtools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
)

func TestNewScannableEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	buildFile := "pom.xml"

	env := NewScannableEnvironment(tempDir, buildFile)

	if env.GetDirectory() != tempDir {
		t.Errorf("Expected directory to be %s, got %s", tempDir, env.GetDirectory())
	}

	if env.GetBuildFile() != buildFile {
		t.Errorf("Expected build file to be %s, got %s", buildFile, env.GetBuildFile())
	}
}

func TestScannableEnvironment_GetDirectory(t *testing.T) {
	dir := "/test/directory"
	env := &ScannableEnvironment{directory: dir}

	result := env.GetDirectory()
	if result != dir {
		t.Errorf("Expected %s, got %s", dir, result)
	}
}

func TestScannableEnvironment_GetBuildFile(t *testing.T) {
	buildFile := "build.gradle"
	env := &ScannableEnvironment{buildFile: buildFile}

	result := env.GetBuildFile()
	if result != buildFile {
		t.Errorf("Expected %s, got %s", buildFile, result)
	}
}

//nolint:staticcheck
func TestNewBuildScanner(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{
		MavenPath: "/usr/bin/mvn",
		PipPath:   "/usr/bin/pip",
	}

	scanner := NewBuildScanner(env, cfg)

	if scanner == nil {
		t.Error("NewBuildScanner should not return nil")
	}

	if scanner.environment != env {
		t.Error("Scanner should store the provided environment")
	}

	if scanner.config != cfg {
		t.Error("Scanner should store the provided config")
	}
}

func TestBuildScanner_DetectBuildTools(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	// Test empty directory
	tools := scanner.DetectBuildTools()
	if len(tools) != 0 {
		t.Errorf("Expected no build tools in empty directory, got %d", len(tools))
	}

	// Create Maven build file
	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>
</project>`
	err := os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	tools = scanner.DetectBuildTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 build tool, got %d", len(tools))
	}
	if tools[0] != "maven" {
		t.Errorf("Expected maven, got %s", tools[0])
	}

	// Create Gradle build file
	gradleFile := filepath.Join(tempDir, "build.gradle")
	gradleContent := `plugins {
    id 'java'
}

group = 'com.example'
version = '1.0.0'`
	err = os.WriteFile(gradleFile, []byte(gradleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create build.gradle: %v", err)
	}

	tools = scanner.DetectBuildTools()
	if len(tools) < 2 {
		t.Errorf("Expected at least 2 build tools, got %d", len(tools))
	}

	// Should detect both maven and gradle
	foundMaven := false
	foundGradle := false
	for _, tool := range tools {
		if tool == "maven" {
			foundMaven = true
		}
		if tool == "gradle" {
			foundGradle = true
		}
	}

	if !foundMaven {
		t.Error("Should detect Maven")
	}
	if !foundGradle {
		t.Error("Should detect Gradle")
	}

	// Create Python requirements file
	reqFile := filepath.Join(tempDir, "requirements.txt")
	reqContent := `numpy==1.21.0
pandas>=1.3.0
requests`
	err = os.WriteFile(reqFile, []byte(reqContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	tools = scanner.DetectBuildTools()
	foundPip := false
	for _, tool := range tools {
		if tool == "pip" {
			foundPip = true
		}
	}

	if !foundPip {
		t.Error("Should detect pip")
	}
}

func TestBuildScanner_ScanDependencies(t *testing.T) {
	tempDir := t.TempDir()

	// Create a Maven project
	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>
    <name>Test Project</name>
    <description>A test project</description>

    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.2</version>
            <scope>test</scope>
        </dependency>
        <dependency>
            <groupId>com.google.guava</groupId>
            <artifactId>guava</artifactId>
            <version>31.1-jre</version>
        </dependency>
    </dependencies>
</project>`

	err := os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	env := NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	if len(dependencies) == 0 {
		t.Error("Expected at least one dependency root")
	}

	// Check the Maven project
	var mavenRoot *model.DependencyRoot
	for _, dep := range dependencies {
		if dep.BuildTool == "maven" {
			mavenRoot = &dep
			break
		}
	}

	if mavenRoot == nil {
		t.Error("Expected to find Maven dependency root")
	} else {
		if mavenRoot.ProjectName != "test-project" {
			t.Errorf("Expected project name 'test-project', got %s", mavenRoot.ProjectName)
		}
		if mavenRoot.ProjectVersion != "1.0.0" {
			t.Errorf("Expected project version '1.0.0', got %s", mavenRoot.ProjectVersion)
		}
		if len(mavenRoot.Dependencies) != 2 {
			t.Errorf("Expected 2 dependencies, got %d", len(mavenRoot.Dependencies))
		}
	}
}

func TestBuildScanner_ScanDependencies_EmptyProject(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	if len(dependencies) != 0 {
		t.Errorf("Expected no dependencies for empty project, got %d", len(dependencies))
	}
}

func TestDetectBuildToolFromFile(t *testing.T) {
	tests := []struct {
		fileName     string
		expectedTool string
		shouldDetect bool
	}{
		{"pom.xml", "maven", true},
		{"build.gradle", "gradle", true},
		{"build.gradle.kts", "gradle", true},
		{"package.json", "npm", true},
		{"requirements.txt", "pip", true},
		{"Pipfile", "pipenv", true},
		{"go.mod", "go", true},
		{"Cargo.toml", "cargo", true},
		{"composer.json", "composer", true},
		{"package-lock.json", "", false},
		{"yarn.lock", "", false},
		{"random.txt", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			tool, detected := detectBuildToolFromFile(tt.fileName)

			if detected != tt.shouldDetect {
				t.Errorf("detectBuildToolFromFile(%s) detection = %v, want %v",
					tt.fileName, detected, tt.shouldDetect)
			}

			if detected && tool != tt.expectedTool {
				t.Errorf("detectBuildToolFromFile(%s) tool = %s, want %s",
					tt.fileName, tool, tt.expectedTool)
			}
		})
	}
}

// Benchmark tests
func BenchmarkBuildScanner_DetectBuildTools(b *testing.B) {
	tempDir := b.TempDir()

	// Create multiple build files
	buildFiles := map[string]string{
		"pom.xml":          "<project></project>",
		"build.gradle":     "plugins { id 'java' }",
		"package.json":     `{"name": "test"}`,
		"requirements.txt": "requests==2.25.1",
		"go.mod":           "module test",
	}

	for fileName, content := range buildFiles {
		err := os.WriteFile(filepath.Join(tempDir, fileName), []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create %s: %v", fileName, err)
		}
	}

	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.DetectBuildTools()
	}
}
