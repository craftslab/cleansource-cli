package model

import (
	"testing"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
)

func TestDependencyID(t *testing.T) {
	depID := &DependencyID{
		Group:   "com.example",
		Name:    "test-library",
		Version: "1.0.0",
		Type:    "jar",
	}

	if depID.Group != "com.example" {
		t.Errorf("Expected Group to be 'com.example', got %s", depID.Group)
	}
	if depID.Name != "test-library" {
		t.Errorf("Expected Name to be 'test-library', got %s", depID.Name)
	}
	if depID.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got %s", depID.Version)
	}
	if depID.Type != "jar" {
		t.Errorf("Expected Type to be 'jar', got %s", depID.Type)
	}
}

func TestDependency(t *testing.T) {
	dep := Dependency{
		ID: &DependencyID{
			Group:   "com.example",
			Name:    "test-lib",
			Version: "1.0.0",
			Type:    "jar",
		},
		Name:    "test-lib",
		Version: "1.0.0",
		Type:    "compile",
		Scope:   "compile",
		Children: []Dependency{
			{
				ID: &DependencyID{
					Group:   "com.example",
					Name:    "child-lib",
					Version: "0.5.0",
					Type:    "jar",
				},
				Name:    "child-lib",
				Version: "0.5.0",
				Type:    "compile",
			},
		},
	}

	if dep.Name != "test-lib" {
		t.Errorf("Expected Name to be 'test-lib', got %s", dep.Name)
	}
	if len(dep.Children) != 1 {
		t.Errorf("Expected 1 child dependency, got %d", len(dep.Children))
	}
	if dep.Children[0].Name != "child-lib" {
		t.Errorf("Expected child name to be 'child-lib', got %s", dep.Children[0].Name)
	}
}

func TestDependencyRoot(t *testing.T) {
	root := DependencyRoot{
		ProjectName:    "test-project",
		ProjectVersion: "2.0.0",
		BuildTool:      "maven",
		Dependencies: []Dependency{
			{
				Name:    "dep1",
				Version: "1.0.0",
				Type:    "compile",
			},
			{
				Name:    "dep2",
				Version: "2.0.0",
				Type:    "test",
			},
		},
	}

	if root.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName to be 'test-project', got %s", root.ProjectName)
	}
	if root.BuildTool != "maven" {
		t.Errorf("Expected BuildTool to be 'maven', got %s", root.BuildTool)
	}
	if len(root.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(root.Dependencies))
	}
}

func TestScanType(t *testing.T) {
	tests := []struct {
		scanType ScanType
		expected string
	}{
		{ScanTypeSource, "source"},
		{ScanTypeDocker, "docker"},
		{ScanTypeBinary, "binary"},
	}

	for _, tt := range tests {
		t.Run(string(tt.scanType), func(t *testing.T) {
			if string(tt.scanType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.scanType))
			}
		})
	}
}

func TestTaskType(t *testing.T) {
	if string(TaskTypeScan) != "scan" {
		t.Errorf("Expected TaskTypeScan to be 'scan', got %s", string(TaskTypeScan))
	}
}

func TestUploadData(t *testing.T) {
	cfg := &config.ScanConfig{
		TaskDir:   "/test/dir",
		ScanType:  "source",
		ServerURL: "https://example.com",
	}

	uploadData := &UploadData{
		WfpFile:     "/tmp/test.wfp",
		BuildFile:   "/tmp/build.json",
		ArchiveFile: "/tmp/archive.zip",
		Config:      cfg,
		DirSize:     1024000,
	}

	if uploadData.WfpFile != "/tmp/test.wfp" {
		t.Errorf("Expected WfpFile to be '/tmp/test.wfp', got %s", uploadData.WfpFile)
	}
	if uploadData.Config.TaskDir != "/test/dir" {
		t.Errorf("Expected Config.TaskDir to be '/test/dir', got %s", uploadData.Config.TaskDir)
	}
	if uploadData.DirSize != 1024000 {
		t.Errorf("Expected DirSize to be 1024000, got %d", uploadData.DirSize)
	}
}

func TestFilePathCollect(t *testing.T) {
	collect := &FilePathCollect{
		ProjectLicenseFiles: []string{
			"LICENSE",
			"COPYING",
			"LICENSE.txt",
		},
		SourceFiles: []string{
			"src/main/java/Main.java",
			"src/test/java/Test.java",
		},
		BinaryFiles: []string{
			"lib/library.jar",
			"bin/executable",
		},
	}

	if len(collect.ProjectLicenseFiles) != 3 {
		t.Errorf("Expected 3 license files, got %d", len(collect.ProjectLicenseFiles))
	}
	if len(collect.SourceFiles) != 2 {
		t.Errorf("Expected 2 source files, got %d", len(collect.SourceFiles))
	}
	if len(collect.BinaryFiles) != 2 {
		t.Errorf("Expected 2 binary files, got %d", len(collect.BinaryFiles))
	}

	if collect.ProjectLicenseFiles[0] != "LICENSE" {
		t.Errorf("Expected first license file to be 'LICENSE', got %s", collect.ProjectLicenseFiles[0])
	}
}

