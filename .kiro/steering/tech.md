# MoAI-ADK Technical Stack

## Primary Technology Stack

### Language & Runtime
- **Python 3.11+**: Primary language with Claude Code native integration
- **Click**: CLI framework for command-line interfaces
- **Pydantic**: Data validation and serialization
- **asyncio**: Asynchronous agent execution and parallel processing

### Development Tools
- **pytest**: Testing framework with asyncio support
- **ruff**: Unified linting (replaces black, isort, flake8, pyupgrade)
- **mypy**: Static type checking with strict mode
- **bandit**: Security vulnerability scanning

### Build System
- **pyproject.toml**: Modern Python packaging configuration
- **setuptools + wheel**: Package building and distribution
- **pip**: Dependency management

### External Integrations
- **GitPython**: Git repository operations
- **PyGithub**: GitHub API integration for PR automation
- **watchdog**: File system monitoring for auto-checkpoints
- **structlog**: Structured logging for agent activities

## Common Commands

### Development Setup
```bash
# Install in development mode
pip install -e .

# Install with development dependencies
pip install -e .[dev]

# Run tests with coverage
pytest --cov=moai_adk --cov-report=html tests/

# Lint and format code
ruff check src/ tests/
ruff format src/ tests/

# Type checking
mypy src/
```

### Build & Release
```bash
# Build package
make build

# Clean build artifacts
make clean

# Version management
make version-check
make version-bump-patch
make version-bump-minor

# Run full test suite
make test
```

### Quality Assurance
```bash
# Security scan
bandit -r src/ -f json -o security-report.json

# Check TRUST 5 principles compliance
python .moai/scripts/check_constitution.py

# Update TAG traceability
python .moai/scripts/check-traceability.py --update
```

## Architecture Patterns

### Design Principles
- **Hexagonal Architecture**: External dependencies isolated through interfaces
- **CQRS**: Command and Query responsibility separation
- **Event Sourcing**: @TAG system tracks all development events
- **Plugin Architecture**: Extensible agent system

### Code Quality Standards
- **Test Coverage**: Minimum 85% required
- **Type Hints**: 95%+ coverage with mypy strict mode
- **Cyclomatic Complexity**: Maximum 10 per function
- **Line Length**: 88 characters (Black standard)
- **Function Length**: Maximum 50 lines
- **File Length**: Maximum 300 lines

### Security Requirements
- **Secret Management**: Environment variables and secure file storage
- **Input Validation**: Pydantic models for all data structures
- **Logging**: Structured logging with sensitive data masking
- **Dependencies**: Regular security scanning with bandit and safety

## Multi-Language Support Strategy

While Python is primary, MoAI-ADK supports detection and basic workflows for:
- **JavaScript/TypeScript**: Node.js, npm/pnpm, jest testing
- **Go**: Go modules, go test, golangci-lint
- **Java/Kotlin**: Gradle, JUnit, detekt/SpotBugs
- **Rust**: Cargo, cargo test, clippy
- **.NET**: dotnet CLI, xUnit, Roslyn analyzers

## Performance Guidelines

### Async Processing
- Use `asyncio.gather()` for parallel agent execution
- Maximum 10 concurrent agents to prevent resource exhaustion
- Implement proper error handling with `return_exceptions=True`

### Caching Strategy
- Cache parsed SPEC documents using content hashing
- LRU cache for frequently accessed configuration
- Incremental TAG index updates to minimize I/O

### Memory Management
- Use `aiofiles` for large file operations
- Implement proper cleanup in temporary file handling
- Monitor agent memory usage and implement limits