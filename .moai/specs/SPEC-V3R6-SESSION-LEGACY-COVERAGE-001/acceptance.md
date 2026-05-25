---
id: SPEC-V3R6-SESSION-LEGACY-COVERAGE-001
title: "internal/session 패키지 test coverage 보강 — Acceptance Criteria"
version: "0.2.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "session, coverage, test-only, behavior-preserving, tier-s, sprint-10, acceptance"
sync_commit_sha: "a440b5c2f"
---

# SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — Acceptance Criteria

## §D. AC Matrix (7 ACs — REQ-SLCO-001..007)

| # | AC ID | REQ Covered | Severity | Phase | Verification Command (binary PASS/FAIL) | Expected Output |
|---|-------|------------|---------|-------|---------------------------------------|-----------------|
| 1 | AC-SLCO-001 | REQ-SLCO-001 | Must-Pass | run | `go test -cover ./internal/session/... 2>&1 \| grep -oE 'coverage: [0-9]+\.[0-9]+%' \| awk -F'[ %]' '{print ($2 >= 85.0) ? "PASS" : "FAIL"}'` | `PASS` (coverage ≥ 85.0% — current baseline 77.7% → target ≥85%) |
| 2 | AC-SLCO-002 | REQ-SLCO-002 | Must-Pass | run | `git diff --name-only main..<run-phase-HEAD> -- 'internal/session/' \| grep -v '_test\.go$' \| wc -l` | `0` (zero non-`_test.go` files modified in `internal/session/`) |
| 3 | AC-SLCO-003 | REQ-SLCO-003 | Must-Pass | run | `git diff --name-only main..<run-phase-HEAD> -- 'internal/session/*_test.go' \| xargs grep -L 't\.TempDir()' 2>/dev/null \| xargs -I{} grep -l -E '(os\.MkdirTemp\|ioutil\.TempDir\|/tmp/[a-z])' {} 2>/dev/null \| wc -l` | `0` (no new/modified test file bypasses `t.TempDir()` for temp dirs — checks for forbidden alternatives in changed test files; empty result expected) |
| 4 | AC-SLCO-004 | REQ-SLCO-004 | Must-Pass | run | `go test -race ./internal/session/... 2>&1 \| tail -3 \| grep -E '(ok\|FAIL)'` | `ok ...` (no `DATA RACE` lines; race detector clean) |
| 5 | AC-SLCO-005 | REQ-SLCO-005 | Must-Pass | run | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/session/ \| grep -v '_test\.go' \| grep -v '^[^:]*:[0-9]*:[ \t]*//'` | (empty) — production .go 파일에 AskUserQuestion/mcp__askuser 미도입; `subagent_boundary_test.go` 정적 grep guard PASS |
| 6 | AC-SLCO-006 | REQ-SLCO-006 | Must-Pass | run | `GOOS=windows GOARCH=amd64 go build ./internal/session/... 2>&1; echo "exit:$?"` | `exit:0` (cross-platform Windows build PASS — registry_lock_windows.go + lock_windows.go build tag 정합성) |
| 7 | AC-SLCO-007 | REQ-SLCO-007 | Must-Pass | run | `git diff main..<run-phase-HEAD> -- 'internal/session/*_test.go' \| grep -E '^\+.*t\.Setenv.*OTEL_'` | (empty) — 신규/수정 test 코드에 `t.Setenv("OTEL_*", ...)` 도입 0건 (CLAUDE.local.md §2 [WARN] 준수) |

## §D.1 Severity Legend

- **Must-Pass**: AC가 FAIL이면 phase 전체가 FAIL. plan-auditor PASS verdict 차단 (plan-phase) 또는 run-phase 종료 차단 (run-phase).
- **Should-Pass (Optional)**: AC가 FAIL이어도 phase는 진행 가능하나 self-compliance demerit로 plan-auditor verdict에서 부분 감점.

본 SPEC은 7 AC 모두 Must-Pass (Optional 없음) — test-only scope 의 모든 invariant 가 binary 강제 대상.

## §D.2 Phase Mapping

| Phase | Must-Pass ACs | Should-Pass ACs | 평가 시점 |
|-------|--------------|----------------|---------|
| plan | (none — all ACs are run-phase) | — | plan-phase 는 spec.md/plan.md/acceptance.md/progress.md write 완료만 검증 (구조적 frontmatter validity) |
| run | AC-SLCO-001..007 (all 7) | — | manager-develop run-phase 종료 시점 모두 PASS 필수 |

[NOTE] 본 SPEC 의 모든 AC 는 run-phase 평가. plan-phase 는 4 artifact write completion + 12-field frontmatter schema validity + plan-auditor PASS ≥ 0.75 (Tier S 임계값) 만 검증한다 (§D.5 Definition of Done — plan-phase 절 참조).

## §D.3 Traceability Matrix (REQ ↔ AC ↔ verification)

| REQ ID | EARS / GEARS Pattern | AC ID | Verification Pattern | Coverage % |
|--------|---------------------|-------|---------------------|-----------|
| REQ-SLCO-001 | Ubiquitous (SHALL ≥85%) | AC-SLCO-001 | `go test -cover` awk numeric threshold | 100% |
| REQ-SLCO-002 | Unwanted (SHALL NOT modify production) | AC-SLCO-002 | `git diff --name-only` filter non-`_test.go` = 0 | 100% |
| REQ-SLCO-003 | Ubiquitous (SHALL use t.TempDir) | AC-SLCO-003 | grep forbidden alternatives in changed test files = 0 | 100% |
| REQ-SLCO-004 | Event-Driven (WHEN race test runs, SHALL PASS) | AC-SLCO-004 | `go test -race` tail check | 100% |
| REQ-SLCO-005 | Unwanted (SHALL NOT introduce AskUserQuestion) | AC-SLCO-005 | `grep AskUserQuestion` in production .go = 0 | 100% |
| REQ-SLCO-006 | Ubiquitous (SHALL pass GOOS=windows build) | AC-SLCO-006 | cross-build exit code = 0 | 100% |
| REQ-SLCO-007 | Unwanted (SHALL NOT use t.Setenv OTEL) | AC-SLCO-007 | `git diff` grep `t.Setenv` + `OTEL_` = 0 | 100% |
| REQ-SLCO-008 | Optional (SHOULD NOT touch §24 namespace) | (no AC — Optional) | (informational only — not binary enforced) | (Optional, no AC required) |

7 mandatory REQs (REQ-SLCO-001..007) ↔ 7 mandatory ACs (AC-SLCO-001..007) — **100% bidirectional coverage on mandatory subset**. REQ-SLCO-008 [Optional] 은 self-compliance demerit 없이 informational only. plan-auditor S5 traceability dimension PASS 기대 (1.00).

## §D.4 Edge Cases

| 시나리오 | AC 영향 | 대응 |
|--------|--------|------|
| run-phase 가 신규 테스트 작성 도중 production 결함을 발견 | AC-SLCO-002 FAIL 위험 (production 수정 충동) | manager-develop blocker report return + 별도 SPEC `SPEC-V3R6-SESSION-<DOMAIN>-FIX-001` 분리 권장 (§A.6 R1 mitigation). 본 SPEC scope 보존. |
| coverage 측정 결과 84.9% (간발의 차이로 미달) | AC-SLCO-001 FAIL | (a) 추가 1-2 테스트 보강 후 재측정 — 권장 / (b) PASS-WITH-DEBT 처리 후 follow-up SPEC scope — 차선 / (c) 사용자 결정 필요 |
| `subagent_boundary_test.go` 가 신규 테스트 함수명 (예: `TestAskUserQuestionBoundary`) 을 false-positive trigger | AC-SLCO-005 FAIL (false positive) | 신규 테스트 함수명에 `AskUserQuestion` / `mcp__askuser` 토큰 사용 금지 (§A.6 R8). 함수명 우회로 boundary 정직 준수. |
| GOOS=windows 빌드가 macOS 환경 미설치 SDK 로 실패 | AC-SLCO-006 FAIL (환경 문제) | `GOOS=windows GOARCH=amd64 go build` 는 Go cross-compile 표준 — SDK 불필요. 만약 환경 결함이면 CI 환경에서 재확인 + run-phase 환경 진단 |
| race detector 가 기존 production 코드의 pre-existing race 를 발견 (본 SPEC 도입 아님) | AC-SLCO-004 FAIL (오해 발생) | `git diff main..<HEAD> -- internal/session/` 가 `_test.go` 만 변경했는지 재확인. production 코드 race 는 본 SPEC scope 외 — 별도 SPEC 분리 |
| 신규 테스트가 OTEL exporter mock 도입 필요 | AC-SLCO-007 PASS 유지 | fake/no-op exporter struct 를 test helper 로 추가 (NOT via t.Setenv). 예: `type noopExporter struct{}` + `func (n *noopExporter) Export(...) error { return nil }` |

## §D.5 Definition of Done

본 SPEC **plan-phase** 는 다음을 모두 만족할 때 DONE 으로 간주한다:

- [ ] 4 artifact (`spec.md`, `plan.md`, `acceptance.md`, `progress.md`) 작성 완료
- [ ] 12-field frontmatter schema 4 파일 모두 준수 (id / title / version / status / created / updated / author / priority / phase / module / lifecycle / tags — snake_case alias 0건)
- [ ] SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` 매칭 (Pre-Write Self-Check Protocol PASS — decomposition 결과 본 prompt 응답 본문에 기록)
- [ ] L46 path-specific staging — `git diff --cached --name-only` 가 정확히 4 paths (SPEC directory 하의 4 .md 파일만)
- [ ] L44 pre-spawn fetch + post-commit fetch 모두 `0 0` (clean)
- [ ] plan-auditor self-audit verdict ≥ 0.75 (Tier S 최소) — 목표 ≥ 0.90 (skip-eligible)
- [ ] `MissingExclusions` self-compliance — 본 spec.md §4 에 `### §4.N Out of Scope — <topic>` H3 + `-` list item ≥1 (4개 sub-section 모두 충족)

본 SPEC **run-phase** 는 별도 사이클에서 다음을 모두 만족할 때 DONE 으로 간주한다:

- [ ] AC-SLCO-001..007 (run-phase Must-Pass 7개) 모두 PASS
- [ ] coverage `internal/session` 패키지 ≥ 85.0%
- [ ] production .go 파일 변경 0건 (binary `git diff --name-only` filter)
- [ ] race detector clean (`go test -race ./internal/session/...` PASS)
- [ ] cross-platform `GOOS=windows GOARCH=amd64 go build ./internal/session/...` exit 0
- [ ] subagent boundary preserved (`grep AskUserQuestion internal/session/` non-test = 0 lines)
- [ ] commit message 에 SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 attribution + Conventional Commits prefix

## §D.6 Indirect Verification (auxiliary signals)

기본 AC matrix 외 다음 보조 signal 들이 run-phase 품질 평가에 활용된다 (binary 가 아니므로 informational):

- **Coverage delta per function** (`go tool cover -func=cover.out | grep -E '(state|store|registry|blocker)'`): 우선순위 P1~P3 함수 (MarshalJSON / mergePhaseStates / WriteRunArtifact / RecordBlocker / ResolveBlocker) coverage 가 baseline 대비 monotonically increase 했는지 확인.
- **Test execution time** (`go test -count=1 ./internal/session/... 2>&1 | grep PASS | awk '{print $3}'`): 신규 테스트 도입 후 패키지 테스트 실행 시간이 baseline (≈ 1.87s) 대비 2x 이내 유지. 큰 증가 시 fixture 비효율 가능.
- **Per-file test count** (`grep -c '^func Test' internal/session/*_test.go`): file-별 테스트 함수 개수 분포가 한 파일에 과집중 (예: state_test.go +50개) 되지 않고 P1~P3 우선순위 비율 따라 분산.

