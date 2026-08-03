# plan.md — SPEC-AUDIT-SNAPSHOT-001

> Implementation plan. Order is decision-reversibility-first: the decisions most likely to change (snapshot key contract, 4-dim binding semantics) lead; mechanical wiring follows.

## §A. Context

This SPEC is a **redesign codification**, not greenfield. The design authority is `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 items A1-A4. The implementation re-wires existing knobs; it does not introduce new machinery. Four change points span Go runtime, JS workflow, skill YAML/rule doc, and shell hook.

Per the report's P0 decomposition (§3.5 closeout table, row "SPEC-AUDIT-SNAPSHOT"), this single SPEC bundles A1-A4 because each is small in isolation and they share the verification-claim integrity invariant (§3.5 risk callout). Splitting would create 4 SPECs each carrying the same tree-state-binding constraint.

## §B. Known Issues

- **K-1**: The current skip-policy condition 4 ("Within 24h") is documented as authoritative in `spec-workflow.md` AND likely restated in skill YAML / agent body text. Removing it requires sweeping all restatements; a single remaining "Within 24h" mention creates a divergent skip path. (Audit step: `grep -rn "Within 24h\|24 hour\|24h" .claude/ internal/` after the rule-doc edit.)
- **K-2**: The `tasks.md` entry in the legacy hash subject set is a V3R4-era artifact retained for backward compat with grandfathered SPECs. Extending the set to include `design.md` + `research.md` for Tier L must not DROP `tasks.md` for grandfathered SPECs; the extension is tier-conditional, not a flat replacement.
- **K-3**: The 4-dim workflow today emits a structured verdict, but whether that verdict carries a discrete "must-pass-dim-0" signal vs only a harmonic mean is an OQ (OQ-1 in spec.md). If absent, A3 requires a workflow-output augmentation, expanding the change footprint.
- **K-4**: Concurrent consumer reads of the snapshot (Stop hook firing while sync-auditor mid-read, or 4-dim judges racing) may race on the recording step. Need to confirm whether `moai verify check --key-current` is atomic (OQ-2).
- **K-5**: The cold sync-auditor agent currently OWNS the binding PASS/FAIL. Promoting the 4-dim workflow to binding on the happy path changes the audit-trail provenance — sync-phase quality records now cite the workflow run ID, not the auditor agent ID. Anywhere that grep-matches `sync-auditor` as the verdict source needs to also accept the workflow run.

## §C. Pre-flight (read-only reconnaissance — before M1)

1. Read `internal/runtime/audit_cache.go` in full — confirm `planArtifactNames`, `ComputeHash`, `Lookup`, `Store`, and the cache key derivation.
2. Grep `.claude/rules/moai/workflow/spec-workflow.md` for ALL skip-policy restatements (the § Phase Transitions block + any later echoes).
3. Read `.claude/workflows/sync-audit-4dim.js` header + verdict-emission block (lines 5-9, 48-56, and the harmonic-mean section) — resolve OQ-1.
4. Read `.claude/skills/moai/workflows/sync/quality-gates-quality.md` Step 0.5.2 — confirm the snapshot interface signature.
5. Read `.claude/hooks/moai/sync-phase-quality-gate.sh` — identify where `go vet` / `go build` / lint invocations live (report cites L222-335).
6. Grep `internal/verify/` for the snapshot store (`store.go:38` per report §3.6 row) — confirm key schema and atomicity.

## §D. Constraints (recap from spec.md §D — binding on the plan)

1. Snapshot key = HEAD SHA. No stale-SHA service.
2. A3 = attributable diff-check, not deletion. Cold auditor retained as fallback.
3. A4 = wiring change. No parallel snapshot store.
4. Backward compat for minimal harness + grandfathered SPECs.
5. No new CLI surface.
6. Audit semantics (what is measured) immutable.

## §E. Self-Verification (run-phase — what manager-develop must demonstrate)

- `go test ./internal/runtime/...` — A1 hash/cache tests pass.
- `go test ./internal/verify/...` — A4 snapshot tests pass (including SHA-invalidation).
- A workflow-run fixture demonstrating A3 no-spawn behavior (acceptance.md AC-AUDIT-SNAPSHOT-003).
- A sweep grep proving NO residual "Within 24h" / "score ≥ 0.90" restatements remain in `.claude/` or `internal/` (K-1 closure).
- Lint clean, build clean, no type errors.

## §F. Milestones

### Milestone M1 — Hash + skip (A1 + A2)

Highest reversibility: the skip contract is a documented invariant with multiple restatement surfaces; getting the sweep wrong leaves divergent paths.

**Files (expected):**
- `internal/runtime/audit_cache.go` — drop the 24h condition in the cache-consult path; add tier-conditional hash subject extension (`design.md`, `research.md` for Tier L; keep `tasks.md` for grandfathered).
- `internal/runtime/audit_cache_test.go` — new test: past-24h unchanged-hash skip still fires (AC-AUDIT-SNAPSHOT-001); new test: 0.78 Tier M SPEC is skip-eligible (AC-AUDIT-SNAPSHOT-002).
- `.claude/rules/moai/workflow/spec-workflow.md` — § Phase Transitions skip policy: retire condition 4 ("Within 24h"); update condition 2 from `≥ 0.90` to per-tier PASS; document the Tier L hash subject extension. Also update the "Plan-artifact hash subject list (Go verbatim)" note to reflect the tier-conditional extension.
- Sweep: any other surface restating the skip contract (`grep -rn "0.90\|Within 24h\|24 hour" .claude/ internal/`).

**Exit:** A1 + A2 ACs green; no residual restatement.

### Milestone M2 — 4-dim binding promotion (A3)

Second-highest reversibility: the binding semantics determine the audit-trail provenance (K-5). Resolving OQ-1 (must-pass-dim-0 signal) is a precondition — if the workflow must emit a new signal, this milestone expands.

**Files (expected):**
- `.claude/workflows/sync-audit-4dim.js` — emit (or surface) a discrete must-pass-dim-0 flag if OQ-1 finds it absent. Verdict schema otherwise unchanged.
- `.claude/skills/moai/workflows/sync.md` FO-SYNC-1 (lines 60-69) — replace the unconditional cold sync-auditor spawn with: on happy-path verdict (all dims above floor, not INCOMPLETE, no contested flag), treat workflow verdict as BINDING and skip the cold spawn; on failure mode (INCOMPLETE / dim-0 / contested), spawn cold sync-auditor.
- Provenance surface: anywhere the sync-phase quality record cites `sync-auditor` as verdict source — extend to also accept the workflow run ID.

**Exit:** AC-AUDIT-SNAPSHOT-003 green (clean-sync fixture → binding verdict, no cold spawn); failure-mode fixture → cold spawn still fires.

### Milestone M3 — Shared snapshot wiring (A4)

Third reversibility tier: the wiring points are mechanical once the snapshot interface is confirmed. OQ-2 was RESOLVED by plan-auditor iter 1 (`internal/verify/store.go` inspection): `Save` (L57-94) is atomic via rename, but `RecordCheck` (L100-120) is a read-modify-write — concurrent same-SHA writers racing on different command dimensions race last-writer-wins. The claim/lock is therefore MANDATORY, not conditional.

**Files (expected):**
- `.claude/skills/moai/workflows/sync/quality-gates-quality.md` Step 0.5.2 — extend the snapshot mechanism to expose keyed-by-HEAD-SHA consumption to the three consumers (sync-auditor Evidence, Stop hook, 4-dim judges). No new store; extend the existing interface.
- `.claude/agents/moai/sync-auditor.md` § Per-Dimension Mechanical Verification — replace direct `go test` / `golangci-lint`` / `go vet` / `go test -cover` invocations with snapshot-consumption calls keyed by HEAD SHA.
- `.claude/hooks/moai/sync-phase-quality-gate.sh` (L222-335 per report) — replace the synchronous `go vet` + `go build` re-execution with snapshot-consumption; retain explicit re-run as fallback when SHA mismatches (per constraint 1).
- `.claude/workflows/sync-audit-4dim.js` — wire the 4 judges' Evidence reads to the snapshot instead of independent re-runs.
- `internal/verify/store.go` `RecordCheck` (L100-120) — add a claim/lock (file lock via `flock` on a per-SHA lockfile, OR an atomic claim-stamp via `O_EXCL` create + rename pattern matching the existing `Save` atomicity) so that concurrent same-SHA writers across different command dimensions (`go test` vs `golangci-lint`) serialize: exactly one writer claims and records all dimensions, the other consumers read the recorded result. Last-writer-wins MUST NOT silently drop a dimension.

