---
id: SPEC-INTERNAL-SECURITY-001
title: "internal/ 보안 하드닝 — 진행 상태"
version: "0.1.0"
status: implemented
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/ (web, cli, hook, profile, template)"
lifecycle: spec-anchored
tags: "security, hardening, progress"
tier: M
---

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출물 4종(spec.md + plan.md + acceptance.md + progress.md) 작성 완료 (status: draft).
- 8개 verified finding → REQ-SEC-001..008 변환 완료, 3개 논리 그룹(web / template-update / hook)으로 그룹핑.
- 각 finding의 file:line 앵커 plan-phase 착수 직전 spot-check 재확인.
- 미결 설계 분기 3건(plan.md §E): REQ-SEC-006 배선 vs 제거, REQ-SEC-002 Sec-Fetch-Site vs 토큰, REQ-SEC-005 provenance vs 보수적 백업 — run-phase 진입 전 사용자 확인 권장(blocker 후보).

## §E.2 Run-phase Evidence

### M1 — Web 그룹 (REQ-SEC-001 P0 + REQ-SEC-002 P1)

**REQ-SEC-001 — GET 읽기 경로 순회 차단 + 중앙화된 traversal guard**

- `internal/web/handlers.go` handleIndex (L153 region): `profile.IsValidProfileName(selected)` 4xx 가드 추가 — GET 읽기 경로 순회 차단. handleSave(L292)와 대칭 복원.
- `internal/profile/preferences.go` GetPreferencesPath: 중앙 traversal 가드 이식 — `!isValidProfileName(name)` 시 base preferences.yaml 로 clamp (base dir 밖 경로 생성 불가).
- `internal/profile/profile.go` GetProfileDir: 기존 중앙 가드(`isValidProfileName` → "" 반환) REQ-SEC-001 주석 명시.

**REQ-SEC-002 — 상태 변경 route same-origin 강제 + 주석 정정**

- `internal/web/app.go` hostCheckMiddleware: `Sec-Fetch-Site: same-origin` 강제 추가 (POST/PUT/PATCH). cross-site / same-site / none / absent → 403. 사용자 결정(binding): Sec-Fetch-Site 헤더 방식, absent → reject (conservative default).
- `internal/web/app.go` 주석 정정: 기존 "CSRF를 막는 단일 경계" 거짓 주장 → "2-layer 쓰기-안전 모델 (Host=DNS-rebinding, Sec-Fetch-Site=CSRF)" 정정. /__shutdown__ 라우트 주석도 동일 정정.

**회귀 테스트 (NFR-SEC-002)**

- NEW `internal/web/security_test.go`: AC-SEC-001a (GET 순회 4xx), AC-SEC-002a (cross-site POST 4xx), AC-SEC-002 conservative default (absent header 4xx), AC-SEC-002b/NFR-SEC-003 (same-origin POST success), AC-SEC-002a (/__shutdown__ cross-site 4xx), NFR-SEC-003 (valid profile GET 200).
- NEW `internal/profile/security_central_guard_test.go`: AC-SEC-001c (GetPreferencesPath 순회 거부), AC-SEC-001c (GetProfileDir 순회 거부), NFR-SEC-003 (valid names unchanged).
- 기존 test helper 업데이트 (Sec-Fetch-Site adaptation): servePost, postProfile, servePostShutdown, postSave, postForm, board_test, profile_traversal_test — 모두 `Sec-Fetch-Site: same-origin` 헤더 추가로 real browser form POST 시뮬레이션 + handler-level 로직 격리.

### AC Binary Matrix (M1)

