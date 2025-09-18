package buildtools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
)

// GoScanner implements scanning for Go projects
type GoScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
}

// GradleScanner handles Gradle project scanning
type GradleScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
}

// NewGradleScanner creates a new Gradle scanner
func NewGradleScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *GradleScanner {
	return &GradleScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind finds the Gradle executable
func (gs *GradleScanner) ExeFind() error {
	// Try to find gradle executable in PATH
	gradleCandidates := []string{"gradle", "gradle.bat", "./gradlew", "./gradlew.bat"}
	for _, candidate := range gradleCandidates {
		if candidate == "./gradlew" || candidate == "./gradlew.bat" {
			// Check for gradle wrapper in project directory
			wrapperPath := filepath.Join(gs.environment.GetDirectory(), candidate)
			if _, err := os.Stat(wrapperPath); err == nil {
				gs.log.Debugf("Found gradle wrapper: %s", wrapperPath)
				return nil
			}
		} else {
			// Check for gradle in PATH
			if path, err := exec.LookPath(candidate); err == nil {
				gs.log.Debugf("Found gradle executable: %s", path)
				return nil
			}
		}
	}
	return fmt.Errorf("gradle executable not found in PATH or as wrapper")
}

// FileFind checks if required Gradle files exist
func (gs *GradleScanner) FileFind() error {
	buildGradle := filepath.Join(gs.environment.GetDirectory(), "build.gradle")
	buildGradleKts := filepath.Join(gs.environment.GetDirectory(), "build.gradle.kts")

	if _, err := os.Stat(buildGradle); err == nil {
		return nil
	}
	if _, err := os.Stat(buildGradleKts); err == nil {
		return nil
	}

	return fmt.Errorf("build.gradle or build.gradle.kts not found")
}

// ScanExecute executes the Gradle dependency scan
func (gs *GradleScanner) ScanExecute() ([]model.DependencyRoot, error) {
	gs.log.Info("Scanning Gradle dependencies...")

	// Parse build.gradle for project info and dependencies
	projectName, projectVersion, dependencies, err := gs.parseBuildGradle()
	if err != nil {
		gs.log.Warnf("Failed to parse build.gradle: %v", err)
		projectName = "unknown"
		projectVersion = "unknown"
		dependencies = []model.Dependency{}
	}

	root := model.DependencyRoot{
		ProjectName:    projectName,
		ProjectVersion: projectVersion,
		BuildTool:      "gradle",
		Dependencies:   dependencies,
	}

	return []model.DependencyRoot{root}, nil
}

// PipenvScanner handles Python pipenv project scanning
type PipenvScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
}

// NewPipenvScanner creates a new pipenv scanner
func NewPipenvScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *PipenvScanner {
	return &PipenvScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind finds the pipenv executable
func (ps *PipenvScanner) ExeFind() error {
	// Try to find pipenv executable in PATH
	pipenvCandidates := []string{"pipenv", "pipenv.exe"}
	for _, candidate := range pipenvCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			ps.log.Debugf("Found pipenv executable: %s", path)
			return nil
		}
	}
	return fmt.Errorf("pipenv executable not found in PATH")
}

// FileFind checks if required pipenv files exist
func (ps *PipenvScanner) FileFind() error {
	pipfile := filepath.Join(ps.environment.GetDirectory(), "Pipfile")
	if _, err := os.Stat(pipfile); os.IsNotExist(err) {
		return fmt.Errorf("pipfile not found")
	}
	return nil
}

// ScanExecute executes the pipenv dependency scan
func (ps *PipenvScanner) ScanExecute() ([]model.DependencyRoot, error) {
	ps.log.Info("Scanning pipenv dependencies...")

	// Parse Pipfile for project info
	projectName, projectVersion, err := ps.parsePipfile()
	if err != nil {
		ps.log.Warnf("Failed to parse Pipfile: %v", err)
		projectName = "unknown"
		projectVersion = "unknown"
	}

	// Get dependencies using pipenv commands
	dependencies, err := ps.getPipenvDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to get pipenv dependencies: %w", err)
	}

	root := model.DependencyRoot{
		ProjectName:    projectName,
		ProjectVersion: projectVersion,
		BuildTool:      "pipenv",
		Dependencies:   dependencies,
	}

	return []model.DependencyRoot{root}, nil
}

