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

**(6) Whole-change CI criterion — authored as `AC-IFU-031`.** As authored at v0.2.0 it read from
the PR head's own CI run and asserted a docs-site build — **both superseded**: re-sited at v0.2.1
onto the `origin/develop` head carrying this lane's merge SHA, and the build clause dropped at
v0.2.2. This bullet is the v0.2.0 record; the current text is the criterion body. It cites the full requirement set
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

### plan-audit iter2 — PASS-WITH-DEBT 0.92, both remaining items repaired (v0.2.2)

Report: `.moai/reports/t1259/plan-audit-iter2.md` at `28476f1a9`, against tree `bac73d358`. Five of
six iter1 repairs confirmed closed and mechanically re-verified. Three judgment calls were
vindicated on the auditor's own terms and are not revisited: equal-delta over narrowing (it
re-measured all 24 section counts independently, byte-for-byte match, and worked the three
staleness cases), applying both D2 options (verified genuinely red at `bac73d358`), and **not**
splitting `AC-IFU-029` — both halves assert the same property on the same requirement, so splitting
would put two criteria on `REQ-IFU-008` and regress the one-to-one traceability just achieved.

**D3-residual — repaired by grounding the whole clause against the workflow files once**, which is
what the audit asked for after this criterion's evidence pointed at something absent twice.
Measured here, not carried:

```
$ grep -rn 'hugo\|vercel' .github/workflows/
(no output; exit 1)          # across all 20 workflow files
$ grep -n '^on:' -A4 .github/workflows/ci.yml
  push:  branches: [main, develop]
$ sed -n '229p' .github/workflows/ci.yml
  go test -json -coverprofile=coverage.out -covermode=atomic ./... > test-stream.json || rc=$?
```

The criterion now names what the develop head actually produces: `CI` (whose `test` job runs
`go test ./...`, covering `./internal/cli/...`) and `spec-lint` (which has its own `push` trigger on
`.moai/specs/**`).

**The docs-site build clause is dropped, not re-sited** — and the reason improved mid-repair. I
wrote "whether Vercel builds is unmeasured"; the orchestrator then measured it at `0d7c7e44e`
(`.moai/reports/t1259/d3-ci-surface.md`) and Vercel **does** build it, configured in-repo
(`docs-site/vercel.json`, `framework: hugo`, `buildCommand: hugo --minify --gc`). I corrected the
note rather than shipping a stale claim — asserting "unmeasured" about something now measured is
the same defect class one more time.

The measurement strengthens the drop rather than reversing it, with three grounded reasons in place
of one absent one: the deployment is a different system with a different head (not a job in the
Actions run the criterion names); `ignoreCommand` makes it conditional on the push touching
`docs-site`, so an unconditional assertion is false on most pushes; and `github.silent: true` means
it never reports onto the commit, so its result is not readable from the surface the other checks
are read from. The premise a re-sited clause would still need — **whether the Vercel project
deploys `develop` at all** — is project-side configuration absent from this tree and remains unread.

**A second gap closed by that same commit, and it changed my wording.** `develop` is **not
protected**: `gh api …/branches/develop/protection` → `404 Branch not protected`. So "every required
check passes" has no referent on `develop` — there is no required-check set, only workflow runs that
block nothing. The criterion names workflow runs by file, and a `[HARD]` note now records why that
phrasing is deliberate and warns against carrying `main`'s protected posture onto `develop`.

**But one finding the audit did not have changes what replaces it.** `docs i18n parity check` fires
on a `develop` push filtered to `docs-site/content/**`, so it *does* run for M4 — and naming it as a
required check would have repeated the defect in a third form, because it is **advisory by
construction**:

```
$ sed -n '71,74p' .github/workflows/docs-i18n-check.yml
          elif [[ "${{ github.event_name }}" == "push" ]]; then
            # push to main/develop: warn-only during Phase 1 rollout ...
            echo "strict=false" >> "$GITHUB_OUTPUT"
```

