# MoAI-ADK Technology Stack

> **Last Updated**: 2026-01-10
> **Version**: 0.41.2

---

## Runtime Environment

### Python
- **Supported Versions**: Python 3.11, 3.12, 3.13, 3.14
- **Target Version**: Python 3.14 (for development)
- **Package Manager**: uv (recommended), pip (supported)

### Node.js (Web UI)
- **Version**: Node.js 20+ LTS
- **Package Manager**: pnpm

---

## Core Dependencies

### CLI Framework
| Package | Version | Purpose |
|---------|---------|---------|
| click | >=8.1.0 | CLI command framework |
| rich | >=13.0.0 | Rich text formatting and terminal UI |
| pyfiglet | >=1.0.2 | ASCII art text generation |
| questionary | >=2.0.0 | Interactive CLI prompts |
| InquirerPy | >=0.3.4 | Modern CLI prompts with fuzzy search |

### Git Integration
| Package | Version | Purpose |
|---------|---------|---------|
| gitpython | >=3.1.45 | Git operations and repository management |

### Configuration & Templating
| Package | Version | Purpose |
|---------|---------|---------|
| pyyaml | >=6.0 | YAML parsing and serialization |
| jinja2 | >=3.0.0 | Template engine for code generation |

### HTTP & Async
| Package | Version | Purpose |
|---------|---------|---------|
| requests | >=2.28.0 | HTTP client for REST API |
| aiohttp | >=3.13.2 | Async HTTP client |

### System & Utilities
| Package | Version | Purpose |
|---------|---------|---------|
| psutil | >=7.1.3 | System and process utilities |
| packaging | >=21.0 | Version parsing and comparison |

### AI Integration
| Package | Version | Purpose |
|---------|---------|---------|
| google-genai | >=1.0.0 | Gemini API for image generation (nano-banana) |
| pillow | >=10.0.0 | Image processing |
| anthropic | >=0.40.0 | Claude API SDK (optional, for web chat) |

---

## Development Dependencies

### Testing
| Package | Version | Purpose |
|---------|---------|---------|
| pytest | >=8.4.2 | Test framework |
| pytest-cov | >=7.0.0 | Coverage reporting |
| pytest-xdist | >=3.8.0 | Parallel test execution |
| pytest-asyncio | >=1.2.0 | Async test support |
| pytest-mock | >=3.15.1 | Mocking support |

### Code Quality
| Package | Version | Purpose |
|---------|---------|---------|
| ruff | >=0.1.0 | Fast Python linter and formatter |
| mypy | >=1.7.0 | Static type checking |
| black | >=24.0.0 | Code formatting |
| types-PyYAML | >=6.0.0 | Type stubs for PyYAML |

### Security
| Package | Version | Purpose |
|---------|---------|---------|
| pip-audit | >=2.7.0 | Dependency vulnerability scanning |
| bandit | >=1.8.0 | Security linter for Python |

---

## Web UI Dependencies

### Backend (FastAPI)
| Package | Version | Purpose |
|---------|---------|---------|
| fastapi | >=0.115.0 | Web framework |
| uvicorn | >=0.34.0 | ASGI server |
| websockets | >=14.0 | WebSocket support |
| aiosqlite | >=0.20.0 | Async SQLite |
| pydantic | >=2.9.0 | Data validation |
| ptyprocess | >=0.7.0 | PTY for terminal integration |

### Frontend (Next.js)
| Technology | Version | Purpose |
|------------|---------|---------|
| Next.js | 16.x | React framework with App Router |
| React | 19.x | UI library |
| TypeScript | 5.9+ | Type-safe JavaScript |
| Tailwind CSS | 3.x | Utility-first CSS |
| shadcn/ui | latest | UI component library |

---

## External Tools

### AST-Grep
- **Tool**: ast-grep (sg CLI)
- **Purpose**: Structural code search and refactoring
- **Languages**: 40+ programming languages supported
- **Usage**: Security scanning, code pattern matching, large-scale refactoring

### Claude Code
- **Platform**: Anthropic Claude Code CLI
- **Version**: Latest
- **Purpose**: AI-powered development assistant
- **Integration**: Via `.claude/` configuration directory

---

## Build System

### Package Building
- **Build Backend**: hatchling
- **Configuration**: pyproject.toml (PEP 517/518)

### Distribution
- **Package Name**: moai-adk
- **CLI Entry Points**:
  - `moai-adk` - Main CLI
  - `moai` - Alias for moai-adk
  - `moai-worktree` - Git worktree management

---

## Quality Configuration

### Ruff (Linting)
```toml
line-length = 120
target-version = "py314"
select = ["E", "F", "W", "I", "N"]
```

### Coverage
```toml
fail_under = 85
precision = 2
show_missing = true
```

### MyPy
```toml
python_version = "3.14"
ignore_missing_imports = true
```

---

## Architecture Patterns

### Agent Architecture (5-Tier)
1. **Tier 0**: Core agents (Alfred orchestrator)
2. **Tier 1**: Manager agents (workflow coordination)
3. **Tier 2**: Expert agents (domain specialists)
4. **Tier 3**: Builder agents (creation specialists)
5. **Tier 4**: Utility agents (helper functions)

### Skill Architecture
- **Foundation Skills**: Core principles and patterns
- **Domain Skills**: Frontend, Backend, Database
- **Language Skills**: Python, TypeScript, Go, Rust, etc.
- **Platform Skills**: Firebase, Supabase, Vercel, etc.
- **Workflow Skills**: TDD, SPEC, Project management

### Configuration Architecture
- **Modular Sections**: Separated by concern (user, language, project, git, quality)
- **YAML Format**: Human-readable configuration
- **Template Variables**: `{{VARIABLE}}` for distribution
- **Runtime Resolution**: Environment variables and dynamic values

---

## Integration Points

### Claude Code Integration
- Custom agents in `.claude/agents/`
- Slash commands in `.claude/commands/`
- Hooks in `.claude/hooks/`
- Skills in `.claude/skills/`

### GitHub Integration
- Workflow templates for CI/CD
- SPEC to GitHub Issue sync
- Branch protection and PR automation

### LSP Integration
- Language Server Protocol for diagnostics
- Real-time error detection
- Auto-fix suggestions

---

## Security Considerations

### Dependency Security
- Regular `pip-audit` scanning
- Bandit static analysis
- GitHub Dependabot alerts

### Code Security
- AST-grep security patterns
- OWASP vulnerability detection
- XSS, CSRF, SQL injection prevention

### Configuration Security
- `.env` files for secrets
- Git-ignored sensitive files
- Secure file permissions (0o600)

---

## Performance Optimizations

### Token Efficiency
- Conditional skill auto-loading
- Quick reference for simple tasks
- Context compaction for long sessions

### Build Performance
- Parallel test execution (pytest-xdist)
- Incremental type checking (mypy)
- Fast linting (ruff over flake8)

### Runtime Performance
- Async operations where possible
- Caching with configurable TTL
- Lazy loading of resources
