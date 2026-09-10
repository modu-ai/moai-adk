# SPEC Run-Entry Plan Audit Gate: SPEC-TODO-ARCHIVE-QUERY-001

Stream: run-gate (date-based, per spec-workflow.md § Report Persistence) — written by plan-auditor at run entry, 2026-09-01T08:56Z.
Verdict: **PASS**
Overall Score: **0.91** (Tier M threshold **0.80**)

**Why this gate re-executed**: skip condition 3 (artifact-hash unchanged) is broken by the
lead-mandated premise repair (correction round 3); conditions 1 (verdict PASS-WITH-DEBT) and
2 (0.875 ≥ 0.80) held at plan-audit iteration 2. The re-execution is a **full re-audit** of
the corrected artifact set, with priority on (1) whether the corrections introduced new
defects, (2) the standing must-pass firewall, (3) fresh adversarial eyes at the new tree
state.

**M1 Context Isolation**: the iteration history and the author's responses were read only as
claims to re-verify; every load-bearing value below was re-measured in this run unless
explicitly attributed. The reasoning context passed in the dispatch was ignored; this audit
proceeded from the artifact files alone.

**Tree audited**: worktree `.claude/worktrees/t394`, branch `WT-todo-done-history`, HEAD
`e8ae9798a` (fast-forwarded from `2c18091d1` at run entry). Tracked-file modifications at
audit start: **0** (`git status --porcelain | grep -v '^??' | wc -l` → `0`).

---

## Claim / Evidence / Baseline-attribution

Every measurement below was taken in this audit run, at HEAD `e8ae9798a` (source-tree
citations) and against the **primary checkout's live queue**
(`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`, read at
2026-09-01T08:5xZ via plain `sqlite3`; the DB does not exist in this worktree — `ls -d
.moai/state/todo` check: *No such file or directory*, confirming the provenance
block's checkout attribution).

### Part 1 — The correction-round-3 premises, independently re-measured

| Correction claim | Independent verification run | Observed | Verdict |
|---|---|---|---|
| Installed binary is the develop build `v3.1.2-1033-g64bba61aa` (mtime 06:37 KST); `moai version` prints banner **plus** descriptor | `ls -la /Users/goos/go/bin/moai` + `moai version` | mtime `Sep 1 06:37`; output carries `moai-adk v3.1.2` banner **and** the same-output line `v3.1.2   v3.1.2-1033-g64bba61aa   built 2026-08-31T21:37:48Z`; exit 0 | **confirmed** — the banner-only-reading root cause is real and reproducible: the two lines sit on the same output |
| Main-era `v3.1.2` deleted on `done` (the withdrawn part was only the "installed" claim, not the behaviour) | `git show origin/main:internal/cli/todo.go` around the `done` handler | `Use: "done <n>"`, `Short: "Remove a card from the backlog queue by id"`, `rec.Items = append(rec.Items[:i], rec.Items[i+1:]...)` + `rec.RemoveFindingsNaming(id)` — row removal, not archive | **confirmed** — §A.4's pre-archive-era destruction narrative rests on a real source fact |
| Snapshot C: 408 / 108 / 11 archived (t400, t332, t333, t338, t343, t350, t356, t357, t358, t362, t399), 5 tables | `sqlite3` at read time | The queue **moved since C**: `last_seq` **410**, `count(*)` **108**, `archived_items` **13** — C's 11 ids are present verbatim as seq 1–11, plus seq 12–13 (`t376`, `t382`) closed after C. States `queued 76 / picked 12 / dropped 20` vs C's `74/14/20`. The movement decomposes exactly as +2 added / +2 closed-and-archived | **confirmed at C; movement is coherent and does not disturb the frozen estimate** |
| Corrected arithmetic 408 − 108 − 11 = **289**, identical to snapshot B's 289 → no destruction since the install | arithmetic + my own counters | 408−108 = 300; 300−11 = 289 ✓; snapshot B: 402−113 = 289 ✓ (archive-tables-absent condition held there). **My counters: 410 − 108 − 13 = 289** — the same 289 after the queue moved | **confirmed, strengthened** — the "frozen at the pre-archive era" claim survived an independent re-measure at a later instant |
| t393 never lost: live in `items`, state `picked`, added `2026-08-31T15:43:00Z` | `SELECT id, state, added_at FROM items WHERE id='t393'` | `t393\|picked\|2026-08-31T15:43:00Z` — both fields exact | **confirmed** |
| t81/t83/t88/t89 measured-unrecoverable (absent from both stores) | `SELECT id FROM items WHERE id IN (...)` + archived scan | absent from `items`; absent from the 13-row `archived_items` | **confirmed** |
| The conditional-subtraction condition statement (§A.4: `last_seq − count(*)` valid only while no archive tables) | read + arithmetic above | correctly stated; correctly applied retroactively to A/B and correctly corrected for C's post-install world | **confirmed** |

Snapshot C is correctly framed: *"the figures are live counters at that instant, under the
same re-measure discipline as A and B"* (spec.md:215-216), attributed to the lead dispatch
(`16:20–16:25 KST`) — an honest attribution of a handed-down measurement, not a
measurement-masquerade. My re-measure at a later instant confirms the discipline works: every
C-era count that moved is accounted for by observable queue events, and the one figure that
matters (289 destroyed, pre-archive era) did not move.

### Part 2 — Must-pass firewall (re-run at e8ae9798a)

- **[PASS] MP-1 REQ number consistency** — `grep -oE 'REQ-TAQ-[0-9]{3}' spec.md | sort -u` →
  `REQ-TAQ-001`…`REQ-TAQ-015`; definition count 15; duplicates in defs: 0; uniform 3-digit
  padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the `REQ-XXX` requirement layer
  (spec.md §B) only; all 15 match a canonical pattern: event-driven `When…shall` (001, 002,
  003, 006, 008, 009), compound `Where…when…shall` (004), `Where…shall` (013),
  ubiquitous `shall` (005, 007, 010, 011, 015), canonical `shall not` negatives (012, 014).
  The verification layer's Given-When-Then entries in `acceptance.md` are correctly NOT
  graded here (M3 § Scope two-layer table).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present, correct types:
  `id`, `title`, `version: "0.2.2"` (quoted semver — bump from 0.2.1 recorded in HISTORY),
  `status: draft`, `created/updated: 2026-09-01`, `author`, `priority: P2`, `phase:
  "v3.1.5 target"`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string).
  No rejected snake_case alias. Optional `tier: M`, `related_specs` present.
