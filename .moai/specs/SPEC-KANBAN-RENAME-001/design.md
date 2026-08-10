---
id: SPEC-KANBAN-RENAME-001
title: "Design — decisions underlying the Factory Mode to Kanban Mode rename"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: cli
lifecycle: spec-anchored
tags: "rename, design, decisions, rejected-alternatives, template-mirror, behavior-preserving"
tier: L
---

## §A. What this file is

The decisions this rename rests on, each with the alternative that was rejected and the **measured** reason it was rejected. It is a Tier L artifact added at v0.2.0 with the promotion; `spec.md` carries the requirements those decisions produce, and `research.md` carries the raw measurements both cite.

A rename looks like a decision-free operation, which is exactly why its decisions go unrecorded and get re-litigated. Every section below therefore names what was *not* chosen. Where a choice was forced by a measurement rather than merely preferred, the measurement is cited by its `research.md` section rather than restated.

---

## §B. Vocabulary: what the rename covers, and what it deliberately does not

### B.1 The mapping is token-scoped, not word-scoped

**Decided.** The rename covers a closed set of **Kanban Mode tokens** — the flag pair, two environment variables, a package path, a state directory, a sentinel, a goal preset, a skill document path, a set of Go identifiers, and the prose phrase "Factory Mode". It does not cover the English word *factory*.

**Rejected — renaming every occurrence of the word.** Measured (`research.md` §D): the bare-word grep over `internal/` and `.claude/` matches **134** files, of which **108** are unrelated vocabulary — `clientFactory` throughout `internal/lsp/core`, the deliberate "Interface + Factory for Single Implementation" anti-pattern example, an `@MX:ANCHOR` renderer-factory comment, per-language rule files. Renaming those would be a behavior change dressed as a rename, and leaving them while claiming word-level completion would require a 108-entry allowlist — a judgment call wearing the costume of a mechanical check.

The consequence is that **the completion criterion is a property of the token set, not of the word**, which is what makes it mechanically decidable. The pattern was falsified before adoption: run against six trees this SPEC does not touch it returns zero every time, while the bare word still returns nine files in `internal/lsp` alone (`research.md` §D.2). A pattern that matched the unrelated vocabulary would produce a criterion no rename could satisfy.

### B.2 The `AC-FM-*` citations are vocabulary the rename must not touch

**Decided.** The **50** `AC-FM-` identifiers in the test corpus (`research.md` §G) stay verbatim.

**Rejected — renaming them for consistency.** They are citations to the acceptance criteria of `SPEC-FACTORY-MODE-001`, a closed record this SPEC does not amend. Rewriting a citation to match a rename in the citing document breaks the link to the cited one — the identifiers would then point at criteria that never existed under those names. The rename changes what the code is called; it does not get to edit history's index.

This is why the criterion is a **count** rather than an absence: `AC-KR-013` asserts the post-rename count equals the pre-rename baseline. An absence check would pass a corpus where they had all been renamed, and a spot check would pass one where some had.

---

## §C. The entry switch

### C.1 `-k`, and why the collision question is a gate rather than an assumption

**Decided.** The short flag is `-k`. The plan-phase probe was **run**, not reasoned about: `claude --help` on `2.1.226` defines the short-flag set `-c -d -h -n -p -r -v -w`, and `-k` is free (`research.md` §E).

**Rejected — assuming the letter is free because it is unusual.** The launcher **strips its own switch before passing the remaining argv through** to `claude`. A collision therefore does not produce an error; it produces silent shadowing — the user's `-k` reaches the launcher, the launcher consumes it, and the underlying flag never fires. A failure with no error message is the class of defect that survives a whole release, which is why the answer had to be measured rather than inferred.

**Rejected — dropping the M0 gate now that the answer is known.** The `claude` CLI surface drifts between versions and the run-phase tree may sit on a different one. `REQ-KR-003` keeps the gate, but its character changed at v0.1.1: M0 now **re-confirms a recorded answer** instead of resolving an open question. The recorded limitation travels with it — the probe pattern matches `-k ` and `-k,` renderings and would miss a `-k=<value>` form, so a null result is strong evidence and not proof.

**Rejected — silently picking a different letter if a collision appears.** The mapping was user-confirmed. A collision is a blocker to surface, not a letter to swap.

### C.2 No deprecation alias, and the basis is release exposure rather than judgment

**Decided.** No hidden `-f`, no `--factory` alias, no deprecation-warning path.

