# README Generation Reference

Detailed technical reference for README generation patterns, metadata parsing, and validation.

---

## Metadata Parsing Specifications

### Python: pyproject.toml (PEP 621)

**Standard Fields**:
```toml
[project]
name = "package-name"              # Required
version = "0.1.0"                  # Required
description = "One-line summary"   # Required for PyPI
readme = "README.md"               # Auto-detected
requires-python = ">=3.12"         # Python version constraint
license = {text = "MIT"}           # SPDX identifier
keywords = ["key1", "key2"]        # Searchable keywords
authors = [
    {name = "Author Name", email = "author@example.com"}
]
maintainers = [
    {name = "Maintainer Name", email = "maintainer@example.com"}
]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3.12"
]

[project.urls]
Homepage = "https://example.com"
Repository = "https://github.com/user/repo"
Documentation = "https://docs.example.com"
"Bug Tracker" = "https://github.com/user/repo/issues"
Changelog = "https://github.com/user/repo/blob/main/CHANGELOG.md"

[project.scripts]
package-name = "package_name.cli:main"  # CLI entry point

[project.dependencies]
dependency1 = ">=1.0.0"
dependency2 = "^2.0.0"

[project.optional-dependencies]
dev = ["pytest>=8.0", "ruff>=0.1"]
docs = ["mkdocs>=1.5"]
```

**Parsing Logic**:
```python
import tomli

with open("pyproject.toml", "rb") as f:
    data = tomli.load(f)

project = data.get("project", {})
name = project.get("name")
version = project.get("version")
description = project.get("description")
requires_python = project.get("requires-python")
license_text = project.get("license", {}).get("text")
urls = project.get("urls", {})
dependencies = project.get("dependencies", [])
scripts = project.get("scripts", {})

# CLI detection: has scripts = CLI tool
is_cli_tool = bool(scripts)
```

### JavaScript/TypeScript: package.json

**Standard Fields**:
```json
{
  "name": "package-name",
  "version": "1.0.0",
  "description": "One-line summary",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "bin": {
    "package-name": "./bin/cli.js"
  },
  "scripts": {
    "build": "tsc",
    "test": "vitest",
    "lint": "biome check .",
    "start": "node dist/index.js"
  },
  "keywords": ["keyword1", "keyword2"],
  "author": "Author Name <author@example.com>",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/user/repo.git"
  },
  "bugs": {
    "url": "https://github.com/user/repo/issues"
  },
  "homepage": "https://github.com/user/repo#readme",
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0"
  },
  "dependencies": {
    "dependency1": "^1.0.0"
  },
  "devDependencies": {
    "typescript": "^5.3.0",
    "vitest": "^1.0.0"
  }
}
```

**Parsing Logic**:
```javascript
const pkg = JSON.parse(fs.readFileSync("package.json", "utf8"));

const name = pkg.name;
const version = pkg.version;
const description = pkg.description;
const license = pkg.license;
const repository = pkg.repository?.url;
const homepage = pkg.homepage;
const nodeVersion = pkg.engines?.node;

// CLI detection: has "bin" field
const isCLITool = !!pkg.bin;

// Test framework detection
const hasVitest = "vitest" in (pkg.devDependencies || {});
const hasJest = "jest" in (pkg.devDependencies || {});
const testFramework = hasVitest ? "Vitest" : hasJest ? "Jest" : "unknown";
```

### Go: go.mod

**Standard Format**:
```go
module github.com/user/repo

go 1.22

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/lib/pq v1.10.9
    github.com/stretchr/testify v1.8.4
)

require (
    // indirect dependencies
    github.com/dependency/indirect v0.1.0 // indirect
)

replace github.com/old/module => github.com/new/module v1.0.0
```

**Parsing Logic**:
```go
// Parse go.mod
content := readFile("go.mod")
lines := strings.Split(content, "\n")

var moduleName string
var goVersion string
var dependencies []string

for _, line := range lines {
    line = strings.TrimSpace(line)
    
    if strings.HasPrefix(line, "module ") {
        moduleName = strings.TrimPrefix(line, "module ")
    }
    
    if strings.HasPrefix(line, "go ") {
        goVersion = strings.TrimPrefix(line, "go ")
    }
    
    if strings.Contains(line, "/") && !strings.HasSuffix(line, "// indirect") {
        dependencies = append(dependencies, line)
    }
}

// CLI detection: has cmd/main.go or main.go
isCLITool := fileExists("cmd/main.go") || fileExists("main.go")

// Framework detection from dependencies
hasGin := containsDep(dependencies, "gin-gonic/gin")
hasEcho := containsDep(dependencies, "labstack/echo")
framework := ""
if hasGin {
    framework = "Gin"
} else if hasEcho {
    framework = "Echo"
}
```

