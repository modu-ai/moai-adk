# SPEC-INSTRUCTION-FILES-UNIFY-001 — implementation plan

## §A Context

Card t1243, worktree `.claude/worktrees/t1243`, branch `WT-instruction-files`, base
develop `553e224f3`. Tier L, class C. Plan phase only; the run phase is blocked (§B).

Scope as of v0.3.0: this SPEC carries 16 requirements and 20 criteria — the contract, the
budget ceilings, the template-mirror invariant, and the guard/learner surfaces. The nine
requirements touching a user-owned file are `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`'s (card
**t1259**), and §C.1 below states the one place the two SPECs touch the same function.

## §B Blocking dependencies and baseline

**[HARD] t1175 (rules diet) must land on develop before the run phase starts.** t1175 is
rewriting the always-loaded rule tree, which is the same surface this SPEC's `AGENTS.md`
reconciliation touches. Running both concurrently produces a merge whose combined state
neither card measured. `AC-IFU-025`'s always-loaded clause is what turns a collision between
the two into a red test rather than a silent overrun.

**The baseline is re-measured after t1175 lands.** Every figure in research.md §B was read
against `553e224f3` and is attributed to that tree only. After t1175 lands, the byte counts,
section counts, and symbol locations are re-read before any milestone begins — they are the
inputs to REQ-IFU-018 and REQ-IFU-019, and a carried-over figure would make the cap check
unattributed.

**Source locations are cited by symbol, not by line.** `origin/develop` is roughly 93 commits
ahead of this worktree's base and has already moved one cited location: card t1224
(PowerShell deny parity) shifted `frozenInstructionFiles` by two lines while leaving the
symbol intact. Any line number in these artifacts is illustrative of 2026-09-26 against
`553e224f3`; the symbol is the address. Absorption happens at the integration window, not now.

## §C Sequencing risks

### C.1 The sibling SPEC — one shared function

`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` holds the Codex `CLAUDE.local.md` fallback branch and
its deprecation advisory (REQ-IFU-007, REQ-IFU-008). Both SPECs therefore edit the launcher's
local-instruction loop — the `for` range over `codexClaudeLocalName` /
`codexLocalInstructionName` in `internal/cli/codex_launcher.go`. This SPEC changes the
**iteration order**; the sibling adds the **advisory on the fallback branch**.

The two edits are compatible but not independent, so they are ordered rather than parallel:
this SPEC's M2 lands the order change first, and the sibling's advisory work builds on the
reordered loop. Running both lanes concurrently against that one function is the case to
avoid, and the lead's dispatch decision is what prevents it.

### C.2 SPEC A

The approved design names SPEC A (codex factory retirement) as overlapping this work at
`AGENTS.md` §3 and §8, and recommends A landing first with B absorbing it. **SPEC A does
not exist as a card.** Neither order is assumed here:

- If A lands first, M3 (the `AGENTS.md` body rewrite) absorbs its §3/§8 changes and the
  reconciliation is done once.
- If this SPEC lands first, A rewrites §3/§8 again on top of the reconciled body, and the `-f`
  removal from §8 that this SPEC performs may be re-touched.
- If they land concurrently, the §3/§8 region conflicts.

Surfacing this is the plan's job; choosing is the lead's.

## §D Constraints

- **Template-First.** `internal/template/templates/` changes first, then `make build`, then
  the root copies. Never the reverse.
- **Two mirrors, two commands.** Every deployed-file check runs against both paths with
  separate exit codes (design.md §F).
- **Every test assertion carries `-v` and a `--- PASS:` read.** Exit `0` alone is not evidence
  a test ran (acceptance.md preamble).
- **No sweep-staging.** Explicit pathspec only; `git status --short` re-read immediately
  before staging.
- **The lane does not push.** Integration is a lead-granted window; push is the lead's
  batch.

## §E Milestones

Ordered by decision-reversibility: the unsettled design decisions first, then the mechanical
code changes, then the body rewrite that depends on both.

