# SPEC Review Report: SPEC-TODO-SQLITE-001

Iteration: 1/3 (Tier L ceiling = 3, `plan_audit_tier_ceilings`)
Auditor: plan-auditor (independent, fresh context)
Tree audited: `.claude/worktrees/t306` @ `d29b8942e` (branch WT-todo-sqlite)

## Must-Pass Results

- **[FAIL] MP-1 REQ number consistency** — spec.md §B uses REQ-TOSQ-001..006, 010..013, 020..024, 030..031, 040: gaps at 007–009, 014–019, 025–029, 032–039 with no declared reservation policy anywhere in the document. Firewall is letter-mechanical ("no gaps"). Precedent: sibling SPEC-WEB-TODO-QUEUE-001 numbers REQ-WTQ-001..008 strictly sequentially — this repo's modern-SPEC convention IS contiguous, so the decade grouping is indistinguishable from lost-numbering to any downstream consumer. Trivially repairable: renumber 001..N and update acceptance.md §D.2 + cross-refs.
- **[FAIL] MP-2 GEARS format compliance (requirement layer only)** — REQ-TOSQ-022 (spec.md:113–116) carries a When-trigger but its response clause is indicative with no modal: "the database is authoritative; completing the quarantine is a best-effort…" — matches none of the five patterns. REQ-TOSQ-024 (spec.md:122–129): of three operative clauses only clause 2 carries "shall"; clause 1 ("relocated layouts apply uniformly") and clause 3 ("the new directory wins…") are modal-less. Secondary style notes (not independently decisive): REQ-TOSQ-005 passive subject; REQ-TOSQ-006 pronoun "it". Band 16/18 conformant ≈ 0.75, but MP-2 is binary. Layer judgment made against the REQ layer exclusively; Given-When-Then AC entries were NOT penalized here (graded under Group 4 / Testability, per M3 §Scope).
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:2–13: all 12 canonical fields present and typed correctly (`version` quoted semver, ISO dates, enum values valid, `tags` comma-string). No rejected snake_case aliases. `related_specs` is a tolerated extra field. Evidence: field-by-field read at spec.md:L2–L13.
- **[N/A] MP-4 Section 22 language neutrality** — N/A: single-language Go storage SPEC; no multi-language tooling enumeration involved. Auto-passes per MP-4 precedent.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — refs extracted: SPEC-KANBAN-TODO-CLI-001, SPEC-WEB-TODO-QUEUE-001 (frontmatter L14, plan §H). Both directories exist; statuses read `in-progress` and `completed` respectively — neither retired/superseded/archived, no reconciliation required. Verified by direct grep this session.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → 0 hits; D8 auto-PASS (D8-4 branch).
- **[FAIL] MP-7 clarification gate** — `plan.md:29` contains the literal token: `[NEEDS CLARIFICATION: none — version pinning is a mechanical run-time lookup …]`. The gate grep is deliberately non-semantic and fires on ANY match; worse, it violates the SPEC's own SC-4 ("No [NEEDS CLARIFICATION] markers remain in plan/research artifacts at kickoff"). The token is misused as decorative emphasis for a resolved item. Fix: delete the bracket token, keep the plain sentence.

3 must-pass failures ⇒ overall verdict FAIL regardless of scores (M5 firewall), iteration 1.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | minor ambiguity cluster | Wrong cross-ref: spec.md:93 (REQ-TOSQ-012) points statusline latency budget at "§C-4"; the latency budget lives in C-2 (C-4 is engine/binary-size). Pronoun "it" in 006. Decade-gap numbering degrades scan-comparability. Otherwise single-interpretation, measurable language throughout. |
| Completeness | 0.75 | one axis sparse | All sections + 12/12 frontmatter + six concrete Out-of-Scope H3 bullets + History table present. Sparse axis: research.md §3's consumer table omits real production consumers (D5 below) while §6.3 claims a count its own table does not enumerate. |
| Testability | 0.75 | one AC not precisely binary-testable today | 15/16 ACs carry WHEN-to-run + named red input + green path faithfully. AC-014's RED-now is defined counterfactually ("absent new code measures trivially below bar once landed-without-tests"); AC-013 concedes its red is "in the future sense" — honest, but neither observes a red on tree d29b8942e. Weasel words: none found. |
| Traceability | 0.50 | multiple REQs lack ACs | REQ-TOSQ-001 (db-file persistence incl. schema-version/meta), REQ-TOSQ-003 (WAL/busy_timeout pragma configuration), REQ-TOSQ-004 (Load insertion-order preservation) have NO directly-citing AC anywhere in acceptance.md — yet §D.2 asserts "mapping complete". All AC citations point at existing REQs (no orphans). |