The file's own header says `ADVISORY ONLY — NOT BLOCKING`, Phase 1 of a declared rollout with 35
pre-existing drifts. So it is named at its real weight: it must have **run and its log been read**,
never that it passed. The path-filter coupling is stated rather than assumed.

The residual gap is written into the criterion rather than left implicit: M4 lands 24 locale files
whose content no *blocking* check inspects. `AC-IFU-023` is what actually decides that work; a
future card flipping the parity check to Phase 2 strict would close it, and this SPEC does not
pretend to.

Sidelight worth recording: those 35 declared pre-existing drifts independently corroborate `D1`'s
baseline — the `claude-md-guide.md` and `quickstart.md` divergence is known, tracked, and owned
elsewhere, which is why equal-delta rather than equality was the right shape.

**N1 — repaired, and the repair corrected an overclaim of my own.** The v0.2.1 alternation made
this file's `no tests to run` guard permanently silent, since one branch always matches. Split into
two commands, two symbols, two reads. I first wrote "two exit codes" — then measured, and it is
wrong:

```
$ go test ./internal/cli/ -run '^TestCodexLocalInstructions_FallbackAdvisory$' -v
testing: warning: no tests to run
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  0.813s [no tests to run]
exit=0
```

`go test` exits `0` and prints `PASS` on a selector matching nothing, so the exit code is not the
discriminator — the marker is, which is why the head block prescribes reading it. The criterion now
says so explicitly with this output recorded. The generalized lesson is in the v0.2.2 note: **an
alternation is the wrong shape for a conjunction** — `-run` takes a disjunction, so each added
branch weakens the selector's ability to report absence while the assertion it serves is "both must
pass". N symbols need N commands.

### Verification after these repairs

```
$ AC counter → 10 ;  grep -c '^\*\*AC-IFU-' → 10          (agree; COUNT unmoved)
$ diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
       <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
  (empty output, exit 0)
$ moai spec lint SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 → ✓ No findings, exit 0
```

No criterion was added or removed, so no baseline cascade applies (this SPEC remains an
absent-from-snapshot, report-only row).

### Gaps — carried forward, NOT closed by assertion

- **Whether the Vercel project deploys `develop` at all — unread.** The build command and git
  integration are in-repo and measured; which branches trigger a deployment is project-side
  configuration, absent from this tree. This is the premise a re-sited build clause would have
  needed, and the reason dropping was the sound close.
- **Both external readings decay silently.** Branch protection and Vercel project settings are
  mutable outside this repository, and nothing in the tree changes when they move. The two figures
  recorded in `AC-IFU-031`'s notes (`develop` unprotected; Vercel builds conditionally) are pinned
  to `0d7c7e44e` and are to be **re-read at close**, not cited from there.
- **Full `internal/cli` state unrun.** Only the two named selectors were run, per affected-packages
  discipline.
- Parent-SPEC M2 and t1175 landing states — measured unmet at `a9e5f9d5a`; unchanged.

**Tree divergence, second occurrence — reported, not absorbed.** `HEAD` moved `28476f1a9` →
`0d7c7e44e` mid-repair. Inspected before staging: one file,
`.moai/reports/t1259/d3-ci-surface.md`, disjoint from the three artifacts staged here; the writer is
the orchestrator closing the two gaps iter2 recorded. Its content contradicted a claim I had already
written, which is why the correction above exists — the divergence check is what caught it.

### plan-audit iter3 (scoped) — FAIL 0.88, the audit ceiling; all sites repaired (v0.2.3)

Report: `.moai/reports/t1259/plan-audit-iter3.md` at `8d73a2a88`. N1 and the whole D3 reasoning were
confirmed repaired — every factual claim in the new `AC-IFU-031` body reproduced, no stale
"unmeasured" survived, and dropping rather than re-siting was called correct for the reason given.
The N1 self-correction was judged to improve on the instruction that prompted it: the audit had also
written "two exit codes", which the measurement falsifies.

