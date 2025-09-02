package model

import "github.com/craftslab/cleansource-sca-cli/internal/config"

// UploadData represents data to be uploaded to the server
type UploadData struct {
	WfpFile     string             `json:"wfpFile"`
	BuildFile   string             `json:"buildFile"`
	ArchiveFile string             `json:"archiveFile"`
	Config      *config.ScanConfig `json:"config"`
	DirSize     int64              `json:"dirSize"`
}

// Dependency represents a single dependency
type Dependency struct {
	ID       *DependencyID `json:"id"`
	Name     string        `json:"name"`
	GroupID  string        `json:"groupId,omitempty"` // Add GroupID for compatibility
	Version  string        `json:"version"`
	Type     string        `json:"type"`
	Scope    string        `json:"scope,omitempty"`
	Children []Dependency  `json:"children,omitempty"`
}

// DependencyID represents a unique identifier for a dependency
type DependencyID struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

// DependencyRoot represents the root of a dependency tree
type DependencyRoot struct {
	ProjectName    string       `json:"projectName"`
	ProjectVersion string       `json:"projectVersion"`
	BuildTool      string       `json:"buildTool"`
	Dependencies   []Dependency `json:"dependencies"`
}

// ScanType represents different types of scans
type ScanType string

const (
	ScanTypeSource ScanType = "source"
	ScanTypeDocker ScanType = "docker"
	ScanTypeBinary ScanType = "binary"
)

// TaskType represents different types of tasks
type TaskType string

const (
	TaskTypeScan TaskType = "scan"
)

// FilePathCollect represents collected file paths during scanning
type FilePathCollect struct {
	ProjectLicenseFiles []string `json:"projectLicenseFiles"`
	SourceFiles         []string `json:"sourceFiles"`
	BinaryFiles         []string `json:"binaryFiles"`
}

// FilterCondition represents a filter condition for dependencies
type FilterCondition struct {
	Path      string `json:"path"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

// BinaryFilterParam represents parameters for binary file filtering
type BinaryFilterParam struct {
	MixedBinaryScanFlag         int      `json:"mixedBinaryScanFlag"`
	MixedBinaryScanFilePathList []string `json:"mixedBinaryScanFilePathList"`
	BinaryScanList              []string `json:"binaryScanList"`
	BinaryRealScanList          []string `json:"binaryRealScanList"`
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	TaskID     string `json:"taskId,omitempty"`
	ResultFile string `json:"resultFile,omitempty"`
}

// ExecutableInfo represents information about an executable
type ExecutableInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

// ProjectInfo represents project information
type ProjectInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	License     string `json:"license"`
	BuildTool   string `json:"buildTool"`
}
