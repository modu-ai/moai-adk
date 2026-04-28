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

`docs-site/`는 `adk.mo.ai.kr` 공식 사용자 문서. moai-adk-go 개발자는 코드 변경 시 이 섹션을 반드시 준수.

### §17.1 URL 표준 & 블랙리스트

공식 URL은 단 하나:

| 리소스 | URL |
|--------|-----|
| 문서 홈페이지 | `https://adk.mo.ai.kr` |
| GitHub | `https://github.com/modu-ai/moai-adk` |
| Discord | `https://discord.gg/moai-adk` |
| NPM | `https://www.npmjs.com/package/moai-adk` |

### [HARD] 금지 URL (사용 시 CI 실패)

- `docs.moai-ai.dev` (구 주소)
- `adk.moai.com` (오타)
- `adk.moai.kr` (오타: 정확히 `adk.mo.ai.kr`)

### URL 변경 체크리스트 (docs-site 내)

- `docs-site/hugo.yaml` → `baseURL`, `params.og`
- `docs-site/static/robots.txt` → `Sitemap:` 라인
- `docs-site/static/llms.txt` → 문서 링크
- `docs-site/layouts/partials/head.html` → 메타 태그
- `docs-site/vercel.json` → redirects

### §17.2 Markdown/Hextra 작성 규칙

**[HARD] 강조 표기와 괄호 사이 공백 필수** (Nextra 시대 유래, Hextra/Goldmark 검증 완료 후 유지 여부 재평가):

- 잘못: `**바이브코딩(Vibe Coding)**`
- 올바름: `**바이브코딩** (Vibe Coding)`

**[HARD] Mermaid 방향: TD only**

- 사용: `flowchart TD`, `graph TB`
- 금지: `flowchart LR`, `graph LR`

이유: 모바일 가독성 + 사이드바 좁은 화면 대응.

### §17.3 4개국어 동기화 규칙

### [HARD] Canonical source는 ko

번역 체인: `ko` → `en` → `zh`/`ja` 병렬. ko-only 머지 금지.

### 지원 locale (4개 고정)

- `ko` (한국어, 기본) / `en` (영어) / `ja` (일본어) / `zh` (중국어 간체)

### [HARD] 동시 업데이트 의무

기능 추가/변경/제거로 ko 문서가 바뀌면 **동일 PR 내에서 en/zh/ja 모두 반영**. 예외: 5,000 단어 이상 대규모 콘텐츠는 ko 머지 후 48h 이내 en, 추가 72h 이내 zh/ja. 해당 파일 front matter에 `translation_status: pending` 필수.

### PR 체크리스트

- [ ] ko/en/zh/ja 4개 locale 해당 파일 모두 수정
- [ ] 각 locale front matter의 `title`, `description`, `weight` 대응
- [ ] `CHANGELOG.md` 반영
- [ ] 버전 스냅샷 필요 여부 판단 (§17.4)
- [ ] Mermaid 노드 라벨 번역 (구문/방향 미변경)
- [ ] Anchor 링크가 번역된 heading slug 가리킴
- [ ] §17.1 금지 URL 재도입 없음
- [ ] 본문에 이모지 유입 없음

### 불일치 탐지 스크립트

`scripts/docs-i18n-check.sh` (Phase 5 산출) 실행 — pre-commit + CI에서 자동:

- 4개 locale 간 파일 개수/경로 일치
- 각 파일 front matter `title` 존재 (비어있지 않음)
- H1 heading 존재
- MoAI 용어집 (glossary) 준수 — "MoAI-ADK" 등은 모든 locale에서 불번역 유지

### §17.4 버전 스냅샷

| 릴리스 타입 | 스냅샷 생성 | 대상 |
|---------------|----------------|------|
| Major (vX.0.0) | YES | 4 locale 전체 `content/{locale}/v{X}/` 동결 복사 |
| Minor (v2.Y.0) | YES | 동일 |
| Patch (v2.12.Z) | NO | `latest`만 업데이트 |

