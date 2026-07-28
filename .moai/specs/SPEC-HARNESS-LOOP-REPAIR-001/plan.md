# SPEC-HARNESS-LOOP-REPAIR-001 — Implementation Plan

> Tier L · Version 0.2.0 · 2026-07-27 · manager-spec
> Companion to `spec.md` (scope + requirements) and `acceptance.md` (AC SSOT).

Sections are ordered by **decision reversibility**: the decisions most likely to change sit at the top (§A.6, §F.1), and mechanical work is deferred to the bottom (§F.5). Review attention belongs at the top.

---

## §A Context

### A.1 Where the work happens

| | |
|---|---|
| Worktree | `.claude/worktrees/harness-loop-repair` |
| Branch | `feat/SPEC-HARNESS-LOOP-REPAIR-001`, based on `origin/main` = `760f09f73` |
| HEAD | `c996eb294` (M1 landed, unpushed at time of writing) |
| Primary checkout | held another session's branch during M1; **do not commit this SPEC's work there** |

### A.2 What is already complete

**M1 is complete and verified at `c996eb294`.** It repaired the producer/consumer layout mismatch (`spec.md` §A.3.1):

| File | Change |
|---|---|
| `internal/harness/proposalgen/layout.go` | NEW — shared accessor (`ProposalDirRel`, `MetadataFileName`, `ProposalDir`, `ListDraftIDs`, `ProposalPath`), placed beside the producer that defines the layout |
| `internal/cli/harness.go` | C1 `countProposals` + C2 apply selector routed through the accessor; duplicate path literal removed |
| `internal/cli/harness/execute.go` | C3 `resolveProposalPath` delegates path + traversal validation to the accessor; `draftLabel` added |
| `internal/cli/harness_layout_repro_test.go` | NEW — C1/C2 reproduction + AC-HLR-006 flat-derivation guard |
| `internal/cli/harness/layout_repro_test.go` | NEW — C3 reproduction, traversal guard, M2 schema characterization |

Live result: `moai harness status` went from `0 items` to `52 items` against the same tree. AC-HLR-001 / 002 / 003 / 006 all PASS (`acceptance.md` §C).

### A.3 What M1 revealed

M1's verification surfaced a **second, independent cause** that the original diagnosis missed: the fields `Applier.Apply()` requires were never collected anywhere upstream (`spec.md` §A.3.2), and the two artifacts are different species (`spec.md` §A.4). Repairing the layout made drafts *visible*; it did not make them *applicable*. This is why `spec.md` was amended to v0.2.0 before this plan was written.

### A.4 The decision this plan implements

Route `proposalgen` drafts to their **designed** consumer — manager-spec SPEC authoring — and drop the `execute`→draft wiring. The `Applier` frontmatter-enrichment path is **preserved and left unfed**; whether it ever gets a producer is deferred (`spec.md` §G question 4).

### A.5 Artifact completeness (Tier L)

`spec.md`, `plan.md`, `acceptance.md`, `progress.md` exist. `design.md` and `research.md` — nominally expected at Tier L — **do not exist**. This is recorded as known debt rather than silently ignored. The M2 design decisions that would normally live in `design.md` are carried in §A.6 and §F.1 of this file; if M2's design space grows beyond what those sections hold, author `design.md` before implementing rather than expanding this file further.

### A.6 Decisions — resolved at the Implementation Kickoff gate (2026-07-28)

These were the highest-change-likelihood items in the SPEC. Both were surfaced to the user at the Implementation Kickoff Approval gate on 2026-07-28 and resolved there. The resolutions below feed M2 implementation; they do not alter any REQ-HLR-* requirement or AC-HLR-* criterion — the requirements already accommodated every option on each question (AC-HLR-004 names outcomes, not mechanism; REQ-HLR-004c requires only "before createSnapshot").

**RESOLVED (Implementation Kickoff gate, 2026-07-28): promotion path surface → option (a), a new CLI verb `moai harness promote --id <ID>`.** Rationale: mechanical, testable, discoverable; AC-HLR-004 and AC-HLR-005 become falsifiable via CLI exit code + SPEC-directory existence — the strongest falsifiability available, and the property §A.4 identifies as missing in the predecessor recurrence. Options (b) orchestrator-side and (c) both were considered and rejected: (b) makes the AC behavior-based and weakly falsifiable (the exact failure mode §A.4 diagnoses); (c) adds a second surface to maintain without any REQ benefit. AC-HLR-004 continues to be satisfiable (it names outcomes, not mechanism), so this decision does not invalidate the AC.

