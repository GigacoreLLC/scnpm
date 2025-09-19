package types

// PackageLock represents the structure of a package-lock.json file
type PackageLock struct {
	Name            string                `json:"name"`
	Version         string                `json:"version"`
	LockfileVersion int                   `json:"lockfileVersion"`
	Dependencies    map[string]Dependency `json:"dependencies,omitempty"`
	Packages        map[string]Package    `json:"packages,omitempty"`
}

// Dependency represents a dependency in the old format (lockfileVersion 1)
type Dependency struct {
	Version      string                `json:"version"`
	Resolved     string                `json:"resolved,omitempty"`
	Integrity    string                `json:"integrity,omitempty"`
	Dev          bool                  `json:"dev,omitempty"`
	Dependencies map[string]Dependency `json:"dependencies,omitempty"`
}

// Package represents a package in the new format (lockfileVersion 2+)
type Package struct {
	Version          string            `json:"version,omitempty"`
	Resolved         string            `json:"resolved,omitempty"`
	Integrity        string            `json:"integrity,omitempty"`
	Dev              bool              `json:"dev,omitempty"`
	DevOptional      bool              `json:"devOptional,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
	Engines          any               `json:"engines,omitempty"`
	License          string            `json:"license,omitempty"`
	Bin              any               `json:"bin,omitempty"`
	Scripts          map[string]string `json:"scripts,omitempty"`
}

// PackageQuery represents a package to search for
type PackageQuery struct {
	Name    string
	Version string
}

// ScanResult represents the result of scanning for a package
type ScanResult struct {
	Package        PackageQuery
	Found          bool
	Instances      []PackageInstance
	TotalInstances int
}

// PackageInstance represents a single instance of a package found
type PackageInstance struct {
	Version          string            `json:"version"`
	Path             string            `json:"path"`
	IsDev            bool              `json:"isDev"`
	IsNested         bool              `json:"isNested"`
	Depth            int               `json:"depth"`
	LineNumber       int               `json:"lineNumber,omitempty"`       // Line number in package-lock.json
	Resolved         string            `json:"resolved,omitempty"`
	Integrity        string            `json:"integrity,omitempty"`
	License          string            `json:"license,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
	Engines          any               `json:"engines,omitempty"`
	Bin              any               `json:"bin,omitempty"`
	Scripts          map[string]string `json:"scripts,omitempty"`
	IsReference      bool              `json:"isReference,omitempty"`   // True if found as dependency reference
	ReferencedBy     string            `json:"referencedBy,omitempty"`  // Package that references this
	ReferenceType    string            `json:"referenceType,omitempty"` // "dependencies", "peerDependencies", etc.
}