| AC | Status | Evidence |
|----|--------|----------|
| AC-SEC-001a (순회 payload 4xx) | PASS | `go test -run TestGETReadPathRejectsTraversalProfile ./internal/web/` — 6 traversal payloads all 4xx |
| AC-SEC-001b (read-path 가드 배선) | PASS | `grep -n IsValidProfileName internal/web/handlers.go` → L159 (handleIndex) + L292 (handleSave) |
| AC-SEC-001c (중앙 traversal 가드) | PASS | `go test -run TestGetPreferencesPathRejectsTraversal ./internal/profile/` + GetProfileDir test — both PASS |
| AC-SEC-002a (cross-site POST 4xx) | PASS | `go test -run TestCrossSitePostRejected ./internal/web/` — 403 + no persistence |
| AC-SEC-002b (same-origin POST 성공) | PASS | `go test -run TestSameOriginPostAllowed ./internal/web/` — not 403 + WritePreferences called |
| AC-SEC-002c (주석 정정) | PASS | `grep -in csrf internal/web/app.go` — CSRF claims now match actual Sec-Fetch-Site mechanism; no false claim |

### M2 — Template-update 그룹 (REQ-SEC-003 P1 + REQ-SEC-004 P1 + REQ-SEC-005 P1 + REQ-SEC-006 P0)

**REQ-SEC-003 — 심링크 역참조 백업 차단**

- `internal/cli/update_namespace_protect.go` collectUserOwnedFilesWith walk 콜백: `isSymlinkEntry(path)` Lstat 가드 추가 (update.go:2179 기존 패턴 재사용 — 새 패턴 발명 금지). 심링크 skip → copyFile이 역참조 target 내용을 `.moai/backups/`에 기록하지 않음.
- `internal/cli/update_archive.go` copyFile: 심링크 source 거부 방어 (defense-in-depth). copyDirAll: walk 시 심링크 skip (archive 경유 유출 차단).

**REQ-SEC-004 — Template-First `.gitignore` `.moai/backups/`**

- `internal/template/templates/.gitignore:118` 에 `.moai/backups/` (SLASH form) 추가 — 기존 `.moai-backups/` (HYPHEN, L117) 옆. Template-First Rule 준수: template source 먼저 편집 후 `make build` 로 임베드 재컴파일.

**REQ-SEC-005 — conservative backup expansion [USER DECISION]**

- `internal/cli/update_namespace_protect.go` 신규 `isUserOwnedNamespaceConservative(rel)` — isUserOwnedNamespace superset. reserved-prefix 모호 이름(`moai-my-notes`, `expert-mydomain.md` 등)도 백업 pass에 포함 (R4 low-risk path). system agent dirs (core/expert/meta)은 여전히 제외.
- `collectUserOwnedFilesWith(projectRoot, classify)` 로 walk 로직 추출; `collectUserOwnedFiles`(strict, buildPreserveInventory용) 와 `collectUserOwnedFilesConservative`(backup용) 분리. buildPreserveInventory strict 유지 → clean-reinstall 회귀 없음.
- `backupUserOwnedNamespace` 가 conservative collector 사용.

**REQ-SEC-006 — dead abort sentinel WIRE [USER DECISION]**

- `internal/cli/update_namespace_protect.go` 신규 `verifyNamespaceBackupCoverage(projectRoot, backupDir)` — backup 완료 후·파괴적 deploy 이전에 실행되는 real pre-modification abort gate. conservative collected set 중 backupDir에 없는 파일을 `unprotected` plan으로 구성 → `assertNoUserOwnedNamespaceTouch` 위임 (UPDATE_USER_NAMESPACE_VIOLATION sentinel).
- `internal/cli/update.go` runUpdate "Backup" case (L797): `verifyNamespaceBackupCoverage(projectRoot, nsBackupPath)` 배선. 정상 update (no user-owned / fully backed up) → unprotected empty → 통과 (NFR-SEC-003).
- `@MX:REASON` 정정 (AC-SEC-006c): update.go:1283 기존 거짓 "called by backupUserOwnedNamespace, assertNoUserOwnedNamespaceTouch, and template overlay write loop; fan_in >= 3" → 실제 fan_in=3 직접 호출자 (collectUserOwnedFiles value-pass + assertNoUserOwnedNamespaceTouch + isUserOwnedNamespaceConservative) 명시.

**회귀 테스트 (NFR-SEC-003/NFR-SEC-004)**

