---
id: SPEC-MEMORY-STORE-RECONCILE-001
title: "Auto-memory store reconciliation and index-budget premise correction"
version: "0.3.0"
status: completed
created: 2026-08-31
updated: 2026-09-01
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: "internal/config, .claude/rules/moai, .claude/output-styles, internal/template/templates"
lifecycle: spec-anchored
tags: "auto-memory, memory-store, token-budget, vacuous-guard, doctrine-correction"
tier: M
---

# SPEC-MEMORY-STORE-RECONCILE-001 — auto-memory store reconciliation and index-budget premise correction

## HISTORY

> **Provenance rule for this table.** `version:` equals the latest row, and every row states a
> change the document actually contains. Every figure in this SPEC is attributed to
> `.moai/reports/t383/measurements.md`, to a command a reader can re-run, or is explicitly
> labelled an unattributed recollection.

| Date | Version | Change |
|---|---|---|
| 2026-08-31 | 0.1.0 | Initial draft (plan-phase). Card t383. Written against `.moai/reports/t383/measurements.md`; the card's dispatched premise (index length is the binding constraint) recorded as falsified in §A.2. |
| 2026-08-31 | 0.2.0 | Plan-audit iteration 1 (FAIL 0.72) revision. **D1**: the spine's headline figure was a unit error — "47% of index entries (58 of 123)" read an occurrence count as a line count. Restated throughout as *58 dangling occurrences → 44 unique missing files → 40 of 123 entry lines (32.5%)* per measurements.md §M5a, and §D's diet-decline argument re-justified at 32.5%. **D3**: every `moai memory doctor` citation now carries `--dir` plus an `exists`/`index_lines` precondition (§M5b: the bare invocation measures an absent worktree-slugged store and returns a passing 0). **D4**: Template-First mirror obligation added (C6, REQ-MSR-014 note, §A.3 mirror form). **D7**: the target doctrine file is paths-scoped and does not load for the session it targets — REQ-MSR-004 re-hosted (C5). **D8**: commit-target prohibition promoted from a constraint to REQ-MSR-015. **D11**: REQ-MSR-005 wording aligned to "absent from". **D9**: the unattributed "19 markers" figure replaced by the measured present count. Gaps G2 and G4 recorded in §D. |
| 2026-08-31 | 0.3.0 | Plan-audit iteration 2 (FAIL 0.78) revision. **N1**: the 0.2.0 "fix" replaced one wrong relation with another and mandated it [HARD] — the 58→44 gap is a **line-population** difference, not deduplication (dedup is worth 2). Re-measured independently in this worktree (`.moai/reports/t383/measure-n1.sh`); §A.2.1 now records **both** wrong revisions, C3 corrected from "177 → 221" to a **+58 delta**, and every restatement now names scope **and** unit. **N2**: the doctor JSON is case-asymmetric — `.code` yields 0 while `.Code` yields 58 — so AC-MSR-009 carries the `jq` selector verbatim plus a before-count ≥ 1 red direction. **N3**: AC-MSR-014's grep passed on a missing file (`grep` exits 2); now `test -f` first, exit code exactly 1, with a planted red direction. **N4**: the two-form R4 prescription was self-contradictory and made AC-MSR-002 vacuous over an empty set — collapsed to ONE form with reference values living only in this SPEC and the report. **N5**: REQ-MSR-011 gains a second mandatory delta (unique `.md` targets file-wide), because the entry-line metric sees 135 of 190 targets. **G5**: copy set decided — all 58 (§A.2.2). **N8**: G4's enumeration extended to three surfaces incl. `moai.md:165`. **N13**: `Where` → `When` on REQ-MSR-011/014. Optional N9-N11 applied. |

## §A Context

### A.1 The motivating incident — a lesson that was never written

A session (lane-8) had a lesson to record and wrote **nothing**. Its reasoning, as reported:
adding an index line pushes `MEMORY.md` further past its ceiling, and a topic file written
*without* an index line is undiscoverable — so neither branch looked available, and the lesson was
dropped.