Overall Score: 0.69 (simple mean 0.6875). Below Tier L PASS threshold 0.85 — fails on both axes (score AND must-pass firewall), mutually consistent.

## Judgment Calls & Focus Items Disposition

1. **Keeping `backlog.lock` as outer serializer post-cutover (spec.md REQ-TOSQ-011; design §5) — CHALLENGE ANSWERED, KEEP is justified.** (a) Cross-version safety is real, not vestigial: a downgraded binary knows only the file lock; `export-json` writes the legacy filename an old binary may concurrently mutate — SQLite's busy_timeout is invisible to that process, so no shared fence would exist between generations (design §5(b,c)). (b) Mechanical identity with the proven ten-lane factory protocol and every existing concurrency test is the reuse-rung of the simplicity ladder; dropping the lock would force inventing a second protocol for exactly the upgrade window where lock semantics are least tested. (c) Cost ≈ µs, already paid today. Deadlock posture clean: single lock always acquired outermost across SQL transactions (design §4), no inversion. Correctly disclosed in research §8 rather than buried.
2. **Operator non-negotiables** — all six covered: (a) lossless migration incl. findings + crash-window/live-growth tolerance (REQ-020/021/023, design §4 states A–E + windows, AC-001..005); (b) fallback READ (design §3, AC-008 EXDEV-injected seam); (c) rollback sound (quarantine-rename + additive `export-json` + documented downgrade with artifact-meaning table; config-knob rejection argued on divergence-class grounds, and correctly framed as capability-irrelevant; D.4 gate 3 reviews doc accuracy); (d) CLI compat frozen twice over (REQ-010 verb freeze + REQ-013 exported-surface freeze; exported symbol list spot-checked against code — NewBacklogStore:240, BacklogPathForRoot:249, ResolveTodoQueueRoot:64/Adopting:76, RecordPath:186 all exist as claimed); (e) ~200 home-side keys explicit Out of Scope H3 + measured census (200 dirs / 197 queues — re-measured by me, matches); (f) windows release-blocking honored via two-tier evidence doctrine (local vet = compile only, CI job = behavioral verdict).
3. **Rebutted premises did NOT silently narrow requirements**: screens_templ.go hit confirmed comment-only (screen line 1242) so dropping it removes no functional site; literal-count correction went UP in rigor; global side resolved root-parametrically (same lazy rules every root gets — uniform treatment, not exemption); live-growth observation (re-measured: 82 items / last_seq 310 this session) STRENGTHENS REQ-020's shape-tolerance mandate rather than freezing today's census.
4. **M6-last reversibility ordering — accepted.** The genuinely release-blocking unknown (CGO-disabled windows compilation) was retired analytically BEFORE milestone ordering by pinning the pure-Go driver to the measured CI fact (ci.yml GOOS matrix + `CGO_ENABLED: "0"` in both vet and cross-compile steps — verified this session), and M1's own gate already demands cross-vet darwin/linux/windows clean. M6 defers confirmation vetting of an already-selected compatible stack, not a discovery risk; rework-explosion scenario does not obtain.
5. **Statusline budget (C-2)** — accepted as an authored gate, not thin air: the 10 ms/25 ms figures are analyst-set ceilings against a same-≥500-item-fixture baseline harness measured at M0 pre-flight, and AC-012 classifies *missing* measurement as RED ("no evidence = fail"), so no unverified number can enter the evidence record. Plausibility model supplied (design §8 cold-WAL-open argument). Binary-size gate (C-4) likewise owned: measured at M1, delta>+12 MB triggers review, recording location named.
6. **Template neutrality exposure — honored.** Three template skill targets verified at exactly those locations; plan M5 sequences template-edit-first → `make build` → embedded-bundle freshness check (AC-016 strings-test) → local sync; neutrality guard cited; research §3 correctly identifies `moai update` wholesale redeploy making the template edit the durable fix.
7. **Driver facts — independently verified**: no sqlite dependency in go.mod (direct or the mattn hits are isatty/runewidth/colorable — unrelated non-cgo packages); CGO_ENABLED=0 claim verified in ci.yml.

