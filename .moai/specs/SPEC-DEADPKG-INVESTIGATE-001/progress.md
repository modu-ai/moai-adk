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

_<pending — dispositions execution + verification>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
