# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-05-25

---

<!-- LSEL:LOCAL-ONLY-DO-NOT-DISTRIBUTE — INVARANTS KERNEL (SPEC-LSEL-LOCAL-EVOLUTION-001 M2) -->
## §0. INVARANTS Kernel (read-only goal — LSEL)

> 이 블록은 LSEL 루프가 절대로 수정하지 않는 읽기 전용 골 커널이다. 모든 LSEL 제안(proposal)은
> 이 커널에 명시된 불변량을 위반하면 CSA forced-gate(REQ-LSEL-005) 대상이 된다. distributed doctrine
> (`CLAUDE.md`, `.claude/rules/moai/**`)이 아니다 — 이 파일은 local-only 가이드이며 템플릿에 미러되지 않는다.

**절대 위반 불가 (CSA forced-gate, 동기식 AskUserQuestion 강제):**

1. **분산 doctrine 불변량** — `internal/template/templates/**`, `CLAUDE.md`, `.claude/rules/moai/**`,
   retained agents (`.claude/agents/moai/**`), `moai*`/`moai-*` skills, frozen Go applier
   (`internal/harness/applier.go` with `enableTriggerInjectionWrites=false`)은 byte-for-byte 동결.
   LSEL 루프는 이들 표면에 한 글자도 쓰지 않는다 (REQ-LSEL-001/003).
2. **6 evolvable surfaces만 쓰기 허용** — `CLAUDE.local.md`, `.claude/settings.local.json`,
   `memory/`, `.claude/agents/harness/**` + `hns-*` skills, `.moai/lessons-inbox.jsonl`,
   `.moai/state/` (+ `.moai/state/lsel/`). 그 외 모든 경로는 hard-reject (REQ-LSEL-001).
3. **self-amending-handcuffs 금지** — frozen allowlist(`.claude/lsel/frozen-allowlist.json`),
   applier/curator 스킬 본체, `lsel-apply.sh`, `settings.local.json` hook-registration subblock은
   execution-meta 카테고리로서 동기식 승인 마커 없이는 기계적으로 거부된다 (REQ-LSEL-002/005 D3).
4. **AskUserQuestion은 orchestrator 전용** — LSEL 메커니즘 스킬(`hns-lsel-*`)은 subagent 맥락에서
   동작하며 절대 AskUserQuestion을 호출하지 않는다. blocker report 반환 (CLAUDE.md §8).

**LSEL Tier-4 DEAD 회피 (AC-LSEL-012):** `moai-harness-learner` Tier-4 AskUserQuestion flow는
production invocation layer에서 DEAD이다 (`tier4_firing_test.sh`로 검증). 따라서 LSEL PROPOSE→APPROVE
handoff는 Tier-4에 의존하지 않고 M3 fresh path(`hns-lsel-applier` + `decision.json`)로 간다.

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
moai hook <event>       # Execute hook handler
moai glm                # GLM worker mode
moai version            # Show version

# NOTE: There is NO top-level `moai build` command. To rebuild the binary
# after editing templates, run `make build` (templates are embedded via
# //go:embed all:templates in internal/template/embed.go — recompiled into
# the binary). See §2 Embedded Template System.
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

**§2.1 Template Content Neutrality — Acceptable Content Range for Templates**: When editing template source files in `internal/template/templates/`, ensure content adheres to the **acceptable** kept-classes (C1/C2/C4/C5/C6/C8) and excludes the FORBIDDEN content classes (SPEC IDs, REQ tokens, Audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references), enforced by CI guard (`.github/workflows/template-neutrality-check.yaml` trigger on path change). The canonical C1-C8 acceptable-vs-forbidden content-class catalogue lives in `.moai/docs/template-internal-isolation-doctrine.md §25.1` (cross-referenced by **§25 (Template Internal-Content Isolation)** of this file, now a stub). This ensures 16-language template distribution remains neutral to moai-adk internal development state.

