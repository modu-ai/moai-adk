# decision-index.md — SPEC-MOAI-HYGIENE-001

Stateless companion (no status field — the SPEC's lifecycle lives in spec.md). One row per decision surfaced during assembly that the operator has not settled. Card t1518 explicitly delegated the rotation shape and liveness design to the SPEC author; those delegations are recorded as unresolved pending the operator's verdict, with the author's working default carried in spec.md/plan.md so run phase can proceed.

### Q1: Should audit-sink rotation be a standalone locked pass, and are 10 MiB / keep-1 the right shape?
Label: FOUNDER
Authority anchor: n/a — no committed authority decides rotation shape
Why unresolved: Card t1518 delegates the rotation design to this SPEC ("rotation shape is the SPEC's design decision — justify it against the measured sizes"); no prior completed SPEC's HISTORY and no config section settles standalone-pass vs in-writer rotation or the threshold values. Working default (spec.md §F): standalone pass under a lockfile, 10 MiB threshold (aligns with the trace writer's REQ-OBS-006 constant), keep-1 — justified by the measured ~1 MB/day peak growth and the ≈20 MiB per-sink bound. An operator preferring in-writer rotation or deeper retention changes spec.md §F and plan.md M1.
Operator verdict:

### Q2: Are 48 h (transcript activity window) and 7 days (minimum-age floor) the right liveness/age values?
Label: FOUNDER
Authority anchor: n/a — values are author-chosen defaults
Why unresolved: The 48 h window reproduces the hygiene audit's own live/dead discrimination baseline and the 7-day floor guards clock skew and in-flush sessions, but both are judgment defaults, not measured optima; no committed authority fixes them. Working default: named constants (`HygieneTranscriptActivityWindow = 48h`, `HygieneMinAgeDays = 7`), operator-tunable via `workflow.yaml`. Both err toward keeping (fail-closed direction).
Operator verdict:

### Q3: Should the rotator also cover `lessons-inbox.jsonl`?
Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 (Enforce Simplicity — reuse-before-build ladder: "does a helper, type, or pattern already exist here — reuse it")
Why unresolved as written: The question was raised during assembly because the audit lists the inbox among unbounded sinks; it resolves by the constitution's reuse clause — `internal/hook/inbox_lifecycle.go` and the `moai inbox` verbs already own the inbox's size lifecycle, so duplicating it in the sink registry would add a second owner for one file. Recorded here because the exclusion is a scope decision an operator could overturn.
Operator verdict:

### Q4: May the GC remove an aged `spec-close-*.lock` when the fd probe cannot run?
Label: DECIDED
Authority anchor: SPEC-WORKTREE-SWEEP-001 (status: completed), REQ-WS-006 — "lsof absent or failing returns an error, which the sweep renders as an unanswerable predicate: the tree is preserved, never reported unoccupied"
Why resolved: The identical question — may deletion proceed when the external probe that would establish safety cannot run — is decided fail-closed by the completed sweep SPEC for the same probe family. This SPEC adopts the same disposition: probe unavailable → INDETERMINATE → kept (spec.md REQ-HYG-013).
Operator verdict:

### Q5: Should the operator surface be a `moai clean` extension or a dedicated verb?
Label: FOUNDER
Authority anchor: n/a — no committed authority fixes the CLI surface for hygiene
Why unresolved: Extending `moai clean` with `--audit-logs` / `--session-state` / `--apply` reuses an existing dry-run-default cleanup verb; a dedicated `moai hygiene` verb would read better in help text but adds a surface. Working default: extend `moai clean` (spec.md REQ-HYG-017). An operator preferring a dedicated verb changes plan.md M4 and REQ-HYG-017 only.
Operator verdict:

### Q6: Should `state/verify` scratch and `state/todo` per-session residue be GC targets in v1, given their on-disk shapes are the least uniform of the target classes?
Label: FOUNDER
Authority anchor: n/a — inclusion is a scope judgment the card leaves to the SPEC
Why unresolved: The audit measured real residue in both (911 aged verify files; 201 dead todo files), and the design contains the shape risk (session-keyed names only; the shared `backlog.json`/`backlog.db` stores are on the never-touch negative list; undatable/unresolvable candidates spared). An operator preferring a narrower first cut would drop both from the target registry (one-list change in plan.md M3) and re-card them. Working default: included (spec.md REQ-HYG-006).
Operator verdict:
