# SPEC-V3R2-RT-002 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)      | Initial task breakdown — 28 tasks (T-RT002-01..28) across M1-M5         |

---

## Task ID Convention

- ID format: `T-RT002-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- DDD/TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M5 (GREEN/REFACTOR)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase) — Priority P0

목적: research.md §2.1 갭에 대응하는 ~10 개의 실패 테스트 + 1 audit lint 추가. spec-workflow.md TDD 원칙: 실패 테스트 먼저.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT002-01 | `internal/permission/resolver_test.go` 에 `TestResolve_FrontmatterUnknownPermissionMode` 추가 — `permissionMode: ultra-bypass` 같은 값에 대한 sentinel 검증 (frontmatter walker 가 별도; 본 테스트는 resolver-side 동작 — `ParsePermissionMode` 가 invalid 값 receive 시 default 로 fallback 하는지 검증). | expert-backend | `internal/permission/resolver_test.go` (extend, ~25 LOC) | none | 1 file (extend) | RED — current ParsePermissionMode 가 invalid 시 ModeDefault + error 반환 — error path 의 caller 가 default 로 fallback 하는지 단위 검증 |
| T-RT002-02 | `internal/permission/resolver_test.go` 에 `TestResolve_HookUpdatedInputReMatch` 추가 — hook 가 `/dangerous/path` → `/safe/path` 로 mutate 시 pre-allowlist 가 mutated path 에 매칭되는지 검증. AC-10. | expert-backend | same file (extend, ~40 LOC) | T-RT002-01 | 1 file (extend) | RED initially — re-match 무한루프 방지 가드 검증 위해 nested mutation 시도 추가 |
| T-RT002-03 | `internal/permission/resolver_test.go` 에 `TestResolve_ForkDepth4DegradeToBubble` 추가 — depth=4, mode=acceptEdits → systemMessage emit + decision=ask + Origin="fork depth limit". AC-14. | expert-backend | same file (extend, ~30 LOC) | T-RT002-01 | 1 file (extend) | RED — current resolver.go:209-219 의 systemMessage sentinel 이 spec 과 약간 다름 (spec.md AC-14 expected "mode degraded" vs current "bypassPermissions degraded") — M4 에서 통일 |
| T-RT002-04 | `internal/permission/resolver_test.go` 에 `TestResolve_BubbleParentClosed` 추가 — IsFork=true, ParentAvailable=false → deny + "parent unavailable" message. AC-08. | expert-backend | same file (extend, ~25 LOC) | T-RT002-01 | 1 file (extend) | RED — current handleBubbleAsk 의 sentinel 검증; 일부는 현재 GREEN 일 수 있음 |
| T-RT002-05 | `internal/permission/resolver_test.go` 에 `TestResolve_NonInteractiveAskBecomesDeny` + `TestLogUnreachablePrompt_FilePath` 추가 — IsInteractive=false + 매치 실패 → DecisionDeny + log entry at `.moai/logs/permission.log`. AC-15. | expert-backend | same file (extend, ~50 LOC) | T-RT002-01 | 1 file (extend) | RED initially — log path 검증 강화 |
| T-RT002-06 | `internal/permission/resolver_test.go` 에 `TestResolve_ConflictSpecificityThenFsOrder` 추가 — 동일 SrcLocal tier 의 두 매치, specificity 우선 + fs-order 우선. AC-12. | expert-backend | same file (extend, ~40 LOC) | T-RT002-01 | 1 file (extend) | RED — current checkTier 의 conflict tiebreak 부재 → 첫 매치 반환 |
| T-RT002-07 | `internal/permission/resolver_test.go` 에 `TestResolve_PolicyDenyOverridesProjectAllow` 추가 — SrcPolicy deny vs SrcProject allow → policy wins. AC-13. | expert-backend | same file (extend, ~25 LOC) | T-RT002-01 | 1 file (extend) | RED initially — 8-tier walk 의 priority order 검증 |
| T-RT002-08 | `internal/permission/resolver_test.go` 에 `TestResolve_BypassPermissionsRejectedInStrictMode` 추가 — strict_mode=true 에서 ValidateMode 호출 시 PermissionModeRejected 반환. AC-07. | expert-backend | same file (extend, ~25 LOC) | T-RT002-01 | 1 file (extend) | RED initially — 일부 GREEN; 호출 사이트 wire 부재로 인한 통합 테스트 RED |
| T-RT002-09 | `internal/permission/resolver_test.go` 에 `TestResolve_LegacyBypassActionMigrated` 추가 — v2 `bypassPermissions` action 발견 시 acceptEdits 로 reroute + deprecation warning. AC-11. | expert-backend | same file (extend, ~35 LOC) | T-RT002-01 | 1 file (extend) | RED — MigrateLegacyBypassRules 함수 부재 |
| T-RT002-10 | `internal/permission/resolver_test.go` 에 `TestResolve_SessionRulesLoadedAsSrcSession` 추가 — REQ-030 의 session_rules 키가 SrcSession tier 로 적재되는지 검증. | expert-backend | same file (extend, ~30 LOC) | T-RT002-01 | 1 file (extend) | RED — RT-005 reader 미머지 시 hardcoded fixture |
| T-RT002-11 | `internal/cli/doctor_permission_test.go` 신규 — `TestDoctorPermission_AllTiersFlag`, `TestDoctorPermission_TraceJSONFormat`, `TestDoctorPermission_DryRun`, `TestDoctorPermission_NoMatchTrace`. AC-05. | expert-backend | new file (~150 LOC) | T-RT002-01 | 1 file (create) | RED — `--all-tiers/--mode/--fork/--format` 플래그 부재 |
| T-RT002-12 | `internal/template/agent_frontmatter_audit_test.go` 확장 — `TestAgentFrontmatter_PermissionModeStrictEnum` 추가. `.claude/agents/**/*.md` 의 permissionMode 키 5-enum strict-validation. AC-09. Sentinel: `PERMISSION_MODE_UNKNOWN_VALUE: <file> declares permissionMode: <value>; allowed: default|acceptEdits|bypassPermissions|plan|bubble.` | expert-backend | existing file (extend, ~80 LOC) | T-RT002-01 | 1 file (extend) | RED — walker 의 permissionMode 검증 분기 부재 |
| T-RT002-13 | `go test ./internal/permission/ ./internal/cli/ ./internal/template/` 실행. 새 테스트 ≥10 RED + 기존 테스트 GREEN baseline 유지 검증. | manager-tdd | n/a (verification only) | T-RT002-01..12 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. TDD discipline.

T-RT002-01..10 은 모두 같은 `resolver_test.go` 를 extend → sequential 실행 권장 (parallel 시 merge conflict). T-RT002-11 은 신규 파일 → independent. T-RT002-12 도 별도 파일 → independent.

---

## M2: PreAllowlist sync.Once + ValidateMode wire + frontmatter strict lint + security.yaml (GREEN, part 1) — Priority P0

목적: 5-enum strict validation, pre-allowlist hot-path 최적화, bypassPermissions strict_mode reject, security.yaml 신규 키. AC-07, AC-09 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT002-14 | `internal/permission/stack.go:182-233` 의 `PreAllowlist()` 를 sync.Once + cached slice 패턴으로 변환. 첫 호출 시에만 8-rule slice 빌드. | expert-backend | `internal/permission/stack.go` (extend, ~15 LOC) | T-RT002-13 | 1 file (edit) | GREEN — REFACTOR (hot path 최적화) |
| T-RT002-15 | `internal/permission/spawn.go` 신규 — `RejectIfStrict(mode PermissionMode, strictMode bool) error` 함수. ModeBypassPermissions + strictMode → `permission mode rejected: bypassPermissions not allowed in strict mode`. 다른 mode 또는 strictMode=false → nil. | expert-backend | new file (~30 LOC) | T-RT002-13 | 1 file (create) | GREEN — AC-07 |
| T-RT002-16 | `internal/permission/spawn_test.go` 신규 — `TestRejectIfStrict_RejectsBypass`, `TestRejectIfStrict_AllowsAcceptEdits`, `TestRejectIfStrict_StrictModeFalse`. | expert-backend | new file (~50 LOC) | T-RT002-15 | 1 file (create) | GREEN — AC-07 단위 |
| T-RT002-17 | `internal/template/agent_frontmatter_audit_test.go` 확장 (T-RT002-12 의 walker 함수 구현 — RED → GREEN 전환). frontmatter delimiter `---` 사이 파싱 + permissionMode 키 5-enum 검증. | expert-backend | existing file (~80 LOC 추가) | T-RT002-13 | 1 file (edit) | GREEN — AC-09 |
| T-RT002-18 | `.moai/config/sections/security.yaml` 확장 — `permission.strict_mode: false`, `permission.pre_allowlist: [<8 패턴>]`, `permission.session_rules: []` 추가. + `internal/template/templates/.moai/config/sections/security.yaml` 미러 동시 업데이트. | expert-backend | 2 files (edit, ~20 LOC each) | T-RT002-13 | 2 files (edit) | GREEN — Template-First 정합성 (CLAUDE.local.md §2) |

**M2 priority: P0** — blocks M3-M5. AC-07, AC-09 turn GREEN.

T-RT002-14, T-RT002-15+16 (paired), T-RT002-17, T-RT002-18 은 독립 파일을 건드리므로 parallel 가능.

---

## M3: hook UpdatedInput 가드 + IsWriteOperation 보강 + conflict tiebreak (GREEN, part 2) — Priority P0

목적: hook 재진입 방지, write 패턴 정밀화, 동일 tier 충돌 해소. AC-04, AC-10, AC-12, AC-13 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT002-19 | `internal/permission/resolver.go:147-160` 의 hook UpdatedInput re-match 가드 검증 — 단일 재실행 enforcement. 코멘트로 invariant 명시 ("// Single re-execution; nested mutation rejected by HookResponse=nil clear"). | expert-backend | `internal/permission/resolver.go` (extend, ~5 LOC + comment) | T-RT002-18 | 1 file (edit) | GREEN — AC-10 |
| T-RT002-20 | `internal/permission/stack.go:244-269` 의 `IsWriteOperation` write 패턴 보강:<br>- `mv ` 중복 항목 제거 (line 252).<br>- 추가 패턴: `git reset --hard`, `git clean`, `npm run build`, `make install`, `dd of=`.<br>- `echo ` 매처를 `strings.HasPrefix(strings.TrimSpace(input), "echo ")` 로 정밀화.<br>- `cat >` 등 redirect 매처 동일 정밀화. | expert-backend | `internal/permission/stack.go` (extend, ~25 LOC) | T-RT002-18 | 1 file (edit) | GREEN — AC-06 의 plan-mode write 분기 정합성 |
| T-RT002-21 | `internal/permission/conflict.go` 신규 — `resolveConflict(rules []*PermissionRule, tool, input string) *PermissionRule` 함수. specificity 점수 (와일드카드 개수의 역수) 우선, 동률 시 Origin lexicographic later 우선. specificity 함수 + tiebreak + log emit (`.moai/logs/permission.log`) 모두 포함. | expert-backend | new file (~80 LOC) | T-RT002-18 | 1 file (create) | GREEN — AC-12 |
| T-RT002-22 | `internal/permission/conflict_test.go` 신규 — `TestResolveConflict_SpecificityWins`, `TestResolveConflict_FsOrderTiebreak`, `TestResolveConflict_SingleMatchNoLog`, `TestResolveConflict_LogPath`. | expert-backend | new file (~120 LOC) | T-RT002-21 | 1 file (create) | GREEN — AC-12 단위 |
| T-RT002-23 | `internal/permission/resolver.go:258-299` 의 `checkTier` 변경 — 동일 tier 에서 ≥2 매치 발생 시 `resolveConflict` 호출 분기 추가. (현재 첫 매치 즉시 반환 → ≥2 매치 누적 후 conflict 평가로 변경). | expert-backend | `internal/permission/resolver.go` (extend, ~20 LOC) | T-RT002-21 | 1 file (edit) | GREEN — AC-12 통합 |

**M3 priority: P0** — blocks M4. AC-04, AC-10, AC-12, AC-13 turn GREEN.

T-RT002-19, T-RT002-20, T-RT002-21+22 (paired), T-RT002-23 은 독립적이지만 T-RT002-23 은 T-RT002-21 (resolveConflict 함수 정의) 에 dependent.

---

## M4: bubble dispatch IPC contract + fork depth degrade + non-interactive log (GREEN, part 3) — Priority P0

목적: bubble routing API contract 동결, fork depth >3 sentinel 정렬, non-interactive 로그 기록 정합성. AC-08, AC-14, AC-15 GREEN.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT002-24 | `internal/permission/bubble.go:94-102` 의 `DispatchToParent` placeholder 격상 — godoc 에 contract 명시 (RT-001 hook channel 재사용; stdin JSON 으로 BubbleRequest, stdout JSON 으로 BubbleResponse). 본 SPEC 은 contract 만 동결, 실제 IPC wire 는 orchestrator integration (skill body) 의 영역으로 명시. placeholder return 은 그대로 유지. | expert-backend | `internal/permission/bubble.go` (godoc 보강, ~10 LOC) | T-RT002-23 | 1 file (edit) | GREEN — contract freezing |
| T-RT002-25 | `internal/permission/resolver.go:209-219` 의 fork depth >3 게이트 systemMessage sentinel 정렬 — current "Fork depth N exceeds limit - mode degraded to bubble" 로 통일 (handleBypassInFork 의 메시지도 동일 패턴 적용). AC-14. | expert-backend | `internal/permission/resolver.go` (extend, ~5 LOC + 동일 sentinel 통일) | T-RT002-24 | 1 file (edit) | GREEN — AC-14 sentinel 정합성 |
| T-RT002-26 | `internal/permission/resolver.go:241-247` 의 non-interactive fail-closed 분기 검증 — IsInteractive=false 시 default decision 이 DecisionDeny 로 변환 + `logUnreachablePrompt` 호출 path 가 `.moai/logs/permission.log` 와 정확히 일치. 추가 단위 테스트로 검증 + log entry format 정렬 ("[<ISO 8601>] Unreachable prompt: tool=<tool> input=<truncated>"). AC-15. | expert-backend | `internal/permission/resolver.go` (extend, ~10 LOC) + `resolver_test.go` extend | T-RT002-24 | 2 files (edit) | GREEN — AC-15 정합성 |
| T-RT002-27 | `internal/permission/bubble.go:150-154` 의 `IsParentAvailable` placeholder 보강 — godoc 에 "registry lookup contract" 명시. 본 SPEC 의 구현은 ParentSessionID 비어있지 않음 으로 단순화 (registry 통합은 RT-004 SessionStore 와 합류 시 wire). | expert-backend | `internal/permission/bubble.go` (godoc, ~5 LOC) | T-RT002-24 | 1 file (edit) | GREEN — contract |

**M4 priority: P0** — blocks M5. AC-08, AC-14, AC-15 turn GREEN.

T-RT002-24, T-RT002-25, T-RT002-26, T-RT002-27 은 sequential within `bubble.go`/`resolver.go`; merge 충돌 회피 위해 순서대로 실행.

---

## M5: doctor_permission CLI 보강 + Legacy migration + Session rules + CHANGELOG + MX tags (GREEN, part 4 + REFACTOR + Trackable) — Priority P1

목적: 사용자 노출 surface, anti-regression CI lint, 문서화, MX tag 삽입.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT002-28 | `internal/cli/doctor_permission.go` 보강 — `--all-tiers`, `--mode <m>`, `--fork`, `--format <human\|json>` 플래그 추가. 기존 `--tool/--input/--trace/--dry-run` 유지. `result.ExportTrace()` 출력의 hook tier sentinel (`config.Source(999)`) 을 `"tier": "hook"` 으로 stringify. AC-05. | expert-backend | `internal/cli/doctor_permission.go` (extend, ~80 LOC) | T-RT002-27 | 1 file (edit) | GREEN — AC-05 |
| T-RT002-29 | `internal/permission/migration.go` 신규 — `MigrateLegacyBypassRules(rules []PermissionRule) ([]PermissionRule, []string)` 함수. legacy `Action == "bypassPermissions"` rule 발견 시 `Action: DecisionAllow` 로 reroute + deprecation warning slice 반환 (Origin file path 명시). 호출자가 stderr WARN + `.moai/logs/permission.log` 양쪽에 emit. AC-11. | expert-backend | new file (~50 LOC) | T-RT002-28 | 1 file (create) | GREEN — AC-11 |
| T-RT002-30 | `internal/permission/migration_test.go` 신규 — `TestMigrateLegacyBypassRules_HappyPath`, `TestMigrateLegacyBypassRules_NoLegacy`, `TestMigrateLegacyBypassRules_MultipleOrigins`. | expert-backend | new file (~80 LOC) | T-RT002-29 | 1 file (create) | GREEN — AC-11 단위 |
| T-RT002-31 | MX tags 삽입 per `plan.md` §6: 3 `@MX:ANCHOR` (resolver.go::Resolve, stack.go::PreAllowlist, bubble.go::DispatchToParent) + 2 `@MX:NOTE` (stack.go::PermissionMode, conflict.go::resolveConflict) + 2 `@MX:WARN` (resolver.go:147-160 hook re-match, resolver.go:241-247 non-interactive). 5 distinct files. | manager-docs | 5 files (edit, MX tag insertion only) | T-RT002-28..30 | 5 files (edit) | Trackable — plan §6 |
| T-RT002-32 | CHANGELOG entry 추가 — `## [Unreleased] / ### 추가` 아래 한국어 entry per plan.md §M5d 텍스트. | manager-docs | `CHANGELOG.md` (extend) | T-RT002-31 | 1 file (edit) | Trackable |
| T-RT002-33 | 워크트리 루트에서 `make build` 실행 — `internal/template/embedded.go` 재생성. security.yaml mirror 변경이 embed 에 반영되었는지 diff 검증. | manager-docs | `internal/template/embedded.go` (regenerated) | T-RT002-32 | 1 file (regenerated) | Build verification |
| T-RT002-34 | 워크트리 루트에서 `go test ./...` 전체 실행. ALL GREEN + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule). 추가로 `go vet ./...`, `golangci-lint run` 0 warnings. | manager-tdd | n/a (verification only) | T-RT002-33 | 0 files | GREEN gate (final) |
| T-RT002-35 | `progress.md` 의 `run_complete_at: <timestamp>` + `run_status: implementation-complete` 갱신. | manager-docs | `progress.md` (extend) | T-RT002-34 | 1 file (edit) | Trackable closure |

