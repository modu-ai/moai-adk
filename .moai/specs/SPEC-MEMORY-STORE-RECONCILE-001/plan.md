---
id: SPEC-MEMORY-STORE-RECONCILE-001
title: "Implementation plan — auto-memory store reconciliation and index-budget premise correction"
version: "0.3.0"
created: 2026-08-31
---

# Implementation plan — SPEC-MEMORY-STORE-RECONCILE-001

Sections are ordered by decision-reversibility: the decisions most likely to change on review come
first, and the mechanical steps sit at the bottom. Read §A-§C before §F.

## §A Context

Evidence base: `.moai/reports/t383/measurements.md` (tree `9328a5242`), including the iteration-1
corrections §M5a (unit split) and §M5b (the tool measures nothing without `--dir`). Card t383.

[HARD] **Every command in this plan has been executed in this worktree during plan-phase, or is
marked as not-yet-run.** Iteration 1 failed on four defects (D1, D2, D3, D5) sharing one cause: a
figure or a command carried forward without being re-run against the tree it will execute on.

Iteration 2 added a second, sharper cause, and it is the one to watch: **a fix can install a new
error while removing the old one.** D1's repair replaced a wrong unit with a wrong *relation* and
then mandated it `[HARD]` for copying into doctrine; the byte-cap correction resolved a two-form
conflict in a direction that emptied AC-MSR-002 over an empty set. Both were introduced by the
correction, not inherited. The discipline that follows: after fixing a figure or a criterion,
**re-derive everything downstream of it** — C3's arithmetic, M0's population, the AC that reads it —
rather than assuming the fix is local.

## §B Decisions taken, and what would reverse them

### B.1 The spine is store divergence, not index length — and the card is split

**Decided.** M5 is the dominant defect and is this SPEC's spine. Index dieting is declined outright
(REQ-MSR-012). The store work is split: the **deterministic** half lands here, the
**judgement-bearing** half is deferred.