**The FAIL is that the D3 repair stopped at the criterion body.** `grep -rn "PR head"` found four
live normative sites neither repair reached, so `acceptance.md` §D.3 and `plan.md` §D/§E gave close
instructions contradicting the criterion they govern — and `plan.md` agreed with the superseded
wording. Third appearance of one defect, which is what made it blocking rather than a wording nit.
The lesson is the one the defect keeps teaching in a new place each round: **repairing a criterion
is not repairing the SPEC.** A criterion is cited from the sections that decide when it is read, and
those sections do not update themselves.

All four repaired to name the `origin/develop` head carrying this lane's merge SHA:

| Site | Was | Now |
|---|---|---|
| `acceptance.md` §D.3 | "`AC-IFU-031` is read from the PR head's own CI run" | develop head + merge SHA, with "not from a PR head" and why |
| `plan.md` §D Close | "whole-change CI on the PR head" | develop head, plus the no-PR consequence spelled out |
| `plan.md` §E | "The full-suite verdict is CI's, on the PR head" | the `test` job of `ci.yml` on the develop head |
| `plan.md` §E close note | "read from the PR head's CI run" | records the v0.2.1 re-siting as history rather than restating it |

**§D.3 gained the four close-time duties the earlier repairs established but never propagated.** Each
existed only inside the note of the criterion that produced it, which means the close would not have
performed any of them: the `no tests to run` marker read (N1's finding — the PASS line alone is
necessary, not sufficient); a **named home** for recording the docs-i18n log reading (`progress.md`
§E.4 — the criterion asks that the check *ran and was read*, and an unrecorded reading is
indistinguishable from none); the close-time **re-read** of the two decaying external readings
(`develop`'s protection state, Vercel's build configuration — both mutable outside this repo, both
pinned to `0d7c7e44e` as evidence of what was true then); and a handover pointer for the docs-parity
residual.

**The docs-parity residual gets a pointer, not a new owner** — per the audit's explicit judgment,
which I take as correct: `AC-IFU-023` is a real gate (24 greps, existence checks, equal-delta, all
decidable) even though it runs in the lane, so writing the residual in IS sufficient for this SPEC's
slice. The general condition already has an owner — **`SPEC-V3R3-DOCS-PARITY-001`**, named in
`docs-i18n-check.yml:74` as the Phase 2 strict-flip route. Requiring this card to own that flip would
be the scope creep refused twice already.

**N2 — both figures re-measured here, both were wrong, both corrected.**

```
$ grep -n 'strict=false' .github/workflows/docs-i18n-check.yml
75:            echo "strict=false" >> "$GITHUB_OUTPUT"      # the push branch
78:            echo "strict=false" >> "$GITHUB_OUTPUT"      # the PR/else branch
$ ls -1 .github/workflows/ | grep -c ''          # 19 files (an earlier `wc -l` counted . and ..)
19
```

Cited `:71-74` was the elif head plus comments, not the assignment; `:75` is the line. And 20 → 19.
Neither touches a pass condition, and both are exactly the class of error this SPEC has spent three
iterations being right about — a miscited line number in a document whose whole argument is that
cited evidence must resolve.

### The close condition, and one property of it worth stating

In place of a fourth audit the close condition is mechanical: `grep -rn "PR head"` over this
directory returning only past-tense historical hits, plus a read of the amended §D.3. Measured after
these repairs: **zero live normative assertions.** Every hit falls into one of four classes, and the
class — not a count — is what the reader judges:

| Class | Where it occurs | Why it is not a live assertion |
|---|---|---|
| **HISTORY row** | `spec.md` v0.2.1 and v0.2.3 rows | Describes a repair that happened |
| **Repair record** | `acceptance.md` v0.2.1 note; `plan.md` §E close note; this section's own was/now table | Quotes the superseded wording to say it was superseded |
| **Negation** | the amended §D.3 — "**not** from a PR head" | Asserts the opposite of the defect |
| **The condition's own statement** | `spec.md` and this section, both stating the grep | Contains the string by construction |

