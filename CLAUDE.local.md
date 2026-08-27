# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-08-27

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
# (.agency/ 템플릿 뿌리는 제거됨 — legacy archive; [2026-08-27 감사 정정])
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

### [HARD] §2.0 에이전트 정의는 사본이 셋이고, 셋째는 손으로 고치지 않는다

에이전트 정의는 세 벌 있는데 **하나만 손편집 대상이 아니다.**

| 사본 | 경로 | 성격 |
|---|---|---|
| C1 | `.claude/agents/moai/*.md` | 로컬 도그푸드, 손편집 |
| C2 | `internal/template/templates/.claude/agents/moai/*.md` | 배포 미러, 손편집 — **중립 원본 층** |
| C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **C2 로부터 기계 방출** (`internal/template/agentemit`) |

[HARD] **`internal/template/templates/.claude/agents/moai/*.md` 를 고쳤으면 `make agents-emit` 을 돌린다.** 그러지 않으면 `internal/template/templates/.codex/agents/moai/*.toml` 이 스테일한 채로 남고, 그 상태로 만든 바이너리가 옛 정의를 임베드한 채 돌아간다. C1↔C2 는 바이트 동일 관계가 **아니다**(의도된 분기) — 생성 관계는 C2 → C3 한 방향뿐이다.

[HARD] **C3 를 손으로 고치지 않는다.** 손편집은 다음 방출에서 말없이 덮인다. 고칠 것이 있으면 C2 를 고치고 재생성한다.

빠뜨렸을 때 어디서 잡히는가 — 세 지점이 각각 다른 축을 본다:

| 검사 | 무엇을 보는가 | 언제 |
|---|---|---|
| `make build` (선행 `agents-emit-check`) | 소스 층: 커밋된 `.toml` vs `.md` 방출 결과 | 로컬 빌드마다. **읽기전용 — 재생성하지 않는다** |
| `go test ./internal/template/agentemit/...` | 같은 축 | CI 매 실행 |
| `make embed-check` / `moai doctor --check "Agent Emit Embed"` | **임베드 축**: 이미 빌드된 바이너리가 실은 바이트 vs 커밋본 | 수동. `BIN=<path>` 로 설치본도 겨눌 수 있다 |

셋째가 따로 있는 이유: `go test` 는 테스트 바이너리를 매번 새로 컴파일하고 `//go:embed` 가 그 시점 커밋본을 읽으므로, 소스↔소스 비교로는 **노후한 바이너리를 원리상 볼 수 없다**. 그래서 `make embed-check` 는 `build` 를 선행으로 갖지 않는다 — 갓 빌드한 바이너리는 정의상 일치하므로 build 직후에만 도는 검사는 동어반복이다. 같은 이유로 CI 빌드 잡에도 붙이지 않는다(CI 는 자신이 검사하는 커밋에서 빌드한다).

재생성은 **`make agents-emit` 이라는 명시적 동사로만** 일어난다. build 가 조용히 덮으면 손편집이 있었다는 사실 자체가 사라진다.

**§2.1 Template Content Neutrality — Acceptable Content Range for Templates**: When editing template source files in `internal/template/templates/`, ensure content adheres to the **acceptable** kept-classes (C1/C2/C4/C5/C6/C8) and excludes the FORBIDDEN content classes (SPEC IDs, REQ tokens, Audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references), enforced by CI guard (`.github/workflows/template-neutrality-check.yaml` trigger on path change). The canonical C1-C8 acceptable-vs-forbidden content-class catalogue lives in `.moai/docs/template-internal-isolation-doctrine.md §25.1` (cross-referenced by **§25 (Template Internal-Content Isolation)** of this file, now a stub). This keeps the template neutral across the 16 supported **programming languages** (§15) and free of moai-adk internal development state. Distinct axis: user-facing locales (ko/en/ja/zh) are governed by the §8 Localization Contract in the active output style — see §15's disambiguation note.

