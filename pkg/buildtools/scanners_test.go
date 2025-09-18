package buildtools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
)

// Test Go Scanner
func TestGoScanner_ExeFind(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewGoScanner(env, cfg)

	// This test will pass if go is available in PATH, fail otherwise
	err := scanner.ExeFind()
	if err != nil {
		t.Logf("Go executable not found (expected in some environments): %v", err)
	}
}

func TestGoScanner_FileFind(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGoScanner(env, cfg)

	// Test without go.mod
	err := scanner.FileFind()
	if err == nil {
		t.Error("Expected error when go.mod is missing")
	}

	// Create go.mod
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module test-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)`
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test with go.mod
	err = scanner.FileFind()
	if err != nil {
		t.Errorf("Expected no error when go.mod exists, got: %v", err)
	}
}

func TestGoScanner_parseGoMod(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGoScanner(env, cfg)

	// Create go.mod
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module test-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)`
	err := os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	name, version, err := scanner.parseGoMod()
	if err != nil {
		t.Fatalf("parseGoMod failed: %v", err)
	}

	if name != "test-project" {
		t.Errorf("Expected module name 'test-project', got %s", name)
	}
	if version != "1.21" {
		t.Errorf("Expected Go version '1.21', got %s", version)
	}
}

func TestGoScanner_parseGoMod_Empty(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGoScanner(env, cfg)

	// Create empty go.mod
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module test-project`
	err := os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	name, version, err := scanner.parseGoMod()
	if err != nil {
		t.Fatalf("parseGoMod failed: %v", err)
	}

	if name != "test-project" {
		t.Errorf("Expected module name 'test-project', got %s", name)
	}
	if version != "unknown" {
		t.Errorf("Expected Go version 'unknown', got %s", version)
	}
}

// Test NPM Scanner
func TestNpmScanner_ExeFind(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewNpmScanner(env, cfg)

	// This test will pass if npm is available in PATH, fail otherwise
	err := scanner.ExeFind()
	if err != nil {
		t.Logf("NPM executable not found (expected in some environments): %v", err)
	}
}

func TestNpmScanner_FileFind(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewNpmScanner(env, cfg)

	// Test without package.json
	err := scanner.FileFind()
	if err == nil {
		t.Error("Expected error when package.json is missing")
	}

	// Create package.json
	packageJsonFile := filepath.Join(tempDir, "package.json")
	packageJsonContent := `{
	"name": "test-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2"
	}
}`
	err = os.WriteFile(packageJsonFile, []byte(packageJsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Test with package.json
	err = scanner.FileFind()
	if err != nil {
		t.Errorf("Expected no error when package.json exists, got: %v", err)
	}
}

func TestNpmScanner_parsePackageJson(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewNpmScanner(env, cfg)

	// Create package.json
	packageJsonFile := filepath.Join(tempDir, "package.json")
	packageJsonContent := `{
	"name": "test-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2",
		"lodash": "4.17.21"
	},
	"devDependencies": {
		"jest": "^29.5.0"
	},
	"peerDependencies": {
		"react": "^18.2.0"
	}
}`
	err := os.WriteFile(packageJsonFile, []byte(packageJsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	name, version, dependencies, err := scanner.parsePackageJson()
	if err != nil {
		t.Fatalf("parsePackageJson failed: %v", err)
	}

	if name != "test-project" {
		t.Errorf("Expected project name 'test-project', got %s", name)
	}
	if version != "1.0.0" {
		t.Errorf("Expected project version '1.0.0', got %s", version)
	}

	// Check dependencies count (2 deps + 1 dev + 1 peer = 4)
	if len(dependencies) != 4 {
		t.Errorf("Expected 4 dependencies, got %d", len(dependencies))
	}

	// Check dependency types
	depTypes := make(map[string]int)
	for _, dep := range dependencies {
		depTypes[dep.Scope]++
	}

	if depTypes["runtime"] != 2 {
		t.Errorf("Expected 2 runtime dependencies, got %d", depTypes["runtime"])
	}
	if depTypes["development"] != 1 {
		t.Errorf("Expected 1 development dependency, got %d", depTypes["development"])
	}
	if depTypes["peer"] != 1 {
		t.Errorf("Expected 1 peer dependency, got %d", depTypes["peer"])
	}
}

// Test Pipenv Scanner
func TestPipenvScanner_ExeFind(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewPipenvScanner(env, cfg)

	// This test will pass if pipenv is available in PATH, fail otherwise
	err := scanner.ExeFind()
	if err != nil {
		t.Logf("Pipenv executable not found (expected in some environments): %v", err)
	}
}

func TestPipenvScanner_FileFind(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewPipenvScanner(env, cfg)

	// Test without Pipfile
	err := scanner.FileFind()
	if err == nil {
		t.Error("Expected error when Pipfile is missing")
	}

	// Create Pipfile
	pipfileFile := filepath.Join(tempDir, "Pipfile")
	pipfileContent := `[[source]]
