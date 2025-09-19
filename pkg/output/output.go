package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"scnpm/pkg/types"
)

// OutputConfig contains configuration for output formatting
type OutputConfig struct {
	ShowSafe bool
	RiskOnly bool
}

// OutputTable displays results in table format
func OutputTable(results []types.ScanResult, config OutputConfig) {
	fmt.Printf("%-30s %-15s %-8s %-15s %-8s %-8s %s\n", "Package", "Target Ver", "Status", "Found Ver", "Dev", "Line#", "Path")
	fmt.Println(strings.Repeat("-", 120))

	for _, result := range results {
		if !result.Found {
			// Only show safe packages if showSafe is true and riskOnly is false
			if config.ShowSafe && !config.RiskOnly {
				fmt.Printf("%-30s %-15s %-8s %-15s %-8s %-8s %s\n",
					result.Package.Name,
					result.Package.Version,
					"✅ SAFE",
					"Not Found",
					"-",
					"-",
					"Package not detected in project",
				)
			}
			continue
		}

		// Group instances by version for cleaner output
		versionGroups := make(map[string][]types.PackageInstance)
		for _, instance := range result.Instances {
			versionGroups[instance.Version] = append(versionGroups[instance.Version], instance)
		}

		first := true
		for version, instances := range versionGroups {
			for i, instance := range instances {
				packageName := result.Package.Name
				expectedVersion := result.Package.Version

				if !first || i > 0 {
					packageName = ""
					expectedVersion = ""
				}

				devStatus := "-"
				if instance.IsDev {
					devStatus = "✓"
				}

				lineStatus := "-"
				if instance.LineNumber > 0 {
					lineStatus = fmt.Sprintf("L%d", instance.LineNumber)
				}

				// Determine security status
				status := "🚨 RISK"
				if instance.IsReference {
					status = "⚠️ REF"
				}

				fmt.Printf("%-30s %-15s %-8s %-15s %-8s %-8s %s\n",
					packageName,
					expectedVersion,
					status,
					version,
					devStatus,
					lineStatus,
					instance.Path,
				)
				first = false
			}
		}

		if result.TotalInstances > 1 {
			fmt.Printf("%-30s %-15s %-8s %-15s %-8s %-8s %s\n",
				"",
				"",
				"",
				fmt.Sprintf("(%d total)", result.TotalInstances),
				"",
				"",
				"",
			)
		}
	}

	// Security Summary
	totalRisks := 0
	totalSafe := 0
	for _, result := range results {
		if result.Found {
			totalRisks++
		} else {
			totalSafe++
		}
	}

	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("SECURITY SUMMARY: 🚨 %d RISKS DETECTED | ✅ %d PACKAGES SAFE\n", totalRisks, totalSafe)
	if totalRisks > 0 {
		fmt.Printf("⚠️  WARNING: Found %d potentially compromised packages in your project!\n", totalRisks)
	} else {
		fmt.Printf("✅ GOOD: No known compromised packages detected in your project.\n")
	}
}

// OutputJSON displays results in JSON format
func OutputJSON(results []types.ScanResult) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}