- **[N/A] MP-4 language neutrality** — single-programming-language (Go) SPEC; auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — four `related_specs` all exist:
  `SPEC-TODO-DESTRUCTIVE-GUARD-001` **completed**, `SPEC-TODO-SQLITE-001` **completed**,
  `SPEC-KANBAN-TODO-CLI-001` **in-progress**, `SPEC-TODO-ANALYSIS-001` **completed**.
  None ∈ {retired, superseded, archived} → no reconciliation obligation, no BLOCKING
  finding. (First probe output ordering was ambiguous; settled by a third measurement —
  value-set identical across all three probes, and matching iteration 2's record.)
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over
  spec.md/plan.md/acceptance.md/progress.md → no match. (`research.md` does not exist —
  Tier M set; N/A per precedent, stated.)

The firewall does not fire.

### Part 3 — Citation integrity at the new head

- **The ff claim verifies**: `git diff --stat 2c18091d1 e8ae9798a -- <the 8 cited files>` →
  **empty** (exit 0) — the fast-forward touched none of them, so `2c18091d1`-pinned
  citations carry to `e8ae9798a` byte-identically.
- **All probed citations resolve with the claimed content**: `state_dir.go` :37
  (`const stateDirName = "todo"`), :43 (`legacyStateDirName`), :79 (`func resolveStateDir`),
  :129 (`func BacklogPathForRoot`); `backlog_sqlite.go` :93 (IF-NOT-EXISTS comment in
  range), :113 (`state` CHECK in DDL), :124/:133 (`archived_items`/`archived_findings`);
  `backlog_store.go` :16 (never-derived-mark comment in range), :223
  (`func (r *BacklogRecord) ArchiveCard`), :772-778 (raise-only mark logic);
  `backlog_migrate.go` :102 (`e.readArchive`), :106-110 (`readLastSeq` → `rec.LastSeq`),
  :411 (`func loadLegacyBacklogJSON`); `todo.go` :148-152 (16 `newTodoXxxCmd`
  registrations — count re-derived = **16**), :310-311 (unfiltered render loop), :409
  (`rec.ArchiveCard(id)`); `todo_why.go` :34-37 (`no findings` conflation);
  `todo_export.go` :75 (`writeExportAtomic`), :100-110 (`discloseArchiveDowngradeCost`);
  `todo_undone.go` :67-68 (`rec.RestoreCard(id)`).
- The progress.md §E.2 pre-flight record ("all 22 cited spots resolve; 0 refreshed") is
  **consistent with this audit's independent probes** at every spot I opened.

### Part 4 — Iteration-2 defect retention (regression check)

All seven iteration-2 defects remain closed after correction round 3; none re-opened:

- **D1** (fetch-depth ledger) — audit-response-iter1.md untouched since 04:28 (mtime), the
  correction round did not touch reports; progress.md §E.1 records the closure. Retained.
- **D2** (AC-TAQ-011 modify-path escape) — `acceptance.md:217`
  `git diff --exit-code "$C" -- "$GOLDENS"` present; prose narrowed to what clause 1
  establishes (`acceptance.md:190-201`). Retained.
- **D3** (singleness + merge-method + develop-drift) — singleness assertion
  (`acceptance.md:210`, `-eq 1`), clause-1 scoping to the pre-integration branch
  (`acceptance.md:242-251`), develop-drift residual (`acceptance.md:277-292`),
  plan.md §D merge-method dependency. Retained.
- **D4** (live counters) — `[MEASUREMENT PROVENANCE]` block; 112/113 reconciled as two
  instants (spec.md:183-189); correction round 3 **strengthened** this with the
  conditional-subtraction paragraph (spec.md:231-245). Retained.
- **D5** (checkout attribution) — provenance block names the primary checkout explicitly
  (spec.md:127-131); my worktree absence probe confirms it. Retained.
- **D6** (fixture convention exceptions) — two named exceptions at `acceptance.md:8-19`,
  both still consistent with the corrected world ("The current CLI archives instead of
  deleting" — true of the develop build). Retained.
- **D7** (constructor naming contract) — plan.md M1 naming contract (plan.md:193-211).
  Retained.

No iteration-2 defect regressed; no stagnation.

### Part 5 — Adversarial pass over the corrections (new-defect hunt)

Checked and found sound:

1. **Arithmetic coherence across all surfaces** — §A.4 (408−108=300, −11=289), plan §B.1
   (corrected arithmetic named, 289 frozen), HISTORY 0.2.2 (identical figures). No surface
   carries a contradictory number.
2. **t393 split consistency** — spec §A.2 incident note ("asserts no destruction" — the
   incident is recorded as observed, the re-measurement is a separate clause), §A.4
   (twice), HISTORY 0.2.2, progress §E.1 addendum: all agree (`t393`, live `items`,
   `picked`). plan.md does not mention t393 — no contradiction, since plan §B.1's "~289
   destroyed" is set arithmetic that excludes t393 (it sits in `items`). The 289
   arithmetic and the never-lost claim are the same world.
3. **"then-installed" stamps** — spec §D (both measurement paragraphs) and plan §B.3/§B.4
   correctly mark the v3.1.2 measurements as taken under the then-installed binary, with
   the 06:37 replacement named.
4. **acceptance.md sweep** — its pre-archive references (AC-TAQ-004's "the shape a
   pre-archive `done` leaves"; AC-TAQ-013's "the shape a pre-archive database has") are
   properties of older artifacts, correctly remain, and the one place that speaks of the
   current CLI ("archives instead of deleting") is now true of it. No AC text depended on
   the corrected premises — verified by reading every AC.
5. **Snapshot C's causal sentence** — "absent at 04:11 and 04:30, present at 16:20, with
   the 06:37 install between — the IF NOT EXISTS DDL … doing what it says" is a
   correlation-plus-mechanism statement, not an overclaimed direct observation; the DDL
   quote it leans on resolves (`backlog_sqlite.go:93`).
6. **Frontmatter/HISTORY** — 0.2.1 → 0.2.2 with a complete HISTORY row recording the
   trigger, the root cause, and the figure corrections.

Two observations, neither a defect:

- **OBS-1 (informational)** — plan §B.1 "11 rows" and spec §A.4 "The 11 cards closed since
  the install" are snapshot-C-instant values. Both sit inside the re-measure discipline with
  provenance named; my later measurement (13) is exactly the staleness the discipline
  predicts, and the load-bearing 289 is unaffected. No change required.
- **OBS-2 (informational)** — the cross-SPEC status probe needed three measurements to
  settle (first probe's output ordering was ambiguous). Value-set identical across probes;
  recorded for probe-hygiene, not against the SPEC.

---

## Category Scores (rubric-anchored, current tree state)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 | Every requirement has a single reading; the four-fates vocabulary (§A.3) and the conditional-subtraction paragraph (§A.4:231-245) remove the two readings a naive reader would otherwise take. Small deduction: §D still quotes `112` standalone (accurately attributed to its instant via provenance, but a §D-only reader meets one value of a two-instant reconciliation whose explanation lives in §A.4). |
| Completeness | 0.90 | 0.75–1.0 | All sections present; frontmatter complete; five `### Out of Scope — <topic>` headings with specific bullets; snapshot C + the frozen-estimate restatement complete the measurement story across three instants (A, B, C) plus this audit's fourth. Small deduction: the narrative now spans four measurement instants and a reader must hold the re-measure discipline in mind throughout — the discipline is stated (provenance block) rather than repeated at each site. |
| Testability | 0.90 | 0.75–1.0 | All 15 ACs binary with a named deciding command; reachable failing inputs stated for 004/011/013/014; the weasel-word scan finds only "a **proper** ancestor" (git terminology, not a weasel word). Small deduction: AC-TAQ-011 clause 2 remains a manual DoD gate — a designed, documented residual, not a defect. |
| Traceability | 0.95 | 0.75–1.0 | 15 REQ ↔ 15 AC strictly 1:1 (`acceptance.md` §D matrix); no orphan AC, no uncovered REQ (id-set comparison). The load-bearing symbol `newTodoHistoryCmd` is now a recorded plan contract (plan.md M1), closing iteration-2 D7's basis for deduction. |

**Aggregate: 0.91** (arithmetic mean 0.9125; harmonic mean 0.912 — agree to two digits).
Threshold for Tier M: **0.80**. Trajectory: 0.775 (iter1) → 0.875 (iter2) → 0.91 (this gate,
post-correction). Monotonic; no regression in any dimension.

---

## Must-Pass Results

- [PASS] MP-1 REQ number consistency
- [PASS] MP-2 GEARS format compliance
- [PASS] MP-3 YAML frontmatter validity
- [N/A] MP-4 Section 22 language neutrality
- [PASS] MP-5 D7 cross-SPEC reconciliation
- [PASS] MP-6 D8 cross-platform discipline
- [PASS] MP-7 clarification gate

## Defects Found

No defects found. Two informational observations recorded (OBS-1, OBS-2 above), both
Class: optional, neither requiring a fix.

## Gaps (what this audit explicitly did NOT observe)

- **AC-TAQ-011 clause 2** (golden regeneration from `C`) and **AC-TAQ-014's RED
  plant-and-remove pair** were not executed — both are run-phase Definition-of-Done
  obligations by the SPEC's own design (`acceptance.md` §D.7), not plan-phase defects.
- **The `done`-archives behaviour was not executed** (a write; prohibited here). It is
  established by source (`todo.go:409` → `ArchiveCard`; `backlog_sqlite.go:124`), by
  presence (13 archived rows), and by movement (11 → 13 between snapshot C and this audit,
  with `last_seq` 408 → 410) — indirect but three-way consistent.
- **The frozen 289** was re-derived from my counters (410−108−13), not from snapshot C's
  (408−108−11). Both give 289; the identity-of-result across two instants is the evidence,
  and neither measurement directly observed "no destruction happened in between" — the
  between-instant zero-destruction is an inference from the unchanged difference, which the
  SPEC itself correctly labels a conditional estimate.

## Residual-risk

- The queue is live; any figure in this report that names it is stale on read (by design,
  per the provenance block).
- `origin/develop` moves; the citation-carry guarantee is pinned to `2c18091d1..e8ae9798a`
  touching none of the 8 cited files — a future ff that touches them re-opens the
  re-verification obligation (plan.md §C already mandates re-running it).
- AC-TAQ-011 clause 3 remains develop-drift-sensitive by design (documented residual,
  `acceptance.md:277-292`).

---

## Recommendation

**PASS at 0.91 (Tier M threshold 0.80). Proceed to M0.**

The correction round 3 did not merely avoid new defects — it measurably strengthened the
artifact set: the premise that iter-2's provenance block carried (a deleting installed
binary) was replaced by a measured, source-backed boundary (the 06:37 install), the
destroyed-card figure gained a conditional-arithmetic guard that survived an independent
re-measure at a later instant, and the harness-selection card's status moved from a wrong
"grouped as lost" to an exactly-verified live row. The one gap iteration 2 carried forward
by construction (citations unverified at the then-current develop head) is now closed at
`e8ae9798a` by the run-entry pre-flight and independently re-confirmed by this audit.

The run phase's own obligations are unchanged and already recorded: M0 before M1
(mechanically, per AC-TAQ-011 clause 1), the clause-2 regeneration and the AC-TAQ-014 RED
pair observed and cited in progress.md §E.2, and CI on the pushed head as the full verdict.
