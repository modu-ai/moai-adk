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
| Skills | `allowed-tools` | ❌ 공백 구분 금지 | `Read Grep Glob` → YAML이 단일 string으로 파싱 (§18.1) |
| Skills | `effort` | String | `effort: low` |
| Skills | `metadata.*` | Quoted strings | `version: "1.0.0"` |

### §18.1 [CRITICAL] Space-separated `allowed-tools` 금지

YAML 스펙상 `allowed-tools: Read Grep Glob`는 **단일 string scalar** `"Read Grep Glob"`으로 파싱된다. Claude Code가 이 string을 공백으로 분리하는지, 쉼표로만 분리하는지는 공식 문서상 명시되지 않았으나 **CSV가 유일한 공식 권장 형식** (skill-authoring.md).

**Go yaml.v3 파서 검증** (2026-04-11 세션):
- `allowed-tools: Read Grep Glob` → `TYPE: string`, `VALUE: "Read Grep Glob"` → 쉼표 분리 시 1개 token "`Read Grep Glob`" (존재하지 않는 도구)
- `allowed-tools: Read, Grep, Glob` → `TYPE: string`, `VALUE: "Read, Grep, Glob"` → 쉼표 분리 시 3개 token 정상

공백 구분은 Claude Code가 **1개의 존재하지 않는 tool 이름**으로 해석하여 silently no tools allowed 상태가 될 수 있다. 검증 불가능한 영역이므로 **CSV만 허용**.

**금지 예시**:
```yaml
# WRONG — YAML 단일 scalar로 파싱됨
allowed-tools: Read Grep Glob Bash(git:*)

# CORRECT — CSV로 3+1 tool token 분리
allowed-tools: Read, Grep, Glob, Bash(git:*)
```

**탐지 grep 패턴**:
```bash
for f in .claude/skills/*/SKILL.md internal/template/templates/.claude/skills/*/SKILL.md; do
  line=$(grep "^allowed-tools:" "$f" 2>/dev/null)
  if [ -n "$line" ]; then
    val="${line#allowed-tools: }"
    if ! echo "$val" | grep -q ","; then
      nf=$(echo "$val" | awk '{print NF}')
      [ "$nf" -gt 1 ] && echo "DEFECT: $f"
    fi
  fi
done
```

### Validation Checklist

새 규칙/에이전트/스킬 파일을 생성하거나 수정할 때:

- [ ] `paths:` 필드가 CSV string 형식인지 확인
- [ ] `tools:` 필드가 CSV string 형식인지 확인
- [ ] `allowed-tools:` 필드가 CSV string 형식인지 확인 (공백 구분 절대 금지, §18.1 참조)
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

## 21. Go 패키지 하드코딩 방지 규칙

### [HARD] 패키지 코드(internal/, pkg/)에서 하드코딩 금지

moai-adk-go는 **범용 패키지**로 배포되므로, 사용자 환경에 종속되는 하드코딩이 절대 금지된다.
개발용 로컬 설정(`CLAUDE.local.md`, `.claude/settings.local.json`)과 패키지 코드(`internal/`, `pkg/`)를 혼동하지 않도록 주의.

### 구분: 개발 환경 vs 패키지 코드