### M1 — settle the two unmeasured questions (measurement, no code)

Two independent measurements, both recorded in `progress.md` §E.2 with command and verbatim
output **whichever way they come out**. A negative result is a result; neither is retried
until it produces the convenient answer.

**M1a — real-worktree ancestor discovery (AC-IFU-021).** Answers research.md Q1 and Q2. The
M0 P7 line-3 observation is a synthetic-fixture observation and, per the lead's ruling, may
not serve as a premise — this measurement is what settles it. Method mirrors the M0 probes
(headless `claude -p`, distinct sentinels, throwaway repository) with one change that is the
whole point: the tree is created by a real `git worktree add`, so `.git` in it is a file
rather than a directory. Outcome routes per design.md §A.5.

If discovery IS confirmed, the finding is handed to card **t1219** with its evidence, and
**the handoff itself is recorded** — acceptance.md §D.3 carries it as a conditional Definition
of Done item, so it is not left to memory. This SPEC does not resolve the duplicate.

**M1b — Codex discovery of `AGENTS.local.md` (AC-IFU-022).** Answers research.md Q5. The
criterion cannot be "run Codex and see that it works": the failure mode is silent tail
truncation, so a passing session proves nothing. Four sentinels — head and tail in both
`AGENTS.md` and `AGENTS.local.md` — make truncation observable, and `CONTRACT_HEAD` is the
**positive control**: its absence means the render failed and no other branch may be read. A
`LOCAL_HEAD` hit blocks — it means the content is counted twice and the design's budget
arithmetic is wrong.

Gate: M2 does not start until both measurements' evidence is on disk. M1a's outcome does not
gate M2 (design.md §A.5: every branch leaves the primary-checkout shape unchanged); M1b's
does, because a positive result changes the byte budget M3 has to fit inside.

### M2 — Codex read order, guard set, learner target

Three independent, mechanically small changes, each with its test:

- `codex_launcher.go` local-instruction loop: **iteration order only** (REQ-IFU-006). The
  fallback advisory on the same branch is the sibling SPEC's (§C.1) and is not written here.
- `pre_tool.go` `frozenInstructionFiles` gains two basenames (REQ-IFU-013). The set matches on
  basename, so two strings suffice — no path normalization is implied.
- `curator/dispatch.go` Tier-3 path (REQ-IFU-014).

**A new test is written here, not assumed:** nothing in `internal/hook` currently guards
`frozenInstructionFiles` (verified 2026-09-26 — see acceptance.md `AC-IFU-016`), so M2 creates
`TestFrozenInstructionFiles` with one sub-case per entry. Adding the two basenames without it
would leave the guard set unguarded.

Invert the `codex_contract_link_test.go` `@AGENTS.local.md`-imports assertion and the
`codex_local_instructions_test.go` read-order assertions here, not later — leaving them for
M3 would put the tree red across two milestones.

### M3 — the `AGENTS.md` body rewrite and the `CLAUDE.md` thinning

Depends on M1 (the worktree paragraph the body must state) and on the §C.2 sequencing
decision.

- Reconcile root (8 sections) against template (12) onto one **section set** (REQ-IFU-018);
  decide whether the four template-only reference sections belong in a contract
  (research.md Q3), under both ceilings.
- **[HARD] Do not touch the mirror's filename.** It is `AGENTS.md.tmpl`, and the suffix is
  what keeps it out of Codex's filename discovery. Card t925 renamed it after measuring the
  failure: chain sum 33,738 / 32,768, the mirror's last section dropped and the preceding
  one cut mid table row, no warning, exit 0. Reconciliation means the section set, never the
  name (design.md §D.2, AC-IFU-004).
- Content is NOT reconciled: the two mirrors diverge intentionally — measured 2026-09-26
  against `553e224f3` as 57 template-only and 17 root-only lines — so the check is a
  section-set diff (AC-IFU-008), not a byte diff.