**Pre-PR Verification (template contributor-checklist)** — before opening a PR that touches `internal/template/templates/**`, run the canonical 5-item pre-commit self-check (the CI guard `template-neutrality-check.yaml` is the safety net). See `.moai/docs/template-internal-isolation-doctrine.md` §25.3 for the full 5-item checklist and §25.1 for the forbidden/allowed content-class catalogue (C1-C8). (C3 dates + C7 commit-hashes are owned by the sibling `internal_content_leak_test.go` per §25, not this neutrality checklist.)

### Local-Only Files (Never in Templates)
```
.claude/settings.local.json    # Personal settings — runtime-managed, NEVER template
.claude/settings.json          # Rendered from .json.tmpl
.claude/agent-memory/          # Per-project agent memory
.claude/hooks/moai/handle-*.sh # Generated hook wrappers (not templates)
.claude/rules/moai/workflow/lifecycle-sync-gate.md         # Dev-only: maintainer lifecycle sync-gate rule (no template mirror, unreferenced by any shipped template file — intentional local-only)
.claude/rules/moai/workflow/repo-local-pr-policy.md        # Dev-only: repo-local all-tier PR policy override (Route A main-direct disabled by branch protection enforce_admins:true; no template mirror — intentional local-only)
.claude/commands/harness/{release-update,github,release}*  # Dev-only: split maintainer harness entries (§21)
.claude/commands/harness/release-update/manifest.json      # Dev-only: release-update harness manifest (§21)
.claude/workflows/hns-release-update-run.js                # Dev-only: release-update harness Runner (§21)
.claude/agents/harness/hns-{release-update,github,release}-specialist.md  # Dev-only: split harness specialists (§21, user-owned per §24)
scripts/ci-watch/              # Dev-only: CI watch loop scripts (5) — not distributed
scripts/ci-autofix/            # Dev-only: CI auto-fix scripts (4) — not distributed
.claude/skills/moai-workflow-ci-loop/                      # Dev-only: CI watch+autofix skill (removed from template; mirror kept)
.claude/rules/moai/workflow/ci-watch-protocol.md           # Dev-only: governs scripts/ci-watch (removed from template; mirror kept)
.claude/rules/moai/workflow/ci-autofix-protocol.md         # Dev-only original form: governs scripts/ci-autofix (template rewritten script-free)
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
.moai/config/astgrep-rules/sgconfig.yml                                    # Dogfood-only: experimental multi-lang ast-grep config (SPEC-ID 포함, mirror 시 §25 위반)
.moai/config/astgrep-rules/{cpp,csharp,elixir,flutter,go,java,javascript,kotlin,php,python,r,ruby,rust,scala,security,swift,typescript}/  # Dogfood-only: 실험적 언어별 ast-grep 룰(10/17 빈 stub); 배포 룰셋은 root go-hardcoding.yml만. 16-언어 룰셋 정식 배포는 후속 SPEC
```

