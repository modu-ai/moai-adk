# acceptance.md — SPEC-AUDIT-SNAPSHOT-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-AUDIT-SNAPSHOT-NNN); this file does NOT restate them as requirements.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-AUDIT-SNAPSHOT-001 | REQ-AUDIT-SNAPSHOT-001 | Sticky cache — past-24h unchanged hash | MUST | plan M1 |
| AC-AUDIT-SNAPSHOT-002 | REQ-AUDIT-SNAPSHOT-002 | Per-tier skip threshold (0.78 Tier M) | MUST | plan M1 |
| AC-AUDIT-SNAPSHOT-003 | REQ-AUDIT-SNAPSHOT-003 | Clean sync → binding verdict, no cold spawn | MUST | plan M2 |
| AC-AUDIT-SNAPSHOT-004 | REQ-AUDIT-SNAPSHOT-004 | Shared snapshot, 3 consumers, SHA-invalidation | MUST | plan M3 |

### §D.1 Severity model

All four ACs are MUST-pass. There are no SHOULD or NICE-TO-HAVE ACs in this SPEC — each maps to a P0 redesign item where partial implementation would leave the audit pipeline in a worse state than the status quo (e.g., A1 without A2 still over-spawns; A3 without A4 still over-executes test suites).

### §D.2 AC definitions (Given-When-Then)

#### AC-AUDIT-SNAPSHOT-001 — Sticky cache survives past-24h (A1)

**Given** a SPEC whose plan-phase audit produced a verdict with overall score ≥ the SPEC's per-tier PASS threshold, AND the plan-artifact hash recorded at verdict time is `H1`, AND the current time is MORE than 24 hours after the verdict was recorded.

**When** the orchestrator consults the plan-audit verdict cache for this SPEC at the current time, AND the current plan-artifact hash is still `H1` (unchanged).

**Then** the cache returns the verdict as VALID (skip-eligible), AND the skip decision rationale records the hash-match as the validity basis with NO "Within 24h" time-window check applied.

**Test shape:** unit test in `internal/runtime/audit_cache_test.go` — arm a verdict with `recordedAt = now - 25h`, advance the clock past 24h, call the cache consult with an unchanged hash, assert the verdict is returned as valid; assert no time-window predicate is evaluated.

#### AC-AUDIT-SNAPSHOT-001b — Tier L hash subject extension (A1 Tier L)

**Given** a SPEC with `tier: L` whose `design.md` and `research.md` are present and non-empty.

**When** `ComputeHash` is invoked on the SPEC directory.

**Then** the resulting hash includes contributions from `design.md` AND `research.md` in addition to the legacy `{spec.md, plan.md, acceptance.md, tasks.md}` set, AND modifying `design.md` produces a different hash than the unmodified case.

**Test shape:** unit test — compute hash, mutate `design.md`, re-compute, assert inequality; same for `research.md`. Also assert `tasks.md` remains in the subject set for a grandfathered fixture SPEC.

#### AC-AUDIT-SNAPSHOT-002 — Per-tier skip threshold (A2)

**Given** a SPEC with `tier: M` whose plan-phase audit produced an overall score of 0.78.

**When** the orchestrator evaluates the plan-audit skip-policy condition 2 (overall score threshold) for this SPEC.

**Then** the condition evaluates to SATISFIED (0.78 ≥ 0.80 is FALSE — wait, re-check: for Tier M the PASS threshold is 0.80, so 0.78 is NOT skip-eligible). Correction: use a score of 0.81 for Tier M, OR use Tier S with a score of 0.78 (0.78 ≥ 0.75 → satisfied).

**Revised Given:** a SPEC with `tier: S` whose plan-phase audit produced an overall score of 0.78.

**Revised Then:** the skip condition 2 evaluates to SATISFIED (0.78 ≥ 0.75 — the Tier S PASS threshold), AND the same 0.78 score against a `tier: M` SPEC evaluates to NOT SATISFIED (0.78 < 0.80), AND the same score against `tier: L` evaluates to NOT SATISFIED (0.78 < 0.85).

**Test shape:** table-driven unit test over (tier, score) pairs asserting skip-condition-2 truth values match the per-tier threshold table. Explicitly assert the retired flat `≥ 0.90` predicate is NOT consulted (e.g., a 0.81 Tier M SPEC is skip-eligible even though 0.81 < 0.90).

#### AC-AUDIT-SNAPSHOT-003 — Clean sync emits binding verdict without cold sync-auditor spawn (A3)

**Given** a sync cycle whose `.claude/workflows/sync-audit-4dim.js` workflow run returns a verdict of PASS with all four dimensions scoring above their must-pass floor AND the verdict is not `INCOMPLETE` AND the orchestrator raises no contested-finding flag.

**When** the orchestrator consumes the workflow verdict to make the sync-phase quality decision.

**Then** the orchestrator treats the workflow verdict as BINDING (the sync-phase quality decision is PASS), AND the orchestrator DOES NOT spawn the cold sync-auditor subagent, AND the audit-trail record for this sync cycle cites the 4-dim workflow run ID (not a sync-auditor agent ID) as the verdict source.

**Test shape:** workflow-run fixture returning 4 passing dimension scores + harmonic-mean PASS + not-INCOMPLETE; assert the orchestrator's sync-phase decision path takes the binding branch and the spawn-counter for the cold sync-auditor is 0 for this cycle.

#### AC-AUDIT-SNAPSHOT-003b — Failure mode still spawns cold sync-auditor (A3 fallback)

