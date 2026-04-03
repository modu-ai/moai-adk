# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-02-20

---

## 1. Quick Start

### Work Location
```bash
# Primary work location (template development)
/Users/goos/MoAI/moai-adk-go/internal/template/templates/

# Local project (testing & git)
/Users/goos/MoAI/moai-adk-go/
```

### Development Cycle
```
1. Work in internal/template/templates/
2. Run `make build` to regenerate embedded files
3. Test in local project
4. Git commit from local root
```

### [CRITICAL] moai CLI vs /moai Slash Command

**DO NOT CONFUSE** these two completely different things:

| | `moai` (Terminal CLI) | `/moai` (Slash Command) |
|---|---|---|
| **Where** | Terminal shell | Claude Code chat input |
| **What** | Go binary (`~/go/bin/moai`) | Claude Code skill invocation |
| **Purpose** | Project setup, template deployment | AI-assisted development workflows |
| **Example** | `moai init myproject` | `/moai plan "add auth"` |
| **Scope** | File system operations | AI agent orchestration |

**Terminal `moai` commands:**
```bash
moai init <project>     # Initialize new project with templates
moai update             # Sync templates to current project
moai build              # Build embedded templates
moai hook <event>       # Execute hook handler
moai glm                # GLM worker mode
moai version            # Show version
```

**Claude Code `/moai` commands:**
```
/moai plan "feature"    # Create SPEC document
/moai run SPEC-XXX      # Implement SPEC
/moai sync SPEC-XXX     # Generate docs & PR
/moai fix               # Auto-fix errors
/moai loop              # Iterative fix loop
/moai project           # Generate project docs
/moai feedback          # Create GitHub issue
```

**Common mistake to avoid:**
- WRONG: Running `/moai init` in Claude Code chat (not a valid slash command)
- CORRECT: Running `moai init` in terminal
- WRONG: Running `moai plan` in terminal (not a CLI command)
- CORRECT: Running `/moai plan` in Claude Code chat

---

## 2. File Synchronization

### Protected Directories (Never Modify During Template Sync)
```bash
# CRITICAL: These directories contain user data and must NEVER be deleted
.claude/        # Local Claude Code configuration
.moai/project/  # Project documentation (product.md, structure.md, tech.md)
.moai/specs/    # SPEC documents (active development files)
```

### Template Source (Single Source of Truth)
```bash
# All template changes MUST be made here
internal/template/templates/.claude/
internal/template/templates/.moai/
internal/template/templates/.agency/
internal/template/templates/CLAUDE.md
```

### [HARD] Template-First Rule

When adding new files to `.claude/`, `.moai/`, or `.agency/`:

1. **Add to template FIRST**: `internal/template/templates/<path>`
2. **Run `make build`** to regenerate embedded files
3. **Then sync to local**: `moai update` or manual copy

Never add files directly to the local project directories without also adding them to the template source. This includes:
- New agents (`.claude/agents/`)
- New skills (`.claude/skills/`)
- New commands (`.claude/commands/`)
- New rules (`.claude/rules/`)
- New config files (`.moai/config/`)
- New agency files (`.agency/`)

**Verification**: Before committing, check that every new file under `.claude/`, `.moai/`, or `.agency/` has a corresponding file in `internal/template/templates/`.

### Local-Only Files (Never in Templates)
```
.claude/settings.local.json    # Personal settings
.claude/settings.json          # Rendered from .json.tmpl
.claude/agent-memory/          # Per-project agent memory
.claude/hooks/moai/handle-*.sh # Generated hook wrappers (not templates)
.claude/commands/98-*.md       # Dev-project-specific commands
.claude/commands/99-*.md       # Dev-project-specific commands
CLAUDE.local.md                # This file
.moai/cache/                   # Cache
.moai/logs/                    # Logs
.moai/state/                   # Session state storage
.moai/specs/                   # Active SPEC documents
.moai/plans/                   # Session plans
.moai/reports/                 # Generated reports
.moai/manifest.json            # Generated at runtime
.moai/status_line.sh           # Rendered from .sh.tmpl
```

### Embedded Template System