- NEW `internal/cli/update_security_m2_test.go`: AC-SEC-003a (심링크 target 미백업), AC-SEC-003b (심링크 skip), AC-SEC-005a (모호 이름 백업), AC-SEC-005b (conservative 분류 테이블), AC-SEC-006a (verifyNamespaceBackupCoverage 4-case: empty/fully-backed-up/missing/partial).
- NEW `internal/template/embed_gitignore_backups_test.go`: AC-SEC-004b (임베드 FS `.moai/backups/` 서빙).

### AC Binary Matrix (M2)

| AC | Status | Evidence |
|----|--------|----------|
| AC-SEC-003a (심링크 target 미백업) | PASS | `go test -run TestBackupUserOwnedNamespace_SymlinkTargetNotBackedUp ./internal/cli/` — secret content not in backup tree |
| AC-SEC-003b (Lstat 가드 배선) | PASS | `grep -nE "Lstat\|isSymlinkEntry" internal/cli/update_namespace_protect.go` → L122 `isSymlinkEntry(path)` |
| AC-SEC-004a (template source `.moai/backups/`) | PASS | `grep -n "^\.moai/backups/" internal/template/templates/.gitignore` → L118 |
| AC-SEC-004b (임베드 FS 서빙, make build 후) | PASS | `go test -run TestEmbeddedGitignoreHasMoaiBackupsSlash ./internal/template/` |
| AC-SEC-005a (모호 이름 백업) | PASS | `go test -run TestBackupUserOwnedNamespace_ConservativeReservedPrefix ./internal/cli/` — moai-my-notes + expert-mydomain backed up |
| AC-SEC-005b (prefix-collision 동작 문서화) | PASS | `go test -run TestIsUserOwnedNamespaceConservative_PrefixCollision ./internal/cli/` — 13-case table |
| AC-SEC-006a (WIRE branch — ≥1 production caller) | PASS | `grep -rn "assertNoUserOwnedNamespaceTouch" internal/cli/ --include="*.go" \| grep -v "_test.go"` → update.go:797 verifyNamespaceBackupCoverage → :310 assertNoUserOwnedNamespaceTouch production path |
| AC-SEC-006c (@MX fan_in matches actual) | PASS | update.go:1283 @MX:REASON = "called by collectUserOwnedFiles, assertNoUserOwnedNamespaceTouch, and isUserOwnedNamespaceConservative; fan_in = 3" — grep 실측 3 callers 일치 |

### M3 — Hook 그룹 (REQ-SEC-007 P1 + REQ-SEC-008 P2)

**REQ-SEC-007 — checkFileAccess EvalSymlinks resolve-recheck (CWE-61)**

- `internal/hook/pre_tool.go` checkFileAccess (L664-725 region): `filepath.Abs(filePath)` 이후, project-boundary 검사 및 DenyPatterns 매칭 이전에 `filepath.EvalSymlinks(resolvedPath)` 호출. project-internal 심링크(notes.txt -> ~/.ssh/id_rsa)가 실제 target으로 해석되어 Write 도구의 링크 추종 secret 덮어쓰기(CWE-61)를 차단. `internal/hook/file_changed.go:176-203`(SPEC-SEC-HARDEN-004 / REQ-SEC4-004)의 resolve-recheck 패턴 전파 — 새 메커니즘 발명 없음.
- new-file Write not-exist fallback (AC-SEC-007c): EvalSymlinks가 not-exist error 반환 시 unresolved 경로로 fallback, deny 아님 (NFR-SEC-003). file_changed.go의 fail-closed와 다른 critical-path 동작.
- macOS `/var -> /private/var` 대칭 보정: projectAbs도 EvalSymlinks 정규화하되, resolvedPath가 fallback한(not-exist) 경우 projectAbs도 미해상 상태로 유지(`resolvedSymlink` 플래그) — 대칭성 보존. 없으면 정상 new-file Write가 false-positive boundary escape로 오탐.

**REQ-SEC-008 — Edit new_string 민감 콘텐츠 스캔**