## Defects Found

| ID | Finding | Anchor | Severity | Class | Required fix |
|----|---------|--------|----------|-------|--------------|
| D1 | MP-7 gate-tripping token: `[NEEDS CLARIFICATION: none…]` present; violates own SC-4 | plan.md:29 | critical | blocking | Delete the bracket token; retain the plain resolution sentence |
| D2 | MP-1 gap structure 001..040 with undeclared decade reservation; repo convention is contiguous (sibling WEB-TODO-QUEUE-001) | spec.md §B (L47–151) | major | blocking | Renumber REQ-TOSQ-001..N sequentially; propagate into acceptance.md §D.2 and cross-refs |
| D3 | MP-2: REQ-TOSQ-022 fully modal-less; REQ-TOSQ-024 clauses 1/3 modal-less | spec.md:113–116, 122–129 | major | blocking | Restore shall-form (or convert first clauses into non-normative rationale moved out of the REQ bullet) |
| D4 | Three REQs uncovered despite §D.2 "mapping complete": REQ-001, 003, 004 have no citing AC | acceptance.md §D matrix vs §D.2 | major | blocking-on-reaudit | Add AC rows (schema/meta presence; pragmas asserted at open; ORDER BY seq insertion-order round-trip) or amend D.2 honestly |
| D5 | Consumer inventory incomplete: internal/cli/kanban.go:330 (companions.json) and :368 (leads.json) are real hand-joins into the renamed directory — absent from research §3 table AND plan.md M4 sweep list; research §6.3 says "eight sites" while its own table enumerates six; additionally cli/todo.go:48/:80 carry old-path comment/help-text literals | research.md §3/§6.3; plan.md M4 | major | blocking-on-reaudit | Add both sites (plus the prose/help literals) to §3/M4; E4/AC-015 remains the mechanical backstop but discovery belongs in the plan |
| D6 | Wrong budget cross-reference: statusline latency budget anchored to §C-4 instead of C-2 | spec.md:93 | minor | optional | Point at C-2 |
| D7 | Optional `tier:` field absent from frontmatter (Tier L recovered via backward-compat default; progress.md §E.1 records it) | spec.md L2–13 | minor | optional | Add `tier: L` |

No regression history (iteration 1).

## Recommendation

Fix route for manager-spec (single revision suffices; nothing substantive is structurally wrong):

1. Delete the NEEDS-CLARIFICATION token wrapper at plan.md:29 (D1).
2. Renumber REQs contiguously and update all cross-references (D2) — mechanical.
3. Restore shall-modality in REQ-TOSQ-022/-024 operative clauses (D3).
4. Add three covering ACs (or amend D.2) for REQ-001/003/004 (D4).
5. Extend research §3 / plan M4 with internal/cli/kanban.go companion/lead registry joins + cli/todo.go prose literals (D5).
6. Two one-word citation fixes (D6/D7).

The engineering substance (driver selection, schema, migration state machine, lock-retention decision, rollout gates) survived adversarial scrutiny intact — every ground-truth claim I could re-measure checked out.

## Residual Risks Worth Carrying Into Run-Phase

1. Directory-level atomic rename while another lane holds flock on `backlog.lock` INSIDE the renamed directory: POSIX tolerates fd-surviving rename; Windows LockFileEx semantics across directory rename deserve an explicit M6 test (not currently enumerated in design §9).
2. modernc.org/sqlite version pin happens at run-phase start (correctly deferred); lint-noise mitigation R1 must produce evidence before scoping, never blanket suppression.
3. `-wal`/`-shm` interaction with arbitrary backup/sync tooling beyond the documented downgrade flow (R2 partially covers; operator education stays docs-side).
4. Ten-lane fleet validation at true process scale remains CI/dogfood observational (acceptance §D.3 discloses the -race approximation honestly).

---

# Iteration 2 — Re-audit of Revised Artifacts

Iteration: 2/3 (Tier L ceiling = 3). Scope per Retry Loop Contract: enumerated defect delta from iteration 1 + regression sweep over prior defects. Tree audited: same worktree @ `d29b8942e`; artifacts revised 2026-08-27 18:19–18:26 (mtimes observed).