[HARD] **This grep can never return empty, and a reader expecting empty will misread a clean SPEC as
a failing one.** Two independent reasons, both structural rather than incidental: the sentence
stating the condition contains the search string, and every repair record of this defect must quote
the wording it removed or the record says nothing. So **the passing condition is "no live normative
hit", judged per hit by the table above — never "no output", and never a hit count.**

The count in particular is a trap worth naming: documenting the repair *raises* it. An earlier draft
of this section recorded "nine hits" and was falsified by the act of writing the paragraph that
recorded it. A threshold that moves when you describe it is not a threshold.

Recorded because a check whose passing state reads as failure gets worked around, and the only
workaround available here would be to obfuscate the string in the one sentence obliged to state it
plainly — trading a real defect for a hidden one.

### Verification after these repairs

```
$ AC counter → 10 ;  grep -c '^\*\*AC-IFU-' → 10          (agree; COUNT unmoved)
$ diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
       <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
  (empty output, exit 0)
$ moai spec lint SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 → ✓ No findings, exit 0
```

No criterion was added or removed, so no baseline cascade applies (this SPEC remains an
absent-from-snapshot, report-only row).

### Gaps — carried forward, NOT closed by assertion

- **Whether the Vercel project deploys `develop` at all — unread.** The build command and git
  integration are in-repo and measured; which branches trigger a deployment is project-side
  configuration, absent from this tree. This is the premise a re-sited build clause would have
  needed, and the reason dropping was the sound close.
- **Both external readings decay silently.** Branch protection and Vercel project settings are
  mutable outside this repository, and nothing in the tree changes when they move. The two figures
  recorded in `AC-IFU-031`'s notes (`develop` unprotected; Vercel builds conditionally) are pinned
  to `0d7c7e44e` and are to be **re-read at close**, not cited from there.
- **Full `internal/cli` state unrun.** Only the two named selectors were run, per affected-packages
  discipline.
- Parent-SPEC M2 and t1175 landing states — measured unmet at `a9e5f9d5a`; unchanged.

**Tree divergence, second occurrence — reported, not absorbed.** `HEAD` moved `28476f1a9` →
`0d7c7e44e` mid-repair. Inspected before staging: one file,
`.moai/reports/t1259/d3-ci-surface.md`, disjoint from the three artifacts staged here; the writer is
the orchestrator closing the two gaps iter2 recorded. Its content contradicted a claim I had already
written, which is why the correction above exists — the divergence check is what caught it.

### plan-audit iter3 (scoped) — FAIL 0.88, the audit ceiling; all sites repaired (v0.2.3)

Report: `.moai/reports/t1259/plan-audit-iter3.md` at `8d73a2a88`. N1 and the whole D3 reasoning were
confirmed repaired — every factual claim in the new `AC-IFU-031` body reproduced, no stale
"unmeasured" survived, and dropping rather than re-siting was called correct for the reason given.
The N1 self-correction was judged to improve on the instruction that prompted it: the audit had also
written "two exit codes", which the measurement falsifies.

**The FAIL is that the D3 repair stopped at the criterion body.** `grep -rn "PR head"` found four
live normative sites neither repair reached, so `acceptance.md` §D.3 and `plan.md` §D/§E gave close
instructions contradicting the criterion they govern — and `plan.md` agreed with the superseded
wording. Third appearance of one defect, which is what made it blocking rather than a wording nit.
The lesson is the one the defect keeps teaching in a new place each round: **repairing a criterion
is not repairing the SPEC.** A criterion is cited from the sections that decide when it is read, and
those sections do not update themselves.

All four repaired to name the `origin/develop` head carrying this lane's merge SHA:

| Site | Was | Now |
|---|---|---|
| `acceptance.md` §D.3 | "`AC-IFU-031` is read from the PR head's own CI run" | develop head + merge SHA, with "not from a PR head" and why |
| `plan.md` §D Close | "whole-change CI on the PR head" | develop head, plus the no-PR consequence spelled out |
| `plan.md` §E | "The full-suite verdict is CI's, on the PR head" | the `test` job of `ci.yml` on the develop head |
| `plan.md` §E close note | "read from the PR head's CI run" | records the v0.2.1 re-siting as history rather than restating it |