That is a real loss, and it is why this card exists. The loss was not caused by a full index. It
was caused by a session believing two things the plan-phase measurement does not support: that the
ceiling is a settled 25,600 bytes, and that it was binding. The corrective is therefore aimed at
the **premise**, not at the file size.

Where the belief came from is only partly reachable from this repository. Three in-repo files
assert the cut (§A.5), and this SPEC corrects all of them. A session's own agent-level instructions
also state a truncation cap, and nothing in this repository can edit those — named in §D so the
card does not claim a completeness it cannot deliver.

### A.2 What the measurement overturned

Evidence base: `.moai/reports/t383/measurements.md` (tree `9328a5242`, tool `moai v3.1.2` build
`343399d2f`; the report records that the judging build's `memory doctor` code path is byte-identical
to the tree's, so the build lag does not affect M4/M5). Figures below are cited from that file and
are not re-derived here.

| Ref | Finding |
|---|---|
| M1 | The active index is **26,280 bytes**, but **18,463 characters** and **163 lines**; 45% of its bytes carry 3,970 of its characters (CJK). |
| M2 | The 25,600-byte cut is **not confirmed**. The file exceeds it by 680 bytes, yet the measuring session's injected context contained the file's final line — no cut occurred. Three explanations remain undistinguished: a character-based cap, a line-only cap, or a larger byte cap. |
| M3 | `internal/config/token_budget_guard.go` measures `repoRoot/MEMORY.md`, which **does not exist**. The MEMORY.md slot always contributes 0 tokens; its unit tests pass only because they supply a fixture. |
| M4 | `moai memory doctor --dir <active store>`: 177 topic files (cap 50), 163 index lines, **46** `MEMORY_ORPHAN_NOT_INDEXED`, **58** `MEMORY_DANGLING_INDEX_LINK`. |
| M5 | There are **two stores**. The active one (`$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory`, 178 files) is what a session loads. A legacy one (`$HOME/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory`, 1,098 files, its own 38,304-byte index) is not loaded. **58 of 58** dangling occurrences resolve in the legacy store. |
| M5a | The dangling figure has **two scopes and three units**, and this SPEC has now collapsed them wrongly twice — see §A.2.1. |
| M5b | `moai memory doctor` **without `--dir`** measures nothing in this worktree. The store is derived from `CLAUDE_PROJECT_DIR` else `os.Getwd()`, so inside a worktree the slug gains the worktree path and both candidate stores report `exists:false, index_lines:0, findings:null`. A count read from that invocation is 0 because nothing was measured. |

### A.2.1 The relation, corrected twice — scope before unit

Two wrong relations have been written into this SPEC, and both are recorded rather than
overwritten, because the second was introduced while fixing the first:

- **rev 0** wrote "47% of index entries (58 of 123)" — read an occurrence count as a line count.
- **rev 1** wrote "58 occurrences resolving to 44 unique files" — attributed the 58→44 gap to
  **deduplication**. Dedup is worth **2**, not 14. The real gap is a **line population**
  difference: `grep -c '^- \['` does not match the grouped line shapes that carry the other 14.

Re-measured independently in this worktree at iteration 3 (`.moai/reports/t383/measure-n1.sh`,
whose verbatim output is in `.moai/reports/t383/measurements.md`):

| Scope | occurrences | unique targets | unique **missing** |
|---|---|---|---|
| Entry lines only (`^- \[`) | 137 | 135 | **44** |
| Whole file, any line shape | 192 | 190 | **58** ← what `moai memory doctor` reports |

Missing targets reachable **only** from non-`^- \[` lines: **14**. Entry lines carrying at least
one missing target: **40**, of which **0** are partially dangling — every affected entry line is
entirely dead.

[HARD] The correct sentence, used wherever this figure appears: **58 unique missing targets
file-wide; 44 of them reachable from `^- \[` entry lines and 14 only from other line shapes;
and 40 entry lines carry at least one, none of them partially.** Never restate any of these
numbers without naming **both its scope and its unit** — scope first, because rev 1 got the
unit right and the scope wrong.

**That sentence carries no denominator, and the omission is the point.** 40 is a *defect*
figure and has now held across five readings; an entry-line total is a *size* figure, and this
SPEC's own stability result (below) is that size figures rot. The total has moved four times —
123 → 124 → 126 → **130** (run-phase M0) — so had this [HARD] sentence been copied into
doctrine reading "40 of 123", it would have been false before the run-phase began, and
uncorrectable afterwards because doctrine is not re-measured. Where a proportion is genuinely
wanted, derive it at read time rather than pinning it:

```bash
# 40 is the measured defect figure; the denominator is measured, never quoted.
echo "$(( 100 * 40 / $(grep -c '^- \[' "$D/MEMORY.md") ))%"
```

This is the same R4 moving-coordinate remedy (§A.3) the rest of the SPEC applies: the command
decides, and any value is a dated reference sitting in this file and in `.moai/reports/t383/`.

**A stability result that decides which figures may be pinned.** Between the M1 measurement and
this one a concurrent session added an index entry: the index went 26,280 → 26,577 bytes,
123 → 124 entry lines, 189 → 190 unique targets file-wide. Across that mutation the **defect**
figures held exactly — 58 / 44 / 14 / 40 unchanged. Size figures move under a live writer; defect
figures did not. That is why the acceptance criteria pin no size literal (AC-MSR-010) while
AC-MSR-009 may target an exact 0.

Consequence: 40 entry lines lead nowhere, each entirely dead — a session following one gets
nothing, not a degraded result — plus 14 further dead targets on line shapes the entry-count metric
cannot see at all. Separately, 46 active topic files carry no index entry. The prior compression
optimized against a constraint the measurement does not confirm was binding.

The dispatched framing is corrected in one place: **index length is not the defect; store
divergence is.** Dispatch candidates (a) and (b) — sectioning the working-discipline block out, and
grouping-and-shortening entries — are both index diets and are declined in §D. Candidate (c) was
executed and produced M1-M5b above.

### A.2.2 The copy set is all 58 — decision, not an open question

The 14 targets invisible to `^- \[` split cleanly, and the split is the reason this needed
deciding rather than assuming:

| Group | Count | Where | Shape |
|---|---|---|---|
| Live working-discipline lessons | 6 | `## 작업 규율`, lines 38 / 50 / 54 | `·`-separated thematic group lines |
| Archive roll-ups | 8 | `## 종료/보관 (참조용)`, lines 155-162 | `- <topic>: [representative](file.md) 외 N편` |

**Decided: copy all 58.** Four reasons, in the order that decided it:

1. An archive roll-up **names a representative file the reader is meant to open** — "외 N편" means
   "and N more", so the named file is the entry point into the archived set. Deliberately archived
   does not mean deliberately unreadable.
2. `moai memory doctor` counts all 58 either way. Excluding the 8 would force AC-MSR-009 off an
   exact 0 and onto "8 remaining, each individually excused" — a criterion whose target is an
   exclusion list is weaker to check and invites the next reader to grow the list.
3. The error cost is asymmetric. Copying 8 unwanted archive files costs 8 files in a store this
   card already accepts pushing further over its cap (C3); not copying them leaves 8 permanently
   dead links **plus** a bespoke exclusion mechanism to maintain.
4. Copy-only and never-overwrite (REQ-MSR-010) makes the decision reversible by deleting the copies.

**Three defects meet at one place, and the coincidence is not incidental.** The three lines
carrying the live 6 are exactly the grouped lines that `grep -c '^- \['` is blind to (N5 /
REQ-MSR-011), they sit in the very section dispatch candidate (a) proposed to move out wholesale
(§D), and they hold 6 of the 14 targets the rev-1 relation mis-explained (§A.2.1). A diet applied
to that section would delete lines whose contents the entry-count guard cannot see, removing live
lessons while every metric in this SPEC's earlier revisions reported no change. That is the single
strongest argument for REQ-MSR-012, and it was invisible until the scope error was corrected.

### A.3 M2 is evidence against a strict byte cut — not proof of the cap's shape

[HARD] This SPEC does **not** assert the loader's cap shape. The Claude Code loader is not in this
repository, so its behaviour cannot be read from source, and one non-truncation observation cannot
distinguish the three surviving explanations. Every budget figure here is written in the
moving-coordinate remedy R4 form of `.claude/rules/moai/core/verification-claim-integrity.md`
§2.1 — the re-measuring command first, any value second, dated and marked a reference.

**ONE form, written identically into both doctrine copies** — the local file
`.claude/rules/moai/workflow/moai-memory.md` and its template mirror:

```bash
# Re-measure your own index. `moai memory doctor` reports the store it resolved;
# pass that path back with --dir, and read a count only when the JSON says exists: true.
moai memory doctor
moai memory doctor --json --dir "<the store path the line above printed>"
```

[HARD] **No reference value appears in either doctrine copy.** Every dated figure for this machine
lives in this SPEC (§A.2, §A.2.1) and in `.moai/reports/t383/`, and nowhere else. Two reasons, and
the second was found by plan-audit rather than by design:

1. The mirror ships to users and is scanned by the template-neutrality CI guard, so a
   machine-specific store path, an internal date, or a SPEC ID in it is a guard violation.
2. An earlier revision prescribed *two* forms — a local one carrying dated values and a neutral
   mirror one — while simultaneously requiring the two files to be byte-identical (AC-MSR-013).
   Those cannot both hold. Worse, resolving the conflict toward the mirror form emptied the local
   file of numeric values, which made AC-MSR-002's universal ("every numeric value carries its
   date") **vacuously true over an empty set** — the SPEC's own named failure mode, occurring
   inside the criterion written to enforce R4. AC-MSR-002 is accordingly restated as a conjunction
   whose positive half (the section *contains* the re-measuring command) cannot be satisfied by
   absence of subject matter.

The R4 discipline is unchanged and is discharged one level up: the command decides, and the values
sit in the SPEC and the report where they are dated, attributed, and outside the distributed
artifact.

### A.4 The stores are outside the repository

[HARD] Both memory stores live outside the git repository. **No file under either store is a commit
target of this SPEC** (REQ-MSR-015). The only committed artifacts are this SPEC directory, the
evidence report under `.moai/reports/t383/`, and the in-repo doctrine and code files listed in
§A.5. Every store mutation this SPEC prescribes is an operator act performed in run-phase whose
before/after evidence is committed as a report, never as a diff of the store.

### A.5 The in-repo surfaces this SPEC touches

Measured on this tree (all six file pairs verified byte-identical local↔mirror by `diff -q`, rc=0):

| Surface | Why it is in scope |
|---|---|
| `.claude/rules/moai/workflow/moai-memory.md` lines 29, 117, 175 | Three separate assertions of the "200 lines OR 25KB" cut in one file. M1 rewrites all three; correcting one and leaving two is a doctrine that contradicts itself. |
| `.claude/output-styles/moai/moai.md` line 165 | Restates "MEMORY.md 200-line/25KB loading" and is an **always-loaded fixed slot**, so it reaches every session — a stronger surface than the paths-scoped file above. |
| `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol | Always-loaded (no frontmatter, verified), and the surface a session actually reads when writing a lesson. Host for REQ-MSR-004 (see C5). |
| `internal/config/token_budget_guard.go` | The vacuous slot and the constant encoding the unconfirmed cut. |
| The template mirror of each of the three doctrine files | Template-First [HARD]; see C6. |

## §B Requirements (GEARS)

### B.1 Premise correction (doctrine)

- **REQ-MSR-001** — Every in-repo assertion of the "200 lines OR 25KB" cut enumerated in §A.5 shall
  be replaced by a re-measuring command plus a dated reference value; no such assertion shall
  remain stated as settled fact.
- **REQ-MSR-002** — The doctrine shall state that compressing the index means making entry lines
  **shorter**, never removing them: removing an index entry makes its topic file unreadable to
  every future session.
- **REQ-MSR-003** — The doctrine shall record the two-store hazard of M5 and shall name
  `moai memory doctor` as the tool that reports both candidate stores, including the `--dir`
  requirement and the `exists` precondition of M5b.
- **REQ-MSR-004** — **When** a session finds the index at or beyond its reference size and has a
  lesson to record, the guidance shall direct it to write the topic file **and** its index line,
  and shall not present dropping the lesson as an available branch. **Where** that guidance is
  placed, it shall be a surface the writing session actually loads (C5) — placement in a
  paths-scoped file that the session does not load fails this requirement.

### B.2 Vacuous-guard repair (Go)

- **REQ-MSR-005** — The always-loaded token guard shall not enumerate a fixed surface slot naming a
  path **absent from** the repository tree.
- **REQ-MSR-006** — **When** the fixed-slot enumeration names a path absent from the repository
  tree, the guard's test suite shall fail.
- **REQ-MSR-007** — The removal shall not change the enumerated always-loaded surface other than by
  deleting the `MEMORY.md` slot: the surface list before and after shall differ by exactly that one
  element, and the token total shall change by exactly the tokens that element contributed.
- **REQ-MSR-008** — The guard shall not carry a constant encoding the unconfirmed 25,600-byte cut.

### B.3 Store reconciliation (the deterministic half of M5)

- **REQ-MSR-009** — The active memory store shall contain a file for every `.md` target its own
  `MEMORY.md` names.
- **REQ-MSR-010** — The reconciliation shall be **copy-only**: no file in the legacy store shall be
  deleted, moved, or modified, and no file already present in the active store shall be overwritten.
- **REQ-MSR-011** — **When** a change touches `MEMORY.md`, the change record shall report, before
  and after and captured inside the same measurement window, **two** reachability metrics plus the
  byte count: (a) the `grep -c '^- \['` entry-line count, and (b) the count of unique `.md` link
  targets file-wide. **Neither** metric shall decrease; a decrease in either shall be recorded as a
  failure, not as a saving. Metric (b) is mandatory because (a) is measurably blind: it sees only
  the targets sitting on `^- \[` lines, while the file carries substantially more on other line
  shapes, so deleting a grouped line carrying six links leaves (a) unmoved (§A.2.1, §A.2.2). The
  blindness is a structural property of the anchor, not a function of the file's current size, so
  it is stated without a literal; the gap is re-derived by `.moai/reports/t383/measure-n1.sh`
  (dated reference 2026-08-31: 141 entry-line targets against 196 file-wide, a 55-target blind
  spot that includes 14 of the 58 missing ones).

### B.4 Decisions this SPEC closes

- **REQ-MSR-012** — This SPEC shall not shorten, group, section out, or remove any `MEMORY.md`
  index entry.
- **REQ-MSR-013** — The `(session: <8-char>)` correlation markers that a prior compression removed
  shall **not** be restored retroactively; the forward-looking obligation in
  `.claude/rules/moai/workflow/session-handoff.md` § Auto-Memory Integration item 5 shall be left
  unchanged, so newly written entries continue to carry the marker.

### B.5 Distribution and safety

- **REQ-MSR-014** — **When** this SPEC edits a file under `.claude/` that has a mirror under
  `internal/template/templates/`, the mirror shall be updated in the same change and the binary
  rebuilt; neither copy shall carry a machine-specific path, an internal date, or a SPEC ID (§A.3).
- **REQ-MSR-015** — No file under either memory store shall be a commit target of this SPEC.

Requirement count: 15 (Tier M ceiling 16).

## §C Constraints

- **C1** — No memory-store file is committed (§A.4, REQ-MSR-015). Store evidence is committed as a
  report.
- **C2** — The reconciliation runs while other sessions may be reading **and writing** the store.
  Copy-only and never-overwrite (REQ-MSR-010) is what makes a concurrent reader safe: the operation
  is purely additive and is reversible by deleting the copies. A concurrent *writer* is expected —
  REQ-MSR-004's own new guidance tells sessions to add index lines — which is why REQ-MSR-011 binds
  a delta rather than a literal count.
- **C3** — The active store's topic-file count rises by **+58** as a direct consequence of
  REQ-MSR-009, moving further from the cap of 50. Accepted and named rather than hidden: the cap's
  own validity is unexamined and its reduction is deferred (§D).

  The figure is stated as a **delta, not an endpoint**, and that is deliberate on two counts.
  First, 58 is the whole-file missing-target count (§A.2.1) — an earlier revision wrote "177 → 221"
  by using the 44 entry-line figure, which would leave `moai memory doctor` reporting 14 and
  AC-MSR-009 unachievable. Second, the base itself moves: the store measured 177 topic files at M4
  and **178** at the iteration-3 re-measurement, because a concurrent session added one. An
  endpoint literal would be stale before run-phase starts; the delta is not. For orientation only,
  dated 2026-08-31: 178 + 58 = 236 — a reference, not a target.
- **C4** — No claim is made about the loader's cap shape (§A.3).
- **C5** — **Measured loading scope.** `.claude/rules/moai/workflow/moai-memory.md` carries
  `paths: "**/.moai/specs/**,**/.claude/agent-memory/**"`, so it is paths-scoped rather than
  always-loaded. A session writing a topic file into the memory store touches neither path, so the
  file this card corrects is **not in that session's context**. Widening `paths:` cannot fix this:
  the store lives outside the repository, so no repo-relative glob can name it. REQ-MSR-004's
  write-anyway clause therefore goes into `.claude/rules/moai/core/moai-constitution.md`
  § Lessons Protocol, which has no frontmatter (verified) and is the surface a session reads when
  writing a lesson. The paths-scoped file still receives REQ-MSR-001..003, which serve the reader
  who opens it deliberately.
- **C6** — **Template-First is [HARD]** (`CLAUDE.local.md` §2). Every `.claude/` file edited here has
  a byte-identical mirror under `internal/template/templates/` today (`diff -q`, rc=0 on all three
  pairs). Editing the local copy alone leaves the mirror stale, and per §2.3 the next `moai update`
  **overwrites the corrected local file with the stale template copy** — a path-collision overwrite
  invisible to `git status | grep '^ D'`, so the correction would silently revert.

## §D Exclusions

The items below are **out of scope** for this SPEC. Each names where it goes instead.

### Out of Scope — index dieting

- Shortening, grouping, or re-sectioning index entries (dispatch candidates (a) and (b)). Three
  reasons, each measured. **(i)** The constraint such a diet optimizes is unconfirmed (M2).
  **(ii)** 40 entry lines are **entirely** dead — 0 partially dangling — so shortening them
  produces shorter dead lines. **(iii)** The decisive one, and it only became visible once the
  scope error was corrected: a diet **cannot be verified** by this SPEC's own primary metric, which
  sees only the entry-line subset of the file's unique targets (§A.2.1; dated reference
  2026-08-31, 141 of 196 — re-measure with `.moai/reports/t383/measure-n1.sh`).

  The fraction has been restated twice and the argument does not rest on it. Rev 0 said
  "approximately half dead links" (wrong unit); the corrected figures are 40 affected entry lines,
  and 58 of the file's unique targets file-wide. Reason (i) is independent of any fraction, and reason
  (iii) is a verifiability argument rather than a magnitude one — which is why the decline stands
  at 33% exactly as it did at the mistaken 47%. Dieting is reconsidered only after a follow-up
  settles the cap's shape.
- Moving the "작업 규율" block into its own topic file behind a single index line. That is entry
  removal wearing a diet's clothes, and §A.2.2 measures the damage precisely: three of that
  section's grouped lines carry 6 live lesson targets that `grep -c '^- \['` does not count, so the
  move would delete them while every entry-count metric reported no change.

### Out of Scope — judgement-bearing store work

- Triage of the 46 `MEMORY_ORPHAN_NOT_INDEXED` files (index each, or archive it). Each is a
  per-file judgement about whether a memory is still wanted. Deferred to a named follow-up card.
- Reducing the active store from 177 topic files toward the cap of 50. Same reason, plus the cap
  itself has not been shown to be the right number — and C3 moves the count the other way.
- Disposition of the legacy store's 1,098 files (archive, migrate, or leave). Deferred; the legacy
  store is untouched here by REQ-MSR-010.

### Out of Scope — loader behaviour

- Determining whether the loader's cut is character-based, line-based, or a larger byte cap. The
  loader is not in this repository; settling it needs a deliberate truncation experiment, which is
  its own card.

### Out of Scope — the part of the belief this card cannot reach

- A session's own agent-level instructions state a truncation cap independently of this repository.
  No card here can edit them, so the incident's belief is corrected **only for the in-repo
  surfaces enumerated in §A.5**. Recording this is what keeps AC-MSR-005 an honest claim about
  doctrine rather than a claim about every session's behaviour.

### Out of Scope — store-derivation divergence (gap G4)

- **Three** in-repo statements about where the store lives disagree with the code, and this SPEC
  edits all three files, so it must not re-assert any of them:

  | Surface | What it says | Status |
  |---|---|---|
  | `.claude/rules/moai/workflow/moai-memory.md:27` | the path derives from the **git repository root**, so all worktrees share one memory directory | contradicted by the code |
  | `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol | `~/.claude/projects/{project-hash}/memory/` | names the **legacy** store |
  | `.claude/output-styles/moai/moai.md:165` | `~/.claude/projects/{hash}/memory/` | names the **legacy** store; M1 step 2 edits this exact line |

  The code derives the path from `CLAUDE_PROJECT_DIR` else `os.Getwd()`
  (`internal/cli/session.go:263-272`), which is exactly what produces M5b.

  [HARD] The rule already applied elsewhere binds all three: every edit this card makes to these
  files refers to the store **by the command that resolves it**, never by a literal path. The
  `moai.md:165` line is the one most easily missed, because M1 rewrites its *cap* clause and could
  leave its *path* clause standing — rewriting a sentence while preserving a wrong claim inside it
  is re-assertion, not preservation. Correcting the three statements — and deciding which of
  doctrine and code is wrong — is deferred to a named follow-up.

### Out of Scope — session-handoff doctrine

- Changing the forward-looking `(session: …)` obligation in `session-handoff.md`. REQ-MSR-013
  leaves it as it stands; only the retroactive-restoration question is answered here.

## §E Success Criteria

Acceptance criteria live in `acceptance.md`. In summary, this SPEC closes when every in-repo
assertion of the unconfirmed cut is replaced, the write-anyway clause sits in a surface the writing
session loads, the Go guard no longer carries a slot it cannot measure, all six local/mirror pairs
are byte-identical again, the active store's dangling occurrences are 0 against a store the tool
confirms `exists`, and the before/after evidence is committed under `.moai/reports/t383/`.

## §F Cross-References

- `.moai/reports/t383/measurements.md` — the plan-phase evidence base (M1-M5b).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.1 — the R4 moving-coordinate remedy
  used for every figure here, and §2.2 tool-provenance, which gap G2 (plan.md §I) applies to lint.
- `internal/cli/memory.go` `memoryCandidateStores` — enumerates both stores; `internal/cli/session.go`
  `resolveProjectDir` — the derivation behind M5b and gap G4.