**Pre-PR Verification (template contributor-checklist)** — before opening a PR that touches `internal/template/templates/**`, run the canonical 5-item pre-commit self-check (the CI guard `template-neutrality-check.yaml` is the safety net). See `.moai/docs/template-internal-isolation-doctrine.md` §25.3 for the full 5-item checklist and §25.1 for the forbidden/allowed content-class catalogue (C1-C8). (C3 dates + C7 commit-hashes are owned by the sibling `internal_content_leak_test.go` per §25, not this neutrality checklist.)

### Local-Only Files (Never in Templates)
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
.claude/skills/hns-workflow-ci-loop/                       # Dev-only: CI watch+autofix skill (removed from template; mirror kept). §2.3에 따라 moai-workflow-ci-loop → hns-* 로 이동(2026-08-15): `.claude/skills/moai*` 글롭이 매 update마다 삭제했음
.claude/rules/local/ci-watch-protocol.md                     # Dev-only: governs scripts/ci-watch (removed from template; mirror kept)
.claude/rules/local/ci-autofix-protocol.md                 # Dev-only 원본: scripts/ci-autofix 를 지배. 배포판(.claude/rules/moai/workflow/ 의 같은 이름, script-free)과 **의도적 쌍둥이** — SPEC-CI-LOOP-DEVONLY-001 의 결정이며 미해결 상태가 아니다. 둘은 `paths:` 범위가 서로 겹치지 않아 함께 로드되지 않는다(로컬판=데브 스킬 SKILL.md, 배포판=manager-develop + .github/workflows/**). #1557(ed04e40e6)이 이 파일을 관리 대상 뿌리 밖으로 옮겨 §2.3 경로 충돌도 해소됐다. 다만 배포판이 update 때마다 `.claude/rules/moai/workflow/` 에 미추적으로 재생성돼 git status 노이즈로 남는다
CLAUDE.local.md                # This file
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
.moai/astgrep-rules/                                                       # Dogfood-only: 실험적 ast-grep 룰셋 전체. §2.3에 따라 .moai/config/astgrep-rules → .moai/astgrep-rules 로 이동(2026-08-15): `.moai/config` 통째 삭제가 로컬 전용 6개(go/{concurrency,error-handling,idioms,resource-safety}.yml, security/{secrets,web}.yml)를 매번 지웠음. gate.yaml `ast_grep_gate.rules_dir` 로 연결. 주의: (2026-08-27 감사 정정) `moai ast-grep`/`moai ast-edit` CLI의 `--rules-dir` 기본값은 현재 `gate.yaml`의 `ast_grep_gate.rules_dir`이며, 빈 값이면 기본 경로 폴백 없이 0룰 스캔을 한다(t50)
```

### [HARD] §2.3 moai update는 관리 대상 뿌리 안의 로컬 전용 파일을 통째로 삭제한다

**위 Local-Only 목록에 적혀 있다는 사실만으로는 파일이 보호되지 않는다.** 2026-08-15 `moai update --yes` 실행에서 이 목록에 명시된 파일 12개가 실제로 삭제됐다 — 목록은 사람과 AI가 읽는 문서일 뿐, 삭제를 수행하는 Go 코드는 이 파일을 읽지 않는다.

**삭제 주체**: `CleanMoaiManagedPaths` (`internal/cli/update/deploy/deploy.go:107`; [2026-08-27 감사 정정]). diff 기반이 아니다. 템플릿 재배포 **전에** 아래 뿌리를 `os.RemoveAll`로 **통째 삭제**한 뒤 임베드 템플릿에 있는 것만 다시 깐다. 템플릿에 대응 파일이 없으면 복구되지 않는다.

```
.claude/settings.json      .claude/commands/moai     .claude/agents/moai
.claude/skills/moai*(글롭) .claude/rules/moai         .claude/output-styles/moai
.claude/hooks/moai         .moai/config              (deploy.go:187-190; [2026-08-27 감사 정정])
```

`.moai-skip-cleanup` 마커는 이 함수가 **참조하지 않는다**(참조 0회 — v2→v3 clean-reinstall 경로 전용). 사용자가 편집할 수 있는 보호 목록 설정은 **존재하지 않는다**.

**[HARD] 규율 — 로컬 전용 파일을 위 뿌리 안에 두지 않는다.** 새 로컬 전용 파일을 만들 때 위치부터 정한다:

| 용도 | 금지 (삭제됨) | 안전 |
|---|---|---|
| 룰 | `.claude/rules/moai/**` | `.claude/rules/` 하위의 **비-moai** 디렉터리 |
| 스킬 | `.claude/skills/moai*` | `moai`로 **시작하지 않는** 이름 (글롭 회피) |
| ast-grep 룰셋 | `.moai/config/astgrep-rules/` | `.moai/` 하위 **`config/` 밖** + `gate.yaml`의 `ast_grep_gate.rules_dir` 지정 (빈 값이면 기본 경로 폴백 없음 — t50, `internal/cli/astgrep.go:69,105`; [2026-08-27 감사 정정]) |
| 하네스 | — | `.claude/skills/hns-*`, `.claude/agents/harness/`, `.claude/commands/harness/`, `.moai/harness/` (`IsUserOwnedNamespace` 백업 대상) |

**[HARD] update 실행 후 매번 검증한다.** 전제: 실행 **전** 추적 파일 수정이 0이어야 diff 귀속이 가능하다.

```bash
git status --porcelain | grep -v '^??' | wc -l        # 실제 변경 수
git status --porcelain | grep '^ D'                   # 삭제된 파일 — 0이어야 정상
# 삭제가 있으면 (전부 추적 파일이므로 git이 안전망):
git status --porcelain | grep '^ D' | sed 's/^...//' | tr '\n' '\0' | xargs -0 git restore --
```

**[HARD] update 후 `git-strategy.yaml`의 git-flow 키를 반드시 재적용한다.** `.moai/config`는 위 wipe 대상이므로, `moai update` 는 `.moai/config/sections/git-strategy.yaml` 을 템플릿 기본값(`workflow: github-flow`, develop/release 키 없음)으로 되돌린다. 이 파일은 **템플릿에 미러하지 않는다** — 미러하면 16개 언어 배포판 전체에 이 프로젝트의 사설 워크플로가 실려 나간다(§15). 그러니 매 update 후 로컬에서 다시 넣는다:

```bash
# 확인 — git-flow 가 아니면 되돌아간 것
grep -n 'workflow: git-flow' .moai/config/sections/git-strategy.yaml || echo 'REVERTED — 재적용 필요'
# 재적용 (git_strategy.manual 블록)
git restore --source=HEAD -- .moai/config/sections/git-strategy.yaml
```

`git restore` 가 통하지 않는 상황(커밋 전 상태)이면 `git_strategy.manual` 아래를 손으로 되돌린다: `workflow: git-flow` [2026-08-27 감사 정정], 그리고 `main_branch:` 바로 아래에 `develop_branch: develop` / `release_branch_prefix: release/` / `rc_version_format: vX.Y.Z-rc.N` 세 줄.

**[HARD] 보고된 파일 수를 믿지 않는다.** `Updated N files`의 N은 관리 대상 뿌리 **밖** 파일만 센다(`internal/cli/update/plan/plan.go:73` `if IsMoaiManaged(...) { continue }`). 2026-08-15 실측: 보고 32, 실제 175. **삭제는 이 요약에 전혀 나타나지 않는다.**

**삭제만이 손실이 아니다 — 덮어쓰기 2종** (2026-08-15 실측):

- **경로 충돌**: 템플릿과 **내용이 달라야 하는** 로컬 파일이 템플릿 경로에 있으면, 삭제가 아니라 템플릿판으로 **덮어써진다**. `ci-autofix-protocol.md`(dev 원본 vs 템플릿 script-free판)가 그렇게 유실됐다. 파일이 남아 있어 삭제 검사(`git status | grep '^ D'`)로는 안 잡힌다 — **내용이 달라야 하는 파일은 반드시 관리 대상 밖에 둔다.**
- **`.sh` / `.sh.tmpl` 쌍 드리프트**: 배포되는 것은 `.tmpl` 쪽인데 편집이 `.sh` 쪽에만 들어가면, update가 배포본을 **구버전으로 되돌린다**. 실측: `handle-{agent-hook,task-completed,teammate-idle,stop-goal}.sh` 4쌍에서 `.sh`에만 있던 SPEC-STOPCHAIN-TRIM-001 가드(41줄)가 사라짐. `sync-phase-quality-gate.sh`는 `.tmpl`이 없어 무사. **훅 래퍼를 고칠 때는 `.sh`와 `.sh.tmpl`을 함께 고친다.**
  ```bash
  # 쌍 존재 확인 + 드리프트 점검
  for f in internal/template/templates/.claude/hooks/moai/*.tmpl; do b=${f%.tmpl}; [ -f "$b" ] && diff -q "$b" "$f" >/dev/null || echo "DRIFT $(basename $b)"; done
  ```

**참고 — 관련 결함 3건** (별도 카드 소관): ① `CleanMoaiManagedPaths`에 보호 목록 부재(근본 원인) ② `archiveLegacySkills`가 wipe **이후**(`internal/cli/update.go:554`; [2026-08-27 감사 정정])에 호출돼 원본이 이미 없어 `0 archived`로 조용히 통과 ③ `--dry-run`이 `CleanMoaiManagedPaths` 삭제 예정 목록을 미리보기하지 않음.

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
- [ ] 변경 대상 패키지 테스트 통과 (`go test ./internal/<pkg>/...`) — **전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다**. 레인 여러 개가 동시에 돌려 load 413까지 치솟고 다른 워크스페이스를 마비시킨 사고(2026-08-15)가 있다. 전 패키지 판정은 CI 몫이며, 깨끗한 환경에서 PR head를 돌리므로 근거로도 더 강하다. 예외는 §4.1의 `develop` 통합 검증 — 그때도 영향 패키지만, 병합 창을 쥔 레인 **1개만**
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

### §4.1 통합 레인 (git-flow, `develop` 원격 공유)

> **2026-08-27 전환 — 운영자 지시로 GitHub Flow → git-flow.** 이 절은 2026-08-26 백로그 카드 t281이 정한 "로컬 전용·일회용 develop, 원격 push 금지"를 **명시적으로 뒤집는다**. 되돌린 근거는 아래 §4.1.3에 남긴다.

카드별로 각각 검증해 머지했는데 **합쳐진 상태는 아무도 보지 않는** 구멍을 막는다. 2026-08-15에 PR 12개가 각각 초록불로 main에 들어갔고, 합류 후에야 `moai update`가 로컬 전용 파일을 지운다는 사실이 드러났다. 새 모델은 그 합류 지점을 **원격에 올려** CI가 판정하게 하고, 그 위에서 rc 빌드를 잘라 손으로 시험한다.

```
main (릴리스 PR로만 갱신 — 브랜치 보호 enforce_admins:true)
 ↑ 릴리스 PR (유일한 main 진입로)
release/vX.Y.Z ← develop 에서 분기
 ↑
develop (원격 공유: origin/develop, CI가 판정)
 ↑ merge --no-ff (통합 워크트리 안에서, 레인이 직접)
 ├─ worktree: 카드A ─┐
 ├─ worktree: 카드B ─┼→ 전부 develop 에서 분기, 카드 PR 없음
 └─ worktree: 카드C ─┘
```

#### §4.1.1 [HARD] 여섯 가지 규율

1. **카드 워크트리는 `develop`에서 판다.** `main`이 아니다.
2. **카드 PR은 없다.** 검증을 마친 카드는 통합 워크트리에서 로컬 `develop`에 `git merge --no-ff`로 직접 합친다. 카드 단위 CodeRabbit 리뷰도 없다.
3. **`develop`은 원격에 올린다.** `git push origin develop` — **원격 CI가 통합 판정의 주체**다. 로컬 통과는 조기 신호일 뿐이다.
4. **rc 빌드는 운영자 요청 시 `develop`에서 자른다.** `make build VERSION=vX.Y.Z-rc.N` → clean 재설치 → 손으로 시험(절차는 `.claude/rules/local/gitflow-lane-protocol.md`).
5. **로컬 시험을 통과하면 `release/vX.Y.Z`를 `develop`에서 분기한다.** main으로 가는 것은 이 브랜치뿐이다.
6. **`main`은 릴리스 PR로만 갱신된다.** 브랜치 보호(`enforce_admins: true`)가 이를 기계적으로 강제한다 — §23 참조.

레인/리드가 따르는 운영 규칙 정본은 `.claude/rules/local/gitflow-lane-protocol.md`(로컬 전용 룰)다. 이 절은 모델과 근거만 기록한다.

#### §4.1.2 운영 절차

```bash
# 리드: 통합 워크트리 1개 provisioning (raw `git worktree add` 금지 — 런처 경유)
moai cc -w develop            # .claude/worktrees/develop

# 레인: 카드 워크트리는 develop 에서 분기
#   EnterWorktree(<card-id>) → git branch -m WT-<slug>

# 레인: 병합 창 확보 → 통합 워크트리 진입 → 병합 → push
moai integration acquire --name <lane>
#   EnterWorktree(.claude/worktrees/develop)
git merge --no-ff WT-<slug>
go build ./... && go vet ./... && go test ./internal/<영향패키지>/...
git push origin develop
moai integration release
```

#### §4.1.3 2026-08-14 GitFlow 기각 사유 — 지금 어디까지 해소됐나

당시 기각 근거 두 가지의 현재 상태를 그대로 남긴다. 하나는 이번 전환으로 답이 됐고, 하나는 **여전히 살아 있다**.

- **해소됨 — `.github/workflows/`의 `branches: [main]` 트리거.** 원격에 없는 ref에는 job이 오지 않아 develop 검증이 무용하다는 지적이었다. 이번 전환에서 push 트리거 6개(`ci.yml`, `codeql.yml`, `graph-freshness.yml`, `lsel-leak-guard.yaml`, `template-neutrality-check.yaml`, `test-install.yml`)에 `develop`을 추가해 origin/develop이 CI 판정을 받는다. `pull_request` 트리거는 카드 PR이 없으므로 그대로 `[main]`이다.
- **미해소 — Vercel 프로덕션 브랜치 바인딩.** docs-site의 Vercel 프로젝트는 여전히 특정 브랜치에 묶여 있고, 이번 변경은 이를 건드리지 않았다. `develop`에 docs-site 변경이 들어갈 때 프리뷰/프로덕션 배포가 어떻게 반응하는지는 **미검증**이다 — docs-site를 만지는 카드는 이 점을 별도로 확인한다.
- **여전히 유효한 실측** — 카드별 PR의 지연은 CI가 아니라 **리뷰(CodeRabbit rate limit 포함)**가 지배했다(2026-08-15 t26). 따라서 카드 PR 폐지가 주는 이득은 **속도가 아니다**. 이번 전환의 이득은 rc 빌드를 자를 수 있는 **상시 스테이징 면(origin/develop)**을 갖는 것이다. 속도를 근거로 이 모델을 정당화하지 않는다.

#### §4.1.4 로컬 CI를 두지 않는 이유 — 검토 후 기각(2026-08-15, 유지)

- **self-hosted runner**: `modu-ai/moai-adk`는 **공개 저장소**라, 러너를 붙이면 누구나 포크 PR로 이 머신에서 임의 코드를 실행할 수 있다. 공개 저장소에 권장되지 않는 구성이다.
- **`act`**: 리눅스 컨테이너 한정이라 이 리포의 darwin×2 / windows 빌드와 macOS·Windows 통합 테스트를 재현하지 못한다. 실제 CI와 어긋나면 진단이 틀어진다.
- **비용 근거 없음**: GitHub 공식 문서 — *"GitHub Actions usage is free for self-hosted runners and for public repositories that use standard GitHub-hosted runners."* 공개 저장소는 **분 수 제한 없이 무료**다. CI를 아낄 이유가 없다.

**알려진 마찰** — BranchGuard가 읽기 전용 `git branch --list` / `git branch -vv`까지 막는다(패턴 `\bgit\s+branch\b`가 조회를 구분하지 않음). §18 독트린은 `git branch -vv`를 허용 조회로 명시하므로 과다 매칭이다(`git merge-base`는 제외 처리돼 있는 것과 대비). 우회: `git show-ref --verify refs/heads/<name>`.

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

- Package-level: 85% minimum coverage (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`)
- Critical packages (cli, template, hook): 90%+ coverage — 목표치로서 유효하나 기계적 근거는 strict 평가 프로필의 전역 "Coverage >= 90%" 게이트뿐이다(`.moai/config/evaluator-profiles/strict.md`; 패키지별 룰은 미발견 [2026-08-27 감사 정정])

### Go Test Execution Rules

- [HARD] After fixing ANY test, run the AFFECTED packages (`go test ./internal/<pkg>/...`), then push and read CI for the full-suite verdict — see §4. Do NOT run `go test ./...` locally: parallel lanes doing so drove load to 413 and stalled the machine (2026-08-15)
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

---

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

---

## 10. Build and Development Commands

빌드/테스트/린트 타깃은 `Makefile`에 `##` 도움말과 함께 정의돼 있다 — `make help` 로 조회. 템플릿 편집 → `make build` → 테스트 → 커밋 순서는 §2 [HARD] Template-First Rule 참조.

---

## 11. Frequent Issues and Solutions

자주 발생 4건 — 상세는 각 섹션 참조:
- **Templates not updated after editing** → §2 Embedded Template System (`make build` 필수, `//go:embed all:templates`).
- **Tests modify ~/.claude/settings.json** → §6 `t.TempDir()` 격리 위반. 테스트가 project root에 파일 만드는지 확인.
- **Hook timeout** → settings.json `{"timeout": 60}` (기본 5초).
- **`moai version` exit 137 (SIGKILL) after binary reinstall** → `cp bin/moai ~/go/bin/moai`만으로 부족; 기존 binary 잔재(go install buildinfo·mmap 캐시)가 꼬여 SHA가 같아도 crash. **반드시 `rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai`**(또는 `make install`)로 inode 갱신하며 clean 재설치. 맨손 `go install ./cmd/moai`는 금지 — `LDFLAGS`(Makefile:9)를 안 실어서 `pkg/version`의 컴파일 기본값(`Commit="none"`, `Date="unknown"`)이 박히고, 그러면 `strings ~/go/bin/moai | grep <sha>` 기반 binary lag 검증 자체가 불가능해진다. `make install`(Makefile:40-41; [2026-08-27 감사 정정])은 `go install $(LDFLAGS) ./cmd/moai`라 안전. 직후 `~/go/bin/moai version; echo $?`로 exit 0 확인(137이면 rm+cp 재시도). 진단 징후: `bin/moai version`=0인데 `~/go/bin/moai version`=137. binary lag 검증은 §6 검증 규율(clear → 측정).

### [HARD] 사용 중 버그·개선 발견 → 즉시 `/moai:feedback`

MoAI-ADK를 사용하다 버그나 개선이 필요한 부분을 발견하는 족족 `/moai:feedback`으로 피드백을 제출한다 — 세션을 마친 뒤 몰아서 남기지 않는다. 대상: `moai` CLI 동작, 훅, 템플릿, 스킬, 에이전트, 팩토리·칸반 운영 결함 전반. 재현 명령과 관측된 출력을 함께 남긴다. 구분: 유지자에게 보고할 사안은 `/moai:feedback`, 작업으로 예정할 사안은 `/moai todo add`.

---

## 12. YAML Frontmatter 빠른 참조

범용 형식 규칙은 `.claude/rules/moai/development/` 내 `skill-authoring.md`, `agent-authoring.md`에 정의.

### 로컬 개발 체크리스트

- [ ] `tools:`, `allowed-tools:` → CSV string (공백 구분 절대 금지)
- [ ] `skills:` → YAML array (유일한 예외)
- [ ] `metadata.*` → quoted string
- [ ] Template 수정 후 `make build` 실행
- [ ] Local copy (`.claude/`)도 동기화

탐지 스크립트: 전역 프로젝트 메모리(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)의 `audit_sweep_patterns.md` Pattern A 참조.

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

## 15. 템플릿 프로그래밍-언어 중립성

> **[HARD] 용어 — "언어"는 두 축을 가리킨다. 수식어 없이 쓰지 않는다.**
>
> | 축 | 값 | 요구되는 것 | 지배 규칙 |
> |---|---|---|---|
> | **프로그래밍 언어** (16) | go, python, typescript, … swift | **중립** — 어느 하나를 PRIMARY로 두지 않음 | 이 §15 |
> | **사용자 대화 로케일** (4) | ko, en, ja, zh | **번역** — 로케일마다 자연스러운 원어 | 활성 output style §8 Localization Contract |
>
> 이 둘은 서로 독립이다. 템플릿은 16개 프로그래밍 언어에 중립이면서, 동시에 4개 로케일로 번역된다.
> "16개 언어"를 배포 로케일 수로 읽는 것이 반복 관측된 오독이므로, 문서에서는 항상
> **"프로그래밍 언어"** 또는 **"로케일"** 로 수식해 쓴다.

### [HARD] `internal/template/templates/` 하위는 16개 프로그래밍 언어를 동등 취급

도구의 구현 언어(Go)와 사용자 프로젝트의 프로그래밍 언어는 별개. 템플릿은 모든 사용자를 위한 것.

- 프로그래밍-언어 편향 허용: `CLAUDE.local.md`, `settings.local.json`, 로컬 `.moai/config/`
- 프로그래밍-언어 편향 금지: `internal/template/templates/**` 전체

### 16개 지원 프로그래밍 언어 (모두 동등)

```
go, python, typescript, javascript, rust, java, kotlin, csharp,
ruby, php, elixir, cpp, scala, r, flutter, swift
```

Dart/Flutter 캐논 이름: **"flutter"** (not "dart").

### 체크리스트 (템플릿 수정 시)

- [ ] 특정 언어를 "PRIMARY"로 배치하지 않았는가?
- [ ] 16개 프로그래밍 언어가 동등 수준으로 나열되어 있는가?
- [ ] 특정 언어만 "enabled", 나머지 "planned"로 격하하지 않았는가?
- [ ] project_markers 기반 자동 감지 로직이 포함되어 있는가?
- [ ] 로컬 config와 템플릿이 달라도 정상 (같으면 오히려 의심)

상세 교훈: 전역 프로젝트 메모리(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)의 `lessons.md` #5 참조.

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

상세 교훈 및 5 Whys: 전역 프로젝트 메모리(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)의 `lessons.md` #4 참조.

---

## 17. docs-site 4개국어 문서 동기화 규칙

docs-site는 `adk.mo.ai.kr` 공식 사용자 문서. URL 표준, 4-locale 동기화 의무, Mermaid TD-only, Vercel 프로젝트 바인딩, 빌드/배포 체크리스트 등 전체 doctrine은 외부 파일 참조.

See: `.moai/docs/docs-site-i18n-rules.md`

### §17.1 디자인 컴포넌트 + 아이콘 규약 → See `.moai/docs/docs-site-design-components.md`

docs-site Claude Warm Editorial [HARD] 세부(아이콘 shortcode·코드블록·Mermaid·푸터·CSS 캐시 버스팅·라이트 단일 테마)는 외부 파일로 이관. 핵심만: 본문 장식 이모지 금지→`{{</* icon */>}}` shortcode; 사이드바 `icon:` 값은 `menu.html` SVG case 필수; CSS 수정 후 dev 반영 안 되면 hugo 재시작.

---

## References

Sections §18-27 were consolidated into external `.moai/docs/` files to reduce launch-time context. Each entry below is the authoritative location for its domain.

- **§5 Version Management** (SemVer pre-release, ldflags injection, release process): `.moai/docs/version-management.md`
- **§7 Hook Development** (shell-script-only pattern, settings.json format, quoting rules): `.moai/docs/hook-development.md`
- **§18 Git Workflow** (Enhanced GitHub Flow 본문 + [2026-08-27] git-flow 전환 상위모델 노트 — 정본은 §4.1, branch protection `enforce_admins: true`, Hybrid Trunk RETIRED): `.moai/docs/git-workflow-doctrine.md`
- **§19 AskUserQuestion Enforcement + §19.1 Implementation Kickoff Approval Mandatory Restoration** (REQ-ATR-015): canonical SSOT at `.claude/rules/moai/core/askuser-protocol.md` + `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E (the gate is mandatory and score-independent; plan-auditor PASS never auto-bypasses it)
- **§20 Vercel Build Cost Guard** [HARD]: all Vercel projects MUST use Elastic build machine ($0.0035/CPU min vs Turbo $0.126/min); check Build Machine setting first on cost anomalies
- **§21 Dev-Only Commands Isolation** (split harnesses, `SPLIT_HARNESS_NAMESPACE_LEAK` sentinel): `.moai/docs/dev-only-commands-isolation.md`
- **§22 Dev Settings Intent** (settings.json key semantics): `.moai/docs/local-dev-settings-intent.md`
- **§23 Local Git Workflows** (PR-mandatory 1-person OSS, all tiers via PR — 릴리스 PR 경로; [2026-08-27] 카드 작업은 git-flow `develop` 병합으로 전환, 상위모델 노트 참조): `.moai/docs/git-local-workflow-doctrine.md`
- **§24 Harness Namespace** (template-managed vs user-owned separation): `.moai/docs/harness-namespace-doctrine.md`
- **§25 Template Internal-Content Isolation** (neutrality catalogue, CI guard): `.moai/docs/template-internal-isolation-doctrine.md`
- **§26 Linear 연동** (local-only): `.moai/docs/local-linear-integration.md`
- **§27 Agent-Skill Architecture** [HARD]: every agent gets ≥1 skill set (4 elements: workflow skill + knowhow reference + scripts + trigger); all skill bodies in English; `/moai:<sub>` slash-wrapping maintained; `/moai:harness` meta-harness (v4 Builder); Analyze-First execution plan before any work

---

## 28. LSEL 드레인 운영 (지역 자가진화 루프)

> 복원 2026-08-26 (SPEC-LSEL-DRAIN-STALL-001 M2). 이 섹션이 과거 진행 문서들이 참조하던 "§28" 앵커의 실체다 — 3주 드레인 정지(2026-08-04~25) 당시 이 섹션이 소실돼 있었다. 정본은 `.claude/skills/hns-lsel-curator/SKILL.md` § Durable operations.

### [HARD] 내구 트리거 — 세션 시작 배선

드레인은 어떤 Claude 세션의 생사에도 의존하지 않는다. `.claude/settings.local.json` `.hooks.SessionStart`에 2개 lsel 항목이 배선돼 있다(각 `"timeout": 30`): (1) `session_drain.sh` 래퍼 — 배타잠금 → 기존 clusters.json 무조건 보존(`clusters-history/`) → `drain.sh` 실행 → 1행 상태 → fail-open. (2) `backlog_check.sh` advisory — 읽지 않은 백로그가 임계(기본 25, `LSEL_BACKLOG_THRESHOLD`) 초과면 system-reminder 발화. **이 배선이 settings.local.json(로컬 전용)에 있는 것은 의도적이다** — tracked `settings.json`은 `moai update`가 통째 재배포해 배선이 매 업데이트 유실된다(§2.3). 검증: `jq '.hooks.SessionStart' .claude/settings.local.json`.

### [HARD] 모든 드레인은 래퍼 경로로

`drain.sh`를 **직접** 호출하지 않는다 — 직접 호출은 호출-전 보존을 우회해 스테이징된 후보를 조용히 유실시킨다(drain.sh는 드레인 경로든 no-op 경로든 clusters.json을 덮어쓴다). 항상 `session_drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel`.

### PROPOSE는 archived 사본 판독

세션 시작마다 드레인이 도는 체제에서 live `clusters.json`은 휘발성이다(no-op 드레인조차 `candidates: []`로 덮어쓴다). 후보 제안(PROPOSE)은 `.moai/state/lsel/clusters-history/`의 사본(최신순)을 읽는다. 검증 레시피·mutant guard 포함 전체 절차는 SKILL.md § Verification.
