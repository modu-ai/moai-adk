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

- **Owner**: manager-develop (`draft → in-progress` 전이, M1 commit)
- **cycle_type**: tdd (reproduction-first) — 각 결함 재현 AC를 fix 전 RED로 확인 후 GREEN.
- **Base**: 5bdc95bfd (orchestrator plan-patch on plan-phase 83c1d46b8).

### §E.2 Run-phase Evidence (AC PASS/FAIL matrix)

| AC ID | 결함 | 종류 | 테스트 함수 | Actual Output | Status |
|-------|------|------|-------------|---------------|--------|
| AC-SEC3-001a | C-F1 | reproduction | `TestRunMXScan_RejectsUncontainedFilePath` | RED(fix 전 사이드카 생성) → PASS | PASS |
| AC-SEC3-001b | C-F1 | reproduction | `TestRunMXScan_RejectsUncontainedSidecarCWD` | RED(fix 전 C/.moai/state 탈출) → PASS | PASS |
| AC-SEC3-002 | C-F1 | containment | (001a/001b post-fix + 계약) | 봉쇄 동작·panic 없음·빈 payload 계약 보존 | PASS |
| AC-SEC3-003 | C-F1 | no-regression | `TestRunMXScan_AllowsInProjectPath` | PASS (정상 경로 사이드카 갱신) | PASS |
| AC-SEC3-004a | C-F2 | reproduction | `TestRestoreMoaiConfigLegacy_SkipsSymlinkEntry` | RED(fix 전 LEAKED 복원) → PASS | PASS |
| AC-SEC3-004b | C-F2 | reproduction | `TestRestoreMoaiConfigLegacy_RejectsTraversalTarget` | RED(fix 전 victim overwrite) → PASS | PASS |
| AC-SEC3-004c | C-F2 | reproduction(sibling) | `TestRestoreMoaiConfig_SkipsSymlinkEntry` | RED(fix 전 모던 walk LEAKED 복원) → PASS | PASS |
| AC-SEC3-005 | C-F2 | containment | (004a/b/c post-fix) | 링크 추종·configDir 탈출 쓰기 없음·panic 없음 | PASS |
| AC-SEC3-006 | C-F2 | no-regression | `TestRestoreMoaiConfigLegacy_AllowsRegularInConfigFile` | PASS (정규 파일 정상 복원) | PASS |
| AC-SEC3-007 | 공통 | no-new-surface | grep 정적 가드 | specid import 0 / 신규 cobra 플래그 0 | PASS |

검증 명령 실측:
- `grep -c resolveProjectRootFromInputOrEnv internal/hook/file_changed.go` → 2 (≥1, AC-SEC3-002).
- `grep -n os.Getenv internal/hook/file_changed.go` → 0 match (NFR-SEC3-002).
- `grep -c asyncDeadline internal/hook/file_changed.go` → 5 (NFR-SEC3-001 계약 보존).
- `grep -c os.Lstat internal/cli/update.go` → 2 (AC-SEC3-005 symlink 가드).
- `grep -n internal/cli/specid internal/cli/update.go` → 0 match (AC-SEC3-007).

### §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-14T05:00:00Z
run_commit_sha: <M2 commit — backfill if needed>
run_status: audit-ready
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "0 1 (plan-patch local-ahead, expected)"
l44_post_push_fetch: <backfill on push>
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: exit 0
  windows: exit 0
total_run_phase_files: 4  # file_changed.go, file_changed_test.go, update.go, update_fileops_test.go (+ spec.md/progress.md frontmatter)
m1_to_mN_commit_strategy: "M1 (C-F1, hook) → M2 (C-F2, cli) 분리 commit, main 직진 push"
```

관측 (무관, pre-existing): `internal/hook` 풀 병렬 suite에서 `TestHookWrapper_ValidJSON` /
`TestHookWrapper_MoaiBinaryFallback` 2건이 "signal: killed" / fallback-binary-not-found로 간헐 FAIL.
격리 실행 시 둘 다 PASS. wrapper_test.go가 실 moai 바이너리를 PATH로 exec하는 sandbox 자원-민감 결함이며
file_changed.go(runMXScan)와 무관. baseline에서 동일 관측됨 — 본 SPEC 회귀 아님.

## §F.2.1 Phase 0.95 Mode Selection

- Input: tier=S, scope=4 files (2 src + 2 test), domain count=2 (internal/hook + internal/cli), file mix=Go, concurrency benefit=LOW (coding-heavy), Agent Teams prereqs=not evaluated.
- Decision: sub-agent (Mode 5, sequential per-milestone).
- Justification: Tier S coding-heavy 2-file containment work. Per Anthropic coding-task parallelism caveat, sequential sub-agent is the correct default; M1(hook)→M2(cli) sequential split matches the disjoint-package milestone structure.

## §F.3 Sync-phase

(미착수 — manager-docs가 `in-progress → implemented` 전이로 채움)

## §F.5 Mx-phase

(미착수 — `implemented → completed` 4-phase close)
