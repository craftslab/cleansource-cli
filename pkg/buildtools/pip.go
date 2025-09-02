package buildtools

import (
	"bufio"
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

// PipScanner handles Python pip project scanning
type PipScanner struct {
	environment *ScannableEnvironment
	config      *config.ScanConfig
	log         *logrus.Logger
	pipPath     string
	pythonPath  string
}

// NewPipScanner creates a new pip scanner
func NewPipScanner(env *ScannableEnvironment, cfg *config.ScanConfig) *PipScanner {
	return &PipScanner{
		environment: env,
		config:      cfg,
		log:         logger.GetLogger(),
	}
}

// ExeFind finds the pip and python executables
func (ps *PipScanner) ExeFind() error {
	// Find Python executable
	if ps.config.PipPath != "" {
		// Extract python path from pip path if configured
		ps.pythonPath = strings.Replace(ps.config.PipPath, "pip", "python", 1)
	} else {
		// Try to find python in PATH
		pythonCandidates := []string{"python3", "python", "py"}
		for _, candidate := range pythonCandidates {
			if path, err := exec.LookPath(candidate); err == nil {
				ps.pythonPath = path
				break
			}
		}
		if ps.pythonPath == "" {
			return fmt.Errorf("python executable not found")
		}
	}

	// Find pip executable
	if ps.config.PipPath != "" {
		if _, err := os.Stat(ps.config.PipPath); err == nil {
			ps.pipPath = ps.config.PipPath
			ps.log.Debugf("Using configured pip path: %s", ps.pipPath)
			return nil
		}
	}

	// Try to find pip in PATH
	pipCandidates := []string{"pip3", "pip"}
	for _, candidate := range pipCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			ps.pipPath = path
			ps.log.Debugf("Found pip in PATH: %s", ps.pipPath)
			return nil
		}
	}

	// Try using python -m pip
	cmd := exec.Command(ps.pythonPath, "-m", "pip", "--version")
	if err := cmd.Run(); err == nil {
		ps.pipPath = ps.pythonPath
		ps.log.Debug("Using python -m pip")
		return nil
	}

	return fmt.Errorf("pip executable not found")
}

// FileFind checks if required pip files exist
func (ps *PipScanner) FileFind() error {
	projectDir := ps.environment.GetDirectory()

	// Check for requirements.txt
	reqPath := filepath.Join(projectDir, "requirements.txt")
	if ps.config.PipRequirementsPath != "" {
		reqPath = ps.config.PipRequirementsPath
	}

	if _, err := os.Stat(reqPath); err == nil {
		return nil
	}

	// Check for setup.py
	setupPath := filepath.Join(projectDir, "setup.py")
	if _, err := os.Stat(setupPath); err == nil {
		return nil
	}

	// Check for pyproject.toml
	pyprojectPath := filepath.Join(projectDir, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		return nil
	}

	return fmt.Errorf("no pip requirement files found (requirements.txt, setup.py, pyproject.toml)")
}

// ScanExecute executes the pip dependency scan
func (ps *PipScanner) ScanExecute() ([]model.DependencyRoot, error) {
	ps.log.Info("Scanning pip dependencies...")

	var dependencies []model.Dependency
	var projectName = "unknown"
	var projectVersion = "unknown"

	// Try to parse requirements.txt first
	reqPath := filepath.Join(ps.environment.GetDirectory(), "requirements.txt")
	if ps.config.PipRequirementsPath != "" {
		reqPath = ps.config.PipRequirementsPath
	}

	if _, err := os.Stat(reqPath); err == nil {
		reqDeps, err := ps.parseRequirementsFile(reqPath)
		if err == nil {
			dependencies = append(dependencies, reqDeps...)
		} else {
			ps.log.Warnf("Failed to parse requirements.txt: %v", err)
		}
	}

	// Try to get installed packages using pip list
	installedDeps, err := ps.getInstalledPackages()
	if err == nil {
		// Merge with requirements, preferring requirements versions
		dependencies = ps.mergeDependencies(dependencies, installedDeps)
	} else {
		ps.log.Warnf("Failed to get installed packages: %v", err)
	}

	// Try to get project info from setup.py
	setupPath := filepath.Join(ps.environment.GetDirectory(), "setup.py")
	if _, err := os.Stat(setupPath); err == nil {
		name, version := ps.parseSetupPy(setupPath)
		if name != "" {
			projectName = name
		}
		if version != "" {
			projectVersion = version
		}
	}

	root := model.DependencyRoot{
		ProjectName:    projectName,
		ProjectVersion: projectVersion,
		BuildTool:      "pip",
		Dependencies:   dependencies,
	}

	return []model.DependencyRoot{root}, nil
}