릴리스 태그 푸시 → `manager-git` 트리거 → `scripts/docs-release-snapshot.go` 실행 → 과거 버전 URL 자동 생성.

사이트 상단 배너: 과거 버전 조회 시 "Viewing v2.X. Go to latest" 자동 표시.

### §17.5 실행 주체

- **기본**: `manager-docs` subagent (`/moai sync` 워크플로우)
- **번역**: `manager-docs` → en 완료 확인 → zh/ja 병렬 위임 (expert-docs)
- **검증**: `plan-auditor` — §17.3 스크립트 결과로 PR pass/fail
- **스냅샷**: `manager-git` — 릴리스 태그 자동 후처리

docs-site 파일을 직접 수정하는 에이전트 (`expert-frontend`, `manager-docs` 등)는 `isolation: worktree` 사용 권장.

### §17.6 빌드/배포 체크리스트

- [ ] `cd docs-site && hugo --minify` 성공 (zero warnings)
- [ ] `docs-site/public/sitemap.xml` 생성 확인
- [ ] Vercel Preview URL 수동 검증 — 4개 locale 홈 페이지
- [ ] Mermaid 다이어그램 렌더링 (light/dark mode)
- [ ] `robots.txt`, `llms.txt` 도메인 일치
- [ ] 언어 스위처 작동 (locale A → locale B navigation)

### [HARD] Vercel 프로젝트 바인딩

- 프로젝트 ID: `prj_EZaVdfE3gJeXVbizafBEECpniINP` (moai-docs에서 계승, 절대 변경 금지)
- 도메인: `adk.mo.ai.kr` → production branch 배포
- Root Directory: `docs-site/` (Vercel Dashboard 설정)
- Framework Preset: Hugo
- Node.js version: Hugo는 Node 불요, 단 `vercel.json`에서 Hugo 버전 명시 필수

---

## 18. Git Workflow — Enhanced GitHub Flow

v2.14.0 릴리스 이후 공식 채택. Gitflow 대안 비교 분석 결과 (v2.14 후 세션) 이 프로젝트의 1-2명 팀 + 2일/릴리스 cadence + 단일 CLI 배포 환경에서 Gitflow의 `develop` 이중 관리 부담을 상회하는 장점이 부재. 현재 de-facto 패턴을 유지하되 **formalization + automation**으로 개선.

### §18.0 [HARD] 운영 원칙 — 5가지 즉시 개선 Framework

이 프로젝트는 **다음 5가지 개선을 공식 운영 방침으로 채택**한다. 모든 git 작업은 이 5개 축을 준수해야 한다.

| # | 개선 항목 | 현재 구현 | 참조 |
|---|----------|-----------|------|
| 1 | **Branch protection rule 강제** (`main` + `release/*`) | ⏳ `gh api` 명령어 준비 완료, admin 적용 대기 | §18.7 |
| 2 | **Label 3축 체계** (`type:` / `priority:` / `area:`) + `status:` 보조축 | ✅ `.github/labels.yml` 정의 완료, 25 labels 추가됨 | §18.6 |
| 3 | **Merge strategy 명시** (release = merge commit, feature = squash) | ✅ `.github/PULL_REQUEST_TEMPLATE.md` + `git-strategy.yaml` | §18.3 |
| 4 | **Release Drafter로 CHANGELOG 자동화** | ✅ `.github/release-drafter.yml` + workflow 구성 완료 | §18.9 |
| 5 | **Hotfix 브랜치 명명 공식화** (`hotfix/vX.Y.Z-*`) | ✅ 스크립트 `--hotfix` 플래그 + `Makefile release-hotfix` target | §18.5 |

### §18.0.1 운영 방침 (매 PR/Release마다 점검)

