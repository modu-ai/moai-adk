# SPEC Review Report: SPEC-TODO-DESTRUCTIVE-GUARD-001

Card: t330 · Tree: `.claude/worktrees/t330` · Branch `WT-todo-destructive-guard` · HEAD `812ee01fc`
Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.75** (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. Artifacts read per the Tier M input contract: spec.md, plan.md, acceptance.md (+ progress.md).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-TDG-001..014 at `spec.md:145-168`, sequential, no gaps, no duplicates, uniform 3-digit padding.
- **[PASS] MP-2 GEARS/EARS format compliance** (judged against the REQUIREMENT layer, `spec.md` §C) — all 14 match a GEARS pattern: Ubiquitous (001, 004, 005, 006, 014), Event-driven (002, 003, 011), Where/capability-gate (008, 009), State-driven (010), Unwanted (007, 012, 013). The Given-When-Then form in `acceptance.md` is the verification layer and is correct there; graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-14`): `version: "0.1.0"` quoted, `created`/`updated` ISO, `priority: P2`, `lifecycle: spec-anchored`, `tags` comma-separated string. No rejected snake_case alias present. `tier: M` and `related_specs` are permitted extras.
- **[N/A] MP-4 language neutrality** — single-language (Go) project SPEC; no multi-language tooling surface.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — RAN `grep -H '^status:'` on the three referenced SPECs. `SPEC-KANBAN-TODO-CLI-001: in-progress`, `SPEC-TODO-SQLITE-001: completed`, `SPEC-KANBAN-QUEUE-PR-SYNC-001: in-progress`. All exist, none in {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — RAN `grep -c syscall spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate** — RAN `grep -rn 'NEEDS CLARIFICATION' plan.md` → no match (rc=1). No `research.md` in the Tier M set.

**No must-pass failure.** The FAIL below is defect-driven and score-driven, not firewall-driven.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Decisions explicit and boundaries stated, but `spec.md:72` asserts a predicate value that is false as-shipped and contradicts `plan.md:61` (D1); `spec.md:153` names a backend the measured architecture does not run (D3). |
| Completeness | 1.00 | 1.0 | HISTORY (23-27), Context (31), Decisions (84), Requirements (141), Exclusions (173) with four `### Out of Scope — <topic>` H3s each carrying specific `-` bullets, Traceability (203). Frontmatter complete. |
| Testability | 0.50 | 0.50 | Three of fourteen ACs are non-executable as written against the measured system: AC-TDG-003 legacy arm (`acceptance.md:54`), AC-TDG-006 (`71-75`), AC-TDG-007 (`77-81`). No weasel words anywhere; the remaining eleven carry a command plus a determinate observable. |
| Traceability | 0.75 | 0.75 | The `spec.md:205-217` table is complete and correct — every REQ has ≥1 AC and every AC names an existing REQ. Docked one band because REQ-TDG-006 and REQ-TDG-007 are covered only by ACs that cannot execute (D2, D3): an AC that cannot be run is not coverage. |

Aggregate (arithmetic mean) = **0.75** < 0.80.

---

## Defects Found

**D1 — `spec.md:72` (§A.4) — the load-bearing measurement is stated false and contradicts `plan.md:61` — Severity: critical — Class: blocking**

The SPEC asserts: *"The run commit `3cb258d62` names t306. `Landed("t306")` was therefore already **true** at the moment of the premature `done`."* The evidence block above it (`spec.md:65`) is a `git log **origin/develop**` query, but `Landed` does not ask about develop.

VERIFIED BY RUNNING and READING:

- READ `internal/kanban/prlink_landed.go:28` — `const LandedRef = "origin/main"`.
- READ `prlink_landed.go:39-49` — `LandedGrepArgs` substitutes `LandedRef` into the argv; `Landed` (`:68`) consumes exactly that.
- RAN `git log origin/main --perl-regexp --grep='\bt306\b' --oneline | wc -l` → **0**.
- RAN the same against `origin/develop` → **13**.

So as shipped, `Landed("t306")` was **false**, not true. `plan.md:61` states the correct reading — *"the predicate would report every develop-landed card as not-landed"* — the direct negation of `spec.md:72`. Two sections of the same SPEC give opposite values for the same predicate at the same moment.

The two failure modes are distinct and both make a default-on refusal wrong, but for opposite reasons: **as shipped** it refuses every develop-integrated card (it would have blocked the incident — by blocking everything, uselessly); **after the obvious ref correction** it passes silently on any mention. The SPEC's stated reason for the opt-in ruling (§B.2 C3, `spec.md:126`) rests only on the second and never states the ref-correction condition it depends on. The *conclusion* (opt-in, not default-on) survives; its *evidence* does not.

Required fix: rewrite §A.4 to state both modes under their conditions — (a) with `LandedRef = "origin/main"` as shipped, `Landed("t306")` is false and a default-on check refuses every develop-landed card; (b) only after correcting the ref to the live integration branch does the "already true, passes silently" form hold. Reconcile the wording with `plan.md` §C so the two do not contradict.

**D2 — `acceptance.md:77-81` (AC-TDG-007) vs `plan.md:83` (M1) — `export-json` serializes the whole record, so top-level archive fields appear in it by construction — Severity: major — Class: blocking**

VERIFIED BY READING `internal/cli/todo_export.go:69`: `encoded, err := json.MarshalIndent(rec, "", "  ")` — the exporter marshals the entire `*kanban.BacklogRecord`, field-set and all. `plan.md:83` (M1) prescribes `Archived []BacklogItem` and `ArchivedFindings []BacklogFinding` as top-level `BacklogRecord` fields. Those fields will therefore land in `export-json` output automatically.

AC-TDG-007 requires that `export-json` **not** name an archived card. `plan.md:101-103` (M5) says invisibility "should be true by construction … M5 is the assertion, not the fix" — for `export-json` the opposite is true by construction, and M5 budgets no fix.

The SPEC has an unmade decision here, and both branches carry a cost it does not state: excluding the archive from `export-json` requires custom marshaling or a separate export struct (unbudgeted work, and it silently drops the archive on the downgrade route — see D4); including it breaks AC-TDG-007 and puts unknown fields into a file whose declared contract (`todo_export.go:1-15`) is "a valid **legacy-format** `backlog.json`".

Required fix: make the decision in §B, add a requirement stating whether `export-json` carries the archive, correct AC-TDG-007 to match, and add the corresponding step to M1 or M5 in `plan.md`.

**D3 — `spec.md:153` (REQ-TDG-006), `acceptance.md:71-75` (AC-TDG-006), `:54` (AC-TDG-003 legacy arm), `:27` (§A.3 legacy basis) — there is no live JSON backend; the SPEC treats a migration source as a coequal backend — Severity: major — Class: blocking**

VERIFIED BY READING `internal/kanban/backlog_store.go:437-455` (`openEngine`, tagged `@MX:ANCHOR: the layout resolver every store operation enters through`) and `:492` (`Mutate` enters through it). State B (json present, no db) calls `migrateUnderLock` and then `return openBacklogEngine(backlogSQLitePath(s.path))`. State D (both present) treats the database as authoritative. **Every path returns a SQLite engine.** RAN `grep -rn 'MOAI_TODO\|forceLegacy\|useLegacy'` over `backlog_store.go`/`backlog_migrate.go` → no legacy-only mode exists. `todo_export.go:3-6` states the swap is one-way by design, "no config knob selecting an engine".

Consequences:

- AC-TDG-006's first repository ("holding only a legacy `backlog.json`") migrates to SQLite on the first `done`. Both arms then exercise the same engine, so the criterion verifies nothing it claims to.
- AC-TDG-003's "on the legacy backend by reading the archived top-level field" has no live file to read after migration.
- §A.3's byte-identity basis "the record file itself on the legacy backend" does not exist at comparison time.

The *design* instruction is unaffected and remains correct — expressing the archive at the `BacklogRecord` level (`plan.md:71`) is right because that is the in-memory record every engine loads and saves through. Only the requirement's framing and the ACs' procedures are wrong.

Required fix: restate REQ-TDG-006 as "expressed at the `BacklogRecord` level so it survives the migration path and the `export-json` downgrade route", drop the "two live backends" framing, and rewrite AC-TDG-006 and AC-TDG-003's legacy arm and §A.3's legacy basis against what is actually observable (a pre-migration `backlog.json` that migrates, then round-trips) — or delete the legacy arm.

**D4 — no requirement or AC covers archive survival across a JSON downgrade — Severity: major — Class: blocking**

REQ-TDG-005 (`spec.md:152`) and AC-TDG-005 (`acceptance.md:65-69`) claim downgrade survival for the **database** only. `todo_export.go:1-15` establishes that the downgrade route is `export-json` → an older binary reads that file exclusively. An older binary unmarshals it into a `BacklogRecord` without archive fields (Go silently ignores unknown keys) and re-serializes without them on the next write — the archive is destroyed with no signal. That is the same silent-loss failure §A.2 (`spec.md:52`) identifies as worse than losing the card.

Required fix: either add a requirement + AC covering the downgrade path, or add an explicit `### Out of Scope` bullet stating that the archive does not survive a downgrade-and-write cycle and why that is acceptable. Silence is not a ruling.

**D5 — `spec.md:35` (§A.1) and `acceptance.md:43` (AC-TDG-001 base tree) — the verb-surface enumeration omits `why` — Severity: minor — Class: blocking**

VERIFIED BY READING `internal/cli/todo.go:137-141`: `newTodoWhyCmd()` is registered; `internal/cli/todo_why.go:22` declares `Use: "why <n>"`. The actual surface is 15 verbs, not the 14 the SPEC lists. AC-TDG-001 presents the 14-verb list as a *base-tree observable* — an assertion that fails if actually checked. Classified blocking rather than optional because it is a stated observable inside an acceptance criterion, not prose.

Required fix: add `why` to both lists.

**D6 — `spec.md:124` (§B.2 C2) — the `--expect` convention list omits `next` (pick) — Severity: minor — Class: optional**

RAN `grep -rn '"expect"' internal/cli/*.go` → four owners: `todo_edit_move.go:91`, `todo.go:441`, `todo_drop.go:133`, `todo_drop.go:190`. READ `todo.go:425-443` — the fourth is `next` (pick), documented at `todo.go:15-18`. The omission understates the convention's breadth, so it strengthens rather than weakens C2's argument; correcting it is cheap.

**D7 — citation drift across four anchors — Severity: minor — Class: optional**

- `acceptance.md:49` cites `internal/cli/todo.go:341` for `RemoveFindingsNaming`; RAN `grep -rn RemoveFindingsNaming internal/` → the call is at **`todo.go:347`**. (`spec.md:47`'s citation of `:332` for `newTodoDoneCmd` is correct, and `:50`'s `backlog_store.go:201` is correct.)
- `acceptance.md:69` cites `backlog_sqlite.go:243-247` for the mismatch abort; READ — that range is the *stamping* `INSERT`; the abort is the `default:` case at **`:251-253`**.
- `spec.md:62` and `plan.md:52` cite `prlink_landed.go:27` for `LandedRef`; the declaration is at **`:28`**.
- `plan.md:57-59` records `git log origin/develop … | wc -l` → `10`; I measure **13**. Plausibly explained by `origin/develop` advancing since authoring, which is exactly why a frozen count is weaker evidence than the command. Not treated as a false claim.

**D8 — `acceptance.md:45-49` (AC-TDG-002) does not verify that `undone` clears the archive — Severity: minor — Class: optional**

If D2 is resolved by excluding the archive from `export-json`, the byte-identity oracle (§A.3) is blind to archive residue, and a `done → undone → undone` or stale-archive-entry hazard goes unasserted. Worth one added assertion once D2 is ruled.

---

## What I checked and found sound

Recorded so the author does not re-litigate these.

- **Decision 1's foundation holds.** READ `backlog_sqlite.go:231-235` — `ensureSchema` executes the whole `backlogDDL` via `ExecContext` on every open; READ `:89-111` — every statement in it is `IF NOT EXISTS`. A new table is genuinely free on an existing database. READ `:100` — the CHECK is exactly `CHECK (state IN ('queued','picked','dropped'))` inside `CREATE TABLE IF NOT EXISTS items`. The additive-table-vs-changed-CHECK asymmetry the ruling rests on is real.
- **The `schema_version` freeze reasoning holds.** READ `backlog_sqlite.go:50` — `backlogSchemaVersion = "1"`; READ `:248-254` — the `switch` handles `""` (stamp) and the current value, and its `default:` returns `ErrBacklogCorrupt` for **any** other value, a newer stamp included. Bumping to `"2"` would indeed make an older binary refuse the queue. REQ-TDG-005 is correct and its stated reason is correct.
- **The frozen contracts are as claimed.** READ `backlog_store.go:49-70` — `BacklogState` carries exactly three values, `BacklogItem` exactly five fields, with REQ-TODO-013 restated in the doc comment at `:43-47`. AC-TDG-004's freeze assertion is well-founded, and its self-flagged base-tree identity is **adequate mitigation**: a freeze assertion is legitimately base-identical, and `acceptance.md:63` says exactly why it is meaningful (it fails if the fourth-state design is built).
- **Findings restoration is a real assertion, not an aspiration.** READ `todo_export.go:69` and `:81-82` — the export marshals `rec` whole and reports `len(rec.Findings)`, so `findings` is in the serialized bytes. A restore that recovered the card row but dropped its findings *would* change the export bytes and AC-TDG-002 *would* fail.
- **AC-TDG-010's self-flagging is adequate.** `acceptance.md:101` states the pairing requirement with AC-TDG-009 explicitly. The pairing is a suite convention rather than a mechanism — acceptable debt at this scale; noted, not charged.
- **Non-interactivity is properly bound.** READ `internal/cli/todo.go:20-22` — the `SUBAGENT BOUNDARY` discipline is present verbatim. REQ-TDG-012 + AC-TDG-012 + `plan.md:69` cover it, and nothing in the design introduces a prompt path.
- **Acceptance preconditions are standing, not footnotes.** `acceptance.md` §A.1/§A.2 are `[HARD]` at the head of the file and restated at `plan.md:122-134` (§F.2/§F.3). Every non-zero-exit criterion (AC-TDG-008 `:87`, 009 `:93`, 013 `:119`) uses `out=$(cmd 2>&1); rc=$?`; AC-TDG-011 inherits the form from §A.2.
- **Template-First is recorded with both paths.** `plan.md:113-120` (§F.1) names `.claude/skills/moai/workflows/todo.md` and `internal/template/templates/.claude/skills/moai/workflows/todo.md`, requires `make build`, and requires `cmp` parity; AC-TDG-014 asserts both halves. RAN `cmp` on the pair → clean, and `wc -c` → **13709** bytes each, exactly as claimed.
- **Scope discipline holds.** `spec.md:177-199` excludes `drop`/`undrop`, `next`/`unpick`, the persisted landing-state field (t331), flipping `--require-landed` to default-on, the operator-authority `[HARD]` doctrine, `moai todo pr`, and any schema bump — each with a stated reason. No creep found.
- **Tier M is justified.** Nine files at `plan.md:11-23`, none constitutional, 400-700 LOC estimated; 14 REQ and 14 AC both under the ceiling of 16. Resolving D2 and D4 may add a file or two and the classification still holds.

---

## Recommendation

FAIL at iteration 1. The SPEC is unusually well-argued and its central storage ruling is correct on measurements I re-verified — the failure is concentrated in the evidence for Decision 2 and in three acceptance criteria written against an architecture the SPEC misdescribes. All the fixes are edits to prose and criteria; none reopens Decision 1.

Fix in this order:

1. **D1** — rewrite `spec.md` §A.4 to state both failure modes under their conditions and reconcile with `plan.md` §C. This is the one that changes what a reader believes.
2. **D3** — restate REQ-TDG-006 at the `BacklogRecord` level, drop the "two live backends" framing, and rewrite AC-TDG-006, AC-TDG-003's legacy arm, and §A.3's legacy basis against what is observable.
3. **D2** — rule on whether `export-json` carries the archive; add the requirement, correct AC-TDG-007, and budget the step in `plan.md` M1 or M5.
4. **D4** — add a requirement + AC for downgrade survival, or an explicit out-of-scope bullet.
5. **D5, D6, D7** — mechanical corrections.
6. **D8** — one added assertion, once D2 is ruled.

Re-audit at iteration 2 is scoped to this enumerated delta plus a regression check; it is not a fresh full audit.