moai-adk-go uses Go's `go:embed` directive:
- **Source**: `internal/template/templates/` (edit here)
- **Generated**: `internal/template/embedded.go` (auto-generated, DO NOT EDIT)
- **Build**: Run `make build` after editing templates

---

## 3. Code Standards

### Language: English Only

**Source Code (Go):**
- All code, comments, godoc in English
- Package names: lowercase, single word
- Exported names: PascalCase
- Private names: camelCase
- Constants: PascalCase or UPPER_SNAKE_CASE
- Commit messages: English (Conventional Commits)

**Configuration Files (English ONLY):**
- Command files (.claude/commands/**/*.md): English only
- Agent definitions (.claude/agents/**/*.md): English only
- Skill definitions (.claude/skills/**/*.md): English only
- Hook scripts (.claude/hooks/**/*.sh): English only
- CLAUDE.md: English only

**Why**: Command/agent/skill files are code, not user-facing content. They are read by Claude Code (English-based) and must be in English for consistent behavior.

**User-facing vs Internal:**
- User-facing: README, CHANGELOG, documentation (can be localized)
- Internal: Commands, agents, skills, hooks (MUST be English)

### Go-Specific Standards

**File Naming:**
- Go files: `snake_case.go` (e.g., `template_deployer.go`)
- Test files: `snake_case_test.go` (e.g., `settings_test.go`)

**Error Handling:**
- Always wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Use error wrapping, not string concatenation
- All godoc comments in English

---

## 4. Git Workflow

### Before Commit
- [ ] Code in English
- [ ] Tests passing (`go test ./...`)
- [ ] Linting passing (`golangci-lint run`)
- [ ] Templates regenerated (`make build`)

### Before Push
- [ ] Branch rebased
- [ ] Commits organized
- [ ] Commit messages follow format (Conventional Commits)

### Commit Message Format
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:** feat, fix, docs, style, refactor, perf, test, chore, revert

**Examples:**
```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
```

---

## 5. Version Management

### Single Source of Truth

- [HARD] `go.mod` module version + git tags are the authoritative sources
- [HARD] `pkg/version/version.go` reads from git tags at build time

**Version Reference:**
- Authoritative Source: Git tags (e.g., `v1.0.0`)
- Runtime Access: `pkg/version/version.go` via `git describe`
- Config Display: `.moai/config/sections/system.yaml` (updated by release process)

### Build Version Injection

Version is injected at build time using ldflags:

```bash
# Build with version injection
go build -ldflags="-X github.com/modu-ai/moai-adk/pkg/version.Version=v1.0.0"

# Makefile handles this automatically
make build VERSION=1.0.0
```

### Files Requiring Version Sync

When releasing new version, update:

**Documentation Files:**
- README.md (Version line)
- README.ko.md (Version line)
- CHANGELOG.md (New version entry)

**Configuration Files:**
- .moai/config/sections/system.yaml (moai.version)
- internal/template/templates/.moai/config/config.yaml (moai.version)

### Release Process

1. Update CHANGELOG.md with new version entry
2. Create git tag: `git tag v1.0.0`
3. Push tag: `git push origin v1.0.0`
4. Build release binaries: `make release VERSION=1.0.0`

---

## 6. Testing Guidelines

### ⚠️ IMPORTANT: Prevent Accidental File Modifications

When running tests, **always check if they modify project files**.

### Test Execution
```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestEnsureGlobalSettingsEnv ./internal/cli/
```

### Test Isolation

**[HARD] All test temp directories MUST be created under `/tmp` and cleaned up automatically.**

Use `t.TempDir()` for all temporary directories. It creates dirs under `os.TempDir()` and registers automatic cleanup.

```go
func TestSomething(t *testing.T) {
    tempDir := t.TempDir()  // Auto-cleanup after test - ALWAYS use this
    // Work in tempDir instead of project root
}
```

**Why this matters - `filepath.Join` vs absolute paths:**