// parseRequirementsFile parses a requirements.txt file
func (ps *PipScanner) parseRequirementsFile(reqPath string) ([]model.Dependency, error) {
	file, err := os.Open(reqPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var dependencies []model.Dependency
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip lines starting with -r, -e, --find-links, etc.
		if strings.HasPrefix(line, "-") {
			continue
		}

		dep, err := ps.parseRequirementLine(line)
		if err == nil {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, scanner.Err()
}

// parseRequirementLine parses a single requirement line
func (ps *PipScanner) parseRequirementLine(line string) (model.Dependency, error) {
	// Handle various formats:
	// package==1.0.0
	// package>=1.0.0
	// package~=1.0
	// package

	var name, version string

	// Split on version specifiers
	for _, sep := range []string{"==", ">=", "<=", "~=", ">", "<", "!="} {
		if strings.Contains(line, sep) {
			parts := strings.SplitN(line, sep, 2)
			name = strings.TrimSpace(parts[0])
			version = strings.TrimSpace(parts[1])
			break
		}
	}

	if name == "" {
		name = strings.TrimSpace(line)
		version = "unknown"
	}

	// Remove any extras (e.g., requests[security])
	if idx := strings.Index(name, "["); idx != -1 {
		name = name[:idx]
	}

	return model.Dependency{
		ID: &model.DependencyID{
			Group:   "",
			Name:    name,
			Version: version,
			Type:    "pip",
		},
		Name:    name,
		Version: version,
		Type:    "pip",
		Scope:   "runtime",
	}, nil
}

// getInstalledPackages gets installed packages using pip list
func (ps *PipScanner) getInstalledPackages() ([]model.Dependency, error) {
	var cmd *exec.Cmd
	if ps.pipPath == ps.pythonPath {
		cmd = exec.Command(ps.pythonPath, "-m", "pip", "list", "--format=freeze")
	} else {
		cmd = exec.Command(ps.pipPath, "list", "--format=freeze")
	}

	cmd.Dir = ps.environment.GetDirectory()
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run pip list: %w", err)
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
				Type:    "pip",
			},
			Name:    name,
			Version: version,
			Type:    "pip",
			Scope:   "runtime",
		})
	}

	return dependencies, nil
}

// mergeDependencies merges requirements and installed packages
func (ps *PipScanner) mergeDependencies(requirements, installed []model.Dependency) []model.Dependency {
	depMap := make(map[string]model.Dependency)

	// Add installed packages first
	for _, dep := range installed {
		depMap[dep.Name] = dep
	}

	// Override with requirements (they have more accurate version constraints)
	for _, dep := range requirements {
		depMap[dep.Name] = dep
	}

	var result []model.Dependency
	for _, dep := range depMap {
		result = append(result, dep)
	}

	return result
}

// parseSetupPy tries to extract project name and version from setup.py
func (ps *PipScanner) parseSetupPy(setupPath string) (string, string) {
	file, err := os.Open(setupPath)
	if err != nil {
		return "", ""
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var name, version string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "name=") {
			if start := strings.Index(line, "name="); start != -1 {
				nameExpr := line[start+5:]
				name = ps.extractQuotedValue(nameExpr)
			}
		}

		if strings.Contains(line, "version=") {
			if start := strings.Index(line, "version="); start != -1 {
				versionExpr := line[start+8:]
				version = ps.extractQuotedValue(versionExpr)
			}
		}
	}

	return name, version
}

// extractQuotedValue extracts a quoted value from a string
func (ps *PipScanner) extractQuotedValue(expr string) string {
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
