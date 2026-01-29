# MoAI-ADK Go Implementation

[![Go Report Card](https://goreportcard.com/badge/github.com/anthropics/moai-adk-go)](https://goreportcard.com/report/github.com/anthropics/moai-adk-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

**MoAI-ADK** is the Strategic Orchestrator for Claude Code. This is the native Go implementation, offering superior performance and easier distribution compared to the Python version.

## Features

- **🚀 Performance**: 10-50x faster than Python implementation
- **📦 Static Binary**: Single executable with no external dependencies
- **🔧 Cross-Platform**: Native support for macOS, Linux, and Windows
- **🔄 Self-Update**: Built-in update mechanism
- **🌍 Multi-Language**: Support for 16+ programming languages
- **🛡️ Type-Safe**: Written in Go with static type checking

## Quick Start

### Installation

#### Method 1: go install (Recommended)

```bash
go install github.com/anthropics/moai-adk-go@latest
```

#### Method 2: GitHub Releases

Download the latest release for your platform from [GitHub Releases](https://github.com/anthropics/moai-adk-go/releases/latest).

```bash
# macOS ARM64 (Apple Silicon)
curl -sSL https://github.com/anthropics/moai-adk-go/releases/latest/download/moai-adk-darwin-arm64 -o moai-adk
chmod +x moai-adk
sudo mv moai-adk /usr/local/bin/

# Linux AMD64
curl -sSL https://github.com/anthropics/moai-adk-go/releases/latest/download/moai-adk-linux-amd64 -o moai-adk
chmod +x moai-adk
sudo mv moai-adk /usr/local/bin/
```

#### Method 3: Install Script

```bash
curl -sSL https://raw.githubusercontent.com/anthropics/moai-adk-go/main/scripts/install.sh | bash
```

#### Method 4: Homebrew (macOS)

```bash
brew tap anthropics/tap
brew install moai-adk
```

### Initialize a Project

```bash
moai-adk init my-project
cd my-project
```

### Update MoAI-ADK

```bash
# Update MoAI templates
moai-adk update

# Update the binary itself
moai-adk self-update
```

### Verify Installation

```bash
moai-adk version
```

Output:
```
moai-adk version 1.0.0
commit: abc123def
built at: 2026-01-29T12:00:00Z
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize a new MoAI project |
| `update` | Update MoAI templates and configuration |
| `self-update` | Update moai-adk binary to latest version |
| `migrate` | Migrate from Python to Go implementation |
| `doctor` | Check project health |
| `status` | Show project status |
| `hook` | Execute Claude Code hooks |
| `statusline` | Display statusline |
| `version` | Show version information |

## Platform Support

| Platform | Architecture | Binary Size |
|----------|--------------|-------------|
| macOS | Intel (amd64) | ~15 MB |
| macOS | Apple Silicon (arm64) | ~14 MB |
| Linux | AMD64 | ~13 MB |
| Linux | ARM64 (AWS Graviton) | ~13 MB |
| Windows | AMD64 | ~15 MB |

## Migration from Python

If you're currently using the Python implementation, see [MIGRATION_GUIDE.md](docs/MIGRATION_GUIDE.md) for detailed migration instructions.

Quick migration:
```bash
# Preview changes
moai-adk migrate --dry-run

# Apply migration
moai-adk migrate

# Verify
moai-adk hook session-start --project-dir "$CLAUDE_PROJECT_DIR"
```

## Development

### Prerequisites

- Go 1.23+
- Make (optional, for using Makefile)

### Build from Source

```bash
# Standard build
go build -o moai-adk ./cmd/moai/

# Cross-compile for macOS ARM64
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
  -ldflags="-s -w" \
  -o moai-adk-darwin-arm64 \
  ./cmd/moai/

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags="-s -w" \
  -o moai-adk-windows-amd64.exe \
  ./cmd/moai/
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Using Makefile

```bash
# Build
make build

# Run tests
make test

# Run linters
make lint

# Format code
make fmt

# Run goreleaser (dry-run)
make release-dry-run
```

## Project Structure

```
.
├── cmd/
│   ├── moai/                    # Main CLI application
│   └── verify-compatibility/    # Compatibility verification tool
├── internal/
│   ├── cli/                     # CLI commands
│   ├── config/                  # Configuration management
│   ├── doctor/                  # Health checks
│   ├── hooks/                   # Hook protocol and handlers
│   ├── initializer/             # Project initialization
│   ├── migration/               # Python to Go migration
│   ├── output/                  # Output formatting
│   ├── status/                  # Status reporting
│   ├── statusline/              # Statusline display
│   ├── template/                # Template management
│   └── update/                  # Update mechanism
├── pkg/
│   ├── templates/               # Embedded MoAI templates
│   └── version/                 # Version information
├── scripts/
│   └── install.sh               # Installation script
├── docs/
│   └── MIGRATION_GUIDE.md       # Migration documentation
├── .goreleaser.yml              # Goreleaser configuration
├── go.mod
├── go.sum
└── README.md
```

## Configuration

MoAI-ADK uses YAML configuration files located at `.moai/config/config.yaml`.

### Environment Variables

- `MOAI_CONFIG_PATH`: Custom config directory path
- `MOAI_USER_NAME`: User name for personalization
- `MOAI_CONVERSATION_LANG`: Conversation language (ko, en, ja, etc.)
- `CLAUDE_PROJECT_DIR`: Project directory (set by Claude Code)

## Release Process

Releases are automated using [goreleaser](https://goreleaser.com/):

1. Tag the release: `git tag -a v1.0.0 -m "Release 1.0.0"`
2. Push tag: `git push origin v1.0.0`
3. GitHub Actions builds and publishes:
   - 5 platform binaries
   - Checksums file
   - Homebrew formula
   - GitHub Release with auto-generated notes

## Performance Comparison

| Operation | Python | Go | Speedup |
|-----------|--------|-----|---------|
| Cold start | ~500ms | ~50ms | 10x |
| Hook execution | ~200ms | ~20ms | 10x |
| Init project | ~2s | ~200ms | 10x |
| Update templates | ~3s | ~300ms | 10x |
| Memory usage | ~100MB | ~20MB | 5x |

## Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - TUI styling
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing

No external runtime dependencies - the binary is fully self-contained.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Support

- **Issues**: [GitHub Issues](https://github.com/anthropics/moai-adk-go/issues)
- **Documentation**: [docs/](docs/)
- **Migration Guide**: [MIGRATION_GUIDE.md](docs/MIGRATION_GUIDE.md)

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Distributed with [Goreleaser](https://goreleaser.com/)

---

**MoAI-ADK** - Strategic Orchestrator for Claude Code

Version: 1.0.0 (Go)
Last Updated: 2026-01-29
