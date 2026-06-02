# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-05-25

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

**§2.1 Template Content Neutrality — Acceptable Content Range for Templates**: When editing template source files in `internal/template/templates/`, ensure content adheres to the **acceptable** kept-classes (C1/C2/C4/C5/C6/C8 per `.claude/rules/moai/development/coding-standards.md` MUST constraints). FORBIDDEN content classes (SPEC IDs, REQ tokens, Audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references) are enforced by CI guard (`.github/workflows/template-neutrality-check.yaml` trigger on path change). See `.claude/rules/moai/development/coding-standards.md` § MUST and **§25 (Template Internal-Content Isolation)** of this file for the canonical acceptable-vs-forbidden content range. This ensures 16-language template distribution remains neutral to moai-adk internal development state.

**Pre-PR Verification (template contributor-checklist)** — before opening a PR that touches `internal/template/templates/**` (the CI guard `template-neutrality-check.yaml` is the safety net):

- [ ] No `/Users/` or OS-specific absolute path (C1) — use `$HOME` / `~`
- [ ] No bare-narrative `V3R[0-9]` dev-version sigil (C2) outside the doctrine allow-list
- [ ] No `feedback_` / `memory.md` ref (C4) outside the canonical-doctrine allow-list
- [ ] No `CLAUDE.local.md` reference (C5)
- [ ] No `PR #N` reference (C6)
- [ ] `GOOS=` cross-compile env vars preserved (C8)
- [ ] `go test ./internal/template/... -run TestTemplateNeutralityAudit` passes in isolation

See §25.3 for the full 5-item pre-commit self-check and §25.1 for the forbidden/allowed content-class catalogue. (C3 dates + C7 commit-hashes are owned by the sibling `internal_content_leak_test.go` per §25, not this neutrality checklist.)

### Local-Only Files (Never in Templates)
```
.claude/settings.local.json    # Personal settings — runtime-managed, NEVER template
.claude/settings.json          # Rendered from .json.tmpl
.claude/agent-memory/          # Per-project agent memory
.claude/hooks/moai/handle-*.sh # Generated hook wrappers (not templates)
.claude/commands/97-release-update.md            # Dev-only: CC upstream tracker (§21)
.claude/commands/98-*.md       # Dev-only: maintainer commands — 98-github.md 등 (§21)
.claude/commands/99-*.md       # Dev-only: future maintainer commands (§21 reserved)
.claude/skills/moai/workflows/release-update.md  # Dev-only: release-update workflow body (§21)
CLAUDE.local.md                # This file
.moai/state/last-cc-version.json # Dev-only: CC tracking state (§21)
.moai/research/cc-update-*.md  # Dev-only: CC update reports (§21)
.moai/cache/                   # Cache
.moai/logs/                    # Logs
.moai/state/                   # Session state storage
.moai/specs/                   # Active SPEC documents
.moai/plans/                   # Session plans
.moai/reports/                 # Generated reports
.moai/manifest.json            # Generated at runtime
.moai/status_line.sh           # Rendered from .sh.tmpl
```

### [HARD] settings.local.json Separation

`settings.local.json` is **runtime-managed**. Never put it in templates.

- Modified by `moai glm`, `moai cc`, `moai cg` commands at runtime
- Modified by SessionStart hook (GLM credentials, teammateMode, CLAUDE_ENV_FILE)
- Contains per-machine values: tmux pane IDs, API tokens, absolute paths
- **Never** add effortLevel, teammateMode, or env tokens to the template

If you accidentally commit `settings.local.json`, run `git rm --cached .claude/settings.local.json`.

### [WARN] OpenTelemetry / OTEL in Tests

Do NOT use `t.Setenv` with OTEL environment variables (`OTEL_EXPORTER_*`, `OTEL_SERVICE_NAME`) in tests. Setting these in parallel tests causes data races because the OTEL SDK initializes global state from env vars on first use.

- Use a fake/no-op exporter instead of env-var configuration in tests
- If the test must set OTEL vars, make the parent test non-parallel and use `t.Setenv` only in non-parallel subtests

### Embedded Template System

moai-adk-go uses Go's `go:embed` directive:
- **Source**: `internal/template/templates/` (edit here)
- **Generated**: `internal/template/embedded.go` (auto-generated, DO NOT EDIT)
- **Build**: Run `make build` after editing templates

---

## 3. Code Standards

Language policy는 `.claude/rules/moai/development/coding-standards.md`에 정의 (auto-loaded).

### Go-Specific (이 프로젝트 전용)

- File naming: `snake_case.go`, `snake_case_test.go`
- Error wrapping: `fmt.Errorf("operation: %w", err)` (string concatenation 금지)
- All code, comments, godoc in English

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

## 11. Frequent Issues and Solutions

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

---

## 12. YAML Frontmatter 빠른 참조

범용 형식 규칙은 `.claude/rules/moai/development/` 내 `skill-authoring.md`, `agent-authoring.md`에 정의.

### 로컬 개발 체크리스트

- [ ] `tools:`, `allowed-tools:` → CSV string (공백 구분 절대 금지)
- [ ] `skills:` → YAML array (유일한 예외)
- [ ] `metadata.*` → quoted string
- [ ] Template 수정 후 `make build` 실행
- [ ] Local copy (`.claude/`)도 동기화

탐지 스크립트: `memory/audit_sweep_patterns.md` Pattern A 참조.

---

## 13. GLM Integration Testing

### [HARD] Dev 프로젝트에서 GLM 통합 테스트 실행 금지

`moai cc`/`moai glm` 커맨드 플로우는 실제 settings 파일을 수정하므로 dev project에서 절대 실행 금지.

- Unit tests: dev project (`go test ./...`), `t.TempDir()` 내 파일만
- Integration tests: `/tmp/test-project`에서 `claude -p`로 실행
- Auth token: `loadGLMKey()` (reads `~/.moai/.env.glm`), 없으면 `t.Skip()`
- 금지: `t.Setenv("HOME", tmpDir)` (병렬 테스트 오염), 하드코딩 fake key

---

## 14. 하드코딩 방지

### [HARD] Go 코드 (internal/, pkg/) 하드코딩 금지

- URL, 모델명, 조직명, API 헤더 → `const`로 추출
- 환경변수명 → `internal/config/envkeys.go`에 상수 정의 후 참조
- 임계값 → `config/defaults.go`에 단일 원천 정의, 중복 금지
- 크로스 플랫폼 → `$HOME`, `HOMEBREW_PREFIX` 등 환경변수 우선

### [HARD] .sh.tmpl 폴백 경로에 `.HomeDir` 금지

