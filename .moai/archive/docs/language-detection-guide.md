# Language Detection Guide

## Overview

MoAI-ADK automatically detects programming languages in your project to provide language-specific CI/CD workflows, toolchain recommendations, and configuration.

## Supported Languages

MoAI-ADK detects **20 programming languages** with intelligent priority-based matching.

### Languages with Dedicated CI/CD Workflows (4)

These languages have fully configured GitHub Actions workflows with language-specific testing, linting, and coverage tools:

1. **Python** - `pyproject.toml`, `requirements.txt`, `setup.py`
2. **JavaScript** - `package.json`
3. **TypeScript** - `tsconfig.json` + `package.json`
4. **Go** - `go.mod`

### Additional Detected Languages (16)

These languages are detected but use generic workflows:

- **Ruby** - `Gemfile`, `.rb` files, `config/routes.rb` (Rails)
- **PHP** - `composer.json`, `artisan` (Laravel), `.php` files
- **Java** - `pom.xml`, `build.gradle`, `.java` files
- **Rust** - `Cargo.toml`, `.rs` files
- **Dart** - `pubspec.yaml`, `.dart` files
- **Swift** - `Package.swift`, `.swift` files
- **Kotlin** - `build.gradle.kts`, `.kt` files
- **C#** - `.csproj`, `.cs` files
- **Elixir** - `mix.exs`, `.ex` files
- **Scala** - `build.sbt`, `.scala` files
- **Clojure** - `project.clj`, `.clj` files
- **Haskell** - `.cabal`, `.hs` files
- **C** - `Makefile`, `.c` files
- **C++** - `CMakeLists.txt`, `.cpp` files
- **Lua** - `.lua` files
- **Shell** - `.sh`, `.bash` files

## How Detection Works

### Detection Algorithm

1. **Scans project directory** for language indicator files
2. **Uses priority-based matching** (framework-specific files ranked higher)
3. **Returns first matched language** from priority list
4. **Falls back to None** if no indicators found

### Detection Priority Rules

When multiple language indicators are present, MoAI-ADK follows this priority order:

```
Ruby/PHP (framework detection) > Python > TypeScript > JavaScript > Go > ...
```

**Example**:
- Project with `package.json` + `tsconfig.json` → **TypeScript** (not JavaScript)
- Project with `config/routes.rb` + `.rb` files → **Ruby** (Rails framework detected)
- Project with `artisan` + `composer.json` → **PHP** (Laravel framework detected)

## Usage

### Python API

```python
from moai_adk.core.project.detector import LanguageDetector

# Initialize detector
detector = LanguageDetector()

# Detect project language
language = detector.detect("/path/to/project")
print(f"Detected language: {language}")

# Detect multiple languages
languages = detector.detect_multiple("/path/to/project")
print(f"All languages: {languages}")

# Get workflow template path (for supported languages only)
if language in ["python", "javascript", "typescript", "go"]:
    workflow_path = detector.get_workflow_template_path(language)
    print(f"Workflow template: {workflow_path}")
```

### CLI Usage (via `/alfred:0-project`)

Language detection runs automatically during project initialization:

```bash
# MoAI-ADK detects language and configures appropriate workflows
/alfred:0-project
```

## Troubleshooting

### Problem: Detection returns None

**Cause**: No language indicator files found in project directory

**Solution**: Add at least one language indicator file:
- Python: `pyproject.toml` or `requirements.txt`
- JavaScript: `package.json`
- TypeScript: `package.json` + `tsconfig.json`
- Go: `go.mod`

### Problem: Wrong language detected

**Cause**: Multiple language indicators present, priority rules applied

**Solution**: Check priority order (Ruby/PHP > Python > TypeScript > JavaScript). Remove unnecessary indicator files or accept mixed-language project detection.

### Problem: TypeScript project detected as JavaScript

**Cause**: Missing `tsconfig.json`

**Solution**: Add `tsconfig.json` to project root. TypeScript detection requires both `package.json` and `tsconfig.json`.

## Advanced Features

### Package Manager Detection (JavaScript/TypeScript)

For JavaScript/TypeScript projects, MoAI-ADK automatically detects the package manager:

```python
detector = LanguageDetector()
package_manager = detector.detect_package_manager("/path/to/project")
# Returns: 'bun' | 'pnpm' | 'yarn' | 'npm'
```

**Priority order**: bun > pnpm > yarn > npm

**Detection method**: Checks for lock files (bun.lockb, pnpm-lock.yaml, yarn.lock, package-lock.json)

### Multi-Language Projects

For projects with multiple languages, `detect_multiple()` returns all detected languages:

```python
languages = detector.detect_multiple("/path/to/project")
# Example: ['python', 'javascript', 'shell']
```

Primary language is determined by `detect()` method (uses priority order).

## Performance Considerations

- **Detection speed**: ~10-50ms for typical projects
- **Large projects**: Consider caching detection results
- **Recursive scanning**: Scans all subdirectories for language indicators
- **Early termination**: Stops after first match in `detect()` mode

## Quick Reference

### Language Detection Cheat Sheet

| Language   | Primary Indicator     | Secondary Indicators           | Workflow Support |
|------------|----------------------|--------------------------------|------------------|
| Python     | `pyproject.toml`     | `requirements.txt`, `.py`      | ✅ Full          |
| TypeScript | `tsconfig.json`      | `package.json`, `.ts`          | ✅ Full          |
| JavaScript | `package.json`       | `.js`                          | ✅ Full          |
| Go         | `go.mod`             | `.go`                          | ✅ Full          |
| Ruby       | `Gemfile`            | `config/routes.rb`, `.rb`      | ⚠️ Generic       |
| PHP        | `composer.json`      | `artisan`, `.php`              | ⚠️ Generic       |
| Java       | `pom.xml`            | `build.gradle`, `.java`        | ⚠️ Generic       |
| Rust       | `Cargo.toml`         | `.rs`                          | ⚠️ Generic       |

### Common Detection Patterns

```python
# Single language detection (fastest)
language = detector.detect()

# Multiple language detection (comprehensive)
all_languages = detector.detect_multiple()

# Check workflow support
supported = detector.get_supported_languages_for_workflows()
# Returns: ['python', 'javascript', 'typescript', 'go']

# Get workflow path
if language in supported:
    workflow = detector.get_workflow_template_path(language)
```

## Related Documentation

- [Workflow Templates Guide](./workflow-templates.md) - CI/CD workflow configuration
- [tdd-implementer Agent](./.claude/agents/alfred/tdd-implementer.md) - Workflow generation
- [SPEC-LANGUAGE-DETECTION-001](./.moai/specs/SPEC-LANGUAGE-DETECTION-001/) - Original specification
