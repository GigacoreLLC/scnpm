package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"scnpm/pkg/output"
	"scnpm/pkg/scanner"
	"scnpm/pkg/types"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)


var rootCmd = &cobra.Command{
	Use:     "scnpm [badpak.json]",
	Short:   "Security scanner for malware-affected npm packages",
	Version: version,
	Long: `A security CLI tool to scan package-lock.json files for potentially compromised npm packages.
This tool helps identify packages that may have been affected by malware, supply chain attacks,
or other security vulnerabilities. Finding packages indicates potential security risks.

Usage examples:
  scnpm badpak.json                                      # Scan bad packages from JSON file
  scnpm --file /path/to/package-lock.json badpak.json   # Custom package-lock path
  scnpm --packages-file /path/to/badpak.json            # Alternative flag syntax with path
  scnpm --file ~/project/package-lock.json ~/lists/badpak.json  # Files from different directories
  scnpm package@1.0.0 another@2.0.0                      # Direct package arguments`,
	Run: runScan,
}

var (
	packageLockPath string
	packagesFlag    []string
	packagesFile    string
	outputFormat    string
	showAllVersions bool
	showDevOnly     bool
	showNestedOnly  bool
	minDepth        int
	showMetadata    bool
	showDependencies bool
	showEngines     bool
	searchInDeps    bool
	riskOnly        bool
	showSafe        bool
)

func init() {
	rootCmd.Flags().StringVarP(&packageLockPath, "file", "f", "package-lock.json", "Path to package-lock.json file")
	rootCmd.Flags().StringSliceVarP(&packagesFlag, "packages", "p", []string{}, "List of packages to scan (format: package@version)")
	rootCmd.Flags().StringVar(&packagesFile, "packages-file", "", "Path to JSON file containing array of bad packages to scan (e.g., badpak.json)")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")
	rootCmd.Flags().BoolVar(&showAllVersions, "all-versions", false, "Show all versions found, not just first match")
	rootCmd.Flags().BoolVar(&showDevOnly, "dev-only", false, "Show only development dependencies")
	rootCmd.Flags().BoolVar(&showNestedOnly, "nested-only", false, "Show only nested dependencies")
	rootCmd.Flags().IntVar(&minDepth, "min-depth", 0, "Minimum nesting depth to show")
	rootCmd.Flags().BoolVar(&showMetadata, "metadata", false, "Include comprehensive metadata (resolved, integrity, license)")
	rootCmd.Flags().BoolVar(&showDependencies, "show-deps", false, "Include dependencies and peerDependencies")
	rootCmd.Flags().BoolVar(&showEngines, "show-engines", false, "Include engines and other technical metadata")
	rootCmd.Flags().BoolVar(&searchInDeps, "search-in-deps", true, "Search within dependency requirements of other packages (enabled by default for comprehensive malware detection)")
	rootCmd.Flags().BoolVar(&riskOnly, "risk-only", false, "Show only packages that pose security risks (hide safe packages)")
	rootCmd.Flags().BoolVar(&showSafe, "show-safe", true, "Show packages that were not found (safe packages)")

	// Add version template
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
Build: {{printf "%s" .Version}}
Commit: ` + commit + `
Date: ` + date + `
`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) {
	// Parse package queries from various sources
	var packageQueries []types.PackageQuery
	var packagesToScan []string
	
	// 1. Check if first argument is a JSON file (new positional syntax)
	if len(args) > 0 && strings.HasSuffix(args[0], ".json") {
		packages, err := readPackagesFromFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading packages file '%s': %v\n", args[0], err)
			os.Exit(1)
		}
		packagesToScan = append(packagesToScan, packages...)
		args = args[1:] // Remove the JSON file from args
	}
	
	// 2. Check --packages-file flag
	if packagesFile != "" {
		packages, err := readPackagesFromFile(packagesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading packages file '%s': %v\n", packagesFile, err)
			os.Exit(1)
		}
		packagesToScan = append(packagesToScan, packages...)
	}
	
	// 3. Add packages from --packages flag
	packagesToScan = append(packagesToScan, packagesFlag...)
	
	// 4. Add remaining command line arguments as packages
	packagesToScan = append(packagesToScan, args...)
	
	// Parse all packages into queries
	for _, pkg := range packagesToScan {
		query, err := parsePackageQuery(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing package '%s': %v\n", pkg, err)
			continue
		}
		packageQueries = append(packageQueries, query)
	}
	
	if len(packageQueries) == 0 {
		fmt.Fprintf(os.Stderr, "No packages specified. Use one of the following methods:\n")
		fmt.Fprintf(os.Stderr, "  scnpm badpak.json\n")
		fmt.Fprintf(os.Stderr, "  scnpm --packages-file badpak.json\n")
		fmt.Fprintf(os.Stderr, "  scnpm --packages package@1.0.0,another@2.0.0\n")
		fmt.Fprintf(os.Stderr, "  scnpm package@1.0.0 another@2.0.0\n")
		os.Exit(1)
	}
	
	// Resolve package-lock.json path (support both relative and absolute paths)
	absPackageLockPath, err := filepath.Abs(packageLockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path '%s': %v\n", packageLockPath, err)
		os.Exit(1)
	}

	// Check if package-lock.json exists
	if _, err := os.Stat(absPackageLockPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: package-lock.json not found at '%s'\n", absPackageLockPath)
		os.Exit(1)
	}
	
	// Read and parse package-lock.json
	packageLock, err := readPackageLock(absPackageLockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading package-lock.json: %v\n", err)
		os.Exit(1)
	}
	
	// Create filter and output configs
	filterConfig := scanner.FilterConfig{
		ShowDevOnly:    showDevOnly,
		ShowNestedOnly: showNestedOnly,
		MinDepth:       minDepth,
	}

	outputConfig := output.OutputConfig{
		ShowSafe: showSafe,
		RiskOnly: riskOnly,
	}

	// Scan for packages
	results := scanner.ScanPackages(packageLock, packageQueries, filterConfig)

	// Output results
	switch outputFormat {
	case "json":
		output.OutputJSON(results)
	case "table":
		output.OutputTable(results, outputConfig)
	default:
		fmt.Fprintf(os.Stderr, "Unknown output format: %s\n", outputFormat)
		os.Exit(1)
	}
}

func parsePackageQuery(input string) (types.PackageQuery, error) {
	parts := strings.Split(input, "@")
	if len(parts) < 2 {
		return types.PackageQuery{}, fmt.Errorf("invalid format, expected package@version")
	}

	// Handle scoped packages like @types/node@1.0.0
	if strings.HasPrefix(input, "@") && len(parts) >= 3 {
		return types.PackageQuery{
			Name:    "@" + parts[1],
			Version: strings.Join(parts[2:], "@"),
		}, nil
	}

	return types.PackageQuery{
		Name:    parts[0],
		Version: strings.Join(parts[1:], "@"),
	}, nil
}

func readPackageLock(path string) (*types.PackageLock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var packageLock types.PackageLock
	if err := json.Unmarshal(data, &packageLock); err != nil {
		return nil, err
	}

	return &packageLock, nil
}

// readPackagesFromFile reads packages from a JSON file
func readPackagesFromFile(filePath string) ([]string, error) {
	// Resolve to absolute path for better error messages and consistency
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path '%s': %v", filePath, err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages file '%s': %v", absPath, err)
	}

	var packages []string
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, fmt.Errorf("failed to parse packages JSON from '%s': %v", absPath, err)
	}

	return packages, nil
}