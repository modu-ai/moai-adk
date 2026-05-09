# SPEC-V3R2-RT-007 Deep Research (Phase 0.5)

> Research artifact for **Hardcoded Path Fix + Versioned Migration**.
> Companion to `spec.md` (v0.1.0). Authored against branch `plan/SPEC-V3R2-RT-007` from `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007` (worktree mode).

## HISTORY

| Version | Date       | Author                                  | Description                                                                                  |
|---------|------------|-----------------------------------------|----------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research. Documents the gap between `spec.md` (drafted 2026-04-23) and the current codebase: the hardcoded literal `/Users/goos/go/bin/moai` no longer appears in any of the 28 shell wrapper templates — the template-side fix is already shipped; this SPEC's path-fix half therefore *affirms and hardens* the template parity, while the migration-framework half remains entirely new work (no `internal/migration/` package exists yet). |

---

## 1. Goal of Research

Substantiate `spec.md` §1 (Goal), §2 (Scope), §3 (Environment), §4 (Assumptions), §7 (Constraints), §8 (Risks) with concrete file:line evidence so that the run phase can implement REQ-V3R2-RT-007-001..061 against a known-good baseline.

The research answers seven questions:

1. **현재 하드코딩 상태 검증**: 28개 shell wrapper 템플릿에 `/Users/goos/go/bin/moai` 리터럴이 실제로 몇 번 등장하는가? (스펙 작성 시점 가정과 코드 현실의 불일치 여부)
2. **GoBinPath 리졸버 현황**: `go env GOBIN` → `go env GOPATH/bin` → `$HOME/go/bin` 폴백 체인이 어디에 어떻게 구현되어 있는가?
3. **`internal/migration/` 패키지 부재 확인**: 마이그레이션 프레임워크의 어떤 부분이 이미 존재하고, 무엇을 새로 만들어야 하는가?
4. **`moai migrate agency` 패턴 분석**: 기존 일회성 마이그레이션 명령은 어떻게 구성되어 있고, 이를 framework로 일반화하기 위한 차이는 무엇인가?
5. **CC `CURRENT_MIGRATION_VERSION = 11` X-5 패턴 차이**: Claude Code의 silent preAction 마이그레이션 모델과 moai의 명시적 CLI 모델 간 격차?
6. **session-start hook preAction 통합**: `internal/hook/session_start.go`에 마이그레이션 러너 호출을 어떻게 끼워넣어야 하는가?
7. **RT-001 / RT-006 의존성 상태**: 두 SPEC의 머지 상태가 RT-007 run phase 진입에 미치는 영향?

---

## 2. Inventory of `internal/template/` and hook wrappers (existing)

### 2.1 Hook wrapper template inventory

`ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007/internal/template/templates/.claude/hooks/moai/` 결과 (verified 2026-05-10):

- 총 **28개** `*.sh.tmpl` 파일 (spec.md §1의 "26 shell wrappers" 표현과 ±2 오차)
- 모든 파일이 동일한 fallback chain 구조 (검증: `handle-session-start.sh.tmpl`, `handle-session-end.sh.tmpl`, `handle-subagent-stop.sh.tmpl` 인스펙트)

목록 (28 wrappers):

```
handle-agent-hook.sh.tmpl              handle-permission-denied.sh.tmpl
handle-compact.sh.tmpl                 handle-permission-request.sh.tmpl
handle-config-change.sh.tmpl           handle-post-compact.sh.tmpl
handle-cwd-changed.sh.tmpl             handle-post-tool-failure.sh.tmpl
handle-elicitation-result.sh.tmpl      handle-post-tool.sh.tmpl
handle-elicitation.sh.tmpl             handle-pre-tool.sh.tmpl
handle-file-changed.sh.tmpl            handle-session-end.sh.tmpl
handle-harness-observe.sh.tmpl         handle-session-start.sh.tmpl
handle-instructions-loaded.sh.tmpl     handle-stop-failure.sh.tmpl
handle-notification.sh.tmpl            handle-stop.sh.tmpl
                                       handle-subagent-start.sh.tmpl
                                       handle-subagent-stop.sh.tmpl
                                       handle-task-completed.sh.tmpl
                                       handle-task-created.sh.tmpl
                                       handle-teammate-idle.sh.tmpl
                                       handle-user-prompt-submit.sh.tmpl
                                       handle-worktree-create.sh.tmpl
                                       handle-worktree-remove.sh.tmpl
```