**Given** a sync cycle whose 4-dim workflow run returns `INCOMPLETE`, OR any must-pass dimension scores 0, OR a **contested finding** is detected.

**Contested finding** is defined mechanically (machine-evaluable from the structured per-judge output the workflow already emits): (i) any one of the 4 parallel judges reports a finding at `critical` severity, OR (ii) two or more judges return conflicting severity classifications for the same dimension (e.g. one judge marks Functionality `critical` while another marks it `minor` for the same SPEC). No orchestrator judgment is required.

**When** the orchestrator consumes the workflow verdict.

**Then** the orchestrator spawns the cold sync-auditor subagent as the fallback binding-verdict owner, AND the cold auditor's PASS/FAIL verdict is treated as binding for that cycle.

**Test shape:** three failure-mode fixtures: (a) INCOMPLETE verdict, (b) a must-pass dimension scoring 0, (c) a contested-finding fixture constructed by emitting judge outputs where one judge assigns `critical` to a dimension the other judges mark `minor` (predicate (ii)); optionally a fourth fixture exercising predicate (i) (a single judge emitting a `critical` finding). For each, assert the cold sync-auditor spawn count is 1 and the auditor's verdict is the one recorded.

#### AC-AUDIT-SNAPSHOT-004 — Single snapshot shared across 3 consumers (A4)

**Given** a sync cycle at HEAD SHA `S1` where no prior diagnostic snapshot exists for `S1`, AND three consumers will request test/lint/vet/coverage results during this cycle: the sync-auditor Evidence cell, the `sync-phase-quality-gate.sh` Stop hook, and the 4-dim workflow judges.

**When** the first consumer requests the snapshot for `S1`.

**Then** a single fresh recording is triggered (exactly one `go test` run, one `golangci-lint` run, one `go vet` run, one `go test -cover` run), AND the result is keyed by `S1` in the snapshot store.

**And when** the second and third consumers request the snapshot for `S1`.

**Then** both consumers read the recorded result WITHOUT triggering a re-execution, AND the total execution count across the sync cycle is exactly 1 per diagnostic dimension (not 3-4× as in the status quo).

**Test shape:** snapshot-recording fixture — install a counting wrapper around each diagnostic command, run the 3 consumers in sequence (and in a parallel variant for the concurrency case), assert the per-dimension execution count is 1; assert each consumer's Evidence cell carries the recorded exit code + output identical to the recording. The parallel variant also exercises the claim/lock serialization mandated by REQ-AUDIT-SNAPSHOT-004 paragraph 4 — two consumers concurrently invoking `RecordCheck` on the same SHA with different command dimensions MUST NOT race last-writer-wins (one claim wins, the other reads); assert both dimensions end up recorded rather than one silently dropping the other.

#### AC-AUDIT-SNAPSHOT-004b — HEAD SHA change invalidates the snapshot (A4 integrity)

**Given** a snapshot recorded for HEAD SHA `S1`, AND a new commit has landed so the current HEAD SHA is `S2` (≠ `S1`).

**When** any consumer requests the snapshot for the current HEAD (`S2`).

**Then** the consumer does NOT receive the `S1`-recorded result, AND the consumer either (a) triggers a fresh recording for `S2`, or (b) surfaces an explicit error indicating the snapshot is absent for `S2`. In neither case does the consumer silently serve the stale `S1` result.

**Test shape:** record a snapshot at `S1`, advance HEAD to `S2`, request the snapshot, assert it is not the `S1` result; assert the consumer path either re-records or errors explicitly (per the consumer's documented contract), never silent stale service.

### §D.3 Indirect verification

The audit-semantics-immutability constraint (spec.md §D.6) is verified indirectly: the existing plan-auditor and sync-auditor test suites (`internal/runtime/audit_*_test.go`, any sync-auditor AC tests) MUST continue to pass unchanged. A regression in those suites indicates this SPEC accidentally changed WHAT is measured, not just WHEN/HOW OFTEN.

### §D.4 Closure gates

- All four MUST ACs green with verbatim command + observed output evidence pinned to the test run.
- K-1 closure: `grep -rn "Within 24h\|24 hour\|score >= 0.90\|≥ 0.90" .claude/ internal/` returns no restatement outside the SPEC's own `spec.md`/`plan.md`/`acceptance.md` (the SPEC documents the retirement; it does not re-apply it).
- K-5 closure: the sync-phase quality audit-trail record schema accepts BOTH the workflow run ID and the sync-auditor agent ID as verdict-source values (provenance compatibility).
- LSP gate: zero errors, zero type errors, lint clean (per the project's standard quality gate).

### §D.5 Forward-looking checks (advisory, non-blocking for this SPEC)

- The snapshot store's growth — if snapshots are persisted to disk rather than memory, confirm a retention/cleanup path exists so the store does not grow unbounded across sync cycles. (Belongs to a future hygiene SPEC if absent.)
- The 4-dim workflow verdict schema — if OQ-1 finds a must-pass-dim-0 signal is missing and M2 adds one, the schema augmentation should be documented for the future MCP-surface SPEC (epic P1) so that surface does not re-derive the signal.

### §D.6 Definition of Done

This SPEC is DONE when:

1. All four MUST ACs (001, 001b, 002, 003, 003b, 004, 004b) pass with attributed evidence.
2. The K-1 residual-restatement sweep is clean.
3. The project's standard quality gate (lint + type-check + test + race) is green.
4. The spec.md frontmatter `status` transition `draft → in-progress → implemented → completed` is owned by manager-develop and manager-docs per the Status Transition Ownership Matrix; this plan-phase authoring only emits `draft`.