`.HomeDir`/`.GoBinPath`는 `moai init` 시점의 절대 경로로 굳어짐. 폴백에는 `$HOME` 사용:
- Primary: `{{posixPath .GoBinPath}}/moai` (OK, init-time)
- Fallback: `$HOME/go/bin/moai` (MUST use `$HOME`)
- `renderer.go`: `$HOME`은 `claudeCodePassthroughTokens`에 이미 등록

### 하드코딩 허용 영역

`CLAUDE.local.md`, `settings.local.json`, `_test.go` (t.TempDir() 내).

---

## 15. 템플릿 언어 중립성

### [HARD] `internal/template/templates/` 하위는 16개 언어 동등 취급

도구의 구현 언어(Go)와 사용자 프로젝트 언어는 별개. 템플릿은 모든 사용자를 위한 것.

- 언어 편향 허용: `CLAUDE.local.md`, `settings.local.json`, 로컬 `.moai/config/`
- 언어 편향 금지: `internal/template/templates/**` 전체

### 16개 지원 언어 (모두 동등)

```
go, python, typescript, javascript, rust, java, kotlin, csharp,
ruby, php, elixir, cpp, scala, r, flutter, swift
```

Dart/Flutter 캐논 이름: **"flutter"** (not "dart").

### 체크리스트 (템플릿 수정 시)

- [ ] 특정 언어를 "PRIMARY"로 배치하지 않았는가?
- [ ] 16개 언어가 동등 수준으로 나열되어 있는가?
- [ ] 특정 언어만 "enabled", 나머지 "planned"로 격하하지 않았는가?
- [ ] project_markers 기반 자동 감지 로직이 포함되어 있는가?
- [ ] 로컬 config와 템플릿이 달라도 정상 (같으면 오히려 의심)

상세 교훈: `memory/lessons.md` #5 참조.

---

## 16. 오케스트레이터 자가 점검

### [HARD] 자가 점검 4 질문 (복잡 작업 시작 전 필수)

1. 이 작업은 전문 에이전트의 고유 도메인인가?
2. 해당 전문 에이전트가 카탈로그에 존재하는가? (CLAUDE.md Section 4)
3. 직접 수행보다 위임이 품질/독립성/편향 방지에 유리한가?
4. 이 작업의 일부를 read-only sub-agent 병렬 spawn으로 분해할 수 있는가? (Anthropic "Exploration-First Pattern": WebFetch + Explore subagent로 분석 → main agent로 종합)

**3개 이상 YES → 직접 수행 금지**. 4번째 질문도 YES인 경우 Exploration-First Pattern (read-only sub-agent 병렬 spawn) 우선 적용. AskUserQuestion으로 위임 방식 확인 후 실행.

### 수량 기반 트리거

- 같은 종류 파일 **5+** 생성 → `manager-develop` 또는 `builder-harness` 위임 강제
- Go 코드 **500+ LOC** 신규 → `manager-develop` (cycle_type=tdd, domain context: backend) 강제 — 과거 `expert-backend`는 archived per SPEC-V3R6-AGENT-TEAM-REBUILD-001
- 에이전트/스킬 **3+** 생성 → `builder-harness` 강제

### 허용되는 직접 수행

Typo/포맷 수정, 설정 1개 편집, 사용자 명시 요청, 위임 대상 부재, 오케스트레이션 자체, git 작업, `/tmp` 작업.

### 순서: Rule 5 → §16 → Rule 1

Rule 5(WHAT) → §16(WHO) → Rule 1(HOW) → 실행

상세 교훈 및 5 Whys: `memory/lessons.md` #4 참조.

---

## 17. docs-site 4개국어 문서 동기화 규칙

docs-site는 `adk.mo.ai.kr` 공식 사용자 문서. URL 표준, 4-locale 동기화 의무, Mermaid TD-only, Vercel 프로젝트 바인딩, 빌드/배포 체크리스트 등 전체 doctrine은 외부 파일 참조.

See: `.moai/docs/docs-site-i18n-rules.md`

---

## 18. Git Workflow — Enhanced GitHub Flow

v2.14.0 릴리스 이후 공식 채택. 5-axis 즉시 개선 (branch protection / label 3축 / merge strategy / Release Drafter / hotfix naming) + Enhanced GitHub Flow 11 branch prefix + Merge strategy 표 + BODP 3-Signal Evaluation + v2.14.0 Case Study + AskUserQuestion Enforcement Protocol 등 전체 doctrine은 외부 파일 참조.

See: `.moai/docs/git-workflow-doctrine.md`

---

## 19. AskUserQuestion Enforcement Protocol

> **[CANONICAL]** 본 섹션의 모든 enforcement 룰 — deferred tool preload 의무, pre-response self-check 4항목, anti-pattern 카탈로그, recovery protocol — 은 `.claude/rules/moai/core/askuser-protocol.md` 에 단일 진실 공급원(SSOT)으로 존재합니다. 본 §19은 cross-reference만 유지하며, 규칙 갱신 시 canonical 파일을 수정하세요.

### Quick Pointer

| 항목 | Canonical 위치 |
|------|----------------|
| Channel Monopoly + Free-form 금지 | `askuser-protocol.md` § Channel Monopoly |
| ToolSearch Preload 절차 (의무) | `askuser-protocol.md` § ToolSearch Preload Procedure |
| Socratic Interview 구조 (≤4Q × ≤4 options, `(권장)` first) | `askuser-protocol.md` § Socratic Interview Structure |
| Option Description 표준 + Bias Prevention | `askuser-protocol.md` § Option Description Standards |
| Orchestrator–Subagent 비대칭 boundary | `askuser-protocol.md` § Orchestrator–Subagent Boundary |
| Ambiguity Trigger 4종 + Exception 5종 | `askuser-protocol.md` § Ambiguity Triggers and Exceptions |
| Free-form Circumvention 금지 + "Other" 메커니즘 | `askuser-protocol.md` § Free-form Circumvention Prohibition |

### Local Notes

본 incident 기록 (2026-04-24): `~/.claude/projects/{hash}/memory/feedback_askuserquestion_enforcement.md`. v3.4.0부터 enforcement 정책 HARD 운영. 위반 탐지 시 즉시 canonical §Recovery Protocol 적용 + memory 추가 기록.

상위 정책 참조:
- CLAUDE.md §1 HARD Rules (AskUserQuestion-Only + Deferred Tool Preload)
- CLAUDE.md §8 User Interaction Architecture
- `.claude/skills/moai/SKILL.md` § Red Flags + Verification

### §19.1 GATE-2 Mandatory Restoration (REQ-ATR-015 — SPEC-V3R6-AGENT-TEAM-REBUILD-001)

[HARD] **GATE-2 (plan-to-implement HUMAN GATE)는 자율 bypass 대상이 아니다.** Plan-phase 산출물이 audit-ready 상태로 PASS 되었더라도, run-phase 진입 직전 orchestrator는 자율 흐름을 중단하고 사용자에게 명시적 진행 승인을 `AskUserQuestion`으로 받아야 한다. 이는 Anthropic Claude Code의 Ctrl+G plan editor mandate (plan-to-implement 경계에서 사용자 개입 의무)와 정합한다.

