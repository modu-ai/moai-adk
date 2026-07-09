---
id: SPEC-INTERNAL-TEST-004
title: "Regenerate stale doctor/status golden testdata for version bump rc7→rc10 (whole-repo green)"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0-rc10"
module: "internal/cli/testdata"
lifecycle: spec-anchored
tags: "golden-test, test-fix, debt-cleanup, version-bump"
tier: S
depends_on: []
related_specs: [SPEC-INTERNAL-TEST-003, SPEC-INTERNAL-ARCH-001, SPEC-WEB-CONSOLE-011]
---

# SPEC-INTERNAL-TEST-004 — Regenerate stale doctor/status goldens (rc7→rc10) for whole-repo green

## §A. Context / Problem Statement

`internal/cli` 의 6개 golden 테스트가 FAIL 상태이며, 이것이 whole-repo `go test ./...` exit 1 의 유일한 원인이다. 이 부채는 SPEC-INTERNAL-TEST-003 AC-006 에서 TEST-004 로 이관되었고, TEST-004 의 완료는 SPEC-INTERNAL-ARCH-001 plan-audit M0 재진입의 whole-repo-green 전제 조건을 unblock 한다.

### Verified failure evidence (manager-spec research-ran, not inferred)

명령: `go test ./internal/cli/ -run 'TestDoctor_Current_Light|TestStatus_Current_Light' -count=1`
결과: 6 FAIL. 축어 출력 + byte-level diff 는 `research.md` §2 + `.moai/state/verify/538fe6ae/test-004-golden-fail-detail.log` 에 보존됨.

### Root cause (decisive, via `UPDATE_GOLDEN=1` + `git diff`)

Golden testdata 가 마지막으로 재생성된 시점(TEST-002 M1 `ffea91710`, rc6→rc7) 이후 `pkg/version/version.go` 가 `rc7 → rc8 → rc9` 로 2회 bump 되었고(`9edb72af5 chore(version): bump to v3.0.0-rc9` 가 HEAD-committed 값), golden testdata 는 rc7 에 고착되어 있다. `UPDATE_GOLDEN=1` 재생성 후 `git diff internal/cli/testdata/` 를 실행하면 6개 파일 각각 정확히 1줄만 변경되며, 변경 내용은 버전 문자열 `v3.0.0-rc7 → v3.0.0-rc9` 이외의 어떤 byte 도 포함하지 않는다. 코드 출력(rc9)이 정확하고 golden(rc7)이 stale 이다 — 코드 회귀가 아니다.

### Scope-boundary note (cache-hit emoji is orthogonal)

TEST-003 progress.md 는 이 부채를 "statusline golden drift ... vs uncommitted `internal/statusline/renderer.go` working-tree changes" 로 특성화했으나, 이는 부정확하다. `internal/statusline/renderer.go` 의 💾→♻️ emoji 변경(`internal/statusline/cache_hit_test.go` 동반 sync 포함)은 `SPEC-WEB-CONSOLE-011` 조상 commit `22220186c` 에 속하며, golden 테스트(fixture 가 `cacheStrategy.enabled: false` 이므로 cache-hit 세그먼트가 렌더링되지 않음)에 어떤 영향도 주지 않는다. TEST-004 는 emoji 변경을 다루지 않는다 — 이 변경은 PRESERVE 이다. 상세는 `research.md` §3.

## §B. Requirements (GEARS)

### REQ-GOLD-001 — 6 golden tests PASS (Ubiquitous)

The 6 golden tests (`TestDoctor_{Current_Light,Current_Dark,NoColor}` in `internal/cli/doctor_golden_test.go` + `TestStatus_{Current_Light,Current_Dark,NoColor}` in `internal/cli/status_golden_test.go`) shall PASS after golden testdata regeneration reflecting the HEAD-committed version `v3.0.0-rc9`.

### REQ-GREEN-001 — Whole-repo exit 0 (Ubiquitous)

The whole-repo test suite (`go test ./...`) shall exit 0 with no package FAIL, unblocking `SPEC-INTERNAL-ARCH-001` plan-audit M0 whole-repo-green precondition.

