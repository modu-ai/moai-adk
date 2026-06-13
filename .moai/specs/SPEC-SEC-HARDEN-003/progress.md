# SPEC-SEC-HARDEN-003 — 진행 상황 (progress.md)

> 4-phase lifecycle: plan → run → sync → Mx. era V3R6 (H-4 auto-detect on §E.2/§E.5 sidecar markers when populated).

## §F.1 Plan-phase

- **Owner**: manager-spec (`(none) → draft` 전이)
- **Tier**: S (minimal)
- **cycle_type**: tdd (reproduction-first)
- **Scope**: 정확히 2 MEDIUM 결함 봉쇄 + 1 in-scope sibling.
  - C-F1 — `internal/hook/file_changed.go:runMXScan` MX 사이드카 비격리 경로 봉쇄.
  - C-F2 — `internal/cli/update.go:restoreMoaiConfigLegacy` 심볼릭 링크/traversal 봉쇄.
  - C-F2 sibling — `internal/cli/update.go:restoreMoaiConfig` 동일 심볼릭 링크 클래스(in-scope, ground-truth 검증).
- **Artifacts**: spec.md (GEARS, 12 frontmatter fields, §F Exclusions) + plan.md (M1/M2) + acceptance.md (AC SSOT) + progress.md.
- **재사용 봉쇄 seam**:
  - C-F1: `internal/hook/path_resolve.go` `resolveProjectRootFromInputOrEnv` (B7 canonical) + 신규 root-relative additive guard.
  - C-F2: SEC-HARDEN-002 `internal/cli/specid` idiom 모델 + 신규 `os.Lstat`/`filepath.Rel` 사적 봉쇄 헬퍼 (specid import 없음).
- **Exclusions**: §F.1 `${IFS}` shell-aware (mvdan.cc/sh 의존 결정 대기), §F.2 env-trust, 새 추상화/패키지/플래그 표면.
- **In-scope 판정**: 모던 복원 경로는 제외가 아니라 in-scope sibling (REQ-SEC3-006) — 레거시만 봉쇄 시 주 경로에 동일 취약점이 남아 security theater가 되므로 포함.

### Milestones (plan-phase 산출)

| Milestone | 대상 | REQ | reproduction 테스트 |
|-----------|------|-----|---------------------|
| M1 | `internal/hook/file_changed.go:runMXScan` | REQ-SEC3-001/002/003/004 | `TestRunMXScan_RejectsUncontainedFilePath`, `TestRunMXScan_RejectsUncontainedSidecarCWD` |
| M2 | `internal/cli/update.go` 레거시+모던 복원 | REQ-SEC3-005/006/007/008/009 | `TestRestoreMoaiConfigLegacy_SkipsSymlinkEntry`, `TestRestoreMoaiConfigLegacy_RejectsTraversalTarget`, `TestRestoreMoaiConfig_SkipsSymlinkEntry` |

### Plan-phase audit-ready signal

plan_complete_at: 2026-06-14T00:00:00Z
plan_status: audit-ready
plan_auditor_verdict: PASS-WITH-DEBT 0.81 (iter-1, Tier S thresh 0.75) — 모던 경로 in-scope 확장 = 정당(제외 시 security theater). D1(tier:S)/D2 BLOCKING(restoreMoaiConfigModern 팬텀명 → restoreMoaiConfig 익명 walk 콜백) /D3(AC-005 vacuous filepath.Rel grep) + lint MissingExclusions, 전부 orch-direct 패치 후 spec-lint clean(✓ No findings). iter-2 재spawn 생략(기계적 수정), run-phase Phase 0.5 재감사 예정.

## §F.2 Run-phase

(미착수 — manager-develop가 `draft → in-progress` 전이로 채움)

## §F.3 Sync-phase

(미착수 — manager-docs가 `in-progress → implemented` 전이로 채움)

## §F.5 Mx-phase

(미착수 — `implemented → completed` 4-phase close)