**M5 priority: P1** — completes the SPEC. AC-05, AC-11 turn GREEN.

T-RT002-28, T-RT002-29+30 (paired) 은 parallel 가능. T-RT002-31 은 모든 source code 변경 land 후 (MX tag 위치는 final code structure 기준).

---

## Task summary by milestone

| Milestone | Task IDs | Total tasks | Priority | Owner role mix |
|-----------|----------|-------------|----------|----------------|
| M1 (RED) | T-RT002-01..13 | 13 | P0 | expert-backend (12) + manager-tdd (1 verification) |
| M2 (GREEN part 1) | T-RT002-14..18 | 5 | P0 | expert-backend |
| M3 (GREEN part 2) | T-RT002-19..23 | 5 | P0 | expert-backend |
| M4 (GREEN part 3) | T-RT002-24..27 | 4 | P0 | expert-backend |
| M5 (GREEN part 4 + REFACTOR + Trackable) | T-RT002-28..35 | 8 | P1 | expert-backend (3) + manager-docs (4) + manager-tdd (1 verification) |
| **TOTAL** | T-RT002-01..35 | **35 tasks** | — | — |

> NOTE: 35 tasks span the 5 milestones. M1 (RED) opens with 13 test-only tasks; M2-M4 (GREEN core) deliver the typed-permission subsystem in 14 tasks; M5 (final GREEN + REFACTOR + Trackable) closes with 8 tasks of CLI, migration, MX tags, and trackability work.

---

## Dependency graph (critical path)

```
T-RT002-01..12 (M1 tests, sequential within shared file resolver_test.go)
   ↓
T-RT002-13 (M1 verification gate — RED 확인)
   ↓
T-RT002-14, 15+16, 17, 18 (M2 parallel)
   ↓
T-RT002-19, 20 (M3 parallel)
   ↓
T-RT002-21+22 (conflict.go + test) → T-RT002-23 (checkTier wire)
   ↓
T-RT002-24 (bubble dispatch contract)
   ↓
T-RT002-25, 26, 27 (M4 sequential within bubble.go/resolver.go)
   ↓
T-RT002-28 (doctor CLI 보강)
   ↓
T-RT002-29+30 (migration.go + test)
   ↓
T-RT002-31 (MX tags) → T-RT002-32 (CHANGELOG) → T-RT002-33 (make build) → T-RT002-34 (go test ./... + vet + lint) → T-RT002-35 (progress.md closure)
```

Critical path: 11 sequential gates from T-RT002-13 → T-RT002-35 (M1 → M5 closure).

---

## Cross-task constraints

[HARD] All file edits use absolute paths under the worktree root `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002` per CLAUDE.md §Worktree Isolation Rules.

[HARD] All tests use `t.TempDir()` per `CLAUDE.local.md` §6 — no test creates files in the project root.

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; no `filepath.Join(cwd, absPath)` patterns per `CLAUDE.local.md` §6.