### Rust: Cargo.toml

**Standard Fields**:
```toml
[package]
name = "package-name"
version = "0.1.0"
edition = "2021"
authors = ["Author Name <author@example.com>"]
description = "One-line summary"
readme = "README.md"
homepage = "https://example.com"
repository = "https://github.com/user/repo"
license = "MIT OR Apache-2.0"
keywords = ["keyword1", "keyword2"]
categories = ["command-line-utilities"]

[[bin]]
name = "package-name"
path = "src/main.rs"

[dependencies]
clap = { version = "4.4", features = ["derive"] }
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.35", features = ["full"] }

[dev-dependencies]
assert_cmd = "2.0"
predicates = "3.0"

[profile.release]
opt-level = 3
lto = true
codegen-units = 1
```

**Parsing Logic**:
```rust
use toml::Value;

let cargo_toml = fs::read_to_string("Cargo.toml")?;
let value: Value = toml::from_str(&cargo_toml)?;

let package = value["package"].as_table()?;
let name = package["name"].as_str()?;
let version = package["version"].as_str()?;
let description = package["description"].as_str().unwrap_or("");
let license = package["license"].as_str().unwrap_or("");
let repository = package["repository"].as_str();

// CLI detection: has [[bin]] section or src/main.rs
let is_cli_tool = value.get("bin").is_some() 
    || Path::new("src/main.rs").exists();

// Dependencies
let deps = value["dependencies"].as_table()?;
let has_clap = deps.contains_key("clap");
let has_tokio = deps.contains_key("tokio");
```

---

## Badge Generation

### Version Badges

**PyPI** (Python):
```markdown
[![PyPI version](https://img.shields.io/pypi/v/PACKAGE-NAME)](https://pypi.org/project/PACKAGE-NAME/)
```

**npm** (JavaScript/TypeScript):
```markdown
[![npm version](https://img.shields.io/npm/v/PACKAGE-NAME)](https://www.npmjs.com/package/PACKAGE-NAME)
```

**Crates.io** (Rust):
```markdown
[![Crates.io](https://img.shields.io/crates/v/PACKAGE-NAME)](https://crates.io/crates/PACKAGE-NAME)
```

**Go** (no package registry badge):
```markdown
[![Go Reference](https://pkg.go.dev/badge/github.com/USER/REPO.svg)](https://pkg.go.dev/github.com/USER/REPO)
```

### License Badges

```markdown
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
```

### Language Badges

```markdown
[![Python](https://img.shields.io/badge/Python-3.12+-blue)](https://www.python.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.3+-blue)](https://www.typescriptlang.org/)
[![Go](https://img.shields.io/badge/Go-1.22+-blue)](https://golang.org/)
[![Rust](https://img.shields.io/badge/Rust-2021-orange)](https://www.rust-lang.org/)
```

### CI/CD Badges

**GitHub Actions**:
```markdown
[![Tests](https://github.com/USER/REPO/actions/workflows/test.yml/badge.svg)](https://github.com/USER/REPO/actions/workflows/test.yml)
```

**GitLab CI**:
```markdown
[![pipeline status](https://gitlab.com/USER/REPO/badges/main/pipeline.svg)](https://gitlab.com/USER/REPO/-/commits/main)
```

### Coverage Badges

**Codecov**:
```markdown
[![codecov](https://codecov.io/gh/USER/REPO/branch/main/graph/badge.svg)](https://codecov.io/gh/USER/REPO)
```

**Coveralls**:
```markdown
[![Coverage Status](https://coveralls.io/repos/github/USER/REPO/badge.svg?branch=main)](https://coveralls.io/github/USER/REPO?branch=main)
```

---

## Installation Command Templates

### Python