## Must-Pass Results (fresh greps, not author claims)

- **[PASS] MP-1** — mechanical extraction of `^- REQ-TOSQ-[0-9]+` yields exactly 18 entries sorting contiguously REQ-TOSQ-001..018, zero gaps, zero duplicates. Numbering-convention note added under §B header (spec.md:47–49); History row documents the revision (spec.md:254).
- **[PASS] MP-2 (requirement layer only)** — both decisive offenders repaired: ex-022 → REQ-TOSQ-013 now carries three shall-clauses covering db-authority, best-effort quarantine retry, and failure containment (spec.md:120–125); ex-024 → REQ-TOSQ-015 has every operative clause normative — "shall apply uniformly" / "shall relocate" / "every code path shall resolve … and shall leave … strictly untouched" (spec.md:131–139). REQ-TOSQ-005 restored to active subject ("The store shall enforce id integrity…"); REQ-TOSQ-006 gained an inverted shall-negation second clause. GWT ACs again NOT penalized here.
- **[PASS] MP-3** — 12 canonical fields unchanged and valid; `tier: L` added at spec.md:15 (valid enum value per optional-field schema).
- **[N/A] MP-4** — unchanged single-language scope.
- **[PASS] MP-5** — referenced SPECs SPEC-KANBAN-TODO-CLI-001 / SPEC-WEB-TODO-QUEUE-001 statuses unchanged (in-progress / completed) since iter-1; no new cross-SPEC references introduced by the revision.
- **[PASS] MP-6** — `syscall` count in revised spec.md: 0.
- **[PASS] MP-7** — gate-surface grep over plan.md/research.md returns ZERO matches. Full-artifact scan finds exactly one occurrence: spec.md:194, SC-4's own prohibition text naming the marker — verified self-referential (it states the rule; it does not carry an open question). Author's reading STANDS on both grounds: outside the gate surface AND semantically a rule statement, not a clarification request.

No must-pass failure ⇒ firewall clear.

## Defect-Delta Verification (D1–D7 from iteration 1)

| ID | Status | Evidence |
|----|--------|----------|
| D1 | **RESOLVED** | plan.md:29–30 now reads "Open question status: none remain —…" — token deleted, plain sentence retained |
| D2 | **RESOLVED** | Contiguous 001..018 verified mechanically (above); all four artifacts propagated |
| D3 | **RESOLVED** | REQ-TOSQ-013/-015 fully modal (spec.md:120–139); REQ-TOSQ-005 active subject (72–75) |
| D4 | **RESOLVED** | New AC-TOSQ-017 (cites REQ-001+004: schema/meta presence + insertion-order fidelity) and AC-TOSQ-018 (cites REQ-003: pragma asserts journal_mode=wal, busy_timeout≥5000 per connection-establishment path) at acceptance.md:29–30; independent union of the matrix Verifies column covers exactly {001..018} — no REQ lost citation coverage during renumbering, no orphan AC. §D.2 claim now true |
| D5 | **RESOLVED** | research.md §3 gains dual-pattern methodology note (54–58), four new rows (kanban.go:330 companionRegistryPath, :368 leadRegistryPath, todo.go:48 comment, :80 user-visible help text), plus explicit site-count summary reconciling "8 structural sites + enumerated prose/comment literals" — matches my own iter-1 independent measurement exactly; §6.3 rewritten transparently crediting the dual-pattern correction; plan M4 has dedicated bullets (106–110); E4 grep upgraded to dual-spelling (57–58), mirrored into AC-TOSQ-015 |
| D6 | **RESOLVED** | spec.md:100 points statusline budget at §C-2 |
| D7 | **RESOLVED** | `tier: L` present |

ADVISORY fold-in VERIFIED: plan.md M6 lock-held-directory-rename test (135–141) covers foreign flock/LockFileEx, fallback-read degradation with no data loss/debris assertion, and retry-on-release completion; design.md R6 extended accordingly (LockFileEx-across-rename divergence named). Cross-layer revision sweep clean: design.md §4.6/§6 now cite REQ-TOSQ-012, §9 testing matrix fully renumbered (011/012/014 · 015 · 008 · 007/010 · 006 · new 003 pragma row), research §8 cites REQ-TOSQ-008 — zero stale references to retired numbers anywhere in the six artifacts (grep NEGATIVE for `REQ-TOSQ-01[9]|02x|03x|04x`).