**skip-eligible 0.90 autonomous bypass 정책의 적용 범위**: `skip-eligible` (score ≥ 0.90) autonomous bypass는 **Phase 0.5 plan-auditor verdict 재실행에만** 적용된다 — CONST-V3R5-026 + `.claude/rules/moai/workflow/spec-workflow.md` § Plan Audit Gate skip policy 참조. **GATE-2 (plan-to-implement HUMAN GATE)에는 적용되지 않는다**. Phase 0.5 SKIP과 GATE-2 SKIP은 서로 다른 결정 — Phase 0.5는 plan-auditor의 verdict 재실행 여부 (자동화 가능), GATE-2는 사용자가 run-phase 진입을 승인할지 여부 (사용자 결정 필수).

**오케스트레이터 의무 (GATE-2 entry)**:
1. Plan-phase 산출물 + plan-auditor verdict 요약을 사용자에게 prose로 제시
2. `ToolSearch(query: "select:AskUserQuestion")` preload
3. `AskUserQuestion` 으로 "run-phase 진입 / 추가 검토 / 중단" 3-option 제시 (첫 옵션 "(권장)" 라벨)
4. 사용자 응답 수신 후 run-phase 진입 (또는 중단)

**위반 anti-pattern**: Phase 0.5 verdict가 PASS skip-eligible (≥ 0.90)이라는 이유만으로 사용자 승인 없이 `/moai run`을 자율 시작하는 행위. GATE-2는 plan-auditor 점수와 무관한 별도 사용자 의지 확인 절차다.

상위 SPEC 참조:
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md` REQ-ATR-015 (GATE-2 restoration)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E (GATE-2 vs Phase 0.5 vs Phase 0.95 boundary)

---

**Status**: Active (Local Development)
**Version**: 3.9.0 (§19.1 GATE-2 mandatory restoration cross-reference 추가 per REQ-ATR-015)
**Last Updated**: 2026-05-25

---

## 20. Vercel Build Cost Guard

### [HARD] Build Machine = Elastic 유지

- Vercel Team default + 각 프로젝트 모두 **Elastic** 머신 사용. Turbo($0.126/min) 또는 Standard로 변경 금지 — Elastic은 $0.0035/CPU min로 약 40배 저렴
- 새 프로젝트 추가 시 Settings → Build and Deployment → Build Machine = Elastic 확인
- 비용 폭탄 의심 시 **가장 먼저 Build Machine 설정 점검**
- docs-site는 §17.6 Vercel 프로젝트 바인딩과 함께 운영 — 비용 의심 시 §17.6과 본 정책 동시 점검

---

## 21. Dev-Only Commands Isolation (97/98/99 Series)

`97-*`, `98-*`, `99-*` prefix 슬래시 커맨드 + 산출물은 로컬 moai-adk 개발 전용. `internal/template/templates/` 어디에도 흔적 금지. 배포 금지 파일 일람, 검증 체크리스트, 위반 시 영향, 신규 dev-only 워크플로우 추가 절차 등 전체 doctrine은 외부 파일 참조.

See: `.moai/docs/dev-only-commands-isolation.md`

---

## 22. Dev Settings Intent — local settings.json 의도 명문화

Workflow audit 2026-05-16 finding M2 후속. 로컬 `.claude/settings.json`의 몇 가지 키는 template baseline과 의도적으로 다르게 운용되며, 본 섹션은 그 의도를 명문화한다.

### §22.1 defaultMode

- **로컬값**: `"bypassPermissions"` 또는 `"acceptEdits"` (개발자 선호)
- **Template 기본값**: 미지정 (Claude Code 기본 `"default"` 사용)
- **의도**: 메인테이너는 빠른 실험 + bypass 모드 빈번 사용. 사용자 프로젝트는 안전한 prompt-each-time 기본값을 따른다.

### §22.2 enableAllProjectMcpServers

- **로컬값**: `true`
- **Template 기본값**: 미지정 (false 효과)
- **의도**: 메인테이너 머신에는 dev tool 다수 (pencil, chrome-devtools, claude-in-chrome 등)가 등록되어 있어 모두 자동 활성화하는 것이 효율적. 사용자 프로젝트는 `mcp__context7`만 `alwaysLoad`되고 나머지는 ToolSearch preload 경로를 따른다. (Sequential Thinking MCP는 SPEC-V3R6-SEQ-THINKING-RETIRE-001에서 retired — `ultrathink` 키워드로 대체.)

### §22.3 teammateMode

- **로컬값**: `"glm"` 또는 `"claude"` (runtime-managed by `moai cg` / `moai glm` 명령)
- **Template 기본값**: 미지정
- **의도**: 메인테이너는 CG mode (Claude leader + GLM teammates)로 cost-optimization 검증을 빈번하게 수행. 사용자 프로젝트는 leader 단독으로 시작 후 필요시 `moai cg` 진입.
- **주의**: §2 [HARD] settings.local.json Separation 참조 — teammateMode는 `settings.local.json`에 위치하며 template에 절대 포함 금지.

### §22.4 env.PATH

- **로컬값**: `$HOME/...` 패턴 (사용자 절대경로 `/Users/goos/...` 금지, 2026-05-17 정정 — workflow audit finding F-009/M5)
- **Template 기본값**: `settings.json.tmpl`이 `{{.GoBinPath}}` 등 Go template 변수로 렌더링
- **의도**: 로컬 settings는 fork/clone 시에도 깨지지 않도록 `$HOME` 환경변수로 추상화. Claude Code가 PATH 키 값을 expand하므로 `$HOME` 직접 사용 가능.

### §22.5 운영 원칙

- [HARD] 메인테이너 머신에서 위 키들을 변경할 때 template 자동 동기화 금지 (§2 settings.local.json Separation 적용)
- [HARD] 위 4개 키의 의도가 변경되면 본 §22를 즉시 갱신
- [HARD] 사용자 프로젝트에 위 키들이 누락된 것이 정상 — 누락은 결함이 아니라 의도된 격리

---

## 23. Local Git Workflows + Hook Setup (Hybrid Trunk 1-person OSS)

2026-05-22 commit `cd9eead14` (`chore(config)`)로 1인 OSS Hybrid Trunk policy 채택. main 직접 push 허용 + auto_branch/auto_pr 활성. 본 섹션은 정책 운영 시 마주치는 6가지 오류/경고 패턴과 처리 절차를 정리.

### §23.1 pre-push hook (manual setup — local infra)

`.git/hooks/pre-push`는 git infra (local-only). Template 동기화 안 됨. 다른 머신 clone 시 수동 설치 필요.

**현재 (2026-05-22~) — Warn-only + 5s sleep**:
- main 직접 push 시 경고 출력 + 5초 대기 (Ctrl+C로 취소 가능)
- ALLOW_MAIN_PUSH=1 escape hatch 불필요 (차단 모드 폐기)
- 보호 장치 4중: pre-commit hook (enforce) + CI workflows (main push 트리거) + GitHub branch protection (4 status checks) + Conventional Commits / Release Drafter (audit)

**다른 머신 manual setup**:

```bash
cat > .git/hooks/pre-push <<'EOF'
#!/bin/bash
while read local_ref local_sha remote_ref remote_sha; do
  if echo "$remote_ref" | grep -qE "refs/heads/main$"; then
    echo "⚠️  main 직접 push — Hybrid Trunk (모든 tier 허용) | CI 자동 트리거" >&2
    sleep 5
  fi
