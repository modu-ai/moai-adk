---
id: SPEC-INTERNAL-SECURITY-001
title: "internal/ 보안 하드닝 — 진행 상태"
version: "0.1.0"
status: in-progress
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

## §E.3 Run-phase Audit-Ready Signal

- run_status: M2-complete (template-update 그룹); M1 web + M2 template-update done, M3 hook pending
- run_commit_sha: (pending — M2 commit)
- M2 AC 8/8 PASS (AC-SEC-003a/b, AC-SEC-004a/b, AC-SEC-005a/b, AC-SEC-006a/c)
- cross-platform build: `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (post `make build`)
- coverage: internal/cli 71.5% (package-level, baseline 유지 — pre-existing statusline-golden 6 FAIL은 본 SPEC 범위 외); internal/template 85.9%. M2 신규 함수 91–100% (isUserOwnedNamespace 96.7%, verifyNamespaceBackupCoverage 91.7%, isUserOwnedNamespaceConservative 100.0%, assertNoUserOwnedNamespaceTouch 100.0%)
- subagent boundary (C-HRA-008): 0 matches in M2 신규/편집 파일 (update_namespace_protect.go, update_archive.go, update_security_m2_test.go, embed_gitignore_backups_test.go)
- lint: golangci-lint 0 issues (NEW: 0, pre-existing baseline: 0)
- NFR-SEC-003 (no false-positive deny): verifyNamespaceBackupCoverage empty/fully-backed-up cases pass → 정상 update 통과; conservative backup does not break EC-UNP-001 (no user-owned → no backup)
- NFR-SEC-004 (테스트 격리): 모든 심링크/secret fixture t.TempDir() 내
- M1 web (commit 5f33dfaa9) done; M3 hook pending — separate spawn

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_

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
