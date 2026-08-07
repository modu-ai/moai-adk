# acceptance.md — SPEC-SYNC-PARALLEL-DOCS-001

> Given-When-Then acceptance criteria. Each AC is binary-testable and labeled `AC-SPD-XXX`. Per the spec-lint two-layer rule, acceptance criteria use GWT format (NOT GEARS); the GEARS obligation binds the `REQ-SPD-*` entries in `spec.md` only.

## §A. AC Matrix

### A.5 — docs ∥ audit parallelization

#### AC-SPD-001 — Docs drafter fan-out concurrent with Phase 7 audit

**Given** the sync scope spans several independent document families (the `FO-SYNC-4` 5-drafter condition, `doc-execution.md` L126)
**When** the orchestrator enters Phase 7 (Quality Verification)
**Then** a read-only docs drafter fan-out launches in the SAME turn as the Phase 7 audit fan-out (not serially after the audit completes)
**And** the concurrency is observable in the progress.md §F mode-selection log or the sync-phase trace (`MOAI_TRACE_PHASES=1`).

#### AC-SPD-002 — Docs drafter input independence

**Given** the docs drafter fan-out is running concurrently with the Phase 7-10 audit
**When** a docs drafter (D1-D5 per `doc-execution.md` L128-134) reads its input
**Then** it reads SPEC artifacts + git diff + the Phase 11 Step 1.5 divergence report
**And** it does NOT read the concurrent audit's quality report, verdict, or per-dimension scores.

#### AC-SPD-003 — gate-sync-2 merge with single-writer applier

**Given** both the docs draft and the audit verdict are produced by the concurrent fan-outs
**When** `gate-sync-2` (HUMAN GATE 2: Documentation Scope) fires
**Then** `manager-docs` applies the five drafts sequentially as the sole write-capable agent
**And** at no point during the concurrent fan-out do two write-capable agents run simultaneously
**And** the audit verdict is surfaced to the user at the same gate (no extra human round-trip).

### A.7 — MX Tag early + parallel

#### AC-SPD-004 — MX Tag scan concurrent with Phase 7 audit

**Given** the orchestrator enters Phase 7 (Quality Verification)
**When** the Phase 7 audit fan-out launches
**Then** the Phase 9 MX Tag scan (existing `FO-SYNC-2` sharded structure, `quality-gates-quality.md` L186) launches concurrently in the same turn
**And** the scan does NOT wait for Phase 8 (Security) to complete.

#### AC-SPD-005 — P1/P2 violations halt before coverage execution

**Given** the concurrent MX scan detects P1 (missing `@MX:ANCHOR` on fan_in≥3 exported function) or P2 (missing `@MX:WARN` on goroutine/async pattern) violations
**When** the orchestrator checks the scan result
**Then** sync halts BEFORE Phase 10 (Coverage Analysis) executes
**And** the "30-min coverage then 1 missing tag aborts all" worst case is eliminated (the coverage command never runs).

#### AC-SPD-006 — No P1/P2 violations → Phase 10 coverage proceeds (no regression)

**Given** the concurrent MX scan detects zero P1/P2 violations (only P3/P4 advisory findings, or no findings)
**When** the orchestrator proceeds past the MX gate
**Then** Phase 10 (Coverage Analysis) executes exactly as it does today
**And** A7 introduces no false aborts (the coverage phase is not blocked by A7).

### A.9 — §E + 7-batch integration (attributable diff-check)

#### AC-SPD-007 — §E evidence carries attribution per VCI §2

**Given** `manager-develop` reports run-phase completion with §E self-verification
**When** the orchestrator inspects the §E evidence (E1-E8 items per `manager-develop-prompt-template.md` § Section E)
**Then** each §E item names (a) the command that was run, (b) the observed verbatim output, (c) the baseline-attribution (this run, this tree)
**And** the §E item content satisfies `verification-claim-integrity.md` §2 (Attribution: command + observed output + baseline).

