package scanners

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
	"github.com/craftslab/cleansource-sca-cli/pkg/buildtools"
)

//nolint:staticcheck
func TestNewMavenScanner(t *testing.T) {
	tempDir := t.TempDir()
	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{
		MavenPath: "/usr/bin/mvn",
	}

	scanner := NewMavenScanner(env, cfg)

	if scanner == nil {
		t.Error("NewMavenScanner should not return nil")
	}

	if scanner.environment != env {
		t.Error("Scanner should store the provided environment")
	}

	if scanner.config != cfg {
		t.Error("Scanner should store the provided config")
	}
}

func TestMavenScanner_IsApplicable(t *testing.T) {
	tempDir := t.TempDir()

	// Test without pom.xml
	env := buildtools.NewScannableEnvironment(tempDir, "")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	if scanner.IsApplicable() {
		t.Error("Maven scanner should not be applicable without pom.xml")
	}

	// Create pom.xml
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

	if !scanner.IsApplicable() {
		t.Error("Maven scanner should be applicable with pom.xml")
	}
}

func TestMavenScanner_GetProjectInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create a detailed pom.xml
	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.2.3</version>
    <name>Test Project</name>
    <description>A Maven test project for unit testing</description>

    <licenses>
        <license>
            <name>Apache License, Version 2.0</name>
            <url>http://www.apache.org/licenses/LICENSE-2.0.txt</url>
        </license>
    </licenses>
</project>`

	err := os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	projectInfo, err := scanner.GetProjectInfo()
	if err != nil {
		t.Fatalf("GetProjectInfo failed: %v", err)
	}

	if projectInfo.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got %s", projectInfo.Name)
	}

	if projectInfo.Version != "1.2.3" {
		t.Errorf("Expected project version '1.2.3', got %s", projectInfo.Version)
	}

	if projectInfo.Description != "A Maven test project for unit testing" {
		t.Errorf("Expected description to match, got %s", projectInfo.Description)
	}

	if projectInfo.BuildTool != "maven" {
		t.Errorf("Expected build tool 'maven', got %s", projectInfo.BuildTool)
	}

	if projectInfo.License != "Apache License, Version 2.0" {
		t.Errorf("Expected license 'Apache License, Version 2.0', got %s", projectInfo.License)
	}
}

func TestMavenScanner_GetProjectInfo_MinimalPom(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal pom.xml
	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>minimal-project</artifactId>
    <version>1.0.0</version>
</project>`

	err := os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	projectInfo, err := scanner.GetProjectInfo()
	if err != nil {
		t.Fatalf("GetProjectInfo failed: %v", err)
	}

	if projectInfo.Name != "minimal-project" {
		t.Errorf("Expected project name 'minimal-project', got %s", projectInfo.Name)
	}

	if projectInfo.Version != "1.0.0" {
		t.Errorf("Expected project version '1.0.0', got %s", projectInfo.Version)
	}

	// Description should be empty for minimal pom
	if projectInfo.Description != "" {
		t.Errorf("Expected empty description, got %s", projectInfo.Description)
	}
}

func TestMavenScanner_ScanDependencies(t *testing.T) {
	tempDir := t.TempDir()

	// Create pom.xml with dependencies
	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>

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
        <dependency>
            <groupId>org.apache.commons</groupId>
            <artifactId>commons-lang3</artifactId>
            <version>3.12.0</version>
            <scope>compile</scope>
        </dependency>
    </dependencies>
</project>`

	err := os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	if len(dependencies) != 3 {
		t.Errorf("Expected 3 dependencies, got %d", len(dependencies))
	}

	// Check specific dependencies
	dependencyMap := make(map[string]*model.Dependency)
	for i, dep := range dependencies {
		dependencyMap[dep.Name] = &dependencies[i]
	}

	// Check JUnit dependency
	if junit, exists := dependencyMap["junit"]; exists {
		if junit.Version != "4.13.2" {
			t.Errorf("Expected JUnit version '4.13.2', got %s", junit.Version)
		}
		if junit.Scope != "test" {
			t.Errorf("Expected JUnit scope 'test', got %s", junit.Scope)
		}
		if junit.ID.Group != "junit" {
			t.Errorf("Expected JUnit group 'junit', got %s", junit.ID.Group)
		}
	} else {
		t.Error("JUnit dependency not found")
	}

	// Check Guava dependency
	if guava, exists := dependencyMap["guava"]; exists {
		if guava.Version != "31.1-jre" {
			t.Errorf("Expected Guava version '31.1-jre', got %s", guava.Version)
		}
		if guava.Scope != "" { // Default scope should be empty or "compile"
			t.Errorf("Expected Guava scope to be empty, got %s", guava.Scope)
		}
		if guava.ID.Group != "com.google.guava" {
			t.Errorf("Expected Guava group 'com.google.guava', got %s", guava.ID.Group)
		}
	} else {
		t.Error("Guava dependency not found")
	}
}

func TestMavenScanner_ScanDependencies_NoDependencies(t *testing.T) {
	tempDir := t.TempDir()

	// Create pom.xml without dependencies
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

	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	dependencies, err := scanner.ScanDependencies()
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	if len(dependencies) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(dependencies))
	}
}

func TestMavenScanner_ScanDependencies_InvalidPom(t *testing.T) {
	tempDir := t.TempDir()

	// Create invalid pom.xml
	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `This is not valid XML content`

	err := os.WriteFile(pomFile, []byte(pomContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	_, err = scanner.ScanDependencies()
	if err == nil {
		t.Error("ScanDependencies should return error for invalid XML")
	}
}

func TestMavenScanner_parsePomXML(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>
    <name>Test Project</name>
    <description>Test Description</description>

    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.2</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>`

	scanner := &MavenScanner{}
	pom, err := scanner.parsePomXML([]byte(xmlContent))
	if err != nil {
		t.Fatalf("parsePomXML failed: %v", err)
	}

	if pom.GroupID != "com.example" {
		t.Errorf("Expected GroupID 'com.example', got %s", pom.GroupID)
	}

	if pom.ArtifactID != "test-project" {
		t.Errorf("Expected ArtifactID 'test-project', got %s", pom.ArtifactID)
	}

	if pom.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got %s", pom.Version)
	}

	if len(pom.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(pom.Dependencies))
	}

	dep := pom.Dependencies[0]
	if dep.GroupID != "junit" {
		t.Errorf("Expected dependency GroupID 'junit', got %s", dep.GroupID)
	}
}

// Benchmark tests
func BenchmarkMavenScanner_GetProjectInfo(b *testing.B) {
	tempDir := b.TempDir()

	pomFile := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>
    <name>Test Project</name>
    <description>Test Description</description>
</project>`

	_ = os.WriteFile(pomFile, []byte(pomContent), 0644)

	env := buildtools.NewScannableEnvironment(tempDir, "pom.xml")
	cfg := &config.ScanConfig{}
	scanner := NewMavenScanner(env, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.GetProjectInfo()
	}
}