**모든 PR 작성자**:
1. 브랜치 명명 §18.2 준수 (11 prefix 중 선택)
2. PR에 type/priority/area 3축 label 부착 (필수)
3. PR template 내 "Merge Strategy" 체크박스 선택 (`--squash` vs `--merge`)
4. Commit 메시지 Conventional Commits 형식 (type:scope) — Release Drafter 자동 분류용

**모든 릴리스 담당자**:
1. Release Drafter draft 확인 → `CHANGELOG.md`에 영문+한국어 섹션 작성
2. `./scripts/release.sh vX.Y.Z` 또는 `make release V=vX.Y.Z` 실행 (수동 tag push 금지)
3. Hotfix는 `./scripts/release.sh vX.Y.Z --hotfix`
4. GoReleaser workflow 완료 후 GitHub Release 5 플랫폼 assets 확인
5. Post-release: docs-site 4개국어 reference sync (별도 PR) — §17 규칙 준수

**모든 리뷰어**:
1. PR의 merge strategy가 브랜치 유형과 일치하는지 확인 (release → `--merge`, feature → `--squash`)
2. 3축 label 부착 여부 확인 (type 축 최소 1, area 축 최소 1, priority 축 최소 1)
3. CI all-green 확인 (Lint / Test ubuntu/macos/windows / Build 5 / CodeQL)
4. v2.14.0 Case Study (§18.11) 재발 방지 체크 (release squash 금지, stacked PR base 전환)

**금지 사항** (§18.10 전체 위반 금지):
- ❌ `main` 직접 push
- ❌ `develop` 브랜치 생성 (Gitflow 패턴)
- ❌ Release PR squash merge (history 손실)
- ❌ `--rebase` merge 전략
- ❌ 수동 `gh release create` (GoReleaser와 충돌)
- ❌ 브랜치 명명 관례 위반

### §18.1 브랜치 구조

```
main ──●──●──●──●──●──  (protected, 항상 배포 가능, tags 부착)
       ↑  ↑     ↑
       │  │     release/vX.Y.Z (며칠 이내)
       │  │
       │  feat/SPEC-XXX-description (1-3일)
       │  fix/issue-NNN-description (1일)
       │  docs/topic-description (1일)
       │  chore/task-description (1일)
       │  plan/vX.Y-topic (SPEC 초안 모음)
       │
       hotfix/vX.Y.Z-description (main의 tag에서 분기)
```

### §18.2 [HARD] 브랜치 명명 규칙

| 접두사 | 용도 | 수명 | 예시 |
|--------|------|------|------|
| `feat/SPEC-XXX-*` | 기능 개발 (SPEC 기반) | 1-3일 | `feat/SPEC-AUTH-001-oauth2` |
| `fix/issue-NNN-*` | 이슈 기반 버그 수정 | 1일 | `fix/issue-683-race-condition` |
| `fix/*` | 이슈 없는 버그 수정 | 1일 | `fix/ci-lint-errcheck` |
| `docs/*` | 문서 작업 | 1일 | `docs/v2.14-release-notes` |
| `chore/*` | 유지보수 (의존성, cleanup) | 1일 | `chore/bump-go-1.26` |
| `ci/*` | CI/GitHub Actions 변경 | 1일 | `ci/add-windows-runner` |
| `refactor/*` | 리팩토링 (동작 불변) | 1-2일 | `refactor/validator-cleanup` |
| `release/vX.Y.Z` | 릴리스 준비 (CHANGELOG, version) | 1-3일 | `release/v2.14.0` |
| `hotfix/vX.Y.Z-*` | 프로덕션 긴급 수정 | 1일 이내 | `hotfix/v2.14.1-import-crash` |
| `plan/vX.Y-*` | SPEC draft 모음 | 1-3일 | `plan/v2.15-util-perf-backlog` |
| `audit/*` | 보안/품질 감사 | 1-3일 | `audit/askuserquestion-propagation` |

### §18.3 [HARD] Merge Strategy

각 브랜치 유형별 머지 방식 **반드시** 준수:

| 머지 유형 | 전략 | gh 명령어 | 이유 |
|-----------|------|-----------|------|
| feature/fix/chore/docs → main | **squash** | `gh pr merge N --squash` | WIP commit 정리, 1 PR = 1 main commit |
| release/* → main | **merge commit** ⭐ | `gh pr merge N --merge` | 릴리스 마일스톤 + 개별 SPEC commit 보존 |
| hotfix/* → main | **merge commit** | `gh pr merge N --merge` | 긴급 변경 이력 명확성 |
| plan/* → main | **squash** | `gh pr merge N --squash` | 초안 commit 정리 |
| dependabot/* → main | **squash** | (auto-merge) | 단순 버전 bump |

**공통 규칙**:
- [HARD] `--delete-branch=true` 기본 (머지 후 head branch 자동 삭제)
- [HARD] `--rebase`는 금지 (main history linear가 아니어도 OK, merge commit 보존 중요)
- [HARD] Force push to `main` 금지 (branch protection으로 차단)

### §18.4 Release Cadence 공식화

| 타입 | 기준 | 주기 | 브랜치 |
|------|------|------|--------|
| **Patch (vX.Y.Z)** | 버그 수정만 | 필요 시 즉시 | `fix/*` 여러 개 직접 main → tag bump |
| **Minor (vX.Y.0)** | SPEC cluster (2-4 SPECs) 또는 주요 feature | 1-2주 | `release/vX.Y.0` 경유 |
| **Major (vX.0.0)** | Breaking changes | 3-6개월 | `release/vX.0.0` + migration guide |
| **Hotfix (vX.Y.Z+1)** | 프로덕션 긴급 수정 | 24h 이내 | `hotfix/vX.Y.Z+1-*` → main |

### §18.5 Hotfix Workflow

프로덕션에서 발견된 긴급 이슈 처리:

```bash
# 1. 최신 production tag에서 분기
git fetch origin --tags
git checkout -b hotfix/v2.14.1-crash-on-startup v2.14.0

# 2. 수정 + 테스트 (로컬)
# ...

# 3. Commit + push
git commit -m "fix(hotfix): crash on startup when config missing (#NNN)"
git push -u origin hotfix/v2.14.1-crash-on-startup

# 4. PR 생성 (base: main)
gh pr create --base main --title "hotfix(v2.14.1): ..." --body "..."

# 5. CI green 확인 후 merge commit
gh pr merge <PR> --merge --delete-branch

# 6. Tag 생성 + push (로컬 릴리스 스크립트 사용)
./scripts/release.sh v2.14.1 "Hotfix release"
```

### §18.6 Label 3축 체계

모든 Issue/PR은 다음 3축 label 부착:

**축 1: type (필수)**
- `type:feature` — 새 기능
- `type:fix` — 버그 수정
- `type:docs` — 문서만 변경
- `type:chore` — 유지보수 (의존성, 빌드 등)
- `type:ci` — CI/CD 변경
- `type:refactor` — 리팩토링
- `type:security` — 보안 이슈/수정
- `type:test` — 테스트 추가/개선

**축 2: priority (필수)**
- `priority:P0` — Critical (즉시 대응)
- `priority:P1` — High (당일)
- `priority:P2` — Medium (이번 sprint)
- `priority:P3` — Low (backlog)
- `priority:P4` — Icebox (장기 보류)

**축 3: area (필수)**
- `area:mx` — MX validator
- `area:astgrep` — ast-grep integration
- `area:lsp` — LSP subsystem
- `area:hooks` — Hook system
- `area:docs-site` — docs-site (Hugo)
- `area:templates` — internal/template/templates/
- `area:cli` — CLI 명령어
- `area:config` — .moai/config/*
- `area:workflow` — SPEC workflow
- `area:ci` — GitHub Actions
- `area:security` — 보안 관련
- `area:deps` — 외부 의존성

**축 4: status (선택)** — 진행 상태 추적
- `status:in-progress` · `status:review` · `status:blocked` · `status:needs-info`

### §18.7 Branch Protection Rule (GitHub)

[HARD] `main` 브랜치 보호 설정 (admin 실행 필요):

```bash
gh api -X PUT /repos/modu-ai/moai-adk/branches/main/protection \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Test (ubuntu-latest)", "Test (macos-latest)", "Test (windows-latest)", "Build (linux/amd64)", "CodeQL"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_linear_history": false,
  "required_conversation_resolution": true
}
EOF
```

[HARD] `release/*` 패턴 보호 (feature freeze 기간 안정성):

```bash
gh api -X PUT /repos/modu-ai/moai-adk/branches/release%2F*/protection \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Test (ubuntu-latest)", "Test (macos-latest)"]
  },
  "allow_force_pushes": false,
  "allow_deletions": false
}
EOF
```

### §18.8 Release 프로세스

[HARD] 릴리스는 `scripts/release.sh` 스크립트 경유. 수동 tag push 금지 (GoReleaser 자동 릴리스와 충돌 가능).

**Minor/Major Release (release 브랜치 경유)**:

```bash
# 1. release 브랜치 생성 + 작업
git checkout -b release/v2.15.0 main
# CHANGELOG.md, SPEC status 업데이트, docs-site update 등

