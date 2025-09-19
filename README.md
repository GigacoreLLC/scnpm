# scnpm - Security Scanner for NPM Packages

A fast / manual Go CLI tool that scans `package-lock.json` files for potentially compromised npm packages.

## Quick Start

```bash
# Install
brew tap GigacoreLLC/scnpm https://github.com/GigacoreLLC/scnpm && brew install scnpm

# Move to your project directory
cd /path/to/your/package-lock.json

# Create a list of suspicious packages
echo '["debug@4.3.4", "chalk@5.3.0"]' > badpak.json

# Example list from recent supply chain attacks
[
  "ansi-styles@6.2.2",
  "debug@4.4.2",
  "chalk@5.6.1",
  "supports-color@10.2.1",
  "strip-ansi@7.1.1",
  "ansi-regex@6.2.1",
  "wrap-ansi@9.0.1",
  "color-convert@3.1.1",
  "color-name@2.0.1",
  "is-arrayish@0.3.3",
  "slice-ansi@7.1.1",
  "color@5.0.1",
  "color-string@2.1.1",
  "simple-swizzle@0.2.3",
  "supports-hyperlinks@4.1.1",
  "has-ansi@6.0.1",
  "chalk-template@1.1.1",
  "backslash@0.2.1"
]

# Scan your project
scnpm --file package-lock.json badpak.json
```

## Features

- 🔍 **Comprehensive Detection** - Finds all instances across the entire dependency tree
- 🎯 **Exact Version Matching** - Precise version detection with regex-based search
- 🚨 **Risk Classification** - Distinguishes between installed packages and references
- 📊 **Multiple Formats** - Table or JSON output
- 🔧 **Advanced Filtering** - Filter by dev dependencies, nesting depth, and more
- 📁 **Cross-Directory** - Scan files from anywhere on your filesystem

## Installation

### Homebrew

```bash
brew tap GigacoreLLC/scnpm https://github.com/GigacoreLLC/scnpm
brew install scnpm
```

### From Source

```bash
git clone https://github.com/GigacoreLLC/scnpm.git
cd scnpm
go build -o scnpm
```

## Usage

### Basic Scan

```bash
# Using a badpak.json file (recommended)
scnpm badpak.json

# Scan specific packages directly
scnpm react@18.2.0 lodash@4.17.21

# Specify custom package-lock.json location
scnpm --file /path/to/package-lock.json badpak.json
```

### Create badpak.json

```json
["debug@4.3.4", "chalk@5.3.0", "lodash@4.17.21"]
```

### Options

- `-f, --file` - Path to package-lock.json (default: "./package-lock.json")
- `-o, --output` - Output format: "table" or "json" (default: "table")
- `--dev-only` - Show only development dependencies
- `--nested-only` - Show only nested dependencies
- `--min-depth N` - Show dependencies at minimum depth N

## Example Output

```bash
$ scnpm badpak.json

Package                        Target Ver      Status   Found Ver       Dev      Line#    Path
------------------------------------------------------------------------------------------------------------------------
debug                          4.3.4           ✅ SAFE   Not Found       -        -        Package not detected in project
chalk                          5.3.0           ⚠️ REF   ^5.3.0          -        -        node_modules/svgo/node_modules/chalk -> chalk
lodash                         4.17.21         🚨 RISK   4.17.21         -        -        node_modules/lodash
========================================================================================================================
SECURITY SUMMARY: 🚨 1 RISK DETECTED | ✅ 2 PACKAGES SAFE
```

### Status Indicators

- ✅ **SAFE** - Package not found in your project
- 🚨 **RISK** - Package is installed (investigate immediately)
- ⚠️ **REF** - Package referenced in dependencies (potential risk)

## Advanced Features

### Detection Capabilities

- Finds all instances across entire dependency tree
- Handles scoped packages (`@types/node`, `@babel/core`)
- Detects nested dependencies at any depth
- Distinguishes development vs production dependencies
- Supports all npm lockfile formats (v1, v2, v3)

### Cross-Directory Scanning

```bash
# Scan files from different locations
scnpm --file ~/project/package-lock.json ~/security/badpak.json

# Works from any directory
cd /tmp && scnpm --file /app/package-lock.json /lists/badpak.json
```

## Development

```bash
# Build
go build -o scnpm

# Test
go test ./...

# Install locally
go install .
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

MIT - See [LICENSE](LICENSE)