Spec.md (drafted 2026-04-23) cites 26; the count grew to 28 in interim work. 본 SPEC의 모든 측정 가치 진술은 28을 사용하며, 정확한 산출은 `ls .../handle-*.sh.tmpl | wc -l` 한 줄로 검증 가능.

### 2.2 Hardcoded path literal — current state

`grep -rln "/Users/goos/go/bin/moai" internal/template/templates/.claude/hooks/moai/` 결과: **0 hits** (28개 파일 모두 깨끗).

`grep -rln "/Users/goos\|/home/goos" internal/template/templates/`:

```
internal/template/templates/.claude/rules/moai/workflow/session-handoff.md   # docs example
internal/template/templates/.claude/skills/moai/SKILL.md                      # docs example
```

→ 두 문서는 워크트리 paste-ready 예시 + 외부 절대경로 cross-reference 용도. **셸 wrapper에서는 하드코딩이 이미 제거되어 있다.** `handle-session-start.sh.tmpl:13-34` 발췌:

```bash
# Try moai command in PATH
if command -v moai &> /dev/null; then
    exec moai hook session-start < "$temp_file" 2>/dev/null
fi

# Try detected Go bin path from initialization
if [ -f "{{posixPath .GoBinPath}}/moai" ]; then
    exec "{{posixPath .GoBinPath}}/moai" hook session-start < "$temp_file" 2>/dev/null
fi

# Try default ~/go/bin/moai
if [ -f "$HOME/go/bin/moai" ]; then
    exec "$HOME/go/bin/moai" hook session-start < "$temp_file" 2>/dev/null
fi

# Try ~/.local/bin/moai (Linux install location)
if [ -f "$HOME/.local/bin/moai" ]; then
    exec "$HOME/.local/bin/moai" hook session-start < "$temp_file" 2>/dev/null
fi

exit 0
```

