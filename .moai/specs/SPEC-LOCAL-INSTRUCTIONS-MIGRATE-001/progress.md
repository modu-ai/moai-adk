# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

**Plan phase completed by card t1259, 2026-09-26 (SPEC v0.2.0). Audit-ready.** The SPEC was
created 2026-09-26 by carve from `SPEC-INSTRUCTION-FILES-UNIFY-001` at `1140bcd1d`, not by a plan
phase; this section records the plan phase that has now run. Every item the carve left open is
below, each applied-and-recorded or explicitly deferred with its reason. Nothing was absorbed
silently.

### Carried forward from the carve, unchanged

- SPEC ID regex check executed as Bash against the Go `specIDPattern`
  (`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`), output `PASS`. Directory-collision check returned `0`.
- Ids retain the `IFU` infix; the gaps (`001~006`, `013~019`, `023~025` on requirements) are the
  carve's footprint. New criteria take `029`-`031`, above the parent's current maximum (`028`,
  measured), so no id collides with the parent SPEC.
- Status remains `draft`. No implementation.
- **Sequencing constraint**, unchanged: `SPEC-INSTRUCTION-FILES-UNIFY-001` M2 must land before
  this SPEC's **M2** (renumbered — it was M1 at the carve). Both edit the local-instruction loop
  in `internal/cli/codex_launcher.go`. plan.md §B holds it.

### Decisions taken, item by item

**(0) The AC-counter discrepancy — closed as observed, and it was the parent's, not this SPEC's.**
The orchestrator's finding is confirmed by re-measurement rather than accepted. Against this
SPEC's `acceptance.md` at v0.2.0:

```
$ AC_FILE=.moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/acceptance.md bash <extracted MOAI-AC-COUNTER>
live=10 excluded=3 ambiguous=0
10
$ grep -c '^\*\*AC-IFU-' .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/acceptance.md
10
```

Counter output and declared count agree at **10**, and `ambiguous=0`. The three exclusions are
sibling-SPEC `AC-` tokens in prose, `[REF]`-marked at v0.1.1. The "23 vs 20" discrepancy belongs
to the parent SPEC (card t1243) and was fixed there at plan-audit iter3; it has no residue here.

**(1) `AC-IFU-014` vs `REQ-IFU-010` — applied. The requirement splits; refusal survives.**
The contradiction was re-observed on the current text before acting, not taken on the audit's
word: `REQ-IFU-010` commanded resolution, `AC-IFU-014` refusal, and in the both-files case a
correct implementation of either violates the other.

Decision: **refuse when both files exist, and fix the requirement.** Both files are user-authored;
a tool that picks one either discards content the user wrote or concatenates two documents whose
relationship it cannot know, and the failure is silent. Refusing costs one command; choosing wrong
costs a file. `REQ-IFU-010` now carries `010a` (resolution, unambiguous case) and `010b` (refusal,
both present) — two clauses under one id, because splitting the id would break the
cross-references the carve exists to preserve. Reasoning: `design.md` §A.

**The coverage-gap check the brief asked for was run, and no new criterion is needed.** Both
clauses already had one: `AC-IFU-013` fixtures the one-file case (`010a`), `AC-IFU-014` the
both-files case (`010b`). §D.2 now records them on separate rows so the mapping is visible rather
than inferred. `AC-IFU-014` did gain a sha256 before/after comparison — "without modifying either
file" was previously unasserted, so a refusal that wrote and *then* exited non-zero would have
passed.

Deliberately not decided: whether the verb should offer a `--force` escape. Nothing requires one,
and adding it re-opens the data-integrity question (`design.md` §A).