func TestFilterCondition(t *testing.T) {
	condition := &FilterCondition{
		Path:      "/src/main",
		Condition: "exclude",
		Value:     "*.test",
	}

	if condition.Path != "/src/main" {
		t.Errorf("Expected Path to be '/src/main', got %s", condition.Path)
	}
	if condition.Condition != "exclude" {
		t.Errorf("Expected Condition to be 'exclude', got %s", condition.Condition)
	}
	if condition.Value != "*.test" {
		t.Errorf("Expected Value to be '*.test', got %s", condition.Value)
	}
}

func TestBinaryFilterParam(t *testing.T) {
	param := &BinaryFilterParam{
		MixedBinaryScanFlag:         1,
		MixedBinaryScanFilePathList: []string{"/path1", "/path2"},
		BinaryScanList:              []string{"file1.jar", "file2.so"},
		BinaryRealScanList:          []string{"/real/path1", "/real/path2"},
	}

	if param.MixedBinaryScanFlag != 1 {
		t.Errorf("Expected MixedBinaryScanFlag to be 1, got %d", param.MixedBinaryScanFlag)
	}
	if len(param.MixedBinaryScanFilePathList) != 2 {
		t.Errorf("Expected 2 file paths, got %d", len(param.MixedBinaryScanFilePathList))
	}
	if len(param.BinaryScanList) != 2 {
		t.Errorf("Expected 2 binary scan files, got %d", len(param.BinaryScanList))
	}
}

func TestScanResult(t *testing.T) {
	result := &ScanResult{
		Success:    true,
		Message:    "Scan completed successfully",
		TaskID:     "task-12345",
		ResultFile: "/tmp/result.json",
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Message != "Scan completed successfully" {
		t.Errorf("Expected Message to be 'Scan completed successfully', got %s", result.Message)
	}
	if result.TaskID != "task-12345" {
		t.Errorf("Expected TaskID to be 'task-12345', got %s", result.TaskID)
	}
}

func TestExecutableInfo(t *testing.T) {
	exec := &ExecutableInfo{
		Name:    "maven",
		Path:    "/usr/bin/mvn",
		Version: "3.8.1",
	}

	if exec.Name != "maven" {
		t.Errorf("Expected Name to be 'maven', got %s", exec.Name)
	}
	if exec.Path != "/usr/bin/mvn" {
		t.Errorf("Expected Path to be '/usr/bin/mvn', got %s", exec.Path)
	}
	if exec.Version != "3.8.1" {
		t.Errorf("Expected Version to be '3.8.1', got %s", exec.Version)
	}
}

func TestProjectInfo(t *testing.T) {
	project := &ProjectInfo{
		Name:        "my-project",
		Version:     "1.0.0",
		Description: "A test project",
		License:     "MIT",
		BuildTool:   "maven",
	}

	if project.Name != "my-project" {
		t.Errorf("Expected Name to be 'my-project', got %s", project.Name)
	}
	if project.BuildTool != "maven" {
		t.Errorf("Expected BuildTool to be 'maven', got %s", project.BuildTool)
	}
	if project.License != "MIT" {
		t.Errorf("Expected License to be 'MIT', got %s", project.License)
	}
}

// Test new scanner types
func TestDependencyID_NewScannerTypes(t *testing.T) {
	tests := []struct {
		name     string
		depID    *DependencyID
		expected string
	}{
		{
			name: "Go Module",
			depID: &DependencyID{
				Group:   "",
				Name:    "github.com/gin-gonic/gin",
				Version: "v1.9.1",
				Type:    "go",
			},
			expected: "go",
		},
		{
			name: "NPM Package",
			depID: &DependencyID{
				Group:   "",
				Name:    "express",
				Version: "^4.18.2",
				Type:    "npm",
			},
			expected: "npm",
		},
		{
			name: "Gradle Dependency",
			depID: &DependencyID{
				Group:   "org.springframework",
				Name:    "spring-core",
				Version: "5.3.21",
				Type:    "gradle",
			},
			expected: "gradle",
		},
		{
			name: "Pipenv Package",
			depID: &DependencyID{
				Group:   "",
				Name:    "requests",
				Version: "2.25.1",
				Type:    "pipenv",
			},
			expected: "pipenv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.depID.Type != tt.expected {
				t.Errorf("Expected Type to be '%s', got %s", tt.expected, tt.depID.Type)
			}
		})
	}
}