현재 fallback 체인: **PATH → `{{posixPath .GoBinPath}}/moai` → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` → exit 0**

이는 spec.md §5 REQ-003가 명시하는 *target* 체인 — `moai PATH → $HOME/go/bin/moai → {{posixPath .GoBinPath}}/moai → silent exit` — 와 **순서가 다르다**. 현재 코드는 `.GoBinPath`(init-time 측정값) 우선, `$HOME` 폴백; spec.md는 `$HOME` 우선, `.GoBinPath` 후순. RT-007 run phase는 이 순서 결정을 다시 검토해야 한다 (§6 참조). spec.md v0.1.0 본문은 이미 코드 현실에 맞게 보정되어 있음 (REQ-003 텍스트 reorder).

### 2.3 What spec.md §1 still gets right

스펙 §1 본문이 완전히 무효화된 것은 아니다. 다음 두 사실은 그대로 유효:

1. v2.x 사용자 프로젝트의 `.claude/hooks/moai/handle-*.sh` 가 (release 시점 hardcoded 였다면) 여전히 그대로 디스크에 남아 있다. 이들에게 fix를 retroactively 전파할 메커니즘이 필요하다 → 마이그레이션 1.
2. CI lint (REQ-V3R2-RT-007-051)는 회귀 방지용으로 여전히 정당하다. *"이미 깨끗하니 lint는 불필요"* 라는 주장은 위험 — 미래 contributor가 ad-hoc Build script로 절대경로를 다시 emit할 수 있다.

따라서 path-fix half의 본질은 변경되었다:

- **이전 가정**: "26개 wrapper 모두 새로 작성 + 사용자에게 fix 전파"
- **현재 현실**: "wrapper 28개는 이미 깨끗 + CI lint로 회귀 잠금 + 마이그레이션 1로 v2.x 사용자에게도 retroactively 전파 + fallback 체인 순서 재확인"

### 2.4 GoBinPath resolver — current locations

`grep -rn "GoBinPath\|detectGoBin" internal/`:

| File:line | Purpose | Notes |
|-----------|---------|-------|
| `internal/template/context.go:51` | `TemplateContext.GoBinPath string` field declaration | "Detected Go binary installation path (e.g., \"/home/user/go/bin\")" |
| `internal/template/context.go:209` | `WithGoBinPath(path) ContextOption` setter | functional-option pattern |
| `internal/core/project/initializer.go:257` | `goBinPath := detectGoBinPath(homeDir)` (init path) | called from `Initializer.Initialize()` |
| `internal/core/project/initializer.go:286-303` | `func detectGoBinPath(homeDir string) string` 정의 | go env GOBIN → GOPATH/bin → $HOME/go/bin → /usr/local/go/bin (or C:\Go\bin on Windows) |
| `internal/cli/update.go:528, 560` | `goBinPath := detectGoBinPathForUpdate(homeDir)` (update path) | called from update flow |
| `internal/cli/update.go:2515-2540` | `func detectGoBinPathForUpdate(homeDir string) string` 정의 | duplicate of `detectGoBinPath` with identical logic |

**중요 관찰**: 동일한 detection 로직이 **두 군데** (`initializer.go:286` + `update.go:2515`) 중복 존재. 공통 helper로 추출하면 REQ-V3R2-RT-007-001의 "단일 GoBinPath 리졸버" 의도와 정합. RT-007 run phase는 이를 `internal/runtime/gobin/resolver.go` 로 통합 후 두 호출처가 import하도록 리팩터.

### 2.5 Renderer + claudeCodePassthroughTokens

`internal/template/renderer.go:39-50` (`claudeCodePassthroughTokens`):

```go
var claudeCodePassthroughTokens = []string{
    "$CLAUDE_PROJECT_DIR",
    "$CLAUDE_SKILL_DIR",
    "$ARGUMENTS",
    "$HOME",                    // ← already registered (CLAUDE.local.md §14 affirmed)
    "$GITHUB_OUTPUT",
    "$GITHUB_ENV",
    "$GITHUB_STEP_SUMMARY",
    "$LANG_COUNT",
}
```

`$HOME`는 이미 등록되어 있어 REQ-V3R2-RT-007-006 ("$HOME shall be registered in claudeCodePassthroughTokens") 은 **affirm-only** — 회귀 방지 테스트(`internal/template/renderer_passthrough_test.go`) 만 추가하면 충분. 신규 코드 변경은 0 LOC.

---

## 3. `internal/migration/` 패키지 부재 + 기존 `migrate agency` 분석

### 3.1 `internal/migration/` 디렉터리 부재 확인

```
$ ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007/internal/migration/
ls: ...: No such file or directory
```

Spec.md §3에서 명시한 "Affected modules" 4개 중 다음은 **새로 만들어야 한다**:

- `internal/migration/runner.go` — 신규
- `internal/migration/migrations/m001_hardcoded_path.go` — 신규
- `internal/migration/registry.go` — 신규
- `internal/migration/version.go` — 신규
- `internal/migration/log.go` — 신규

### 3.2 기존 `moai migrate agency` 명령 분석

`internal/cli/migrate_agency.go:1-700+` 검사 결과:

- 단일 cobra 명령 `migrateCmd` (`migrate_agency.go:573-578`) 에 서브명령 `migrateAgencyCmd` (`migrate_agency.go:580-598`) 가 등록됨.
- Group ID: `"project"`.
- 플래그: `--dry-run`, `--force`, `--resume <txID>`.
- 진입함수: `runMigrateAgency` at `migrate_agency.go:609-`.
- **체크포인트**: `migrationCheckpoint{TxID, ProjectRoot, CompletedPhases, RemainingFiles, Timestamp}` (line 63-69) — 트랜잭션 진행 상태 영속화 메커니즘이 이미 존재.
- **롤백**: `transactionLog{createdDirs, createdFiles}` (line 71-74) — 생성된 경로 기록 후 atomic 롤백.
- **에러 코드 enum**: `ErrMigrate*` 9종 (line 19-39) — 머신리더블 에러 분류 베이스라인.

이 패턴은 RT-007의 마이그레이션 프레임워크에 그대로 차용 가능:

- `migrationCheckpoint` ↔ RT-007의 `.moai/state/migration-version` (단순화: 단일 정수 버전 + per-migration 상태 로그).
- `transactionLog` ↔ RT-007의 per-migration `Rollback(projectRoot string) error` (선언적 롤백 함수).
- `ErrMigrate*` enum 패턴 ↔ RT-007의 `MigrationNotRollbackable`, `MigrationReadOnly`, `DuplicateMigrationVersion` 등 (REQ-V3R2-RT-007-052/053/024).

### 3.3 신규 vs 기존 라우팅

| 명령 | 위치 | 책임 |
|------|------|------|
| `moai migrate agency` (기존) | `internal/cli/migrate_agency.go` | 일회성, .agency/ → .moai/ 한 번 옮기고 종료 |
| `moai migration run` (신규, REQ-040) | `internal/cli/migration.go` (신규) | 등록된 모든 pending 마이그레이션 실행 |
| `moai migration status [--json]` (신규, REQ-041) | 동일 | 현재 버전 + pending + last-applied 출력 |
| `moai migration rollback <version>` (신규, REQ-024, -042) | 동일 | 선언된 Rollback 호출, 버전 -1 |
| `moai doctor migration-status` (신규, REQ-015) | `internal/cli/doctor.go` 확장 | doctor 출력에 마이그레이션 섹션 추가 |

`moai migrate` (cobra group) 와 `moai migration` (신규 cobra group) 의 **이름 충돌 회피**: 기존 `migrateCmd` 는 별도 그룹으로 보존 — 의미상 `migrate agency` 는 일회성, `migration *` 는 framework. RT-007 run phase에서 최종 결정.

---

## 4. CC `CURRENT_MIGRATION_VERSION = 11` X-5 패턴

### 4.1 Reference 자료 위치

- `r3-cc-architecture-reread.md` §2 Decision 10: "Versioned migrations with preAction auto-apply. CURRENT_MIGRATION_VERSION = 11 + preAction hook + per-migration idempotency guards."
- `pattern-library.md` X-5 row: "Versioned migration auto-apply at preAction. ADOPT — moai's current `moai migrate agency` is explicit; silent preAction align with CC and reduce support burden."

### 4.2 CC와 moai의 차이

| 차원 | Claude Code (X-5) | moai 현재 (`migrate agency`) | moai RT-007 (목표) |
|------|------------------|---------------------------|--------------------|
| 트리거 | preAction hook (silent, automatic) | 사용자 CLI 명시 호출 | session-start hook (silent) |
| 멱등성 | Per-migration guard | 트랜잭션 단위 (모두 or 무) | Per-migration guard + version file |
| 버전 관리 | `CURRENT_MIGRATION_VERSION` 정수 | 없음 (단일 마이그레이션) | `.moai/state/migration-version` 정수 |
| 레지스트리 | Compile-time static | 없음 | Compile-time static (REQ-V3R2-RT-007-016) |
| 관측성 | Internal logs | stdout (CLI 모드) | `moai doctor migration-status` + `.moai/logs/migrations.log` |
| 롤백 | 없음 (forward-only) | 트랜잭션 단위 자동 롤백 | Opt-in `Rollback` 함수 (REQ-024, -042) |

### 4.3 차용 결정

- **silent preAction**: ADOPT (REQ-020)
- **per-migration idempotency guard**: ADOPT (REQ-011)
- **monotonic 정수 version**: ADOPT (REQ-013)
- **forward-only**: REJECT — moai는 opt-in `Rollback` 지원 (REQ-024, -042). CC가 안 한 이유는 LSP 클라이언트라는 운영 환경 차이; moai는 CLI 도구라 사용자 통제권이 더 큼.

---

## 5. session-start hook 통합 점

### 5.1 `internal/hook/session_start.go` 구조

`internal/hook/session_start.go:23-60` 발췌:

```go
type sessionStartHandler struct {
    cfg ConfigProvider
}