**(2) Milestone re-sequencing — applied.** Re-ordered by dependency, not by the parent's
numbering, with an explicit old→new mapping table in plan.md §D so nothing is lost by the
renumber. New order: **M1** migration verb + advisories (nothing depends on it; two milestones
depend on it) → **M2** Codex fallback advisory (blocked on the parent SPEC's M2; sequencing it
first would stall the card behind another card's merge) → **M3** this repository's own migration,
operator-gated (now explicitly downstream of M1, because it is performed **with** the verb) →
**M4** docs-site (last: it documents the verb name, the refusal behaviour, and the advisory text,
all settled by M1-M2; writing 24 files against an unlanded design is the expensive way to discover
a rename).

The brief's note about M1 opening with research is resolved by item (7) rather than by moving the
milestone: the research is answered, so the milestone no longer starts with it.

**(3) Tier — revised M → L.** The carve's provisional `M` rested on the requirement count, which
is not the binding axis. Measured file count is ~35 (24 docs-site files alone; `research.md` §D
carries the arithmetic) against a Tier L threshold of `> 15`. The LOC axis agrees:
`migrate_agency.go`, the precedent the verb is modelled on, is 25,790 bytes with a 30,580-byte
test file. Consequences applied: artifact set becomes 5 files (`design.md` and `research.md`
authored, deliberately thin and cross-referencing the parent for shared context), and the
plan-auditor PASS threshold rises 0.80 → **0.85**.

**(4) Both recorded debts — unfolded.** Each note stated the fold was a budget compromise at the
parent's Tier L 25/25 ceiling, "not tidiness", and asked that the split be decided by whoever
owned the scope. That owner is t1259 and the stated reason is gone (10 criteria against a ceiling
of 25), so the folds are undone along exactly the lines the notes drew:

- `AC-IFU-011` → `AC-IFU-011` (advisory, `REQ-IFU-007`) + `AC-IFU-029` (provenance preamble
  literal filename, `REQ-IFU-008`). The v0.1.1 both-end anchoring is preserved through the split.
- `AC-IFU-015` → `AC-IFU-015` (sha256 immutability, `REQ-IFU-011`) + `AC-IFU-030` (advisory parity
  across both commands, `REQ-IFU-012`).

Result: one-to-one requirement↔criterion coverage, which is what §D.2 prefers and what makes a
failure say which defect occurred.

**(5) `AC-IFU-015` promoted to blocking — applied, and it resolves *with* the unfold.** While the
criterion bundled one data-integrity outcome with two UX outcomes, promoting it would have made a
missing advisory block a milestone boundary — which is why the parent's table correctly left it
non-blocking. Unfolded, the blocking half is exactly the data-integrity assertion (`moai update`
altering a user-authored file) and nothing travels with it. `AC-IFU-030` stays non-blocking: a
missing advisory is recoverable, a mutated file is not. Blocking set is now `AC-IFU-013`,
`AC-IFU-014`, `AC-IFU-015`.

**(6) Whole-change CI criterion — authored as `AC-IFU-031`.** Read from the PR head's own CI run,
covering `go test ./internal/cli/...` and the docs-site build. It cites the full requirement set
and is **deliberately excluded from the §D.2 coverage table** — counting a whole-change criterion
as coverage would let it stand in for a missing per-requirement one. It is a close gate, not a
milestone gate. The parent's `AC-IFU-025` correctly stayed with the parent: it asserts that SPEC's
always-loaded budget clause, which this SPEC does not own.

**(7) Parent `research.md` Q4 — answered, cheaply, by reading this tree.** Yes, the fallback
branch has a diagnostic surface. `codexLocalDeveloperInstructionArgs` is a pure producer
(`(projectRoot string) ([]string, error)` — no writer, no `cobra.Command`) whose every output is
JSON-encoded into `developer_instructions`, so emitting inside it would pollute the payload. Its
**sole** caller (`grep` confirms one definition, one call site) is `runCodexLaunch`, which holds
`cmd.ErrOrStderr()` and already writes operator diagnostics to it — the install hint and the
worktree error both go there. Full readings: `research.md` §A; the plumbing consequence and the
preferred shape: `design.md` §B. plan.md M2 no longer opens with this task.

**(8) Every carved figure re-measured — and two do not reproduce.**

- **`CLAUDE.local.md` size.** `git show origin/develop:CLAUDE.local.md | wc -m` → **44,740** in
  this tree, 2026-09-26; `git show develop:CLAUDE.local.md | wc -m` agrees. The carve recorded
  **44,381** — short by 359 characters. The file is a live maintainer document that sibling cards
  keep editing, so **any** absolute figure written into a criterion about it is stale by
  construction. `AC-IFU-007`'s fixed reduction floor is therefore **withdrawn**, not re-repaired:
  the binding condition is `after <= 39,999`, which does not drift, and the before-value is
  measured at the milestone. The same off-by-one correction is applied in `REQ-IFU-021`. The
  recording obligation (both values, each with its command) is unchanged — it is what stops the
  criterion being discharged against an already-compliant copy.

  Worth naming as a class, not an arithmetic slip: v0.1.1 correctly repaired the floor's
  off-by-one, and that repair is what made a drifting constant look authoritative.

- **The docs-site paths.** `AC-IFU-023` cited six stems in `docs-site/content/{ko,en,ja,zh}/`.
  Measured: **no page sits at that depth** — every one is one or two directories deeper — so the
  glob matches the empty set and the criterion passed vacuously. Worse, `memory` resolves to
  **two** files per locale and the criterion named neither:
  `grep -c 'CLAUDE.local.md\|CLAUDE.md'` reports `0` for `ko/cli-reference/memory.md` (the
  `moai memory` CLI reference) and `28` for `ko/claude-code/context-memory/memory.md` (the Claude
  Code memory-file concept page). `REQ-IFU-020` now names all six by **path**, targeting the
  latter. Editing the wrong one would have satisfied a stem-based grep while leaving the
  documented structure untouched. Full readings: `research.md` §C.

- **Symbols still resolve.** `codexLocalInstructionName` / `codexClaudeLocalName`
  (`codex_contract.go:33-34`), the local-instruction loop (`codex_launcher.go:125`), and the
  `migrate_agency_*` precedent are all present. plan.md's cite-by-symbol discipline holds.

### Verification run at plan close

```
$ AC counter → live=10 excluded=3 ambiguous=0 ; grep -c '^\*\*AC-IFU-' → 10   (agree)
$ diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
       <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
  (empty output — passing condition)
```

**AC baseline snapshot — checked, not assumed, because this repository's CI has failed on it
three times (t1108, t1122, t1148/t1154).** The criterion count changed 7 → 10, which is exactly
the trigger. The guard lives in `./internal/spec` (`ac_baseline_commit_guard_test.go`,
`ac_baseline_check_armed_test.go`, `ac_count_clause_test.go`), so the guard was run rather than
reasoned about:

```
$ go test ./internal/spec/...
ok  	github.com/modu-ai/moai-adk/internal/spec	149.110s
```

Green with the edits in the tree, so **no snapshot is owed for this SPEC** — the baseline the
guard keys on does not enumerate it. Recorded with the command because a green run here is the
whole evidence for not shipping a snapshot.

### Left open, deliberately

- **No plan-audit has run against this SPEC.** That is the next step, at the Tier L threshold of
  0.85.
- The parent SPEC M2 dependency and the t1175 dependency (plan.md §B) are unmeasured here — they
  are the lead's read at dispatch, not plan-phase facts.
- The `.moai/docs/` split point for M3 (which sections of `CLAUDE.local.md` are procedure, which
  are rules) is M3's first act. Deciding it here would re-decide the document's content, which
  spec.md §D puts out of scope.
- Whether `TestCodexLocalInstructions_DualFileMatrix` currently passes was not run; `AC-IFU-029`'s
  own `--- PASS:` assertion settles it at run time.

### For plan-audit to judge

- The `REQ-IFU-010a`/`010b` sub-clause shape: two clauses under one id, to preserve
  cross-references. An auditor may prefer two ids at the cost of breaking them.
- Excluding `AC-IFU-031` from the §D.2 coverage table. Deliberate, and stated in the table.
- The Tier raise's consequence for the route: Tier L is Route B (PR per phase), which differs from
  the Route A the carve's `M` implied. This card is a lane card under the git-flow protocol, where
  integration is a lead-granted window into `develop` rather than a per-phase PR; the two
  regimes are not reconciled here and the lead's dispatch governs.

### plan-audit iter1 — PASS-WITH-DEBT 0.85, four blocking defects repaired (v0.2.1)

Report: `.moai/reports/t1259/plan-audit-iter1.md`, written at `b8fb023e8` against tree
`5ba87003f`. Tier L threshold 0.85; all seven must-pass criteria clean; traceability 1.00;
Testability 0.65 carried the four defects. All nine of the plan phase's own re-measurements
reproduced under the auditor's independent runs.

Three shapes the plan phase explicitly referred to the audit were **accepted** and are not
revisited: the `REQ-IFU-010a`/`010b` split under one id, excluding `AC-IFU-031` from the coverage
table, and withdrawing the fixed reduction floor (judged to *strengthen* the criterion).

**D1 — `AC-IFU-023` locale parity — repaired, equality replaced by equal-delta.** Re-measured in
this tree rather than carried: `advanced/claude-md-guide.md` ko=10 vs en/ja/zh=18;
`getting-started/quickstart.md` ko=14 vs 10; the other four pages equal. Two of six already
unequal, in opposite directions, by 8 and by 4 sections — so an equality assertion either absorbs
an unscoped twelve-section restructure into M4 or fails a correct implementation.

Chose **equal-delta against a recorded baseline** over the audit's narrower alternative. Narrowing
to "the sections the change adds" would have left M4 free to land 24 files whose locale structure
diverges *further* with nothing to catch it — the audit's own residual-risk note. Equal-delta keeps
the property the clause existed to protect (the four locales move in step) without asserting one
that was false before the change. The baseline table is recorded in `acceptance.md` and is
**re-measured at M4** against that milestone's base, not read from the table; what the table pins is
the *shape* (two pages unequal, four equal), so a differing shape is a signal to re-read the
criterion rather than to adjust numbers silently.

Deliberately **not** done, per the dispatch: widening `REQ-IFU-020` to cover the restructure. It is
real work needing an owner who can decide whether ko is missing eight sections or en/ja/zh carry
eight they should not — a question this SPEC has no basis to answer.

**D2 — `AC-IFU-029` discriminated nothing — repaired with BOTH of the audit's options.**
Reproduced at `5ba87003f` with none of the work done:

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED \
    && go test ./internal/cli/ -run '^TestCodexLocalInstructions_DualFileMatrix$' -v
--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.01s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli
```

`REQ-IFU-008` is satisfied at `codex_launcher.go:138`. It is now declared a **preservation
requirement** in `spec.md` §C.1 and `AC-IFU-029` is labelled a **regression guard** with that
baseline recorded — so a later failure is meaningful and §D.3's "all ten criteria pass" no longer
hands out a free pass. The audit offered declaration *or* extension; both are applied, because
declaration alone leaves the criterion unable to fail. The extension requires the new
`FallbackAdvisory` test to assert the preamble too, which is red today and closes the real risk the
preservation framing exposes: the advisory is emitted from the same launch path that builds the
payload, so this SPEC's own change is the one most likely to break that preamble.

**D3 — `AC-IFU-031` cited a head this lane never produces — repaired by re-siting the evidence.**
The criterion read "Given the whole change on its PR head" while `plan.md` §C states the lane does
not push. Under the git-flow lane protocol the card branch opens no PR; only `release/vX.Y.Z` PRs
to `main`, and that head carries many cards. Evidence re-sited on the `origin/develop` run carrying
this lane's merge SHA — how every other card here closes. The criterion's substance is untouched:
clean environment, both the Go suite and the docs-site build, a named head rather than "CI was
green". The §E.1 Route A/B disclosure below was the right call and is not what this repairs — the
disclosure lived here while the unsatisfiable obligation lived in `acceptance.md`, which is the
artifact the run phase reads.

**D4 — `AC-IFU-024` `grep -c` → `grep -n -C1`.** A count cannot show which occurrences they are or
what sentence surrounds them, so the assertion's second half was not decidable from the output the
verdict was read from. Passing condition restated per line.

**D5 (optional) — applied.** Measured the premise rather than accepting it: `doctor.go:74`
`out := cmd.OutOrStdout()` with `printer.New(printer.WithWriters(out, cmd.ErrOrStderr()))` at `:80`;
`update.go:153` likewise stdout. Recorded the rule this implies — **each command's own report
stream**, not "diagnostics go to stderr" — and why the launcher differs (its stdout is
payload-adjacent, so an advisory there becomes model context). `design.md` §B needs no change: it
scopes its stderr claim to the launcher's fallback branch.

**D6 (optional) — applied.** `REQ-IFU-012` now leads with `When`, matching its eight siblings.

### Gaps the audit left open — two closed, three not mine to close

**`moai spec lint` — CLOSED, and the auditor's budget problem is explained.** It did not finish in
the auditor's 120s because it was run without an argument, which scans the whole corpus. Scoped to
this SPEC it completes in seconds:

```
$ timeout 180 moai spec lint SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
✓ No findings — all SPEC documents are valid
exit=0
```

Run twice — before the repairs and after — with the same result. This is an observed pass, not an
absence of signal.

**AC counter — CLOSED, re-run here rather than carried from the orchestrator.** After the repairs:
counter `10`, `grep -c '^\*\*AC-IFU-'` `10` — agree. The criterion COUNT did not move (10 before,
10 after; no criterion was added or removed by these repairs), and this SPEC is an
absent-from-snapshot row and therefore report-only, so no baseline cascade applies. Traceability
diff re-run: empty output, exit 0.

**Blocking dependencies — CLOSED by the orchestrator, not by me, at `a9e5f9d5a`.** That commit
landed on this branch mid-repair (see the divergence note below) and records both measurements in
`.moai/reports/t1259/blocking-dependencies.md`: the parent SPEC is `status: draft` on
`origin/develop` AND its loop at `codex_launcher.go:125` still reads `codexClaudeLocalName` first,
so REQ-IFU-006 is absent from the code as well as the status; t1175 is likewise absent. **Both
run-phase dependencies in `plan.md` §B are unmet**, which does not affect the plan phase but does
bind M2's dispatch.

**Still open, and not this lane's to close:** whether `internal/cli` is green as a whole at this
HEAD — only the named tests were run, per the affected-packages-only discipline.

**Tree divergence observed and reported, not absorbed quietly.** `HEAD` moved from `b8fb023e8` to
`a9e5f9d5a` between the start of these repairs and staging. Inspected before proceeding:
`git diff --name-only b8fb023e8..HEAD` → one file, `.moai/reports/t1259/blocking-dependencies.md`,
disjoint from the three artifacts this commit stages. The writer is the orchestrator closing an
audit gap, not a stray session. Staging by explicit pathspec means the foreign commit cannot be
lost by this one. Recorded because a HEAD move on an actively-worked tree is reportable whatever
its cause.

### For iter2

- **D1's residual risk is reduced, not eliminated.** Equal-delta catches a page whose locales move
  by different amounts. It does not catch M4 leaving `claude-md-guide.md` and `quickstart.md` as
  unequal as it found them — by design, since repairing that is out of scope. If iter2 judges the
  pre-existing divergence itself unacceptable to ship alongside this change, the answer is a
  sibling card, not a wider `REQ-IFU-020`.
- **`AC-IFU-029` is now half guard, half verification.** That is deliberate and labelled, but it
  means one criterion carries two severities — the guard half cannot fail before the work, the
  extension half cannot pass before it. An auditor preferring one severity per criterion would
  split it; the Tier L ceiling (25) leaves room.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