### REQ-VER-001 — Golden diff is version-string-only (Event-detected)

**When** the 6 golden testdata files are regenerated via `UPDATE_GOLDEN=1`, the `git diff internal/cli/testdata/` shall show ONLY the version-string line change (`v3.0.0-rc7` → `v3.0.0-rc9`) — 6 files, 1 line each, no other byte changes — confirming no unintended golden mutation.

### REQ-PRESERVE-001 — Scope discipline (Event-detected)

**When** the golden regeneration is performed, the implementation agent shall modify ONLY the 6 files under `internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden` and shall NOT touch any Go source file, any `internal/statusline/` file, `pkg/version/version.go`, or any working-tree path listed in `plan.md` §D PRESERVE list.

### REQ-SCOPE-001 — Cache-hit emoji scope boundary (Capability gate)

**Where** the cache-hit emoji change (💾→♻️) in `internal/statusline/renderer.go` + `internal/statusline/cache_hit_test.go` is concerned, the implementation agent shall preserve those working-tree edits uncommitted — they belong to `SPEC-WEB-CONSOLE-011` (ancestor commit `22220186c`), are orthogonal to the golden drift (cache-hit segment is not rendered in golden fixtures), and are not a TEST-004 concern.

## §C. Constraints

- **Approach**: golden regenerate only (`UPDATE_GOLDEN=1`). No Go source code changes — version.go rc9 is the intended release version, not a regression.
- **PRESERVE**: ~15 unrelated working-tree paths (WEB-CONSOLE-011 statusline edits, template files, config files, other SPEC dirs) — see `plan.md` §D for the exhaustive list.
- **No commit from plan-phase**: manager-spec authors artifacts only; the golden regeneration + commit happens in run-phase (manager-develop).
- **version.go is frozen**: `v3.0.0-rc9` is the HEAD-committed release value. Do NOT bump, do NOT revert.

## §D. Success Criteria

1. 6 golden tests PASS (`go test ./internal/cli/ -run 'TestDoctor_|TestStatus_' -count=1` exit 0).
2. Whole-repo `go test ./...` exit 0.
3. `git diff` on the 6 goldens shows only the version-string line change.
4. No PRESERVE path modified.
5. SPEC-INTERNAL-ARCH-001 M0 whole-repo-green precondition is unblocked.

## §E. Out of Scope

### Out of Scope — Cache-hit emoji change (💾→♻️)

- `internal/statusline/renderer.go` line 312 emoji change and the `internal/statusline/cache_hit_test.go` sync — these belong to SPEC-WEB-CONSOLE-011 (ancestor `22220186c`) and are orthogonal to the `internal/cli/` golden drift. TEST-004 preserves them uncommitted.

### Out of Scope — version.go bump or revert

- `pkg/version/version.go` is the release-process-owned value (`v3.0.0-rc9`, committed via `9edb72af5`). TEST-004 regenerates goldens to MATCH it, not to change it.

### Out of Scope — Section-rendering logic changes

- The doctor sections (System / MoAI-ADK / Workspace) and status sections (Project / Configuration) renderers are NOT modified — the `UPDATE_GOLDEN=1` + `git diff` evidence proves the section structure is identical between got/want; only the version string differs.

### Out of Scope — ARCH-001 M0 execution

- TEST-004 unblocks the ARCH-001 whole-repo-green precondition but does not execute any ARCH-001 milestone. ARCH-001 is a separate SPEC with its own plan-audit re-entry.

## §F. References

- `research.md` — drift-cause analysis, scope decision, approach selection
- `plan.md` — milestones, PRESERVE list, self-verification
- `acceptance.md` — AC matrix with Given-When-Then
- SPEC-INTERNAL-TEST-003 `progress.md` §E.2 AC-006 — debt transfer provenance
- SPEC-INTERNAL-ARCH-001 — whole-repo-green M0 precondition consumer
- SPEC-WEB-CONSOLE-011 (`22220186c`) — cache-hit emoji scope owner