On macOS, `t.TempDir()` returns paths starting with `/var/folders/...`.
Go's `filepath.Join(cwd, absPath)` does NOT strip the leading `/` from the second arg:
```
filepath.Join("/a/b", "/var/folders/x") = "/a/b/var/folders/x"  // WRONG!
filepath.Abs("/var/folders/x") = "/var/folders/x"                // CORRECT
```

Always use `filepath.Abs()` when resolving user-supplied paths in CLI commands.
Never use `filepath.Join(cwd, userPath)` when `userPath` can be absolute.

### Coverage Targets

- Package-level: 85% minimum coverage
- Critical packages (cli, template, hook): 90%+ coverage

### Go Test Execution Rules

- [HARD] After fixing ANY test, run the FULL test suite (`go test ./...`) to catch cascading failures
- Do not declare success after fixing only the initially failing tests
- Run `go test -count=1 ./...` to disable test caching when debugging flaky tests
- Run `go test -race ./...` for concurrency safety on any code touching goroutines or channels
- Run `go vet ./...` before committing to catch static analysis issues

### Table-Driven Tests (Go Convention)

```go
func TestBuildRequiredPATH(t *testing.T) {
    tests := []struct {
        name    string
        goBin   string
        goPath  string
        want    string
    }{
        {"default", "", "", wantDefault},
        {"custom bin", "/custom/bin", "", wantCustom},
        {"custom path", "", "/custom/path", wantPath},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## 7. Hook Development Guidelines

### [HARD] Shell Script Hooks Only

moai-adk-go uses shell scripts for hooks, NOT Python:

**Hook Wrapper Pattern:**
```bash
#!/bin/bash
# .claude/hooks/moai/handle-session-start.sh

# Read stdin JSON from Claude Code
INPUT=$(cat)