**Rejected — a compatibility window.** The default instinct for a user-facing flag rename, and it is wrong here for a measured reason (`research.md` §F): `internal/cli/factory.go` is **absent** from `v3.0.1`, the latest stable tag, and present only in `v3.1.0-rc.0` and `-rc.1`; `CHANGELOG.md` carries **zero** case-insensitive occurrences of `factory`. No released user can depend on `-f`, and no release note ever told one it existed. An alias would be dead code from the moment it shipped, carrying a permanent maintenance cost and a permanent question ("when can we remove this?") for a population of zero.

**The timing is the argument.** Renaming now costs one mechanical pass; renaming after a stable release costs a deprecation window. That asymmetry is the whole reason this SPEC exists ahead of the three board SPECs — `SPEC-KANBAN-BOARD-001`, `SPEC-KANBAN-WORKTREE-001`, `SPEC-KANBAN-BOOTSTRAP-001` — rather than inside any of them. Each declares this SPEC in its `dependencies:` and writes every identifier in its post-rename form, which is only sound because the rename lands first.

### C.3 No state-record migration

**Decided.** Records under `.moai/state/factory/` are not migrated, moved, or read.

**Rejected — migrating them.** A record is session-scoped and best-effort by design: a launch never depends on one, and an unreadable record resolves in the safe direction (run the check rather than skip it). Orphaned records under the old path are therefore **inert**, not a correctness hazard. Migration code would exist to move data whose absence is already handled, and would need its own failure path for a case the design has already declared harmless.

Note what is *not* claimed: this is not "migration is hard". It is "the value of the data is zero by construction", which is a stronger reason and a falsifiable one — it fails the moment a launch is made to depend on a record.

---

## §D. Behavior preservation as the governing constraint

**Decided.** The rename introduces no functional change: no flag semantics, no state-record schema field, no gate ordering, and no goal bound differs from the pre-rename behavior (`REQ-KR-023`). Only names and prose move.

**What "behavior-preserving" is taken to exclude, stated because the exclusions are where such claims fail:**

- **Test assertions.** Not one is added, weakened, or removed (`REQ-KR-013`). This is not conservatism for its own sake — a changed assertion makes the preservation claim *unverifiable*, because the suite is the instrument measuring preservation and adjusting the instrument mid-measurement destroys the reading.
- **Test function names.** These *do* change (`REQ-KR-011`), and the distinction from the point above is deliberate: a name is how a reader finds the test from the function under test, and leaving `TestParseFactoryFlag` beside `parseKanbanFlag` breaks that path. Names are navigation; assertions are the measurement.
- **`captureEnvState`.** Explicitly *not* renamed (`REQ-KR-006`), because its name carries no mode-specific token. Renaming it would be scope creep with a consistency rationale — the same rationale §B.1 rejects at tree scale, applied to one symbol.

**Rejected — verifying preservation with an affected-packages test run.** `go test ./internal/cli/... ./internal/kanban/...` is the tempting shape, and it is blind exactly where a rename touching template source does damage: the template guards live in `internal/template`. `go list ./...` reports **115** packages (`research.md` §I), so a narrow run measures a fraction of the surface. `REQ-KR-022` mandates the full suite, and the reason is a recorded incident in this repository rather than a hypothetical — a prior run-phase missed a cross-cutting template guard by testing narrowly.

**Rejected — reading the suite's exit code after a pipe.** Recorded as a design decision rather than a style note because it decides whether the preservation claim is observable at all: `go test ./... 2>&1 | tail -20; echo "exit=$?"` reports `tail`'s status, so a fully red suite prints `exit=0`. Every criterion whose exit code decides it redirects to a log and reads `$?` before any pipe, then asserts a `^FAIL` count against the **whole** log rather than its tail.

---

## §E. The template mirror

### E.1 Delta preservation, never convergence

**Decided.** For each mirrored `.claude/` ↔ `internal/template/templates/.claude/` pair, the rename must leave that pair's measured `diff` **unchanged except for the renamed tokens**. A pair that was byte-identical stays byte-identical; a sanitized pair retains exactly the content it had stripped.

**Rejected — a byte-parity invariant.** This is the premise the SPEC started from and it is **false at HEAD** (`research.md` §C): three of six pairs are byte-identical and three are *sanitized pairs* whose local copies carry `Updated:` dates and SPEC-ID-bearing paragraphs that the template copies strip. The measured deltas are 0, 0, 0 lines for `contract` / `run` / `goal` and 5, 7, 9 lines for `moaidoc` / `modeorch` / `qgates`.

The failure mode a parity invariant would cause is specific and worth naming: an implementer told the pairs are identical "restores parity" by copying the local file over the template file, which **re-introduces §25-forbidden content into template source** and trips the neutrality guard. The invariant would instruct the implementer to commit the violation.