위 3 signal 은 run-phase Section E (Self-Verification) deliverables 에 informational 으로 추가 권장.

## §D.7 Forward-looking checks (post-merge / next-SPEC integration)

- [ ] **Sprint 10 cohort tracking**: 본 SPEC 4-phase close 후 Sprint 10 entry SPEC 진행 상황을 `MEMORY.md` 에 반영 (Tier S minimal 1-pass cohort 14 → 15 sustain 검증).
- [ ] **Coverage trend baseline 갱신**: post-sync 시점 `internal/session` 패키지 coverage 신규 baseline (예: 85.X%) 을 향후 SPEC 들의 기준점으로 기록. 후속 SPEC 이 coverage 를 다시 떨어뜨릴 경우 regression detection.
- [ ] **Sibling 패키지 coverage 정리 후속 SPEC**: 본 SPEC 마감 후 `internal/hook/` / `internal/cli/` / `internal/harness/` 등 coverage 갭 패키지에 대한 동일 패턴 SPEC 제안 (Sprint 11+ 후보).

## §G. B12 Self-Test (sync-phase 시점 평가 항목)

B12 (manager-develop-prompt-template.md §B12 Sync-phase CHANGELOG emission discipline) 준수 의무. sync-phase 종료 시점 다음 항목 모두 PASS 필요:

- **CHANGELOG count**: `grep -c 'SPEC-V3R6-SESSION-LEGACY-COVERAGE-001' CHANGELOG.md` 결과 = `1` (NEW entry referencing 본 SPEC, pre-sync = 0 검증)
- **AC count match**: CHANGELOG entry 에 명시된 AC 개수 = `7` (acceptance.md §D AC Matrix 1~7 일치, NOT progress.md deferred AC)
- **Frontmatter status (all 4 artifacts)**: sync-phase 종료 후 spec.md / plan.md / acceptance.md / progress.md 모두 `status: implemented`
- **sync_commit_sha field**: 4 artifacts 모두 `sync_commit_sha: "<full-hash>"` 정상 backfill (manager-docs L60 partial backfill defect 방지 — 4 paths ALL 동시 업데이트 의무)
- **File paths claimed in CHANGELOG**: `internal/session/*_test.go` 만 reference (production 파일 reference 0건 — REQ-SLCO-002 정합)

## §H. Cross-references

- spec.md — canonical SSOT (REQ-SLCO-001..008 anchored).
- plan.md — Tier S minimal Section A only (§A.4 EXTEND table + §A.5 run-phase delegation contract).
- progress.md — Phase 0.5 plan-auditor verdict 기록 + lifecycle audit-ready signals.
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate authority + skip-eligible policy.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` §B12 — sync-phase CHANGELOG emission discipline (본 §G B12 self-test 출처).
- `CLAUDE.local.md §6` — Coverage Targets (85% project minimum) + Test Isolation (`t.TempDir()` HARD).
- `CLAUDE.local.md §2 [WARN]` — OTEL t.Setenv 데이터 레이스 회피.
- SPEC-LINT-CLEANUP-001 acceptance.md §D.7 — forward-looking checks 패턴 선례.
- SPEC-V3R6-MULTI-SESSION-COORD-001 — adjacent SPEC (방금 close, 동일 `internal/session/` 패키지). 본 SPEC 의 subagent_boundary_test.go preserve 의무 출처.