**RESOLVED (Implementation Kickoff gate, 2026-07-28): applicability guard placement → in-`Apply` pre-flight.** The guard lives at the top of `applier.go`'s `Apply`, before Step 2 / `createSnapshot`. Rationale: `spec.md` §A.7 explicitly names "every future caller including ones bypassing the loader" as the protection target — in-`Apply` is defense-in-depth and the strongest guard. Both placements (load-time and in-`Apply`) satisfy REQ-HLR-004c (reject before any snapshot directory is created); in-`Apply` is chosen for the broader protection. The sequencing hazard in §F.1 is unchanged: this guard MUST land before any change that makes the producer payload parseable (§A.7).

A third question — the fate of `moai harness execute` once de-wired (`spec.md` §G question 4) — was **deliberately not** a blocker for M2 entry: AC-HLR-014 requires only an honest diagnostic, which is satisfiable whether the verb survives or is later removed. That question is also now resolved (`spec.md` §G q4 — leave dormant); the "not a blocker" characterization remains accurate because AC-HLR-014 needs only a diagnostic.

---

## §B Known issues carried into implementation

Filtered to the categories that actually apply to this SPEC's surface.

- **B2 Cross-SPEC policy conflict.** Four predecessor SPECs own code in this area (LOOP-CLOSURE, OUTCOME-CAPTURE, APPLY-EXECUTE, EVO-PIPE-REPAIR). M2 **reverses the wiring** SPEC-HARNESS-APPLY-EXECUTE-001 introduced. That reversal is deliberate and justified in `spec.md` §A.4 — it MUST be stated in the commit body, not performed silently.
- **B3 Subagent boundary (C-HRA-008).** `internal/harness/`, `internal/cli/harness/` and `internal/hook/` MUST NOT invoke `AskUserQuestion`. Verification: `grep -rn 'AskUserQuestion\|mcp__askuser' <pkg> | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` yields no invocation. Prose comments and flag descriptions are permitted and currently account for all matches.
- **B5 CI 3-tier.** spec-lint, golangci-lint and per-OS Test can each fail independently; classify pre-existing baseline vs NEW defect before reporting.
- **B6 spec-lint headings.** A bare `## Out of Scope` (h2) triggers `MissingExclusions`; an `### Out of Scope — <topic>` (h3) sub-heading with `-` bullets is required.
- **B8 Working-tree hygiene.** `.moai/harness/**` and `.moai/state/**` are runtime-managed. Do not commit them; stage by explicit pathspec, never `git add -A`.
- **B9/B10 Scope discipline.** Touch only the §D.2 target list. A parallel session may hold the primary checkout.
- **B11 Blocker reporting.** On a blocker, return a structured report; never prompt the user directly.

### B.13 SPEC-specific trap — the drafts are gitignored

`.gitignore:229` ignores `.moai/harness/proposals/`. The worktree therefore holds **0** drafts while the primary checkout holds **52**. Running `moai harness status` from inside the worktree reports `0 items` **even after a correct fix**, which reads as a failed repair.

- Unit fixtures MUST build their own `proposals/<ID>/proposal.json` under `t.TempDir()`.
- Any live-count check MUST pass `--project-root` at the primary checkout **and** use a binary built from the tree under test (`go build`), not the installed `moai`.

This trap already cost one false negative during M1; it is restated here because M2 re-enters the same surface.

---

## §C Pre-flight

Run before any code change; a single parallel batch.

```bash
# 1. Confirm tree identity and baseline (never assume the session-start snapshot)
git -C <worktree> rev-parse --show-toplevel
git -C <worktree> branch --show-current
git -C <worktree> rev-parse --short HEAD          # expect c996eb294 or a descendant

# 2. Build feasibility, both platforms
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint + test baseline (to separate NEW defects from pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5
go test ./... 2>&1 | tail -20

# 4. Confirm the M1 accessor is intact before extending it
go test -run 'TestNoFlatProposalPathDerivation|TestCountProposals_NestedLayout' ./internal/cli/...

# 5. Live draft inventory (primary checkout — the worktree has none, see §B.13)
ls -1d /Users/goos/MoAI/moai-adk-go/.moai/harness/proposals/*/ | wc -l   # expect 52
```

---

## §D Constraints

### D.1 PRESERVE — do not modify

| Path | Why |
|---|---|
| `internal/harness/proposalgen/layout.go` | M1 deliverable; the single layout accessor. Extend only through its exported surface |
| `internal/harness/proposalgen/scaffolder.go` producer semantics | The nested layout is the resolved decision (`spec.md` §G q1). M4 may change *identity derivation* (`mapper.go`), not the directory shape |
| `internal/harness/applier.go` apply semantics | The Applier path is preserved unfed. M2 adds a guard *ahead* of it; it does not alter `applyFileModification`, the 5-layer ordering, or snapshot/rollback |
| `internal/harness/safety/**` | The 5-layer pipeline is out of scope. §A.7's finding is that it does not check *applicability* — the fix is a guard before it, not a change inside it |
| `.moai/harness/proposals/**` | Runtime data. The 52 drafts (including the 7 duplicate pairs) are NOT retroactively rewritten |
| `.moai/harness/usage-log.jsonl`, `.moai/state/**` | Runtime-managed |
| The `goal` surface | Out of scope (`spec.md` §B.2); a concurrent SPEC owns it |

### D.2 Expected write targets by milestone

| M | Files |
|---|---|
| M2 | promotion path (new file, package per §A.6 decision); `internal/cli/harness/execute.go`; `internal/cli/harness/layout_repro_test.go` (tripwire retirement); new tests |
| M3 | orchestrator-side dispatch recording; `.claude/rules/**` or skill workflow surface |
| M4 | `internal/harness/proposalgen/mapper.go` (identity + narrowing); tests |
| M5 | `.claude/rules/moai/core/moai-constitution.md` (Lessons Protocol); lesson-store docs |
| M6 | `internal/cli/harness.go` (help text, `list`/`doctor` agreement); lesson entries |

### D.3 Forbidden

- `--no-verify`, `--amend`, force-push.
- `git add -A` (stage by explicit pathspec — a parallel session may share the checkout).
- Adding `MarshalJSON`/`UnmarshalJSON` to `harness.Tier` (AC-HLR-016 clause 1).
- Making the producer payload parseable **before** the applicability guard exists — that ordering is what activates the §A.7 hazard.
- Deleting `TestLoadProposalByID_ProducerSchemaMismatch` without the recorded rationale (AC-HLR-016 clause 2).

---

## §E Self-verification deliverables

Report per `.claude/rules/moai/core/verification-claim-integrity.md` §3 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