**pip (standard)**:
```bash
pip install PACKAGE-NAME

# With extras
pip install PACKAGE-NAME[dev]

# From Git
pip install git+https://github.com/USER/REPO.git

# Editable mode (development)
pip install -e .
```

**uv (recommended)**:
```bash
# Global tool installation
uv tool install PACKAGE-NAME

# Project dependency
uv add PACKAGE-NAME

# Dev dependency
uv add --dev PACKAGE-NAME

# From Git
uv add git+https://github.com/USER/REPO
```

### JavaScript/TypeScript

**npm**:
```bash
npm install PACKAGE-NAME

# Global
npm install -g PACKAGE-NAME

# Dev dependency
npm install --save-dev PACKAGE-NAME
```

**pnpm**:
```bash
pnpm add PACKAGE-NAME

# Global
pnpm add -g PACKAGE-NAME

# Dev dependency
pnpm add -D PACKAGE-NAME
```

**yarn**:
```bash
yarn add PACKAGE-NAME

# Global
yarn global add PACKAGE-NAME

# Dev dependency
yarn add --dev PACKAGE-NAME
```

### Go

```bash
# Install binary
go install github.com/USER/REPO@latest

# Add to project
go get github.com/USER/REPO

# Update
go get -u github.com/USER/REPO
```

### Rust

```bash
# Install binary
cargo install PACKAGE-NAME

# Install from Git
cargo install --git https://github.com/USER/REPO

# Add to project
cargo add PACKAGE-NAME
```

---

## Section Templates

### Problem-Solution Pattern

```markdown
## What is [Project]?

### The Problem: [Context Setting]

[2-3 sentences describing the pain point users face]

Examples of challenges:
- Challenge 1
- Challenge 2
- Challenge 3

### The Solution: [Your Approach]

[2-3 sentences explaining how your project solves these problems]

Key innovations:
- Innovation 1
- Innovation 2
- Innovation 3

### Core Promises

1. **Promise 1**: [Specific, measurable benefit]
2. **Promise 2**: [Specific, measurable benefit]
3. **Promise 3**: [Specific, measurable benefit]
```

### Feature Checklist Pattern

```markdown
## Features

### Core Capabilities
- ✅ **Feature 1**: Brief description
- ✅ **Feature 2**: Brief description
- ✅ **Feature 3**: Brief description

### Advanced Features
- ✅ **Advanced 1**: Description
- 🚧 **In Progress**: Work in progress feature
- 📋 **Planned**: Future roadmap item

### Integration
- ✅ **Integration 1**: Third-party tool
- ✅ **Integration 2**: Platform support
```

### Quick Start Pattern (3-Step)

```markdown
## Quick Start ([Total Time])

### Step 1: [Action] ([Time Estimate])

[Clear instructions with exact commands]

```bash
# Command 1
command-to-run

# Verify
verification-command
```

### Step 2: [Action] ([Time Estimate])

[Next step instructions]

```bash
# Command 2
next-command
```

### Step 3: [Action] ([Time Estimate])

[Final step instructions]

```bash
# Command 3
final-command
# Expected output
```

### Verify Installation

[How to confirm everything works]

```bash
# Health check
check-command
# ✅ Expected output
```
```

---

## Validation Rules

### File Size Constraints

**GitHub Limit**: 500 KiB (512,000 bytes)

```bash
# Check README size
FILE_SIZE=$(wc -c < README.md)
MAX_SIZE=512000

if [ "$FILE_SIZE" -gt "$MAX_SIZE" ]; then
    echo "⚠️  README exceeds GitHub limit: ${FILE_SIZE} bytes"
    echo "Recommendation: Move detailed content to docs/"
fi
```

### Required Sections

**Minimum viable README**:
- [ ] H1 title (project name)
- [ ] One-line description
- [ ] Installation instructions
- [ ] Basic usage example
- [ ] License

**Complete README**:
- [ ] Project overview (What/Why/How)
- [ ] Quick Start (< 5 minutes)
- [ ] Features list
- [ ] Usage examples
- [ ] Documentation links
- [ ] Community/support section

### Link Validation

**Prefer relative links**:
```markdown
✅ GOOD: [Contributing](CONTRIBUTING.md)
✅ GOOD: [API Docs](docs/API.md)
❌ BAD: [Contributing](https://github.com/user/repo/blob/main/CONTRIBUTING.md)
```