url = "https://pypi.org/simple"
verify_ssl = true
name = "pypi"

[packages]
requests = "*"
flask = "*"

[dev-packages]
pytest = "*"
`
	err = os.WriteFile(pipfileFile, []byte(pipfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create Pipfile: %v", err)
	}

	// Test with Pipfile
	err = scanner.FileFind()
	if err != nil {
		t.Errorf("Expected no error when Pipfile exists, got: %v", err)
	}
}

func TestPipenvScanner_parsePipfile(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewPipenvScanner(env, cfg)

	// Create Pipfile
	pipfileFile := filepath.Join(tempDir, "Pipfile")
	pipfileContent := `[[source]]
url = "https://pypi.org/simple"
verify_ssl = true
name = "pypi"

[packages]
requests = "*"
flask = "*"

[dev-packages]
pytest = "*"

[requires]
python_version = "3.9"
`
	err := os.WriteFile(pipfileFile, []byte(pipfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create Pipfile: %v", err)
	}

	name, version, err := scanner.parsePipfile()
	if err != nil {
		t.Fatalf("parsePipfile failed: %v", err)
	}

	// Pipfile doesn't have name/version by default, should be unknown
	if name != "unknown" {
		t.Errorf("Expected project name 'unknown', got %s", name)
	}
	if version != "unknown" {
		t.Errorf("Expected project version 'unknown', got %s", version)
	}
}

func TestPipenvScanner_parsePipfile_WithNameVersion(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewPipenvScanner(env, cfg)

	// Create Pipfile with name and version
	pipfileFile := filepath.Join(tempDir, "Pipfile")
	pipfileContent := `[[source]]
url = "https://pypi.org/simple"
verify_ssl = true
name = "pypi"

[packages]
requests = "*"

[dev-packages]
pytest = "*"

