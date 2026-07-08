---
id: SPEC-INTERNAL-PERF-001
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-PERF-001 Progress

## §E.1 Plan-phase Audit-Ready Signal

plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 작성 완료, `status: draft`. plan-audit 및 Implementation Kickoff Approval 대기.

plan_complete_at: 2026-07-08
plan_status: audit-ready

## §E.2 Run-phase Evidence

### Commits (M1-M6, chronological)

| Milestone | REQ | Commit SHA | Subject |
|-----------|-----|------------|---------|
| M2 | REQ-PERF-004 | 9d0f0acfa | regex package-level promotion (era.go, status.go) |
| M5 | REQ-PERF-005 | 48a5a9ff7 | merge diff size guard (differ.go) |
| M6 | REQ-PERF-006 | de4d1762f | template single render (deployer.go) |
| M1 | REQ-PERF-001 | d69d4a868 | spec-lint git-query cache (gitquery_cache.go, lint.go) |
| M3 | REQ-PERF-002 | ad0ddaed0 | mx fan-in single-traversal index (validator.go) |
| M4 | REQ-PERF-003 | 96e0e1026 | CLI lazy initialization (root.go) |
| lint | — | 38f614960 | lint cleanup (differ_perf_test.go) |

> **V8 evidence-SHA correction (sync-phase)**: 상기 7 SHA는 sync-phase에서 `git log --oneline --grep="SPEC-INTERNAL-PERF-001"`로 재도출한 실측값. manager-develop가 기록한 원본 SHA(85cf27a08 / 34f5efbaf / d90a458f1 / 96d624041 / f2dfe4a8c / a5981768a / 43ea47ab0)는 병렬 세션 rebase로 교체되어 git log에 존재하지 않는 stale 값이었다 (verification-claim-integrity §2 baseline-attribution 위반 — sync-phase에서 실측값으로 정정). 각 SHA는 `git log --oneline | grep <sha>`로 존재 확인 완료.

### Coverage (measured 2026-07-08, `go test -cover ./internal/{pkg}/`)

| Package | Baseline | After | Status |
|---------|----------|-------|--------|
| internal/spec | 87.9% | 88.3% | >= baseline |
| internal/merge | 87.1% | 87.4% | >= baseline |
| internal/template | 85.8% | 86.0% | >= baseline |
| internal/hook/mx | 87.5% | 88.5% | >= baseline |

### Cross-platform build
```
$ go build ./...                           -> exit 0
$ GOOS=windows GOARCH=amd64 go build ./... -> exit 0
```

### Lint
```
$ golangci-lint run --timeout=2m -> 0 issues
```