// NpmScanner handles Node.js npm project scanning
type NpmScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
}

// NewNpmScanner creates a new npm scanner
func NewNpmScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *NpmScanner {
	return &NpmScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind finds the npm executable
func (ns *NpmScanner) ExeFind() error {
	// Try to find npm executable in PATH
	npmCandidates := []string{"npm", "npm.cmd"}
	for _, candidate := range npmCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			ns.log.Debugf("Found npm executable: %s", path)
			return nil
		}
	}
	return fmt.Errorf("npm executable not found in PATH")
}

// FileFind checks if required npm files exist
func (ns *NpmScanner) FileFind() error {
	packageJson := filepath.Join(ns.environment.GetDirectory(), "package.json")
	if _, err := os.Stat(packageJson); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found")
	}
	return nil
}

// ScanExecute executes the npm dependency scan
func (ns *NpmScanner) ScanExecute() ([]model.DependencyRoot, error) {
	ns.log.Info("Scanning npm dependencies...")

	// Parse package.json for project info and dependencies
	projectName, projectVersion, dependencies, err := ns.parsePackageJson()
	if err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	root := model.DependencyRoot{
		ProjectName:    projectName,
		ProjectVersion: projectVersion,
		BuildTool:      "npm",
		Dependencies:   dependencies,
	}

	return []model.DependencyRoot{root}, nil
}

// NewGoScanner creates a new Go scanner
func NewGoScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *GoScanner {
	return &GoScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind finds the Go executable
func (gs *GoScanner) ExeFind() error {
	// Try to find go executable in PATH
	goCandidates := []string{"go"}
	for _, candidate := range goCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			gs.log.Debugf("Found go executable: %s", path)
			return nil
		}
	}
	return fmt.Errorf("go executable not found in PATH")
}

// FileFind checks if required Go files exist
func (gs *GoScanner) FileFind() error {
	goMod := filepath.Join(gs.environment.GetDirectory(), "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found")
	}
	return nil
}

// ScanExecute executes the Go dependency scan
func (gs *GoScanner) ScanExecute() ([]model.DependencyRoot, error) {
	gs.log.Info("Scanning Go modules dependencies...")

	// Get project info from go.mod
	projectName, projectVersion, err := gs.parseGoMod()
	if err != nil {
		gs.log.Warnf("Failed to parse go.mod: %v", err)
		projectName = "unknown"
		projectVersion = "unknown"
	}

	// Get dependencies using go list
	dependencies, err := gs.getGoDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to get Go dependencies: %w", err)
	}

	root := model.DependencyRoot{
		ProjectName:    projectName,
		ProjectVersion: projectVersion,
		BuildTool:      "go",
		Dependencies:   dependencies,
	}

	return []model.DependencyRoot{root}, nil
}

// parseGoMod parses go.mod file to extract module name and version
func (gs *GoScanner) parseGoMod() (string, string, error) {
	goModPath := filepath.Join(gs.environment.GetDirectory(), "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = file.Close() }()

	var moduleName, goVersion string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
		} else if strings.HasPrefix(line, "go ") {
			goVersion = strings.TrimSpace(strings.TrimPrefix(line, "go"))
		}
	}

	if moduleName == "" {
		moduleName = "unknown"
	}
	if goVersion == "" {
		goVersion = "unknown"
	}

	return moduleName, goVersion, scanner.Err()
}