**Exception**: External resources (package registries, CI badges, etc.)

### Code Block Guidelines

**Always specify language**:
```markdown
✅ GOOD:
```bash
npm install package-name
```

❌ BAD:
```
npm install package-name
```
```

**Keep examples concise**:
- Basic examples: < 10 lines
- Advanced examples: < 20 lines
- Complex examples: Link to `examples/` directory

---

## Multi-Language Support

### Language Badge Row

```markdown
[English](README.md) | [한국어](README.ko.md) | [日本語](README.ja.md) | [中文](README.zh.md) | [Español](README.es.md)
```

### Language Detection Strategy

```pseudocode
1. Check for existing README translations
   - README.ko.md → Korean
   - README.ja.md → Japanese
   - README.zh.md → Chinese
   - README.es.md → Spanish
   - README.fr.md → French

2. If translations exist:
   - Add language badge row at top of README
   - Generate all translations consistently
   - Ensure section structure matches across languages

3. If no translations:
   - Generate English README only
   - Add comment: "<!-- Multi-language support: Add README.{lang}.md -->"
```

---

## Update Conflict Resolution

### Merge Strategy

**When updating existing README**:

```pseudocode
1. Parse existing README into sections
   sections = parseMarkdown(existingREADME)

2. Identify section types:
   for each section in sections:
       if section.title in STANDARD_SECTIONS:
           section.type = "standard"
       else:
           section.type = "custom"

3. Generate new sections:
   newSections = generateSections(metadata)

4. Merge:
   result = {}
   for section in STANDARD_SECTIONS:
       if section exists in sections:
           # Update if outdated
           if isOutdated(sections[section], newSections[section]):
               result[section] = newSections[section]
               preserveCustomSubsections(sections[section], result[section])
           else:
               result[section] = sections[section]
       else:
           # Add missing section
           result[section] = newSections[section]
   
   # Preserve custom sections
   for section in sections where type == "custom":
       result[section] = sections[section]

5. Reorder sections:
   orderedResult = reorderSections(result, BEST_PRACTICE_ORDER)

6. Output:
   write(orderedResult, "README.md")
```

### Outdated Detection

```pseudocode
function isOutdated(existingSection, newSection):
    # Version mismatch
    if existingVersion != currentVersion:
        return true
    
    # Missing dependencies
    if newDependencies not in existingDependencies:
        return true
    
    # Outdated badges
    if badgeVersion != currentVersion:
        return true
    
    return false
```

---

## Error Handling

### No Metadata Files

**Fallback Strategy**:
```
1. Search for README.md, README, readme.md
2. If exists: Extract project name from H1 title
3. Prompt user:
   - Project name?
   - Version?
   - Description?
   - Primary language?
4. Generate minimal README with placeholders
5. Add TODO comments for manual completion
```

### Conflicting Metadata

**Example**: Both `package.json` (Node) and `pyproject.toml` (Python) exist

**Resolution**:
```
1. Detect conflict: Multiple primary metadata files
2. Analyze project structure:
   - More files in src/*.py → Python
   - More files in src/*.ts → TypeScript
   - Check for main entry point
3. Ask user for confirmation:
   "Detected both Python and TypeScript. Primary language?"
4. Use confirmed language for README generation
```

### Complex Existing README

**When to suggest "update" mode**:
- Existing README > 10 KB
- Has > 10 custom sections
- Complex markdown (nested lists, tables, diagrams)
- Non-standard structure

**User prompt**:
```
⚠️  Existing README is complex (15 KB, 12 sections).

Recommendation: Use "update" mode to preserve custom content.

Options:
1. Update mode (merge new info, keep custom sections)
2. Create new (replace entire README)
3. Cancel

Choice: [1]
```

---

## References

- GitHub. "About READMEs." https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes
- Python Packaging Authority. "PEP 621 – Storing project metadata in pyproject.toml." https://peps.python.org/pep-0621/
- npm Docs. "package.json." https://docs.npmjs.com/cli/v10/configuring-npm/package-json
- Go Modules Reference. https://go.dev/ref/mod
- Cargo Book. "The Manifest Format." https://doc.rust-lang.org/cargo/reference/manifest.html
- Shields.io. "Badge generation service." https://shields.io/