# 2. PR + review + merge (merge commit 전략)
gh pr create --base main --title "release: v2.15.0" --body "..."
gh pr merge <PR> --merge --delete-branch

# 3. Tag + release (로컬 스크립트)
git checkout main && git pull
./scripts/release.sh v2.15.0
# → tag 생성, push, GoReleaser 자동 실행, GitHub Release 생성
```

**Patch Release (fix 직접 main)**:

```bash
# 1. fix 브랜치에서 수정 + PR + squash merge
# 2. main pull + tag + push (release 스크립트)
./scripts/release.sh v2.14.1
```

### §18.9 자동화 도구

**구성 완료된 도구 (Enhanced GitHub Flow 공식 인프라)**:

| 도구 | 역할 | 트리거 | 설정 파일 |
|------|------|--------|-----------|
| **GoReleaser** | Tag push 시 5 플랫폼 바이너리 빌드 + GitHub Release 생성 | `push: tags: v*` | `.goreleaser.yml` + `.github/workflows/release.yml` |
| **Release Drafter** ⭐ | PR merge 시 next release draft 자동 업데이트 (CHANGELOG 작성 보조) | `push: branches: main`, `pull_request: [opened, synchronize, edited]` | `.github/release-drafter.yml` + `.github/workflows/release-drafter.yml` |
| **Dependabot** | Go modules + GitHub Actions 자동 버전 업데이트 PR | 주간 | `.github/dependabot.yml` |
| **Auto-merge** | Dependabot PR CI pass 시 자동 머지 (squash) | PR + CI success | `.github/workflows/auto-merge.yml` |
| **Labeler** | PR 파일 패턴 기반 자동 라벨 부착 (area 축 추론) | PR opened/synchronized | `.github/labeler.yml` |
| **Release Drafter autolabeler** | PR title/branch/files 기반 type 축 자동 라벨 | 위 Release Drafter와 통합 | `.github/release-drafter.yml` § `autolabeler` |

**Release Drafter ↔ GoReleaser 역할 분담** (공존 설계):
- **Release Drafter** = "다음 릴리스 preview" 자동 축적 → 인간이 `CHANGELOG.md`에 영문+한국어로 정제
- **GoReleaser** = tag push 시 final release 생성 (commit 기반 changelog 포함). Release Drafter draft와 무관하게 tag 기준 독립 운영
- **실제 workflow**: PR merge → Release Drafter가 draft 축적 → 릴리스 시 draft 확인 → `CHANGELOG.md` 업데이트 → `./scripts/release.sh` → GoReleaser final release

**Release Drafter Version Resolver** (자동 SemVer 추정):
- `breaking` / `type:breaking` label → major bump
- `type:feature` label → minor bump
- `type:fix` / `type:security` / `type:performance` / 기타 → patch bump

**추가 도구 (v2.15+ 검토)**:
- **EndBug/label-sync**: `.github/labels.yml`을 GitHub과 주기적 동기화 (현재 수동)
- **Commitlint GitHub Action**: Conventional Commits 형식 CI 강제
- **Changesets (대안)**: CHANGELOG 관리 자동화 심화 — 현재 Release Drafter로 충분

### §18.10 [HARD] 공식 위반 금지 사항

- ❌ `main`에 직접 push (반드시 PR 경유)
- ❌ `develop` 브랜치 생성 (Gitflow 패턴 금지)
- ❌ release PR squash merge (merge commit 필수 — 개별 SPEC commit 보존)
- ❌ tag 수동 push 없이 `gh release create` (GoReleaser와 충돌)
- ❌ branch prefix 관례 위반 (`feat/SPEC-*` 대신 `add-something` 등)
- ❌ `--rebase` merge 전략 (history rewrite 방지)
- ❌ CI fail 상태 PR merge (admin override 필요, 최소화)

### §18.11 이전 세션 학습 (v2.14.0 Case Study)

v2.14.0 릴리스 과정에서 다음 문제 발생 → v2.15부터 방지:

1. **Squash merge로 release history 손실**
   - PR #703 (`release/v2.14.0 → main`)을 squash로 머지하여 9개 의미 있는 commit (UTIL-001/002/003 + review fix + CI fix + ETXTBSY)이 단일 commit으로 압축됨
   - **교훈**: release → main은 항상 `--merge` 사용

2. **Stacked PR close 후 reopen 실패**
   - PR #704 (base: release/v2.14.0)가 PR #703 merge 과정에서 auto-close됨. `gh pr reopen` 실패 → 새 PR #705 생성으로 우회
   - **교훈**: stacked PR은 parent merge 전 base를 main으로 미리 전환하는 것이 안전

3. **CI Ubuntu 환경 의존성 누락**
   - `sg` command가 Ubuntu의 `util-linux` `newgrp` symlink와 충돌
   - `TestRuleSeed` skip 로직이 `LookPath`만으로 판정하여 false positive
   - **교훈**: 환경 의존 테스트는 **command signature 검증** 필수 (예: `sg --version` 출력의 "ast-grep" 확인)

4. **Pre-existing flaky test가 릴리스 block**
   - `TestSupervisor_NonZeroExit`의 ETXTBSY race가 Ubuntu CI에서 간헐 실패
   - **교훈**: flaky 테스트는 CLAUDE.local.md에 기록 + retry 로직 즉시 적용 (ETXTBSY mitigation pattern 참조)

---

---

## 19. AskUserQuestion Enforcement Protocol

2026-04-24 세션에서 `AskUserQuestion` 미사용으로 [HARD] 규칙(§1, §8) 위반 반복 발생. 근본 원인은 **deferred tool 사전 로드 부재** + **산문 질문 편의주의**. v3.4.0부터 본 Enforcement Protocol을 [HARD] 운영 방침으로 고정.

### §19.1 근본 원인 체인

1. **1차 원인**: `AskUserQuestion`은 deferred tool. 세션 시작 시 schema 미로드 → 직접 호출 시 `InputValidationError` → agent가 회피 → 산문 질문으로 우회.
2. **2차 원인**: Red Flags / Verification 체크리스트에 "응답 말미 `?` + AskUserQuestion 미호출" 탐지 규칙 부재.
3. **3차 원인**: 규칙은 존재하나 매 세션 agent 해석에 의존. 편의주의("짧은 질문은 산문으로") 허용 관행.

### §19.2 [HARD] 세션 시작 Preload (의무)

모든 MoAI 세션은 첫 사용자 입력 수신 직후 다음 `ToolSearch` 호출을 실행해야 한다:

```
ToolSearch({
  query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet",
  max_results: 5
})
```

Preload 완료 후에만 해당 tool 호출 가능. Preload 이전 호출 = HARD 위반 + `InputValidationError`.

### §19.3 [HARD] Pre-Response Self-Check (응답 전송 전 필수)

모든 사용자 응답 전송 전, 다음 4항목 자가 점검:

1. **Question mark detection**: 응답에 `?`가 포함되었는가? → 있으면 `AskUserQuestion` tool call 동반 필수
2. **Option list detection**: 응답에 선택지 구조(`- A:`, `- B:`, `Option X:`, `1.`)가 있는가? → structured `AskUserQuestion` 경유 필수
3. **Schema load check**: `AskUserQuestion` schema 로드 상태 확인. 미로드 시 `ToolSearch` 선행
4. **Silent-wait detection**: 산문 질문 후 사용자 응답 대기 상태인가? → `AskUserQuestion`으로 전환 필수

점검 실패 시 = [HARD] 규칙 위반. 즉시 수정 필요.

### §19.4 [HARD] Anti-Patterns (금지)

| 패턴 | 왜 금지 | 올바른 방법 |
|------|---------|-------------|
| 응답 말미 `?`로 끝나는 산문 질문 | 사용자에게 무엇을 원하는지 불명확, silent wait | `AskUserQuestion` 호출 필수 |
| "진행할까요?", "A or B?" 자연어 선택 요청 | structured 응답 아닌 free-form 기대 → 파싱 불가 | `AskUserQuestion` with 2-4 options |
| 선택지를 markdown `- A:`, `- B:`로만 나열 | 사용자가 자연어로 답해야 함 | `AskUserQuestion` structured options |
| `AskUserQuestion` 호출 전 `ToolSearch` 생략 | `InputValidationError` 발생 | §19.2 Preload 먼저 |
| Silent "wait for next message" after prose | 사용자에게 정확한 응답 형식 미제공 | `AskUserQuestion` 또는 구체적 지시 |

### §19.5 운영 적용 (Role별)

**MoAI orchestrator**:
- [HARD] 모든 사용자 결정 요청은 `AskUserQuestion` 경유
- [HARD] 세션 시작 시 §19.2 preload 실행
- [HARD] 응답 전송 전 §19.3 self-check 통과

**Subagent (manager-*, expert-*, builder-*)**:
- [HARD] `AskUserQuestion` 호출 금지 (agent-common-protocol §User Interaction Boundary)
- 사용자 결정 필요 시 "missing inputs" 보고서로 orchestrator에 반환

**User**:
- 산문 질문 탐지 시 즉시 지적 (2026-04-24 세션 사례처럼)
- 반복 위반 시 memory/feedback_askuserquestion_*.md에 기록

### §19.6 Recovery Protocol (위반 발생 시)

[HARD] 위반 탐지 즉시 다음 순서로 복구:

1. 위반 인정 + 근본 원인 명시 (1/2/3차)
2. `ToolSearch`로 schema preload
3. 동일 질문을 `AskUserQuestion`으로 재구성 + 재전송
4. `memory/feedback_askuserquestion_*.md`에 incident 기록
5. 다음 응답부터 §19.3 self-check 엄격 적용

### §19.7 관련 정책 참조

- CLAUDE.md §1 HARD Rules (AskUserQuestion-Only + Deferred Tool Preload)
- CLAUDE.md §8 User Interaction Architecture § Deferred Tool Preload Protocol
- .claude/skills/moai/SKILL.md § Red Flags + Verification (deferred tool 관련 5+4 항목)
- ~/.claude/projects/{hash}/memory/feedback_askuserquestion_enforcement.md (이번 incident 기록)

---

**Status**: Active (Local Development)
**Version**: 3.4.0 (Phase 9: §19 AskUserQuestion Enforcement Protocol 공식 채택)
**Last Updated**: 2026-04-24