done
exit 0
EOF
chmod +x .git/hooks/pre-push
```

**Hook 동작 검증** (dry-run):

```bash
echo "refs/heads/main 0000 refs/heads/main 0000" | .git/hooks/pre-push  # warn + 5s + exit 0
echo "refs/heads/feat/test 0000 refs/heads/feat/test 0000" | .git/hooks/pre-push  # silent + exit 0
```

### §23.2 GitHub branch protection 현황 (modu-ai/moai-adk main)

`gh api repos/modu-ai/moai-adk/branches/main/protection` 조회 결과 (2026-05-22):

| 설정 | 값 | 의도 |
|------|------|------|
| `required_status_checks.strict` | `true` | up-to-date 강제 (병합 전 rebase) |
| `required_status_checks.contexts` | 4개 (Test ubuntu / Lint / Build linux/amd64 / CodeQL) | CI 보호 (main push에도 작동) |
| `required_approving_review_count` | `0` | 1인 OSS — self admin merge 허용 |
| `enforce_admins` | `false` | admin이 정책 bypass 가능 |
| `allow_force_pushes` | `false` | history 보호 |
| `allow_deletions` | `false` | branch 삭제 보호 |
| `required_conversation_resolution` | `true` | PR 대화 해결 필수 |
| `required_signatures` | `false` | GPG signing 강제 안함 |

조정 필요 시: `gh api -X PATCH repos/modu-ai/moai-adk/branches/main/protection ...`

### §23.3 운영 패턴 — A4: `gh pr merge --delete-branch` fatal

**증상**: PR admin merge 후 `fatal: Not possible to fast-forward, aborting`

**근본 원인**: gh CLI가 머지 직후 자동 `git pull --ff-only` 시도. 로컬 main이 머지된 PR squash commit과 분기되어 fast-forward 불가.

**핵심**: **실제 머지는 GitHub에서 완료된 상태** (`gh pr view <PR> --json state` → MERGED 확인). 로컬 동기화만 별도 필요.

**처리 절차**:

```bash
gh pr view <PR> --json state,mergedAt    # MERGED 확인
git fetch origin main
git reset --keep origin/main             # --hard 차단 우회 (§23.5)
```

### §23.4 운영 패턴 — A5: `git stash pop` 부분 적용 silent skip

**증상**: `git stash pop`이 일부 파일만 복원 + 나머지 파일 silent skip + "stash entry is kept in case you need it again."

**근본 원인**: stash 파일과 working tree 파일이 충돌하지 않더라도, git이 정책상 일부 적용 후 stash 보존. Silent skip은 표면화 안 됨.

**처리 절차** (명시적 복원):

```bash
git stash show --stat stash@{0}                              # 누락 진단
git checkout stash@{0} -- <missing-path-1> <missing-path-2>  # 명시 복원
git restore --staged <paths>                                 # unstage (필요 시)
git stash drop stash@{0}                                     # cleanup
```

### §23.5 운영 패턴 — A6: `git reset --hard` sandbox 자동 차단

**증상**: Claude Code sandbox에서 `git reset --hard` 명령 자동 거부 (Permission Denied)

**근본 원인**: claude-code sandbox가 destructive 명령 (`--hard`, `--force`, `rm -rf`, …)를 명시적 사용자 권한 없이 차단.

**우회 절차** (--keep equivalent + 안전):

```bash
# 1. dirty preserve
git stash push --include-untracked -m "phase-d $(date -u +%Y%m%dT%H%M%SZ)"

# 2. safe reset (--hard 대신 --keep)
git fetch origin main
git reset --keep origin/main   # local modifications 자동 보호

# 3. stash pop + 누락 명시 복원 (§23.4)
git stash pop || git checkout stash@{0} -- <paths>
```

`--keep`는 `--hard`와 달리 working tree에 commit되지 않은 변경이 있으면 reset 자체를 거부하지만, stash로 working tree가 clean한 상태에서는 `--hard`와 동등 효과.

### §23.6 운영 패턴 — Late-Branch Phase D 2중 보호

orphan commits 보존 + dirty 보존 + reset + stash pop 5단계:

```bash
git branch save-orphan-$(date +%Y-%m-%d) <latest-local-commit>             # 1) orphan 보존
git stash push --include-untracked -m "phase-d-$(date -u +%Y%m%dT%H%M%SZ)" # 2) dirty 보존
git fetch origin main                                                       # 3) 원격 최신
git reset --keep origin/main                                                # 4) 정렬 (§23.5)
git stash pop || git checkout stash@{0} -- <missing-paths>                  # 5) 복원 (§23.4)
```

선례: SPEC-V3R6-HARNESS-RENAME-001 sync (PR #1043) + chore PR #1044 (2026-05-22).

### §23.7 [HARD] 운영 원칙

- [HARD] pre-push hook은 `.git/hooks/`에 위치 — template 동기화 불가, 다른 머신 manual setup 필수
- [HARD] GitHub branch protection 변경은 `gh api -X PATCH` 명시적 수정으로만 (Settings UI 사용 시 audit trail 손실)
- [HARD] `git reset --hard` 대신 `--keep` 사용 (sandbox 안전)
- [HARD] `gh pr merge --delete-branch` 후 fatal 발생 시 `gh pr view --json state` 별도 확인 (실제 머지 여부)
- [HARD] `git stash pop` 결과는 `git status` 별도 검증 필수 (silent skip 가능성)
- [HARD] 1-person OSS Hybrid Trunk: 모든 tier (S/M/L) main 직진 push 허용 — CI 4 status checks + pre-push hook 5s warn + Conventional Commits + Release Drafter 4중 보호 (§23.0 chore commit `cd9eead14`, 2026-05-22 채택). feat 브랜치 + 자동 PR은 사용자가 명시적으로 review round 필요하다고 결정한 경우 (예: cross-team review, security-sensitive change) opt-in으로만 사용

### §23.9 Tier-based PR Routing (REQ-ATR-020 — SPEC-V3R6-AGENT-TEAM-REBUILD-001)

[HARD] **§23.7의 "모든 tier main 직진" 일반화에 대한 Tier-based 예외 명문화.** Hybrid Trunk 1-person OSS 정책의 기본은 모든 tier (S/M/L) main 직진 push이지만, SPEC tier 가 L 이거나 사용자가 명시적으로 `--pr` 플래그를 사용한 경우 `manager-git` 서브에이전트로 PR 생성을 routing한다.

| Tier / 조건 | 기본 routing | Owner | 비고 |
|------------|-------------|-------|------|
| Tier S (< 300 LOC, < 5 files) | main 직접 push | manager-develop / manager-docs (commit 직접 수행) | Hybrid Trunk 기본 — CI 4 status checks + pre-push hook 5s warn |
| Tier M (300-1000 LOC, 5-15 files) | main 직접 push | manager-develop / manager-docs (commit 직접 수행) | Hybrid Trunk 기본 |
| Tier L (> 1000 LOC OR > 15 files OR constitutional) | `feat/SPEC-XXX` 브랜치 + `gh pr create` | **manager-git** | Tier L 또는 사용자 `--pr` 플래그 시 PR routing |
| Tier S/M + 사용자 `--pr` opt-in | `feat/SPEC-XXX` 브랜치 + `gh pr create` | **manager-git** | 사용자 명시적 review round 요구 시 (cross-team review, security-sensitive change 등) |

**Owner 명시 (REQ-ATR-020 정합)**: Tier L OR `--pr` 케이스에서 PR 생성은 `manager-git` 의 책임이다. `manager-develop` 또는 `manager-docs` 는 PR 생성을 직접 수행하지 않으며, commit 만 수행 후 `manager-git` 에게 PR 생성을 위임한다. 이는 Anthropic 2026 SRP (Single Responsibility Principle) 정합 — 각 retained agent 가 명확한 phase boundary 를 가진다.

**Late-Branch 4-Phase Pattern**: Tier L PR routing 시 `manager-git` 은 `.moai/docs/git-workflow-doctrine.md` §18.3.1 의 Late-Branch 4-Phase 패턴 (A: branch creation / B: commit / C: PR creation / D: Late-Branch closure)을 따른다. Phase D Late-Branch closure 는 PR 머지 후 local main 정렬 의무 — `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern 참조.

**Routing 결정 흐름**:
1. SPEC tier 가 L → `manager-git` routing (자동)
2. 사용자가 `/moai sync --pr` 또는 `/moai run --pr` 명시적 사용 → `manager-git` routing
3. 그 외 (Tier S/M without `--pr`) → main 직접 push (manager-develop/manager-docs commit 직접 수행)

상위 SPEC 참조:
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md` REQ-ATR-020 (manager-git PR doctrine reconciliation)
- `.moai/docs/git-workflow-doctrine.md` §18.3.1 [HARD] Tier-based PR Routing (SPEC-V3R6-AGENT-TEAM-REBUILD-001 REQ-ATR-020) — M5 NEW section
- `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern
- `.claude/skills/moai/workflows/sync.md` § Phase Owners (Tier L OR `--pr` 플래그 시 manager-git)

### §23.8 [HARD] Multi-Session Race Mitigation

