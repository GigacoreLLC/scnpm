package scanner

import (
	"reflect"
	"testing"

	"scnpm/pkg/types"
)

func TestMatchesPackageName(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		queryName   string
		want        bool
	}{
		{
			name:        "exact match",
			packageName: "react",
			queryName:   "react",
			want:        true,
		},
		{
			name:        "scoped package exact match",
			packageName: "@types/node",
			queryName:   "@types/node",
			want:        true,
		},
		{
			name:        "scoped package without @ prefix",
			packageName: "@types/node",
			queryName:   "node",
			want:        true,
		},
		{
			name:        "query scoped, package not",
			packageName: "node",
			queryName:   "@types/node",
			want:        true,
		},
		{
			name:        "partial match - contains",
			packageName: "react-dom",
			queryName:   "react",
			want:        true,
		},
		{
			name:        "no match",
			packageName: "vue",
			queryName:   "react",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchesPackageName(tt.packageName, tt.queryName); got != tt.want {
				t.Errorf("matchesPackageName(%q, %q) = %v, want %v", tt.packageName, tt.queryName, got, tt.want)
			}
		})
	}
}

func TestMatchesPackageInPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		packageName string
		want        bool
	}{
		{
			name:        "simple package in node_modules",
			path:        "node_modules/react",
			packageName: "react",
			want:        true,
		},
		{
			name:        "scoped package in node_modules",
			path:        "node_modules/@types/node",
			packageName: "@types/node",
			want:        true,
		},
		{
			name:        "nested package",
			path:        "node_modules/express/node_modules/debug",
			packageName: "debug",
			want:        true,
		},
		{
			name:        "package not in path",
			path:        "node_modules/react",
			packageName: "vue",
			want:        false,
		},
		{
			name:        "empty path for root packages",
			path:        "",
			packageName: "react",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesPackageInPath(tt.path, tt.packageName); got != tt.want {
				t.Errorf("matchesPackageInPath(%q, %q) = %v, want %v", tt.path, tt.packageName, got, tt.want)
			}
		})
	}
}

func TestApplyFilters(t *testing.T) {
	instances := []types.PackageInstance{
		{Version: "1.0.0", IsDev: true, IsNested: false, Depth: 0},
		{Version: "2.0.0", IsDev: false, IsNested: true, Depth: 1},
		{Version: "3.0.0", IsDev: true, IsNested: true, Depth: 2},
	}

	tests := []struct {
		name           string
		config         FilterConfig
		expectedCount  int
		expectedVersions []string
	}{
		{
			name: "no filters",
			config: FilterConfig{
				ShowDevOnly:    false,
				ShowNestedOnly: false,
				MinDepth:       0,
			},
			expectedCount: 3,
			expectedVersions: []string{"1.0.0", "2.0.0", "3.0.0"},
		},
		{
			name: "dev only",
			config: FilterConfig{
				ShowDevOnly:    true,
				ShowNestedOnly: false,
				MinDepth:       0,
			},
			expectedCount: 2,
			expectedVersions: []string{"1.0.0", "3.0.0"},
		},
		{
			name: "nested only",
			config: FilterConfig{
				ShowDevOnly:    false,
				ShowNestedOnly: true,
				MinDepth:       0,
			},
			expectedCount: 2,
			expectedVersions: []string{"2.0.0", "3.0.0"},
		},
		{
			name: "minimum depth 2",
			config: FilterConfig{
				ShowDevOnly:    false,
				ShowNestedOnly: false,
				MinDepth:       2,
			},
			expectedCount: 1,
			expectedVersions: []string{"3.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := applyFilters(instances, tt.config)

			if len(filtered) != tt.expectedCount {
				t.Errorf("applyFilters() returned %d instances, want %d", len(filtered), tt.expectedCount)
			}

			var gotVersions []string
			for _, inst := range filtered {
				gotVersions = append(gotVersions, inst.Version)
			}

			if !reflect.DeepEqual(gotVersions, tt.expectedVersions) {
				t.Errorf("applyFilters() returned versions %v, want %v", gotVersions, tt.expectedVersions)
			}
		})
	}
}

func TestScanPackages(t *testing.T) {
	// Create a mock package-lock structure
	packageLock := &types.PackageLock{
		LockfileVersion: 2,
		Packages: map[string]types.Package{
			"node_modules/react": {
				Version: "18.2.0",
				Dev:     false,
			},
			"node_modules/@types/node": {
				Version: "18.0.0",
				Dev:     true,
			},
			"node_modules/express/node_modules/debug": {
				Version: "2.6.9",
				Dev:     false,
			},
		},
	}

	queries := []types.PackageQuery{
		{Name: "react", Version: "18.2.0"},
		{Name: "@types/node", Version: "18.0.0"},
		{Name: "vue", Version: "3.0.0"},
	}

	// Create default filter config
	config := FilterConfig{
		ShowDevOnly:    false,
		ShowNestedOnly: false,
		MinDepth:       0,
	}

	results := ScanPackages(packageLock, queries, config)

	if len(results) != 3 {
		t.Errorf("scanPackages() returned %d results, want 3", len(results))
	}

	// Check react was found
	if !results[0].Found {
		t.Error("Expected react to be found")
	}
	if results[0].TotalInstances != 1 {
		t.Errorf("Expected 1 instance of react, got %d", results[0].TotalInstances)
	}

	// Check @types/node was found
	if !results[1].Found {
		t.Error("Expected @types/node to be found")
	}

	// Check vue was not found
	if results[2].Found {
		t.Error("Expected vue to not be found")
	}
}