## Corruption Disclosure Spot-Check (author-reported script damage)

NEGATIVE. Four touched files scanned for `<<`, `<>`, `< >`, `> >` artifacts: zero hits. Load-bearing placeholders render correctly in current state: `t<last_seq+1>` (AC-TOSQ-002), `<uuid>.json`, `<key>/…`, `<queue-root>`. Consistent with the stated rewrite-from-authored-source recovery rather than patch-over-corruption.

## New Findings (iteration 2)

| ID | Finding | Anchor | Severity | Class |
|----|---------|--------|----------|-------|
| N1 | Author revision claim partially inaccurate: "REQ-006 pronoun removed" is false — clause 1 still reads "**it** shall map" (antecedent "the store" sits in the immediately preceding When-clause, so reference is unambiguous; clause 2's inverted form "under no such fault shall the store delete or overwrite…" is a genuine improvement). Nonconformity residual is cosmetic, single instance, previously classified secondary/non-decisive — do not re-open as MP-2 material absent ambiguity or testability damage | spec.md:77–82 | MINOR | optional |
| N2 | progress.md §E.1 `plan_complete_at: 2026-08-27T18:05Z` predates the revised artifact mtimes (18:19–18:26); the audit-ready signal was not refreshed after the revision. Cosmetic — skip-eligibility keys on artifact hash and THIS verdict, not that timestamp — but the field misstates when the final artifact state existed | progress.md §E.1 | MINOR | optional |

## Category Scores (iteration 2)

| Dimension | Score | Rubric Band | Change rationale |
|-----------|-------|-------------|------------------|
| Clarity | 1.00 | — | Both concrete defects eliminated (C-4→C-2 fix; contiguous numbering + convention note). Single interpretation throughout; the one remaining pronoun has an in-sentence antecedent and resolves uniquely, failing even the 0.75 band definition ("a reasonable engineer would resolve differently") |
| Completeness | 1.00 | — | Consumer inventory complete and internally consistent (dual-pattern methodology closed the exact gap iter-1 found); all sections/frontmatter/tier complete; History records the revision |
| Testability | 0.75 | unchanged axis | AC-013/AC-014 RED-now cells remain honest-but-unobservable-today (counterfactual/future-sense) — these were never enumerated as blocking defects and were not part of the delta; the 16 other ACs including both newcomers are exemplary binary-testable pairs |
| Traceability | 1.00 | — | Independently recomputed union covers all 18 REQs; no orphan AC; numbering makes every citation exact |

Overall Score: **0.94** (mean 0.9375). Monotonic increase from iter-1 (0.69 → 0.94): no STOP trigger. Above Tier L threshold 0.85 ⇒ skip-eligible (Verdict PASS required, which is met).

## Must-Pass Summary

MP-1 PASS · MP-2 PASS · MP-3 PASS · MP-4 N/A · MP-5 PASS · MP-6 PASS · MP-7 PASS

## Recommendation

Proceed to Implementation Kickoff Approval. Optional polish for whoever next touches these artifacts (both Class optional, neither blocks): align REQ-006 clause 1 wording to "the store shall map" and refresh `plan_complete_at` when artifacts change.

## Verdict

**PASS** (iteration 2, score 0.94, Tier L threshold 0.85 cleared, all seven must-pass criteria green, no regression against any iteration-1 defect).

## Residual Risks (carried forward, updated)

1. ~~Windows LockFileEx-vs-POSIX rename-divergence test~~ CLOSED by this revision (M6 dedicated test asserted by plan.md:135–141 + R6 pointer); its execution evidence remains a run-phase deliverable.
2. modernc.org/sqlite version pin at run-phase start; R1 lint-noise mitigation stays evidence-first (never blanket-suppressed without written reason).
3. `-wal`/`-shm` vs arbitrary backup/sync tooling beyond the documented downgrade flow (R2 docs-side).
4. Ten-lane process-scale contention stays CI/dogfood observational (-race approximation honestly disclosed in acceptance §D.3).
5. Binary-size 62 MB baseline remains release knowledge until C-4's measured-at-M1 figure replaces it.