**§D.3 gained the four close-time duties the earlier repairs established but never propagated.** Each
existed only inside the note of the criterion that produced it, which means the close would not have
performed any of them: the `no tests to run` marker read (N1's finding — the PASS line alone is
necessary, not sufficient); a **named home** for recording the docs-i18n log reading (`progress.md`
§E.4 — the criterion asks that the check *ran and was read*, and an unrecorded reading is
indistinguishable from none); the close-time **re-read** of the two decaying external readings
(`develop`'s protection state, Vercel's build configuration — both mutable outside this repo, both
pinned to `0d7c7e44e` as evidence of what was true then); and a handover pointer for the docs-parity
residual.

**The docs-parity residual gets a pointer, not a new owner** — per the audit's explicit judgment,
which I take as correct: `AC-IFU-023` is a real gate (24 greps, existence checks, equal-delta, all
decidable) even though it runs in the lane, so writing the residual in IS sufficient for this SPEC's
slice. The general condition already has an owner — **`SPEC-V3R3-DOCS-PARITY-001`**, named in
`docs-i18n-check.yml:74` as the Phase 2 strict-flip route. Requiring this card to own that flip would
be the scope creep refused twice already.

**N2 — both figures re-measured here, both were wrong, both corrected.**

```
$ grep -n 'strict=false' .github/workflows/docs-i18n-check.yml
75:            echo "strict=false" >> "$GITHUB_OUTPUT"      # the push branch
78:            echo "strict=false" >> "$GITHUB_OUTPUT"      # the PR/else branch
$ ls -1 .github/workflows/ | grep -c ''          # 19 files (an earlier `wc -l` counted . and ..)
19
```

Cited `:71-74` was the elif head plus comments, not the assignment; `:75` is the line. And 20 → 19.
Neither touches a pass condition, and both are exactly the class of error this SPEC has spent three
iterations being right about — a miscited line number in a document whose whole argument is that
cited evidence must resolve.

### The close condition, and one property of it worth stating

In place of a fourth audit the close condition is mechanical: `grep -rn "PR head"` over this
directory returning only past-tense historical hits, plus a read of the amended §D.3. Measured after
these repairs — **nine hits, zero of them live normative assertions**:

| Site | Why it is not a live assertion |
|---|---|
| `spec.md:24`, `spec.md:26` | HISTORY rows describing the v0.2.1 and v0.2.3 repairs |
| `spec.md:37` | states the close condition itself — see the note below |
| `plan.md:162` | past tense: the wording "proved to name something this regime never produces" |
| `progress.md:106` | the v0.2.0 authoring record, explicitly marked **both superseded** |
| `progress.md:249` | quotes the old text inside the v0.2.1 repair record |
| `acceptance.md:436`, `:437` | the v0.2.1 repair note, quoting the superseded wording |
| `acceptance.md:515` | a **negation** — "not from a PR head" — in the amended §D.3 |

[HARD] **The close-condition grep can never return empty, and that is not a miss.** `spec.md:37`
states the condition, so it contains the search string by construction; a reader expecting zero hits
would read a correctly-closed SPEC as failing. The condition is therefore "no live normative hit",
judged per line against the table above — not "no output". Recording this because a check whose
passing state is misread as failure gets worked around, and the workaround here would be to obfuscate
the string in the one sentence that has to state it plainly.

### Verification after these repairs

```
$ AC counter → 10 ;  grep -c '^\*\*AC-IFU-' → 10          (agree; COUNT unmoved)
$ traceability diff → empty output, exit 0
$ moai spec lint SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 → ✓ No findings, exit 0
```

No criterion added or removed, so no baseline cascade (this SPEC remains an absent-from-snapshot,
report-only row).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