#### AC-SPD-008 — Attributable diff-check consumed (no re-execution)

**Given** §E evidence cites a snapshot key matching the current HEAD SHA AND the §E-cited command matches the snapshot's recorded command AND the §E-cited output matches the snapshot's recorded output
**When** the orchestrator trust-but-verify batch (`agent-common-protocol.md` § Parallel Execution canonical 7-command batch) runs
**Then** the batch consumes the attributable §E evidence for that dimension
**And** the batch does NOT re-execute the corresponding command (test/lint/vet/cover)
**And** the batch records the snapshot key + cited evidence path as its baseline-attribution per VCI §2.

#### AC-SPD-009 — Diff-check fallback to re-execution on any mismatch

**Given** any of the three diff-check matches fails — (a) snapshot key mismatch (HEAD SHA changed since §E recording), OR (b) §E-cited command does not match the snapshot's recorded command, OR (c) §E evidence is missing or cites no observable output
**When** the orchestrator trust-but-verify batch detects the mismatch
**Then** the batch falls back to re-execution of the affected verification dimension
**And** the fallback is logged with the mismatch reason (snapshot key drift / command drift / missing §E)
**And** the batch NEVER silently skips verification (the `verification-claim-integrity.md` §1.1 invariant holds).

### A.6 — plan-auditor retry Tier ceilings

#### AC-SPD-010 — Tier S plan-auditor retry ceiling = 1

**Given** a SPEC with `tier: S` in frontmatter
**When** the plan-auditor Retry Loop Contract (`plan-auditor.md` § Retry Loop Contract L386-418) is consulted
**Then** the effective retry ceiling is 1 (single spawn, no iteration 2+)
**And** on a FAIL verdict at iteration 1, the orchestrator escalates via AskUserQuestion (3-option path preserved).

#### AC-SPD-011 — Tier M plan-auditor retry ceiling = 2

**Given** a SPEC with `tier: M` in frontmatter
**When** the plan-auditor Retry Loop Contract is consulted
**Then** the effective retry ceiling is 2 (up to two spawns)
**And** on a FAIL verdict at iteration 2, the orchestrator escalates via AskUserQuestion.

#### AC-SPD-012 — Tier L plan-auditor retry ceiling = 3 (legacy fallback)

**Given** a SPEC with `tier: L` in frontmatter OR a SPEC with no `tier:` field (backward-compat Tier L treatment per § SPEC Complexity Tier)
**When** the plan-auditor Retry Loop Contract is consulted
**Then** the effective retry ceiling is 3 (up to three spawns — the legacy behavior)
**And** pre-A6 SPECs without `tier:` see no behavior change.

### Cross-cutting

#### AC-SPD-013 — Concurrency guard preserved (no two write-capable agents)

**Given** the A5 docs drafter fan-out and the A7 MX scan run concurrently with the Phase 7 audit
**When** any concurrent agent is inspected (D1-D5 drafters, FO-SYNC-2 MX shards, FO-SYNC-1 4-dim judges, sync-auditor fallback)
**Then** every concurrent agent is read-only (drafters return draft text; shards return findings; judges return structured verdicts; auditors read tree state)
**And** the single write-capable pass (manager-docs applying the docs drafts at gate-sync-2) runs AFTER both fan-outs return
**And** the `[HARD]` concurrency guard (`agent-common-protocol.md` § Background Agent Execution) holds throughout.

#### AC-SPD-014 — Audit semantics unchanged (WHEN/HOW OFTEN, not WHAT)

**Given** A5/A7/A9/A6 are applied to a sync cycle
**When** the audit dimensions are evaluated (Functionality 40%, Security 25%, Craft 20%, Consistency 15% per `quality-gates-quality.md` L70-73)
**Then** the 4-dimension weights, the PASS thresholds (Tier S 0.75 / M 0.80 / L 0.85), the severity definitions (critical/warning/suggestion), and the AC content are identical to the pre-A5/A7/A9/A6 baseline
**And** only the scheduling (concurrent vs serial), the iteration ceiling (Tier-aware vs flat), and the orchestrator consumption mode (diff-check vs re-execution) are changed.

