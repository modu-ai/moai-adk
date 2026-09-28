# moai-adk-go Local Development Guide — Detail Companion

> Dev-only companion of `AGENTS.local.md` (no template mirror — this file is load-bearing for the
> maintainer only). Owns the reference bodies relocated from `AGENTS.local.md` by the always-loaded
> diet (card t1303): quick start, the local-only file registry, the embedded-template and
> command-publication references, the commit format, coverage targets, template variables,
> configuration, frequent-issue details, and the YAML frontmatter checklist. `AGENTS.local.md`
> keeps every [HARD] rule and a pointer here.

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
4. Git commit from the card worktree — never from the primary checkout (AGENTS.md §2·§3)
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

## 2a. Local-Only Files (Never in Templates)
```
.claude/settings.local.json    # Personal settings — runtime-managed, NEVER template
.claude/settings.json          # Rendered from .json.tmpl
.claude/agent-memory/          # Per-project agent memory
.claude/hooks/moai/handle-*.sh # Hook wrappers deployed from templates (.sh/.sh.tmpl pairs in internal/template/templates/.claude/hooks/moai/ — edit both sides together, §2.3)
.claude/rules/local/lifecycle-sync-gate.md                 # Dev-only: maintainer lifecycle sync-gate rule (no template mirror, unreferenced by any shipped template file — intentional local-only)
.claude/rules/local/repo-local-pr-policy.md               # Dev-only: repo-local all-tier PR policy override (Route A main-direct disabled by branch protection enforce_admins:true; no template mirror — intentional local-only)
.claude/rules/local/gitflow-lane-protocol.md              # Dev-only: git-flow lane operational rule (2026-08-27 transition; no template mirror — added [2026-08-27 감사])
.claude/commands/harness/{release-update,github,release}*  # Dev-only: split maintainer harness entries (§21)
.claude/commands/harness/release-update/manifest.json      # Dev-only: release-update harness manifest (§21)
.claude/workflows/hns-release-update-run.js                # Dev-only: release-update harness Runner (§21)
.claude/agents/harness/hns-{release-update,github,release}-specialist.md  # Dev-only: split harness specialists (§21, user-owned per §24)
scripts/ci-watch/              # Dev-only: CI watch loop scripts (5) — not distributed
scripts/ci-autofix/            # Dev-only: CI auto-fix scripts (4) — not distributed
scripts/jev/                   # Dev-only: TypeSafe(Jev) 로컬 전용 도구 (§29) — 템플릿 미러 없음, 사용자 프로젝트로 배포되지 않음
scripts/ac-baseline/           # Dev-only: 커밋타임 AC-snapshot 가드(check-staged.sh·install-hook.sh) — git config 기반 pre-commit 훅으로 develop 병합 후 리드가 1회 설치, 템플릿 미러 없음, 사용자 프로젝트로 배포되지 않음 (SPEC-ACSNAPSHOT-COMMIT-GUARD-001)
~/.moai/.env.typesafe          # Dev-only: TypeSafe API 키 (저장소 밖, chmod 600). settings/config/템플릿에 넣지 않는다 (§29)
.claude/skills/hns-workflow-ci-loop/                       # Dev-only: CI watch+autofix skill (removed from template; mirror kept). §2.3에 따라 moai-workflow-ci-loop → hns-* 로 이동(2026-08-15): `.claude/skills/moai*` 글롭이 매 update마다 삭제했음
.claude/rules/local/ci-watch-protocol.md                     # Dev-only: governs scripts/ci-watch (removed from template; mirror kept)
.claude/rules/local/ci-autofix-protocol.md                 # Dev-only 원본: scripts/ci-autofix 를 지배. 배포판(.claude/rules/moai/workflow/ 의 같은 이름, script-free)과 **의도적 쌍둥이** — SPEC-CI-LOOP-DEVONLY-001 의 결정이며 미해결 상태가 아니다. 둘은 `paths:` 범위가 서로 겹치지 않아 함께 로드되지 않는다(로컬판=데브 스킬 SKILL.md, 배포판=manager-develop + .github/workflows/**). #1557(ed04e40e6)이 이 파일을 관리 대상 뿌리 밖으로 옮겨 §2.3 경로 충돌도 해소됐다. 다만 배포판이 update 때마다 `.claude/rules/moai/workflow/` 에 미추적으로 재생성돼 git status 노이즈로 남는다
AGENTS.local.md                # This file's parent guide
.moai/state/last-cc-version.json # Dev-only: CC tracking state (§21)
.moai/research/cc-update-*.md  # Dev-only: CC update reports (§21)
.moai/cache/                   # Cache
.moai/logs/                    # Logs — 내용물만 로컬 전용. 빈 디렉터리 스캐폴드(.gitkeep)는 템플릿에 있다 [2026-08-27 감사 정정]
.moai/state/                   # Session state storage — 위와 같음(템플릿에 .gitkeep + state/chain/ 스캐폴드 존재)
.moai/specs/                   # Active SPEC documents
.moai/plans/                   # Session plans
.moai/reports/                 # Generated reports — 내용물만 로컬 전용(템플릿에 reports/plan-audit/ 스캐폴드 존재) [2026-08-27 감사 정정]
.moai/manifest.json            # Generated at runtime
.moai/status_line.sh           # Rendered from .sh.tmpl
.moai/docs/update-local-file-survival.md  # Dev-only: moai update 관리 대상 삭제 실측·생존 규칙 (AGENTS.local.md §2.3 본문 이관, card t750; no template mirror)
.moai/docs/gitflow-integration-chain.md   # Dev-only: GitFlow 통합 체인 운영 절차·실측 근거 (AGENTS.local.md §4.1 절차 본문 이관, card t750; no template mirror)
.moai/docs/local-dev-guide.md             # Dev-only: this file — AGENTS.local.md 참조 본문 이관 (card t1303; no template mirror)
.moai/astgrep-rules/                                                       # Dogfood-only: 실험적 ast-grep 룰셋 전체. §2.3에 따라 .moai/config/astgrep-rules → .moai/astgrep-rules 로 이동(2026-08-15): `.moai/config` 통째 삭제가 로컬 전용 6개(go/{concurrency,error-handling,idioms,resource-safety}.yml, security/{secrets,web}.yml)를 매번 지웠음. gate.yaml `ast_grep_gate.rules_dir` 로 연결. 주의: (2026-08-27 감사 정정) `moai ast-grep`/`moai ast-edit` CLI의 `--rules-dir` 기본값은 현재 `gate.yaml`의 `ast_grep_gate.rules_dir`이며, 빈 값이면 기본 경로 폴백 없이 0룰 스캔을 한다(t50)
```

