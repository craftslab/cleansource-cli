package buildtools

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	mavenPath   string
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
func (ms *MavenScanner) ExeFind() error {
	// Use configured path if available
	if ms.config.MavenPath != "" {
		if _, err := os.Stat(ms.config.MavenPath); err == nil {
			ms.mavenPath = ms.config.MavenPath
			ms.log.Debugf("Using configured Maven path: %s", ms.mavenPath)
			return nil
		}
	}

	// Try to find mvn in PATH
	path, err := exec.LookPath("mvn")
	if err == nil {
		ms.mavenPath = path
		ms.log.Debugf("Found Maven in PATH: %s", ms.mavenPath)
		return nil
	}

	// Try to find mvnw (Maven wrapper) in project directory
	mvnwPath := filepath.Join(ms.environment.GetDirectory(), "mvnw")
	if _, err := os.Stat(mvnwPath); err == nil {
		ms.mavenPath = mvnwPath
		ms.log.Debugf("Using Maven wrapper: %s", ms.mavenPath)
		return nil
	}

	// Try Windows Maven wrapper
	mvnwCmdPath := filepath.Join(ms.environment.GetDirectory(), "mvnw.cmd")
	if _, err := os.Stat(mvnwCmdPath); err == nil {
		ms.mavenPath = mvnwCmdPath
		ms.log.Debugf("Using Maven wrapper (Windows): %s", ms.mavenPath)
		return nil
	}

	return fmt.Errorf("maven executable not found")
}

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
	ms.log.Info("Scanning Maven dependencies...")

	// First, try to parse POM.xml directly
	pomPath := filepath.Join(ms.environment.GetDirectory(), "pom.xml")
	projectInfo, err := ms.parsePOM(pomPath)
	if err != nil {
		ms.log.Warnf("Failed to parse POM.xml: %v", err)
	}

	// Use Maven dependency:tree command to get full dependency tree
	dependencies, err := ms.getMavenDependencyTree()
	if err != nil {
		ms.log.Warnf("Failed to get Maven dependency tree: %v", err)
		// Fallback to parsing POM.xml only
		if projectInfo != nil {
			return []model.DependencyRoot{*ms.pomToDepencyRoot(projectInfo)}, nil
		}
		return nil, err
	}

	// Create dependency root
	root := model.DependencyRoot{
		BuildTool:    "maven",
		Dependencies: dependencies,
	}

	if projectInfo != nil {
		root.ProjectName = projectInfo.ArtifactID
		root.ProjectVersion = projectInfo.Version
	}

	return []model.DependencyRoot{root}, nil
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
func (ms *MavenScanner) getMavenDependencyTree() ([]model.Dependency, error) {
	cmd := exec.Command(ms.mavenPath, "dependency:tree", "-DoutputType=text")
	cmd.Dir = ms.environment.GetDirectory()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute Maven dependency:tree: %w", err)
	}

	return ms.parseDependencyTreeOutput(string(output))
}

// parseDependencyTreeOutput parses the output from Maven dependency:tree command
func (ms *MavenScanner) parseDependencyTreeOutput(output string) ([]model.Dependency, error) {
	var dependencies []model.Dependency
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}

		// Parse lines like: [INFO] +- group:artifact:type:version:scope
		if strings.Contains(line, "+-") || strings.Contains(line, "\\-") {
			dep, err := ms.parseDependencyLine(line)
			if err == nil {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, nil
}

// parseDependencyLine parses a single dependency line from Maven output
func (ms *MavenScanner) parseDependencyLine(line string) (model.Dependency, error) {
	// Remove Maven tree indicators and [INFO] prefix
	cleaned := strings.ReplaceAll(line, "[INFO]", "")
	cleaned = strings.ReplaceAll(cleaned, "+-", "")
	cleaned = strings.ReplaceAll(cleaned, "\\-", "")
	cleaned = strings.ReplaceAll(cleaned, "|", "")
	cleaned = strings.TrimSpace(cleaned)

	// Parse format: groupId:artifactId:type:version:scope
	parts := strings.Split(cleaned, ":")
	if len(parts) < 3 {
		return model.Dependency{}, fmt.Errorf("invalid dependency format: %s", line)
	}

	groupId := parts[0]
	artifactId := parts[1]

	var version, scope, depType string
	if len(parts) >= 4 {
		if len(parts) == 4 {
			// format: groupId:artifactId:type:version or groupId:artifactId:version:scope
			if strings.Contains(parts[3], "compile") || strings.Contains(parts[3], "test") ||
				strings.Contains(parts[3], "provided") || strings.Contains(parts[3], "runtime") {
				version = parts[2]
				scope = parts[3]
				depType = "jar"
			} else {
				depType = parts[2]
				version = parts[3]
				scope = "compile"
			}
		} else {
			// format: groupId:artifactId:type:version:scope
			depType = parts[2]
			version = parts[3]
			if len(parts) > 4 {
				scope = parts[4]
			}
		}
	} else {
		version = parts[2]
		depType = "jar"
		scope = "compile"
	}

	return model.Dependency{
		ID: &model.DependencyID{
			Group:   groupId,
			Name:    artifactId,
			Version: version,
			Type:    depType,
		},
		Name:    artifactId,
		Version: version,
		Type:    depType,
		Scope:   scope,
	}, nil
}

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