**Exit:** AC-AUDIT-SNAPSHOT-004 green (single recording shared by 3 consumers; SHA change invalidates; concurrent-writer variant does not drop dimensions).

### Milestone M4 — Verify

Lowest reversibility: the verification pass itself.

- Full `go test ./...` + `go test -race ./internal/verify/... ./internal/runtime/...`.
- The K-1 residual-restatement sweep.
- AC matrix green; evidence pinned to commands + verbatim output.

## §G. Anti-Patterns (specific to this SPEC)

- **AP-1**: Silently serving a stale-SHA snapshot "because the tree probably didn't change" — violates the core invariant. The Stop hook MUST re-trigger or error explicitly, never fall through to a stale read.
- **AP-2**: Deleting the cold sync-auditor agent or its spawn path because "A3 makes it redundant on the happy path" — the cold auditor is the fallback for INCOMPLETE / dim-0 / contested; A3 promotes the workflow verdict on happy path only.
- **AP-3**: Introducing a parallel snapshot store instead of extending Step 0.5.2 — A4 is a wiring change by design; a second store creates the exact divergence the SPEC eliminates.
- **AP-4**: Dropping `tasks.md` from the hash subject set when adding `design.md` + `research.md` for Tier L — `tasks.md` is retained for grandfathered SPEC backward compat.
- **AP-5**: Restating the skip contract in a new surface (a new skill YAML, a new agent body) — the four-condition contract (minus the retired time-window) has ONE authoritative home (`spec-workflow.md`); new surfaces cross-reference it.

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-AUDIT-SNAPSHOT-001/spec.md` (this SPEC).
- acceptance.md: `.moai/specs/SPEC-AUDIT-SNAPSHOT-001/acceptance.md` (AC matrix + GWT).
- Design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 + §1.3 + §3.5 risk callout.
- Neighbor SPEC: `SPEC-AUDIT-GATE-INTEGRITY-001` (skip-policy invariants — extended, not superseded).
- Epic sibling (planned): `SPEC-STOPCHAIN-TRIM-001` (A10 + A8 + A11 — the other P0 of the epic).
- Integrity invariant: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 + §2.