### Pre-existing FAIL (Out of Scope per acceptance.md D.4)
- internal/cli: TestDoctor_Current_Light/Dark/NoColor, TestStatus_Current_Light/Dark/NoColor (statusline env flaky)
- internal/template: TestOutputStylesTemplateLiveParity (moai-easy.md parity - parallel-session residue in PRESERVE list)

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-07-08
run_commit_sha: 38f614960
run_status: run-complete (6/6 REQ implemented, 0 NEW FAIL)
ac_pass_count: 17 (AC-PERF-001a..006b - see E1 matrix in final report)
ac_fail_count: 0
new_warnings_or_lints_introduced: 0 (golangci-lint clean)
cross_platform_build.darwin_arm64: PASS
cross_platform_build.windows_amd64: PASS
total_run_phase_files: 14 (8 production + 6 test/doc)
m1_to_mN_commit_strategy: per-milestone independent commit (7 commits)

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: sync-complete (manager-docs sync-phase, single sync commit 3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 / Status Transition Ownership Matrix)
- sync_complete_at: 2026-07-08
- sync_commit_sha: <backfill pending — SHA cannot be in its own commit; populated in follow-up chore commit per established convention (SPEC-BRAND-DIR-REMOVE-001 `2c637ece2` / SPEC-INTERNAL-SECURITY-001 `f3193bac8` / SPEC-HANDOFF-GOALFIX-001 `bf983aaf5`)>
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Added` — SPEC-INTERNAL-PERF-001 entry (6 REQ / 17 AC sub-criteria; M1-M6 + lint cleanup)
- frontmatter_status_transitions: spec.md `in-progress → completed` + plan.md/acceptance.md/progress.md `draft → completed` atomic on single sync commit; `updated: 2026-07-08` refreshed in all 4 frontmatter blocks
- evidence_corrections:
  - V8 (§E.2/§E.3 stale-SHA fix): 7 run-phase commit SHAs re-derived from `git log --grep="SPEC-INTERNAL-PERF-001"` — every SHA now `git log`-verified (no stale, no carry-over; verification-claim-integrity §2 baseline-attribution restored)
  - V9 (AC-PERF-001a mechanical-gap): recorded as §K Run-phase Debt (fresh letter §K, no §E/§J reuse)
- mx_tag_validation: PASS — 모든 신규 PERF-001 구조가 적절한 @MX 태그 + REQ 역참조를 보유 (run-phase가 추가; sync-phase는 validate-only): `gitQueryCache` @MX:ANCHOR REQ-PERF-001-A (`gitquery_cache.go:18`), `progressFieldYAMLPattern` @MX:ANCHOR REQ-PERF-004-A (`era.go:217`), status regexes @MX:NOTE REQ-PERF-004-A (`status.go:16`), `fanInIndex` @MX:ANCHOR REQ-PERF-002-A (`validator.go:124`), `renderCache` @MX:NOTE REQ-PERF-006-A (`deployer.go:68`), `diffLinesThreshold` @MX:NOTE REQ-PERF-005-A (`differ.go:15`)

## §F Phase 0.95 Mode Selection

> Implementation Kickoff Approval 획득 후(IGGDA §H 4-condition 충족, 사용자 승인 2026-07-08), 첫 run-phase `Agent()` spawn 직전 기록. Phase 0.95는 자율 결정(orchestration-mode-selection.md §B).

### Input parameters

- tier: M
- scope (file count): ~15 (프로덕션 8-10 + 테스트 ~6 — plan.md §G file-touch inventory)
- domain count: 5 (`internal/spec`, `internal/hook/mx`, `internal/cli`, `internal/merge`, `internal/template`)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy — Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: FAIL (`workflow.team.enabled: false` — Sonnet 5 / Opus 4.8 re-design default-disabled)

### Mode evaluation

| Mode | selected | rationale |
|------|----------|-----------|
| 1 trivial | No | 6 REQ · 7 milestone · Go semantic 코드 변경 |
| 2 background | No | write 작업 — 동기 실행 필요 |
| 3 agent-team | No | `team.enabled: false` (default-disabled) |
| 4 parallel | No | coding-heavy — Anthropic caveat (Mode 5 우선) |
| 5 sub-agent | **Yes** | coding-heavy 순차 per-milestone, default fallback |
| 6 workflow | No | semantic/inter-file-dep 코드 (Mode 5 우선) |

### Decision

Decision: sub-agent (Mode 5 — sequential)

### Justification

코딩 집약적(semantic 코드 변경 + inter-file dependency) 작업으로 Anthropic coding-task parallelism caveat 적용 — 병렬 fan-out(Mode 4)보다 순차 sub-agent(Mode 5)가 안전 기본. 5 domain이나 coding-heavy이므로 Mode 4보다 Mode 5. `team.enabled: false`로 Mode 3 배제. 7 milestone(M0 측정 인프라 → M1 P0 → M2~M6 → M7 종합 검증)을 manager-develop 1위임에 순차 진행.

## §J Plan-phase Debt

> plan-auditor Phase 0.5 audit MINOR 결함 3건(D2/D4/D5)의 run/sync-phase 가시화 기록. 본 섹션은 기록 전용(plan-phase debt registry)이며 본SPEC의 plan-phase 범위에서 수정하지 않는다. `§E` namespace는 era.go 파서가 lifecycle 구조로 예약하므로 별도 letter(§J)를 사용한다.

### D2 (MINOR) — REQ-PERF-001-A cache key에 query-kind 누락 가능성

- **현상**: REQ-PERF-001-A가 캐시 키를 `(BaseDir, SpecID)`로 명시하나 query-kind(질의 종류)는 생략. plan §F M1 방향과 §H R1 risk는 query-kind 포함을 요구.
- **영향**: REQ-level WHAT이 silent → 두 규칙이 서로 다른 git 질의 결과를 동일 키로 오염 공유할 수 있는 cross-rule cache-contamination 해석 간극.
- **bounding**: plan §H R1(캐시 키에 query-kind 포함 의무) + REQ-PERF-001-C(doc/code parity AC)가 run-phase 구현을 종속. REQ 본문 수정 없이도 run-phase에서 R1이 구현 방향을 보정.
- **선택적 run-phase 액션**(필수 아님): REQ-PERF-001-A 본문을 query-kind 포함으로 강화하거나, cross-rule isolation AC sub-criterion을 acceptance.md에 추가.
- **결정**: 사용자 판단으로 REQ-PERF-001-A 본문 미수정(plan-phase debt 기록만). run-phase에서 R1 + REQ-PERF-001-C가 기능적 격리를 보장.

### D4 (MINOR) — AC가 strict GEARS 대신 Given/When/Then test-scenario 형식

- **현상**: acceptance.md의 AC들이 strict GEARS pattern 대신 Given/When/Then test-scenario 형식으로 작성됨.
- **분류**: mislabel 아님 — acceptance.md 헤더가 이들을 sub-criteria로 명시. GEARS 무게는 spec.md §B의 REQ가 담당.
- **판정**: borderline PASS — action 불필요, audit trail 목적 기록.
- **결정**: plan-phase 수정 없음.

### D5 (MINOR) — plan.md §A.1 strategies.go line-number drift

- **현상**: plan.md §A.1이 `internal/merge/strategies.go` `computeLineChanges`를 L107-108로 인용하나 실제 def는 L222 (~115줄 offset).
- **분류**: content-token match는 유효(`computeLineChanges` 함수명 자체는 정확). plan §A.1 footnote가 line-drift asymmetry를 self-acknowledge함.
- **bounding**: run-phase M5 착수 시 content-token 앵커 우선 재검증 의무(plan §A.1 footnote + CLAUDE.local.md line-number drift 교훈). line-number는 보조 신호일 뿐.
- **결정**: plan-phase 수정 없음. run-phase에서 content-token 기준 재검증 예정.

## §K Run-phase Debt

> sync-phase에서 발견된 run-phase 증거 부채 레지스트리. 본 섹션은 기록 전용이며, `§E` namespace(era.go 파서가 lifecycle 구조로 예약) 및 `§J`(plan-phase debt)와 충돌하지 않는 별도 letter(§K)를 사용한다 (spec-frontmatter-schema.md § progress.md Section Map의 "fresh-letter allocation rule" 준수).

### AC-PERF-001a-mechanical-gap

- **현상**: AC-PERF-001a (P0)는 "N=20 합성 catalog + git subprocess counting test double → spawn count ≤ 2×N+C, ≥50% reduction"의 기계적 증명을 요구. run-phase는 code inspection으로만 검증 (`cachedGitDirAvailable`/`cachedMainBranch`가 per-SPEC `exec.Command`를 per-run cached lookup으로 대체). 정량적 상한(quantitative upper bound)이 N=20 합성 catalog counting test double로 기계적으로 측정되지 않음.
- **영향**: 캐싱 메커니즘 자체는 sound (per-SPEC 4-spawn baseline → per-run 2-spawn(디렉터리 캐시 + 브랜치 캐시)으로 전환). 그러나 AC 본문이 요구하는 "≤ 2×N + C" 상한이 기계적으로 증명되지 않아, 엄격한 AC 증거가 code inspection에만 의존함.
- **bounding**: (1) 방향성(≥50% reduction)은 구조적으로 보장 — per-SPEC 4-spawn이 per-run 상수-spawn으로 축소되므로 N 증가에 무관하게 spawn 수는 상수에 수렴; (2) REQ-PERF-001-C doc/code parity AC + 기존 lint 테스트 무변경 PASS가 기능적 보존을 보장; (3) 본 SPEC 게이트는 "신규 FAIL 0" (acceptance.md §D.4)로 정의되어 본 gap은 게이트 위반이 아님 — 정량 증명은 본 SPEC의 MUST-PASS gate가 아닌 증거 품질 영역.
- **결정**: 사용자 결정으로 debt 수락 (code-inspection verification + sound per-run caching mechanism으로 방향성 충분 확보; 정량 상한의 기계적 증명은 후속으로 연기). 선택적 후속: `getGitImpliedStatus` 계열을 injectable command-runner interface로 refactor + N=20 합성 catalog + counting test double로 spawn 수를 기계적으로 측정하는 별도 SPEC/마일스톤 (본 SPEC scope 외).