| 영역 | 하드코딩 가능? | 예시 |
|------|-------------|------|
| **CLAUDE.local.md** | OK (로컬 전용) | `/Users/goos/MoAI/...` |
| **settings.local.json** | OK (로컬 전용) | `afplay /System/Library/Sounds/Glass.aiff` |
| **internal/, pkg/** Go 코드 | **절대 금지** | URL, 모델명, 환경변수명, 임계값 |
| **internal/template/templates/** | 컨텍스트별 | Go 템플릿 변수 + `$HOME` 규칙 준수 (Section 20) |
| **_test.go** 테스트 코드 | 조건부 OK | 테스트 픽스처, `t.TempDir()` 내 |

### 하드코딩 방지 체크리스트

새 Go 코드 작성 또는 기존 코드 수정 시 반드시 확인:

- [ ] **URL**: 상수로 추출 (`const apiURL = "https://..."`)
- [ ] **모델명**: 상수로 추출 (`const probeModel = "claude-haiku-..."`)
- [ ] **환경변수명**: `config/envkeys.go`에 상수 정의 후 참조 (`config.EnvXxx`)
- [ ] **임계값**: `config/defaults.go`에 정의, 중복 금지 (단일 원천)
- [ ] **조직/저장소명**: 상수로 추출 (`const githubReleasesURL = "..."`)
- [ ] **API 헤더/버전**: 상수로 추출 (반복 사용 시 중복 방지)
- [ ] **크로스 플랫폼**: `HOMEBREW_PREFIX`, `PSModulePath` 등 환경변수 우선 확인

### 환경변수 상수 파일 위치

모든 환경변수명은 `internal/config/envkeys.go`에 집중 정의:
- `MOAI_*` 계열: `config.EnvConfigDir`, `config.EnvTestMode` 등
- `CLAUDE_*` 계열: `config.EnvClaudeProjectDir`, `config.EnvClaudeConfigDir` 등
- `ANTHROPIC_*` 계열: `config.EnvAnthropicBaseURL`, `config.EnvAnthropicAuthToken` 등

새 환경변수 추가 시 반드시 `envkeys.go`에 상수 먼저 정의 후, 코드에서 참조.

### 중복 상수 방지

동일한 값을 여러 패키지에서 독립 정의하지 말 것:

```go
// WRONG: 같은 값 85.0이 두 패키지에 독립 정의
// internal/loop/state.go
const DefaultCoverageTarget = 85.0
// internal/hook/teammate_idle.go
const defaultCoverageThreshold = 85.0

// CORRECT: config/defaults.go의 단일 원천 참조
float64(config.DefaultTestCoverageTarget)
```

---

## 22. 템플릿 언어 중립성 규칙

### [HARD] 패키지 템플릿은 특정 언어 전용으로 하드코딩 금지

`internal/template/templates/` 하위의 모든 파일은 **범용 패키지로 배포되어 사용자 프로젝트에 적용**되므로, 특정 언어 전용 가정이나 편향을 담으면 안 된다.

moai-adk-go는 Go로 작성된 도구이지만, **도구 자체의 언어와 사용자 프로젝트 언어는 완전히 별개**다. Python 개발자가 `moai init`을 실행하면 그 사람에게 필요한 것은 pylsp이지 gopls가 아니다.

### 허용 vs 금지

| 위치 | 언어 편향 허용? | 이유 |
|------|-------------|------|
| **CLAUDE.local.md** | ✅ 허용 | 이 프로젝트(moai-adk-go) 개발자 전용 |
| **`.moai/config/sections/*.yaml`** (로컬) | ✅ 허용 | 이 프로젝트 자체가 Go |
| **`.claude/settings.local.json`** | ✅ 허용 | 로컬 세션 전용 |
| **`internal/template/templates/**/*.tmpl`** | ❌ **금지** | 16개 언어 모두 동등 취급 |
| **`internal/template/templates/**/*.yaml`** | ❌ **금지** | 사용자 환경 감지 or 중립 |
| **`internal/template/templates/CLAUDE.md`** | ❌ **금지** | 다국적 언어 중립 |
| **Go 코드 `internal/`, `pkg/`** | ❌ 금지 | URL/model 상수화 (Section 21) |

### 체크리스트

새 템플릿 파일 추가 또는 수정 시 반드시 확인:

- [ ] 특정 언어 바이너리(gopls, pylsp, rust-analyzer 등)를 "PRIMARY"로 배치하지 않았는가?
- [ ] 16개 지원 언어(go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)가 동등 수준으로 나열되어 있는가?
- [ ] 특정 언어만 "enabled: true", 다른 언어는 "planned"/"deferred"로 격하하지 않았는가?
- [ ] "Primary/Secondary/Tertiary" 분류가 있다면, 언어가 아닌 "기능"(문법 검사, 타입 체크, 린트) 기준인가?
- [ ] 사용자 환경 감지 로직(project_markers, extension-based)이 포함되어 있는가?

### 16개 지원 언어 목록

Template 파일 작성 시 반드시 16개 모두 동등 처리:

```
go, python, typescript, javascript, rust, java, kotlin, csharp,
ruby, php, elixir, cpp, scala, r, flutter, swift
```

**캐논 이름 주의**: Dart/Flutter는 캐논 이름이 **"flutter"** (not "dart").
이는 `.claude/skills/moai/workflows/sync.md` Phase 0.6.1 "Language Detection" 테이블과 일치하며, `.claude/rules/moai/languages/flutter.md` 파일 위치와도 일치. MoAI 프로젝트에서 Flutter는 주로 모바일 앱 개발 맥락에서 사용되므로 "dart"보다 "flutter"를 선호.

이 목록은 `.claude/skills/moai/workflows/sync.md` Phase 0.6.1 "Language Detection" 테이블과 일치해야 하며, 변경 시 두 파일 모두 동기화 필요.

### 사용자 환경 감지 패턴

템플릿은 project marker 파일로 사용자 프로젝트 언어를 감지하여 해당 도구만 활성화해야 한다:

```yaml
# CORRECT: 16개 언어 동등, project_markers로 자동 감지
servers:
  go:
    project_markers: ["go.mod", "go.sum"]
    binary: gopls
  python:
    project_markers: ["pyproject.toml", "setup.py", "Pipfile", "requirements.txt"]
    binary: pylsp
  typescript:
    project_markers: ["tsconfig.json", "package.json"]
    binary: typescript-language-server
  rust:
    project_markers: ["Cargo.toml"]
    binary: rust-analyzer
  # ... (16개 모두 동일 스키마, 동일 우선순위)
