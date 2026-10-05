---
id: SPEC-AC-BASELINE-REFRESH-001
title: "Acceptance — durable AC-count corpus baseline refresh"
version: "0.1.0"
created: 2026-09-22
author: lane agent-20 (t1068)
---

# Acceptance — SPEC-AC-BASELINE-REFRESH-001

Every criterion is observable by command output or file inspection. Baseline attribution per `verification-claim-integrity.md` §2: each PASS names the command and the observed output, measured in the claiming run, against the tree under measurement.

## §1 Requirements → Criteria coverage

| REQ | AC(s) |
|---|---|
| REQ-ABR-001 (regeneration mode, default-off) | AC-ABR-001, AC-ABR-002 |
| REQ-ABR-002 (provenance header) | AC-ABR-002 |
| REQ-ABR-003 (remedy in failure output) | AC-ABR-003 |
| REQ-ABR-004 (tracked cascade procedure doc) | AC-ABR-004 |
| REQ-ABR-005 (named-cause cascade commits) | AC-ABR-004, AC-ABR-007 |
| REQ-ABR-006 (judgment semantics untouched) | AC-ABR-006 |
| REQ-ABR-007 (mutation preservation) | **AC-ABR-005** |

## §2 Criteria

### AC-ABR-001 — Regeneration mode exists and is default-off

**Given** the run-phase tree with M1 landed, **When** `go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1` runs WITHOUT `MOAI_AC_BASELINE_REGENERATE`, **Then** it exits 0 and the snapshot file's content hash is unchanged (no write). **When** the same command runs WITH `MOAI_AC_BASELINE_REGENERATE=1`, **Then** it exits 0 and the snapshot at `.moai/reports/t338/ac-count-baseline.txt` is rewritten in place, with `grep -c '^\.moai'` on the result equal to the live corpus match count (798 at the 2026-09-22 post-authoring measurement — recount at generation, never carry; the population moves with every authored `acceptance.md`) and every line parseable by `parseACBaseline`.

### AC-ABR-002 — Provenance header and format round-trip

**Given** the regenerated snapshot, **Then** its header carries: (a) the frozen corpus glob statement verbatim (`depth-1 glob .moai/specs/*/acceptance.md`, `_archive` excluded), (b) the tree commit SHA equal to `git rev-parse --short HEAD` at generation time, (c) the regeneration date, (d) the exact regeneration command; **and** the dead-recipe string `run-scratch/gen-baseline.sh` appears nowhere on the surfaces this SPEC OWNS — the regenerated snapshot header and the new procedure document `.moai/docs/ac-count-baseline-refresh.md` (`grep -n 'run-scratch/gen-baseline'` over those two files → 0 rows). **Scope note (plan-audit iter-1 D2):** two further tracked hits survive BY DESIGN and are outside this SPEC's edit rights — `.moai/reports/t348/verdict.md:15` (a completed card's historical verdict record; rewriting observed evidence is prohibited) and `SPEC-AC-COUNT-DISCRIMINATOR-001/progress.md:374` (a completed SPEC's progress body; ownership-barred to manager-develop and manager-docs). The criterion's intent is that no LIVE instruction points at the dead recipe; historical evidence records are not live instructions and are not remediated by this SPEC. **And** the emitter's output fed through `parseACBaseline` yields one entry per emitted line with equal live/excluded values (unit-asserted on a fixture writer).

### AC-ABR-003 — The remedy rides the failure

**Given** the refreshed snapshot, **When** a recorded `acceptance.md` is deleted without a cascade (the AC-ABR-005 mutation), **Then** the failing test output contains the in-tree regeneration command (`MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate`). The two other recorded-file failure sites (halting problem, count/state problem at `acComparison` call sites) carry the same remedy suffix — verified by code inspection recorded in the run evidence.

### AC-ABR-004 — Tracked cascade procedure document

**Given** M3 landed, **Then** `.moai/docs/ac-count-baseline-refresh.md` exists and `git ls-files .moai/docs/ac-count-baseline-refresh.md` resolves (tracked, not scratch); **and** the document names all three trigger events (superseded/split `acceptance.md` removal, `_archive/` directory move, count-affecting corpus rewrite), the same-commit rule, the diff-review discipline, and the named-cause commit-message rule; **and** references to it resolve from the snapshot header, the test file's doc comment, and `internal/spec/CLAUDE.md` (`git grep -l 'ac-count-baseline-refresh'` covers all three surfaces).

### AC-ABR-005 — MUTATION VERIFICATION: a real vanish still fails (card [HARD] 5)

**Given** the refreshed snapshot (M2 complete), **When** `.moai/specs/SPEC-MODEL-MATRIX-CORE-001/acceptance.md` is removed from the working tree and `go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1` runs, **Then** the test FAILS with a line naming exactly that file at the vanish site (`present in the snapshot but no longer matched by the corpus glob`) AND the remedy command — the gate is not hollow. **When** the file is restored from `/tmp` backup, **Then** the same command exits 0. **And** `go test ./internal/spec -run TestACBaselineComparisonTransitions -count=1` exits 0 both before and after the mutation — the v0.5.0 absence-first narrowing (`absent-and-counts` / `absent-and-halts` rows report-not-fail) survives the disposition intact.

### AC-ABR-006 — Judgment-semantics diff audit

**Given** the completed implementation, **Then** `git diff <pre-SPEC-HEAD> -- internal/spec/ac_count_clause_test.go` shows: `acComparison`'s body unchanged; the corpus test's error conditions unchanged (`!known` → report-not-fail, `!seen[rel]` → error, live/excluded comparison → error); additions confined to the regeneration mode, the message remedy suffixes, the doc comment, and the header emission. Verified by reading the diff and recording the inspection in run evidence — not asserted from memory.

### AC-ABR-007 — Cascade commit discipline demonstrated

**Given** the M2 catch-up commit, **Then** `git show --stat <cascade-commit>` lists only `.moai/reports/t338/ac-count-baseline.txt`, and the commit message names a cause per changed-line class (the N additions — every line attributable to a SPEC directory absent from the old snapshot; the 1 removal with its origin commit `20cdeb6bd`; the header format). Any count or state change in the diff whose cause the message cannot name fails this criterion. This attribution predicate — not any expected total — is M2's only stop rule.

### AC-ABR-008 — Affected-package and consumer green

**Then** `go test ./internal/spec -count=1` exits 0, and `go test ./internal/cli -run 'TestTodoTriage' -count=1` exits 0 (the snapshot path — its only coupling — is unchanged). Full-suite and platform-matrix verdicts belong to CI on the pushed head and are not claimed locally.
