---
id: SPEC-OBSERVE-HYGIENE-001
title: "Observation Sink Hygiene — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "observation-sinks, log-hygiene, pruning, sync-gate, dormancy-annotation, workflow-reflex, plan"
---

# SPEC-OBSERVE-HYGIENE-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. Tier S minimal envelope; function names/signatures are run-phase discretion.

## §A Context

### §A.1 Problem summary

Six observation sink groups write with no reader, or promise behavior that never fires: status-transition-audit.log (~389KB, 0 consumers), fact-force-skip.log (~73KB), task-metrics.jsonl (stale since 2026-05-30), 314 zero-byte trace files (no age pruning), a sync gate whose deterministic vet/build checks block only behind an env var nobody sets, an 8-reference loop.md snapshot/resume promise with a never-materialized directory, and two dormant config blocks with zero in-file annotation. Disposition per sink: consumer / documented-write-only / prune.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
ls -la .moai/logs/                              → status-transition-audit.log 397,935 B; fact-force-skip.log 75,128 B;
                                                  task-metrics.jsonl 467,600 B (mtime 2026-05-30 12:56)
ls .moai/logs | grep -c trace-                  → 643;  find -size 0 → 314 zero-byte
grep -rn "status-transition-audit" --include='*.go' internal/ cmd/  → 0 (no Go consumer)
status-transition-ownership.sh:68               → "Advisory hook (never blocks; exit 2 reserved for future ...)" ; append at :79
sync-phase-quality-gate.sh:218-220,325          → go vet + go build steps; MOAI_SYNC_GATE_BLOCKING:-0 opt-in
internal/hook/trace/writer.go                   → size-based rotation only (REQ-OBS-006); no age pruning
internal/hook/post_tool.go:166-169              → task-metrics writer (Agent/Task subagent metrics, SPEC-MONITOR-001)
ls .moai/cache/loop-snapshots                   → No such file or directory; loop.md refs at 5,60,75,82,150,198,201,257
sunset.yaml                                     → zero comments; DORMANT per audit_loader_completeness_test.go:23 (REQ-MIG003-006)
                                                  + loader.go:97 once-per-session DORMANT notice (REQ-MIG003-018)
harness.yaml:77 model_upgrade_review            → consumer emitModelUpgradeReminder (harness_validate.go, REQ-HRN-001-016)
                                                  gated on CLAUDE_MODEL_PREVIOUS/CLAUDE_MODEL — no setter found (dormant-in-practice)