func (h *sessionStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
    slog.Info("session started", ...)
    data := map[string]any{
        "session_id": input.SessionID,
        "status": "initialized",
    }
    cfg := h.getConfig()
    if cfg != nil {
        if cfg.Project.Name != "" {
            data["project_name"] = cfg.Project.Name
        }
        ...
    }
    return &HookOutput{...}, nil
}
```

Handle 진입 직후 (project config load 와 동등 위계) 가 마이그레이션 러너 호출의 자연스러운 자리. 패턴:

```go
// 마이그레이션 자동 적용 (RT-007 REQ-020)
if !migrationsDisabled(cfg) {
    runner := migration.NewRunner(input.ProjectDir)
    if applied, err := runner.Apply(ctx); err != nil {
        slog.Warn("migration apply failed (non-blocking)", "error", err)
        // SystemMessage로 surface (RT-001 HookResponse 의존)
        output.SystemMessage = fmt.Sprintf("migration error: %v", err)
    } else if len(applied) > 0 {
        slog.Info("migrations applied", "versions", applied)
    }
}
```

### 5.2 RT-001, RT-006 의존성 상태

- **RT-001 (Hook JSON protocol with SystemMessage)**: spec.md §9.1 Blocked-by에 명시. SystemMessage 필드를 통한 마이그레이션 실패 surfacing은 RT-001의 HookResponse 스키마 확정에 의존. RT-001 이 머지 안 되었으면 임시로 `slog.Warn` 사용 + RT-001 머지 후 SystemMessage로 swap (single-line change).
- **RT-006 (SessionStart handler completeness)**: spec.md §9.1 Blocked-by에 명시. RT-006는 현재 핸들러의 완성도 (config load, evolution scheduler 통합 등) 를 보장. RT-007의 마이그레이션 호출은 RT-006 이후에 들어가는 것이 안전 — RT-006 미완 상태에서 추가하면 두 SPEC 의 변경이 같은 함수 본문에 충돌.

### 5.3 Run-phase 진입 조건

RT-007 run phase 진입 시 plan-auditor가 다음을 verify해야 한다:

1. RT-001 의 PR 머지 상태 — 머지 완료 시 SystemMessage 필드 사용 가능, 미완 시 stub.
2. RT-006 의 PR 머지 상태 — 머지 완료 시 RT-007 변경이 conflict-free, 미완 시 두 SPEC의 stash 합류 필요.
3. RT-007 의 templates wrapper 28개 모두 hardcoded 0 검증 (회귀 방지 baseline).

---

## 6. fallback chain 순서 재검토

### 6.1 spec.md §5 REQ-003 vs 현재 구현

코드 현실 (`handle-session-start.sh.tmpl:13-34`): *"PATH → `{{posixPath .GoBinPath}}/moai` → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` → exit"*.

