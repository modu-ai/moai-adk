# plan.md — SPEC-NAVIGATOR-SYNC-004 (BAS Epic M2 — Falconer Route)

> Plan-phase implementation plan. Status: draft. Tier M. Author: manager-spec.
> Run-phase entry requires Implementation Kickoff Approval. Do NOT write code from this plan.

## §A. Context

M0 (`SPEC-NAVIGATOR-SYNC-001`, PR #1375 `f4bd58acbd`, MERGED 2026-08-05) delivered the BAS binding-token trio + the graph-join layer at `internal/navigator/sync/` producing `nav-graph.json`. M1 (`SPEC-NAVIGATOR-SYNC-002`, PR #1379 `304907b6d`, MERGED 2026-08-06) delivered the **Detect** element (PostToolUse changed-path → affected-graph-rows mapping at `internal/navigator/detect/` + `internal/hook/navigator_detect*.go`, writing `.moai/state/navigator-detect/<session-id>.jsonl`). M4 (`SPEC-NAVIGATOR-SYNC-003`, PRs #1384/#1385, MERGED 2026-08-06/07) delivered the 4-tier map (`tiers.json` overlay). M2 is the **Route** element — the second of the Falconer loop (detect → route → fix): promote audit missing/orphan + detect findings into actionable work items, each owner-bound to a code path.

This plan resolves the 6 design questions posed by the orchestrator (§C.1–C.6), names the run-phase template-first touch list (§D), enumerates the asset-reuse map (§C.7), and defines the ≥70% accuracy measurement methodology (§E).

**Recon conducted (read-only, evidence-cited)**:
- `internal/navigator/sync/schema.go` — `Edge` struct carries `SourcePath` (absolute) + `LineNumber`. `Node` carries `EntityType` / `Identifier` / `DisplayName`. The owner-resolution path for `audit-missing` consults these to resolve a symbol node's declaration path.
- `internal/navigator/detect/traverse.go` — pure `Traverse(graph, changedPath) → *Result{Nodes, Edges}`. The JSONL writer (in `internal/hook/navigator_detect.go`) emits `{changed_path, changed_at, affected_nodes[], affected_edges[]}` per detection. M2 consumes the accumulated JSONL across all sessions.
- `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:553-591` — the audit-report.json emitter. Verified schema: `missing[]` entries carry `{design_name, source:{file, heading_path}, closest_match}`; `orphan[]` entries carry `{spec_id, title, implementation_path}`; `matched[]` entries carry `{design_name, spec_id, match_basis}`. The `implementation_path` field (line 575) is the load-bearing owner anchor for `audit-orphan`.
- `internal/cli/navigator_sync.go:22-32` — `newNavigatorSyncCmd` creates the M0 Hidden subcommand (`Use: "navigator-sync"`, `Hidden: true`). The sibling pattern M2 mirrors.
- `internal/cli/navigator_tiers.go:35-41` — M4's `navigator-tiers` Hidden subcommand; `internal/cli/navigator_tiers_test.go:65-70` asserts it is Hidden. M2's test mirrors this assertion.
- `internal/navigator/sync/write.go` — `atomicWrite` (`.tmp` + `os.Rename`) + `NAVIGATOR_PRE_RENAME_BARRIER`. M2's writer reuses this atomic-rename guarantee (REQ-NS4-008).
- No `nav-graph.json` / `audit-report.json` currently exists in the worktree (the repo's navigator state was not committed) — this is the expected fail-open baseline M2 must tolerate (REQ-NS4-009).

## §B. Known Issues + Pre-flight Decisions

- **B1 — audit-report.json may be absent**: 002's audit is produced on-demand by `/moai project --audit`. On a fresh checkout or before the first audit run, the file is absent. REQ-NS4-009 fail-open covers this: M2 writes an empty work-item set (or no output) and returns exit 0. M2 does NOT trigger an audit run — that is the user's call (`/moai project --audit` then `/moai project --route`, or a future combined invocation).
- **B2 — detect state directory may be absent or empty**: M1's `.moai/state/navigator-detect/` is populated only when a session with the M1 hook has edited files. On a fresh checkout, the directory is absent. M2 treats this as "no detect findings" (not an error) and still promotes audit findings. This is the graceful partial-input path (REQ-NS4-009: "a partial input set SHALL produce work items from the available inputs only").
- **B3 — detect JSONL spans multiple sessions**: M1 writes one `<session-id>.jsonl` per session. M2 reads ALL `*.jsonl` files in the directory, not just one. Deduplication by `changed_path` (latest `changed_at` wins) prevents the same path edited across N sessions from producing N duplicate work items.
- **B4 — `implementation_path` may be empty for an orphan**: 002's audit emits `implementation_path` from the capability-map row, but a SPEC with no capability-map entry (a SPEC that exists in `.moai/specs/` but was never mapped) yields an empty path. M2 marks such an orphan's `confidence` as `low` and falls back to the SPEC directory path (`.moai/specs/<spec-id>/`) as the `owner_path` — the SPEC itself is the owner when no code path is mapped. This is a defensible fallback (the SPEC's maintainer is the owner), and it counts toward the denominator but not the numerator of the ≥70% accuracy metric (honestly reflecting that no code path was bound).
- **B5 — the missing→symbol resolution requires a line-range heuristic**: the audit `missing` entry carries `source.heading_path` (a markdown heading anchor), not a line number. M2 scans the design doc at `source.file` for `@NAV:SYM:<symbol>` tokens (via the M0 graph's `sym-edge` records whose `source_path` matches the design doc), and if a symbol's edge `line_number` falls within the heading's line range (resolved by re-parsing the doc's headings), M2 binds the owner to that symbol's package. **Run-phase reconnaissance step (D3 lift)**: the M0 graph's `sym-edge` records DO carry `line_number`, but the audit's `heading_path` is a string anchor (e.g. `## Authentication > ### OAuth2`), not a line range. The M1.1 run-phase MUST verify, BEFORE relying on the heading-range heuristic, that (1) re-parsing the design doc to map heading anchors → line ranges is reliable, and (2) the `sym-edge` `line_number` is accurate enough for a range-containment check. **Decision branch — recorded as an explicit run-phase decision point**: IF the recon shows heading→line-range mapping is unreliable (e.g. heading anchors are ambiguous, or the doc was edited since the graph was built), THEN downgrade the missing→symbol resolution to a **whole-doc fallback**: M2 binds the owner to the design-doc `source.file` directly (skipping the symbol lookup), and records the decision in `progress.md §E.2` at M2.2. In that case, REQ-NS4-004's missing-resolution path still holds — it just always takes the "else design-doc source.file" branch. The missing→symbol lookup becomes a SHOULD refinement, not a MUST.
- **B6 — no separate `moai hook navigator-route` subcommand**: the Route layer is a CLI subcommand invoked from `/moai project`, not a hook. Adding a hook wrapper would fork the chain (and M2 is on-demand, not PostToolUse — REQ-NS4-001). No new wrapper script.

## §C. Design Questions Resolved

### C.1 — Trigger surface: on-demand CLI (NOT PostToolUse real-time)

**Decision**: on-demand Hidden cobra subcommand `navigator-route`, invoked from `/moai project`. NO PostToolUse real-time path in M2.

**Rationale (REQ-NS4-001)**: the design report §6 M2 card (line 365-369) says "M1과 병렬 가능" (parallel with M1) — but "parallel" here means *parallel in the milestone graph* (M2 does not depend on M1's completion; both depend only on M0), NOT *parallel in the trigger cadence*. The trigger question is distinct from the dependency question:

1. **The audit-report.json is itself on-demand** (produced by `/moai project --audit`). A PostToolUse real-time trigger for M2 would usually find no audit to promote (the audit is not re-run on every edit), or a stale one. Promoting a stale audit into fresh work items is worse than not promoting at all.
2. **Promoting each individual PostToolUse detection into a work item on every edit would flood the engineer** — one work item per keystroke is the wrong cadence for a persisted, owner-bound artifact. M1 already owns the real-time advisory surface (`systemMessage` + JSONL); M2 owns the coarser-cadence persisted work-item surface. The two layers compose: M2 consumes M1's accumulated detect state across all sessions when invoked, but does not replicate M1's real-time trigger.
3. **Clean separation of concerns**: M1 = detect (real-time, ephemeral advisory); M2 = route (on-demand, persisted work-item). M3 = fix (on-demand or trigger-from-work-item, AI-drafted regen). The Falconer loop is detect→route→fix, and "route" is the moment the engineer asks "what do I owe the doc?" — that is an on-demand question, not a per-keystroke one.
4. **The design report's "실시간 영향 범위 표시" (real-time impact display) phrasing in the M1 card (line 363) belongs to M1**, not M2. M2's card (line 368) says "audit 권고 → 작업 항목 승격" (audit advisory → work item promotion) — a promotion act, not a real-time display act.

A future PostToolUse real-time path for M2 is explicitly deferred (acceptance.md §G Forward-Looking Checks) — it would require a debounce/cadence policy (e.g. "promote only on session end" or "promote only when ≥N detections accumulate for one path") that is out of scope for the on-demand milestone.

**Integration point (REQ-NS4-011)**: a new Hidden cobra subcommand at `internal/cli/navigator_route.go`, wired into the existing `/moai project` skill step sequence the same way M0's `navigator-sync`, 003's `navigator-enrich`, and M4's `navigator-tiers` are wired. The Route layer's engine lives in `internal/navigator/route/` (new package).

### C.2 — Owner-binding contract (three resolution paths)

**Decision**: three resolution paths, one per `source_kind`, each binding to a code path or a design-doc path (never a person). The contract is specified exhaustively in spec.md §C REQ-NS4-004; this section records the design rationale + the confidence taxonomy.

**Why three paths, not one**: the three `source_kind`s carry different owner-anchor fields in their source data. An `audit-orphan` already has an `implementation_path` (002 emits it directly); an `audit-missing` has a design-doc location but no code path (the design feature is unfulfilled — the question is whether code exists for it); a `detect` finding has the `changed_path` (the file just edited). Forcing all three through one resolution path would discard the natural owner anchor each one carries. Three paths, each using the most-direct anchor, is the honest mapping.

**Confidence taxonomy** (the `confidence` field on each work item):
- **high** — owner resolved from a direct path field in the source data (`audit-orphan.implementation_path`, `detect.changed_path`). No graph traversal, no heuristic.
- **medium** — owner resolved via M0 graph traversal (`audit-missing` → `@NAV:SYM` symbol → symbol's declaration `source_path`). One hop through the graph.
- **low** — owner fell back to a path with no code binding (`audit-missing` → design-doc `source.file` with no symbol; `audit-orphan` with empty `implementation_path` → SPEC directory path). The owner is a doc/spec path, not a code path.

The ≥70% accuracy metric (REQ-NS4-010) counts only `high` + `medium` toward the numerator. A `low`-confidence owner is still a valid work item (it names *something* responsible), but it honestly reflects that the Route layer could not bind a code path — and that is the signal M3 (or a human) should look at it next.

**Why owner = path, not person** (falconer binding, design report §6 line 368): a person moves on (author turnover, role change, departure); a code path stays. Binding the owner to `git blame`'s author would couple the work item to a transient fact. Binding it to the code path means the work item survives the author — whoever next touches that path inherits the work item. This is the falconer principle's load-bearing design choice, and M2 honors it verbatim.

### C.3 — Output surface (work-items.{md,json})

**Decision**: two surfaces, produced from the same in-memory work-item set in a single pass.

1. **`work-items.json`** (machine-readable, the SSOT for downstream consumers — a future tracker integration, an M3 fix-loop, a dashboard): schema
   ```json
   {
     "provenance": {"route_commit_sha":"…","captured_at":"…","inputs":{"audit_commit":"…","nav_graph_commit":"…","detect_sessions":["…"]}},
     "work_items": [
       {"source_kind":"audit-orphan","source_entry":{"spec_id":"SPEC-X","title":"…","implementation_path":"…"},"owner_path":"/abs/…/internal/foo.go","action":"link this SPEC to a design feature or document its design rationale","confidence":"high"},
       {"source_kind":"audit-missing","source_entry":{"design_name":"…","source":{"file":"tech.md","heading_path":"## Auth"}},"owner_path":"/abs/…/internal/auth/","action":"create a SPEC for this design feature or link existing code","confidence":"medium"},
       {"source_kind":"detect","source_entry":{"changed_path":"…","changed_at":"…","affected_nodes":[…]},"owner_path":"/abs/…/internal/auth/login.go","action":"verify the affected doc rows still hold after this edit","confidence":"high"}
     ]
   }
   ```
   The `provenance.inputs` block records WHICH audit commit + nav-graph commit + detect sessions fed this run, so a later consumer can diff-check the inputs (verification-claim-integrity §2 attribution).

2. **`work-items.md`** (human-readable, grouped by `source_kind`):
   ```markdown
   # Navigator Work Items

   _Provenance: route_commit_sha <sha>, captured_at <date>, inputs audit=<sha> nav-graph=<sha> detect=<N sessions>_

   ## Orphan SPECs (audit — SPEC with no matching design feature)

   | SPEC | owner (code path) | action |
   |------|-------------------|--------|
   | SPEC-X | `internal/foo.go` | link this SPEC to a design feature… |

   ## Missing SPECs (audit — design feature with no SPEC)

   | design feature | owner (code/doc path) | action |
   |----------------|----------------------|--------|
   | OAuth2 strategy | `internal/auth/` | create a SPEC for this design feature… |

   ## Detect findings (M1 — code edit touched bound rows)

   | changed path | owner (code path) | affected rows | action |
   |--------------|-------------------|--------------|--------|
   | `internal/auth/login.go` | `internal/auth/login.go` | 3 nodes, 2 edges | verify the affected doc rows… |
   ```

**NOT emitted** (forbidden): a GitHub issue, a Linear ticket, a `.moai/specs/` draft, a TODO file, a mutation of any source file / SPEC / producer artifact.

### C.4 — Atomic-write + idempotence

**Decision**: reuse M0's `atomicWrite` (`.tmp` + `os.Rename`) for both `work-items.md` and `work-items.json`. Idempotence by deterministic ordering + no wall-clock.

**Mechanism (REQ-NS4-008)**: the work-item array is sorted by `(source_kind, owner_path, source_entry.identifier)` before serialization. The `provenance` block uses `route_commit_sha` (`git rev-parse HEAD`) + `captured_at` (committer date of that SHA) — NO wall-clock `time.Now()`. Two runs on the same HEAD with the same inputs produce byte-identical output. This carries forward M0's REQ-NS-009 + M4's REQ-NS3-019.

**Test hook**: no new test hook needed. M0's `NAVIGATOR_PRE_RENAME_BARRIER` is the precedent, but M2's writes are not concurrent with a reader the way M0's graph writes are (M2 is on-demand, single-invoker). The atomic-rename is still required for crash-safety (a reader never sees a half-written file), but no barrier test is needed.

### C.5 — Fail-open semantics

**Decision**: the Route layer is fail-open on EVERY error mode (REQ-NS4-009). Failure modes and their handling:
- `audit-report.json` absent → log one line, promote from detect + nav-graph only (partial input).
- `audit-report.json` unparseable JSON → log one line, skip audit promotion, promote from detect + nav-graph.
- `nav-graph.json` absent → log one line, owner resolution degrades (no symbol→package lookup; all `audit-missing` owners fall back to design-doc path with `confidence: low`), promote from audit + detect with degraded owners.
- detect state directory absent/empty → log one line, promote from audit only (this is the common case before M1 runs).
- detect JSONL unparseable → skip the malformed line (per-line fail-open, like M1's per-edge fail-open), continue with well-formed lines.
- owner-resolution error (symbol node references non-existent path) → mark the work item `confidence: low`, owner_path = the fallback path, continue.
- ALL inputs absent → write NO output (do not emit an empty `work-items.json`), log one summary line, return exit 0. (Distinct from "some inputs absent" — if nothing is available, there is nothing to route.)
- per-run timeout exceeded (context cancellation) → return whatever partial work-item set was collected (possibly empty), no log (context cancellation is not an error to advertise).

In ALL cases: exit 0, no user-facing error, no cascade into sibling `/moai project` steps. The Route layer wraps its entire body in a `defer recover()` + a `context.WithTimeout(ctx, 500ms)` — a panic in the promotion loop cannot escape into the project step.

### C.6 — Hidden cobra subcommand wiring

**Decision**: `navigator-route` registers as a Hidden cobra subcommand at `internal/cli/navigator_route.go`, mirroring `navigator_sync.go` / `navigator_tiers.go` exactly. It is invoked from the existing `/moai project` skill step sequence (a new `--route` flag or a sibling step alongside `--audit`), NOT a new top-level `moai` subcommand.

**Test guard (REQ-NS4-011)**: `internal/cli/navigator_route_test.go` asserts `navigator-route` is Hidden, mirroring `navigator_tiers_test.go:65-70`:
```go
if !cmd.Hidden {
    t.Errorf("navigator-route is not Hidden (must mirror navigator-sync/navigator-tiers)")
}
```

### C.7 — Asset-reuse map (consumer-only, bridge-not-absorb)

| Existing asset | File:line | Reuse in M2 |
|---|---|---|
| M0 `sync.Graph` / `sync.Edge` / `sync.Node` types | `internal/navigator/sync/schema.go` | The Route layer imports `internal/navigator/sync` and uses these types directly to parse `nav-graph.json` for owner resolution (symbol → declaration `source_path`). NO re-declaration. |
| M0 `atomicWrite` | `internal/navigator/sync/write.go` | The atomic-rename pattern M2's writer reuses for `work-items.{md,json}` (C.4). The function itself is re-implemented in M2's package (it is package-private in M0) — the *pattern* is reused, cited by file:line. |
| M0 `Provenance` model | `internal/navigator/sync/schema.go` (the `Provenance` struct) | The shape M2's `work-items.json` provenance block follows (C.4). NO wall-clock; `route_commit_sha` + `captured_at`. |
| 002 audit-report.json schema | `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:553-591` | The input schema M2 consumes. The `orphan[].implementation_path` (line 575) is the load-bearing owner anchor for `audit-orphan`. The `missing[].source.file` + `source.heading_path` (line 564-565) is the owner anchor for `audit-missing`. M2 does NOT execute the shell script — it reads the emitted JSON. |
| M1 detect JSONL schema | `internal/hook/navigator_detect.go` (the JSONL writer) | The input schema M2 consumes: `{changed_path, changed_at, affected_nodes[], affected_edges[]}` per line. M2 reads ALL `*.jsonl` in `.moai/state/navigator-detect/`, deduplicating by `changed_path`. |
| Hidden-subcommand sibling pattern | `internal/cli/navigator_sync.go:22-32`, `internal/cli/navigator_tiers.go:35-41` | The CLI registration pattern M2 mirrors (C.6). |
| `decide_action` directive patterns | (NEW — M2 authors these) | The one-line `action` directives (REQ-NS4-003) are M2's own prose, authored per `source_kind`. They are NOT imported from an existing asset; they are the Route layer's value-add. |

**Nothing is duplicated from scratch that already exists**: the graph schema, the atomic-write pattern, the provenance model, the audit JSON schema, the detect JSONL schema, and the Hidden-subcommand pattern are all cited and consumed. M2's new code is the promotion engine (audit/detect → work-item transform) + the owner-resolution logic (three paths) + the action-directive prose + the fail-open wrapper + the two-output serializer.

## §D. Run-Phase Touch List (Template-First, named for §C compliance)

**Primary work location** (CLAUDE.local.md §2): runtime Go files are edited in place (NOT under `internal/template/templates/`); template-managed files are edited template-first.

| Path | Template-managed? | Change |
|---|---|---|
| `internal/navigator/route/promote.go` | NO (runtime Go) | NEW file: the promotion engine (audit/detect → work-item transform). |
| `internal/navigator/route/owner.go` | NO (runtime Go) | NEW file: the three-path owner-resolution logic (REQ-NS4-004). |
| `internal/navigator/route/write.go` | NO (runtime Go) | NEW file: atomic-write of `work-items.{md,json}` + provenance. |
| `internal/navigator/route/promote_test.go` | NO (runtime Go) | NEW file: unit tests per source_kind + owner-resolution path. |
| `internal/navigator/route/owner_test.go` | NO (runtime Go) | NEW file: owner-resolution tests (three paths + confidence taxonomy). |
| `internal/navigator/route/failopen_test.go` | NO (runtime Go) | NEW file: one fail-open test per error mode (REQ-NS4-009). |
| `internal/navigator/route/coverage_test.go` | NO (runtime Go) | NEW file: the ≥70% accuracy fixture corpus + ratio assertion (§E). |
| `internal/navigator/route/nonoverlap_test.go` | NO (runtime Go) | NEW file: grep-based non-overlap assertion (REQ-NS4-006), pattern from M0/M1. |
| `internal/cli/navigator_route.go` | NO (runtime Go) | NEW file: the Hidden cobra subcommand. |
| `internal/cli/navigator_route_test.go` | NO (runtime Go) | NEW file: Hidden-subcommand assertion (mirrors `navigator_tiers_test.go`). |
| `internal/cli/project.go` (or the `/moai project` wiring site) | NO (runtime Go) | ONE new step/flag wiring `navigator-route` into the project skill sequence. |
| `internal/navigator/route/testdata/route-corpus/` | NO (test fixture) | NEW directory: the fixture corpus for the coverage test (§E). |
| `internal/template/templates/.claude/` | YES | (Verify) if a new distributed config key is needed; else no template change. M2 is expected to ship NO distributed surface (pure CLI + runtime Go) — the Route layer is invoked from the existing `/moai project` skill, which is already template-distributed. If confirmed, this row reduces to "no template path in the diff, no catalog regen" — documented in the PR body. |
| `catalog.yaml` regeneration | YES (generated) | `make build` ONLY IF a template file changed. Expected: no regen. |

**Non-overlap grep tests** (run-phase, REQ-NS4-006): `internal/navigator/route/nonoverlap_test.go` asserts the Route source does not write to forbidden surfaces — pattern lifted from `internal/navigator/sync/nonoverlap_test.go` + `internal/hook/navigator_detect_nonoverlap_test.go`.

## §E. ≥70% Route Accuracy Measurement Methodology (REQ-NS4-010)

**Definition of "input finding"**: a single entry from the consumed inputs — one `missing[]` or `orphan[]` element from audit-report.json, OR one deduplicated `changed_path` row from the detect JSONL. `matched[]` entries are NOT findings (they are already-resolved) and are excluded from both numerator and denominator.

**Definition of "actionable work item"**: a promoted work item whose `owner_path` is non-empty AND whose `confidence` is `high` or `medium`. A `low`-confidence work item (design-doc fallback / SPEC-directory fallback) is a valid promotion but does NOT count as actionable — it honestly reflects that no code path was bound.

**Measurement procedure** (the command that produces the percentage — acceptance.md §D.AC-NS4-010 binds the test to this):
1. A fixture corpus at `internal/navigator/route/testdata/route-corpus/` containing:
   - A synthetic `audit-report.json` fixture with 6 `missing[]` entries (3 with a design-doc `@NAV:SYM` token resolvable via the graph fixture → medium-confidence; 3 with no symbol → low-confidence fallback) and 12 `orphan[]` entries (10 with non-empty `implementation_path` → high-confidence; 2 with empty path → low-confidence SPEC-directory fallback).
   - A synthetic detect-state directory with 12 `*.jsonl` lines across 2 session files (all with `changed_path` → high-confidence).
   - A synthetic `nav-graph.json` fixture covering the symbol nodes referenced by the `missing` corpus.
   - **Corpus total**: 6 missing + 12 orphan + 12 detect = 30 input findings.
2. A Go test `TestRouteAccuracy` runs the Route layer against the corpus and emits:
   ```
   accuracy = (actionable work items) / (total input findings)
            = (high + medium) / (missing + orphan + detect dedup'd)
   ```
3. The test asserts `accuracy >= 0.70` and prints the observed percentage on failure. The percentage is the observed output of `go test ./internal/navigator/route/ -run TestRouteAccuracy -v` — NOT a narrative.

**Corpus design rationale + dual-arithmetic (D1 fix)**: the corpus is deliberately seeded with a known mix so the accuracy is computable by hand under BOTH the happy path AND the B5 fallback (missing→symbol lookup collapsing mediums to lows). The explicit arithmetic:
- **Happy path**: actionable = 3 medium-missing + 10 high-orphan + 12 high-detect = 25; total = 30; accuracy = 25/30 = **83.3%** ≥ 70% ✓.
- **B5 fallback worst case** (all 3 medium-missing collapse to low — whole-doc fallback): actionable = 0 medium + 10 high-orphan + 12 high-detect = 22; total = 30; accuracy = 22/30 = **73.3%** ≥ 70% ✓ — the floor survives the fallback with 3.3pp headroom.

The corpus distribution (missing's share lowered to 6/30, orphan-high raised to 10) is the lever that makes the fallback clear the floor: since the fallback zeroes missing's actionable contribution, the orphan-high + detect-high counts (22 actionable) must carry ≥70% on their own — and 22/30 = 73.3% does. A regression that drops owner resolution further (e.g. orphan `implementation_path` lookups failing, downgrading high-orphans to low) would drop below 70% and fail the test — the metric remains sensitive to owner-resolution quality.

**Why fixture-driven, not repo-driven**: same rationale as M1's coverage measurement (SYNC-002 plan.md §E) — a repo-driven measurement would re-scan the live repo's audit + detect state and produce a number that drifts with every commit. A fixture corpus is deterministic and isolates the Route layer's promotion + owner-resolution quality from the repo's current drift shape.

## §F. Milestones (priority-ordered, no time estimates)

Ordered by decision-reversibility — the decisions most likely to change lead.

### M2.1 — Promotion engine (pure function, audit/detect → work-item)

Pure-Go function: input `{auditReport, detectRows, navGraph, projectRoot}`, output `{workItems, errors}`. No I/O side effects. This is the most-likely-to-change decision (the promotion transform shape, the work-item schema); isolating it as a pure function lets the auditor review it independently of the CLI wiring. Tests: unit tests per `source_kind` (audit-missing / audit-orphan / detect), per confidence level (high / medium / low), deduplication correctness.

### M2.2 — Owner-resolution logic (three paths + confidence)

The `owner.go` module: three resolution paths (orphan→implementation_path, missing→symbol-via-graph-else-doc, detect→changed_path) + the confidence taxonomy. This milestone's decision includes the B5 heading-range heuristic (missing→symbol lookup) — the run-phase recon governs whether the symbol lookup is a MUST or a SHOULD. Tests: one test per resolution path + per confidence level + the B5 fallback (heading-range unreliable → whole-doc fallback).

### M2.3 — Output serializer (work-items.{md,json} + provenance + atomic-write)

The `write.go` module: atomic-write of both files from the in-memory work-item set, provenance block (no wall-clock), deterministic ordering. This milestone's decision is the JSON schema (the contract a future tracker / M3 consumes) — schema decisions are reversible-but-disruptive, so they land after the engine but before the hardening.

### M2.4 — Fail-open hardening + CLI wiring

Wrap the engine in `defer recover()` + `context.WithTimeout(ctx, 500ms)`. Add one fail-open test per error mode (audit absent / detect absent / nav-graph absent / unparseable / schema-invalid / owner-error / timeout). Wire `navigator-route` as a Hidden cobra subcommand + into the `/moai project` step sequence. This milestone's decisions are the timeout value + the CLI wiring point — both low-reversibility-risk, land late.

### M2.5 — Coverage harness + non-overlap guards + template-first

The `TestRouteAccuracy` fixture corpus + ratio assertion (§E). The non-overlap grep test (`nonoverlap_test.go`). The Hidden-subcommand test (`navigator_route_test.go`). The template-first verification (§D — confirm no distributed surface, or mirror if there is one). This milestone bundles the verification surfaces together — they are the lowest-change-likelihood decisions and land last.

## §G. Anti-Patterns (carry forward from M0/M1 + add M2-specific)

- **AP-NS4-001 — PostToolUse real-time work-item promotion**: registering a hook branch that promotes each detection into a work item on every edit. FORBIDDEN (REQ-NS4-001, decision defended in §C.1). M2 is on-demand CLI only.
- **AP-NS4-002 — owner = person**: resolving the owner via `git blame`, `CODEOWNERS`, or any person/team/email. FORBIDDEN (REQ-NS4-004 falconer binding). Owner is always a code path or design-doc path.
- **AP-NS4-003 — external-tracker writes**: creating a GitHub issue / Linear ticket / SPEC draft from a work item. FORBIDDEN (§F Out of Scope) — M2 emits a local artifact only.
- **AP-NS4-004 — narrative accuracy claim**: writing "accuracy is ≥70%" in an AC without naming the command + observed output. FORBIDDEN (REQ-NS4-010 + verification-claim-integrity).
- **AP-NS4-005 — overclaiming asset reuse**: claiming the Route layer reuses `navigator-audit.sh`'s matching engine. FORBIDDEN (honest framing per M1's REQ-NS2-010 precedent) — M2 consumes the *emitted JSON*, not the shell script; the promotion engine is M2's own code.
- **AP-NS4-006 — modifying M0/M1/M4 producers**: writing to `internal/navigator/sync/`, `internal/navigator/detect/`, `internal/navigator/tiers/`, `internal/hook/navigator_detect*.go`, or `internal/mx/`. FORBIDDEN (REQ-NS4-005).
- **AP-NS4-007 — wall-clock in provenance**: using `time.Now()` in the `work-items.json` provenance block. FORBIDDEN (REQ-NS4-008 idempotence) — two runs on the same HEAD must produce byte-identical output.
- **AP-NS4-008 — carrying the accuracy percentage across SPECs**: a coverage number from M1 or a prior measurement is a carry-over, not a baseline (verification-claim-integrity §2). The fixture corpus is the baseline.

## §H. Cross-References

- spec.md §C REQ-NS4-001…REQ-NS4-012 — the requirement set this plan operationalizes.
- acceptance.md §D — the AC matrix binding each REQ to a Given-When-Then + the accuracy command.
- `.moai/reports/navigator-redesign-bas-20260805.html` §3.3(C) Falconer loop, §4 asset reuse, §6 M2 milestone, §9 success metrics.
- M0 plan.md (`.moai/specs/SPEC-NAVIGATOR-SYNC-001/plan.md`) — the asset-reuse + non-overlap + fail-open precedents M2 carries forward.
- M1 plan.md (`.moai/specs/SPEC-NAVIGATOR-SYNC-002/plan.md`) — the on-demand-vs-real-time decision pattern + the coverage-measurement methodology M2 adapts.