동일 project root + 동일 memory hash (`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)에서 2개 이상 Claude Code 세션이 동시에 작업할 때 race condition이 빈번 발생. 메모리 공유 + git working tree 공유로 양 세션 모두 같은 paste-ready resume을 보고 같은 SPEC을 동시에 진행 시도.

**선례**: SPEC-V3R6-LEGACY-CLEANUP-001 sync-phase race (2026-05-23) — parallel session이 spec.md frontmatter status `draft → implemented`를 commit `aea0cf7b9`로 별도 push, 내 세션 manager-docs는 "이미 올바른 상태"로 감지 (`git diff` skip). 다행히 conflict 회피했으나 push range mismatch (`aea0cf7b9..19bc873ff` instead of `ccd1fa9cf..19bc873ff`)로 retrospectively 감지.

**완화 정책 4중 (Defense in Depth)**:

1. **[HARD] Pre-spawn fetch obligation**: `.claude/rules/moai/core/agent-common-protocol.md` §Pre-Spawn Sync Check (L1) — implementation Agent spawn 전 `git fetch origin && git rev-list --count --left-right origin/main...HEAD` 의무. `N 0` (origin ahead) 감지 시 STOP + AskUserQuestion (rebase / inspect / abort 3 옵션).

2. **[SHOULD] Multi-session 인지 시 L2/L3 worktree opt-in 권장**: 사용자가 동일 cwd에서 2+ 세션 작업 패턴이면 `/moai plan --worktree` 또는 `moai worktree new SPEC-XXX --base origin/main`으로 SPEC별 working tree 분리. Memory는 여전히 공유되나 git working tree는 분리 → race 원천 차단. CLAUDE.md §14 [SHOULD] worktree advisory + session-handoff.md Block 0 패턴 활용.

3. **[SHOULD] Paste-ready resume 단일 세션 처리 discipline**: 사용자 수동 규율 — paste-ready resume은 1 세션에서만 paste. 다른 세션에서는 별도 SPEC 작업 OR read-only 활동 (`Agent(Explore)` 또는 `Agent(general-purpose)` diagnostic — 과거 `manager-quality` diagnostic은 archived per SPEC-V3R6-AGENT-TEAM-REBUILD-001). Memory hash 공유로 인한 paste-ready 동시 consume 회피.

4. **[INFO] Detection signal**: `git log --oneline` mystery commit 발견 시 (예: 본인이 commit 안 한 SPEC ID commit이 main에 등장) parallel session race 정황. retrospective 감지이지만 향후 sync 전 fetch 필요성 명시.

**Multi-session pattern 판단 기준**:
- 사용자가 명시적으로 2+ terminal/IDE 띄워 사용 중 (예: 한 세션 plan / 다른 세션 review)
- `ps aux | grep claude` 또는 `tmux list-panes` 다중 결과
- mystery commit 1회 이상 발견된 경험 있음

본 정책 적용 시 §23.7 일반화 (모든 tier main 직진)는 single-session 환경 기본값. Multi-session 시 사용자는 L2 worktree로 자발적 분리 OR feat 브랜치 + PR opt-in 활용.

선례: SPEC-V3R6-LEGACY-CLEANUP-001 race incident (2026-05-23) + agent-common-protocol.md §Pre-Spawn Sync Check L1 정책 도입.

---

## 24. Harness Namespace 분리 정책

[HARD] Skills + Agents의 namespace는 **"범용 배포"** vs **"사용자 생성"** 으로 명확히 분리한다.

### §24.1 Skills Namespace

| Prefix | 범위 | Source of Truth | `moai update` 영향 |
|--------|------|-----------------|---------------------|
| `moai-foundation-*` / `moai-workflow-*` / `moai-domain-*` / `moai-ref-*` / `moai-meta-*` | 범용 배포 — moai-adk 패키지에 포함, 모든 사용자 프로젝트에 deploy | template | sync (overwrite local) |
| `moai-harness-*` | **하네스 빌더 (builder/lifecycle)** — moai-adk 패키지가 제공하는 generator/learner. 현재 `moai-meta-harness` + `moai-harness-learner`만 해당 | template | sync |
| **`harness-*`** | **사용자 생성** — `moai-meta-harness`가 `/moai project` Phase 5+ 인터뷰 후 사용자 프로젝트 도메인에 맞춰 generate | user project | **NOT synced (보호)** |

[HARD] 사용자 프로젝트별 도메인 specialist skill은 **`harness-*` prefix만** 사용. `moai-harness-*` 또는 다른 `moai-*` prefix로 emit하면 contract 위반.

### §24.2 Agents Directory

| Path | 범위 | Source of Truth | `moai update` 영향 |
|------|------|-----------------|---------------------|
| `.claude/agents/moai/` | manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness (retained 7, FLAT layout per v.2.x baseline) | template | sync |
| **`.claude/agents/harness/`** | **사용자 생성 domain specialist agents** — `moai-meta-harness`가 `/moai project` Phase 5+ 인터뷰 후 generate | user project | **NOT synced (보호)** |

[HARD] `internal/template/templates/.claude/agents/harness/` 디렉토리는 **존재 자체가 금지**. template에는 `{core,expert,meta}/` 만 mirror. `harness/` directory 등장 시 cleanup chore + 본 §24 cross-reference.

### §24.3 운영 원칙

- [HARD] `moai-harness-*` prefix로 사용자 프로젝트별 skill generate 금지 — `moai-meta-harness`는 `harness-*` prefix만 emit
- [HARD] template (`internal/template/templates/`)에 `harness-*` skill 또는 `.claude/agents/harness/*-specialist.md` 누출 금지
- [HARD] `moai update`의 namespace 보호 contract: `harness-*` skill + `.claude/agents/harness/` 디렉토리는 sync 대상 제외 (user-owned)
- [HARD] `moai-meta-harness` skill 본체는 `moai-*` namespace (generator/builder이므로 범용 배포 대상)
- 선례: chore commit `4f1135684` (2026-05-23) — moai-adk-go 도메인 specialist 4 agent + `moai-harness-cli-template` / `moai-harness-patterns` 2 skill 잘못된 누출을 제거하면서 본 정책 명문화. 정정 전 SPEC-V3R6-HARNESS-RENAME-001 (PR #1043, 2026-05-22)의 my-harness → moai-harness 통합은 본 namespace 분리 정책 도입으로 부분 supersede됨
- 후속: 2026-05-26 prefix doctrine을 `my-harness-*` → `harness-*` 으로 migration (이번 Phase 1 doctrine-only 변경). Go code enforcement (`internal/cli/update.go`, `internal/cli/update_archive.go`, `internal/cli/update_preserve_inventory.go`, `internal/harness/prefix_conflict.go`, test fixtures)는 여전히 `my-harness-*` enforce 상태이며 별개 SPEC (가칭 `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`, Tier M, 39 Go files + 30+ tests scope)에서 catch-up 예정. **그 때까지 `harness-*` prefix actual generation 금지** — 새 prefix는 protection 안 받으므로 `moai update` 시 삭제 위험.

### §24.4 `moai update` 동작 Contract

[HARD] `moai update`는 `.claude/skills/` + `.claude/agents/` 에 대해 다음 동작을 수행한다:

| Namespace / Path | 동작 | 백업 정책 |
|------------------|------|-----------|
| `.claude/skills/moai-*` (incl. `moai-harness-*`, `moai-meta-*`, `moai-foundation-*`, `moai-workflow-*`, `moai-domain-*`, `moai-ref-*`) | **삭제 후 신규 설치** (overwrite) | 백업 불필요 — template-managed, 사용자 수정 시 손실됨 |
| **`.claude/skills/harness-*`** | **절대 삭제 금지 + 절대 modify 금지** | **백업 + 보존** (user-owned, Phase 2 SPEC catch-up 후 Go enforcement 작동) |
| `.claude/agents/moai/` | 삭제 후 신규 설치 (overwrite) | 백업 불필요 — template-managed (FLAT layout per v.2.x baseline) |
| **`.claude/agents/harness/`** | **절대 삭제 금지 + 절대 modify 금지** | **백업 + 보존** (user-owned) |
| 기타 사용자 직접 추가 자산 (`.claude/agents/<custom>.md`, `.claude/skills/<custom>/` 단 prefix가 `moai-` 시작 아닌 것) | 보존 | 백업 + 보존 |
| `.moai/harness/` (main.md, interview-results.md, extensions) | 절대 삭제 금지 | 백업 + 보존 (user-owned) |

[HARD] `moai update` 실행 전 user-owned 자산은 **반드시 백업** — 갑작스러운 process kill 등 비정상 종료 시에도 손실 위험 0이어야 한다. 백업 위치: `.moai/backups/update-{ISO-DATE}/` 권장.

[HARD] 이 contract는 다음 SSOT와 일관성을 유지:
- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation 의 Storage Roots 표
- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention

[HARD] Go 구현 (`internal/cli/update.go`, `internal/cli/update_archive.go`)이 본 contract를 정확히 준수하는지는 SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (별도 작성 예정)에서 검증한다. 현재 본 contract는 정책 명문화이며, 코드 구현 검증은 후속 작업.

### §24.5 Phase 2 Drift Entry-Condition (2026-05-26 ~ Phase 2 SPEC 완료 전)

[HARD] 본 doctrine은 2026-05-26 chore commit으로 `harness-*` prefix를 user-owned namespace로 선언했으나, Go code (`internal/cli/update.go`, `internal/cli/update_archive.go`, `internal/cli/update_preserve_inventory.go`, `internal/harness/prefix_conflict.go`, test fixtures ~39 files)는 여전히 `my-harness-*` enforce 중이다. 이는 **의도된 doctrine-code drift**이며, 별개 SPEC (가칭 `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`, Tier M, 39 Go files + 30+ tests scope)에서 catch-up 한다.

drift 운영 원칙 (Phase 2 SPEC 완료 전):

- [HARD] 새 `harness-*` prefix로 **실제 skill generate 금지** — Go code가 protection 안 하므로 `moai update` 시 삭제 위험. 사용자 데이터 손실 critical.
- [HARD] `moai-meta-harness` actual emission은 **잠정 `my-harness-*` prefix 유지** — generator runtime behavior는 Phase 2 SPEC 완료 후 `harness-*`로 전환. 본 Phase 1 변경은 declarative intent만.
- [HARD] 본 doctrine 변경은 **intent declaration only**, runtime behavior 변경 아님. `moai update`, `moai-meta-harness` Phase 4/5 generation, `internal/harness/prefix_conflict.go` 모든 enforcement layer는 `my-harness-*` 그대로 작동.
- [HARD] CI test `TestNamespaceLeakMyHarnessSkills` (`internal/template/namespace_protection_audit_test.go`)는 Phase 2 SPEC 진행 시 `harness-*` 패턴으로 갱신 — 현재 prefix `my-harness-` substring hardcoded는 본 Phase 1과 무관하게 작동 유지.
- [SHOULD] Phase 2 SPEC entry-condition: `harness-*` 패턴으로 Go enforcement + test fixture + `moai-meta-harness` runtime behavior 전환 + substring conflict 검증 (`harness-*` vs `moai-harness-*` 정확 분리) 모두 atomic하게.

본 §24.5 노트는 Phase 2 SPEC sync-phase 종료 시 제거 (drift 해소 완료 후).

---

## 25. Template Internal-Content Isolation

[HARD] `internal/template/templates/` 하위 산출물은 **moai-adk를 사용하는 외부 사용자에게 deploy되는 범용 자산**이며, **moai-adk 내부 개발 과정의 흔적**을 포함해서는 안 된다. 사용자 프로젝트에 deploy되었을 때 의미를 가지지 않거나 (예: 다른 사용자의 SPEC ID), 잘못된 도메인 신호를 주거나 (예: REQ-ATR-* 등 moai-adk 내부 추적 ID), audit trail을 노출하면 (예: "Audit 3 Findings A1-A6") template의 본분이 깨진다. 본 §25는 **§15 (16-language neutrality)** + **§21 (97/98/99 dev-only commands isolation)** + **§24 (harness namespace separation)** 와 동일한 isolation doctrine 계열에 위치하며, 각각 다른 차원의 "내부 ≠ 배포" 경계를 정의한다.

### §25.1 정의 — Allowed vs Forbidden Content Classes

[HARD] template에 **포함 가능한** content classes (per REQ-TII-013):

1. **Generic prose**: 범용 정책 설명 (예: "범용 배포 vs 사용자 생성 분리", "byte-for-byte mirror parity")
2. **Generic mechanism descriptions**: 메커니즘 설명 (예: "predecessor SPEC supersession via frontmatter status: superseded")
3. **Generic examples**: 도메인-중립적 예시 (예: "When user runs `moai update`, the system shall ...")
4. **External public references**: 공개 자료 인용 (예: "Per Anthropic Claude Code documentation at claude.com/docs/en/sub-agents, subagents cannot spawn other subagents")
5. **Permanent rule citations**: 영구적 규칙 인용 (예: ".claude/rules/moai/core/agent-common-protocol.md § User Interaction Boundary")
6. **MoAI-ADK system identifiers**: 시스템 정체성 자체 (예: "MoAI-ADK", "MoAI orchestrator", "MoAI agent catalog")

[HARD] template에 **절대 포함 금지** content classes (per REQ-TII-001):

1. **moai-adk internal SPEC IDs**: 본 프로젝트의 SPEC 식별자 (예: `SPEC-V3R6-AGENT-TEAM-REBUILD-001`, `SPEC-V3R6-WORKFLOW-OPT-001`)
2. **moai-adk internal REQ tokens**: 본 프로젝트의 requirement 식별자 (예: `REQ-ATR-007`, `REQ-WO-013`, `REQ-COORD-018`)
3. **moai-adk internal AC tokens**: 본 프로젝트의 acceptance criterion 식별자 (예: `AC-ATR-022`, `AC-WO-007`)
4. **Audit/post-mortem citations**: 본 프로젝트의 audit 인용 (예: "Audit 3 Findings A1-A6", "verbatim Anthropic 2026 sources cited in spec.md §B.1")
5. **Internal session dates**: 본 프로젝트 작업 날짜 (예: "2026-05-25", "2026-05-23", "added 2026-05-22") — 단, generic prose 안의 "today" / "yesterday" 같은 상대 표현은 OK
6. **Internal archive paths**: 본 프로젝트 archive 경로 (예: `.moai/backups/agent-archive-2026-05-25/`)
7. **Internal commit SHAs**: 본 프로젝트 commit hash (예: `b957a4d04`, `d9838995d`) — 단, generic prose 안의 short-sha mention `abc1234 `로 끝나는 trailing space 패턴은 grep 노이즈로 허용 (D-007 inline 해소; M3 lint test 참조)
8. **Internal memory file references**: 본 프로젝트 memory hash 경로 (예: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)

### §25.2 Forbidden / Allowed Worked Examples

| Class | Forbidden 예시 | Allowed Substitution 예시 |
|-------|---------------|--------------------------|
| SPEC ID literal | `Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 (2026-05-25), the agent catalog ...` | `Per the canonical MoAI agent catalog policy, the agent catalog ...` |
| REQ token | `REQ-ATR-008 specifies orchestration mode selection` | `The orchestration mode selection rule (.claude/rules/moai/workflow/orchestration-mode-selection.md) specifies ...` |
| AC token | `AC-ATR-022 verifies hook subagent boundary` | `The hook subagent boundary verification rule verifies ...` |
| Audit citation | `Per Audit 3 Finding A1 (verbatim Anthropic)` | `Per the canonical principle from Anthropic Claude Code documentation: "Subagents cannot spawn other subagents." (Source: https://claude.com/docs/en/sub-agents)` |
| Date | `Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 plan-phase commit (2026-05-25)` | `Per the canonical agent catalog policy` |
| Archive path | `12 agents archived at .moai/backups/agent-archive-2026-05-25/` | (omit entirely, or rephrase: `12 archived agents are preserved offline for reference`) |
| Commit SHA | `commit b957a4d04 introduced ...` | (omit entirely, or rephrase: `The canonical rule introduced ...`) |
| Memory ref | `Per ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_template_internal_content_isolation.md` | (omit; memory is per-user, never reference in template) |

### §25.3 Pre-commit Self-Check (5-item Mandatory Checklist)

[HARD] template 영역 (`internal/template/templates/**`) 의 staged changes commit 전에 다음 5개 항목을 **체크리스트로** 수동 통과시켜야 한다. 자동화는 §25.4 의 Go lint test (`TestTemplateNoInternalContentLeak`) 가 수행하며, 본 self-check는 그 lint test가 catch하지 못하는 의미적 누출까지 차단하기 위한 사람-인-더-루프(HITL) layer다.

- [ ] **C1 — SPEC ID literal**: staged diff 안에 `SPEC-V3R6-` / `SPEC-AGENCY-` / 기타 본 프로젝트 SPEC ID prefix가 등장하지 않는다. (단, `internal/template/templates/.moai/specs/.example/` 같은 example fixture 안의 `SPEC-XXX-001` 등 generic placeholder는 OK)
- [ ] **C2 — REQ/AC token**: staged diff 안에 `REQ-XXX-NNN` / `AC-XXX-NNN` 형태의 본 프로젝트 추적 token이 등장하지 않는다. (단, EARS/GEARS 본문 안에 `REQ-EXAMPLE-NNN` 등 generic placeholder는 OK)
- [ ] **C3 — Audit citation**: "Audit N Finding AX" / "verbatim Anthropic 2026 sources" / "spec.md §X.Y" 같은 audit 인용 패턴이 등장하지 않는다.
- [ ] **C4 — Date / short-sha**: 본 프로젝트 작업 날짜 (`2026-MM-DD` ISO 형식) 또는 commit short-sha (`[0-9a-f]{7,8}`) 가 prose 안에 등장하지 않는다. (단, `version: vX.Y.Z` 같은 SemVer는 OK. 또한 short-sha sentence-final 패턴은 D-007 inline 해소 후 lint test가 허용)
- [ ] **C5 — Memory/archive path**: `~/.claude/projects/-Users-goos-` / `.moai/backups/` 같은 본 프로젝트 maintainer-only 경로가 등장하지 않는다.

체크리스트 위반 발견 시: substitution dictionary (`.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md §B`) 의 generic-prose 치환 패턴을 적용한 후 다시 self-check 통과시키고 commit 진행.

### §25.4 Anti-pattern Catalogue (predecessor cleanup 사례 기반)

predecessor cleanup history (`research.md` §B 참조: chore commits `20a66df85` pass 1 + `40dc43f5b` pass 2) 에서 surface된 3개 anti-pattern:

#### Anti-pattern AP-25.1 — Audit/citation 본문 leak (Audit 3 Findings citation 패턴)

원천: M5/M7 작업이 NOTICE.md (`.claude/rules/moai/NOTICE.md` 의 template mirror) 에 "Audit 3 Findings A1-A6" 6개 verbatim Anthropic citation을 통째로 mirror.

위반 양상: template `.claude/rules/moai/NOTICE.md` 가 사용자 프로젝트에 deploy되면, 사용자는 "Audit 3"이 자기 프로젝트의 audit인지 moai-adk의 audit인지 알 수 없다. citation 자체는 공개 자료이지만 "Audit 3" 식별자는 moai-adk 내부 추적 번호.

올바른 패턴: NOTICE.md 본문에 verbatim citation은 유지 (license-required attribution) 하되, "Audit 3 Findings A1-A6" wrapper는 제거하고 "Anthropic 2026 Alignment" 같은 generic header 아래 직접 인용. SPEC ID, plan-phase commit hash, date 도 모두 제거.

탐지 패턴: `grep -rn 'Audit [0-9]\|Finding A[0-9]' internal/template/templates/`

#### Anti-pattern AP-25.2 — SPEC ID 본문 cross-reference leak

원천: 30+ template files (`agents/`, `rules/`, `skills/` 산하) 가 SPEC 본문에서 derive된 정책 변경 후 cross-reference로 "Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 (M1-M8 milestones)" 형태로 SPEC ID 인용을 추가.

위반 양상: 사용자 프로젝트에 deploy된 template에서 사용자가 `SPEC-V3R6-AGENT-TEAM-REBUILD-001`을 검색하면 자기 SPEC 디렉토리에 그런 SPEC이 없어 confusion. SPEC ID 자체는 본 프로젝트 추적용이지 정책의 본질적 일부가 아님.

올바른 패턴: SPEC ID cross-reference를 generic-prose 정책 인용으로 대체. 예: `Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 REQ-ATR-009 (lint + test + coverage delta gate)` → `Per the canonical sync-phase quality gate policy (lint + test + coverage delta)`. 추가 detail이 필요하면 `.claude/rules/moai/...` 파일 자체를 인용 (rule 파일은 template-mirror 형태로 같이 deploy 되므로 reference 유효).

탐지 패턴: `grep -rn 'SPEC-V3R6-\|SPEC-AGENCY-\|SPEC-WORKTREE-' internal/template/templates/`

#### Anti-pattern AP-25.3 — REQ token + date 본문 leak (canonical rule citation 변형)

원천: agent body 파일 (특히 `.claude/agents/moai/manager-*.md` template mirrors) 이 정책 변경 사유 footnote로 "Per REQ-ATR-008 + REQ-ATR-014 (added 2026-05-25)" 같이 REQ token + date 조합을 추가.

위반 양상: REQ token은 본 프로젝트 acceptance.md row 식별자. 사용자 프로젝트에는 acceptance.md 가 없어 token 자체가 의미 없음. date는 본 프로젝트 작업 history 노출.

올바른 패턴: REQ token + date를 모두 제거하고 정책의 essence만 prose로 유지. 정말로 추적이 필요하면 maintainer-only `CLAUDE.local.md` (현재 §25 같은 doctrine 섹션 안) 에 옮기고, template body는 generic 유지.

탐지 패턴: `grep -rn 'REQ-[A-Z][A-Z]\+-[0-9]\{3\}\|AC-[A-Z][A-Z]\+-[0-9]\{3\}\|20[2-9][0-9]-[0-1][0-9]-[0-3][0-9]' internal/template/templates/`

### §25.5 운영 원칙 + Cross-references

- [HARD] template 영역에 staged 변경이 있을 때 §25.3 5-item self-check **수동 통과 의무**. 자동화 (Go lint test `TestTemplateNoInternalContentLeak`) 와 병행하여 의미적 누출 방어막 (in-depth defense).
- [HARD] §25.1 forbidden classes 가 audit-time에 발견되면 즉시 cleanup chore commit + 본 §25 cross-reference. predecessor cleanup history pattern (`20a66df85` + `40dc43f5b`) 답습.
- [SHOULD] 새로운 forbidden class 등장 시 (예: 향후 "Phase 6 Findings" 같은 새 audit 형식) §25.1 forbidden list 확장 + §25.4 anti-pattern 사례 추가.
- Cross-references:
  - §15 — Template language neutrality (16-language equal treatment)
  - §21 — 97/98/99 dev-only commands isolation
  - §24 — Harness namespace separation (`harness-*` user-owned vs `moai-harness-*` template-managed)
  - SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 §B (Substitution Dictionary) — generic-prose 치환 패턴 SSOT
  - `internal/template/internal_content_leak_test.go` — automated regression guard (M3 deliverable)