- Update the touched contract constants in `internal/cli/codex_contract.go`:
  `codexLinkAgentsDirective`, and `codexCreatedClaudeBody` / `codexCreatedAgentsBody` — the
  created stub must emit the new two-import shape or it will not satisfy AC-IFU-002
  (design.md §C).
- Correct the budget statement (REQ-IFU-017); keep the truncation statement.
- **Neutralize the harness-restricted clauses in §1-§7 (REQ-IFU-018, second clause), and add the
  test `TestCodexContractLink_LocalImportMatrix` (REQ-IFU-002, `AC-IFU-012`).** Measured 2026-09-26
  against `553e224f3`, each mirror carries exactly **one** offending paragraph: §3's worktree-entry
  clause, which names `moai cc -w <name>` / `--spawn` with no Codex counterpart. The fix is to name
  **both** launchers in that clause — not to delete the Claude one, since a clause naming both
  harnesses is neutral, and a count-to-zero rule would have failed a correct implementation. Then
  drop `-f` from §8 and fix the §3 "Codex lanes" / `-w` inconsistency. The new test asserts both
  import directions (`AGENTS.local.md` in `CLAUDE.md` = 1, in `AGENTS.md` = 0) in one symbol; the
  existing test's `CLAUDE.md` assertion covers a different directive (`@AGENTS.md`), so this is new
  coverage rather than a rename.

  > Added at v0.3.2. The plan-audit of `4eb5405dc` found `REQ-IFU-016` named by **no** milestone
  > task and tested by no criterion — the only one of the 16 requirements absent from both `plan.md`
  > and `design.md`. The requirement is now retired and its substance absorbed as `REQ-IFU-018`'s
  > second clause (spec.md §C.4), so this task is the owner it previously lacked.
  >
  > [HARD] **This task has no criterion, and that is recorded rather than hidden.** No criterion
  > measures a clause, so closing it means recording the offending paragraph's before/after text in
  > `progress.md` §E.2 — named debt item 1 in `acceptance.md` §D.3.1. Do not report it discharged on
  > `AC-IFU-003`'s import count: that is a declared proxy and says nothing about clause wording.

- Thin `CLAUDE.md` to import + mechanism layer + import (REQ-IFU-002).
- Add the C5 neutrality allowlist change (REQ-IFU-015).
- `make build`, then both mirrors, then **three** budget checks, all with `--- PASS:` reads:
  per-file 24,576 (AC-IFU-005), nested-sum 32,768 (AC-IFU-006), and the always-loaded surface
  (AC-IFU-025). They are three different limits with three different owners and none
  substitutes for another. **The third is run package-scoped locally** —
  `go test ./internal/config/ -run '^TestAlwaysLoadedTokenBudget$' -v` — because
  `AC-IFU-025`'s own command is the CI full suite (`make build && go test ./...`), which §F
  forbids running locally. The criterion is discharged by CI; M3 needs the guard's local
  `--- PASS:` read, and that invocation is what supplies it. Raising `project_doc_max_bytes` is not an available remedy — the
  override is silently ignored until the user is `trusted`, and a distributed user's first
  session is untrusted by construction.

## §F Self-verification

Per-milestone: the affected packages only (`go test ./internal/<pkg>/...`), never
`go test ./...` locally. The full-suite verdict is CI's, on the PR head, in a clean
environment.

[HARD] **Any criterion added or edited during the run phase is anchored by hand.** The mechanical
judge for that class is **card t1269** (`VacuousAssertionRule`) and it has not landed, so nothing
will catch an unanchored `-run` pattern or an undelimited `--- PASS:` line in the interim —
acceptance.md §D.3.1 item 2. Apply the head-block rule by reading it, not by running a check.

At close: the acceptance.md §D.2 traceability diff command, plus a separate re-run of every
two-mirror criterion.