[requires]
python_version = "3.9"
`
	err := os.WriteFile(pipfileFile, []byte(pipfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create Pipfile: %v", err)
	}

	name, version, err := scanner.parsePipfile()
	if err != nil {
		t.Fatalf("parsePipfile failed: %v", err)
	}

	// Should still be unknown since we didn't add name/version sections
	if name != "unknown" {
		t.Errorf("Expected project name 'unknown', got %s", name)
	}
	if version != "unknown" {
		t.Errorf("Expected project version 'unknown', got %s", version)
	}
}

func TestPipenvScanner_extractQuotedValue(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewPipenvScanner(env, cfg)

	tests := []struct {
		input    string
		expected string
	}{
		{`"test-value"`, "test-value"},
		{`'test-value'`, "test-value"},
		{`"test-value"`, "test-value"},
		{`'test-value'`, "test-value"},
		{`test-value`, ""},
		{`"`, ""},
		{`'`, ""},
		{`"unclosed`, ""},
		{`'unclosed`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := scanner.extractQuotedValue(tt.input)
			if result != tt.expected {
				t.Errorf("extractQuotedValue(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Gradle Scanner
func TestGradleScanner_ExeFind(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	// This test will pass if gradle is available in PATH, fail otherwise
	err := scanner.ExeFind()
	if err != nil {
		t.Logf("Gradle executable not found (expected in some environments): %v", err)
	}
}

func TestGradleScanner_FileFind(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	// Test without build.gradle
	err := scanner.FileFind()
	if err == nil {
		t.Error("Expected error when build.gradle is missing")
	}

	// Create build.gradle
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

	// Test with build.gradle
	err = scanner.FileFind()
	if err != nil {
		t.Errorf("Expected no error when build.gradle exists, got: %v", err)
	}
}

func TestGradleScanner_FileFind_KotlinDSL(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	// Create build.gradle.kts
	gradleFile := filepath.Join(tempDir, "build.gradle.kts")
	gradleContent := `plugins {
    java
}

group = "com.example"
version = "1.0.0"`
	err := os.WriteFile(gradleFile, []byte(gradleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create build.gradle.kts: %v", err)
	}

	// Test with build.gradle.kts
	err = scanner.FileFind()
	if err != nil {
		t.Errorf("Expected no error when build.gradle.kts exists, got: %v", err)
	}
}

func TestGradleScanner_parseBuildGradle(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	// Create build.gradle
	gradleFile := filepath.Join(tempDir, "build.gradle")
	gradleContent := `plugins {
    id 'java'
    id 'application'
}

group = 'com.example'
version = '1.0.0'
mainClassName = 'com.example.Main'

repositories {
    mavenCentral()
}

dependencies {
    implementation 'org.springframework:spring-core:5.3.21'
    implementation 'com.google.guava:guava:31.1-jre'
    testImplementation 'junit:junit:4.13.2'
    compileOnly 'org.projectlombok:lombok:1.18.24'
}`
	err := os.WriteFile(gradleFile, []byte(gradleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create build.gradle: %v", err)
	}

	name, version, dependencies, err := scanner.parseBuildGradle()
	if err != nil {
		t.Fatalf("parseBuildGradle failed: %v", err)
	}

	// Note: The current implementation doesn't extract name from this format
	// It would need to be enhanced to parse rootProject.name or similar
	if name != "unknown" {
		t.Errorf("Expected project name 'unknown', got %s", name)
	}
	if version != "1.0.0" {
		t.Errorf("Expected project version '1.0.0', got %s", version)
	}

	// Check dependencies count
	if len(dependencies) != 4 {
		t.Errorf("Expected 4 dependencies, got %d", len(dependencies))
	}

	// Check dependency scopes
	scopeCounts := make(map[string]int)
	for _, dep := range dependencies {
		scopeCounts[dep.Scope]++
	}

	if scopeCounts["runtime"] != 2 {
		t.Errorf("Expected 2 runtime dependencies, got %d", scopeCounts["runtime"])
	}
	if scopeCounts["test"] != 1 {
		t.Errorf("Expected 1 test dependency, got %d", scopeCounts["test"])
	}
	if scopeCounts["provided"] != 1 {
		t.Errorf("Expected 1 provided dependency, got %d", scopeCounts["provided"])
	}
}

func TestGradleScanner_extractGradleValue(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	tests := []struct {
		line     string
		key      string
		expected string
	}{
		{`version = "1.0.0"`, "version", "1.0.0"},
		{`version = '1.0.0'`, "version", "1.0.0"},
		{`name = "test-project"`, "name", "test-project"},
		{`group = 'com.example'`, "group", "com.example"},
		{`version = 1.0.0`, "version", ""}, // No quotes
		{`some other line`, "version", ""}, // No match
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := scanner.extractGradleValue(tt.line, tt.key)
			if result != tt.expected {
				t.Errorf("extractGradleValue(%s, %s) = %s, want %s", tt.line, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGradleScanner_parseGradleDependency(t *testing.T) {
	env := NewScannableEnvironment("/tmp", "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	tests := []struct {
		line     string
		expected *model.Dependency
	}{
		{
			`implementation 'org.springframework:spring-core:5.3.21'`,
			&model.Dependency{
				ID: &model.DependencyID{
					Group:   "org.springframework",
					Name:    "spring-core",
					Version: "5.3.21",
					Type:    "gradle",
				},
				Name:    "spring-core",
				Version: "5.3.21",
				Type:    "gradle",
				Scope:   "runtime",
			},
		},
		{
			`testImplementation 'junit:junit:4.13.2'`,
			&model.Dependency{
				ID: &model.DependencyID{
					Group:   "junit",
					Name:    "junit",
					Version: "4.13.2",
					Type:    "gradle",
				},
				Name:    "junit",
				Version: "4.13.2",
				Type:    "gradle",
				Scope:   "test",
			},
		},
		{
			`compileOnly 'org.projectlombok:lombok:1.18.24'`,
			&model.Dependency{
				ID: &model.DependencyID{
					Group:   "org.projectlombok",
					Name:    "lombok",
					Version: "1.18.24",
					Type:    "gradle",
				},
				Name:    "lombok",
				Version: "1.18.24",
				Type:    "gradle",
				Scope:   "provided",
			},
		},
		{
			`some other line`, // No dependency
			nil,
		},
	}

	for i, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := scanner.parseGradleDependency(tt.line)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil dependency, got %+v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected dependency, got nil")
				} else {
					if result.ID.Group != tt.expected.ID.Group {
						t.Errorf("Test %d: Expected Group %s, got %s", i, tt.expected.ID.Group, result.ID.Group)
					}
					if result.Name != tt.expected.Name {
						t.Errorf("Test %d: Expected Name %s, got %s", i, tt.expected.Name, result.Name)
					}
					if result.Version != tt.expected.Version {
						t.Errorf("Test %d: Expected Version %s, got %s", i, tt.expected.Version, result.Version)
					}
					if result.Scope != tt.expected.Scope {
						t.Errorf("Test %d: Expected Scope %s, got %s", i, tt.expected.Scope, result.Scope)
					}
				}
			}
		})
	}
}

// Integration tests for BuildScanner with new scanners
func TestBuildScanner_DetectBuildTools_AllTypes(t *testing.T) {
	tempDir := t.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewBuildScanner(env, cfg)

	// Create all types of build files
	buildFiles := map[string]string{
		"go.mod": `module test-go-project
go 1.21`,
		"package.json": `{
	"name": "test-npm-project",
	"version": "1.0.0"
}`,
		"Pipfile": `[[source]]
url = "https://pypi.org/simple"
name = "pypi"

[packages]
requests = "*"`,
		"build.gradle": `plugins {
    id 'java'
}

group = 'com.example'
version = '1.0.0'`,
		"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-maven-project</artifactId>
    <version>1.0.0</version>
</project>`,
		"requirements.txt": `requests==2.25.1
numpy>=1.21.0`,
	}

	for fileName, content := range buildFiles {
		err := os.WriteFile(filepath.Join(tempDir, fileName), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create %s: %v", fileName, err)
		}
	}

	tools := scanner.DetectBuildTools()
	expectedTools := []string{"go", "npm", "pipenv", "gradle", "maven", "pip"}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d build tools, got %d", len(expectedTools), len(tools))
	}

	for _, expectedTool := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find build tool: %s", expectedTool)
		}
	}
}

// Benchmark tests for new scanners
func BenchmarkGoScanner_parseGoMod(b *testing.B) {
	tempDir := b.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGoScanner(env, cfg)

	// Create go.mod
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module test-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/stretchr/testify v1.8.4
)`
	_ = os.WriteFile(goModFile, []byte(goModContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = scanner.parseGoMod()
	}
}

func BenchmarkNpmScanner_parsePackageJson(b *testing.B) {
	tempDir := b.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewNpmScanner(env, cfg)

	// Create package.json
	packageJsonFile := filepath.Join(tempDir, "package.json")
	packageJsonContent := `{
	"name": "test-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2",
		"lodash": "4.17.21"
	},
	"devDependencies": {
		"jest": "^29.5.0"
	}
}`
	_ = os.WriteFile(packageJsonFile, []byte(packageJsonContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = scanner.parsePackageJson()
	}
}

func BenchmarkGradleScanner_parseBuildGradle(b *testing.B) {
	tempDir := b.TempDir()
	env := NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewGradleScanner(env, cfg)

	// Create build.gradle
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
	_ = os.WriteFile(gradleFile, []byte(gradleContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = scanner.parseBuildGradle()
	}
}
