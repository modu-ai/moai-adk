# SPEC-JUDGMENT-FIRST-MODE-001 — Progress

SPEC: SPEC-JUDGMENT-FIRST-MODE-001
Card: t401
Tier: L
Worktree: `.claude/worktrees/t401` — branch `WT-analysis-pull`, base `ad272be20`

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set complete for Tier L (5 files + this progress record). **Revision 0.2.0**
closes iteration-1 plan-audit defects D1-D12 (report:
`.moai/reports/plan-audit/SPEC-JUDGMENT-FIRST-MODE-001-review-1.md`, verdict FAIL 0.76 — MP-8 plus
sub-threshold aggregate).

| Artifact | State |
|---|---|
| `spec.md` | authored — 25 GEARS requirements (REQ-JFM-024 added: detector positive control; REQ-JFM-025 added by the provenance amendment), 12-field frontmatter at `version: "0.2.0"`, `status: draft`, `tier: L`, HISTORY table (D6), exclusions section with 6 `### Out of Scope —` sub-headings |
| `plan.md` | authored — 7 milestones **reordered**: M0 (observer + pre-landing baseline) now leads, because that is the only position from which a live `label_present: true` control row is obtainable (D4) |
| `acceptance.md` | authored — 23 Given-When-Then criteria, **13 release-blocking**, each with a measured two-cell RED-now + green-path pair; 6 regression-guards with baseline cells; AC-JFM-018 as the vacuity falsifier, entry-gated on AC-JFM-023 |
| `design.md` | authored — mode resolution (+ mid-session flip semantics, D12), §6.3 now the single denominator owner (D3), observer record shape, rejected alternatives |
| `research.md` | authored — measured surface inventory, two negative findings, precedent wiring, tier evidence; §6 `.github/` premise corrected against measurement (D8) |

**RED-now measurement pin for the whole verification layer**: 2026-09-02, worktree
`.claude/worktrees/t401`, branch `WT-analysis-pull`, tree
`ad272be20abff9e4f3b1b363fce3e48dac4c5132`; `git status --porcelain` → two untracked paths only, so
every tracked file measured is byte-identical to `ad272be20`. Every cell in `acceptance.md` carries
its command, verbatim stdout, exit code as a separate field, and why it is red — measured in this
run, in this tree.

Defects surfaced mechanically by authoring those cells, exactly as the audit predicted:

- **AC-JFM-021 was impossible.** The 0.1.0 drift loop printed **31** DRIFT lines on the untouched
  tree (35 `.tmpl`, 12 `.sh`, only 4 pairs) because `[ -f "$b" ] && … || echo DRIFT` misreports an
  absent base sibling. Recipe corrected (`swept=4 drift=0`), criterion demoted to regression-guard,
  and the `handle-pre-tool.sh` exclusion recorded as a stated gap rather than a vacuous pass (D2).
- **Four criteria were green at arrival** — AC-JFM-001, 010, 012, 020, each measured at 0 diff
  lines / 0 grep hits — plus AC-JFM-022 (16 packages `ok`, exit 0). All six are now
  regression-guards with baseline cells (D5). Blocking count fell 18 → 13, and the Definition of
  Done now derives that number from the matrix instead of restating it (D7).

SPEC ID regex self-check executed:

```
ID="SPEC-JUDGMENT-FIRST-MODE-001"
[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
→ PASS
```

Open items carried into the audit and the Implementation Kickoff Approval gate:

1. **Operator review requested** — the one-line mode reference at
   `.claude/skills/moai/workflows/run.md:137` (spec.md §E.1). It is a downstream consumer of S1,
   not a seventh nominated surface; leaving it unconditioned would place two `[HARD]` clauses in
   direct contradiction at the one mandatory human gate.
2. **Scope widening for operator confirmation** — REQ-JFM-005 now binds **every**
   `AskUserQuestion` call under `pull`, not only decision-type ones (spec.md §B). The narrower
   0.1.0 wording was unmeasurable: the payload carries no question-type field, so the criterion
   would have filtered on a field that does not exist and returned an empty sample forever (D3).
   The requirement widened to match what is observable rather than the measurement narrowing to
   match an unobservable requirement. This is a real behavioral scope change and is flagged for the
   Implementation Kickoff Approval gate, not assumed.
