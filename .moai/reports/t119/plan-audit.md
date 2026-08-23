# SPEC-TODO-ANALYSIS-001 — Plan-Phase Audit (Tier M)

VERDICT: PASS
Score: 0.920 (Tier M PASS threshold 0.80 — `spec-workflow.md:141`)

Auditor: plan-auditor. Reasoning context ignored per M1 Context Isolation — this
audit reads only the four artifacts and the tree.

## Baseline — read this first, the target moved twice

This audit ran against a **live, concurrently-revised** SPEC. The sequence:

1. Audit opened against all four artifacts at HEAD `f1bc39310`.
2. Mid-audit, `spec.md` was revised to **v0.2.0** (uncommitted).
3. Before this report was finalized, `acceptance.md` and `plan.md` were **also**
   revised (uncommitted).
4. And then `progress.md` — resolving F15 after this report first raised it.

The final `git status --porcelain` at report time:

```
 M .moai/specs/SPEC-TODO-ANALYSIS-001/acceptance.md
 M .moai/specs/SPEC-TODO-ANALYSIS-001/plan.md
 M .moai/specs/SPEC-TODO-ANALYSIS-001/progress.md
 M .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md
?? .moai/reports/t119/plan-audit.md
```

`spec.md` v0.2.0's HISTORY row attributes the revision to another auditor's
finding set ("plan-audit iteration 1 결함 반영: … D2, D4, D6, D8, D9, D10, D11,
D13"), so a concurrent review is in flight and its findings substantially
overlap this one's.

**The verdict below is against the current working-tree state of all four
artifacts** — every claim was re-verified after each revision landed, including
the last.

**This report's own history matters, because it is evidence about the SPEC.**
Against the HEAD-`f1bc39310` state this audit found four blocking defects: an
acceptance criterion that could not pass on any tree, two commanded verbs with
no criterion at all, and a Tier premise that did not hold. Three of those four
are now fully resolved and the fourth is materially answered. That is not a
softened verdict — it is a different document. The blocking findings are
recorded in the Findings table with their resolution evidence so the fix route
stays auditable, and so a reader can see the FAIL→PASS movement was earned
rather than argued away.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-TA-001` … `REQ-TA-015`,
  sequential, no gaps, no duplicates, uniform 3-digit padding
  (`spec.md:122-148`).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement
  layer** (`REQ-TA-*` in `spec.md §C`), not the verification layer. All 15
  match a GEARS pattern: Ubiquitous (001, 006, 014, 015), Event-driven (002,
  003, 005, 007, 008, 011, 012), Where/capability-gate (004), Unwanted (009,
  010), State-driven (013). The Given-When-Then entries in `acceptance.md` are
  the correct verification-layer form and are graded under Group 4, not here.
  Granularity note as F11 — form is compliant.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present
  with correct types (`spec.md:2-14`): `id`, `title`, `version: "0.2.0"`
  (quoted semver), `status: draft` (valid enum — `spec-frontmatter-schema.md`
  § Status Enum), `created: 2026-08-22` / `updated: 2026-08-23` ISO, `author`,
  `priority: P1`, `phase: "v3.1.4 target"` (a release target, not a prohibited
  lifecycle-stage name), `module`, `lifecycle: spec-anchored`, `tags`
  comma-separated string. Plus `tier: M`. No rejected snake_case alias.
- **[N/A] MP-4 language neutrality** — single-project SPEC scoped to
  moai-adk-go's own `internal/cli` + `internal/kanban`. No language-specific
  tool names appear. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference,
  `SPEC-KANBAN-TODO-CLI-001`; its `spec.md` exists with `status: in-progress`
  (not retired/superseded/archived). No BLOCKING finding.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` = 0.
  D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over
  the SPEC directory = 0 matches. No `research.md` at Tier M; `plan.md` clean.

Budget check: 15 REQ / 16 AC against the Tier M ceilings of 16 and 16
(`spec-workflow.md` § SPEC Complexity Tier). Both within budget; the AC count
now sits **exactly at** the ceiling, leaving no room for a further criterion
without a tier decision (F16).

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.92 | 0.75-1.0 | Every requirement has a single reading. v0.2.0's scope qualifier (`spec.md:71-73`) stops the criterion from binding `normalizeBacklogRecord`'s legitimate repair; the two mis-cited line numbers are corrected. Deduction: `plan.md:20`'s narrows-not-relaxes framing still stands uncorrected beneath the ruling layered on it (F4). |
| Completeness | 0.90 | 0.75-1.0 | All sections present; 6 `### Out of Scope — <topic>` sub-headings each with `-` bullets; frontmatter complete; the residual risk on the refusal path is declared in both `spec.md:86` and `plan.md:53`; `progress.md §E.1` now carries an accurate count, the Tier ruling, a per-defect resolution table, and its own Gaps section. Deductions: no actor constraint on `relate` (F5), no always-loaded budget accounting (F13). |
| Testability | 0.90 | 0.75-1.0 | The unsatisfiable grep criterion is gone, replaced by a scoped guard with a retained negative control, and the discarded criterion is documented with its measured baseline. AC-TA-012's key-set argument is genuinely sharp (see C6). Deductions: NFC unverified (F9), one grep token with a non-zero baseline (F10). |
| Traceability | 0.95 | 0.75-1.0 | Every AC now carries a `**검증**: REQ-…` line; all 15 REQs are cited by at least one AC (verified mechanically, C7). `analyze` and `unrelate` both gained behavioural criteria. |

Aggregate 0.920.

---

## Evidence (5-section structure per verification-claim-integrity.md)

### 1. Claim

The SPEC's load-bearing factual claims about the tree — including the five new
source citations the revision introduces — plus this audit's own findings about
coverage, doctrine, and the always-loaded budget.

### 2. Evidence — command + verbatim output

**C1 — nine subcommands exist (spec.md:23-24).** CONFIRMED.

```
$ sed -n '206,208p' internal/cli/todo.go
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoNextCmd(),
		newTodoUnpickCmd(), newTodoEditCmd(), newTodoMoveCmd(),
		newTodoDropCmd(), newTodoUndropCmd())
```

add, list, done, next, unpick, edit, move, drop, undrop = 9. Accurate.

**C2 — no analysis layer anywhere in the queue path (spec.md:30).** CONFIRMED.

```
$ grep -rni "jaccard\|similarity\|duplicate\|NFC" internal/cli/todo*.go internal/kanban/backlog_store.go | grep -v _test.go
internal/cli/todo_drop.go:13:// card is still worth keeping — no staleness heuristic, no duplicate
internal/cli/todo_edit_move.go:17://   - move only permutes the item slice; no item is dropped, duplicated, or
internal/cli/todo_edit_move.go:118:duplicated, or altered, so a wrong move is reversed by another move. The
internal/cli/todo.go:128:	// Same-volume rename is atomic and leaves no duplicate behind.
```

Every hit is prose or an unrelated sense of "duplicate". No similarity,
normalization-comparison, or relation logic exists. Accurate.

**C3 — quoted source comments and their line numbers.** All three now correct.

```
$ grep -n "no staleness heuristic" internal/cli/todo_drop.go
13:// card is still worth keeping — no staleness heuristic, no duplicate

$ grep -n "Nothing here infers" internal/cli/todo_edit_move.go
5:// a correction the operator decided on. Nothing here infers what a card

$ grep -n "Order is the only thing" internal/cli/todo_edit_move.go
99:// Order is the only thing the queue records about priority — there are no
```

The current text cites `todo_drop.go:13`, `todo_edit_move.go:5-7`, and
`todo_edit_move.go:99`. All three resolve. (The `:17` and `:110` citations this
audit flagged at HEAD are gone.)

**C4 — the revision's two new source citations.** BOTH CONFIRMED.

```
$ sed -n '322,336p' internal/kanban/backlog_store.go
// normalizeBacklogRecord establishes the invariants every in-memory record
// holds: version 1, a non-nil item slice, and a high-water mark that clears
// every present id (max of persisted and max-present — REQ-TODO-009's
// derive-on-absent and the hand-edited-low-value guard in one rule).
func normalizeBacklogRecord(rec *BacklogRecord) {
	if rec.Version == 0 {
		rec.Version = backlogVersion
	}
	if rec.Items == nil {
		rec.Items = []BacklogItem{}
	}
	if max := maxPresentBacklogSeq(rec.Items); max > rec.LastSeq {
		rec.LastSeq = max
	}
}

$ sed -n '308,311p' internal/cli/todo.go
	for _, it := range rec.Items {
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s\n", it.ID, it.State, it.Text)
	}
```

`backlog_store.go:326` is exactly `func normalizeBacklogRecord`, and it does
silently raise `LastSeq` to the max present id — so the scope-qualifier example
at `spec.md:73` holds. `todo.go:309` is the human-render loop, and the `Fprintf`
prints exactly `ID`, `State`, `Text` — three columns, `added_at` absent — so the
`spec.md:84` claim that a reorder leaves no trace on the operator's default
screen holds too.

**C5 — AC-TA-013's replacement criterion and its discarded predecessor.** BOTH
CONFIRMED.

The predecessor criterion (`grep -rn "AskUserQuestion" internal/cli/` = 0) was
unsatisfiable. Measured:

```
$ grep -rn "AskUserQuestion" internal/cli/ | wc -l
     317
$ grep -rl "AskUserQuestion" internal/cli/ | wc -l
      82
```

`acceptance.md:139` now documents this verbatim — "실측 **317건 / 82파일**
(`f1bc39310`)" — and discards the criterion. Both figures match this audit's
independent measurement exactly, against the same SHA.

The replacement criterion names an existing guard, which exists as described:

```
$ grep -n "todoPromptGuard\|TestTodoCmd_NoAskUserQuestion" internal/cli/todo_test.go
447:// TestTodoCmd_NoAskUserQuestion — AC-TODO-014: the todo command surface
451:func TestTodoCmd_NoAskUserQuestion(t *testing.T) {
456:	if reason, bad := todoPromptGuard(string(data)); bad {
461:	if _, bad := todoPromptGuard("x := AskUserQuestion()"); !bad {
466:// todoPromptGuard reports whether source contains an interactive-prompt
469:func todoPromptGuard(source string) (reason string, bad bool) {
```

Line 461 is the synthetic negative control (`x := AskUserQuestion()`) that
`acceptance.md:137` requires be retained. The AC asks for a **scope extension**
of an existing guard, not a new one — and it says so explicitly
(`acceptance.md:138`).

**C6 — AC-TA-012's key-set argument.** CORRECT, and load-bearing.

`acceptance.md:129` argues that "decoding into the old-schema struct succeeds"
cannot catch an *added* per-item field, because `encoding/json` silently ignores
unknown fields — so the AC counts each item's key set instead, and deliberately
does NOT apply `DisallowUnknownFields()` at the top level, because the additive
`findings` array would redden a correct implementation there. Both halves of
that reasoning are right, and the resulting criterion is strictly stronger than
the round-trip form it replaced.

**C7 — REQ↔AC coverage.** COMPLETE.

```
$ grep -o "REQ-TA-[0-9]*" .moai/specs/SPEC-TODO-ANALYSIS-001/acceptance.md | sort -u | tr '\n' ' '
REQ-TA-001 REQ-TA-002 REQ-TA-003 REQ-TA-004 REQ-TA-005 REQ-TA-006 REQ-TA-007 REQ-TA-008 REQ-TA-009 REQ-TA-010 REQ-TA-011 REQ-TA-012 REQ-TA-013 REQ-TA-014 REQ-TA-015

$ grep -c "^### AC-TA-" .moai/specs/SPEC-TODO-ANALYSIS-001/acceptance.md
16
```

All 15 REQs are cited by at least one AC. 16 ACs against the Tier M ceiling of
16. The two previously-uncovered verbs now have criteria: `unrelate` in
AC-TA-004 (bidirectional relate→unrelate in one fixture, asserting exact
findings length in both directions plus `items` byte-identity), `analyze` in the
new AC-TA-016 (positive control "length > 0" after the first run, then exact
length equality after the second, plus `items` byte-identity).

**C8 — REQ-TODO-013 permits additive change (spec.md:51).** CONFIRMED.

```
$ grep -n "REQ-TODO-013" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
59:- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record shape — `{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}` with `state ∈ {queued, picked, dropped}` — changing it only additively (the high-water mark, per REQ-TODO-009).
```

The parenthetical names the high-water mark (`last_seq`) as the additive
precedent — exactly the precedent `spec.md:51` and `plan.md:38` invoke. A new
TOP-LEVEL `findings` array is within "changing it only additively"; the frozen
surface is the per-item field set:

```
$ grep -n "per-item\|five field" internal/kanban/backlog_store.go | head -2
46:// bump, and no per-item field may ever be added (spec.md §E out-of-scope).
61:// BacklogItem is one queued card. The five fields are the frozen per-item
```

`plan.md:85` now argues the store comment and its `spec.md §E` cross-reference
should both be left alone, because that reference points at
SPEC-KANBAN-TODO-CLI-001's own §E. Verified:

```
$ sed -n '92p' .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
- No version bump and no new per-item fields. The only schema change is the additive high-water mark (REQ-TODO-009); everything else stays load-compatible with the existing production file.
```

The citation resolves and the reasoning holds — re-pointing that comment at this
SPEC would aim it at a section that does not make the claim.

**C9 — REQ-TODO-014 headless contract.** Now quoted correctly in both places.

```
$ grep -n "REQ-TODO-014" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
63:- **REQ-TODO-014** (Ubiquitous) The `moai todo` command shall not prompt: it is headless-safe, asks no questions, and follows the `internal/cli` conventions (structured stdout for `--json`, human-readable stderr, exit 0/1/2, no AskUserQuestion anywhere in the package).
```

REQ-TA-015 (`spec.md:148`) now says "exit 0/1/2", and AC-TA-013's Then-clause
now accepts "`0` / `1` / `2` 중 하나". Both matched the narrower "exit 0/1" at
HEAD; both are corrected.

**C10 — the two colliding [HARD] clauses.** CONFIRMED, quoted accurately.

```
$ sed -n '46,53p' .claude/skills/moai/workflows/todo.md
[HARD] `edit`, `move`, `drop`, and `undrop` are operator acts, exactly like
`add` and the pick. Correct a card's wording, move it, or discard it because
the operator said to — never on inferred priority, never as tidy-up, never to
fold one card into another, and never because a card looks stale. The queue
records the operator's intent; it does not curate it.

The queue is never mutated through any other surface. A missing backlog file is
an empty queue, never an error; a malformed file is reported and left untouched.
```

`kanban-dispatch.md` § Entry into the board is an operator act carries "The lead
is the queue's sole producer", "Promotion is the operator's act, always", and
"never reorders by inferred priority" — matching `spec.md:45`.

**C11 — always-loaded token budget headroom** (this audit's finding, not the
SPEC's).

```
$ go test ./internal/config/ -run 'TokenBudget|AlwaysLoaded' -count=1
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.388s

$ bash scratchpad/budget.sh    # replicates alwaysLoadedSurface() enumeration
bytes=284850 tokens=71212 budget=76000 headroom=4788
```

`internal/config/token_budget_guard.go:32` — `AlwaysLoadedTokenBudget = 76000`.
`kanban-dispatch.md:5` declares itself "Intentionally always-loaded" and is
25,915 bytes. The M1 addition (~700 bytes ≈ 175 tokens) fits inside 4,788 tokens
of headroom. Coordination concern, not capacity — F13.

**C12 — AC-TA-014 grep-token baselines.**

```
$ grep -c "records" .claude/skills/moai/workflows/todo.md .claude/rules/moai/workflow/kanban-dispatch.md
.claude/skills/moai/workflows/todo.md:5
.claude/rules/moai/workflow/kanban-dispatch.md:1

$ grep -c "never folds" .claude/skills/moai/workflows/todo.md .claude/rules/moai/workflow/kanban-dispatch.md
.claude/skills/moai/workflows/todo.md:0
.claude/rules/moai/workflow/kanban-dispatch.md:0
```

"records" already returns hits pre-implementation and observes nothing on its
own; "never folds" has a clean 0 baseline and carries the AC — F10.

**C13 — Tier M sizing and threshold.**

```
$ sed -n '138,142p' .claude/rules/moai/workflow/spec-workflow.md
| Tier | Scope guidance (LOC) | Files affected | Artifact set | plan-auditor PASS threshold |
| S (Simple) | < 300 LOC | < 5 files | **2 files**: spec.md + plan.md (AC inline in spec.md §3) | 0.75 |
| M (Medium) | 300 - 1000 LOC | 5 - 15 files | **3 files**: spec.md + plan.md + acceptance.md | 0.80 |
| L (Large) | > 1000 LOC or constitutional | > 15 files | **5 files**: spec.md + plan.md + acceptance.md + design.md + research.md | 0.85 |
```

`plan.md:13-16`'s 12-14 files / 500-800 LOC / 0.80 threshold all match the SSOT.
This auditor concurs with Tier M — see F4 for the one framing correction still
outstanding.

**C14 — supporting claims in plan.md §C.** CONFIRMED.

```
$ grep -n '"expect"' internal/cli/todo*.go | grep -v _test
internal/cli/todo_drop.go:133 / :190 / internal/cli/todo_edit_move.go:91 / internal/cli/todo.go:440
$ grep -n "DROPPED" internal/cli/todo_drop.go | head -2
47:// already established on the live queue: `[DROPPED — <reason>] <text>`.
52:	todoDropMarkerOpen  = "[DROPPED — "
$ grep -rn "func resolveTodoQueueRoot" internal/cli/
internal/cli/todo.go:67:func resolveTodoQueueRoot() string {
```

`--expect` on four verbs, the text-marker precedent, and the queue-root test
seam (`acceptance.md:18`) all exist as described.

**C15 — `progress.md`'s remeasured-figures table.** ALL VERIFIED.

`progress.md:35-45` tabulates this audit's cited figures against the SPEC
author's own re-measurement. Every row was independently re-checked here:

| Figure | This audit | progress.md remeasure | Re-verified |
|---|---|---|---|
| `grep -rn AskUserQuestion internal/cli/` | 317 / 82 files | 317 / 82 files | match (C5) |
| REQ-TODO-014 exit contract | `0/1/2` at `spec.md:63` | same | match (C9) |
| "Nothing here infers…" | `:5` | `:5-7` | **their correction is right** |
| "Order is the only thing…" | `:99` | `:99` | match (C3) |
| `normalizeBacklogRecord` | `:326` | `:326`, raise at `:333-335` | match (C4) |
| `backlog_store.go:46` referent | KANBAN-TODO-CLI-001 §E | `spec.md:92` | match (C8) |
| `todo list` human render | `added_at` absent | `id`/`state`/`text` 3 cols | match (C4) |
| REQ / AC count | 15 / 16 | 15 / 16 | match (C7) |

Two supporting checks:

```
$ grep -n "AddedAt" internal/cli/todo.go | head -1
271:			AddedAt: time.Now().UTC().Format(time.RFC3339),

$ sed -n '5,7p' internal/cli/todo_edit_move.go
// a correction the operator decided on. Nothing here infers what a card
// should say or where it belongs — no analysis, no absorption, no silent
// promotion. Those would collide head-on with the [HARD] clauses in
```

`AddedAt` is written at creation (`:271`) and never rendered (`:309`), as
`progress.md:27` states. The quoted comment does span lines 5-7, so the SPEC's
`:5-7` citation is correct and this audit's `:5` was one line narrow.

A SPEC that re-measures an auditor's figures rather than accepting them, and
that corrects the auditor where the auditor was imprecise, is doing the thing
`verification-claim-integrity.md` §2 asks for.

### 3. Baseline-attribution

Every command above was run in this audit, against this worktree
(`.claude/worktrees/t119`, branch `WT-todo-auto-analysis`). Source-tree commands
ran against HEAD `f1bc39310`; **no tracked source file is modified by this
audit** — `git status --porcelain` shows the three SPEC artifacts (modified by
the SPEC author, not by this auditor) and this untracked report. Artifact claims
are attributed to the current working-tree state of all four artifacts as of
report time. The budget figure is a re-derivation of `alwaysLoadedSurface()`'s
enumeration (no-`paths:` rules under `.claude/rules/moai` + `CLAUDE.md` +
`output-styles/moai/moai.md` + `MEMORY.md` head), not a number carried from any
prior session; the guard test's own PASS is quoted beside it. No REQ or AC was
sampled — all 15 REQs and all 16 ACs were read in full, and all were re-read
after each of the two revisions landed mid-audit.

### 4. Gaps — what was NOT observed

- The full Go test suite was not run (prohibited by the dispatch and by
  `CLAUDE.local.md §4`). No claim here rests on suite-wide behaviour.
- `moai todo` was not executed against a live or fixture queue. Every claim
  about current CLI behaviour is read from source, not from a run.
- **The revised artifacts are uncommitted.** This verdict attaches to a working
  tree, not to a commit. If the revisions are altered before they land, the
  verdict does not carry over.
- The budget figure is a re-derivation of the guard's enumeration, not the
  guard's own internal total — the guard reports pass/fail, not the number. It
  is a ±15% char/4 estimate by the guard's own design
  (`token_budget_guard.go:40-45`), so 4,788 is an approximation, not a ledger.
- The concurrent auditor's D-numbered finding set was not read. Its findings are
  known only from the `spec.md` HISTORY row, so this report cannot state whether
  its D-numbers fully overlap these F-numbers, nor whether all of its findings
  were addressed. Overlap is evident (D6/D11 correspond to F6/F7/F8; the
  317/82 figure in `acceptance.md:139` matches this audit's measurement) but the
  correspondence is inferred, not verified.
- `acceptance.md:157`'s "~336s for `internal/cli`" was not re-measured; it was
  accepted, consistent with the standing 600s-floor guidance.
- `SPEC-KANBAN-TODO-CLI-001`'s own artifacts were not audited beyond the three
  lines quoted (REQ-TODO-013, REQ-TODO-014, `spec.md:92`) and its `status:`.
- Whether the t170 lane (`plan.md:76`) currently holds edits to
  `workflows/todo.md` was not checked; the conflict risk is stated but
  unmeasured.

### 5. Residual risk

- **Concurrent revision is the largest risk to this verdict's validity.** The
  artifacts changed twice during a single audit. A third change invalidates the
  line citations throughout this report, and possibly the verdict.
- **AC-TA-003 sits 0.033 above the Jaccard threshold** (F15). The fixture scores
  J = 0.8333 against a 0.80 cutoff. Any normalization step added during
  implementation that the SPEC does not currently command — stopword removal
  most obviously, which would take the pair to 1.0 and reclassify it `exact` —
  moves this fixture across a classification boundary. `progress.md:51` records
  it as a run-phase gap; it is carried, not closed.
- **The AC budget is exhausted.** 16 of 16 at Tier M. Two of the open minor
  findings (F9 NFC, F10 grep token) are best fixed by *strengthening existing
  criteria* rather than adding new ones, because there is no room for a
  seventeenth without a tier decision — F16.
- **The `unreviewed` mark remains weaker than §B.3 argues.** The revision
  sharpened the pair comparison to unordered and AC-TA-011 now validates it by
  relating in the *reverse* direction — a genuinely good test-design choice. It
  does not address who may write the agent finding that clears the mark — F5.
- **The doctrine amendment lands in files two other lanes may touch**
  (`plan.md:76`, `plan.md:133`). The additive-only shape genuinely reduces the
  conflict surface, but the mirror-pair obligation means a partial landing
  reverts on the next `moai update`. AC-TA-014 catches this only if run after
  both halves land.
- **The 0.80 Jaccard threshold is declared provisional** (`plan.md:47`) with a
  20%-of-cards noise trigger, but no AC or DoD item requires that measurement to
  be taken. The tuning step can silently be skipped.
- **The declared residual risk on the refusal path is real and accepted, not
  solved.** A script that discards both the exit code and stderr loses the card
  silently. The SPEC says so plainly in two places (`spec.md:86`, `plan.md:53`)
  and declines to build a compensating mechanism. That is a legitimate call —
  the CLI cannot force a caller to read its exit code — but it is a live risk
  carried into run-phase, not an eliminated one.

---

## Findings

Status reflects the current working-tree state of all four artifacts.

| # | Severity | Class | Artifact | Status |
|---|---|---|---|---|
| F1 | critical | blocking | acceptance.md | **RESOLVED** |
| F2 | major | blocking | acceptance.md | **RESOLVED** |
| F3 | major | blocking | acceptance.md | **RESOLVED** |
| F4 | minor | optional | plan.md | open (downgraded — ruling recorded, framing not corrected) |
| F5 | major | optional | spec.md | open |
| F6 | major | optional | spec.md | **RESOLVED** |
| F7 | minor | optional | spec.md | **RESOLVED** |
| F8 | minor | optional | spec.md | **RESOLVED** |
| F9 | minor | optional | acceptance.md | open |
| F10 | minor | optional | acceptance.md | open |
| F11 | minor | optional | spec.md | open (now cosmetic) |
| F12 | minor | optional | spec.md | open |
| F13 | minor | optional | plan.md + acceptance.md | open |
| F14 | minor | optional | spec.md | open (materially narrowed) |
| F15 | minor | optional | progress.md | **RESOLVED** |
| F16 | minor | optional | acceptance.md | open (new — advisory) |

No blocking finding is open. Per M6, the open findings are optional-class: they
are surfaced for the orchestrator's discretion and do not gate the verdict.

### F1 — CRITICAL, blocking — **RESOLVED**

At HEAD, AC-TA-013 required `grep -rn "AskUserQuestion" internal/cli/` to return
0. Measured: **317 hits across 82 files** (C5). The criterion was unsatisfiable
on any implementation, which made the Definition of Done unreachable by
construction.

The revision discards it and substitutes a scope-extension of the existing
`todoPromptGuard` over `internal/cli/todo*.go` non-test sources, retaining the
synthetic negative control at `todo_test.go:461`. The discarded criterion is
documented in place with its measured baseline (`acceptance.md:139`, "실측
317건 / 82파일 (`f1bc39310`)") — matching this audit's independent measurement
exactly, so the AC now carries its own evidence against re-widening. This is the
strongest single fix in the revision.

### F2 — MAJOR, blocking — **RESOLVED**

At HEAD, REQ-TA-002's `analyze` clause had no acceptance criterion; v0.2.0 then
added a third obligation (tuple-level idempotence) to the same uncovered
requirement, widening the gap.

The revision adds **AC-TA-016**, which covers all three obligations in one
fixture: a positive control (findings length > 0 after the first run, so an
implementation that records nothing cannot pass on "length unchanged" alone),
exact length equality after the second run, a zero-duplicate-tuple assertion,
and `items` byte-identity across both runs. The positive-control pairing is
exactly the discipline `acceptance.md §A` sets out for itself.

### F3 — MAJOR, blocking — **RESOLVED**

At HEAD, `moai todo unrelate` was commanded by REQ-TA-008 with no criterion —
leaving the reversibility half of the SPEC's own §B.1 criterion unverified for
agent findings.

The revision rewrites AC-TA-004 as a bidirectional relate→unrelate test in one
fixture, seeded with a pre-existing unrelated finding. It asserts exact findings
length in both directions (2 then 1), that the survivor is the *unrelated*
pair — catching an implementation that deletes the wrong finding — and `items`
byte-identity at both checkpoints. The over- and under-deletion failure modes
are both caught.

### F4 — MINOR (downgraded from MAJOR), optional — `plan.md:20`

`plan.md:22` now records the Tier ruling explicitly: "plan-audit iteration 1
판정: Tier M 유지", with three stated grounds — the amendment targets a workflow
rule plus a skill rather than `moai-constitution.md`; it is additive-only with
AC-TA-014 asserting the existing [HARD] sentences survive; and both the file
count and LOC sit inside the M band. **This auditor concurs with Tier M** on
those grounds.

What remains is narrower than the original finding. `plan.md:20` still asserts
the amendment "금지를 **완화하는 것이 아니라** 무엇이 허용되는지 이름 붙여 좁히는
것" (does not relax; narrows by naming). With respect to `edit`/`move`/`drop`/
`undrop` that is true. With respect to `add` it is not: before the amendment,
`todo.md:50` says without qualification that the queue "records the operator's
intent; it does not curate it", and no automatic queue behaviour is permitted;
after it, an automatic refusal is. Preserving the old sentences verbatim makes
the two coexist, with the newer carving an exception out of the older — which is
a bounded relaxation, and defensible as one.

The Tier conclusion does not depend on this: the ruling at `plan.md:22` rests on
scope, method, and size, none of which the correction touches.

**Required fix (one sentence)**: amend `plan.md:20` to "with respect to
`edit`/`move`/`drop`/`undrop` the amendment narrows; with respect to `add` it
grants a new, bounded permission that did not previously exist" — the Tier M
ruling at `plan.md:22` stands unchanged either way.

### F5 — MAJOR, optional — `spec.md:101`, `spec.md:135`, `spec.md:143`

`spec.md §B.3` assigns the L2 semantic layer to `manager-lead`, and REQ-TA-013's
`unreviewed` mark is built on that assignment. But REQ-TA-008 places no actor
constraint on `relate` — any agent, any lane, any script may write a
`source: agent` finding, and REQ-TA-013 clears the mark on the mere existence of
one for the pair (confirmed by AC-TA-011 ②). So the mark that `spec.md:103`
calls "L2 부재를 다루는 유일한 정직한 방법" can be switched off by any writer,
reviewed or not.

The revision narrowed a neighbouring hole well — pairs are now compared
unordered, and AC-TA-011 validates it by relating in the reverse direction — but
left the actor question untouched.

**Required fix**: either add a requirement constraining who may write
`source: agent` findings (with an AC that a non-lead writer is refused or
recorded distinctly), or amend `spec.md §B.3` to state plainly that `unreviewed`
tracks *the presence of an agent-sourced record*, not *review*, and rename the
marker accordingly. The text currently claims the stronger property while the
requirements deliver the weaker one.

### F6 — MAJOR, optional — **RESOLVED**

At HEAD, REQ-TA-015 claimed to preserve REQ-TODO-014 while restating its contract
as "exit 0/1"; the source says **exit 0/1/2**. Both REQ-TA-015 (`spec.md:148`)
and AC-TA-013's Then-clause now say 0/1/2 (C9).

### F7 — MINOR, optional — **RESOLVED**

`internal/cli/todo_edit_move.go:17` → `:5-7`. The quoted comment sits at line 5;
the citation now resolves (C3).

### F8 — MINOR, optional — **RESOLVED**

`internal/cli/todo_edit_move.go:110` → `:99`. The quoted sentence sits at line
99; the citation now resolves (C3).

### F9 — MINOR, optional — `spec.md:122`, `acceptance.md:26-27`

REQ-TA-001 commands a four-step normalization: Unicode NFC → trim → collapse
internal whitespace → case-fold. AC-TA-001's input `"  fix   the FLAKY gate "`
exercises trim, collapse, and case-fold. **No AC exercises NFC.** A Korean or
accented-Latin card text differing only by composition form (NFD vs NFC) would
pass every criterion while the analyser silently missed the duplicate — and this
is a Korean-language project where composed vs decomposed Hangul is a live
hazard on macOS filesystems, which is precisely where this queue runs.

**Required fix**: extend AC-TA-001's existing fixture with a second pair whose
texts differ only by NFC/NFD composition, asserting the same refusal. Extending
AC-TA-001 rather than adding AC-TA-017 keeps the count at the Tier M ceiling
(F16).

### F10 — MINOR, optional — `acceptance.md:146`

AC-TA-014 requires the amended section to contain phrases corresponding to
"records" and "never folds". Measured baseline: "records" already returns 5 hits
in `todo.md` and 1 in `kanban-dispatch.md`; "never folds" returns 0 in both
(C12). The "records" half observes nothing — it would pass on the untouched
tree. The AC as a whole still discriminates because the "never folds" half has a
clean baseline, so this is a weakening rather than a break.

**Required fix**: replace the "records" token with a phrase whose baseline is 0
(a distinctive fragment of the M1 boundary paragraph at `plan.md:68-70`), and
record both measured baselines in the AC — the same treatment AC-TA-013 now
gives its discarded criterion.

### F11 — MINOR, optional — `spec.md:123`, `spec.md:135`, `spec.md:142`

REQ-TA-002, REQ-TA-008, and REQ-TA-012 each bundle multiple independent
`when`-clauses under one REQ id — REQ-TA-002 now carries three. GEARS form is
satisfied per clause; atomicity is not. This non-atomicity was the mechanical
cause of the original F2 and F3: requirement halves without ids of their own are
easy to leave uncovered.

**Now cosmetic.** The revision closed the consequence by other means — every AC
carries a `**검증**: REQ-…` line and all 15 REQs are cited (C7), so the coverage
hole this defect produces is now visible in the document itself. Splitting the
REQs would raise the count toward the Tier M ceiling of 16 for no verification
gain.

**Required fix**: none required. If the SPEC is ever tiered up, split them then.

### F12 — MINOR, optional — `spec.md:125`

REQ-TA-004 is labelled `(Capability-gate)`. GEARS reframes `Where` as capability
gate / feature flag / static config; a per-invocation CLI flag is an invocation
event, not a static capability. Event-driven is the fitting pattern. Cosmetic —
the sentence parses correctly either way and MP-2 passes regardless.

**Required fix**: relabel as `(Event-driven)` and rephrase as "**When**
`moai todo add --force` runs and the analysis classifies the candidate as
`exact`, the command **shall** …".

### F13 — MINOR, optional — `plan.md §G`, `acceptance.md §D`

M1 adds prose to `.claude/rules/moai/workflow/kanban-dispatch.md`, which declares
itself always-loaded (`kanban-dispatch.md:5`) and is therefore measured by the
`AlwaysLoadedTokenBudget = 76000` guard
(`internal/config/token_budget_guard.go:32`). Neither `plan.md §G` (risks) nor
`acceptance.md §D` (quality gates) mentions the guard.

Measured: current surface ≈ 71,212 tokens, headroom ≈ 4,788; the M1 addition is
~175 tokens. **The change fits comfortably — this is not a capacity problem.** It
is a coordination one: `kanban-dispatch.md` is the largest always-loaded rule at
25,915 bytes and is the standing first target for the stub + lazy-companion diet,
so a card adding to it should say so.

**Required fix**: add one row to `plan.md §G` naming the guard and the measured
headroom, and one line to `acceptance.md §D` requiring
`go test ./internal/config/ -run TokenBudget` to pass after M1.

### F14 — MINOR, optional — `spec.md:77-86`

The §B decision table evaluates four actions against reversibility +
conspicuousness. The revision materially improved the refusal row: it is now
"조건부 자동 허용", its conspicuousness is conditioned on the caller reading exit
code or stderr, and `spec.md:86` declares an explicit residual risk for the
script/pipe path that discards both — mirrored at `plan.md:53`. That is an honest
correction and it closes most of what this finding originally raised.

What remains: §B still never evaluates the refusal against the doctrine sentence
most directly in tension with it — "The queue records the operator's intent; it
does not curate it" (`todo.md:50`). A refusal is precisely the queue declining to
record a stated intent on its own equivalence judgment.

This is an argument gap, not an unscoped change: REQ-TA-014's mandated amendment
text names the refusal explicitly ("it refuses the admission of an exact
duplicate"), so the change is openly scoped — which is what keeps this SPEC clear
of a doctrine-conflict blocking finding.

**Required fix**: one sentence in §B answering it directly — the refusal records
no card and curates none; it declines an admission that would be
indistinguishable from one already recorded, leaves the file byte-identical, and
surfaces immediately, with `--force` preserving the operator's override.

### F15 — MINOR, optional — **RESOLVED**

At the time this finding was raised, `progress.md` had not been revised alongside
the other three artifacts: it recorded "REQ 15 / AC 15" (the count was already
16) and "상태: plan-audit 대기" after an audit had completed. Since `§E.1` is the
plan-phase readiness signal the run-gate reads, a stale count there is a small
but real integrity defect in the artifact whose job is to state readiness.

It has since been revised and now records "REQ 15 / AC 16 (Tier M 상한 각 16)"
(`progress.md:6`), the Tier M ruling with its grounds (`progress.md:5`), a
per-defect resolution table, and a remeasured-figures table.

**The remeasured figures were independently re-verified by this audit and all
match** (C15). Two are worth calling out because they *correct* this report
rather than merely confirm it:

- `todo_edit_move.go:5-7` is more precise than the `:5` this audit first cited —
  the sentence spans three comment lines:
  ```
  $ sed -n '5,7p' internal/cli/todo_edit_move.go
  // a correction the operator decided on. Nothing here infers what a card
  // should say or where it belongs — no analysis, no absorption, no silent
  // promotion. Those would collide head-on with the [HARD] clauses in
  ```
- The AC-TA-003 Jaccard boundary (`progress.md:51`) is a gap this audit did not
  identify. Verified independently:
  ```
  A = {rework, the, auth, middleware, error, paths}   (6)
  B = {rework, auth, middleware, error, paths}        (5)
  |A∩B| = 5, |A∪B| = 6  →  J = 0.8333, margin over 0.80 = 0.0333
  ```
  The fixture clears the threshold by 0.033. Adding a normalization step the SPEC
  does not currently command — stopword removal being the obvious candidate,
  since dropping "the" would take the pair to J = 1.0 and reclassify it as
  `exact` — moves this fixture across a classification boundary. `progress.md`
  flags it as a run-phase gap rather than silently carrying it. That is the
  correct disposition and it is recorded here so it is not lost.

### F16 — MINOR, optional — `acceptance.md` (new, advisory)

The AC count is now **16 of the Tier M ceiling of 16**
(`spec-workflow.md` § SPEC Complexity Tier). The budget is exhausted: no
seventeenth criterion can be added without either tiering up or removing one.

This is not a defect — 16 ≤ 16 passes — but it constrains how the remaining
findings should be fixed. F9 (NFC) and F10 (grep token) are both best addressed
by **strengthening existing criteria** (extend AC-TA-001's fixture; swap
AC-TA-014's token) rather than by adding new ones.

**Required fix**: none. Recorded so the run-phase fix route does not silently
push the SPEC over budget.

---

## Recommendation

**PASS at 0.920**, above the Tier M threshold of 0.80. Proceed to Implementation
Kickoff Approval.

This PASS is grounded in the following must-pass evidence:

- **MP-1**: `REQ-TA-001`…`REQ-TA-015` at `spec.md:122-148` — sequential, no gaps,
  no duplicates, uniform padding.
- **MP-2**: all 15 REQs matched to a named GEARS pattern, requirement layer only;
  the Given-When-Then ACs graded separately under Group 4.
- **MP-3**: all 12 canonical frontmatter fields verified field-by-field at
  `spec.md:2-14`, including `phase: "v3.1.4 target"` as a release label rather
  than a prohibited lifecycle-stage name.
- **MP-5**: `SPEC-KANBAN-TODO-CLI-001` resolves with `status: in-progress`.
- **MP-7**: zero `[NEEDS CLARIFICATION]` markers.
- **MP-4 / MP-6**: N/A, reasons stated.

Four findings that were blocking at HEAD `f1bc39310` (F1, F2, F3, and the
original form of F4) are resolved in the working tree, each verified against the
tree rather than accepted on the revision's own account: the 317/82 measurement
behind F1's fix reproduces exactly, `todoPromptGuard` and its negative control
exist where AC-TA-013 says they do, and all 15 REQs now resolve to at least one
AC.

One caveat bounds this verdict: **it attaches to an uncommitted working tree,
not to a commit.** All four artifacts are modified and unstaged. If any changes
again before landing, the verdict does not carry over and the line citations
throughout this report stop resolving.

Eight optional findings remain (F4, F5, F9, F10, F11, F12, F13, F14, plus the
F16 advisory). None blocks. In judgment order:

- **F5** is the one worth closing before run-phase. The SPEC claims a stronger
  property for `unreviewed` than its requirements deliver, and that gap is far
  cheaper to close in prose now than to discover in implementation.
- **F9** is the highest-value verification addition, and it fits inside
  AC-TA-001's existing fixture without touching the exhausted AC budget (F16).
  It also interacts with the Jaccard boundary noted in F15: NFC/NFD handling and
  stopword handling are both normalization decisions this SPEC leaves partly
  open, and both move classification outcomes.
- **F4, F10, F12, F13, F14** are one-line or one-sentence corrections. The rest
  is the orchestrator's call.

A closing note on process rather than content. This SPEC was revised three times
during a single audit, and each revision re-measured the auditor's figures
instead of accepting them — correcting the auditor's line citation in one case
and surfacing a threshold-margin gap the auditor missed in another. The
resulting artifacts are stronger than either the original SPEC or this report's
first draft. That is what the plan-audit loop is for, and it worked here.
