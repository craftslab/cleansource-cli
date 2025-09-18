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

	// Create Go module file
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module test-go-project
go 1.21`
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create NPM package file
	packageJsonFile := filepath.Join(tempDir, "package.json")
	packageJsonContent := `{
	"name": "test-npm-project",
	"version": "1.0.0"
}`
	err = os.WriteFile(packageJsonFile, []byte(packageJsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create Pipenv file
	pipfileFile := filepath.Join(tempDir, "Pipfile")
	pipfileContent := `[[source]]
url = "https://pypi.org/simple"
name = "pypi"

[packages]
requests = "*"`
	err = os.WriteFile(pipfileFile, []byte(pipfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create Pipfile: %v", err)
	}

	tools = scanner.DetectBuildTools()
	expectedTools := []string{"maven", "gradle", "pip", "go", "npm", "pipenv"}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d build tools, got %d", len(expectedTools), len(tools))
	}

	// Check that all expected tools are detected
	for _, expectedTool := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should detect %s", expectedTool)
		}
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

func TestBuildScanner_ScanDependencies_GoProject(t *testing.T) {
	tempDir := t.TempDir()

	// Create Go project
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module test-go-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/stretchr/testify v1.8.4
)`
	err := os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	// Should have at least one dependency root (Go project)
	if len(dependencies) == 0 {
		t.Error("Expected at least one dependency root for Go project")
	}

	// Check the Go project
	var goRoot *model.DependencyRoot
	for _, dep := range dependencies {
		if dep.BuildTool == "go" {
			goRoot = &dep
			break
		}
	}

	if goRoot == nil {
		t.Error("Expected to find Go dependency root")
	} else {
		if goRoot.ProjectName != "test-go-project" {
			t.Errorf("Expected project name 'test-go-project', got %s", goRoot.ProjectName)
		}
		if goRoot.ProjectVersion != "1.21" {
			t.Errorf("Expected Go version '1.21', got %s", goRoot.ProjectVersion)
		}
		// Note: Dependencies from go list command may not be available in test environment
		t.Logf("Found %d Go dependencies", len(goRoot.Dependencies))
	}
}

func TestBuildScanner_ScanDependencies_NpmProject(t *testing.T) {
	tempDir := t.TempDir()

	// Create NPM project
	packageJsonFile := filepath.Join(tempDir, "package.json")
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
	err := os.WriteFile(packageJsonFile, []byte(packageJsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	// Should have at least one dependency root (NPM project)
	if len(dependencies) == 0 {
		t.Error("Expected at least one dependency root for NPM project")
	}

	// Check the NPM project
	var npmRoot *model.DependencyRoot
	for _, dep := range dependencies {
		if dep.BuildTool == "npm" {
			npmRoot = &dep
			break
		}
	}

	if npmRoot == nil {
		t.Error("Expected to find NPM dependency root")
	} else {
		if npmRoot.ProjectName != "test-npm-project" {
			t.Errorf("Expected project name 'test-npm-project', got %s", npmRoot.ProjectName)
		}
		if npmRoot.ProjectVersion != "1.0.0" {
			t.Errorf("Expected project version '1.0.0', got %s", npmRoot.ProjectVersion)
		}
		if len(npmRoot.Dependencies) != 3 {
			t.Errorf("Expected 3 dependencies (2 deps + 1 dev), got %d", len(npmRoot.Dependencies))
		}
	}
}

func TestBuildScanner_ScanDependencies_GradleProject(t *testing.T) {
	tempDir := t.TempDir()

	// Create Gradle project
	gradleFile := filepath.Join(tempDir, "build.gradle")
	gradleContent := `plugins {
    id 'java'
}

group = 'com.example'
version = '1.0.0'

dependencies {
    implementation 'org.springframework:spring-core:5.3.21'
    testImplementation 'junit:junit:4.13.2'
}`
	err := os.WriteFile(gradleFile, []byte(gradleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create build.gradle: %v", err)
	}

	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	// Should have at least one dependency root (Gradle project)
	if len(dependencies) == 0 {
		t.Error("Expected at least one dependency root for Gradle project")
	}

	// Check the Gradle project
	var gradleRoot *model.DependencyRoot
	for _, dep := range dependencies {
		if dep.BuildTool == "gradle" {
			gradleRoot = &dep
			break
		}
	}

	if gradleRoot == nil {
		t.Error("Expected to find Gradle dependency root")
	} else {
		if gradleRoot.ProjectVersion != "1.0.0" {
			t.Errorf("Expected project version '1.0.0', got %s", gradleRoot.ProjectVersion)
		}
		if len(gradleRoot.Dependencies) != 2 {
			t.Errorf("Expected 2 dependencies, got %d", len(gradleRoot.Dependencies))
		}
	}
}

func TestBuildScanner_ScanDependencies_PipenvProject(t *testing.T) {
	tempDir := t.TempDir()

	// Create Pipenv project
	pipfileFile := filepath.Join(tempDir, "Pipfile")
	pipfileContent := `[[source]]
url = "https://pypi.org/simple"
name = "pypi"

[packages]
requests = "*"
flask = "*"

[dev-packages]
pytest = "*"`
	err := os.WriteFile(pipfileFile, []byte(pipfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create Pipfile: %v", err)
	}

	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	// Should have at least one dependency root (Pipenv project)
	if len(dependencies) == 0 {
		t.Error("Expected at least one dependency root for Pipenv project")
	}

	// Check the Pipenv project
	var pipenvRoot *model.DependencyRoot
	for _, dep := range dependencies {
		if dep.BuildTool == "pipenv" {
			pipenvRoot = &dep
			break
		}
	}

	if pipenvRoot == nil {
		t.Error("Expected to find Pipenv dependency root")
	} else {
		// Note: Dependencies from pipenv run pip freeze may not be available in test environment
		t.Logf("Found %d Pipenv dependencies", len(pipenvRoot.Dependencies))
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