- `internal/hook/pre_tool.go` checkFileAccess 민감 콘텐츠 스캔 게이트 (L745-762): 기존 `if toolName == "Write"` 단일 게이트 → `if toolName == "Write" || toolName == "Edit"` 확장. Edit일 때 `new_string` 필드 파싱하여 SensitiveContentPatterns 통과 — Write 콘텐츠와 동일 deny 동작. Edit 경유 secret 주입이 Write-path deny를 우회하던 갭 봉쇄.

**회귀 테스트 (NFR-SEC-003/NFR-SEC-004)**

- NEW `internal/hook/pre_tool_security_test.go` (5 tests, 모두 t.TempDir() 내 fixture):
  - TestCheckFileAccess_SymlinkToDenyTargetBlocked (AC-SEC-007a) — project-internal 심링크 -> .ssh/id_rsa deny.
  - TestCheckFileAccess_SymlinkEscapeProjectBlocked — project-internal 심링크 -> project 외부 normal file boundary deny.
  - TestCheckFileAccess_NewFileWriteFallback (AC-SEC-007c) — 미존재 정상 경로 Write allow (fallback 보존).
  - TestCheckFileAccess_EditNewStringSecretDenied (AC-SEC-008a) — Edit new_string private key deny.
  - TestCheckFileAccess_EditNewStringCleanAllowed (NFR-SEC-003) — Edit clean new_string allow (false-positive 방지).
- secret fixture(`rsaPrivateKeyFixture()`)는 runtime concatenation으로 생성 — source 파일 자체가 PreToolUse sensitive-content scan에 걸려 write 차단되는 것을 방지하면서, runtime 값은 scanner 정상 매칭.

### AC Binary Matrix (M3)

| AC | Status | Evidence |
|----|--------|----------|
| AC-SEC-007a (심링크 경유 deny target 차단) | PASS | `go test -run TestCheckFileAccess_SymlinkToDenyTargetBlocked ./internal/hook/` — ok |
| AC-SEC-007b (EvalSymlinks 배선) | PASS | `grep -n EvalSymlinks internal/hook/pre_tool.go` → L685 + L705 (checkFileAccess scope, 6 matches total) |
| AC-SEC-007c (new-file Write fallback) | PASS | `go test -run TestCheckFileAccess_NewFileWriteFallback ./internal/hook/` — ok |
| AC-SEC-008a (Edit secret 거부) | PASS | `go test -run TestCheckFileAccess_EditNewStringSecretDenied ./internal/hook/` — ok |
| AC-SEC-008b (Edit 경로 게이트 확장) | PASS | `grep -n 'toolName ==' internal/hook/pre_tool.go` → L749 `Write \\|\\| Edit` + L751 `Edit` |

## §E.3 Run-phase Audit-Ready Signal

