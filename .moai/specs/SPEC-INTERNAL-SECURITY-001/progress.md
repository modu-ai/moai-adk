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

## §E.3 Run-phase Audit-Ready Signal

- run_status: M1-complete (web 그룹)
- run_commit_sha: (pending — M1 commit)
- M1 AC 6/6 PASS (AC-SEC-001a/b/c, AC-SEC-002a/b/c)
- cross-platform build: `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- coverage: internal/web 69.8% (maintained — NFR-SEC-002 threshold ≥ 69.8%), internal/profile 82.6%
- subagent boundary (C-HRA-008): 0 matches in internal/web/ + internal/profile/
- lint: golangci-lint 0 issues (NEW: 0, pre-existing baseline: 0)
- NFR-SEC-003 (no false-positive deny): valid profile GET 200, same-origin POST success — explicit behavior-preservation tests PASS
- M2 (template-update) + M3 (hook) pending — separate spawn

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