The figure, stated with **scope and unit** — this plan has now carried two wrong relations, rev 0
("58 of 123 = 47%", an occurrence count read as a line count) and rev 1 ("58 occurrences resolving
to 44 unique files", the gap attributed to deduplication when dedup is worth 2). Re-measured
independently in this worktree (`.moai/reports/t383/measure-n1.sh`):

| Scope | occurrences | unique targets | unique **missing** |
|---|---|---|---|
| Entry lines only (`^- \[`) | 137 | 135 | **44** |
| Whole file, any line shape | 192 | 190 | **58** ← `moai memory doctor`'s count |

Derived: 14 missing targets reachable only from non-`^- \[` lines; 40 entry lines affected, 0 of
them partially. **The gap is a line population, not deduplication** — `grep -c '^- \['` does not
match the grouped shapes carrying the other 14 (spec.md §A.2.1, §A.2.2).

| Half | Content | Where |
|---|---|---|
| Deterministic | The **58** unique missing files, whole-file scope. The index already names them — the operator's own curation. Copying them from the legacy store needs no per-file judgement, and success is a number (58 → 0). Scope matters here: copying only the 44 entry-line-reachable ones leaves `moai memory doctor` reporting 14 and AC-MSR-009 unachievable. | This card, M3 |
| Judgement-bearing | 46 orphans (index it or archive it?), the 177-vs-50 topic-file cap, disposition of the legacy store's 1,098 files. Every item is a per-file "is this memory still wanted?" question. | Deferred, follow-up card |

**The copy set is all 58, decided rather than assumed** — spec.md §A.2.2 carries the decision and
its four reasons. The 14 split 6 live working-discipline lessons / 8 archive roll-ups, and both
groups are copied: a roll-up names a representative file the reader is meant to open, and excluding
the 8 would force AC-MSR-009 off an exact 0 onto a maintained exclusion list.

**What this card delivers alone**: a session can open every file its index names; no in-repo file
still teaches a phantom cap; the write-anyway instruction sits where a writing session will meet
it; the guard can no longer be green because it measures nothing. **What it defers**: 46 memories
remain unreachable-by-index and the store is further over its topic-file cap (spec.md C3). Both are
in spec.md §D so the follow-up is not rediscovered from scratch.

**Reversal — resolved, not pending.** If the legacy-only files were predominantly superseded,
copying would be wrong and the correct remedy would be triage instead. The M0 gate (§F) tests this
mechanically. The auditor sampled 10 files independently and found all 10 read as live lessons, so
the copy direction stands; M0 re-runs the gate on the record rather than inheriting that result.

### B.2 Requirement 4 — the `(session: <8-char>)` markers are NOT restored

**Decided.** No retroactive restoration. `session-handoff.md`'s forward-looking obligation is left
untouched, so entries written from now on keep carrying it.

**Why.** The marker's value is correlating an entry back to a source session, and that value
survives in the topic file, which is where a reader who wants it goes. Against the live defect —
32.5% of entry lines leading nowhere — a session id on a dead link is worth nothing.

**On the count.** An earlier revision said "19 markers" were stripped. That figure appears in no
measurement and in no re-runnable command, and this SPEC's HISTORY pledges attribution for every
figure, so it is withdrawn. What is measured, on this machine at plan time:

```bash
grep -c '(session: ' "$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md"
# → 6   (dated reference, 2026-08-31; the command decides)
```

How many were removed is unknown and is not needed: the decision is "do not retro-restore", which
does not depend on the count.

**Reversal.** If a workflow emerges that reads the index alone and needs the session id, the
decision inverts and restoration becomes a small mechanical follow-up. No such workflow was
identified.

### B.3 The regression guards — bidirectional, and what is honestly not guarded

**G1 (repo-side, mechanical, bidirectional).** Remove the `MEMORY.md` fixed slot from
`alwaysLoadedSurface`, together with `memoryHead`, `memoryHeadLineCap`, and `memoryHeadByteCap`.
Add a test asserting every fixed surface slot names a path **present in** the repository tree.

- **Root resolution is specified, not left open.** The new test resolves its root exactly as the
  package's existing real-tree tests do — `findRepoRoot(mustGetwd(t))` plus the `t.Skip` guard at
  `token_budget_guard_test.go:53-59` and `:210-213` — and asserts **only against the real tree**.
  Run against a `t.TempDir()` fixture (the pattern at `:293`) all three remaining fixed slots are
  absent and the test would fail for a reason unrelated to the defect.
- **Hermeticity tension is reconciled in the file, not left implicit.** The guard's doc comment
  (`token_budget_guard.go:164-166, 170-172`) says fixed slots are enumerated even when absent, so
  a tree missing one still measures. That stays true — it is about *measurement*. The new rule is
  about *enumeration*: a slot may be absent from a **user's** tree, but a slot naming a path that
  is absent from **this repository** measures nothing here, forever. M2 adds one sentence saying so.
- **Red direction (demonstrated, not asserted):** re-insert the `MEMORY.md` slot → the new test
  fails. The run-phase records that output and reverts. Without the demonstration the test is a
  one-directional check that can never go red — the exact shape M3 exhibits.

Why removal rather than repair. The slot's caps encode the unconfirmed 25,600-byte cut (M2), so
"repairing" the guard to enforce them would harden an unverified premise into code. Pointing it at
the real store is not available either: the store is machine-specific and outside the repo, which
is the hermeticity the guard's own comment cites. REQ-MSR-007 keeps the removal honest as an
**edit-scope invariant** rather than as "the total will not move" — an absent file already
contributes 0, so a no-move assertion is true by construction and falsifies nothing. What is
checked instead: the surface **list** before and after differs by exactly one element.

**G2 (store-side, measured, bidirectional by construction).** `moai memory doctor --json` is the
domain's dedicated tool and it enumerates both candidate stores. Run-phase records its output
**before** the reconciliation (58 dangling occurrences — the red state) and **after** (0 — green).

[HARD] Every invocation carries `--dir "$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory"`,
and every count read from it is gated on `exists: true` **and** `index_lines > 0` in the same JSON.
Measured reason (measurements.md §M5b, reproduced from this worktree): the store is derived from
`CLAUDE_PROJECT_DIR` else `os.Getwd()` (`internal/cli/session.go:263-272`); `CLAUDE_PROJECT_DIR` is
unset here and run-phase executes inside this worktree, so the bare invocation slugs to
`-Users-goos-MoAI-moai-adk-go--claude-worktrees-t383` and **both** candidates return
`exists:false, index_lines:0, findings:null`. A finding count of 0 from that invocation means
nothing was measured — it would satisfy a naive "after-count is 0" criterion while touching nothing.

**What is honestly NOT guarded.** Nothing in this repository can mechanically prevent the live
store from re-diverging tomorrow: it is outside the tree and no CI job can see it. A hermetic Go
test against a fixture would prove only that the detector works, which `linkage_test.go` already
proves. The residual is caught by re-running `moai memory doctor` — which REQ-MSR-003 puts into
doctrine — and by nothing else.

### B.4 Where the doctrine correction goes — measured loading scope

**Measured fact, not an assumption.** `.claude/rules/moai/workflow/moai-memory.md` carries
`paths: "**/.moai/specs/**,**/.claude/agent-memory/**"` (frontmatter lines 1-3, read on this tree).
It is paths-scoped. A session writing a topic file into the memory store touches neither path, so
**the file this card corrects does not load for the session it targets** — REQ-MSR-004 as
originally scoped was undeliverable.

**Decided.** Do not widen `paths:`. The store lives outside the repository, so no repo-relative
glob can name a write into it; widening would enlarge the file's context cost without reaching the
target session. Instead REQ-MSR-004's write-anyway clause goes into
`.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol — verified to have no frontmatter
(the file opens `# MoAI Constitution`), therefore always-loaded, and already the surface that tells
a session to capture a lesson. REQ-MSR-001..003 still land in the paths-scoped file for the reader
who opens it deliberately.

**The occurrence sweep — the reversal condition fired, and here is the handling.** The earlier
revision said "more than one further occurrence tiers the card up rather than silently widening it".
The sweep has now been run and it fired:

```bash
grep -n '25KB' .claude/rules/moai/workflow/moai-memory.md
# → 29, 117, 175   (three assertions in one file)
grep -rn '25KB\|25 \* 1024\|25600\|200-line' .claude/output-styles CLAUDE.md AGENTS.md internal/config
# → .claude/output-styles/moai/moai.md:165   (an ALWAYS-LOADED fixed slot)
#   internal/config/token_budget_guard.go:95,98  (the constant, removed by M2)
```

The card is **not** tiered up, and the reason is stated rather than assumed: the extra occurrences
are three more lines in a file already being edited plus one line in an always-loaded file, i.e.
+2 file pairs, not a new problem class. Had the sweep surfaced a new subsystem, the tier-up would
have stood.

**The file count, enumerated rather than asserted** (an earlier revision said "four sources plus
four mirrors", which is wrong — `internal/config/` has no template mirror, so the totals happened
to agree only by silently counting the test file):

| # | File | Mirror? |
|---|---|---|
| 1 | `.claude/rules/moai/workflow/moai-memory.md` | yes → 5 |
| 2 | `.claude/rules/moai/core/moai-constitution.md` | yes → 6 |
| 3 | `.claude/output-styles/moai/moai.md` | yes → 7 |
| 4 | `internal/config/token_budget_guard.go` | **no** |
| 8 | `internal/config/token_budget_guard_test.go` | **no** |

**Four sources + three mirrors + one test file = 8 non-artifact files**, inside Tier M's 5-15 band.
`progress.md §E.1` states the same count and is the surface to keep consistent with this table.

### B.5 Template-First mirror — [HARD], and the failure mode is silent

**Measured.** All three doctrine files are byte-identical to their template mirrors today:

```bash
diff -q .claude/rules/moai/workflow/moai-memory.md internal/template/templates/.claude/rules/moai/workflow/moai-memory.md          # rc=0
diff -q .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md      # rc=0
diff -q .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md                          # rc=0
```

`CLAUDE.local.md` §2 is [HARD]: a local `.claude/` edit requires the mirror plus `make build`. Per
§2.3 the next `moai update` otherwise **overwrites the corrected local file with the stale template
copy** — a path-collision overwrite invisible to `git status | grep '^ D'`, so the correction
reverts silently. The earlier revision omitted the mirror entirely *and* AC-MSR-012's permitted-path
list actively forbade it.

**Second consequence, specific to this card.** The R4 block is not copied verbatim into the mirror.
It carries a machine-specific store path and dated internal figures, both of which are template
neutrality classes the CI guard scans for. spec.md §A.3 states the two forms; the mirror gets the
self-resolving form with no reference values.

## §C Pre-flight

All read-only; batched in one turn. Every line below was executed in this worktree during
plan-phase.

```bash
D="$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory"
L="$HOME/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory"
wc -c -m -l "$D/MEMORY.md"
grep -c '^- \[' "$D/MEMORY.md"
moai memory doctor --json --dir "$D"          # --dir is MANDATORY (B.3 / M5b)
grep -rn '25KB\|25 \* 1024\|25600\|200-line\|200 lines' \
  .claude/rules .claude/output-styles CLAUDE.md AGENTS.md internal/config \
  internal/template/templates/.claude
```

The sweep list is broadened from the earlier revision's `.claude/rules internal/config`, which
could not see `.claude/output-styles/moai/moai.md:165` or any mirror.

## §D Constraints

Per spec.md §C. The two that shape every step: **no memory-store file is a commit target** (C1),
and **every `.claude/` edit carries its mirror plus `make build`** (C6).

## §E Self-verification

Each milestone names its own check; the aggregate is acceptance.md.

## §F Milestones

### M0 — pre-flight and premise sampling (blocking)

1. Run §C verbatim; write raw output to `.moai/reports/t383/preflight.md`.
2. **Sample the legacy-only targets, by a stated rule, over the RIGHT population.** Take the **58**
   unique missing targets (whole-file scope — the population M3 actually copies) sorted
   lexicographically; select **every 5th**: indices 1, 6, 11, … 56 → exactly **12** files.
   Deterministic and re-runnable; no discretion in selection.

   Two arithmetic corrections from iteration 2, both material. The population was 44, which left
   the 14 outside-only targets structurally unsamplable while M3 copies them. And "every 4th up to
   n=10" over 44 reaches only index 37, so entries 38-44 were unreachable within its own
   population — the auditor was right about that arithmetic and the earlier objection to it was
   wrong: the n=10 cap makes it exactly 10 items, not 11. Every 5th of 58 reaches index 56; the
   residual 57-58 is stated in the record per AC-MSR-015 item 4 rather than left silent.
3. **Per-file decision rule.** A sampled file is **superseded** iff *either* its body carries a
   `[SUPERSEDED by …]` marker *or* its frontmatter `description:` is subsumed by an existing
   active-store file (the subsuming file named in the record). Otherwise it is **live**. Both arms
   are checkable by another reader against the same files.
4. **Stop threshold.** ≥ **4 of 12** superseded → return a blocker to the orchestrator: the copy
   direction is wrong and the work belongs in the deferred triage card. ≤ 3 → proceed. (Holds the
   same ~30% share as the earlier 3-of-10, rounded away from proceeding.)
5. Write `.moai/reports/t383/m0-sample.md`: the 12 selected paths and the rule that produced them,
   a per-file verdict with its deciding evidence, the superseded count, the comparison against the
   threshold, and the coverage limit (which of the 58 the sample did not reach).
6. Report the occurrence count from the broadened sweep (B.4).

### M1 — doctrine correction (REQ-MSR-001..004, REQ-MSR-014)

1. `.claude/rules/moai/workflow/moai-memory.md` — rewrite **all three** occurrences (lines 29, 117,
   175), not only 117. Each becomes the R4 local form (spec.md §A.3). Add REQ-MSR-002's compression
   rule and REQ-MSR-003's two-store hazard including the `--dir` + `exists` requirement.
2. `.claude/output-styles/moai/moai.md:165` — rewrite the "MEMORY.md 200-line/25KB loading" phrase
   so it no longer asserts the cut; it may keep pointing at the doctrine file.
3. `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol — insert REQ-MSR-004's
   write-anyway clause (B.4). The clause refers to the store **by the command that resolves it**,
   never by a literal path, so gap G4 is neither corrected nor re-asserted.
4. Mirror all three edits into `internal/template/templates/.claude/…`, using the mirror form of
   the R4 block (no machine path, no dated figures, no SPEC ID).
5. `make build`.
6. `diff -q` each of the three pairs; all three must return rc=0.

### M2 — vacuous-guard removal (REQ-MSR-005..008)

1. Record the enumerated surface **list** and the total before the edit
   (`alwaysLoadedSurface` + `measureAlwaysLoaded` against the real root).
2. Remove the `MEMORY.md` fixed slot, `memoryHead`, `memoryHeadLineCap`, `memoryHeadByteCap`, and
   the test assertions that supply a MEMORY.md fixture; correct the two hardcoded surface counts
   (`wantRuleCount + 4` → `+ 3`, and the temp-tree count).
3. Add the fixed-slot existence test (G1), resolving its root via `findRepoRoot(mustGetwd(t))` with
   the existing `t.Skip` guard and asserting only against the real tree.
4. Add one sentence to the guard's doc comment reconciling hermeticity with the new rule (B.3).
5. Leave exactly one named tombstone comment recording why there is no MEMORY.md slot, so the next
   author does not re-add it; AC-MSR-006's grep exempts that one comment by name.
6. Re-measure the surface list and total; the list must differ by exactly one element.
7. Run the mutation: re-insert the slot, show the new test failing, revert.
8. `go test ./internal/config/...`

### M3 — dangling-link reconciliation (REQ-MSR-009..011, REQ-MSR-015)

1. Record BEFORE: `moai memory doctor --json --dir "$D"` → `.moai/reports/t383/reconcile-before.json`.
   Assert `exists: true` and `index_lines > 0` in that JSON before reading any count.
2. In the **same measurement window**, record the index byte count and `grep -c '^- \['` entry count.
2b. **Record `MEMORY_ORPHAN_NOT_INDEXED` before AND after, alongside the dangling count** (gap G8).
   Every copied file is index-named, so none *should* become an orphan and the count should hold —
   but nobody has measured it and no acceptance criterion covers the orphan count moving. Recording
   both readings converts an assumption into an observation. A move in either direction is reported
   in the verdict as a finding, not silently absorbed: it would mean the copy step had a side effect
   the card did not predict. Same `--dir` invocation, same `exists`/`index_lines` gate, capital
   `.Code` selector.
3. Derive the target list from the active index's own link targets absent from the active store;
   copy each from the legacy store. Copy-only; never overwrite; never touch the legacy store.
4. Record AFTER with the same `--dir` invocation and the same `exists` gate →
   `.moai/reports/t383/reconcile-after.json`, plus the index byte and entry counts.
5. Compare as a **delta**: entry count after ≥ entry count before; any byte increase attributable
   only to entries a concurrent session added (C2), never to an edit by this milestone, which edits
   the index not at all.
6. `git status --short` shows no path under either memory store.

### M4 — evidence and close

Write `.moai/reports/t383/verdict.md` in the five-section evidence format (Claim / Evidence /
Baseline-attribution / Gaps / Residual-risk), citing commands run and their verbatim output. The
Gaps section names the deferred items from spec.md §D and the two gaps in §I below.

## §G Anti-patterns

- **Citing a command you have not run in this worktree.** The cause of iteration 1's D1/D2/D3/D5.
  A command copied from another card, another tree, or an earlier revision is a hypothesis.
- **Fixing a figure without re-deriving what reads it.** Iteration 2's N1: correcting 58→44 left
  C3's arithmetic, M0's population, and a `[HARD]` verbatim sentence all carrying the old relation.
  A corrected number is not a corrected document.
- **Shipping a criterion you just rewrote without re-running it, in both directions.** The entry
  above catches *propagation* — a fix that leaves stale readers downstream. This one catches the
  *self-inflicted* case: the repair itself plants a fresh defect inside the very criterion written
  to prevent that class. The base rate on this card is **five for five** — three plan-audit rounds
  plus two more planted by the run-phase repairs themselves — which makes it the most reliable
  prediction the card holds about itself:

  | Round | The fix | The defect it planted |
  |---|---|---|
  | iter-1 → 0.2.0 | corrected the 58-of-123 unit error | planted a wrong *relation* (dedup, worth 2 not 14) and mandated it `[HARD]` for copying into doctrine |
  | iter-2 → 0.3.0 | resolved the two-form R4 conflict | emptied AC-MSR-002's universal over an empty set — the SPEC's own named failure mode, inside the criterion written to enforce R4 |
  | iter-3 → run | rewrote AC-MSR-016's lint gate | planted N15: a criterion that was **written and never executed**, vacuous on empty output and cwd-dependent in the other direction |
  | run M2 | wrote the tombstone explaining why the 25KB caps were removed | the explanation **quoted the literal `25KB`**, so `token_budget_guard.go` failed AC-MSR-001 — the criterion the same milestone existed to satisfy |
  | run DEBT 2 | wrote AC-MSR-016's cwd anchor as `cd "$(git rev-parse --show-toplevel)"` | the worktree-isolation guard **refuses** that command, so the repair for an un-run criterion was itself un-runnable in the session that must run it |

  The discipline: after rewriting a criterion, **execute it** — once on a known-passing input and
  once on a deliberately-failing one — before the round is called done. Every defect above would
  have surfaced on first execution; none was caught by re-reading. Reading a criterion tells you
  what it says; only running it tells you what it *measures*, and the gap between those two is
  where all five defects lived.

  The two run-phase rows are the sharpest evidence for the rule, because both were introduced by
  a repair whose *stated purpose* was to close the very hole it reopened, and both were invisible
  to review: the tombstone reads as a correct explanation, and the `cd` anchor reads as the
  textbook way to anchor a working directory. Neither survived first execution.
- **Sweeping for the numeral and missing the idea.** The `25KB` grep does not find "approximately
  half", "roughly 25 kilobytes", or "the 200-line ceiling" written in words. A leakage sweep hunts
  the **claim**, not its digits: grep the numeral, then re-read the surrounding prose for the same
  assertion spelled out. `.moai/reports/t383/measurements.md` carried "approximately half dead
  links" through two numeral-only sweeps.
- **Reading a JSON field without checking its case.** `.code` and `.Code` both parse; one returns 0.
- **Treating a non-zero grep exit as "no match".** `grep` exits 2 on file-not-found.
- **`moai memory doctor` without `--dir`.** It measures an absent store and returns a passing 0.
- **Dieting the index anyway** because it is right there and looks long. REQ-MSR-012 forbids it.
- **Dropping an index line to "fix" a dangling link.** That converts a repairable link into a
  permanently unreachable memory. The repair direction is copy-the-file, always.
- **Deleting from the legacy store** to "consolidate". It is the only copy of 1,098 files whose
  disposition is undecided.
- **Repairing the guard by enforcing 25,600 bytes.** That hardens the premise this card removes.
- **Editing a `.claude/` file without its mirror.** The revert is silent (B.5).
- **Pinning a live-mutating literal in an acceptance criterion.** The store has concurrent writers.
- **Committing store files.** They are outside the repository (spec.md §A.4).

## §H Cross-references

See spec.md §F.

## §I Recorded gaps — NOT closed by this card

These are stated so they are not mistaken for resolved. Neither is claimed fixed.

### G2 — `moai spec lint` is UNMEASURED for this SPEC at plan time

The installed binary is **20 commits behind this tree** for `internal/spec/`:

```bash
git log --oneline 343399d2f..HEAD -- internal/spec | wc -l   # → 20
```

A lint run from PATH is therefore judged by a build that does not carry this tree's rules
(`verification-claim-integrity.md` §2.2 — tool provenance). Plan-phase did run
`moai spec lint` from PATH and observed zero findings naming this SPEC; that observation is
recorded as **weak evidence from a stale build**, not as a pass. acceptance.md §D.4 makes
"no new lint findings" a quality gate, and closing it requires `make build` and invoking the
tree's own binary **by path** (`./bin/moai spec lint`) with the lint's own exit code read — not
a pipeline exit code from `head` or `grep`.

### G4 — doctrine-vs-code divergence on store derivation

**Three** in-repo statements disagree with the code, and this card edits all three files:

| Surface | Statement | Fault |
|---|---|---|
| `.claude/rules/moai/workflow/moai-memory.md:27` | derives from the **git repository root**, so all worktrees share one memory directory | contradicted by the code |
| `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol | `~/.claude/projects/{project-hash}/memory/` | names the **legacy** store |
| `.claude/output-styles/moai/moai.md:165` | `~/.claude/projects/{hash}/memory/` | names the **legacy** store — **and M1 step 2 edits this exact line** |

The code derives the path from `CLAUDE_PROJECT_DIR` else `os.Getwd()`
(`internal/cli/session.go:263-272`), which is what produces M5b.

The third row is the one an earlier revision missed, and it is the most dangerous: M1 rewrites that
line's **cap** clause, so it is trivially possible to leave its **path** clause standing inside a
sentence this card just touched. Rewriting a sentence while preserving a wrong claim inside it is
re-assertion, not preservation. [HARD] Every edit to these three files refers to the store **by the
command that resolves it**, never by a literal path (spec.md §D). Deciding which side is wrong —
doctrine or code — is a follow-up card.
