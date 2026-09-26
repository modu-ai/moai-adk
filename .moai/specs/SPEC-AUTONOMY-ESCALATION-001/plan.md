# plan.md — SPEC-AUTONOMY-ESCALATION-001

Milestones are ordered by decision reversibility: the decisions most likely to change come
first, mechanical wiring last. No time estimates; priority labels only.

## §A — Context

- Base: `develop` at `ca1d5dc43`; worktree `.claude/worktrees/t1235`, branch `WT-escalation-detector`.
- Tier L (ten detection classes across PreToolUse, PostToolUse, Stop and checkpoint surfaces, a
  record layer, a queue-reading resolver, config, template default; expected > 15 files).
- Run-phase entry gate: card t1234 (A1, contract schema) merged into `develop`. Plan may proceed
  in parallel; run may not (card text).
- Measured premises: `research.md` §B. Proposed mechanics: `design.md`. Lead rulings: spec.md §H.
  Items moved to card t1245: spec.md §K.

## §B — Known issues carried into run

- **B-1** Queue store forbids a fourth state (`backlog_schema_freeze_test.go:3-5,70`); the
  resolver only reads it and the needs-decision marking lives only in escalation records.
- **B-2** `graph.FileAPI` is working-tree only and CGO/grammar-dependent (`codequery.go:65-95`).
- **B-3** `harness.escalation` is a homonym (`internal/config/types.go:1199-1243`); do not reuse.
- **B-4** `MOAI_AUTONOMY_TIER` stays untouched (`internal/config/envkeys.go:139-148`).
- **B-5** Hook bind budget: no subprocess on the per-tool-call path; HEAD read from ref files;
  class 4 only at the on-demand checkpoint (spec.md C4).
- **B-6** Changing any `acceptance.md` requires `./internal/spec` in the re-measure scope (AC
  snapshot guard).
- **B-7** The destructive denylist returns early at `internal/hook/pre_tool.go:505-517` under an
  `@MX:ANCHOR` forbidding a conditional return above it; the class 6 call sits above it and never
  returns (design.md §C.3).
- **B-8** The resolver depends on the queue recording `spec_id` for the card
  (`backlog_store.go:81`); a card picked without one stays not-armed, observably.

## §C — Pre-flight (run phase)

1. `git merge-base --is-ancestor <A1 merge commit> HEAD` → exit 0 (A1 present).
2. Re-check **every row of spec.md §F.1**, every open item, and requests R1-R4 and R7 against the
   plan-audit-passed A1 schema — not only the tagged requirements. If a field is renamed or
   missing, or an item resolved differently, stop and return a blocker to the orchestrator for a
   mid-run spec amendment.
3. Capture the AC-AE-001 golden baseline in its own commit before any detector code.

## §D — Constraints

- Detection and reporting only; no denial of any kind (spec.md C3), no closure report (A4), no
  schema design (A1), nothing from spec.md §K (card t1245).
- Default config `guided`; template carries no card id, SPEC id, date, or SHA.
- One writer per tree; change-scoped tests only; no background load.

## §E — Self-verification (per milestone)

E1 AC matrix with command + verbatim output + HEAD, each AC by its named test and package
(acceptance.md §A); E2 cross-platform build; E3 coverage; E4 subagent-boundary grep; E5 lint
delta; E8 RED output before GREEN.

## §F — Milestones

### M1 — Record path, resolver, and activation gate (Priority High, most change-prone)

- Record path and frontmatter schema (spec.md §I); the two-layer resolver as one function
  (design.md §C.8); read `workflow.autonomy.mode` through A1's reader (inert under `guided`); add
  only `escalation.new_api_detector`.
- ACs: AC-AE-001, AC-AE-002, AC-AE-003, AC-AE-019 (path half).

### M2 — Record writer, dedup, fault handling, acceptance-change, first-resolution invalidity

- Writer with frontmatter, occurrence counting and post-resolution re-trip naming,
  `not-armed` / `not-checked` / warning audit lines, card state file, class 1.
- ACs: AC-AE-005, AC-AE-006, AC-AE-019, AC-AE-020, AC-AE-021, AC-AE-023.

### M3 — Path and command classes (PreToolUse / PostToolUse)

- Classes 2a, 2b (frozen union), 3 with post-signing immutability, exemptions and their root
  sources, unreadable field handling.
- ACs: AC-AE-007, AC-AE-008, AC-AE-009, AC-AE-010, AC-AE-022.

### M4 — New-architecture/API detector (Priority Medium, heuristic)

- Card base resolution, base-blob extraction, per-language sub-kinds, `off` switch, not-observed
  labelling.
- ACs: AC-AE-011, AC-AE-012.

### M5 — Evidence, irreversible action, operational trips, contract-void

- Classes 5-10, with contract-void ordered before resolution; the never-alters-tool-call sweep.
- ACs: AC-AE-004, AC-AE-013, AC-AE-014, AC-AE-015, AC-AE-016, AC-AE-017, AC-AE-018.

### M6 — Template default and documentation (mechanical)

- Template `workflow.yaml` gains `autonomy.escalation.new_api_detector: graph` under the block
  A1 introduces (A1 owns `mode: guided` and `budget_default`); `make build`.

## §G — Risks

| Risk | Mitigation |
|---|---|
| A1 plan-audit changes the draft or declines R1-R4/R7 | spec.md §F.1 full-row re-check in pre-flight step 2 |
| Cards picked without a `spec_id` never arm | Observable `not-armed` line naming the queue layer (REQ-AE-002); B-8 |
| PreToolUse latency under contract mode | In-memory checks, file stats, ref-file reads, one queue read (C4) |
| Class 4 false positives on refactors that move declarations | Record lists additions only by kind; sync-audit remains the backstop (C6) |
| Class 6 regex misses an obfuscated push | Residual risk; blocking belongs to A3 |
| Invariant commands never run → never checked | Reported as not-observed (REQ-AE-022, O10) |
| Exemption root undeterminable → outside-root write unjudged | Listed as not-observed, never silently allowed or tripped (REQ-AE-013) |

## §H — Open questions (for the lead)

- **Q1** [RESOLVED — lead ruling 09-26 #4, format superseded by (2) #6].
- **Q2** Is a new CLI verb for on-demand checkpoints acceptable, given it is itself a class-4
  event? Class 4 runs only there (C4).
- **Q3** [RESOLVED by the A1 draft] class 1 consumes A1's verify.
- **Q4** [RESOLVED — lead ruling 09-26 #5].
- **Q5** Does "recorded CI failure" have an existing on-disk producer, or does class 5's CI limb
  stay not-observed until one exists?
- **Q6** [RESOLVED — lead ruling 09-26 #1b, ordering by (2) #4].
- **Q7** [RESOLVED — lead ruling 09-26 #6].
- **Q8** [RESOLVED — lead ruling 09-26 #2, narrowed by (2) #1].
- **Q9** [RESOLVED — lead ruling 09-26 (2) #1: the worktree directory name is the card id].
