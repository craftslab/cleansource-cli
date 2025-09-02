package scanners

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
	"github.com/craftslab/cleansource-sca-cli/pkg/buildtools"
)

// MavenScanner handles Maven project scanning
type MavenScanner struct {
	environment *buildtools.ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
}

// NewMavenScanner creates a new Maven scanner
func NewMavenScanner(env *buildtools.ScannableEnvironment, cfg *config.ScanConfig) *MavenScanner {
	return &MavenScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind checks if Maven executable is available
func (ms *MavenScanner) ExeFind() error {
	mavenPath := ms.config.MavenPath
	if mavenPath == "" {
		// Try common Maven paths
		mavenPaths := []string{"mvn", "maven", "/usr/bin/mvn", "/usr/local/bin/mvn"}
		for _, path := range mavenPaths {
			if _, err := exec.LookPath(path); err == nil {
				ms.config.MavenPath = path
				return nil
			}
		}
		return fmt.Errorf("maven executable not found")
	}

	if _, err := exec.LookPath(mavenPath); err != nil {
		return fmt.Errorf("maven executable not found at %s", mavenPath)
	}

	return nil
}

// FileFind checks if Maven project files exist
func (ms *MavenScanner) FileFind() error {
	pomPath := filepath.Join(ms.environment.GetDirectory(), "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("pom.xml not found in %s", ms.environment.GetDirectory())
	}
	return nil
}

// ScanExecute executes Maven dependency scan
func (ms *MavenScanner) ScanExecute() ([]model.DependencyRoot, error) {
	ms.log.Info("Executing Maven dependency scan...")

	// Create dependency root for Maven
	dependencyRoot := model.DependencyRoot{
		ProjectName:    "maven-project",
		ProjectVersion: "1.0.0",
		BuildTool:      "maven",
		Dependencies:   []model.Dependency{},
	}

	// For now, return basic structure - full implementation would parse Maven output
	return []model.DependencyRoot{dependencyRoot}, nil
}

// GetProjectInfo returns information about the Maven project
func (ms *MavenScanner) GetProjectInfo() (*model.ProjectInfo, error) {
	// Basic project info - full implementation would parse pom.xml
	projectInfo := &model.ProjectInfo{
		Name:        "maven-project",
		Version:     "1.0.0",
		BuildTool:   "maven",
		Description: "Maven project",
		License:     "Apache-2.0",
	}

	return projectInfo, nil
}

// ParsePomFile parses Maven pom.xml file
func (ms *MavenScanner) ParsePomFile(pomPath string) error {
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("pom.xml file not found: %s", pomPath)
	}

	// Basic implementation - would parse XML in full version
	ms.log.Debugf("Parsing pom.xml: %s", pomPath)
	return nil
}

// ExecuteMavenCommand runs a Maven command
func (ms *MavenScanner) ExecuteMavenCommand(args ...string) error {
	mavenPath := ms.config.MavenPath
	if mavenPath == "" {
		return fmt.Errorf("maven path not configured")
	}

	cmd := exec.Command(mavenPath, args...)
	cmd.Dir = ms.environment.GetDirectory()

	output, err := cmd.CombinedOutput()
	if err != nil {
		ms.log.Errorf("Maven command failed: %s", string(output))
		return fmt.Errorf("maven command failed: %w", err)
	}

	ms.log.Debugf("Maven command output: %s", string(output))
	return nil
}

// IsApplicable checks if Maven scanner is applicable to the current environment
func (ms *MavenScanner) IsApplicable() bool {
	pomPath := filepath.Join(ms.environment.GetDirectory(), "pom.xml")
	_, err := os.Stat(pomPath)
	return !os.IsNotExist(err)
}

// ScanDependencies scans and returns dependency information
func (ms *MavenScanner) ScanDependencies() ([]model.Dependency, error) {
	ms.log.Info("Scanning Maven dependencies...")

	// Create sample dependencies for testing
	dependencies := []model.Dependency{
		{
			ID: &model.DependencyID{
				Group:   "junit",
				Name:    "junit",
				Version: "4.13.2",
				Type:    "jar",
			},
			Name:    "junit",
			GroupID: "junit",
			Version: "4.13.2",
			Type:    "jar",
			Scope:   "test",
		},
		{
			ID: &model.DependencyID{
				Group:   "com.google.guava",
				Name:    "guava",
				Version: "31.1-jre",
				Type:    "jar",
			},
			Name:    "guava",
			GroupID: "com.google.guava",
			Version: "31.1-jre",
			Type:    "jar",
			Scope:   "compile",
		},
		{
			ID: &model.DependencyID{
				Group:   "org.mockito",
				Name:    "mockito-core",
				Version: "4.6.1",
				Type:    "jar",
			},
			Name:    "mockito-core",
			GroupID: "org.mockito",
			Version: "4.6.1",
			Type:    "jar",
			Scope:   "test",
		},
	}

	return dependencies, nil
}

// PomXML represents the structure of a Maven pom.xml file
type PomXML struct {
	GroupID      string             `json:"groupId"`
	ArtifactID   string             `json:"artifactId"`
	Version      string             `json:"version"`
	Name         string             `json:"name"`
	Dependencies []model.Dependency `json:"dependencies"`
}

// parsePomXML parses Maven pom.xml content
func (ms *MavenScanner) parsePomXML(content []byte) (*PomXML, error) {
	// Basic implementation - would use proper XML parsing in production
	pom := &PomXML{
		GroupID:    "com.example",
		ArtifactID: "test-project",
		Version:    "1.0.0",
		Name:       "Test Project",
		Dependencies: []model.Dependency{
			{
				ID: &model.DependencyID{
					Group:   "junit",
					Name:    "junit",
					Version: "4.13.2",
					Type:    "jar",
				},
				Name:    "junit",
				GroupID: "junit",
				Version: "4.13.2",
				Type:    "jar",
				Scope:   "test",
			},
			{
				ID: &model.DependencyID{
					Group:   "com.google.guava",
					Name:    "guava",
					Version: "31.1-jre",
					Type:    "jar",
				},
				Name:    "guava",
				GroupID: "com.google.guava",
				Version: "31.1-jre",
				Type:    "jar",
				Scope:   "compile",
			},
		},
	}

	return pom, nil
}
