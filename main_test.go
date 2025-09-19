package main

import (
	"os"
	"path/filepath"
	"testing"

	"scnpm/pkg/types"
)

func TestParsePackageQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    types.PackageQuery
		wantErr bool
	}{
		{
			name:  "simple package",
			input: "react@18.2.0",
			want: types.PackageQuery{
				Name:    "react",
				Version: "18.2.0",
			},
			wantErr: false,
		},
		{
			name:  "scoped package",
			input: "@types/node@18.0.0",
			want: types.PackageQuery{
				Name:    "@types/node",
				Version: "18.0.0",
			},
			wantErr: false,
		},
		{
			name:  "package with complex version",
			input: "package@^1.2.3",
			want: types.PackageQuery{
				Name:    "package",
				Version: "^1.2.3",
			},
			wantErr: false,
		},
		{
			name:    "invalid format - no version",
			input:   "react",
			wantErr: true,
		},
		{
			name:  "multiple @ symbols in version",
			input: "package@1.0.0@beta",
			want: types.PackageQuery{
				Name:    "package",
				Version: "1.0.0@beta",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePackageQuery(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePackageQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Name != tt.want.Name || got.Version != tt.want.Version) {
				t.Errorf("parsePackageQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadPackagesFromFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-packages.json")

	content := `["react@18.2.0", "@types/node@18.0.0", "lodash@4.17.21"]`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	packages, err := readPackagesFromFile(testFile)
	if err != nil {
		t.Errorf("readPackagesFromFile() returned error: %v", err)
	}

	expected := []string{"react@18.2.0", "@types/node@18.0.0", "lodash@4.17.21"}
	if len(packages) != len(expected) {
		t.Errorf("readPackagesFromFile() returned %d packages, want %d", len(packages), len(expected))
	}

	for i, pkg := range packages {
		if pkg != expected[i] {
			t.Errorf("Package[%d] = %q, want %q", i, pkg, expected[i])
		}
	}

	// Test non-existent file
	_, err = readPackagesFromFile("non-existent-file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}

	_, err = readPackagesFromFile(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestReadPackageLock(t *testing.T) {
	// Create a temporary test package-lock.json
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "package-lock.json")

	content := `{
		"name": "test-project",
		"version": "1.0.0",
		"lockfileVersion": 2,
		"packages": {
			"node_modules/react": {
				"version": "18.2.0",
				"dev": false
			}
		}
	}`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	packageLock, err := readPackageLock(testFile)
	if err != nil {
		t.Errorf("readPackageLock() returned error: %v", err)
	}

	if packageLock.Name != "test-project" {
		t.Errorf("Package name = %q, want %q", packageLock.Name, "test-project")
	}

	if packageLock.Version != "1.0.0" {
		t.Errorf("Package version = %q, want %q", packageLock.Version, "1.0.0")
	}

	if packageLock.LockfileVersion != 2 {
		t.Errorf("Lockfile version = %d, want %d", packageLock.LockfileVersion, 2)
	}

	if len(packageLock.Packages) != 1 {
		t.Errorf("Number of packages = %d, want %d", len(packageLock.Packages), 1)
	}

	// Test non-existent file
	_, err = readPackageLock("non-existent-file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}