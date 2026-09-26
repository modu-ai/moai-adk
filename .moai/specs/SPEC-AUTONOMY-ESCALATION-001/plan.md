# plan.md — SPEC-AUTONOMY-ESCALATION-001

Milestones are ordered by decision reversibility: the decisions most likely to change come
first, mechanical wiring last. No time estimates; priority labels only.

## §A — Context

- Base: `develop` at `ca1d5dc43`; worktree `.claude/worktrees/t1235`, branch `WT-escalation-detector`.
- Tier L (ten detection classes across PreToolUse, PostToolUse, Stop and checkpoint surfaces, a
  record layer, a contract-reading resolver, a state file with a tamper chain, config, template
  default; expected > 15 files).
- Run-phase entry gate: card t1234 (A1, contract schema) merged into `develop`, and A1 request R9
  (`card` field) satisfied. Plan may proceed in parallel; run may not (card text).
- Measured premises: `research.md` §B. Proposed mechanics: `design.md`. Lead rulings: spec.md §H.
  Items moved to card t1245: spec.md §K.

## §B — Known issues carried into run

- **B-1** Queue store forbids a fourth state (`backlog_schema_freeze_test.go:3-5,70`); the
  detector never reads or writes it and the needs-decision marking lives only in escalation
  records.
- **B-2** `graph.FileAPI` is working-tree only and CGO/grammar-dependent (`codequery.go:65-95`).
- **B-3** `harness.escalation` is a homonym (`internal/config/types.go:1199-1243`); do not reuse.
- **B-4** `MOAI_AUTONOMY_TIER` stays untouched (`internal/config/envkeys.go:139-148`).
- **B-5** Hook bind budget: no subprocess on the per-tool-call path; HEAD read from ref files;
  A1 verify in-process and cached by contract digest on PreToolUse; class 4 only at the
  on-demand checkpoint (spec.md C4).
- **B-6** Changing any `acceptance.md` requires `./internal/spec` in the re-measure scope (AC
  snapshot guard).
- **B-7** The destructive denylist returns early at `internal/hook/pre_tool.go:505-517` under an
  `@MX:ANCHOR` forbidding a conditional return above it; the class 6 call sits above it and never
  returns (design.md §C.3).
- **B-8** The resolver depends on the contract's `card` field, which A1 does not yet define
  (request R9). Until it lands every card resolves not-armed, observably; the queue `spec_id` is
  deliberately not a fallback (lead ruling 09-26 (3) #1).

## §C — Pre-flight (run phase)

1. `git merge-base --is-ancestor <A1 merge commit> HEAD` → exit 0 (A1 present).
2. Re-check **every row of spec.md §F.1**, every open item, and requests R1-R4 and R7-R10 against
   the A1 that landed — not only the tagged requirements. If a field is renamed or missing, or an
   item resolved differently (in particular R9's field name or R10's store path), stop and return
   a blocker to the orchestrator for a mid-run spec amendment.
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

- Record path and frontmatter schema (spec.md §I); the card-field resolver as one function
  (design.md §C.8); read `workflow.autonomy.mode` through A1's reader (inert under `guided`); add
  only `escalation.new_api_detector`.
- ACs: AC-AE-001, AC-AE-002, AC-AE-003, AC-AE-020 (path half).

### M2 — Record writer, state file, fault handling, acceptance-change, first observation

- Writer with frontmatter, per-class fingerprints (design.md §C.10), occurrence counting and
  post-resolution re-trip naming; `not-armed` / `not-checked` / warning audit lines; card state
  file with the arming snapshot and the audit hash chain (design.md §C.6, §C.11); in-process
  verify at PreToolUse on first observation; class 1.
- ACs: AC-AE-005, AC-AE-006, AC-AE-020, AC-AE-021, AC-AE-022, AC-AE-024.

### M3 — Path and command classes (PreToolUse / PostToolUse)

- Classes 2a, 2b (A1 `frozen_files`), 3 with `effective_never`, exemptions with the
  `.moai/state/escalation/` carve-out and their root sources, unreadable field handling.
- ACs: AC-AE-007, AC-AE-008, AC-AE-009, AC-AE-010, AC-AE-023.

### M4 — New-architecture/API detector (Priority Medium, heuristic)

- Card base resolution, base-blob extraction, per-language sub-kinds, `off` switch, not-observed
  labelling.
- ACs: AC-AE-011, AC-AE-012.

### M5 — Evidence, irreversible action, operational trips, unified disarm

- Classes 5-9 and class 10 `detection-disarmed` with all five disarm reasons, ordered before
  resolution; the never-alters-tool-call sweep.
- ACs: AC-AE-004, AC-AE-013, AC-AE-014, AC-AE-015, AC-AE-016, AC-AE-017, AC-AE-018, AC-AE-019.

### M6 — Template default and documentation (mechanical)

- Template `workflow.yaml` gains `autonomy.escalation.new_api_detector: graph` under the block
  A1 introduces (A1 owns `mode: guided` and `budget_default`); `make build`.

## §G — Risks

| Risk | Mitigation |
|---|---|
| A1 declines or renames R9 (`card`) | Pre-flight step 2 stops the run; the resolver is one function, so only step 2 of design.md §C.8 changes |
| A1 plan-audit changes a consumed field | spec.md §F.1 full-row re-check in pre-flight step 2 |
| A contract copied between SPECs keeps a stale `card` value | Two claimants → not-armed plus a warning naming both (REQ-AE-002); an armed card that gains a second claimant disarms with `card-mismatch` (REQ-AE-017) |
| Bash deletes both the state file and the audit log | Residual until the moai-owned store exists (R10; spec.md §G) |
| PreToolUse latency under contract mode | Contract `card` reads, digest check, cached verify, ref-file reads; verify re-runs only on a digest change (C4) |
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
- **Q6** [RESOLVED — lead ruling 09-26 #1b, generalized by (3) #2].
- **Q7** [RESOLVED — lead ruling 09-26 #6; now read from A1's `frozen_files`].
- **Q8** [RESOLVED — lead ruling 09-26 #2, resolution input replaced by (3) #1].
- **Q9** [RESOLVED — lead ruling 09-26 (3) #1: the worktree directory name is the card id; the
  contract's `card` field, not the queue, links it to a SPEC].
- **Q10** A1 requests R8 (projection owner label), R9 (`card` field inside the signed digest),
  and R10 (moai-owned store path) — accepted by A1/A3, or answered otherwise?