// getGoDependencies gets Go module dependencies using go list command
func (gs *GoScanner) getGoDependencies() ([]model.Dependency, error) {
	// Use go list -m -json all to get all dependencies
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = gs.environment.GetDirectory()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run go list: %w", err)
	}

	var dependencies []model.Dependency
	lines := strings.Split(string(output), "\n")

	var jsonBuffer strings.Builder
	for _, line := range lines {
		jsonBuffer.WriteString(line)
		if strings.TrimSpace(line) == "}" {
			// Parse the complete JSON object
			var moduleInfo struct {
				Path     string `json:"Path"`
				Version  string `json:"Version"`
				Main     bool   `json:"Main"`
				Indirect bool   `json:"Indirect"`
			}

			if err := json.Unmarshal([]byte(jsonBuffer.String()), &moduleInfo); err == nil {
				// Skip the main module
				if !moduleInfo.Main && moduleInfo.Path != "" {
					dependency := model.Dependency{
						ID: &model.DependencyID{
							Group:   "",
							Name:    moduleInfo.Path,
							Version: moduleInfo.Version,
							Type:    "go",
						},
						Name:    moduleInfo.Path,
						Version: moduleInfo.Version,
						Type:    "go",
						Scope:   "runtime",
					}

					if moduleInfo.Indirect {
						dependency.Scope = "indirect"
					}

					dependencies = append(dependencies, dependency)
				}
			}
			jsonBuffer.Reset()
		}
	}

	return dependencies, nil
}

// parsePackageJson parses package.json file to extract project info and dependencies
func (ns *NpmScanner) parsePackageJson() (string, string, []model.Dependency, error) {
	packageJsonPath := filepath.Join(ns.environment.GetDirectory(), "package.json")
	file, err := os.Open(packageJsonPath)
	if err != nil {
		return "", "", nil, err
	}
	defer func() { _ = file.Close() }()

	var packageInfo struct {
		Name         string            `json:"name"`
		Version      string            `json:"version"`
		Dependencies map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		PeerDependencies map[string]string `json:"peerDependencies"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&packageInfo); err != nil {
		return "", "", nil, err
	}

	projectName := packageInfo.Name
	if projectName == "" {
		projectName = "unknown"
	}

	projectVersion := packageInfo.Version
	if projectVersion == "" {
		projectVersion = "unknown"
	}

	var dependencies []model.Dependency

	// Parse dependencies
	for name, version := range packageInfo.Dependencies {
		dependency := model.Dependency{
			ID: &model.DependencyID{
				Group:   "",
				Name:    name,
				Version: version,
				Type:    "npm",
			},
			Name:    name,
			Version: version,
			Type:    "npm",
			Scope:   "runtime",
		}
		dependencies = append(dependencies, dependency)
	}

	// Parse devDependencies
	for name, version := range packageInfo.DevDependencies {
		dependency := model.Dependency{
			ID: &model.DependencyID{
				Group:   "",
				Name:    name,
				Version: version,
				Type:    "npm",
			},
			Name:    name,
			Version: version,
			Type:    "npm",
			Scope:   "development",
		}
		dependencies = append(dependencies, dependency)
	}

	// Parse peerDependencies
	for name, version := range packageInfo.PeerDependencies {
		dependency := model.Dependency{
			ID: &model.DependencyID{
				Group:   "",
				Name:    name,
				Version: version,
				Type:    "npm",
			},
			Name:    name,
			Version: version,
			Type:    "npm",
			Scope:   "peer",
		}
		dependencies = append(dependencies, dependency)
	}

	return projectName, projectVersion, dependencies, nil
}

// parsePipfile parses Pipfile to extract project name and version
func (ps *PipenvScanner) parsePipfile() (string, string, error) {
	pipfilePath := filepath.Join(ps.environment.GetDirectory(), "Pipfile")
	file, err := os.Open(pipfilePath)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = file.Close() }()

	var projectName, projectVersion string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "name = ") {
			projectName = ps.extractQuotedValue(strings.TrimSpace(strings.TrimPrefix(line, "name =")))
		} else if strings.HasPrefix(line, "version = ") {
			projectVersion = ps.extractQuotedValue(strings.TrimSpace(strings.TrimPrefix(line, "version =")))
		}
	}

	if projectName == "" {
		projectName = "unknown"
	}
	if projectVersion == "" {
		projectVersion = "unknown"
	}

	return projectName, projectVersion, scanner.Err()
}

// getPipenvDependencies gets pipenv dependencies using pipenv commands
func (ps *PipenvScanner) getPipenvDependencies() ([]model.Dependency, error) {
	// Use pipenv run pip freeze to get installed packages
	cmd := exec.Command("pipenv", "run", "pip", "freeze")
	cmd.Dir = ps.environment.GetDirectory()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run pipenv run pip freeze: %w", err)
	}

	var dependencies []model.Dependency
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "==")
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])

		dependencies = append(dependencies, model.Dependency{
			ID: &model.DependencyID{
				Group:   "",
				Name:    name,
				Version: version,
				Type:    "pipenv",
			},
			Name:    name,
			Version: version,
			Type:    "pipenv",
			Scope:   "runtime",
		})
	}

	return dependencies, nil
}

// extractQuotedValue extracts a quoted value from a string
func (ps *PipenvScanner) extractQuotedValue(expr string) string {
	expr = strings.TrimSpace(expr)

	for _, quote := range []string{"\"", "'"} {
		if strings.HasPrefix(expr, quote) {
			if end := strings.Index(expr[1:], quote); end != -1 {
				return expr[1 : end+1]
			}
		}
	}

	return ""
}

// parseBuildGradle parses build.gradle file to extract project info and dependencies
func (gs *GradleScanner) parseBuildGradle() (string, string, []model.Dependency, error) {
	// Try build.gradle first, then build.gradle.kts
	buildGradlePath := filepath.Join(gs.environment.GetDirectory(), "build.gradle")
	buildGradleKtsPath := filepath.Join(gs.environment.GetDirectory(), "build.gradle.kts")

	var filePath string
	if _, err := os.Stat(buildGradlePath); err == nil {
		filePath = buildGradlePath
	} else if _, err := os.Stat(buildGradleKtsPath); err == nil {
		filePath = buildGradleKtsPath
	} else {
		return "", "", nil, fmt.Errorf("no build.gradle or build.gradle.kts found")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", "", nil, err
	}
	defer func() { _ = file.Close() }()

	var projectName, projectVersion string
	var dependencies []model.Dependency
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse project name
		if strings.Contains(line, "rootProject.name") || strings.Contains(line, "name =") {
			if name := gs.extractGradleValue(line, "name"); name != "" {
				projectName = name
			}
		}

		// Parse project version
		if strings.Contains(line, "version") && !strings.Contains(line, "dependencies") {
			if version := gs.extractGradleValue(line, "version"); version != "" {
				projectVersion = version
			}
		}

		// Parse dependencies
		if strings.Contains(line, "implementation") || strings.Contains(line, "compile") ||
		   strings.Contains(line, "api") || strings.Contains(line, "testImplementation") {
			if dep := gs.parseGradleDependency(line); dep != nil {
				dependencies = append(dependencies, *dep)
			}
		}
	}

	if projectName == "" {
		projectName = "unknown"
	}
	if projectVersion == "" {
		projectVersion = "unknown"
	}

	return projectName, projectVersion, dependencies, scanner.Err()
}

// extractGradleValue extracts a value from a gradle line
func (gs *GradleScanner) extractGradleValue(line, key string) string {
	// Look for patterns like: name = "value" or name 'value'
	patterns := []string{
		key + `\s*=\s*["']([^"']+)["']`,
		key + `\s*=\s*"([^"]+)"`,
		key + `\s*=\s*'([^']+)'`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// parseGradleDependency parses a gradle dependency line
func (gs *GradleScanner) parseGradleDependency(line string) *model.Dependency {
	// Look for patterns like: implementation 'group:artifact:version'
	// or implementation("group:artifact:version")
	re := regexp.MustCompile(`['"]([^:]+):([^:]+):([^"']+)['"]`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 4 {
		group := matches[1]
		artifact := matches[2]
		version := matches[3]

		scope := "runtime"
		if strings.Contains(line, "testImplementation") || strings.Contains(line, "testCompile") {
			scope = "test"
		} else if strings.Contains(line, "compileOnly") {
			scope = "provided"
		}

		return &model.Dependency{
			ID: &model.DependencyID{
				Group:   group,
				Name:    artifact,
				Version: version,
				Type:    "gradle",
			},
			Name:    artifact,
			Version: version,
			Type:    "gradle",
			Scope:   scope,
		}
	}

	return nil
}