# Call moai binary with hook subcommand
moai hook session-start <<< "$INPUT"
```

**Why Shell Scripts:**
- Faster execution (no Python startup overhead)
- Always available (no dependency on uv/python)
- Cross-platform (bash, /bin/sh)

### Hook Command Format

**settings.json hook configuration:**
```json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
        "timeout": 5
      }]
    }]
  }
}
```

**Key Rules:**
- [HARD] Always quote `$CLAUDE_PROJECT_DIR`: `"$CLAUDE_PROJECT_DIR"`
- [HARD] Use full path to hook wrapper script
- [HARD] Set appropriate timeout (default: 5 seconds)

### Platform Differences

**macOS/Linux:**
```json
"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""
```

**Windows:**
```json
"command": "\"%CLAUDE_PROJECT_DIR%\\.claude\\hooks\\moai\\hook.sh\""
```

---

## 8. Template Variable Strategy

### Template vs Local Settings

moai-adk-go uses different path variable strategies:

**Template settings** (`internal/template/templates/.claude/settings.json`):
- Uses: `{{.GoBinPath}}` template variable (Go template syntax)
- Purpose: Runtime rendering during `moai init`
- Cross-platform: Resolved by `template.TemplateContext`

**Local settings** (`~/.claude/settings.json`):
- Uses: `"$CLAUDE_PROJECT_DIR"` environment variable
- Purpose: Runtime path resolution by Claude Code
- Cross-platform: Automatically resolved by Claude Code

### Template Variables

Available in Go templates (`*.tmpl` files):

```go
type TemplateContext struct {
    GoBinPath string  // Path to Go bin directory
    HomeDir   string  // User home directory
}
```

**Usage in templates:**
```go
// .moai/status_line.sh.tmpl
export PATH="{{.GoBinPath}}:$PATH"
```

**Rendering:**
```go
ctx := template.NewTemplateContext(
    template.WithGoBinPath(detectGoBinPath()),
    template.WithHomeDir(homeDir),
)
deployer.Deploy(ctx, projectRoot, mgr, ctx)
```

---

## 9. Configuration System

### Config File Format

moai-adk-go uses YAML for configuration:

**Project config** (`.moai/config/config.yaml`):
- Main configuration file
- Contains sections for different settings

**Section files** (`.moai/config/sections/*.yaml`):
- `config.yaml` - Main config
- `quality.yaml` - Quality gates, development mode
- `language.yaml` - Language preferences
- `user.yaml` - User information
- `workflow.yaml` - Workflow settings

### Configuration Priority

1. Environment Variables: `MOAI_USER_NAME`, `MOAI_CONVERSATION_LANG`
2. User Configuration: `.moai/config/sections/*.yaml`
3. Template Defaults: From `internal/template/templates/.moai/config/`

---

## 10. Build and Development Commands

### Common Commands

```bash
# Build the project
make build

# Run tests
make test

# Run with race detection
make test-race

# Run linter
make lint

# Format code
make fmt

# Install locally
make install

# Clean build artifacts
make clean

# Run go fix modernizers
make fix
```

### Development Workflow

```bash
# 1. Edit templates
vim internal/template/templates/.claude/skills/moai/SKILL.md

# 2. Regenerate embedded files
make build

# 3. Run tests
go test ./internal/template/...

# 4. Test locally
./moai init test-project

# 5. Commit
git add internal/template/templates/
git commit -m "feat(template): update SKILL.md"
```

---

## 11. Directory Structure

```
moai-adk-go/
├── cmd/                        # Main application entry points
│   └── moai/                   # CLI command
│       └── main.go             # Entry point
├── internal/                   # Private application code
│   ├── cli/                    # CLI commands
│   │   ├── init.go             # moai init command
│   │   ├── update.go           # moai update command
│   │   └── ...
│   ├── core/                   # Core business logic
│   │   └── project/            # Project management
│   ├── foundation/             # Foundation utilities
│   ├── hook/                   # Hook system
│   ├── manifest/               # Template manifest
│   ├── merge/                  # 3-way merge
│   ├── template/               # Template system
│   │   ├── templates/          # SOURCE: Edit templates here ⭐
│   │   │   ├── .claude/        # Claude Code config templates
│   │   │   │   ├── agents/     # Agent definitions
│   │   │   │   ├── commands/   # Slash commands
│   │   │   │   ├── hooks/      # Hook scripts
│   │   │   │   ├── output-styles/ # Output styles
│   │   │   │   ├── rules/      # Rules
│   │   │   │   └── skills/     # Skill definitions
│   │   │   ├── .moai/          # MoAI config templates
│   │   │   │   └── config/     # Config templates
│   │   │   ├── CLAUDE.md       # Main execution directives
│   │   │   └── *.tmpl          # Template files
│   │   ├── deployer.go         # Template deployment
│   │   ├── renderer.go         # Template rendering
│   │   ├── settings.go         # settings.json generation
│   │   └── embedded.go         # Generated: DO NOT EDIT
│   └── ...
├── pkg/                        # Public libraries
│   ├── models/                 # Data models
│   └── version/                # Version info
├── .claude/                    # Local Claude Code config (NOT in template)
├── .moai/                      # Local MoAI state (NOT in template)
├── CLAUDE.md                   # Synced from templates
├── CLAUDE.local.md             # This file (local only)
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── Makefile                    # Build commands
└── README.md                   # Project documentation
```

---

## 12. Frequent Issues and Solutions

### Issue: Templates not updated after editing

**Solution:**
```bash
# Regenerate embedded files
make build

# Verify
ls -la internal/template/embedded.go
```

### Issue: Tests modify ~/.claude/settings.json

**Solution:** Tests should use `t.TempDir()` for isolation. Check if test creates files in project root.

### Issue: Hook timeout

**Solution:** Increase timeout in settings.json:
```json
{"timeout": 60}  # 60 seconds instead of default 5
```

---

## 13. Reference

- CLAUDE.md: Alfred execution directives
- README.md: Project overview
- Skill("moai-foundation-core"): Execution rules
- Skill("moai-foundation-claude"): Plugin development, sandboxing
- Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments
- Effective Go: https://go.dev/doc/effective_go

---

## 15. Multi-Model Architecture (Claude Code 2.1.50+)

### Three Distinct Concepts

```
┌─────────────────────────────────────────────────────────────────┐
│  1. Model Policy (moai init --model-policy)                     │
│  ├── CLI sets each agent's model field individually             │
│  ├── Source: internal/template/model_policy.go                  │
│  └── Mapping: [high_model, medium_model, low_model] per agent   │
├─────────────────────────────────────────────────────────────────┤
│  2. Model Field (Agent Definition)                              │
│  ├── Values: inherit, opus, sonnet, haiku                       │
│  ├── NEVER use: glm, high, medium, low                          │
│  └── Set by: moai init --model-policy or manual edit            │
├─────────────────────────────────────────────────────────────────┤
│  3. CG Mode (CLI Commands)                                      │
│  ├── moai cc: Claude-only                                       │
│  ├── moai glm: GLM-only                                         │
│  └── moai cg: Claude Leader + GLM Teammates (tmux isolation)    │
└─────────────────────────────────────────────────────────────────┘
```

### Model Policy Reference

```bash
moai init --model-policy high      # opus/sonnet/haiku per agent
moai init --model-policy medium    # opus/sonnet/haiku per agent (default)
moai init --model-policy low       # sonnet/haiku only (no opus)
moai update -c --model-policy high # Update existing project
```

Key agent mappings (see model_policy.go for full list):
- Always opus (high/medium): manager-spec, manager-strategy, expert-security
- Always haiku (all policies): manager-quality, manager-git, team-researcher, team-quality
- manager-docs: sonnet/haiku (docs are lightweight)

### GLM Configuration

GLM is configured via environment variable overrides in ~/.claude/settings.json:
```json
{"env": {
  "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.7-air",
  "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
  "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1"
}}
```

Reference: https://docs.z.ai/devpack/tool/claude

### Mode Selection Matrix

| Command | Leader | Workers | Use Case |
|---------|--------|---------|----------|
| `moai cc` | Claude | Claude | Complex work, high quality |
| `moai glm` | GLM | GLM | Cost optimization |
| `moai cg` | Claude | GLM | Best balance (tmux isolation) |

### Agent Definition Pattern

```yaml
# CORRECT
model: inherit              # Uses user's choice or GLM (CG/GLM mode)
# model: opus              # Also OK (set by model_policy.go)

# WRONG
model: glm                  # NEVER: GLM is not a model field value
model: high                 # NEVER: This is a CLI flag, not a model value
```

---

## 16. Claude Code 2.1.50 Worktree Integration

MoAI-ADK uses two complementary worktree systems:
- **Claude Native** (`.claude/worktrees/`): Ephemeral, session-scoped, used by agents with `isolation: worktree`
- **MoAI Worktree** (`.moai/worktrees/`): Persistent, SPEC-scoped, used for multi-session development

For complete details including agent configuration, development checklist, and troubleshooting, see @.claude/rules/moai/workflow/worktree-integration.md.

---

## 17. iTerm2 Notification (작업 완료 알림)

### Claude Code Hooks로 알림 소리 설정

Claude Code의 `Notification` 이벤트를 활용하여 작업 완료 시 macOS 시스템 사운드를 재생한다.
`Notification` 이벤트는 Claude Code가 사용자 입력을 기다릴 때 (작업 완료 포함) 발생한다.

**`.claude/settings.local.json`에 추가:**
```json
{
  "hooks": {
    "Notification": [{
      "hooks": [{
        "type": "command",
        "command": "afplay /System/Library/Sounds/Glass.aiff",
        "timeout": 5
      }]
    }]
  }
}
```

### 사용 가능한 시스템 사운드

```bash
ls /System/Library/Sounds/
# Glass.aiff, Ping.aiff, Pop.aiff, Purr.aiff,
# Sosumi.aiff, Submarine.aiff, Tink.aiff
```

### 대안: iTerm2 Triggers

iTerm2 → Settings → Profiles → Advanced → Triggers:
- Regular Expression: 완료 메시지 패턴
- Action: Run Command
- Parameters: `afplay /System/Library/Sounds/Glass.aiff`

---

## 18. Claude Code YAML Frontmatter Guide

### [HARD] Claude Code Frontmatter Format Rules

Claude Code의 rule, agent, skill 파일에서 YAML frontmatter를 작성할 때 반드시 지켜야 하는 규칙들.

**배경**: Claude Code의 내부 YAML 파서는 일부 필드에서 YAML 배열 지원을 개선하고 있다. v2.1.84 이상에서는 여러 필드가 YAML 배열을 지원한다.

### Rules (.claude/rules/**/*.md)