**[HARD] At close, additionally: the exhaustive noun-comparison pass** (acceptance.md §D.3). This is
a run-phase task, not a review nicety, and it is scoped inside this SPEC rather than deferred as
debt. For each of the 15 requirements, open **every** criterion citing it and record in
`progress.md` §E.2 one row per pair: requirement id, criterion id, the noun the requirement
constrains, the noun the criterion measures, and the verdict (`match`, or `proxy` with the
criterion's own stated reason).

[HARD] **The pair list is the deliverable; a statement that the pass ran is not.** The task is
discharged by that table and by nothing else, and it is written so a later reader can tell a
completed pass from an unstarted one without asking whoever ran it: a declared pair total above the
table, a row count equal to it, every one of the 15 requirement ids present in the id column (a
requirement cited by nothing gets a row reading `no citing criterion`, never an absent row), and
every `proxy` verdict quoting the criterion's own stated reason. A mismatches-only table does not
discharge it — "no mismatches over 23 pairs" and "no mismatches over the 4 pairs I got to" read
identically. Full obligation: acceptance.md §D.3.

The family this pass closes — a citation that is faithful while the noun underneath it differs — is
currently a **sample and not an enumeration**: iter-1 enumerated five instances of its sibling
vacuity class, iter-2 found four more survivors after that repair, and iter-3 did not state whether
it read every pair. An unmechanizable class of unknown size carried as debt is exactly the shape that
survived three audits, which is why this pass stayed in scope rather than being deferred. The pass is
finite: 15 requirements, one open per citing criterion. Anything tree-dependent in a proxy's
reasoning is measured, not argued.

## §G Anti-patterns

- Grepping one mirror and reporting the tree clean.
- Reading the presence of `@AGENTS.local.md` in `CLAUDE.md` as evidence the import
  resolved. M0-1 forbids this; **AC-IFU-019** is the correct assertion — it asserts the
  sentinel arrived. `AC-IFU-020` deliberately asserts only that the *literal line survives*
  when the target is absent, which is the directive-presence check this bullet forbids using
  as resolution evidence; the v0.2.0 draft cited it here and would have routed the
  implementer to the wrong assertion at exactly the moment it was warning them off it.
- Reading a `go test` exit code as evidence the test ran. `-run` on a pattern matching nothing
  exits `0` and prints `no tests to run`; five criteria in the v0.2.0 draft passed that way.
- Guessing a test's name from its subject. Three of those five named a guard that already
  existed under a different symbol.
- Re-running M1a or M1b until it produces the convenient answer.
- Reading `AC-IFU-022`'s decision rule without its positive control. A render that captured
  nothing resolves to "`LOCAL_HEAD` absent" and looks like the benign answer.
- Treating the M0 P7 line-3 observation as settled. It is a synthetic-fixture observation
  and AC-IFU-021 exists because it is not evidence for either answer.
- "Tidying up" the `.tmpl` suffix on the template mirror, or adding a file named `AGENTS.md`
  under `internal/template/templates/`. This is the exact shape of the defect t925 fixed,
  and prose did not stop it the first time — AC-IFU-004 does.
- Resolving the worktree duplicate-load here. It is t1219's, and this SPEC's obligation is
  only not to make it worse.
- Writing the sibling SPEC's fallback advisory into the launcher loop while reordering it
  (§C.1), or citing a source line number as an address rather than a symbol (§B).

## §H Cross-references

- `.moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/` — the sibling SPEC (card t1259).
- `.moai/reports/t1243/m0/verdict.md` — the measurement this plan's M1 extends.
- `.moai/reports/t1243/plan-audit-iter1.md` — the audit of `1140bcd1d` that v0.3.0 answers.
- design.md §A — the option analysis M1 resolves.
- design.md §D.2 — the `.tmpl` invariant and the t925 measurement behind it.
- design.md §A.3 / card **t1219** item (1) — the worktree duplicate load, scoped out.
- `internal/config/token_budget_guard.go` — `CodexContractByteCeiling`.
- `.moai/docs/gitflow-integration-chain.md` — the integration window this lane uses.
