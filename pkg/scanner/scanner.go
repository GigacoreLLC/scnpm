package scanner

import (
	"strings"

	"scnpm/pkg/types"
)

// FilterConfig contains configuration for filtering scan results
type FilterConfig struct {
	ShowDevOnly    bool
	ShowNestedOnly bool
	MinDepth       int
}

// ScanPackages scans for packages in the package-lock.json
func ScanPackages(packageLock *types.PackageLock, queries []types.PackageQuery, config FilterConfig) []types.ScanResult {
	results := make([]types.ScanResult, len(queries))

	for i, query := range queries {
		result := types.ScanResult{
			Package:   query,
			Found:     false,
			Instances: []types.PackageInstance{},
		}

		// Search through the parsed packageLock data instead of re-reading file
		instances := findPackageInstancesInLock(packageLock, query.Name, query.Version)

		for _, instance := range instances {
			result.Instances = append(result.Instances, instance)
		}

		// Apply filters
		result.Instances = applyFilters(result.Instances, config)
		result.TotalInstances = len(result.Instances)
		result.Found = result.TotalInstances > 0

		results[i] = result
	}

	return results
}

// findPackageInstancesInLock searches for package instances in the parsed PackageLock data
func findPackageInstancesInLock(packageLock *types.PackageLock, packageName, version string) []types.PackageInstance {
	var instances []types.PackageInstance

	// Handle different lockfile versions
	if packageLock.LockfileVersion >= 2 {
		// Search in packages field (lockfileVersion 2+)
		for path, pkg := range packageLock.Packages {
			if matchesPackageInPath(path, packageName) && (version == "" || pkg.Version == version) {
				instance := types.PackageInstance{
					Version:     pkg.Version,
					Path:        path,
					LineNumber:  0, // Not available from parsed data
					IsReference: false,
					IsDev:       pkg.Dev,
					IsNested:    strings.Contains(path, "/node_modules/"),
					Depth:       strings.Count(path, "/node_modules/"),
				}
				instances = append(instances, instance)
			}
		}

		// Also check dependencies references in packages
		for path, pkg := range packageLock.Packages {
			for depName, depVersion := range pkg.Dependencies {
				if MatchesPackageName(depName, packageName) && (version == "" || strings.Contains(depVersion, version)) {
					instance := types.PackageInstance{
						Version:       depVersion,
						Path:          path + " -> " + depName,
						LineNumber:    0,
						IsReference:   true,
						ReferenceType: "dependencies",
						IsDev:         pkg.Dev,
						IsNested:      strings.Contains(path, "/node_modules/"),
						Depth:         strings.Count(path, "/node_modules/") + 1,
					}
					instances = append(instances, instance)
				}
			}
		}
	} else {
		// Search in dependencies field (lockfileVersion 1)
		instances = append(instances, searchDependenciesRecursive(packageLock.Dependencies, packageName, version, "")...)
	}

	return instances
}

// matchesPackageInPath checks if a path contains the specified package name
func matchesPackageInPath(path, packageName string) bool {
	// Extract package name from path like "node_modules/package-name" or "node_modules/@scope/package-name"
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "node_modules" && i+1 < len(parts) {
			// Handle scoped packages
			if strings.HasPrefix(parts[i+1], "@") && i+2 < len(parts) {
				fullName := parts[i+1] + "/" + parts[i+2]
				if MatchesPackageName(fullName, packageName) {
					return true
				}
			} else {
				if MatchesPackageName(parts[i+1], packageName) {
					return true
				}
			}
		}
	}
	return false
}

// searchDependenciesRecursive searches through the dependencies tree recursively (lockfileVersion 1)
func searchDependenciesRecursive(deps map[string]types.Dependency, packageName, version, basePath string) []types.PackageInstance {
	var instances []types.PackageInstance

	for depName, dep := range deps {
		currentPath := basePath
		if currentPath == "" {
			currentPath = "node_modules/" + depName
		} else {
			currentPath = currentPath + "/node_modules/" + depName
		}

		// Check if this dependency matches
		if MatchesPackageName(depName, packageName) && (version == "" || dep.Version == version) {
			instance := types.PackageInstance{
				Version:     dep.Version,
				Path:        currentPath,
				LineNumber:  0,
				IsReference: false,
				IsDev:       dep.Dev,
				IsNested:    strings.Contains(currentPath, "/node_modules/"),
				Depth:       strings.Count(currentPath, "/node_modules/"),
			}
			instances = append(instances, instance)
		}

		// Recursively search nested dependencies
		if dep.Dependencies != nil {
			instances = append(instances, searchDependenciesRecursive(dep.Dependencies, packageName, version, currentPath)...)
		}
	}

	return instances
}

// MatchesPackageName checks if a package name matches the query with sophisticated matching logic
func MatchesPackageName(packageName, queryName string) bool {
	// Exact match
	if packageName == queryName {
		return true
	}

	// Handle scoped packages - allow matching with or without @ prefix
	if strings.HasPrefix(packageName, "@") && !strings.HasPrefix(queryName, "@") {
		// Package is scoped, query is not - check if query matches the package part
		parts := strings.Split(packageName, "/")
		if len(parts) == 2 && parts[1] == queryName {
			return true
		}
	}

	if !strings.HasPrefix(packageName, "@") && strings.HasPrefix(queryName, "@") {
		// Query is scoped, package is not - check if package matches the scoped part
		parts := strings.Split(queryName, "/")
		if len(parts) == 2 && parts[1] == packageName {
			return true
		}
	}

	// Partial matching for cases where package names might have variations
	if strings.Contains(packageName, queryName) || strings.Contains(queryName, packageName) {
		return true
	}

	return false
}

// applyFilters applies command-line filters to the found instances
func applyFilters(instances []types.PackageInstance, config FilterConfig) []types.PackageInstance {
	var filtered []types.PackageInstance

	for _, instance := range instances {
		// Apply dev-only filter
		if config.ShowDevOnly && !instance.IsDev {
			continue
		}

		// Apply nested-only filter
		if config.ShowNestedOnly && !instance.IsNested {
			continue
		}

		// Apply minimum depth filter
		if instance.Depth < config.MinDepth {
			continue
		}

		filtered = append(filtered, instance)
	}

	return filtered
}