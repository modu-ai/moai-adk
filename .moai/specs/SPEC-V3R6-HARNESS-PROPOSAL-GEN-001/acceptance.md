---
id: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001
title: "V3R4 Self-Evolving Harness Loop Closure — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Self-Evolution Closure"
module: "internal/harness/proposalgen, internal/cli/harness"
lifecycle: spec-anchored
tags: "harness, proposal, acceptance, tier-m"
---

# Acceptance Criteria — SPEC-V3R6-HARNESS-PROPOSAL-GEN-001

## §A. AC Matrix (SSOT for AC count and binary PASS/FAIL)

| AC | Description | Verification Command | Expected Outcome |
|----|-------------|---------------------|------------------|
| **AC-PGN-001** | Reader parses live 8-record `tier-promotions.jsonl` without error and surfaces 4 unique pattern_key values | `go test -run TestReader_LiveFixture ./internal/harness/proposalgen/...` | `PASS` — test reads `.moai/harness/learning-history/tier-promotions.jsonl` (or testdata snapshot), asserts `len(promotions) == 8` and `len(uniquePatternKeys(promotions)) == 4` |
| **AC-PGN-002** | Mapper produces 0 actionable proposals from current system-event-only data (graceful no-op) | `go test -run TestMapper_CurrentDataNoOp ./internal/harness/proposalgen/...` | `PASS` — test feeds the 4-pattern-key fixture (`agent_invocation:Bash:`, `subagent_stop:unknown:`, `session_stop::`, `user_prompt::`), asserts `len(candidates) == 0` |
| **AC-PGN-003** | Scaffolder creates `.moai/proposals/<draft-id>/` directory tree ONLY when ≥1 actionable candidate exists; no-op skips creation | `go test -run TestScaffolder_NoOpSkipsCreation ./internal/harness/proposalgen/...` | `PASS` — test runs scaffolder against empty candidate slice in `t.TempDir()`, asserts `.moai/proposals/` directory NOT created (`os.Stat` returns `ErrNotExist`) |
| **AC-PGN-004** | CLI `moai harness propose --dry-run` exits 0 with stdout JSON matching `{"proposals":[], "reason":"no-actionable-patterns", "malformed_lines":0, "evaluated_patterns":4, "auto_delegate":false}` against the frozen current-data baseline fixture | `go test -run TestPropose_DryRun_BaselineFixture ./internal/cli/harness/...` | Test uses fixture at `internal/harness/proposalgen/testdata/tier-promotions-current-baseline.jsonl` (snapshot of live data as of 2026-05-24, 8 records, 4 unique pattern_keys). Exit code 0; stdout JSON has `.proposals == []`, `.reason == "no-actionable-patterns"`, `.evaluated_patterns == 4`, `.auto_delegate == false`. Run-phase manager-develop MUST create this fixture file as part of M1 deliverables — copy the 8-record current state of `.moai/harness/learning-history/tier-promotions.jsonl` into the testdata directory before authoring the test. The CLI test uses an environment variable override (e.g., `MOAI_HARNESS_TIER_PROMOTIONS_PATH`) or `--input` flag to point the propose subcommand at the fixture instead of the live file. |
| **AC-PGN-005** | Subagent boundary HARD: `grep -rn 'AskUserQuestion(' internal/cli/harness/` returns 0 matches (excluding test files and comments without invocation form) | `grep -rn 'AskUserQuestion(' internal/cli/harness/ \| grep -v '_test.go'` | Zero output |
| **AC-PGN-006** | `TestPropose_NoAskUserQuestion` static check passes (mirrors `TestNew_NoAskUserQuestion` in `internal/cli/worktree/new_test.go`) | `go test -run TestPropose_NoAskUserQuestion ./internal/cli/harness/...` | `PASS` — test reads each source file in `internal/cli/harness/*.go` (excluding `*_test.go`), asserts `!strings.Contains(src, "AskUserQuestion(")` for each |
| **AC-PGN-007** | Combined test coverage for `internal/harness/proposalgen/` + `internal/cli/harness/` ≥ 85% | `go test -coverprofile=cover.out ./internal/harness/proposalgen/... ./internal/cli/harness/... && go tool cover -func=cover.out \| tail -1` | Final `total:` line shows coverage ≥ `85.0%` |
| **AC-PGN-008** | `go vet ./...` and `golangci-lint run --timeout=2m` exit 0 with zero NEW issues attributable to this SPEC | `go vet ./... && golangci-lint run --timeout=2m` | Both commands exit 0; any pre-existing baseline issues NOT introduced by this SPEC's new files |

