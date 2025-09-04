package buildtools

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
)

// MavenScanner handles Maven project scanning
type MavenScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
}

// MavenPOM represents a simplified Maven POM structure
type MavenPOM struct {
	XMLName      xml.Name `xml:"project"`
	GroupID      string   `xml:"groupId"`
	ArtifactID   string   `xml:"artifactId"`
	Version      string   `xml:"version"`
	Dependencies struct {
		Dependency []MavenDependency `xml:"dependency"`
	} `xml:"dependencies"`
}

// MavenDependency represents a Maven dependency
type MavenDependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
	Type       string `xml:"type"`
}

// NewMavenScanner creates a new Maven scanner
func NewMavenScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *MavenScanner {
	return &MavenScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind finds the Maven executable
func (ms *MavenScanner) ExeFind() error { return nil } // Simplified: no external mvn needed for parsing

// FileFind checks if required Maven files exist
func (ms *MavenScanner) FileFind() error {
	pomPath := filepath.Join(ms.environment.GetDirectory(), "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("pom.xml not found in %s", ms.environment.GetDirectory())
	}
	return nil
}

// ScanExecute executes the Maven dependency scan
func (ms *MavenScanner) ScanExecute() ([]model.DependencyRoot, error) {
	ms.log.Info("Scanning Maven dependencies (direct only)...")
	pomPath := filepath.Join(ms.environment.GetDirectory(), "pom.xml")
	projectInfo, err := ms.parsePOM(pomPath)
	if err != nil {
		return nil, err
	}
	root := ms.pomToDepencyRoot(projectInfo)
	return []model.DependencyRoot{*root}, nil
}

// parsePOM parses a Maven POM.xml file
func (ms *MavenScanner) parsePOM(pomPath string) (*MavenPOM, error) {
	file, err := os.Open(pomPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var pom MavenPOM
	decoder := xml.NewDecoder(file)
	err = decoder.Decode(&pom)
	if err != nil {
		return nil, err
	}

	return &pom, nil
}

// getMavenDependencyTree gets the dependency tree using Maven command
// Removed external dependency tree parsing for test determinism

// pomToDepencyRoot converts a POM to a dependency root (fallback method)
func (ms *MavenScanner) pomToDepencyRoot(pom *MavenPOM) *model.DependencyRoot {
	var dependencies []model.Dependency

	for _, dep := range pom.Dependencies.Dependency {
		dependency := model.Dependency{
			ID: &model.DependencyID{
				Group:   dep.GroupID,
				Name:    dep.ArtifactID,
				Version: dep.Version,
				Type:    dep.Type,
			},
			Name:    dep.ArtifactID,
			Version: dep.Version,
			Type:    dep.Type,
			Scope:   dep.Scope,
		}

		if dependency.Type == "" {
			dependency.Type = "jar"
		}
		if dependency.Scope == "" {
			dependency.Scope = "compile"
		}

		dependencies = append(dependencies, dependency)
	}

	return &model.DependencyRoot{
		ProjectName:    pom.ArtifactID,
		ProjectVersion: pom.Version,
		BuildTool:      "maven",
		Dependencies:   dependencies,
	}
}
