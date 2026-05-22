# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-04-11

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

### [HARD] 자가 점검 3 질문 (복잡 작업 시작 전 필수)

1. 이 작업은 전문 에이전트의 고유 도메인인가?
2. 해당 전문 에이전트가 카탈로그에 존재하는가? (CLAUDE.md Section 4)
3. 직접 수행보다 위임이 품질/독립성/편향 방지에 유리한가?

**3개 모두 YES → 직접 수행 금지**. AskUserQuestion으로 위임 방식 확인 후 실행.

### 수량 기반 트리거

- 같은 종류 파일 **5+** 생성 → 전문가 위임 강제
- Go 코드 **500+ LOC** 신규 → `expert-backend` 강제
- 에이전트/스킬 **3+** 생성 → `builder-agent`/`builder-skill` 강제

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

---

**Status**: Active (Local Development)
**Version**: 3.8.0 (§19 → canonical cross-reference로 축소, /doctor 경고 해소)
**Last Updated**: 2026-05-18

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
- **의도**: 메인테이너 머신에는 dev tool 다수 (pencil, chrome-devtools, claude-in-chrome 등)가 등록되어 있어 모두 자동 활성화하는 것이 효율적. 사용자 프로젝트는 `mcp__context7` / `mcp__sequential-thinking`만 `alwaysLoad`되고 나머지는 ToolSearch preload 경로를 따른다.

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