func TestDependency_NewScannerTypes(t *testing.T) {
	tests := []struct {
		name     string
		dep      Dependency
		expected string
	}{
		{
			name: "Go Module Dependency",
			dep: Dependency{
				ID: &DependencyID{
					Group:   "",
					Name:    "github.com/gin-gonic/gin",
					Version: "v1.9.1",
					Type:    "go",
				},
				Name:    "github.com/gin-gonic/gin",
				Version: "v1.9.1",
				Type:    "go",
				Scope:   "runtime",
			},
			expected: "go",
		},
		{
			name: "NPM Development Dependency",
			dep: Dependency{
				ID: &DependencyID{
					Group:   "",
					Name:    "jest",
					Version: "^29.5.0",
					Type:    "npm",
				},
				Name:    "jest",
				Version: "^29.5.0",
				Type:    "npm",
				Scope:   "development",
			},
			expected: "npm",
		},
		{
			name: "Gradle Test Dependency",
			dep: Dependency{
				ID: &DependencyID{
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
			expected: "gradle",
		},
		{
			name: "Pipenv Runtime Dependency",
			dep: Dependency{
				ID: &DependencyID{
					Group:   "",
					Name:    "requests",
					Version: "2.25.1",
					Type:    "pipenv",
				},
				Name:    "requests",
				Version: "2.25.1",
				Type:    "pipenv",
				Scope:   "runtime",
			},
			expected: "pipenv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dep.Type != tt.expected {
				t.Errorf("Expected Type to be '%s', got %s", tt.expected, tt.dep.Type)
			}
		})
	}
}

func TestDependencyRoot_NewScannerTypes(t *testing.T) {
	tests := []struct {
		name        string
		root        DependencyRoot
		expectedTool string
	}{
		{
			name: "Go Project",
			root: DependencyRoot{
				ProjectName:    "test-go-project",
				ProjectVersion: "1.21",
				BuildTool:      "go",
				Dependencies: []Dependency{
					{
						Name:    "github.com/gin-gonic/gin",
						Version: "v1.9.1",
						Type:    "go",
						Scope:   "runtime",
					},
				},
			},
			expectedTool: "go",
		},
		{
			name: "NPM Project",
			root: DependencyRoot{
				ProjectName:    "test-npm-project",
				ProjectVersion: "1.0.0",
				BuildTool:      "npm",
				Dependencies: []Dependency{
					{
						Name:    "express",
						Version: "^4.18.2",
						Type:    "npm",
						Scope:   "runtime",
					},
				},
			},
			expectedTool: "npm",
		},
		{
			name: "Gradle Project",
			root: DependencyRoot{
				ProjectName:    "test-gradle-project",
				ProjectVersion: "1.0.0",
				BuildTool:      "gradle",
				Dependencies: []Dependency{
					{
						Name:    "org.springframework:spring-core",
						Version: "5.3.21",
						Type:    "gradle",
						Scope:   "runtime",
					},
				},
			},
			expectedTool: "gradle",
		},
		{
			name: "Pipenv Project",
			root: DependencyRoot{
				ProjectName:    "test-pipenv-project",
				ProjectVersion: "unknown",
				BuildTool:      "pipenv",
				Dependencies: []Dependency{
					{
						Name:    "requests",
						Version: "2.25.1",
						Type:    "pipenv",
						Scope:   "runtime",
					},
				},
			},
			expectedTool: "pipenv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.root.BuildTool != tt.expectedTool {
				t.Errorf("Expected BuildTool to be '%s', got %s", tt.expectedTool, tt.root.BuildTool)
			}
		})
	}
}

func TestDependency_Scopes(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{"Runtime scope", "runtime", true},
		{"Development scope", "development", true},
		{"Test scope", "test", true},
		{"Provided scope", "provided", true},
		{"Peer scope", "peer", true},
		{"Indirect scope", "indirect", true},
		{"Compile scope", "compile", true},
		{"Invalid scope", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := Dependency{
				Scope: tt.scope,
			}

			// Test that the scope is set correctly
			if dep.Scope != tt.scope {
				t.Errorf("Expected scope to be '%s', got %s", tt.scope, dep.Scope)
			}
		})
	}
}

func TestDependencyRoot_MultipleScanners(t *testing.T) {
	// Test a project that could have multiple build tools
	roots := []DependencyRoot{
		{
			ProjectName:    "go-module",
			ProjectVersion: "1.21",
			BuildTool:      "go",
			Dependencies: []Dependency{
				{Name: "github.com/gin-gonic/gin", Version: "v1.9.1", Type: "go", Scope: "runtime"},
			},
		},
		{
			ProjectName:    "npm-package",
			ProjectVersion: "1.0.0",
			BuildTool:      "npm",
			Dependencies: []Dependency{
				{Name: "express", Version: "^4.18.2", Type: "npm", Scope: "runtime"},
			},
		},
		{
			ProjectName:    "gradle-project",
			ProjectVersion: "1.0.0",
			BuildTool:      "gradle",
			Dependencies: []Dependency{
				{Name: "org.springframework:spring-core", Version: "5.3.21", Type: "gradle", Scope: "runtime"},
			},
		},
	}

	expectedTools := []string{"go", "npm", "gradle"}

	if len(roots) != len(expectedTools) {
		t.Errorf("Expected %d roots, got %d", len(expectedTools), len(roots))
	}

	for i, root := range roots {
		if root.BuildTool != expectedTools[i] {
			t.Errorf("Root %d: Expected BuildTool to be '%s', got %s", i, expectedTools[i], root.BuildTool)
		}
	}
}