```

```yaml
# WRONG: Go를 특별 취급
gopls_bridge:        # ← Go 전용 최상위 섹션
  enabled: false
  binary: gopls
future_servers:      # ← 다른 언어를 "future"로 격하
  python: { status: "planned" }
  typescript: { status: "planned" }
```

### 로컬과 템플릿 분리 원칙

이 프로젝트(moai-adk-go)는 Go로 작성되었지만, **템플릿은 사용자를 위한 것**이지 이 프로젝트를 위한 것이 아니다.

- `internal/template/templates/.moai/config/sections/lsp.yaml.tmpl`: **범용**, 16개 언어 동등
- `.moai/config/sections/lsp.yaml` (로컬): **moai-adk-go 개발용**, gopls 편향 허용

두 파일의 내용이 달라도 정상이며, 오히려 **같으면 규칙 위반 의심**. 로컬은 이 프로젝트의 실제 개발 필요에 맞게 편향되어야 하고, 템플릿은 사용자 환경에 맞춰 중립이어야 한다.

### 허용되는 언어별 차등 (예외)

다음 경우에만 언어별 차등이 허용된다:

1. **공식 지원 티어**: `.moai/project/tech.md` 또는 `README.md`에 "Tier 1: Go, Python, TS / Tier 2: 나머지" 같은 공식 우선순위가 명시된 경우
2. **상태가 실제로 다를 때**: 특정 언어 서버가 stable, 다른 언어는 experimental인 경우 (status 필드로 명시)
3. **기능 제약**: 특정 기능이 특정 언어에서만 동작하는 경우 (드물게)

이 예외들도 반드시 **명시적 문서화**가 필요하며, 암묵적 편향은 허용되지 않는다.

### 검증 방법

새 템플릿 파일 커밋 전 수행:

```bash
# 1. 템플릿 파일에 언어 이름 등장 횟수 체크
for lang in go python typescript javascript rust java kotlin csharp ruby php elixir cpp scala r dart swift; do
  count=$(grep -c "^\s*$lang:" internal/template/templates/path/to/file.tmpl)
  echo "$lang: $count"
