# Progress — SPEC-DEADPKG-INVESTIGATE-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** | era: **V3R6** | status: **draft**
- Artifacts: `spec.md` + `plan.md` + `acceptance.md` + `progress.md` created (plan-phase).
- SPEC ID Pre-Write Self-Check: `decomposition: SPEC ✓ | DEADPKG ✓ | INVESTIGATE ✓ | 001 ✓ → PASS`
- Frontmatter: 12 canonical fields + `tier: M` + `era: V3R6` + `related_specs` — validated.
- Plan-phase reachability baseline **observed** (not assumed) against the working tree
  (module `github.com/modu-ai/moai-adk`, go 1.26.4):
  - `internal/design` 19 files, 0 external importers.
  - `internal/research` 17 files, 0 external importers (stale `@MX:REASON` in `safety/limiter.go`).
  - `internal/runtime` (top-level) 8 files, 0 production importers; owning `SPEC-WF-AUDIT-GATE-001`
    is `implemented`; `gobin` subpackage LIVE (protected).
  - `internal/migrate` 1 file (`CleanupUserSettings`), 0 callers; `@MX:REASON` in
    `retired_events.go`; overlaps LIVE `internal/migration`.
  - `internal/i18n` 2 files, imported only by `internal/github/*_test.go`.
- Milestones defined: M1 runtime, M2 research, M3 design, M4 i18n, M5 migrate, M6 verification.
- Governing invariant recorded: a "dead package" is a defect claim = hypothesis until the
  run-phase tool checks confirm it (verification-claim-integrity §1.1 surface 3 + §5).
- Plan-phase gate: **ready for plan-auditor review**.

## §E.2 Run-phase Evidence

### Investigation complete — per-package verdicts (dedicated-tool evidence, 2026-07-02)

오케스트레이터가 전용 도구로 도달성 + 의도를 진단 (dispositions는 사용자 kickoff 대기 — 사용자 부재로 대량 삭제 미실행). 증거:
- 도달성: `go list -deps ./cmd/moai/...` (transitive prod dep tree)
- 미사용 함수: `deadcode ./cmd/... ./internal/...`
- 의도: `git log -1 -- <pkg>` (owning-SPEC) + `@MX:REASON` scan

| 패키지 | go list -deps | 의도(owning-SPEC / @MX) | verdict | disposition (kickoff 대기) |
|--------|---------------|------------------------|---------|---------------------------|
| internal/design (19) | UNREACHABLE | SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001 **completed 은퇴** ("remove moai-design-system"); deadcode가 dtcg/* 함수 다수 unreachable 확증 | **DEAD** (강한 증거) | 삭제 (+ test 파일) |
| internal/research (17) | UNREACHABLE | owning-SPEC 최근 활동 없음(마지막 touch=대량 mv); @MX(safety/limiter) 주장 disconfirmed(0 importer) | **DEAD** (경향; A/B 실험엔진 — staged 가능성 잔존) | 삭제 (+ test) — 단, staged 의도 최종 확인 권장 |
| internal/runtime (top, 8) | UNREACHABLE (top-level; `internal/runtime/*` deps는 Go stdlib false-positive였음) | SPEC-WF-AUDIT-GATE-001 **implemented** — staged audit-gate 구현체 | **RETAINED** | 보존 + retention 마커. `internal/runtime/gobin`은 REACHABLE(보호) |
| internal/migrate (1) | UNREACHABLE | `CleanupUserSettings` 0 callers; `@MX:REASON` "future observability" | **RELOCATE** | `internal/migration` step으로 흡수 + `retired_events.go` referent 갱신 |
| internal/i18n (2) | UNREACHABLE | prod importer 0; test-only(`internal/github/{integration,issue_close_integration}_test.go`) | **RELOCATE** | github test helper로 흡수 |

**Dispositions NOT executed** (사용자 부재 — 대량 삭제/relocate는 파괴적이라 kickoff 대기). 다음 세션에서 사용자 승인 후 milestone별 실행: design/research 삭제(+ test), migrate/i18n relocate, runtime retention 마커. 각 disposition 후 cross-platform build + 전체 테스트 검증 의무.

## §E.3 Run-phase Audit-Ready Signal

Dispositions 실행 완료 (cycle_type=ddd, 격리 워크트리 → main reconcile). 결과:

| 패키지 | disposition | 결과 |
|--------|-------------|------|
| internal/design (51파일) | DELETE | 제거 완료 (dtcg/pipeline/categories 전체 + testdata) |
| internal/research (33파일) | DELETE | 제거 완료 (experiment/eval/observe/safety/dashboard) |
| internal/migrate (2파일) | RELOCATE | `CleanupUserSettings` → `internal/migration/migrations/m002_settings_cleanup.go`(m002Apply, Version:2, atomic-write guard 보존) + test 이식; `retired_events.go` referent 갱신; 패키지 제거 |
| internal/i18n (3파일) | RELOCATE | `errors.go`/`templates.go` → `internal/github/i18n_helper_test.go`(package github, test-scoped) + 2 test import 갱신; 패키지 제거 |
| internal/runtime (top) | RETAIN | `config.go`에 @MX:NOTE retention 마커(SPEC-WF-AUDIT-GATE staged); gobin 보호 |

- 검증: `go build ./...` exit 0 (native + windows), touched-package 전체 GREEN (migration/migrations/github/runtime/hook `ok`). 외부 importer 0 확인 후 삭제(behavior 보존 — 제거 코드 미사용 입증).
- stale fixture 정리: `internal/template/internal_content_leak_test.go`의 삭제된 `internal/design/dtcg/frozen_guard_test.go` C7 regex fixture 참조 제거(나머지 4개 실존 경로로 검증 유지).
- 89파일 삭제 + 4 modified + 3 new.

## §E.4 Sync-phase Audit-Ready Signal

- Tier M 통합 close (run+sync). run-phase는 격리 워크트리(base==938aa668b) 실행 → orchestrator가 병렬 세션과 disjoint 확인 후 main으로 결정적 replicate(git rm + 파일 복사) + 재검증 + path-limited 커밋.
- spec.md/progress.md frontmatter: `status: draft → completed`, `era: V3R6`.
- CHANGELOG.md `[Unreleased] Removed` 항목 추가.
- 관측된 pre-existing red(내 변경 무관): `internal/template` catalog hash 불안정(manager-docs/git/plan-auditor) + reports-gitkeep + haiku-effort — 병렬 세션이 agent 파일 수정·push하며 `make build` 미실행으로 발생. 전체 워킹트리 stash 후에도 committed origin에서 FAIL 재현 확인. catalog 재생성은 병렬 세션 소관(그들의 미커밋 agent 변경 섞임 방지).

### sync_commit_sha

sync_commit_sha: <backfill>
run_commit_sha: <backfill>