spec.md v0.1.0 §5.1 REQ-003 (이번 재작성에서 코드 현실에 맞춰 amend됨): `moai` in PATH → `{{posixPath .GoBinPath}}/moai` → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` → silent exit.

### 6.2 우선순위 결정 — `.GoBinPath` 우선 (현재 코드 유지)

근거:

- `.GoBinPath` 는 `moai init` / `moai update` 시점에 측정된 정확한 경로 (실제 사용자 환경에서 `go env GOBIN` → `GOPATH/bin` → `$HOME/go/bin` 폴백 결과). 즉 `.GoBinPath` 가 `$HOME/go/bin` 과 다르면 사용자가 `GOBIN` 또는 `GOPATH` 를 명시 설정한 경우뿐이며, 이 경우 사용자 의도는 `.GoBinPath` 우선.
- `$HOME/go/bin/moai` 는 *`.GoBinPath` 측정 실패 시* 의 최후 폴백.
- spec.md REQ-003 의 의도는 "절대경로 회귀 방지 + 폴백 체인 명시"; 순서는 `.GoBinPath` 우선이 정합.

### 6.3 4번째 폴백: `$HOME/.local/bin/moai`

Linux apt-get / Snap 배포 호환성을 위해 4번째 폴백으로 보존. spec.md REQ-003에 명시.

---

## 7. STALE_SECONDS와 마이그레이션 간 무관성 확인

RT-004 (Typed Session State) 는 `STALE_SECONDS = 3600` 을 도입; 이는 *checkpoint 파일* 의 TTL이지 *migration-version 파일* 의 TTL이 아니다. RT-007의 `.moai/state/migration-version` 은 TTL이 없는 *영속* 정수 (마이그레이션이 advance 시키지 않으면 변하지 않음). RT-004 와 RT-007의 `.moai/state/` 서식은 동일 디렉터리를 공유하지만 두 SPEC의 staleness 의미는 분리됨.

검증: `.moai/state/migration-version` 는 spec.md §3 environment 에서 "version guard"로 명시되며, hydrate stale-check (RT-004 §5.3 REQ-014) 와는 다른 코드 경로에서 검사됨.

---

## 8. Idempotency 설계 — 마이그레이션 1 케이스 스터디

### 8.1 Migration 1 의 트리거 조건

`spec.md` §5.3 REQ-022:

> WHEN migration 1 runs on a user project whose `.claude/hooks/moai/handle-*.sh` contains the literal `/Users/goos/go/bin/moai`, the migration SHALL rewrite the file replacing the literal with `$HOME/go/bin/moai`, preserving all other content.

### 8.2 Idempotent re-run

`spec.md` §5.3 REQ-023:

> WHEN migration 1 runs on a project where shell wrappers do NOT contain the hardcoded literal (already migrated or fresh install), the migration SHALL treat it as a no-op and write a log entry `result: success, details: "already migrated"`.

### 8.3 알고리즘 (run phase 구현 가이드)

```
func m001_hardcoded_path_Apply(projectRoot string) error {
    hookDir := filepath.Join(projectRoot, ".claude", "hooks", "moai")
    files, err := filepath.Glob(filepath.Join(hookDir, "handle-*.sh"))
    if err != nil { return err }

    rewritten := 0
    for _, f := range files {
        body, err := os.ReadFile(f)
        if err != nil { continue }  // skip unreadable
        if !bytes.Contains(body, []byte("/Users/goos/go/bin/moai")) {
            continue  // already clean
        }
        info, _ := os.Stat(f)  // preserve mode
        newBody := bytes.ReplaceAll(body, []byte("/Users/goos/go/bin/moai"), []byte("$HOME/go/bin/moai"))
        if err := writeAtomic(f, newBody, info.Mode().Perm()); err != nil { return err }
        rewritten++
    }
    // log entry: {result: success, details: "rewritten N file(s)" or "already migrated"}
    return nil
}
```

검증: rewriting은 `bytes.ReplaceAll` 로 단일 패스, 다른 콘텐츠 보존. atomic rename 으로 partial-write 방지. `os.ReadFile`/`os.WriteFile` perm 0o755 (executable) 유지.

---

## 9. Risk research (extends spec.md §8)

### 9.1 spec.md drafted 2026-04-23 → 코드 reality 2026-05-10 격차

- **Risk**: spec.md의 "26 wrappers with hardcoded path" 명제가 invalid → run-phase implementer가 spec 텍스트 그대로 따라가다가 *없는 hardcoded literal을 찾는* 작업으로 시간 낭비.
- **Mitigation**: research.md §2.2 + plan.md §3에서 명시. spec.md v0.1.0 본문은 본 plan-phase에서 코드 현실에 맞게 수정됨.

### 9.2 Migration 1 "no-op" 사례가 압도적

- **Risk**: 28개 wrapper가 *이미 깨끗* 한 상태로 v3.0 출시되므로, fresh install 사용자는 migration 1 을 실행해도 모든 파일이 "already migrated"로 처리됨. v2.x → v3 업그레이드 사용자만 의미 있는 rewrite 발생.
- **Mitigation**: 이는 실제로 *기대 동작*이며 risk라기보다 *expected behavior*. 다만 telemetry / metric 차원에서 "fresh install vs upgrade" 구분이 불가능 → migration log entry에 `details: "already migrated (N files scanned, 0 rewritten)"` 로 분포 측정 가능하게 기록.

### 9.3 Cobra 명령 그룹 충돌 (`migrate` vs `migration`)

- **Risk**: `moai migrate agency` (기존) + `moai migration run` (신규) 두 그룹의 이름 유사성으로 사용자 혼동.
- **Mitigation**: `moai migration --help` 첫 줄에 "for one-off agency migration, see `moai migrate agency`" 명시. `moai migrate --help` 도 역으로 cross-reference.

### 9.4 Concurrent migration apply (두 세션 동시 시작)

- **Risk**: 사용자가 두 터미널에서 동시에 `claude` 시작 → 두 session-start hook이 동시 마이그레이션 적용 시도 → version-file 레이스.
- **Mitigation**: spec.md §5.4 REQ-031 advisory lock (RT-004의 `internal/session/lock_*.go` primitive 재사용 가능). `os.Rename` atomic 보장 + advisory lock 으로 critical section 보호. 3-retry / 10ms backoff (RT-004 §5.5 REQ-040 와 동일 패턴).

### 9.5 마이그레이션 1 의 perm 보존

- **Risk**: shell wrapper는 executable 비트 (`0o755`) 가 필수. atomic rename 시 파일 모드 손실 가능.
- **Mitigation**: `writeAtomic(path, data []byte, perm os.FileMode)` helper에 `perm` 파라미터 명시. 호출자가 `os.Stat(path).Mode()` 로 기존 모드 읽고 전달. 테스트: `m001_*_test.go::TestApply_PreservesExecutableBit`.

### 9.6 Windows shell 호환성

- **Risk**: spec.md §5.7 REQ-060 — Windows Git Bash / WSL / MSYS2 환경에서 `$HOME/go/bin/moai` 경로 expansion 차이.
- **Research**: `$HOME` 은 Git Bash, WSL, MSYS2 모두 동일하게 user home으로 expand. Migration 1 은 *문자열 치환만* 수행하고 expansion은 shell이 런타임에 함; 따라서 migration의 platform-independent 보장.
- **Mitigation**: 이미 spec.md REQ-060이 mitigation을 기술. 추가 필요 없음.

### 9.7 `moai doctor migration-status` UX 일관성

- **Risk**: spec.md §5.2 REQ-015는 doctor의 sub-subcommand 형태 (`moai doctor migration-status`). 그러나 `internal/cli/doctor.go:43`의 `doctorCmd` 는 단일 명령 + 플래그 (`--check`) 패턴. sub-subcommand 추가 시 cobra 구조 변경 필요.
- **Mitigation**: `--check migration` 패턴 사용 (UX 일관성 및 cobra 변경 최소화). `moai migration status` (별도 cobra group) 가 main 진입; doctor는 health-check 형태로 보조 surface 제공.

---

## 10. File:line evidence anchors

다음 anchor들은 run-phase에서 load-bearing이며, plan.md §3.4에서 verbatim 인용된다:

1. `spec.md:50-66` — In-scope items (path fix half + migration framework half).
2. `spec.md:122-175` — 32 EARS REQs (REQ-001 through REQ-061).
3. `spec.md:179-194` — 16 ACs (AC-01 through AC-16).
4. `spec.md:198-205` — Constraints.
5. `internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl:13-34` — Current fallback chain (verified 0 hardcoded literals).
6. `internal/template/templates/.claude/hooks/moai/handle-session-end.sh.tmpl:13-34` — Same chain pattern (28 wrappers identical).
7. `internal/template/renderer.go:39-50` — `claudeCodePassthroughTokens` (`$HOME` already registered at index 3).
8. `internal/template/context.go:51` — `TemplateContext.GoBinPath` field.
9. `internal/template/context.go:209-211` — `WithGoBinPath` setter.
10. `internal/core/project/initializer.go:257` — `goBinPath := detectGoBinPath(homeDir)` call.
11. `internal/core/project/initializer.go:286-303` — `detectGoBinPath` definition (init path).
12. `internal/cli/update.go:528, 560` — `detectGoBinPathForUpdate` call sites (update path).
13. `internal/cli/update.go:2515-2540` — `detectGoBinPathForUpdate` definition (duplicate of initializer's).
14. `internal/cli/migrate_agency.go:1-700` — Existing `moai migrate agency` (one-shot pattern, reference for framework design).
15. `internal/cli/migrate_agency.go:573-607` — `migrateCmd` cobra wiring.
16. `internal/cli/doctor.go:43-58` — `doctorCmd` definition (extension target for migration-status).
17. `internal/hook/session_start.go:23-60` — `sessionStartHandler.Handle` (extension target for MigrationRunner.Apply call).
18. `.moai/state/.gitkeep` — Empty state directory (new files written here at runtime).
19. `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary — Subagent prohibition (load-bearing for SystemMessage routing constraint).
20. `.claude/rules/moai/workflow/spec-workflow.md` — Plan Audit Gate.
21. `CLAUDE.local.md:§14` — `.HomeDir` 금지 / `$HOME` 사용 규칙.
22. `CLAUDE.local.md:§2` — 보호 디렉터리 (`.moai/state/` 포함).
23. `CLAUDE.local.md:§6` — Test isolation (`t.TempDir()`).
24. `r3-cc-architecture-reread.md:§2 Decision 10` — CC `CURRENT_MIGRATION_VERSION = 11` 패턴.
25. `pattern-library.md:X-5` — Versioned migration auto-apply at preAction.
26. `problem-catalog.md:P-H04` — Hardcoded absolute path in shell wrappers (CRITICAL).
27. `problem-catalog.md:P-C06` — No silent migration pattern.
28. `master.md:§8 BC-V3R2-008` — AUTO migration commitment.
29. `internal/template/renderer_test.go:359-370` — Existing test 'windows_gobinpath_converted_in_shell' (regression baseline).

Total: **29 distinct file:line anchors** (exceeds plan-auditor minimum of 10).

---

## 11. External library evaluation

이 SPEC은 외부 라이브러리 추가가 거의 0이다. 단 두 가지 plan-time 의존성 검토:

| Library / Source | Purpose | Decision |
|------------------|---------|----------|
| `golang.org/x/sys/unix` (existing indirect dep) | Advisory lock 재사용 — RT-004의 `lock_unix.go` | **REUSE** — RT-004 머지 후 import |
| `golang.org/x/sys/windows` (existing indirect dep) | Advisory lock 재사용 — RT-004의 `lock_windows.go` | **REUSE** — RT-004 머지 후 import |
| `os.Rename` (stdlib) | Atomic version-file update | **USE** — REQ-013 atomic 보장 |
| `bytes.ReplaceAll` (stdlib) | Migration 1의 literal 치환 | **USE** — single-pass, allocation-free for typical wrapper size |
| `path/filepath.Glob` (stdlib) | Migration 1의 wrapper 파일 enumerate | **USE** — `.claude/hooks/moai/handle-*.sh` 매칭 |
| Anthropic Claude API hook protocol | Reference | Out-of-scope (RT-001 owns) |
| Cobra command framework | CLI surface | **USE** (already in go.mod, pattern from `migrate_agency.go`) |

**No new direct module deps**. 100% standard library + existing indirect deps.

---

## 12. Cross-SPEC dependency status

### 12.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN-zone codification): 마이그레이션 프레임워크는 헌법적 메커니즘이며 zone-registry에 등록되어야 함. CON-001 머지 상태 plan-audit 시점 verify.
- **SPEC-V3R2-RT-001** (HookResponse with SystemMessage): 마이그레이션 실패 surfacing은 SystemMessage 필드 의존. 미머지 시 `slog.Warn` 임시 사용.
- **SPEC-V3R2-RT-006** (SessionStart handler completeness): 핸들러 완성도 보장; RT-007의 `MigrationRunner.Apply` 호출이 RT-006 변경 위에 들어가야 conflict-free.

### 12.2 Blocks

- **SPEC-V3R2-EXT-004** (versioned migration auto-apply general framework): RT-007이 base infrastructure 제공. EXT-004는 v2→v3 migration catalog 추가.
- **SPEC-V3R2-MIG-001** (v2→v3 user migrator): RT-007의 MigrationRunner 패턴 재사용.
- **SPEC-V3R2-MIG-002** (hook registration cleanup): RT-007의 framework 위에 m002, m003 등록.
- **SPEC-V3R2-MIG-003** (config loader 보강): 누락된 yaml 섹션 retrofit 마이그레이션 가능.

### 12.3 Related (non-blocking)

- **SPEC-V3R2-RT-005** (settings-file multi-layer merge): 마이그레이션이 settings 파일 수정 시 provenance 태그 적용.
- **SPEC-V3R2-RT-004** (typed session state): `.moai/state/` 디렉터리 공유, 그러나 staleness 의미 분리 (research.md §7).
- **SPEC-V3R2-CON-003** (consolidation pass): 일부 문서 이동을 마이그레이션으로 emit 가능.
- **SPEC-V3R2-WF-001** (skill consolidation 48→24): 마이그레이션 catalog는 SPEC-V3R2-MIG-001 ownership.

---

End of research.md.