done

# 2. "PRIMARY" / "gopls" 등 특정 언어 편향 키워드 검색
grep -i "primary.*gopls\|gopls.*primary" internal/template/templates/**

# 3. 로컬과 템플릿 내용 비교 (동일하면 경고)
diff .moai/config/sections/lsp.yaml internal/template/templates/.moai/config/sections/lsp.yaml.tmpl
```

### 역사적 참고

이 규칙은 2026-04-11 세션에서 PR #624가 병합된 직후 발견된 문제에서 비롯되었다:
- PR #624는 `lsp.yaml.tmpl`을 작성하며 gopls를 "PRIMARY Path C"로 배치
- Python/TypeScript 등 15개 언어는 "future_servers.status: planned"로 격하
- Python 개발자가 `moai init` 실행 시 자신에게 필요 없는 Go-biased config를 받게 됨
- 근본 원인: 이 프로젝트가 Go로 작성되었다는 사실이 무의식 중에 템플릿 설계에 반영됨

**교훈**: 템플릿 파일을 작성할 때 "이 프로젝트가 어떤 언어로 만들어졌는지"를 완전히 잊어버리고, "16개 언어 중 어떤 언어 사용자라도 동등하게 환영받아야 한다"는 관점을 유지해야 한다.

---

## 23. Audit Sweep Patterns

### [HARD] 전수 감사 의뢰 시 반드시 재검증 루프 포함

**배경**: 2026-04-11 세션에서 `Explore` 서브에이전트에게 전체 스킬 frontmatter 감사를 의뢰했으나, 2개 스킬(`moai-platform-auth`, `moai-platform-chrome-extension`)의 `allowed-tools` 공백 구분 결함을 놓쳤다. Iteration 2 재검증 grep에서 추가 발견되었다. **7%의 false-negative rate** (2/27).

### Explore/Plan 등 Claude Code 내장 서브에이전트의 한계

- `Explore`, `Plan`, `general-purpose`는 Claude Code **내장** 서브에이전트로, `.claude/agents/`에 파일이 없어 **직접 프롬프트 수정 불가**.
- 감사 품질은 에이전트에게 전달되는 **프롬프트 지시문의 완결성**에 전적으로 의존한다.
- 따라서 "전수 감사" 의뢰 시 에이전트를 개선하는 것이 아니라 **의뢰 프롬프트와 재검증 루프**를 강화해야 한다.

### [HARD] 감사 프롬프트 필수 구성요소

1. **명시적 카테고리 분류**: 카테고리를 번호(A, B, C...)로 나누고 각 카테고리의 판정 기준을 구체적 정규식 또는 문자열 매칭으로 명시
2. **구체적 grep/glob 힌트**: 에이전트가 직접 실행할 수 있는 명령어를 프롬프트 내 제공
3. **조사 대상 디렉토리 양쪽 모두 명시**: `internal/template/templates/...` + `.claude/...` (Template-First rule 대응)
4. **예상 결함 수량 가이드**: "N개 카테고리에서 최소 M개 결함 예상" — 에이전트가 너무 일찍 멈추는 것을 방지
5. **재검증 3단계 루프 필수** (아래 참조)

### 감사 결과 재검증 3단계 루프

| 단계 | 방법 | 목적 |
|------|------|------|
| 1. 1차 감사 | Explore 에이전트에 sweep 패턴 + 카테고리 명시 의뢰 | 구조적 결함 발견 |
| 2. 2차 재검증 | **동일 sweep을 Bash 스크립트로 직접 재실행** | 에이전트 누락 보완 (이번 세션 +2개 발견) |
| 3. 3차 full-file 파싱 | Go yaml.v3 파서 등으로 실제 YAML 유효성 검증 | 단순 grep이 놓치는 파싱 오류 탐지 |

### 검증된 sweep 패턴 라이브러리

```bash
# 패턴 A: 공백 구분 allowed-tools 탐지
for f in .claude/skills/*/SKILL.md internal/template/templates/.claude/skills/*/SKILL.md; do
  line=$(grep "^allowed-tools:" "$f" 2>/dev/null)
  if [ -n "$line" ]; then
    val="${line#allowed-tools: }"
    if ! echo "$val" | grep -q ","; then
      nf=$(echo "$val" | awk '{print NF}')
      [ "$nf" -gt 1 ] && echo "SPACE-SEP: $f"
    fi
  fi
done

# 패턴 B: Skill() 호출 vs allowed-tools 일관성
for f in .claude/skills/*/SKILL.md internal/template/templates/.claude/skills/*/SKILL.md; do
  if grep -q 'Skill("' "$f" 2>/dev/null; then
    if ! grep -q "^allowed-tools:.*Skill" "$f" 2>/dev/null; then
      echo "MISSING Skill in allowed-tools: $f"
    fi
  fi
done

# 패턴 C: license template vs local 불일치
for f in .claude/skills/*/SKILL.md; do
  base=$(basename $(dirname "$f"))
  tpl="internal/template/templates/.claude/skills/$base/SKILL.md"
  if [ -f "$tpl" ]; then
    local_lic=$(grep "^license:" "$f" | sed 's/license: //')
    tpl_lic=$(grep "^license:" "$tpl" | sed 's/license: //')
    [ "$local_lic" != "$tpl_lic" ] && echo "MISMATCH: $base"
  fi