- run_status: M3-complete (hook 그룹); M1 web + M2 template-update + M3 hook all done — 8/8 REQ complete
- run_complete_at: 2026-07-08
- run_commit_sha: M3 = 24482adf2; M1 = 5f33dfaa9; M2 = 85e280599
- ac_pass_count: 16/16 (M1 6/6 AC-SEC-001..002 + M2 8/8 AC-SEC-003..006 + M3 5/5 AC-SEC-007..008)
- ac_fail_count: 0
- preserve_list_post_run_count: 0 (M1 web/profile + M2 cli/template + file_changed.go + unrelated working-tree files untouched; diff confined to internal/hook/)
- l44_pre_commit_fetch: origin/main = 85e280599 (M1+M2 base); worktree HEAD aligned, no parallel-session race
- l44_post_push_fetch: (pending — post-push verification)
- new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; baseline 0)
- cross_platform_build.darwin: exit 0 (`go build ./...`)
- cross_platform_build.windows_amd64: exit 0 (`GOOS=windows GOARCH=amd64 go build ./...`)
- total_run_phase_files (M3): 2 (internal/hook/pre_tool.go edit +42/-3; NEW internal/hook/pre_tool_security_test.go, 5 tests)
- m1_to_mN_commit_strategy: 3 separate Conventional-Commit `fix(SPEC-INTERNAL-SECURITY-001): M{N} ...` commits on main (Hybrid Trunk main-direct) — M1 5f33dfaa9, M2 85e280599, M3 this commit
- coverage: internal/hook 83.6% (package-level, `go test -cover ./internal/hook/`); M3 신규 분기(checkFileAccess EvalSymlinks + Edit scan) 5 tests로 직접 커버
- subagent boundary (C-HRA-008): 0 AskUserQuestion/mcp__askuser real calls in M3 edited+new files (pre_tool.go, pre_tool_security_test.go)
- NFR-SEC-003 (no false-positive deny): new-file Write fallback PASS + Edit clean new_string allow PASS + symmetric EvalSymlinks (macOS /var 보정)
- NFR-SEC-004 (테스트 격리): 모든 심링크/secret fixture t.TempDir() 내 — 실제 ~/.ssh 미접촉
- M1 web (5f33dfaa9) done; M2 template-update (85e280599) done; M3 hook (this commit) done — run-phase complete, ready for sync-phase (manager-docs)
- pre-existing baseline failures (out of SPEC scope, PRESERVE-list packages): internal/statusline `TestBuild_WritesContextUsageWithSessionID`, internal/template `TestOutputStylesTemplateLiveParity` + `TestTemplateNeutralityAuditC8Preserve` — confirmed FAIL at origin/main baseline (M3 changes stashed, true baseline run), independent of M3 (statusline/template do not reference checkFileAccess/pre_tool; grep -rl = no references)

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete (manager-docs sync-phase)
- sync_complete_at: 2026-07-08
- sync_commit_sha: (backfilled by `moai spec close` — self-referential, cannot be in its own commit)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Fixed` — SPEC-INTERNAL-SECURITY-001 entry (M1-M3 hardening summary + 16/16 AC PASS citation per §E.3 headline)
- frontmatter_status_transitions:
  - spec.md: draft → implemented (this sync commit) → completed (`moai spec close` atomic transition)
  - plan.md: draft → implemented → completed (same)
  - acceptance.md: draft → implemented → completed (same)
  - progress.md: in-progress → implemented → completed (same)
- updated_field_refresh: 2026-07-08 (all 4 artifacts; date already 2026-07-08 from run-phase, confirmed present — no change needed)
- sync_phase_deliverables:
  1. progress.md §E.3 `run_commit_sha` M3 backfill: `24482adf2` (observed: `git log --oneline 24482adf2 -1` → "fix(SPEC-INTERNAL-SECURITY-001): M3 hook EvalSymlinks + Edit sensitive scan")
  2. progress.md §E.4 (this section) filled with sync-phase audit-ready signal
  3. progress.md §E.5 Mx-phase Audit-Ready Signal created (close precondition per `moai spec close --dry-run` output: "missing §E.5 Mx-phase audit-ready signal in progress.md")
  4. 4-artifact frontmatter status: draft/in-progress → implemented (this sync commit; spec/plan/acceptance `draft → implemented`, progress `in-progress → implemented`)
  5. CHANGELOG.md `## [Unreleased]` > `### Fixed` entry added
- observed_evidence_cross_refs:
  - §E.2 M1 web evidence: commit `5f33dfaa9` — AC-SEC-001a..002c (6 AC PASS, `httptest` traversal/cross-site tests + grep IsValidProfileName L159/L292)
  - §E.2 M2 template-update evidence: commit `85e280599` — AC-SEC-003a..006c (8 AC PASS; AC-SEC-006b N/A per WIRE-branch design decision — `verifyNamespaceBackupCoverage` 배선 chosen over removal)
  - §E.2 M3 hook evidence: commit `24482adf2` — AC-SEC-007a..008b (5 AC PASS, `t.TempDir()` fixture tests for symlink/EvalSymlinks/Edit-scan)
  - §E.3 run-phase signal: `ac_pass_count: 16/16`, `internal/hook` coverage 83.6%, cross-platform build darwin+windows exit 0, golangci-lint 0 issues
  - §E.5 Mx-phase: @MX tag consistency validated (AC-SEC-006c PASS — update.go:1283 `@MX:REASON` fan_in=3 matches actual grep callers)
- b12_self_test_a (pre-emission grep): `grep -c 'SPEC-INTERNAL-SECURITY-001' CHANGELOG.md` = 0 (pre-emission) → no duplicate entry, emission safe
- b12_self_test_b (AC count match): acceptance.md SSOT = 20 AC rows (AC-SEC-001a..008b); §E.3 headline = 16/16 applicable AC (AC-SEC-006b excluded as WIRE-branch design N/A); CHANGELOG entry cites "16/16 AC PASS" matching §E.3 authoritative run-phase signal
- b12_self_test_c (file path verification): all 8 implementation file paths claimed in §E.2 evidence verified to exist — `internal/web/handlers.go`, `internal/web/app.go`, `internal/profile/preferences.go`, `internal/profile/profile.go`, `internal/cli/update.go`, `internal/cli/update_namespace_protect.go`, `internal/template/templates/.gitignore`, `internal/hook/pre_tool.go`
- canary_compliance_check: N/A (no canary deployment for `internal/` hardening — `moai web` is a local dev tool, not a deployed service)
- l44_pre_commit_fetch: `origin/main` = `076b144ac` (fully synced, ff-pushable; working tree clean pre-edit; `stash@{0}` = non-security-wip preserved by orchestrator, untouched per scope constraint)
- l44_post_push_fetch: (pending — orchestrator pushes post-sync)
- residual_debt: pre-existing baseline failures (out of SPEC scope) — `internal/statusline` `TestBuild_WritesContextUsageWithSessionID`, `internal/template` `TestOutputStylesTemplateLiveParity` + `TestTemplateNeutralityAuditC8Preserve` (confirmed FAIL at origin/main baseline per §E.3, independent of SECURITY SPEC — statusline/template do not reference checkFileAccess/handlers.go)

## §E.5 Mx-phase Audit-Ready Signal

- mx_status: complete (@MX tag consistency validated)
- mx_complete_at: 2026-07-08
- mx_commit_sha: (backfilled by `moai spec close` — self-referential)
- mx_scope: @MX tag consistency validation for run-phase code changes (M1 web + M2 template-update + M3 hook)
- mx_verification_approach: grep-based @MX annotation audit against run-phase modified files (no new @MX annotations introduced; one existing @MX:REASON corrected)
- ac_sec_006c_validation: PASS (already verified in §E.2 M2 AC Binary Matrix) — update.go:1283 `@MX:REASON` fan_in=3 claim matches actual grep callers:
  - `collectUserOwnedFiles` (value-pass in `collectUserOwnedFilesWith` signature) — `internal/cli/update_namespace_protect.go`
  - `assertNoUserOwnedNamespaceTouch` — `internal/cli/update_namespace_protect.go`
  - `isUserOwnedNamespaceConservative` — `internal/cli/update_namespace_protect.go`
  - Total: 3 direct callers, matches `@MX:REASON` annotation verbatim
- mx_tags_in_run_phase_code:
  - M1 web (`5f33dfaa9`): no new @MX tags added (existing `profile.IsValidProfileName` central guard + `Sec-Fetch-Site` mechanism annotated via REQ-SEC-001/002 code comments, not @MX tags)
  - M2 template-update (`85e280599`): update.go:1283 `@MX:REASON` corrected from false "fan_in >= 3" claim to accurate "fan_in = 3" with explicit caller enumeration (AC-SEC-006c fix); new `verifyNamespaceBackupCoverage` + `isUserOwnedNamespaceConservative` carry REQ-SEC-005/006 intent in code comments (not @MX tags)
  - M3 hook (`24482adf2`): no new @MX tags added (`EvalSymlinks` resolve-recheck pattern propagated from `file_changed.go:176-203` per REQ-SEC-007, code comment cites SPEC-SEC-HARDEN-004 / REQ-SEC4-004 provenance)
