package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"scnpm/pkg/scanner"
	"scnpm/pkg/types"
)

func TestFullScanWorkflow(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	// Create a test package-lock.json
	packageLockContent := `{
		"name": "test-project",
		"version": "1.0.0",
		"lockfileVersion": 2,
		"packages": {
			"node_modules/react": {
				"version": "18.2.0",
				"dev": false
			},
			"node_modules/lodash": {
				"version": "4.17.21",
				"dev": false
			},
			"node_modules/@types/node": {
				"version": "18.0.0",
				"dev": true
			}
		}
	}`

	packageLockPath := filepath.Join(tmpDir, "package-lock.json")
	err := os.WriteFile(packageLockPath, []byte(packageLockContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test package-lock.json: %v", err)
	}

	// Create a test badpak.json
	badpakContent := `["react@18.2.0", "vue@3.0.0", "@types/node@18.0.0"]`
	badpakPath := filepath.Join(tmpDir, "badpak.json")
	err = os.WriteFile(badpakPath, []byte(badpakContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test badpak.json: %v", err)
	}

	// Test reading package-lock.json
	packageLock, err := readTestPackageLock(packageLockPath)
	if err != nil {
		t.Errorf("Failed to read package-lock.json: %v", err)
	}

	// Test scanning
	queries := []types.PackageQuery{
		{Name: "react", Version: "18.2.0"},
		{Name: "vue", Version: "3.0.0"},
		{Name: "@types/node", Version: "18.0.0"},
	}

	config := scanner.FilterConfig{
		ShowDevOnly:    false,
		ShowNestedOnly: false,
		MinDepth:       0,
	}

	results := scanner.ScanPackages(packageLock, queries, config)

	// Verify results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Check react was found
	if !results[0].Found {
		t.Error("Expected react@18.2.0 to be found")
	}

	// Check vue was not found
	if results[1].Found {
		t.Error("Expected vue@3.0.0 to not be found")
	}

	// Check @types/node was found
	if !results[2].Found {
		t.Error("Expected @types/node@18.0.0 to be found")
	}
}

func TestCrossDirectoryScanning(t *testing.T) {
	// Create two separate temp directories
	projectDir := t.TempDir()
	listsDir := t.TempDir()

	// Create package-lock.json in project directory
	packageLockContent := `{
		"name": "cross-dir-test",
		"version": "1.0.0",
		"lockfileVersion": 2,
		"packages": {
			"node_modules/debug": {
				"version": "4.3.4",
				"dev": false
			}
		}
	}`

	packageLockPath := filepath.Join(projectDir, "package-lock.json")
	err := os.WriteFile(packageLockPath, []byte(packageLockContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package-lock.json: %v", err)
	}

	// Create badpak.json in lists directory
	badpakContent := `["debug@4.3.4", "express@4.0.0"]`
	badpakPath := filepath.Join(listsDir, "badpak.json")
	err = os.WriteFile(badpakPath, []byte(badpakContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create badpak.json: %v", err)
	}

	// Verify both files exist and can be read
	if _, err := os.Stat(packageLockPath); os.IsNotExist(err) {
		t.Errorf("package-lock.json not found at %s", packageLockPath)
	}

	if _, err := os.Stat(badpakPath); os.IsNotExist(err) {
		t.Errorf("badpak.json not found at %s", badpakPath)
	}

	// Test that files from different directories work
	packageLock, err := readTestPackageLock(packageLockPath)
	if err != nil {
		t.Errorf("Failed to read package-lock.json from %s: %v", packageLockPath, err)
	}

	if packageLock.Name != "cross-dir-test" {
		t.Errorf("Expected project name 'cross-dir-test', got '%s'", packageLock.Name)
	}
}

// Helper function to read package-lock.json for tests
func readTestPackageLock(path string) (*types.PackageLock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var packageLock types.PackageLock
	err = json.Unmarshal(data, &packageLock)
	if err != nil {
		return nil, err
	}

	return &packageLock, nil
}