done

# 패턴 D: Go yaml 파싱 전체 유효성 (gopkg.in/yaml.v3 이미 go.mod에 있음)
# 임시 Go 스크립트로 모든 SKILL.md frontmatter 파싱 후 에러 집계
```

### 교훈

- **에이전트 보고서는 최소 기준선이며 절대적 진실이 아니다.** 이번 세션에서 에이전트는 25개 결함을 보고했으나 실제는 27개였다 (7% 누락).
- **재검증 grep sweep이 누락을 잡아냈다.** Iteration 2의 final sweep이 `moai-platform-auth`, `moai-platform-chrome-extension`을 추가 발견해 완전한 범위로 확장되었다.
- **Template-First 원칙**: 감사 시 반드시 template + local **양쪽** 모두 스캔. 한쪽만 보면 절반을 놓친다.
- **Go 파서 검증이 grep보다 강력**: CSV 형식 검증만으로는 공백 구분이 실제로 어떻게 파싱되는지 모름. Go yaml.v3 파서로 실제 type + value를 출력하면 false-positive 없이 확정 가능.

### 적용 지침

향후 전수 감사 의뢰 시 반드시:
1. 이 §23 섹션을 감사 의뢰 프롬프트에 include (또는 요약)
2. 감사 결과 수신 후 **Bash 재검증 스크립트를 반드시 실행**
3. 불일치 발견 시 iteration 추가 — "에이전트가 한 말이 전부"라고 절대 가정 금지

---

## 24. 오케스트레이터 자가 점검 (Orchestrator Self-Check)

### [HARD] 복잡 작업 시작 전 위임 우선 판정

**배경**: 2026-04-11 세션에서 MoAI가 7개 SPEC × 3 파일 = 21개 파일을 `manager-spec` 에이전트를 호출하지 않고 Write 툴로 직접 생성한 사례 발생. `moai-constitution.md`의 최상위 원칙 "MoAI is the strategic orchestrator for Claude Code. **Direct implementation by MoAI is prohibited for complex tasks**"를 명백히 위반. 근본 원인: 앞선 세션 내 수많은 직접 수정 패턴이 습관화되어 복잡 작업에도 무비판적으로 적용됨.

### [HARD] 자가 점검 3 질문

MoAI는 복잡 작업 시작 전 다음 질문을 반드시 **스스로** 해야 한다:

1. **도메인 일치**: 이 작업은 전문 에이전트(`manager-*`, `expert-*`, `builder-*`)의 고유 도메인인가?
2. **에이전트 존재**: 해당 전문 에이전트가 카탈로그에 존재하는가? (CLAUDE.md Section 4)
3. **위임 이득**: 직접 수행보다 위임이 품질/독립성/편향 방지에 유리한가?

**3개 모두 YES이면 직접 수행 금지**. 사용자에게 위임 방식 확인 후 실행.

### 강제 위임 대상 (직접 수행 불가)

| 작업 유형 | 필수 위임 에이전트 | 근거 |
|---|---|---|
| SPEC 생성/재작성 (EARS 포맷) | `manager-spec` | SPEC 전문가 |
| 에이전트 정의 생성 (`.claude/agents/`) | `builder-agent` | 에이전트 빌더 |
| 스킬 생성 (`.claude/skills/`) | `builder-skill` | 스킬 빌더 |
| 플러그인/마켓플레이스 생성 | `builder-plugin` | 플러그인 빌더 |
| `internal/`, `pkg/` Go 구현 코드 | `expert-backend` | 백엔드 구현 |
| React/Vue 컴포넌트 구현 | `expert-frontend` | 프런트엔드 구현 |
| 보안 감사/OWASP | `expert-security` | 보안 전문가 |
| 성능 최적화/프로파일링 | `expert-performance` | 성능 전문가 |
| E2E/통합 테스트 작성 | `expert-testing` | 테스트 전문가 |
| 리팩토링/codemod | `expert-refactoring` | 리팩토링 전문가 |
| 디버깅/근본 원인 분석 | `expert-debug` | 디버깅 전문가 |
| 문서 전면 재작성 (README, API docs) | `manager-docs` | 문서 전문가 |
| 프로젝트 초기 설정 (`moai init` 이후) | `manager-project` | 프로젝트 전문가 |
| DDD/TDD 구현 워크플로우 | `manager-ddd` / `manager-tdd` | 구현 워크플로우 |

### 수량 기반 트리거

- **같은 종류 파일 5개 이상 생성** → 전문가 위임 강제 (예: SPEC 5개 → `manager-spec`)
- **같은 종류 파일 10개 이상 수정** → 전문가 위임 권장
- **Go 코드 500+ LOC 신규 작성** → `expert-backend` 강제 위임
- **테스트 파일 10개 이상 작성** → `expert-testing` 강제 위임
- **에이전트 3개 이상 생성** → `builder-agent` 강제 위임
- **스킬 3개 이상 생성** → `builder-skill` 강제 위임

### 허용되는 직접 수행 (예외)

다음 경우에만 MoAI 직접 수행이 허용된다:

1. **Typo/포맷팅 수정**: 단일 파일, 1-5줄 수정
2. **설정 파일 1개 편집**: `.claude/settings.local.json`, `.moai/config/sections/*.yaml` 단건
3. **사용자 명시 요청**: "네가 직접 해", "delegate하지 말고" 등 명시적 지시
4. **위임 대상 부재**: 해당 도메인 전문 에이전트가 카탈로그에 없음
5. **오케스트레이션 자체**: AskUserQuestion, TaskCreate/Update, Agent 호출
6. **감사/검수 결과 통합**: 에이전트 보고서를 사용자용으로 요약·정리
7. **git 작업**: 브랜치, 커밋, push, PR (manager-git이 있으나 MoAI가 직접 처리 허용)
8. **테스트용 임시 수정**: `/tmp`, `.moai/cache/`, worktree 내 작업

### 체크포인트 프로토콜

**복잡 작업이라고 판정되면 반드시 다음 순서**:

1. **위임 제안**: AskUserQuestion으로 "이 작업은 {전문가-에이전트}에 위임하는 것이 원칙입니다. 어떻게 진행할까요?" 질문
2. **옵션 제시** (최소 2개):
   - (권장) 전문가 에이전트에 위임
   - 대안: MoAI 직접 작성 + 전문가 에이전트 검수
3. **사용자 선택 수신**
4. **실행**: 선택에 따라 Agent 호출 또는 직접 수행
5. **검증**: 직접 수행 선택 시 작업 완료 후 전문가 검수 필수

### Rule 5 (Context-First Discovery)와의 관계

Rule 5는 **WHAT** (무엇을 할 것인가)을 명확히 한다. §24는 **WHO** (누가 할 것인가)를 결정한다. 순서:

1. Rule 5: Socratic 인터뷰로 의도 100% 파악
2. **§24: 전문가 위임 여부 결정** (신규)
3. Rule 1 (Approach-First): 실행 방법 설명 + 승인
4. 실행

§24는 Rule 5와 Rule 1 **사이**에 위치하며, 둘 사이의 누락된 연결고리를 채운다.

### 실패 모드 자가 진단

MoAI가 복잡 작업에 직면했을 때 스스로 다음을 물어야 한다:

- [ ] 내가 "이미 알고 있다"는 친숙성 때문에 전문가 호출을 생략하려 하는가?
- [ ] 세션 효율 압박이 위임 결정을 흐리고 있는가?
- [ ] 앞선 직접 수정 패턴의 관성으로 무의식 중 Write를 선택하려 하는가?
- [ ] 연구·계획 수행자(나 자신)와 구현자(에이전트)를 혼동하고 있는가?
- [ ] Agent 툴 존재를 잊고 Write/Edit만 떠올리고 있는가?
- [ ] 사용자가 "범위"를 선택했다고 "방식"까지 동의한 것으로 착각하고 있는가?

하나라도 YES면 **실행 중단 → AskUserQuestion 재검토**.

### 역사적 계기

2026-04-11 세션: 사용자 "2,3 이어서 모두 진행을하자 /moai spec이 모두 생성이 되어 있는가? 없다면 추가를 하자" 요청 → MoAI가 `manager-spec` 호출 대신 Write × 21회 직접 실행 → 사용자 "왜 spec 에이전트가 아닌 직접 스펙을 생성한 원인을 찾아서 보고 해라" 지적. 5 Whys 근본 원인:

1. "내가 내용을 이미 안다" (친숙성 편향)
2. 연구 선행 완료 → 전문가 불필요 착각 (과제 혼동)
3. "컨텍스트"와 "전문성" 혼동 (주제 지식 ≠ SPEC 문법 전문성)
4. 앞선 직접 수정 패턴 관성 (무비판적 패턴 이월)
5. 자가 점검 체크포인트 부재 (§24가 없어서 발생)

**교훈**: 주제 지식이 있다고 전문가 대역이 되지 않는다. manager-spec은 EARS 엄격성·REQ 번호 일관성·YAML 스키마·Section 22 준수 등을 전문화된 관점으로 검증한다. MoAI가 이를 대체할 수 없다.

### 검증 방법

세션 중 복잡 작업 감지 시 self-check 수행:

```bash
# 1. 현재 작업 분류
echo "작업 종류: SPEC 생성 / 에이전트 생성 / 코드 구현 / 문서 작성 / 감사 / ..."

# 2. 수량 측정
echo "예상 파일 수: N"
echo "예상 LOC: N"

# 3. 강제 위임 테이블 조회
# 위 표와 대조하여 위임 대상 에이전트 식별

# 4. 위임 대상 있음 → AskUserQuestion 필수
# 위임 대상 없음 → 직접 수행 허용
```

---

**Status**: Active (Local Development)
**Version**: 2.1.0 (§24 Orchestrator Self-Check 추가)
**Last Updated**: 2026-04-11