## §B. Optional AC (documentary, no executable verification)

| AC | Description | Documentation Anchor |
|----|-------------|---------------------|
| **AC-PGN-009** (Optional) | The orchestrator AskUserQuestion gate contract (Approve/Modify/Reject with `(권장)` on Approve, Other auto-appended for Modify) is documented in spec.md REQ-PGN-010/011 and plan.md §E. | spec.md §2.5; plan.md §A.3 decision 1 |

## §C. Edge Cases (verified within AC tests, no separate ACs)

The following edge cases are exercised inside the AC-PGN-001..008 test bodies; they do not warrant separate top-level ACs because they are internal robustness concerns rather than user-observable contracts:

| Edge case | Covered by AC | Test scenario |
|-----------|---------------|---------------|
| Empty `tier-promotions.jsonl` file | AC-PGN-001 | Reader test sub-case: zero-byte file → returns empty slice, no error |
| Missing `tier-promotions.jsonl` file | AC-PGN-001 | Reader test sub-case: non-existent path → returns empty slice, no error (per REQ-PGN-003) |
| Malformed JSONL line in middle of file | AC-PGN-001 | Reader test sub-case: 3-line fixture with line 2 corrupt → returns 2 valid records, `malformed_lines == 1` |
| Confidence exactly at threshold 0.70 | AC-PGN-002 | Mapper test sub-case: synthetic actionable pattern with Confidence=0.70 → included; with 0.69 → excluded |
| `pattern_key` with trailing colon (e.g., `session_stop::`) | AC-PGN-002 | Mapper test sub-case: current-data fixture explicitly includes this format |
| Multiple candidates exceeding `--limit N` | AC-PGN-004 (extended) | CLI test sub-case: synthetic 7-actionable fixture with `--limit 3` → JSON returns exactly 3 sorted by Confidence desc |
| Directory `.moai/proposals/` already exists with prior runs | AC-PGN-003 | Scaffolder test sub-case: pre-existing directory + 1 new candidate → new subdirectory created without disturbing existing siblings |

## §D. Definition of Done

All 8 mandatory ACs (AC-PGN-001..008) MUST pass before sync-phase invocation. Specifically:

1. **All 8 AC commands listed in §A return their expected outcomes.** Manager-develop's E1 self-verification matrix MUST include each command's actual stdout/exit-code captured verbatim.
2. **Coverage ≥85% verified via `go tool cover -func=cover.out | tail -1`.** Per-file coverage SHOULD be ≥80% for non-trivial files; trivial files (e.g., `types.go` with struct-only definitions) MAY be lower.
3. **Cross-platform build verified: `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.** This is a Section E1 deliverable per Section A-E template (Tier M).
4. **Subagent boundary CI guard test passes** (AC-PGN-006) — this is the C-HRA-008-equivalent regression test for the `internal/cli/harness/` package.
5. **No NEW golangci-lint issues attributable to this SPEC.** Baseline pre-existing issues in unrelated packages remain; new files in `internal/harness/proposalgen/` and `internal/cli/harness/` MUST be clean.
6. **CLI smoke test against live data** (AC-PGN-004) produces the documented JSON shape with `proposals: []` and `reason: "no-actionable-patterns"`. This empirically verifies the current-data graceful no-op (REQ-PGN-014).
7. **No PRESERVE list violations.** Manager-develop's E1 matrix MUST attest that `internal/harness/applier.go`, `internal/harness/classifier.go`, `internal/cli/hook.go`, `.moai/harness/learning-history/tier-promotions.jsonl`, and other PRESERVE entries (plan.md §A.4) are byte-identical to their pre-run-phase state.

Manager-develop's run-phase completion report MUST include the AC binary PASS/FAIL matrix (E1) populated for all 8 mandatory ACs, the coverage measurement (E3), the boundary grep result (E4), and the lint status (E5) per the Section A-E delegation template (E1-E7).

## §E. Decision Rule for Optional MAY Without AC

REQ-PGN-005 (mapper no-op emit with `evaluated_patterns: <total>`) contains an Optional MAY clause ("WHERE the mapper produces zero candidates"). Per the precedent established in `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/acceptance.md` §F (and reinforced by L48 lesson), Optional MAY clauses on Ubiquitous/State-Driven requirements do not require dedicated ACs when:

- The behavior is observable as a side-effect of a Mandatory requirement's AC (here: AC-PGN-002 verifies `len(candidates) == 0` which subsumes the no-op trigger condition)
- The Optional clause adds a diagnostic/observability field rather than a new behavior (here: `evaluated_patterns: <total>` is a count surfaced in CLI JSON output, validated as part of AC-PGN-004's `.evaluated_patterns == 4` assertion)

REQ-PGN-014 (current-data graceful no-op) is verified by AC-PGN-004 and represents the empirical realization of REQ-PGN-005 against live data.

## §F. Parallel Batch Strategy (run-phase verification)

Per `.claude/rules/moai/workflow/verification-batch-pattern.md`, the orchestrator's post-run verification SHOULD batch these 7 commands in a single response turn (canonical AC-WO-007 pattern):

```bash
# 1. Functional — test suite
go test ./internal/harness/proposalgen/... ./internal/cli/harness/...

# 2. Coverage
go test -coverprofile=cover.out ./internal/harness/proposalgen/... ./internal/cli/harness/...

# 3. Subagent-boundary grep (AC-PGN-005)
grep -rn 'AskUserQuestion(' internal/cli/harness/ | grep -v '_test.go'

# 4. Sentinel/frontmatter audit (cross-SPEC sanity)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20

# 5. CLI smoke (AC-PGN-004)
go run ./cmd/moai harness propose --dry-run

# 6. Lint baseline
golangci-lint run --timeout=2m

# 7. go vet
go vet ./...
```

All 7 commands are read-only and independent; safe to batch per § Verification Class Taxonomy in verification-batch-pattern.md.

## §G. AC Count Verification (B12 self-test)

Pre-emission: `grep -cE '^\| \*\*AC-PGN-[0-9]+\*\* \| ' acceptance.md` → expected count = **8** (mandatory only). The pattern `\*\* \| ` requires the bold-closed AC ID to be immediately followed by ` | ` (space-pipe-space) — this matches the §A mandatory row format `| **AC-PGN-NNN** | <description> | ...` and excludes the §B AC-PGN-009 row whose AC ID is followed by ` (Optional) |` instead.

Verification: an alternate broader regex `grep -cE '^\| \*\*AC-PGN-[0-9]+\*\*' acceptance.md` returns **9** (includes AC-PGN-009 Optional). Sync-phase manager-docs should use the tightened regex (`\*\* \| ` suffix) when counting mandatory ACs to compare against the SPEC's DoD criteria; the broader regex may be used to count ALL ACs (mandatory + optional) when reporting total AC coverage.

This pattern was selected over alternates `[^(]` (rejected — would still match because the AC ID is followed by a space, not `(`, regardless of mandatory/optional status) and AC-PGN-009 row format mutation (rejected — preserves row format consistency).