[HARD] No new direct module dependencies — `internal/permission/` 의존성은 stdlib + 기존 `internal/config` + 기존 `internal/hook` + 기존 `github.com/spf13/cobra` (CLI 만) 으로 한정.

[HARD] `internal/template/templates/.moai/config/sections/security.yaml` 미러를 동일 변경으로 유지 (Template-First Rule per CLAUDE.local.md §2).

[HARD] 코드 주석은 한국어 (per `.moai/config/sections/language.yaml` `code_comments: ko`). godoc 와 exported identifier 의 영문 docstring 은 industry standard 로 영어 유지.

[HARD] commit messages 는 한국어 (per `.moai/config/sections/language.yaml` `git_commit_messages: ko`).

[HARD] 본 SPEC 은 `.claude/agents/**/*.md` 의 frontmatter `permissionMode` 키를 *읽기만* 검증; 미선언 agent 에 강제 추가하지 않음 (그것은 SPEC-V3R2-ORC-001 의 영역).

[HARD] `internal/cli/run.go`, `internal/hook/*.go`, `internal/session/*.go`, `internal/config/source.go` 는 본 SPEC 이 직접 수정 금지 — cross-SPEC scope 보호. spawn-side wire 는 `internal/permission/spawn.go` 의 helper 만 노출하고, 호출 사이트 추가는 다른 SPEC 의 영역.

---

End of tasks.md.