moai spec audit                                 → exists (internal/cli/spec_audit.go; audit engine internal/spec/audit.go)
```

Template mirrors verified present (2026-07-09): sunset.yaml, harness.yaml, gateguard-fact-force.sh, sync-phase-quality-gate.sh, status-transition-ownership.sh, loop.md. Byte sizes/counts drift with use — re-measure at run-phase; content anchors authoritative.

### §A.3 Approach — three milestones

- **M1 — spec-audit log consumer (Go, TDD)**: extend the `moai spec audit` path to parse status-transition-audit.log lines, correlate transition sites against the Status Transition Ownership Matrix, and emit INFO findings for mismatches; graceful no-op on absent/corrupt log; table-driven tests with fixture logs.
- **M2 — telemetry pruning (Go, TDD)**: SessionEnd-path pruning of zero-byte `trace-*.jsonl` + age-based retention for non-empty traces (D2 threshold); task-metrics.jsonl root-cause note + disposition per D2; tests under `t.TempDir()`.
- **M3 — decisions + annotations + doc markers + template sync**: sync-gate blocking-default per D3 (vet/build only); fact-force header write-only line; loop.md best-effort marker; sunset.yaml + harness.yaml dormancy comments; template-first mirrors + `make build`.

### §A.4 Tier evidence (S)

- LOC estimate: ~200-280 total (small parser + pruner + tests; everything else is comments/prose) — inside the Tier S <300 LOC band.
- Files: ~7 logical surfaces; counting template mirrors and `_test.go` files the touch count exceeds the nominal <5 band. Recorded honestly per the Tier convention: each change is surgical, no architectural movement, no constitutional surface → S with rationale (Epic brief mandates Tier S).
- Risk: LOW-MED — the only behavior change users can feel is the D3 gate promotion, which is why it is a Kickoff decision point.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| status-transition-ownership.sh (producer, reserved exit-2) | PRESERVE (log format is now a consumer contract — note added at most) |
| internal/spec audit engine + spec_audit.go | EXTEND (new cross-check input, INFO findings only) |
| internal/hook session-end path | EXTEND (pruning step) |
| internal/hook/trace/writer.go size rotation | PRESERVE (age pruning is additive, elsewhere) |
| sync-phase-quality-gate.sh signaling (stdout JSON, exit-0 block channel) | PRESERVE; EXTEND default for vet/build per D3 |
| runtime-recovery-doctrine.md §4 carve-out | PRESERVE (cited invariant) |
| loop.md snapshot sections | EXTEND (marker only; coordinate with SPEC-LOOP-VERDICT sibling) |
| sunset.yaml / harness.yaml semantics | PRESERVE (comment-only edits) |
| .moai/state/verify (TOKEN-BUDGET-STOP) / .moai/harness/* (RATCHET-REWIRE) | UNTOUCHED (sibling scope) |

## §B Known Issues (filtered, Tier S)

- **B1 Cross-platform**: pruning code touches the filesystem — verify `GOOS=windows` build; use filepath-safe APIs.
- **B8 Working-tree hygiene (CRITICAL here)**: `.moai/logs/*` are runtime-managed and actively written DURING the run-phase session (fact-force-skip.log grew mid-audit). Never commit log files; pruning tests operate only in `t.TempDir()`; specific-path `git add` only.
- **B10 Scope discipline / sibling collision**: loop.md is SPEC-LOOP-VERDICT-CONTRACT-001's edit surface (Steps 1/4/9) — sequence the M3 marker hunk after that SPEC's run-phase or coordinate; SPEC-ADVISOR-RUNG-001 also queues a loop.md hunk (three-way surface — orchestrator sequences).
- **B2 Cross-SPEC policy**: REQ-MIG003-006/-018 (sunset dormancy) and REQ-OBS-006 (trace rotation) are owned by closed SPECs — annotations cite them, never re-legislate them. SPEC-HOOK-FACTFORCE-ADVISORY-001 owns fact-force advisory semantics.
- **B5 CI tiers**: new Go code hits spec-lint / golangci-lint / per-OS test jobs independently.

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/spec/... ./internal/hook/... 2>&1 | tail -5        # baseline
ls -la .moai/logs/ | head -8                                           # re-measure sink sizes (drift expected)
grep -n "MOAI_SYNC_GATE_BLOCKING" .claude/hooks/moai/sync-phase-quality-gate.sh
git log --oneline -3 -- .claude/skills/moai/workflows/loop.md          # sibling sequencing check
golangci-lint run --timeout=2m 2>&1 | tail -5
```

## §D Constraints + open decisions

Constraints: see spec.md §Constraints (sibling boundaries; hook contract preservation; comment-only config edits; document-or-prune posture).

**D1 — status-transition-audit.log disposition (REQ-OBH-001).** Both directions:

| Option | Change set | Rationale | Risk |
|--------|-----------|-----------|------|
| **(a) spec-audit consumer (RECOMMENDED)** | Parse log in `moai spec audit`; correlate against Ownership Matrix; INFO findings | The matrix already has a lint-side sibling (OwnershipTransitionRule via git history); the log adds the WRITE-SITE view git can't see; INFO severity keeps it non-breaking | Log format becomes a contract — document the line format in the hook header; parser must tolerate historical format drift |
| (b) document write-only | One-line statements in hook header + doctrine | Cheapest | The 389KB unread-promise problem persists; H5 finding stays open |

**D2 — telemetry retention policy + task-metrics disposition (REQ-OBH-002).** Recommended: prune zero-byte traces unconditionally at SessionEnd; age-out non-empty traces older than 30 days (constant in `internal/config/defaults.go` per the no-hardcoding rule). task-metrics.jsonl: run-phase spends a bounded investigation (≤1 finding note) on why writes stopped 2026-05-30 (writer path exists in post_tool.go; suspect: upstream input shape change for Agent tool events); then EITHER repair the writer (if trivial), OR document write-only + include in age-out, OR retire. Default if inconclusive: document + age-out (option 2).

**D3 — sync-gate blocking promotion (REQ-OBH-004).** Both directions:

| Option | Change set | Rationale | Risk |
|--------|-----------|-----------|------|
| **Promote vet/build to default-blocking (RECOMMENDED)** | Gate defaults MODE=blocking for the two Go deterministic steps; env var becomes the opt-OUT (or scope-widener); tests/coverage stay advisory | Deterministic, fast, once-per-commit sentinel-guarded; a broken build blocking a sync commit is pure signal; anti-death-spiral carve-out (§4 SHOULD) still applies on recovery turns | Any new block point can surprise; must keep the stdout-JSON exit-0 signaling and --skip-hook escape intact |
| Keep opt-in, document dormancy | Comment/doctrine note only | Zero behavior change | H6 stays open; the gate remains ornamental |

The orchestrator MUST surface D3 (and D1 if contested) at Implementation Kickoff Approval before M3.

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E (Tier S minimal form), vci 5-section format:
- E1: AC matrix (acceptance.md §D) with verbatim outputs.
- E2: cross-platform builds exit 0 (incl. `make build` after template edits).
- E3: `go test -cover` for touched packages (new code paths ≥85%).
- E5: lint NEW vs baseline; config-semantics bit-identity evidence for M3 comment edits (`go test ./internal/config/...` green, zero loader diffs).
- E6: commit SHAs + push state; D1/D2/D3 decisions recorded.
- E7: blocker report if D3 was not pre-resolved.

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — spec-audit log consumer | audit engine cross-check + fixtures + graceful degradation | REQ-OBH-001 | AC-OBH-001 PASS |
| M2 — telemetry pruning | SessionEnd zero-byte prune + age retention + task-metrics disposition | REQ-OBH-002 | AC-OBH-002 PASS |
| M3 — decisions + annotations + markers + template sync | D3 gate default; fact-force header; loop.md marker; sunset/harness comments; mirrors + make build | REQ-OBH-003..006 | AC-OBH-003..007 PASS |

Dependency note: M3's gate promotion blocks on the D3 user decision; M1/M2 do not. M3's loop.md hunk sequences after/with the sibling loop.md editors (§B B10).

## §G Anti-Patterns (do NOT)

- Committing anything under `.moai/logs/` or pruning the LIVE logs from test code — `t.TempDir()` only.
- Building readers for sinks other than REQ-OBH-001's (scope creep against the document-or-prune posture).
- Activating status-transition exit-2 enforcement "while we're here" — explicitly out of scope.
- Changing sync-gate signaling semantics (exit-2 stdout discard hazard) — only the default mode flips per D3.
- Semantic edits to sunset.yaml/harness.yaml disguised as comments.
- Restating REQ-MIG003/REQ-OBS-006 contracts instead of citing them.

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), progress.md (§E skeleton).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (M1 correlation target) + § OwnershipTransitionRule (the git-history sibling check).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §4 (preserved carve-out cited by D3).
- SPEC-TOKEN-BUDGET-STOP-001 (verify-dir owner) / SPEC-HARNESS-RATCHET-REWIRE-001 (harness owner) — boundary contracts.
- SPEC-LOOP-VERDICT-CONTRACT-001 + SPEC-ADVISOR-RUNG-001 (loop.md surface co-editors — sequencing).
- `internal/spec/audit.go` + `internal/cli/spec_audit.go` (M1 extension base); `internal/hook` session-end path (M2 base).