## 2b. Embedded Template System

moai-adk-go uses Go's `go:embed` directive:
- **Source**: `internal/template/templates/` (edit here — this is the source of truth)
- **Embed mechanism**: `internal/template/embed.go` carries `//go:embed all:templates` + `//go:embed catalog.yaml`, which compile the `templates/` FS directly into the binary (there is NO generated `embedded.go` file)
- **Build**: Run `make build` after editing templates (recompiles the binary)

## 2c. Command-to-Skill Publication (SPEC-CODEX-COMMAND-SKILLS-001)

`internal/template/commandemit` publishes the 16 `/moai` command sources as codex skill-shaped artifacts at `templates/.agents/skills/moai-<command>/SKILL.md` (committed real files, golden-checked).

- **Regenerate** (after editing any command source or the emitter): `make commands-emit`
- **Drift check**: `make commands-emit-check` — read-only, wired ahead of `build` (same position as `agents-emit-check`); it never writes
- **Boundary**: bodies publish VERBATIM from the command sources, including their Claude-only `Skill("moai")` dispatcher line — the emitter flags this per skill and never repairs it (repair is the command-body layer's concern, sibling card t497). Do not hand-edit the emitted SKILL.md files; edit the command sources and regenerate.
- **gitignore coupling**: the 16 published names are re-included in `templates/.gitignore` (the mirror rule `.agents/skills/moai*` would otherwise ignore them); `TestGitignoreCarriesEveryPublishedName` + `TestPublishedSkillsNamesMatchTree` keep both lists in step with the emitted set.

## 4a. Commit Message Format
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

## 6a. Test Isolation — why, and Coverage Targets

**Why this matters - `filepath.Join` vs absolute paths:**

On macOS, `t.TempDir()` returns paths starting with `/var/folders/...`.
Go's `filepath.Join(cwd, absPath)` does NOT strip the leading `/` from the second arg:
```
filepath.Join("/a/b", "/var/folders/x") = "/a/b/var/folders/x"  // WRONG!
filepath.Abs("/var/folders/x") = "/var/folders/x"                // CORRECT
```

Always use `filepath.Abs()` when resolving user-supplied paths in CLI commands.
Never use `filepath.Join(cwd, userPath)` when `userPath` can be absolute.

```go
func TestSomething(t *testing.T) {
    tempDir := t.TempDir()  # Auto-cleanup after test - ALWAYS use this
    // Work in tempDir instead of project root
}
```

### Coverage Targets

- Package-level: 85% minimum coverage (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`)
- Critical packages (cli, template, hook): 90%+ coverage — 목표치로서 유효하나 기계적 근거는 strict 평가 프로필의 전역 "Coverage >= 90%" 게이트뿐이다(`.moai/config/evaluator-profiles/strict.md`; 패키지별 룰은 미발견 [2026-08-27 감사 정정])

## 8. Template Variable Strategy

### Template vs Local Settings

moai-adk-go uses different path variable strategies:

**Template settings** (`internal/template/templates/.claude/settings.json.tmpl`; [2026-08-27 감사 정정]):
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
type TemplateContext struct { // (요약 발췌 — 전체 구조체는 internal/template/context.go)
    GoBinPath string  // Path to Go bin directory
    HomeDir   string  // User home directory
}
```

**Usage in templates:**
```bash
# .moai/status_line.sh.tmpl — 바이너리 폴백 체인(command -v → ResolvedMoaiPath → $HOME/go/bin)
if [ -f "$HOME/go/bin/moai" ]; then
	exec "$HOME/go/bin/moai" statusline
fi
```

**Rendering:**
```go
ctx := template.NewTemplateContext(
    template.WithGoBinPath(detectGoBinPath()),
    template.WithHomeDir(homeDir),
)
// Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext)
// — 첫 인자는 context.Context, 마지막 인자가 TemplateContext다 [2026-08-27 감사 정정]
deployer.Deploy(context.Background(), projectRoot, mgr, tmplCtx)
```

## 9. Configuration System

### Config File Format

moai-adk-go uses YAML for configuration:

**Project config**: 단일 `config.yaml`은 없다 — 설정은 `.moai/config/sections/*.yaml` 32개 파일로만 존재한다([2026-08-27 감사 정정]; 종전 서술이 가리키던 `.moai/config/config.yaml`은 이 트리에 부재).

**Section files** (`.moai/config/sections/*.yaml`, 32개):
- `quality.yaml` - Quality gates, development mode
- `language.yaml` - Language preferences
- `user.yaml` - User information
- `workflow.yaml` - Workflow settings

### Configuration Priority

1. Environment Variables (override file values) — the actually-implemented
   config overrides: `internal/config/manager.go` `applyEnvOverrides` reads
   `MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR`
   (manager.go:398-411), while `MOAI_CONFIG_DIR` (config directory location) is
   honored separately by the same file's config-dir resolver (manager.go:70;
   [2026-08-27 감사 정정]). All env-var names are
   constants in `internal/config/envkeys.go`.
2. User Configuration: `.moai/config/sections/*.yaml`
3. Template Defaults: From `internal/template/templates/.moai/config/`

> NOTE: `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` are NOT currently
> implemented — no code in `internal/`/`pkg/`/`cmd/` reads them. User name and
> conversation language come from `.moai/config/sections/user.yaml` /
> `language.yaml` only. (Adding these env overrides would be a future
> enhancement, not current behavior.)

## 11a. Frequent Issues — details

자주 발생 4건:

- **Templates not updated after editing** → §2b Embedded Template System (`make build` 필수, `//go:embed all:templates`).
- **Tests modify ~/.claude/settings.json** → `t.TempDir()` 격리 위반. 테스트가 project root에 파일 만드는지 확인.
- **Hook timeout** → settings.json `{"timeout": 60}` (기본 5초).
- **`moai version` exit 137 (SIGKILL) after binary reinstall** → `cp bin/moai ~/go/bin/moai`만으로 부족; 기존 binary 잔재(go install buildinfo·mmap 캐시)가 꼬여 SHA가 같아도 crash. **반드시 `rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai`**(또는 `make install`)로 inode 갱신하며 clean 재설치. 맨손 `go install ./cmd/moai`는 금지 — `LDFLAGS`를 안 실어서 `pkg/version`의 컴파일 기본값(`Commit="none"`, `Date="unknown"`)이 박히고 binary provenance 검증이 무력화된다. clean 재설치 직후 `sh scripts/verify-local-install.sh`를 실행한다(`make verify-local-install`은 같은 script의 편의 alias). 이 script는 `bin/moai`와 설치본을 byte 단위로 비교하고 설치본의 `version`에서 측정 시점 HEAD의 short SHA와 exit 0을 확인한다. 기본 `git`이 Xcode 라이선스에 막히면 Command Line Tools `git`을 자동 재시도하며 macOS `strings`는 호출하지 않는다. `/usr/bin/make`까지 exit 69인 호스트는 `.claude/rules/local/gitflow-lane-protocol.md` §9의 PATH 전처리를 먼저 적용한다. 진단 징후: `bin/moai version`=0인데 설치본 `version`=137. binary lag 검증은 AGENTS.local.md §6 검증 규율(clear → 측정).

## 12. YAML Frontmatter 빠른 참조

범용 형식 규칙은 `.claude/rules/moai/development/` 내 `skill-authoring.md`, `agent-authoring.md`에 정의.

### 로컬 개발 체크리스트

- [ ] `tools:`, `allowed-tools:` → CSV string (공백 구분 절대 금지)
- [ ] `skills:` → YAML array (유일한 예외)
- [ ] `metadata.*` → quoted string
- [ ] Template 수정 후 `make build` 실행
- [ ] Local copy (`.claude/`)도 동기화

탐지 스크립트: 전역 프로젝트 메모리(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/` (**휴면 저장소** — 세션이 실제로 로드하는 것은 `CLAUDE_CONFIG_DIR` profile 저장소이고 이 파일은 거기 없다. 읽으려면 이 경로를 직접 연다. 두 저장소 확인은 `moai memory doctor`))의 `audit_sweep_patterns.md` Pattern A 참조.