> **§2.2 astgrep-rules 로컬 전용 예외 (2026-07-02)**: 로컬 `.moai/config/astgrep-rules/`의 언어별 서브디렉터리 트리 + `sgconfig.yml`은 dogfood-experimental(10/17 빈 `.gitkeep` stub, 나머지는 데모성 스캐폴드, 메시지 언어 혼재 ko/en, `sgconfig.yml`이 존재하지 않는 `utils` ruleDir 참조 + SPEC-ID 포함)이라 템플릿에 미러하지 않는다. 배포 사용자는 template-managed `go-hardcoding.yml`(root, SPEC-ID stripped) 1개를 baseline으로 받는다. **(2026-08-01 정정)** 종전 이 절은 "`gate.yaml`/`gate` 로더 부재 → `AstGrepGate.Enabled` 항상 컴파일 기본값 false"라고 적었으나 **두 주장 모두 현재 main에서 거짓**이다: 로더는 `internal/config/loader_gate.go` 에 존재하며 `internal/config/loader.go` 의 `Loader.Load` 에서 호출된다(SPEC-CONFIG-AUDIT-REPAIR-001 M2, PR #1142). 실제 기본값은 `internal/config/defaults.go` 의 `AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}` — 즉 **차단 없는 권고 모드로 켜져 있고**, 차단(blocking)만이 `gate.yaml` opt-in이다. 단, **릴리스된 `v3.0.1` 에는 로더가 없어**(`loadGateSection` 호출 0회) 해당 버전 사용자에게는 종전 서술이 여전히 사실이다 — 이것이 이슈 #1265 의 내용이며, 해결책은 코드 수정이 아니라 릴리스다. 16-언어 정식 룰셋 배포(메시지 영어 통일 + 데모 stub → 실제 패턴 + `utils` 정리 + SPEC-ID strip + `sg` config-mode 검증)는 별도 후속 SPEC 소관.

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
- **Source**: `internal/template/templates/` (edit here — this is the source of truth)
- **Embed mechanism**: `internal/template/embed.go` carries `//go:embed all:templates` + `//go:embed catalog.yaml`, which compile the `templates/` FS directly into the binary (there is NO generated `embedded.go` file)
- **Build**: Run `make build` after editing templates (recompiles the binary)

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
- [HARD] Set appropriate timeout. MoAI policy default is 5 seconds (the Claude Code platform default is 10 minutes; MoAI tightens this to 5 seconds to avoid stalling the session).

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

1. Environment Variables (override file values) — the actually-implemented
   config overrides read by `internal/config/manager.go` `applyEnvOverrides`:
   `MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR`,
   plus `MOAI_CONFIG_DIR` (config directory location). All env-var names are
   constants in `internal/config/envkeys.go`.
2. User Configuration: `.moai/config/sections/*.yaml`
3. Template Defaults: From `internal/template/templates/.moai/config/`

> NOTE: `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` are NOT currently
> implemented — no code in `internal/`/`pkg/`/`cmd/` reads them. User name and
> conversation language come from `.moai/config/sections/user.yaml` /
> `language.yaml` only. (Adding these env overrides would be a future
> enhancement, not current behavior.)

---

## 10. Build and Development Commands

빌드/테스트/린트 타깃은 `Makefile`에 `##` 도움말과 함께 정의돼 있다 — `make help` 로 조회. 템플릿 편집 → `make build` → 테스트 → 커밋 순서는 §2 [HARD] Template-First Rule 참조.

---

## 11. Frequent Issues and Solutions

### Issue: Templates not updated after editing

**Solution:**
```bash
# Recompile the binary (templates embedded via //go:embed all:templates in embed.go)
make build

# Verify the build succeeded
go build ./...
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

### §17.1 디자인 컴포넌트 + 아이콘 사용 규약 (Claude Warm Editorial)

docs-site는 Claude Warm Editorial 디자인 시스템(코랄 `#cc785c` · Pretendard · Goorm Sans Code · **라이트 단일 테마**)을 따른다. 토큰은 `static/moai-brand.css`(FROZEN) + `static/moai-design.css`, 폰트는 `layouts/partials/head/custom.html`(Pretendard Variable jsdelivr + Goorm Sans Code goorm CDN)에서 로드.

- [HARD] **다크 테마 미사용**: 라이트 단일. `[data-theme="dark"]` 분기는 dead code(신규 추가 금지). 코드 카드의 다크 배경(`#181715`)은 테마가 아닌 컴포넌트 스타일.
- [HARD] **이모지 대신 아이콘**: 본문 콘텐츠의 장식용 이모지(`📖 💡 🚀 ✨ 🎉 🔥 📌` 등) 대신 `{{</* icon <name> [variant] */>}}` shortcode(인라인 SVG)를 사용. 정의: `layouts/shortcodes/icon.html`. variant: `ok|warn|danger|primary|muted`. 아이콘: check, check-circle, x, x-circle, warning, info, bulb, rocket, star, flash, sparkles, target, package, book, search, wrench, database, rotate, clock, arrow-right.
  - **유지(이모지 아님 — 치환 금지)**: 타이포그래피 기호 `→ ← ↓ ✓ ✗`(흐름 서술); MoAI 오케스트레이터 배너 예시 코드블록 내 브랜딩 이모지(`🤖 🗿 📋 🎯` 등 — 출력 스타일 재현 목적). 시맨틱 표 마커(`✅`)는 `{{</* icon check ok */>}}`로 전환 가능하나 의미 보존 필수.
  - 신규/수정 콘텐츠는 본 규약 준수. **기존 콘텐츠 일괄 전환은 의미·4-locale 파리티·브랜딩 보존을 위해 파일별 판단으로 진행(blind sed 금지)**.
- [HARD] **사이드바 아이콘**: `data/menu/main.yaml`의 `icon:` 값은 `layouts/partials/menu.html` SVG switch에 대응 case 필수. 신규 icon 값 추가 시 menu.html에 path case도 함께 추가(미매칭 시 빈 svg 렌더 → 아이콘 누락).
- [HARD] **코드블록**: 모든 fenced code는 `layouts/_default/_markup/render-codeblock.html` render hook이 macOS 다크 카드로 렌더(트래픽라이트·언어 pill·복사 버튼·멀티라인 줄번호). Chroma `noClasses=false` + `.code-card .chroma` 스코프 코랄 신택스. 줄번호는 `lineNumbersInTable=true`. pre 여백은 geekdoc 기본 `padding:0`을 이기기 위해 `!important` 필요.
- [HARD] **Mermaid**: `foot.html`이 CDN UMD(`mermaid@10`, `window.mermaid` 노출)를 로드해 코랄 themeVariables(`theme:'base'`)로 렌더. geekdoc 내장 번들은 `window.mermaid` 미노출이라 사용 금지(기본 라벤더 자체 렌더됨).
- [HARD] **푸터**: geekdoc 기본 `.gdoc-footer` 다크틸 배경을 `.cw-footer` 라이트 캔버스로 오버라이드(`!important`).
- [HARD] **CSS 캐시 버스팅**: custom.html이 `hash.FNV32a (readFile ...)`로 CSS URL에 `?h=` 해시 부여(프로덕션 full build에서 정확). dev `hugo server`는 template 변경 시에만 해시 갱신되므로 CSS만 수정 후 미반영 시 서버 재시작 또는 하드 리로드.

---

## 18. Git Workflow — Enhanced GitHub Flow

v2.14.0 릴리스 이후 공식 채택. 5-axis 즉시 개선 (branch protection / label 3축 / merge strategy / Release Drafter / hotfix naming) + Enhanced GitHub Flow 11 branch prefix + Merge strategy 표 + BODP 3-Signal Evaluation + v2.14.0 Case Study + AskUserQuestion Enforcement Protocol 등 전체 doctrine은 외부 파일 참조. **(2026-07-20) branch protection 축이 `enforce_admins: true`로 적용 완료 — main direct push 전면 차단, 모든 tier PR 경유 (§18.7 + §18.3.1 참조). Hybrid Trunk main-direct는 RETIRED.**

See: `.moai/docs/git-workflow-doctrine.md`

---

## 19. AskUserQuestion Enforcement Protocol

> **[CANONICAL]** 본 섹션의 모든 enforcement 룰 — deferred tool preload 의무, pre-response self-check 4항목, anti-pattern 카탈로그, recovery protocol — 은 `.claude/rules/moai/core/askuser-protocol.md` 에 단일 진실 공급원(SSOT)으로 존재합니다. 본 §19은 cross-reference만 유지하며, 규칙 갱신 시 canonical 파일을 수정하세요.

### Local Notes

본 incident 기록 (2026-04-24): `~/.claude/projects/{hash}/memory/feedback_askuserquestion_enforcement.md`. v3.4.0부터 enforcement 정책 HARD 운영. 위반 탐지 시 즉시 canonical §Recovery Protocol 적용 + memory 추가 기록.

상위 정책 참조:
- CLAUDE.md §1 HARD Rules (AskUserQuestion-Only + Deferred Tool Preload)
- CLAUDE.md §8 User Interaction Architecture
- `.claude/skills/moai/SKILL.md` § Red Flags + Verification

### §19.1 구현 착수 승인 Mandatory Restoration (REQ-ATR-015 — SPEC-V3R6-AGENT-TEAM-REBUILD-001)

[HARD] **구현 착수 승인 (plan-to-implement HUMAN GATE)는 자율 bypass 대상이 아니다.** Plan-phase 산출물이 audit-ready 상태로 PASS 되었더라도, run-phase 진입 직전 orchestrator는 자율 흐름을 중단하고 사용자에게 명시적 진행 승인을 `AskUserQuestion`으로 받아야 한다. 이는 Anthropic Claude Code의 Ctrl+G plan editor mandate (plan-to-implement 경계에서 사용자 개입 의무)와 정합한다.

**skip-eligible 0.90 autonomous bypass 정책의 적용 범위**: `skip-eligible` (score ≥ 0.90) autonomous bypass는 **Phase 0.5 plan-auditor verdict 재실행에만** 적용된다 — CONST-V3R5-026 + `.claude/rules/moai/workflow/spec-workflow.md` § Plan Audit Gate skip policy 참조. **구현 착수 승인 (plan-to-implement HUMAN GATE)에는 적용되지 않는다**. Phase 0.5 SKIP과 구현 착수 승인 SKIP은 서로 다른 결정 — Phase 0.5는 plan-auditor의 verdict 재실행 여부 (자동화 가능), 구현 착수 승인은 사용자가 run-phase 진입을 승인할지 여부 (사용자 결정 필수).

**오케스트레이터 의무 (구현 착수 승인 entry)**:
1. Plan-phase 산출물 + plan-auditor verdict 요약을 사용자에게 prose로 제시
2. `ToolSearch(query: "select:AskUserQuestion")` preload
3. `AskUserQuestion` 으로 "run-phase 진입 / 추가 검토 / 중단" 3-option 제시 (첫 옵션 "(권장)" 라벨)
4. 사용자 응답 수신 후 run-phase 진입 (또는 중단)

**위반 anti-pattern**: Phase 0.5 verdict가 PASS skip-eligible (≥ 0.90)이라는 이유만으로 사용자 승인 없이 `/moai run`을 자율 시작하는 행위. 구현 착수 승인은 plan-auditor 점수와 무관한 별도 사용자 의지 확인 절차다.

상위 SPEC 참조:
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md` REQ-ATR-015 (구현 착수 승인 restoration)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E (구현 착수 승인 vs Phase 0.5 vs Phase 0.95 boundary)

---

## 20. Vercel Build Cost Guard

### [HARD] Build Machine = Elastic 유지

- Vercel Team default + 각 프로젝트 모두 **Elastic** 머신 사용. Turbo($0.126/min) 또는 Standard로 변경 금지 — Elastic은 $0.0035/CPU min로 약 40배 저렴
- 새 프로젝트 추가 시 Settings → Build and Deployment → Build Machine = Elastic 확인
- 비용 폭탄 의심 시 **가장 먼저 Build Machine 설정 점검**
- docs-site는 §17.6 Vercel 프로젝트 바인딩과 함께 운영 — 비용 의심 시 §17.6과 본 정책 동시 점검

---

## 21. Dev-Only Commands Isolation (Split Harnesses)

3개 split 메인테이너 하네스 (`/harness:release-update`, `/harness:github`, `/harness:release`) + 산출물은 로컬 moai-adk 개발 전용. `internal/template/templates/` 어디에도 흔적 금지 (CI guard: `internal/template/split_namespace_test.go` `TestSplitHarnessNamespaceNoLeak`, sentinel `SPLIT_HARNESS_NAMESPACE_LEAK`). 구 `97-*`/`98-*`/`99-*` 번호 커맨드는 한때 단일 unified 하네스로 통합되었다가 SPEC-V3R6-DEV-HARNESS-SPLIT-001 에서 3개 독립 하네스로 분리됨 (release-update 만 Runner+manifest 보유; github/release 는 thin command → specialist 직접). 배포 금지 파일 일람, 검증 체크리스트, 위반 시 영향, 신규 dev-only capability 추가 절차 등 전체 doctrine은 외부 파일 참조.

See: `.moai/docs/dev-only-commands-isolation.md`

---

## 22. Dev Settings Intent — local settings.json 의도 명문화

[HARD] 로컬 `.claude/settings.json`의 키 의도(defaultMode / enableAllProjectMcpServers / teammateMode / env.PATH / outputStyle / model / worktree auto-toggles / branch_guard.enabled)와 "의도된 격리" 정책 전문은 `See: .moai/docs/local-dev-settings-intent.md` (§22.1-§22.9). 본체 다이어트를 위해 외부 파일로 이관(2026-08-03). settings 키 의도가 변경되면 외부 파일을 갱신할 것. 핵심 원칙만 본체에 잔류: (1) `teammateMode`(`.claude/settings.local.json`, `"tmux"`/`""`) != `llm.yaml team_mode`(`cg`/`glm`/`""`) — 위치·값·용도 상이; (2) `env.PATH`는 Claude Code가 `$HOME`을 expand 안 하므로 `settings.local.json`에 절대경로; (3) 로컬 `model` 키 의도적 미탑재(last-choice 존중, 템플릿 `model: sonnet`은 보존); (4) worktree 세 토글 + branch_guard.enabled 모두 분산 기본 `false` — 감사 시 "결함"으로 되돌리지 말 것.

---

## 23. Local Git Workflows + Hook Setup (PR-mandatory 1-person OSS)

[HARD] **(2026-07-20 개정) 1인 OSS PR-mandatory 정책 — `enforce_admins: true`로 main direct push 완전 차단 (admin 포함).** 모든 변경 (daily Tier S/M commit 포함)은 PR 경유; self-merge 허용 (0 approvals, 4개 CI check 통과 시). 종전 "모든 tier(S/M/L) main 직접 push 허용" Hybrid Trunk 정책은 RETIRED. tier는 이제 main-direct 여부가 아니라 PR ceremony 무게(§23.9)에만 영향. tag push(`scripts/release.sh`)는 branch protection 무관 → tag flow 무영향. 다루는 주제: pre-push hook 수동 설치(§23.1, main엔 이제 redundant·harmless), GitHub branch protection 현황(§23.2, enforce_admins:true), 운영 오류 패턴 A4/A5/A6 + Late-Branch Phase D 2중 보호(§23.3–§23.6), [HARD] 운영 원칙(§23.7, PR-mandatory), Tier-based PR Routing(§23.9, 모든 tier PR), Multi-Session Race Mitigation 4중 방어(§23.8).

See: `.moai/docs/git-local-workflow-doctrine.md`

---

## 24. Harness Namespace 분리 정책

[HARD] Skills/Agents namespace는 "범용 배포" vs "사용자 생성"으로 분리한다. `moai-*` / `moai-harness-*` skill + `.claude/agents/{core,expert,meta}/` = template-managed (sync 시 overwrite) vs `hns-*` skill(정식, SPEC-HNS-PREFIX-RENAME-001) + 레거시 `harness-*`/`my-harness-*` skill + `.claude/agents/harness/` = user-owned (`moai update`가 절대 삭제·수정 금지, 반드시 백업+보존). `internal/template/templates/`에 `hns-*`/`harness-*` skill 또는 `.claude/agents/harness/` 디렉터리 누출 금지. §24.4 `moai update` 동작 contract(delete-vs-preserve 매트릭스) + §24.5 Phase 2 drift entry-condition 포함.

See: `.moai/docs/harness-namespace-doctrine.md`

---

## 25. Template Internal-Content Isolation

[HARD] `internal/template/templates/` 산출물은 외부 사용자에게 배포되는 범용 자산이며 moai-adk 내부 개발 흔적을 포함하면 안 된다. 금지 클래스: 내부 SPEC ID, REQ/AC 토큰, audit 인용("Audit N Finding AX"), 내부 작업 날짜, commit SHA, archive/memory 경로. 허용: generic prose, 메커니즘 설명, 공개 자료 인용, 영구 규칙 인용, MoAI-ADK 시스템 식별자. CI guard: `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml`. 5-item pre-commit self-check + Allowed/Forbidden content-class catalogue + anti-pattern catalogue(AP-25.1~25.3)가 포함된다. §15(언어 중립성)·§21(dev-only commands)·§24(harness namespace)와 동일 isolation doctrine 계열.

See: `.moai/docs/template-internal-isolation-doctrine.md`

---

## 26. Linear 연동 (개인/로컬 전용)

[ZONE:Local-Only] 개인/로컬 전용. 전문(워크스페이스 매핑 · 2계층 운영 모델 · Linear↔SPEC 상태 매핑 · MCP 도구 지침 · `idea:`/emoji 트리거)은 `See: .moai/docs/local-linear-integration.md`. 본체 다이어트를 위해 외부 파일로 이관(2026-08-03). 핵심만 잔류: 이 리포(moai-adk-go)의 Linear Project에 `Idea` 라벨 + Backlog(Triage) 이슈로 기록; SPEC 저작은 항상 리포 파일에서만(Linear로 옮기지 않음); `CLAUDE.local.md`는 git-tracked 공개 파일이므로 "로컬 전용"은 템플릿 미러 금지를 뜻함(비공유가 아님).

---

## 27. 에이전트-스킬 아키텍처 필수 조건 (2026-07-15 채택)

[HARD] 신규/수정되는 모든 에이전트·스킬 개발 시 아래 5개 조건을 필수로 준수한다.

### §27.1 에이전트별 스킬 제공 의무

- 모든 에이전트는 **1개 이상의 스킬 세트**를 제공받아야 하며, 스킬은 다음 4요소로 구성된다:
  1. **기본 워크플로우 스킬** — 에이전트의 핵심 작업 절차 (예: `moai-workflow-*`)
  2. **노하우 레퍼런스** — 도메인 지식/패턴 참조 (예: `moai-ref-*`)
  3. **스크립트 수행** — 실행 가능한 스크립트/검증 레시피 (bundled scripts, verify recipes)
  4. **호출 트리거** — 언제 이 스킬을 로드할지의 트리거 조건 (frontmatter description + Conditional Skill Loading)
- 연결 메커니즘: `skills:` frontmatter preload(≤2) + orchestrator 주입 `Skill()` 지시 (`.claude/rules/moai/workflow/skill-routing.md`).

### §27.2 스킬 언어

- [HARD] 모든 스킬 본문(SKILL.md + references + scripts 주석)은 **영어로 작성** (CLAUDE.md §9 "Commands, Agents, Skills Instructions: Always English"와 정합).

### §27.3 /moai:sub-commands 슬래시 래핑 유지

- `/moai:<sub>` 형태의 스킬 래핑 커맨드(deprecated command 기능 활용)를 사용자 편의 UX로 **유지·확장**한다 — `/` 입력 시 커맨드 리스트에 힌트 + 빠른 찾기 제공이 목적.
- 신규 서브커맨드 추가 시 `/moai:<sub>` 래퍼도 함께 생성한다 (template source 동기화 포함, §2 Template-First).

### §27.4 /moai:harness 메타 하네스

- `/moai:harness <자연어 요청>` 은 moai-adk 에이전트+스킬 자산을 활용한 **메타 하네스**로 동작한다: 스킬 체이닝 + 워크플로우 설계 + 에이전트 체이닝/위임 배정을 한 번에 수행 (v4 Builder: ANALYZE / PLAN / GENERATE / ACTIVATE).

### §27.5 분석-우선 실행 계획 (Analyze-First 강화)

- 사용자가 일반 요청을 하든 `/moai '요청'` 을 하든, 실행 전에 반드시: **요구사항 분석 → 계획 수립 → 스킬·에이전트 호출 계획 명시 → 진행**. 근거: CLAUDE.md §2 Request Processing Pipeline (①~⑤).
- 실행 계획에는 어떤 스킬을 로드하고 어떤 에이전트를 어떤 순서로 spawn할지가 포함되어야 하며, 사용자에게 제시 후 진행한다 (Rule 1 Approach-First + 구현 착수 승인 gate 유지).

---

## 28. LSEL (Local Self-Evolution Loop) 운용 지침 — SPEC-LSEL-LOCAL-EVOLUTION-001

> 본 절은 GOOS 로컬 전용 LSEL 루프의 운용 지침이다. 분산 doctrine이 아님 — 템플릿에 미러 금지.
> INVARANTS kernel은 파일 최상단 §0에 있다. 메커니즘 SSOT: design report
> `.moai/reports/moai-local-self-evolution-design-20260804.html`.

### §28.1 루프의 7단계 + REFLECTION

`OBSERVE → CLUSTER → PROPOSE → APPROVE → APPLY → REMEMBER → VERIFY` + 주기적 REFLECTION.
M1 = CLUSTER/drain 마감, M2 = PROPOSE shadow + INVARANTS kernel, M3 = APPLY(bypass)가 critical path.
분산 Go applier는 동결된 채(`enableTriggerInjectionWrites=false`) bypass로 닫는다 — 절대 unfreeze하지 않는다(AP-LSEL-002).

### §28.2 쓰기 허용 표면 (6 evolvable surfaces)

1. `CLAUDE.local.md`(본 파일, §0 INVARANTS kernel 제외), 2. `.claude/settings.local.json`,
3. `memory/`, 4. `.claude/agents/harness/**` + `hns-*` skills, 5. `.moai/lessons-inbox.jsonl`,
6. `.moai/state/` (+ `.moai/state/lsel/`). 그 외는 전부 FROZEN.

### §28.3 기계적 트리거 (REQ-LSEL-007)

- **SessionStart backlog check:** `.claude/skills/hns-lsel-curator/backlog_check.sh`가
  `wc -l .moai/lessons-inbox.jsonl`을 `drain-offset.json + N`(기본 N=25)과 비교하여 초과 시
  LSEL drain을 참조하는 system-reminder를 stderr로 발행(advisory, non-blocking).
- **default `/loop` recipe:** `.claude/workflows/lsel-drain-loop.js`(read-only drain 트리거).
  루프가 orchestrator의 기억에 의존하지 않게 한다(audit의 정확한 실패 모드, report §11 mustFix B#1).

### §28.4 PROPOSE shadow + 자가비판 게이트 (M2)

각 제안은 `.moai/state/lsel/proposals/<id>/{proposal.md, diff.patch, self-critique.md}` triple을
갖는다. `proposal.md`는 8-key schema(`proposal_id`, `target_surface`, `rationale`,
`WHY-not-just-WHAT`, `prediction`, `verify_command`, `blast_radius`, `memory_type`) +
`retrieval_evidence`를 담는다. retrieval-before-propose(Reflexion)는 필수 — 관련 `feedback_*.md`를
draft 전에 읽고 증거를 기록. 자가비판에 UNRESOLVED 이의가 있으면 `status: blocked`로 APPROVE 불가.

### §28.5 CSA forced-gate (REQ-LSEL-005) — 동기식 AskUserQuestion 강제

6 카테고리: INVARANTS kernel, security/validation exception, HIGH-fan-in refs, Bash risk path,
`permissions.allow`, execution-meta files. execution-meta 4종(allowlist, applier/curator 본체,
`lsel-apply.sh`, `settings.local.json` hook-registration)은 `decision.json`의 동기식 승인 마커
부재 시 기계적으로 거부(reject-log 행 추가 + write 없음). forced-gate는 bother-cost 면제.
`csa_refusal_test.sh`가 거부 규칙을 fixture로 검증한다.

### §28.6 Tier-4 DEAD — APPROVE는 fresh path로 (AC-LSEL-012)

`moai-harness-learner` Tier-4 AskUserQuestion flow는 production layer에서 DEAD
(`tier4_firing_test.sh` 검증). 따라서 LSEL의 PROPOSE→APPROVE handoff는 Tier-4를 타지 않고,
M3의 fresh path(`hns-lsel-applier` + `decision.json` 동기식 승인 마커)로 간다.

### §28.7 PR-mandatory + lsel-tagged commit

모든 apply는 feature branch + PR 경유(CLAUDE.local.md §23, `enforce_admins:true`). 각 apply는
`lsel-<proposal-id>`-tagged 단일 commit(REQ-LSEL-004) — rollback은 `git revert <lsel-tag>` 1-liner.