**Rejected — closing the gap deliberately.** The sanitization is not drift. It exists because `CLAUDE.local.md` §25 Template Internal-Content Isolation forbids SPEC IDs and internal dates in distributed template source, and each stripped passage is a live instance of that rule. Converging the pair is not an improvement; it is the violation, arrived at by a tidier route. This is why `AC-KR-017` treats a pair that "became identical" as a **failure**.

**Rejected — trusting this document's classification at run-phase.** The classification is **time-varying**. Any commit touching either side moves it. M0 re-measures and captures the baseline before the first edit, because a delta baseline that was never recorded cannot be reconstructed afterward — the pre-rename state is gone by the time the comparison is wanted.

### E.2 The baseline is keyed by label, not by filename

**Decided.** The six diff baselines are written to `/tmp/base-<label>.diff` under stable labels (`contract`, `run`, `goal`, `moaidoc`, `modeorch`, `qgates`).

**Rejected — keying on the basename.** The contract document's basename **changes during the SPEC** — `factory.md` becomes `kanban.md` at M2. A baseline keyed on the basename is captured under one name and looked up under another, so the lookup misses. Under a naive comparison loop a missing baseline contributes nothing to the output, which reads as clean. `AC-KR-017` therefore carries two explicit `test -f` guards that print `MISSING` rather than letting the absence pass silently.

### E.3 The catalog is regenerated, and the reason it looks unaffected

**Decided.** `make build` runs after template source is edited and the regenerated `internal/template/catalog.yaml` is committed (`REQ-KR-020`).

**Rejected — concluding the catalog is unaffected because grep finds nothing in it.** Measured (`research.md` §J): `grep -c 'workflows/' internal/template/catalog.yaml` returns **0** and `grep -ci factory` returns **0**. Both zeros are true and both are misleading — the catalog indexes skill **directories** with a content hash, not files by path. Renaming `factory.md` inside `.claude/skills/moai/` changes the `moai` skill directory's hash. An uncommitted catalog leaves the committed tree stale and surfaces as a CI parity failure, having produced no local signal at all.

This is why `AC-KR-020` asserts the hash **changed** (a `^[+-].*hash:` count of at least 2) rather than merely that the build ran. An empty catalog diff means either the build did not run or the rename never reached template source — both are failures, and the assertion distinguishes them from success rather than from each other.

---

## §F. The completion grep and its blind spots

### F.1 Scope widened to `.moai/project/`, and why regeneration was rejected

**Decided.** The completion grep covers `internal/`, `.claude/`, and `.moai/project/`. The two project documents that name the renamed package are edited in place (`REQ-KR-024`).

**Rejected — the v0.1.0 scope of `internal/` plus `.claude/` alone.** Run whole-tree, the same pattern returns two further files (`research.md` §B.2): `.moai/project/codemaps/modules.md` and `.moai/project/structure.md`. Each names a package path and a flag the rename deletes, so under the narrow scope the completion criterion would have returned **0 while two documents described a package that no longer exists** — a green light over a known staleness.

**Rejected — excluding them on the grounds that `/moai project` regenerates them.** Two measured objections. Neither file is template-mirrored, so nothing inside this SPEC's Definition of Done regenerates them; and both passages are hand-authored Korean prose carrying measurement narrative — `structure.md` line 139 and `modules.md` line 246 each record a package-count measurement with its command and its correction history — which a regeneration pass would not reliably reproduce. Five lines of edits is cheaper than a documented staleness, and far cheaper than a regeneration that silently rewrites unrelated narrative.

### F.2 `.moai/project/` and `.moai/specs/` are deliberately not treated alike

**Decided.** `.moai/project/` is in scope; `.moai/specs/` is excluded so `SPEC-FACTORY-MODE-001/` survives verbatim as a closed record.

**Rejected — excluding both, for the symmetry of it.** They hold different kinds of document. A closed SPEC is a record of what was decided and delivered under the names in force at the time; rewriting it would falsify the record. A project document is a description of the tree **as it currently is**; leaving it stale makes it wrong. The two directories share a parent and nothing else, and `plan.md` AP-10 names the conflation as an anti-pattern precisely because the symmetry is tempting.

### F.3 What the token grep cannot see, and the criteria that cover it

Two residues are invisible to the pattern **by construction**, and each carries its own criterion rather than a pattern extension:

- **A bare `-f`.** Adding `-f` as a global alternative would match `rm -f`, `grep -f`, and `git commit -f` tree-wide. The check is file-scoped to the six contract documents instead (`REQ-KR-025`, `AC-KR-026`), with a measured positive control of **8** occurrences (`research.md` §H). Without it, an implementer who renames `--factory` → `--kanban` and leaves `-f` in the prose passes every other criterion while shipping documentation that advertises a flag the launcher no longer accepts.
- **Anything under `.moai/specs/`.** Excluded deliberately per §F.2; `AC-KR-024` covers it as a diff-emptiness assertion anchored to the baseline commit.