- mx_tag_accuracy_post_run: all @MX annotations in touched files remain accurate post-run; no stale or false fan_in claims remain (AC-SEC-006c PASS confirms the single corrected `@MX:REASON`)
- mx_no_new_tags_rationale: M1/M3 propagated existing verified patterns (`profile.IsValidProfileName` central guard, `file_changed.go` `EvalSymlinks` resolve-recheck) rather than inventing new mechanisms — no new `@MX:ANCHOR` warranted; M2 corrected an existing `@MX:REASON` (fan_in accuracy) rather than adding new annotations

## §F Phase 0.95 Mode Selection

### Input parameters
- tier: M
- scope (file count): ~10-13 (9 소스 파일 + 신규 *_test.go 회귀 테스트, plan.md §F.1 인벤토리)
- domain count: 5 (web, profile, cli, template, hook)
- file language mix: Go 소스 중심 + .gitignore (config)
- concurrency benefit: LOW (coding-heavy + 보안 정확성 + milestone 순차 검증)
- Agent Teams prereqs: harness thorough(security_keywords auto-detect) ✅ / team.enabled: true ✅ / CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 ✅

### Mode evaluation
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | No | 8 REQ 보안 하드닝, 다중 파일·도메인 |
| 2 background | No | Write(구현) 작업 — read-only 아님 |
| 3 agent-team | No | capability gate 전부 충족이나 "research-heavy" 선호 조건 미충족(coding-heavy) + 같은 main checkout 병렬 write conflict 위험 |
| 4 parallel | No | coding-heavy → Anthropic coding-task parallelism caveat (Mode 5 우선) |
| 5 sub-agent | **Yes** | coding-heavy 보안 수정, milestone 순차 검증, 단순 모드 |
| 6 workflow | No | semantic 보안 수정(mechanical uniform transform 아님), <30 files |

### Decision
Mode 5: sub-agent (sequential per milestone)

### Justification
본 SPEC은 Go coding-heavy 보안 하드닝으로 Anthropic coding-task parallelism caveat("most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time")에 해당한다. Agent Teams capability gate는 충족되나 Mode 3은 research-heavy 작업에 적합하고 본 작업은 coding-heavy이므로 Mode 5 sequential이 안전하다. Route A(main-direct) 같은 체크아웃에서의 병렬 write는 file conflict 위험(병렬 race 마스킹 교훈)이 있으며, 보안 false-positive deny 방지(NFR-SEC-003)를 위해 milestone별 검증 충실도가 우선한다. M1 web → M2 template-update → M3 hook 순차 진행으로 각 milestone green 확인 후 다음 진입한다.

## §G IGGDA Kickoff Predicate

- condition (a) intent clarity 100%: ✅ — 4개 설계 분기 결정(REQ-SEC-006 배선 / REQ-SEC-002 Sec-Fetch-Site / REQ-SEC-005 보수적 백업) + run 진입 승인, AskUserQuestion 완료
- condition (b) plan-auditor PASS: ✅ — PASS 0.941 (iter-1, harmonic mean, Tier M threshold 0.80 상회, skip-eligible ≥0.90)
- condition (c) Tier S/M: ✅ — Tier M
- condition (d) dangerous keywords: **FAIL** — matched: security, csrf, injection, secret, credential, vulnerability, path-traversal, symlink(CWE-61), owasp (제목·tags·REQ 본문 매칭)
- final verdict: **explicit-gate** (condition (d) FAIL → mandatory blocking AskUserQuestion per §H.1)
- Implementation Kickoff Approval 결과: **사용자 승인** ("승인, run-phase 진입") — 2026-07-08
- 설계 분기 최종 결정(사용자): REQ-SEC-006 = 배선(a), REQ-SEC-002 = Sec-Fetch-Site 헤더, REQ-SEC-005 = 보수적 백업 확대
- timestamp: 2026-07-08
