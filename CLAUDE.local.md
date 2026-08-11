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

See: `.moai/docs/version-management.md` — SemVer 2.0.0 pre-release form (`-rc.N`), build version injection via ldflags, files requiring version sync, release process under the PR-mandatory regime.

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

See: `.moai/docs/hook-development.md` — shell-script-only hook pattern, hook wrapper template, settings.json format, `$CLAUDE_PROJECT_DIR` quoting rules, platform differences, timeout policy.

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

## References

Sections §18-27 were consolidated into external `.moai/docs/` files to reduce launch-time context. Each entry below is the authoritative location for its domain.

- **§5 Version Management** (SemVer pre-release, ldflags injection, release process): `.moai/docs/version-management.md`
- **§7 Hook Development** (shell-script-only pattern, settings.json format, quoting rules): `.moai/docs/hook-development.md`
- **§18 Git Workflow** (Enhanced GitHub Flow, branch protection `enforce_admins: true`, Hybrid Trunk RETIRED): `.moai/docs/git-workflow-doctrine.md`
- **§19 AskUserQuestion Enforcement + §19.1 Implementation Kickoff Approval Mandatory Restoration** (REQ-ATR-015): canonical SSOT at `.claude/rules/moai/core/askuser-protocol.md` + `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E (the gate is mandatory and score-independent; plan-auditor PASS never auto-bypasses it)
- **§20 Vercel Build Cost Guard** [HARD]: all Vercel projects MUST use Elastic build machine ($0.0035/CPU min vs Turbo $0.126/min); check Build Machine setting first on cost anomalies
- **§21 Dev-Only Commands Isolation** (split harnesses, `SPLIT_HARNESS_NAMESPACE_LEAK` sentinel): `.moai/docs/dev-only-commands-isolation.md`
- **§22 Dev Settings Intent** (settings.json key semantics): `.moai/docs/local-dev-settings-intent.md`
- **§23 Local Git Workflows** (PR-mandatory 1-person OSS, all tiers via PR): `.moai/docs/git-local-workflow-doctrine.md`
- **§24 Harness Namespace** (template-managed vs user-owned separation): `.moai/docs/harness-namespace-doctrine.md`
- **§25 Template Internal-Content Isolation** (neutrality catalogue, CI guard): `.moai/docs/template-internal-isolation-doctrine.md`
- **§26 Linear 연동** (local-only): `.moai/docs/local-linear-integration.md`
- **§27 Agent-Skill Architecture** [HARD]: every agent gets ≥1 skill set (4 elements: workflow skill + knowhow reference + scripts + trigger); all skill bodies in English; `/moai:<sub>` slash-wrapping maintained; `/moai:harness` meta-harness (v4 Builder); Analyze-First execution plan before any work