**Rejected — extending `$TOK` to cover them.** Both extensions would trade a bounded false-negative for an unbounded false-positive, and a completion criterion that fires on `rm -f` is a criterion that gets disabled.

**A third blind spot, measured at v0.2.0 and closed at v0.3.0** (`research.md` §H.4): the token pattern matches `.moai/project/codemaps/modules.md` at lines **157 and 246 only**. Lines 158 and 161 of that file — which carry `Factory 모드` in Korean and the path `internal/cli/factory.go` — match neither `[Ff]actory [Mm]ode` (the mode phrase is not in English there) nor `internal/factory` (the path is `internal/cli/factory.go`). At **file** granularity the criterion is sound, since lines 157 and 246 hold the file in the match set until they are fixed. At **line** granularity it was not: an implementer who edited only the two matching lines drove `AC-KR-028`'s token count to zero with two stale lines remaining.

**Rejected — carrying it as residual risk.** That was the v0.2.0 disposition, on the reasoning that closing it meant editing a criterion the audit had passed, with `plan.md` M2b's explicit four-line enumeration as the working mitigation. The reasoning does not survive its own premise. The objection that forces `$TOK` to be token-scoped is the **tree-wide** false-positive population — ~110 files of unrelated `factory` vocabulary (§B.1) — and that population does not exist across **two named files**: measured, the bare word matches those two files at five lines, and all five are Kanban Mode. So the stated ground for preferring disclosure over closure was weaker than it appeared; it borrowed a tree-scale objection for a two-file scope. `AC-KR-028` gains a third command, a bare-word grep bounded to the two files (baseline 5, target 0). No criterion and no requirement is added — an existing criterion is strengthened, which the Tier L budget permits and a twenty-sixth requirement would not (`spec.md` §B).

---

## §G. Out of Scope

### Out of Scope — decisions this file does not make

- The kanban board itself. The six columns, the card record, and the board state store belong to `SPEC-KANBAN-BOARD-001`; the per-card worktree lifecycle, holder liveness, and mutual exclusion to `SPEC-KANBAN-WORKTREE-001`; the `lead` and `run` session roles, the topology, bootstrap, and orchestration across N sessions to `SPEC-KANBAN-BOOTSTRAP-001`. All three build on the surface this SPEC renames. No board is designed here, and no decision above constrains one beyond the identifier vocabulary it establishes.
- Where the board keeps its state. `SPEC-KANBAN-BOARD-001` `REQ-KB-005` places it at `.moai/state/kanban-board/`, resolved at the **primary checkout**, and explicitly declines to reuse or amend the `.moai/state/kanban/` that `REQ-KR-009` gives the **per-tree** session record. §C.3's reasoning is what makes that separation correct rather than merely convenient: a session record is session-scoped and best-effort, so per-tree resolution is the right rule for it, while a board is a single origin every session must agree on. One directory name serving both would have carried two occupants under two invisible resolution rules. The sibling moved; this SPEC's path is deliberately unchanged (`spec.md` §C).
- The internal design of anything renamed. §C decides what the entry switch is *called*; it re-decides nothing about what it *does*.
- The repository's pre-commit gate defect (`moai gate` exiting non-zero on pre-existing ast-grep findings, blocking every commit). `plan.md` §C records the `SKIP_MOAI_PRECOMMIT=1` bypass as an environment constraint; the defect is tracked separately.

### Out of Scope — surfaces this rename does not reach

- `docs-site/`. Its four `factory` hits are all `ExecutionFactory` inside a Claude Code best-practices example table, one per locale (`research.md` §B.3). This SPEC commissions no docs-site work, and `AC-KR-023` asserts the tree is untouched.
- `CHANGELOG.md`. Zero occurrences (`research.md` §F), so there is nothing to rename and no released user to notify.
- The 108 files of unrelated "factory" vocabulary (§B.1).

---

## §H. Cross-references

- `spec.md` §A.2, §A.4, §A.5, §A.6, §D.1 — the requirements and verification surfaces these decisions produce.
- `research.md` — the commands and observed outputs every measurement above cites.
- `plan.md` §B (B-1 … B-6), §F (M0 … M4), §G (AP-1 … AP-10) — the same decisions as an execution order, and the route back into each rejected alternative.
- `acceptance.md` AC-KR-004, AC-KR-013, AC-KR-017, AC-KR-020, AC-KR-021, AC-KR-026, AC-KR-028 — the criteria that decide them.
- `CLAUDE.local.md` §2 (Template-First), §14 (env constants), §25 (Template Internal-Content Isolation).