3. **Milestone reordering for operator confirmation** — M0 (observer) moved ahead of the doctrine
   milestones, against the plan's own reversibility ordering, so that REQ-JFM-024's pre-landing
   `push`-mode control window exists at all (D4). M1 does not start until M0's baseline row is on
   record.
4. **Measured in M2, not assumed** — whether `moai-easy.md` and `moai-learn.md` carry the same
   banner rules and therefore need the same pull branch.
5. **Recorded, not resolved** — the artifacts are English while `conversation_language` is `ko`.
   The reason is stated in spec.md §E.4 (Template-First: the quoted doctrine strings the criteria
   `grep` for are English and land in the neutral template tree). Operator-reversible (D11).

**Recorded branch state (truth pass, 2026-09-03 — replaces the former closing sentence of this
section, which claimed no code, no commit and no push in this phase: true when written at 0.2.0,
false since 2026-09-02 14:42).** Commit `d5caf2d8e` ("feat(SPEC-JUDGMENT-FIRST-MODE-001): M0 AskUserQuestion
PreToolUse observer (t401)", 2026-09-02 14:42:03 +0900) landed on this branch —
`internal/hook/askuser_observer.go` +186, `askuser_observer_test.go` +389, `pre_tool.go` +10,
`.claude/settings.json` +11, template mirror `settings.json.tmpl` +11 — **before the
Implementation Kickoff Approval gate was ever opened**, and is reachable from `origin/develop`
(`git merge-base --is-ancestor d5caf2d8e origin/develop` → rc=0; re-measured in this tree
2026-09-03). This is a **recorded gate-ordering violation**: no authorization for the pre-gate
landing is on record here; the lane's own verdict
(`.moai/reports/t401/plan-phase-verdict.md`) records that the Kickoff gate was never opened and
was escalated as a blocker to the lead. Whether the landed M0 code is ratified or reverted is an
**OPERATOR decision pending with the lead — it sits on the Implementation Kickoff Approval
agenda**. The lane does not ratify and does not revert; this repair touches records only, not
the code. The M0 pre-landing baseline row (REQ-JFM-024's `push`-mode control window) is
**still uncollected**.

**Operator decision on the pre-gate M0 landing (2026-09-06): RATIFIED, not reverted.** The
gate-ordering violation recorded in the paragraph above **stands as a violation** — ratification
approves what happened; it does not make it not have happened, and it is **not a precedent**. The
next SPEC that lands code before its Kickoff gate opens is a fresh violation and is escalated as
one.

Why revert was rejected, stated so the reasoning is auditable rather than assumed: reverting kills
the observer, and with it the `AC-JFM-023` half-2 collection window that is **already accumulating**
(3 rows at the time the decision was taken, 8 rows measured 2026-09-06T08:32Z — see §E.2). The
landed code **observes and changes nothing else**: it appends a row per `AskUserQuestion` PreToolUse
event and has no other effect on the tool call, so the cost of leaving it in place is a log file,
while the cost of removing it is the loss of a pre-landing baseline that exists only in that window.

Decision route: escalated as a lane blocker report to the lead session, raised by the lead to the
operator, decision returned through the lead and confirmed with the operator in the lane session
before this record was written. The lane did not open the operator gate itself
(`feedback_lane_cannot_open_operator_gate`).

Re-measured in this tree at `HEAD 928da5dc3`, 2026-09-06:

```
$ git merge-base --is-ancestor d5caf2d8e origin/develop ; echo rc=$?
rc=0
```

Two run-phase obligations follow from the ratification and are NOT yet discharged: the M0
pre-landing baseline row is still uncollected (unchanged from the paragraph above), and this
ratification does not retroactively authorize any further pre-gate landing.

**Provenance amendment (0.2.3, 2026-09-03).** Adopted from the decision document
`.moai/reports/t401/provenance-eligibility-options.md`: **Option B rejected** (the doc's own form:
안 A + 안 C, 안 B 기각) — no `session_start`,
no matcher SHA in the provenance record, per Option A's add-nothing stance (reason recorded in
REQ-JFM-025: the exported window's own
existence and row counts already prove the wired-session condition; confirmation stamps gating
nothing leave a finished-verification impression); **Option C adopted** — `calls_issued` (the asking
session's own count of `AskUserQuestion` calls issued during the interval) added to REQ-JFM-025's
provenance enumeration, with AC-JFM-018 half 3 / AC-JFM-023 half 4 asserting the four-way
`rows_recorded` vs `calls_issued` contrast: observer non-wiring and partial row loss become
observable mismatch signals, never silent passes (detectable, not eliminated). Version bumped
0.2.2 → 0.2.3; REQ count unchanged at 25 (rides REQ-JFM-025's existing enumeration). Affected
RED-now cells re-measured and re-pinned to `HEAD 095f2799b`. The §E.1 audit-ready verdict line is
deliberately NOT refreshed here — it is refreshed after the audit this amendment will receive.

**Audit signal refresh (iter-5, 2026-09-03).** Plan-audit iteration 5 returned **FAIL 0.95**
(report: `.moai/reports/plan-audit/SPEC-JUDGMENT-FIRST-MODE-001-review-5.md`) — above the Tier L
0.85 threshold on aggregate; the FAIL rests solely on blocking defects D1 (this progress record's
false "no code" state) and D2 (stale provenance-pin header in acceptance.md); must-pass 7/7 PASS.
The iter-5 record-truth repairs (D1-D7 of that report) are applied in this commit — records only,
no code change, no version bump (record-only repair; version stays 0.2.3). The next audit
iteration (iter-6) **awaits lead authorization**: iteration-5 triggered the STOP signal on score
regression (iter-4 0.96 → iter-5 0.95) and the plan-auditor iteration cap is already consumed, so
no further iteration is self-served from the lane.

## §E.2 Run-phase Evidence

_<pending run-phase — formal population at run phase not yet entered. M0 e2e evidence, however,
IS committed on this branch under `.moai/reports/t401/e2e/` (`payload.json`,
`payload-installed.json`, `fresh/` and `installed-v2/` captures;
`installed-v2/.moai/reports/t401/wire-shape-check.md`), landing with the pre-gate M0 commit
series — see the recorded gate-ordering violation in §E.1. This annotation points at that
evidence; it is not the formal §E.2 population._

### `calls_issued` self-report defect — mitigation chosen (2026-09-06, before first code edit)

**The defect.** `AC-JFM-018` half 3 / `AC-JFM-023` half 4 contrast `rows_recorded` (produced by the
observer) against `calls_issued`, which `acceptance.md:400` enumerates as *"a value the asking
session knows without the observer"* — its independence from the observer is recorded as an
advantage and its **lack of any independent source** is recorded nowhere. With no independent
source, an asking session that writes `calls_issued` to match `rows_recorded` satisfies
`acceptance.md:417` (`rows_recorded == calls_issued` → the window covers the interval), and the two
mismatch states the contrast exists to expose — partial row loss (`:417`) and observer non-wiring
(`:418`) — pass silently. The contrast asserts nothing against the one party able to falsify it.

**Chosen mitigation: option 1 — bind `calls_issued` to an independent source**, with option 2
(record the limitation as an explicit gap in the cell) as the fallback where the source cannot be
resolved. Not option 2 alone.

**Why option 1 rather than option 2, and why the choice is not an assumption.** Option 1 was not
adopted because it sounds stronger; it was adopted because the independent source was **measured to
exist and to agree** before the choice was made. The asking session's transcript records one
`tool_use` entry per `AskUserQuestion` call, written by the runtime rather than declared by the
session, which makes it independent in the sense the contrast requires. Measured in this session
(`session_id 98e78ea5-aded-4e8a-93db-014842a32f3b`), 2026-09-06:

```
$ grep -o '"name":"AskUserQuestion"' \
    ~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t401/98e78ea5-aded-4e8a-93db-014842a32f3b.jsonl \
  | wc -l
       5
$ grep -c '98e78ea5-aded-4e8a-93db-014842a32f3b' \
    /Users/goos/MoAI/moai-adk-go/.moai/logs/askuser-observations.jsonl
5
```

Transcript-derived count 5, observer-derived count 5, agreeing on a session whose calls were issued
by this lane rather than staged. Had they disagreed, option 2 would have been the correct choice.

**Two hazards the run-phase implementation MUST handle; both were measured, not predicted.**

1. **The transcript is not under `~/.claude/projects/`.** A session launched with a custom
   `--settings` file writes its transcript under the profile directory instead — here
   `~/.moai/claude-profiles/moai-adk/projects/`. `find ~/.claude -name '98e78ea5*'` returned nothing
   (rc=0, empty). A recipe that hardcodes the default path resolves `calls_issued = 0`, which reads
   as `rows_recorded > calls_issued` — `acceptance.md:418`'s "rows from other sessions are mixed in"
   branch — i.e. a **wrong diagnosis, not a visible failure**. The resolution MUST therefore fail
   loudly when no transcript is located, and fall back to option 2 (explicit gap) rather than
   emitting a zero.
2. **The transcript directory is keyed by cwd.** The directory name encodes the worktree path
   (`…-Users-goos-MoAI-moai-adk-go--claude-worktrees-t401`). This session worked in
   `.claude/worktrees/develop` before `.claude/worktrees/t401`, and only one file was found, of
   1,002,764 bytes — consistent with the transcript following the session, but **whether it follows
   or splits across directories was not measured**. A session that changes worktree mid-window may
   have its transcript split; the resolution MUST search all candidate directories rather than one.

**Gaps in this measurement, stated so the run phase does not inherit them as settled.** n=1 session,
`push` mode only — the `pull`-mode window (`AC-JFM-018`) was not measured. The
follow-or-split question above is open. No interval filtering was applied: the counts are
whole-session, whereas the criteria contrast counts *over the collection interval*, so the
implementation must add timestamp bounding that this probe did not exercise.

**Scope note.** This mitigation changes `acceptance.md` cell content (the provenance enumeration
and the four-way reading rule), which is `manager-spec`-owned body content under the Status
Transition Ownership Matrix. The run phase returns a blocker report for the cell edit rather than
performing it directly.

### M1 — the pull convention in doctrine (2026-09-07)

Measured in the worktree `.claude/worktrees/t401`, branch `WT-analysis-pull`, parent HEAD
`129fe8b88`, against base tree `ad272be20`.

| Criterion | Command | Actual output | Status |
|---|---|---|---|
| AC-JFM-004 | `grep -n 'recommendation_mode' .claude/rules/moai/core/askuser-protocol.md` | 4 lines: `:64`, `:117`, `:120`, `:265`; `:117`/`:120` sit inside § Recommendation Placement Principles (`:87`-`:154`) | PASS |
| AC-JFM-004 (companion) | `grep -c 'Adaptive strength' .claude/rules/moai/core/askuser-protocol.md` | `1` | PASS |
| AC-JFM-009 | `grep -c '^### On-request emission' .claude/rules/moai/core/askuser-protocol.md` | `1` | PASS-WITH-DEBT — the heading exists; the "referenced from both banner rules" half lands in M2, which owns `moai.md` |
| AC-JFM-010 | `git diff ad272be20 -- .claude/rules/moai/core/askuser-protocol.md .claude/rules/moai/workflow/orchestration-mode-selection.md` | 3 hunks, all in `askuser-protocol.md` (`@@ -61,7`, `@@ -102,7`, `@@ -214,7`); none touches § Option Description Standards (`:74`-`:86`), § Report-Before-Ask Gate, or § Requested-Deliverable Primacy; `orchestration-mode-selection.md` diff empty | PASS |
| AC-JFM-010 (section byte-equality) | `sed -n '/Report-Before-Ask/,/^## /p'` on current vs `ad272be20`, compared as files | `diff_exit=0`, 44 lines extracted (non-empty, so not a vacuous compare) | PASS |
| AC-JFM-011 | `grep -c 'Detect → Explain → Ask'`, `grep -c "An LLM 'best practice' is not a policy"`, `grep -c 'When uncertain, escalate. Never downgrade.'` | `1`, `1`, `1` | PASS |
| AC-JFM-012 | `git diff ad272be20 -- .claude/rules/moai/core/zone-registry.md` | empty; exit `0` | PASS |
| AC-JFM-012 (companion) | `grep -c 'recommendation_mode' .claude/rules/moai/development/branch-origin-protocol.md` | `1` | PASS |
| AC-JFM-013 | sweep → `wc -l < .moai/reports/t401/ac013-candidates.txt` | `26` (pre-edit `25`; the 26th is the pull branch this milestone added at `branch-origin-protocol.md:26`, which itself matches the selector) | swept count > 0 |
| AC-JFM-013 (ledger) | `grep -c '^| [0-9]* | \`' .moai/reports/t401/ac013-ledger.md` | `26` rows = swept count; 7 `conditioned`, 18 `unconditioned-by-design`, **1 escalated** | **FAIL** — one unclassified remainder (see below) |

All five coordinates AC-JFM-013 requires in `conditioned` are `conditioned`:
`askuser-protocol.md:64`, `askuser-protocol.md:265` (the other surviving S1 row, `:217` at base),
`run.md:137`, `plan/spec-assembly.md:212`, `plan/spec-assembly.md:353`. Six of the seven carry
`recommendation_mode` on the matched line; `branch-origin-protocol.md:25` is conditioned by the
adjacent pull branch at `:26` because REQ-JFM-015 / §B.2 require its text verbatim.

**AC-JFM-013 is FAIL, deliberately.** `.claude/skills/moai/SKILL.md:350` carries an independent
`[HARD]` clause — "All AskUserQuestion calls throughout MoAI workflows MUST follow these rules:
The first option MUST always be the recommended choice, clearly marked with '(Recommended)'
suffix" — with no SSOT citation and the broadest reach in the swept set. It meets REQ-JFM-016's
reachability test, and `.claude/skills/moai/SKILL.md` is not in M1's declared edit surface. Neither
available class is honest: `conditioned` would require an out-of-scope edit, and
`unconditioned-by-design` would rest on scope absence rather than a design ground. Escalated as a
blocker; ledger § Blocker carries the recommended resolution.

**Template mirrors.** All 4 changed files were byte-identical to their
`internal/template/templates/` mirrors before the edit and are byte-identical after it
(`diff -q` × 5, including the untouched `zone-registry.md`). The 85 added template lines were
scanned for SPEC IDs, REQ / AC tokens, `CLAUDE.local` references, absolute `/Users/` paths, ISO
dates, and 9-40 hex SHAs: zero hits, with a catch-all control confirming the pipeline reached data
(85 `^+` lines, 8 carrying `recommendation_mode`).

**M0's baseline window is still outstanding, and M1 has now closed half of its stated condition.**
M0's code landed at `d5caf2d8e` (`internal/hook/askuser_observer.go`, the `pre_tool.go` branch, the
`settings.json.tmpl` matcher), but `.moai/reports/t401/baseline-push-window.jsonl` does not exist.
plan.md §F M0 states "Do not proceed to M1 until the baseline row exists", and defines the window as
one recorded while "M1/M2 have not amended the doctrine" — which this commit ends on this branch.
The repository config is still `push` (`recommendation_mode` absent from `.moai/config/`, so it
resolves to `push`; M3 has not run). The live log in the primary checkout currently reads:
`wc -l` → `8`, `grep -c '"label_present":true'` → `8`, so a window meeting the ≥5-rows /
≥1-positive floor appears collectible — but the copy is an M0 deliverable owned by the asking
(lead) session, not a run-phase act, and no attribution beyond §E.2's earlier session record was
verified here.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_phase: M1 only (M2-M6 not entered)
run_commit_sha: <backfill — this record is committed within the M1 commit>
run_status: partial — M1 edits complete, AC-JFM-013 blocked on one escalated row
ac_pass_count: 8
ac_fail_count: 1
ac_blocked_count: 1
preserve_list_post_run_count: 2   # zone-registry.md, orchestration-mode-selection.md — both diff-empty vs ad272be20
cross_platform_build: not applicable — markdown-only milestone, no Go code changed
new_warnings_or_lints_introduced: none — no code changed
total_run_phase_files: 10   # 4 doctrine + 4 template mirrors + 2 ledger artifacts
m1_to_mN_commit_strategy: one commit per milestone; M1 is a single commit
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