1. **AC matrix** — every AC touched, PASS/FAIL, with the command run and the output observed.
2. **Falsification executed** — for each AC claimed PASS, the revert applied, the failure observed, the revert undone, the suite re-confirmed green. `acceptance.md` §J item 5 makes this mandatory, not optional.
3. **Cross-platform** — `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
4. **Coverage** — `go test -cover ./internal/cli/harness/...`. Note the pre-existing gap: 80.9% against an 85% target, measured during M1.
5. **Subagent boundary** — the B3 grep, with invocation count zero.
6. **Lint** — `golangci-lint run`, NEW findings separated from baseline.
7. **Blockers** — structured report; no direct user prompting.

---

## §F Milestones

Ordered by decision reversibility. §F.1 carries the architectural decision and should absorb most review attention; §F.5 is mechanical.

### F.1 M2 — route drafts to their designed consumer  *(highest reversibility — review here)*

Implements REQ-HLR-004, 004b, 004c, 012. Satisfies AC-HLR-004, 005, 014, 015, 016.

1. §A.6 decisions were resolved at the Kickoff gate (2026-07-28) — confirm the recorded resolutions before continuing. This is now a no-op confirmation step, not a fresh decision.
2. Build the promotion path: draft `<ID>` → SPEC directory carrying `<ID>` as provenance → draft leaves the pending queue.
3. Add the promotion audit record (draft → SPEC → timestamp).
4. Add the applicability guard **before** `createSnapshot`.
5. De-wire `execute`→draft; replace the raw unmarshal error with a diagnostic naming the reason.
6. Retire the tier tripwire with recorded rationale.

**Reproduction-first.** Write the failing test before each change and confirm it fails first (CLAUDE.md Rule 4). For step 4 the reproduction is a fixture proposal with an empty `target_path` that currently reaches `createSnapshot` and leaves a dated directory behind.

**Sequencing hazard.** Step 4 MUST land before any change that makes the producer payload parseable. Reversing that order activates the §A.7 hazard in a live tree.

### F.2 M4 — generator quality

Implements REQ-HLR-009, 011. Satisfies AC-HLR-008, 017. Changes `pattern_key` narrowing and draft-identity derivation in `mapper.go`. Forward-looking only — the 7 existing duplicate pairs are grandfathered.

**§G q2 resolved (2026-07-28, AskUserQuestion): option (a) — exclude the `agent_invocation` event type in `isActionable`.** Rationale recorded in `spec.md` §G q2. The two M4 code changes landed in commit `b010bcfd9`:
1. **Narrowing (AC-008):** `isActionable` rejects promotions whose `pattern_key` event-type prefix is `agent_invocation`, via an `excludedEventTypes` set derived from `harness.EventTypeAgentInvocation`. Independent of the tier gate (an `agent_invocation` with `to_tier:auto_update` + high confidence is still rejected).
2. **Identity (AC-017):** `buildDraftID` drops the `<YYYYMMDD>` date segment → `PROPOSAL-<sha256(pattern_key)[:8]>`.

Downstream cascade (commit `efcb4990c`): `TestRunHarnessObserveStop_ProposeChainAutoRuns` (SPEC-HARNESS-RATCHET-REWIRE-001 AC-HRR-005) seed switched `agent_invocation` → `tool_failure` — the intended AC-008 consequence. Orchestrator independent verification: full `go test ./...` exit 0 (105 ok), `GOOS=windows` build ok, `golangci-lint` 0 issues.

### F.3 M3 — dispatch observation

Implements REQ-HLR-005. Satisfies AC-HLR-007. **Correction (2026-07-28):** the original premise — "what is missing is the orchestrator-side obligation to call it" — was stale. The obligation, the CLI writer, the opt-in self-gate, and the unit tests were all shipped by `SPEC-HARNESS-EVOLVE-001` M3 (commit `1c54cd9c6`, 2026-07-12), 15 days before this plan was written: the obligation lives at `.claude/skills/moai/SKILL.md` (router section) and `workflows/run.md:196`; `moai harness ledger record` self-gates on `isHarnessLearningEnabled` (`internal/cli/hook.go:469`, default ON). M3's actual deliverable (user-approved Option A): verify the mechanics end-to-end, execute the AC-HLR-007 falsification, and correct this stale premise. The real residual gap is that the obligation is LLM-obeyed (no mechanical dispatch-time backstop) — documented as residual risk in `acceptance.md §E.1`, not closed by M3.

### F.4 M5 — lesson channel

Implements REQ-HLR-006, 007. Satisfies AC-HLR-009, 010. Requires resolving `spec.md` §G question 3 (migrate `lessons.md` into topic files, or restore it as an index).

**§G q3 resolved (2026-07-28, AskUserQuestion): option (a) — designate the topic-file convention (`feedback_*.md` + `MEMORY.md` index) as the single lesson store; `lessons.md` (47 KB, 40+ days stale) becomes a `[SUPERSEDED]` legacy artifact (content not migrated).** Rationale recorded in `spec.md` §G q3. The two M5 deliverables:
1. **AC-009 (lesson store):** edit `moai-constitution.md` § Lessons Protocol (local + template mirror + `make build`) to name the topic-file convention instead of `lessons.md`; mark `lessons.md` `[SUPERSEDED]`.
2. **AC-010 (inbox drain):** name the drain actor (orchestrator) + trigger in doctrine, and execute a drain that reduces the `.moai/lessons-inbox.jsonl` backlog (870 lines). The drain mechanism is doctrine-only today (no Go backstop — parallel to the M3 routing-ledger finding).

### F.5 M6 — falsifiability + CLI reporting  *(lowest reversibility — mechanical)*

Implements REQ-HLR-008, 010. Satisfies AC-HLR-011, 012, 013. Help-text completion, `list`/`doctor` agreement, and `prediction:`/`verified:` on this SPEC's own lesson entries.

### F.6 Ordering

M1 → M2 gates nothing downstream except by preference. M2 is sequenced first because it is the only milestone carrying a reversible architectural decision. M3–M6 are mutually independent. M4 reads more usefully after M2 (narrowing matters once a destination exists) but does not depend on it for correctness.

---

## §G Anti-patterns

- **Bridging the species gap.** Adding `target_path` / `field_key` / `new_value` to the observation schema to make drafts applicable. `spec.md` §A.4: those fields are the output of an authoring judgment, not of an observation gap. This is explicitly out of scope (`spec.md` §B.2).
- **Fixing the tier split.** Adding a `harness.Tier` JSON codec. Its only consumer is the path being removed; the split is dissolved, not repaired (AC-HLR-016).
- **Trusting the `unsupported fieldKey` branch as a guard.** It is unreachable for a content-free proposal — `createSnapshot` fails first (`spec.md` §A.7).
- **Running the live count from the worktree.** Reports `0 items` even after a correct fix (§B.13).
- **Bulk-promoting the queue.** Promotion is explicit and per-draft. 56% of the queue is bare tool names.
- **Retroactively rewriting the 52 drafts.** Out of scope, including the 7 duplicate pairs.
- **Claiming a falsification that was reasoned but not executed.** `acceptance.md` §A rule 3 and §J item 5.
- **Silently reversing a predecessor SPEC.** M2 reverses SPEC-HARNESS-APPLY-EXECUTE-001's wiring; the commit body must say so.

---

## §H Cross-references

- `spec.md` §A.3.2 (payload cause), §A.4 (species distinction), §A.6 (duplicate drafts), §A.7 (safety-pipeline hazard), §G (open questions)
- `acceptance.md` — AC SSOT, including the §J Definition of Done
- `progress.md` — M1 run-phase evidence (§E.2); manager-develop-owned, not edited by this plan
- `.claude/rules/moai/core/verification-claim-integrity.md` §3 — the 5-section report format §E requires
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A–E delegation template (required at Tier L)
- `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol — the `prediction:` / `verified:` obligation M6 satisfies and this SPEC is the first test of