**`paths` 필드**: CSV 문자열 권장 (호환성), YAML 배열도 지원됨 (v2.1.84+).

```yaml
# RECOMMENDED - CSV string
---
paths: "**/*.go,**/go.mod,**/go.sum"
---

# ALSO OK - YAML array (supported since v2.1.84)
---
paths:
  - "**/*.go"
  - "**/go.mod"
---
```

MoAI convention: 기존 규칙과의 일관성을 위해 CSV 형식 계속 사용.

### Agents (.claude/agents/**/*.md)

**`tools` 필드**: 반드시 CSV 문자열 사용. YAML 배열 사용 금지.

```yaml
# CORRECT
tools: Read, Write, Edit, Grep, Glob, Bash

# WRONG
tools:
  - Read
  - Write
  - Edit
```

**`skills` 필드**: YAML 배열 사용 (예외적으로 배열이 정상 동작).

```yaml
# CORRECT - skills는 YAML 배열
skills:
  - moai-lang-go
  - moai-domain-backend
```

**`model` 필드 값**: `inherit`, `opus`, `sonnet`, `haiku` 중 하나만 사용.

**`permissionMode` 필드 값**: `default`, `acceptEdits`, `delegate`, `dontAsk`, `bypassPermissions`, `plan` 중 하나만 사용.

**`initialPrompt` 필드**: Agent가 시작할 때 자동으로 제출할 초기 프롬프트. 사용자 입력을 기다리지 않고 즉시 작업을 시작할 수 있음 (v2.1.83+).

```yaml
# CORRECT
initialPrompt: "Analyze the following code for performance issues: @.src/"

# Agent가 시작되면 위의 프롬프트가 자동으로 제출됨
```

### Skills (.claude/skills/**/*.md)

**`allowed-tools` 필드**: 반드시 CSV 문자열 사용. YAML 배열 사용 금지.

```yaml
# CORRECT
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id

# WRONG
allowed-tools:
  - Read
  - Grep
```

**`description` 필드**: YAML folded scalar (>) 사용 권장.

```yaml
# CORRECT
description: >
  Multi-line description here.
  Uses YAML folded scalar for readability.

# ALSO OK (pipe scalar)
description: |
  Multi-line description here.
  Preserves line breaks.
```

**`metadata` 값**: 모든 값은 반드시 quoted string.

```yaml
# CORRECT
metadata:
  version: "1.0.0"
  category: "workflow"

# WRONG - unquoted values
metadata:
  version: 1.0.0
  category: workflow
```

### Quick Reference Table

| 파일 유형 | 필드 | 형식 | 예시 |
|-----------|------|------|------|
| Rules | `paths` | CSV string | `paths: "**/*.go,**/go.mod"` |
| Agents | `tools` | CSV string | `tools: Read, Write, Edit` |
| Agents | `disallowedTools` | CSV string | `disallowedTools: Task, WebSearch` |
| Agents | `initialPrompt` | String | `initialPrompt: "Analyze the code: @.src/"` |
| Agents | `skills` | YAML array | `skills:\n  - moai-lang-go` |
| Skills | `allowed-tools` | CSV string | `allowed-tools: Read, Grep` |
| Skills | `effort` | String | `effort: low` |
| Skills | `metadata.*` | Quoted strings | `version: "1.0.0"` |

### Validation Checklist

새 규칙/에이전트/스킬 파일을 생성하거나 수정할 때:

- [ ] `paths:` 필드가 CSV string 형식인지 확인
- [ ] `tools:` 필드가 CSV string 형식인지 확인
- [ ] `allowed-tools:` 필드가 CSV string 형식인지 확인
- [ ] `metadata:` 모든 값이 quoted string인지 확인
- [ ] Template 수정 후 `make build` 실행했는지 확인
- [ ] Local copy (`.claude/`)도 동일하게 수정했는지 확인

---

## 19. GLM Integration Testing Rules