## §B. Edge Cases

- **EC-1 — Audit verdict INCOMPLETE / contested**: AUDIT-SNAPSHOT-001 REQ-AUDIT-SNAPSHOT-003 fallback to cold sync-auditor still applies. A5's concurrent docs draft is unaffected (input-independent per REQ-SPD-002); the cold auditor runs after the concurrent audit returns. The single-writer applier at gate-sync-2 still sequences correctly.
- **EC-2 — §E evidence drift between §E recording and orchestrator batch**: a commit lands between manager-develop §E and orchestrator verification. AC-SPD-009 fallback fires (snapshot key mismatch); the orchestrator re-executes the affected dimension. The drift is logged, not silently served.
- **EC-3 — P3/P4 MX findings during concurrent scan**: REQ-SPD-005 binds P1/P2 only; P3 (long exported function missing `@MX:NOTE`) and P4 (untested public function missing `@MX:TODO`) remain advisory and do NOT trigger the pre-coverage halt. Phase 10 coverage proceeds (AC-SPD-006).
- **EC-4 — Tier field absent (legacy SPEC)**: REQ-SPD-010 fallback path — `tier:` absent → treated as Tier L → ceiling 3 (AC-SPD-012). Pre-A6 SPECs see no behavior change.
- **EC-5 — Concurrent fan-out input-blocker**: a docs drafter or MX shard missing a required input returns a structured blocker report (`agent-common-protocol.md` § Blocker Report Format); the orchestrator re-delegates that one item while the other shards' results stand. The concurrency contract is not violated by a blocker return.
- **EC-6 — gate-sync-2 user-abort**: if the user selects "Abort" at gate-sync-2 (the existing 4-option gate), the sync terminates with no doc application; the audit verdict is still recorded in `.moai/reports/`. A5 adds no new abort semantics.

## §C. Quality Gate Criteria

- **Coverage**: no Go code changes required for A5/A7 (workflow-skill prose edits); A9 may add a doctrinal clause to `agent-common-protocol.md` (no Go code) unless plan-auditor resolves OQ-1 toward a mechanical hook. A6 may add a per-Tier map to `harness.yaml` (YAML only). Coverage target (85%) applies to any Go change that M1/M2 introduce (currently none anticipated).
- **Lint**: `moai spec lint --strict .moai/specs/SPEC-SYNC-PARALLEL-DOCS-001/` MUST return 0 errors. `golangci-lint run` baseline MUST NOT regress (no new findings in any Go file touched by M1/M2, if any).
- **Cross-platform**: no Go build-tag changes; no cross-platform impact.
- **Subagent boundary**: no `AskUserQuestion` calls in any A5/A7 fan-out agent or in `manager-develop` §E machinery. Grep: `grep -rn 'AskUserQuestion' .claude/skills/moai/workflows/sync/ | grep -v '^.*#'` returns no NEW matches introduced by this SPEC.

## §D. Definition of Done

- [ ] All 14 AC-SPD-* PASS with attributable evidence (command + observed output + baseline per VCI §2).
- [ ] `moai spec lint --strict` on this SPEC directory returns 0 errors.
- [ ] `moai spec audit --json` classifies this SPEC V3R6 (H-4 predicate: §E.2 + §E.4 + sync_commit_sha).
- [ ] No regression in any existing sync-phase flow (audit semantics unchanged — AC-SPD-014).
- [ ] Concurrency guard `[HARD]` verified: no two write-capable agents run concurrently in any A5/A7 fan-out path (AC-SPD-013).
- [ ] A9 fallback to re-execution verified on snapshot-key drift (AC-SPD-009).
- [ ] A6 Tier ceilings verified for S/M/L (AC-SPD-010 / AC-SPD-011 / AC-SPD-012) + legacy fallback (no `tier:` → ceiling 3).