## §G. Resolution via External Absorption (ce2a509dc)

> **Reality-correction addendum (2026-07-09).** The §A–§F analysis above is a correct historical record of the drift diagnosis — the golden drift WAS real (goldens stuck at rc7 while `pkg/version/version.go` advanced). This §G records HOW the drift was actually resolved: by a DIFFERENT SPEC's commit, before TEST-004's own run-phase was invoked.

### §G.1 External-commit golden regeneration

The 6 golden testdata files (`internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden`) were regenerated by commit **`ce2a509dc chore(version): bump to v3.0.0-rc10 and realign tag with source`** — the rc10 version-bump commit — NOT by a TEST-004 run-phase commit. The commit message states verbatim: *"Regenerate the six doctor/status golden fixtures, which still pinned rc7 and were the cause of the standing TestDoctor/TestStatus failures"*. TEST-004's own run-phase was never invoked; the SPEC proceeds directly from plan to a consolidated sync close.

### §G.2 Target version became rc10 (not rc9)

The original plan-phase (§A–§F) targeted `v3.0.0-rc9`. The actual target became **`v3.0.0-rc10`** because a parallel session bumped rc9→rc10 to align the missing git tag (no rc9 git tag was ever created; the Makefile derives `VERSION` from `git describe --tags`, so the source-tree version and the tag had to be realigned). The golden regeneration in `ce2a509dc` therefore regenerated the fixtures to rc10, the current `pkg/version/version.go` value.

### §G.3 AC satisfaction status (mechanically verified)

| AC | Status | Evidence |
|----|--------|----------|
| AC-001..006 (6 golden PASS) | **SATISFIED via `ce2a509dc`** | `internal/cli` golden tests all PASS post-`ce2a509dc`; the 6 fixtures now carry rc10, matching `pkg/version/version.go` |
| AC-008 (version-string-only diff) | **SATISFIED via `ce2a509dc`** | The golden diff in `ce2a509dc` is the version-string line change (rc7→rc10), confirming REQ-VER-001 |
| AC-007 (whole-repo green) | **DEBT-TRANSFERRED** — NOT TEST-004-satisfiable | `go test ./...` still exits 1; the 3 remaining FAILs are in `internal/template` (`TestAllAgentsInCatalog`, `TestTemplateNoInternalContentLeak`, `TestRuleProvenanceAudit`) and belong to SPEC-AGENT-ARCH-V2-001 (super-advisor integration in progress in the working tree). Transferred as debt to SPEC-AGENT-ARCH-V2-001 / SPEC-INTERNAL-ARCH-001. |
| AC-009 (PRESERVE) | **HOLDS** | No TEST-004 run-phase commit touched any PRESERVE path; the 6 golden files are committed inside `ce2a509dc` and the PRESERVE paths in `plan.md` §D.2 remain in their pre-existing state |

### §G.4 REQ satisfaction summary

- **REQ-GOLD-001** (6 golden tests PASS): **MET** — via `ce2a509dc`.
- **REQ-VER-001** (golden diff is version-string-only): **MET** — via `ce2a509dc`.
- **REQ-GREEN-001** (whole-repo exit 0): **DEBT-TRANSFERRED** to SPEC-AGENT-ARCH-V2-001 / SPEC-INTERNAL-ARCH-001 — the remaining 3 FAILs are the super-advisor integration surface, not the golden drift.
- **REQ-PRESERVE-001** / **REQ-SCOPE-001** (scope discipline, cache-hit emoji boundary): **HOLDS** — no TEST-004 run-phase action touched any out-of-scope path.

### §G.5 Honest close posture

TEST-004 is closed honestly: the golden drift it was authored to fix WAS real (§A–§F diagnosis is correct), and the fix WAS applied (`ce2a509dc` applied exactly the golden regeneration TEST-004 specified). The one unsatisfied AC (AC-007 whole-repo green) is outside TEST-004's surface and is debt-transferred to the SPEC that owns the 3 remaining FAILs. This is an "external absorption" close: the work was completed by a sibling SPEC's commit rather than TEST-004's own run-phase.