### [HARD] Never Run GLM Integration Tests in the Dev Project

Running `go test ./internal/cli/` in the development project can invoke `moai cc` / `moai glm` command flows that modify **real settings files**:

- `.claude/settings.local.json` in the project root
- `~/.claude/settings.local.json` (global)

This is destructive and can wipe auth tokens or GLM configuration.

### Unit Tests vs Integration Tests

| Test Type | Where to Run | What's Allowed |
|-----------|-------------|----------------|
| Unit tests (function-level) | Dev project (`go test ./...`) | File manipulation in `t.TempDir()` only |
| Integration tests (full command) | `/tmp/test-project` via `claude -p` | Full `moai cc`, `moai glm`, `moai cg` |

### Rule: Use `~/.moai/.env.glm` for Auth Token Tests

Unit tests that verify ANTHROPIC_AUTH_TOKEN preservation must:
1. Load the real key via `loadGLMKey()` (reads `~/.moai/.env.glm`)
2. Skip with `t.Skip()` if the key is not configured
3. Never hardcode fake keys like `"test-key-123"` in test fixtures

```go
// CORRECT: Load from ~/.moai/.env.glm
realKey := loadGLMKey()
if realKey == "" {
    t.Skip("~/.moai/.env.glm not configured")
}

// WRONG: Hardcoded fake key
"ANTHROPIC_AUTH_TOKEN": "test-key-123"
```

### Rule: Integration Testing with `/tmp` + `claude -p`

For full command integration tests:
```bash
# 1. Create a temp test project
mkdir -p /tmp/moai-test-project && cd /tmp/moai-test-project

# 2. Initialize with moai
moai init .

# 3. Test with claude -p (pipe/programmatic mode)
claude -p "moai cc should restore Claude mode" --output-format json

# 4. Verify the settings file
cat .claude/settings.local.json
```

### What NOT to Do

```bash
# WRONG: Runs moai cc on the real dev project, modifies real settings
go test -run TestCCCmd_Execution ./internal/cli/

# WRONG: Any test that reads/writes to dev project's .claude/ directory
# WRONG: t.Setenv("HOME", tmpDir) — affects all parallel tests
```

---

---

## 20. Template Path Hardcoding Prevention

### [HARD] Never Use `.HomeDir` or `.GoBinPath` for Fallback Paths in Shell Templates

Shell script templates (`.sh.tmpl`) that need to reference the user's home directory or Go bin path MUST use shell environment variables (`$HOME`) instead of Go template variables (`.HomeDir`, `.GoBinPath`).

**Why**: `.HomeDir` and `.GoBinPath` are resolved at `moai init` time on the host machine and baked into generated scripts as absolute paths (e.g., `/Users/username/go/bin`). When the project is cloned or opened on another OS or by another user, these hardcoded paths silently fail.

**Rule for fallback binary lookups in `.sh.tmpl`:**

```bash
# WRONG: bakes in macOS absolute path at init time
if [ -f "{{posixPath .HomeDir}}/go/bin/moai" ]; then

# CORRECT: resolved at runtime per OS/user
if [ -f "$HOME/go/bin/moai" ]; then
```

**`.GoBinPath` is still valid** for the primary path injection (the first detected path at `moai init`), because it is the most specific location. But the `$HOME/go/bin` fallback MUST use `$HOME`.

**Checklist when editing any `.sh.tmpl` file:**
- [ ] Primary path: `{{posixPath .GoBinPath}}/moai` — OK (init-time detection)
- [ ] Fallback path: `$HOME/go/bin/moai` — MUST use `$HOME`, not `{{posixPath .HomeDir}}/go/bin`
- [ ] User-local path: `$HOME/.local/bin/moai` — MUST use `$HOME`
- [ ] After editing template: run `make build` to regenerate embedded files

**`renderer.go` passthrough**: `$HOME` is already in `claudeCodePassthroughTokens`, so the unexpanded token validator will not reject it.

---

**Status**: Active (Local Development)
**Version**: 1.7.0 (Template Path Hardcoding Prevention added)
**Last Updated**: 2026-03-18
