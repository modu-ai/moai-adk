# SPEC Review Report: SPEC-SYNCSHA-BAND-BOUNDARY-001

Iteration: 1/1 (Tier S ceiling = 1 per `harness.plan_audit_tier_ceilings`)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.92** (harmonic mean; Tier S PASS threshold 0.75)

Tree: `.claude/worktrees/t380`, branch `WT-shared-predicate-guard`, HEAD `3f03d9c36`.
Reasoning context ignored per M1 Context Isolation — this audit reads only the four
SPEC artifacts and the repository tree.

---

## Central claim: REPRODUCED under this auditor's own hands

The card rests on one claim — that rule-level boundary fixtures kill the mutant that
survived t299's whole test package. I planted it myself rather than reading the claim
back. Method: created `sha-min7` / `sha-below6` / `sha-above41` from the `sha-short`
template, registered `sha-min7` in `TestSyncSHASlot_SilentOnSHA` and the two
outside-band fixtures in a case-list extension of `TestSyncSHASlot_FlagsProse` (the
plan's own M2 choice), then planted four mutants one at a time at
`internal/spec/lint_syncsha.go:103`, replacing `isCommitSHAToken(token)` with an
inlined `regexp.MustCompile(...).MatchString(token)` and adding the `regexp` import.

Baseline with fixtures, clean source:

```
$ go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	3.201s          (exit 0)
```

| # | Mutant | Direction | Result | Killing fixture |
|---|---|---|---|---|
| 1 | `^[0-9a-fA-F]{8,40}$` | floor narrowed | **RED** | `sha-min7` |
| 2 | `^[0-9a-fA-F]{7,39}$` | ceiling narrowed | **RED** | `sha-full` (existing, reused) |
| 3 | `^[0-9a-fA-F]{6,40}$` | floor widened | **RED** | `sha-below6` (count 1 → 0) |
| 4 | `^[0-9a-fA-F]{7,41}$` | ceiling widened | **RED** | `sha-above41` (count 1 → 0) |

Mutant 1 is the t299 surviving mutant verbatim. Its verbatim FAIL:

```
--- FAIL: TestSyncSHASlot_SilentOnSHA (1.39s)
    lint_syncsha_test.go:131: sha-min7: expected 0 findings on a well-formed SHA, got 1: [{testdata/syncsha/sha-min7/progress.md 9 warning SyncSHASlotFormat `sync_commit_sha` holds "19b6f76", which is neither a commit SHA nor a recognized backfill placeholder, ...}]
FAIL	github.com/modu-ai/moai-adk/internal/spec	3.094s
```

Mutant 3 (the direction a single inside-edge fixture would miss):

```
--- FAIL: TestSyncSHASlot_FlagsProse (0.52s)
    lint_syncsha_test.go:98: sha-below6: expected exactly 1 SyncSHASlotFormat finding, got 0: []
```

**Bidirectionality holds.** §B.1's claim that the 7- and 6-character fixtures catch
opposite directions of the same edge is correct and measured: no band direction stays
green. The pairing argument (inside fixture catches narrowing, outside fixture catches
widening) is sound as written.

**Bonus, unclaimed by the SPEC:** an anchor-dropping mutant
(`` `[0-9a-fA-F]{7,40}` ``, no `^`/`$`) is ALSO killed — by `sha-above41`, and by
nothing in the pre-card corpus. The new fixture set is strictly stronger than the SPEC
claims for it.

Everything was reverted. `git status --porcelain` after revert:

```
?? .moai/reports/t380/
?? .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/
```

`git diff --stat` empty; the three fixture directories removed; the targeted suite green
again (`ok … 2.838s`, exit 0). No commit made.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-SBB-001` … `REQ-SBB-008`, sequential,
  no gaps, no duplicates, uniform zero-padding. Measured:
  `grep -Eoh 'REQ-SBB-[0-9]{3}' spec.md | sort -u` → 8 distinct ids. AC ids
  `AC-SBB-001` … `AC-SBB-009` likewise sequential; `grep -c '^### AC-SBB-'` → 9.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer**
  (`REQ-SBB-*` in `spec.md`), not the verification layer. REQ-SBB-001/002/005/006 are
  Ubiquitous (`The <subject> shall …`, spec.md:113/117/129/134); REQ-SBB-003/004 are
  Event-driven (`When …, the rule shall …`, spec.md:121/125); REQ-SBB-007/008 are
  Unwanted in the GEARS canonical negative form (`The delivery shall not …`,
  spec.md:138/142). The `Given/When/Then` entries in `acceptance.md` are `AC-SBB-*`
  verification-layer criteria and are correctly NOT penalized here (graded under
  Group 4).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with
  correct types (spec.md:2-14): `id`, `title`, `version` (quoted `"0.1.0"`), `status`
  (`draft`, in enum), `created`/`updated` (`2026-08-31`, ISO), `author`, `priority`,
  `phase` (`"v3.2.0"` — a release target, not a lifecycle stage), `module`,
  `lifecycle` (`spec-anchored`), `tags`. Plus optional `tier: S`. No rejected
  snake_case alias present. `plan.md` / `acceptance.md` carry no frontmatter, which
  satisfies the artifact-statelessness rule. See D5 for a note on the `priority` value.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `module:
  internal/spec`), no multi-language tooling claim. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — verb executable, no BLOCKING finding.
  Referenced ids: `SPEC-SYNC-SHA-SLOT-FORMAT-001` →
  `.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/spec.md:5` reads `status: completed`
  (not retired/superseded/archived). `SPEC-SSFA-001` resolves to no directory, which
  would ordinarily be a D7-5 SHOULD — it is a testdata fixture id, declared as such in
  `plan.md` §B and `acceptance.md` §A, so it is a known false positive rather than a
  defect.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall'` over the SPEC
  directory returns nothing (rc=1). D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over
  `plan.md` (present) returns nothing (rc=1). `research.md` is absent, which is correct
  at Tier S. The verb was executable against `plan.md`, so this is a PASS rather than
  an N/A.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 0.75-1.0 | Every requirement single-interpretation; every AC names the exact mutation that must turn it red (acceptance.md:35-38, 50-52, 68-69, 82-83). Minor: AC-SBB-003/004 do not name the test function carrying them — that lives only in plan.md §E M2. |
| Completeness | 0.85 | 0.75-1.0 | All sections present; 6 `### Out of Scope — <topic>` H3 sub-headings each with specific bullets (spec.md:172/183/191/197/209/215); frontmatter 12/12. Gaps: no RED-now four-element cells (D4); the residual non-band rule-level gap is unrecorded (D3). |
| Testability | 0.95 | 1.0 band | Every AC binary-decidable; no weasel words present. AC-SBB-005 declares "Mutation that must turn it red: none" WITH a justification (it measures delivery shape) rather than leaving it silent. AC-SBB-006/007 require the mutant observed red AND reverted AND the verbatim output captured — this auditor executed all four and they decide cleanly. |
| Traceability | 1.00 | 1.0 band | spec.md §D maps all 8 REQ to ACs; every one of AC-SBB-001..009 carries a `Maps REQ-SBB-xxx` line. No orphan AC, no uncovered REQ. |

Harmonic mean = 4 / (1/0.90 + 1/0.85 + 1/0.95 + 1/1.00) = **0.9216**.

---

## Defects Found

**D1.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/plan.md`:L86-93 (M2) — The chosen
home for the outside-band cases is `TestSyncSHASlot_FlagsProse`, but that function is
named in another SPEC's acceptance criterion verbatim:
`.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/acceptance.md:56` reads
`**when** \`go test ./internal/spec/ -run TestSyncSHASlot_FlagsProse\` runs`. Loading it
with `sha-below6` / `sha-above41` means a failure in that function no longer identifies
AC-SSF-001, and the t299 triad doc-block (`lint_syncsha_test.go:6-17`) declares leg (a)
to be one part of an instrument "reported together or not at all". The obvious repair —
rename to `TestSyncSHASlot_FlagsMalformed`, which keeps the `func Test` count at 3 and
so preserves AC-SBB-005 — **falsifies t299's AC-SSF-001 command string**, and
AC-SBB-009 forbids editing that file. The SPEC does not acknowledge this tension. This
is the substantive answer to plan.md §E M2's own open point: "prose" is NOT the right
home, and the constraint that pushed it there (AC-SBB-005's no-new-`func Test` rule) is
what makes the correct home unreachable. — Severity: **major** — Class: **blocking** —
Required fix: surface this to the operator as a three-way choice: (a) add a fourth test
function and relax AC-SBB-005 to "no duplicated assertion body" rather than a `func
Test` count; (b) rename `TestSyncSHASlot_FlagsProse` and accept an amendment to t299's
acceptance.md, which requires relaxing AC-SBB-009; (c) keep the overload and record the
identity loss as declared debt in spec.md §C. Do not let run-phase pick silently.

**D2.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/acceptance.md` (whole file) — Nine
acceptance criteria against a declared `tier: S`, whose acceptance-criterion ceiling is
**8** (`spec-workflow.md` § SPEC Complexity Tier, REQ/AC budget table; the ceilings apply
independently to REQ and AC counts, so the 8 requirements are exactly at ceiling and the
9 criteria are one over). Measured: `grep -c '^### AC-SBB-' acceptance.md` → 9. The rule
states that exceeding a ceiling is a signal to tier up or split, not to relax the
budget. — Severity: **major** — Class: **blocking** — Required fix: fold AC-SBB-005
(delivery shape) and AC-SBB-008 (no production source change) into one criterion — both
are read off the same `git diff --stat` and neither measures runtime behavior — bringing
the count to 8; or re-declare `tier: M` and accept the M artifact set.

**D3.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/spec.md`:L79-91 (§A.2) — §A.2's
diagnosis is correct and general: rule-level coverage is absent because the mutant
*bypasses* the predicate. But that generality is not band-specific, and the card closes
only the band slice of it. Measured in this tree with the full four-fixture set
installed: an alphabet-narrowing rule-level mutant
(`` `^[0-9a-f]{7,40}$` ``, dropping uppercase hex) leaves
`go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1` at `ok … 3.082s`, exit 0 —
every fixture value is lowercase, so nothing notices. The predicate-level table does
carry an uppercase row (`syncsha_test.go:58` `{"A6BBBF82B", true}`), which is exactly
the layer the mutant bypasses. The SPEC does not over-claim (REQ-SBB-006 is scoped to
"BOTH edges" of the band, honestly), but leaving the residual unstated invites the next
reader to conclude the rule-level hole is closed. — Severity: **minor** — Class:
**optional** — Required fix: add one sentence to spec.md §C recording that the
alphabet and anchor clauses of the §D.1 grammar remain uncovered at the rule level,
with the measurement above as its evidence, so the residual is declared rather than
discovered later. (Note: the anchor clause turns out to be covered incidentally by
`sha-above41` — see the bonus finding above — so the residual is alphabet-only.)

**D4.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/acceptance.md`:L28-99 — No criterion
carries a RED-now cell in the four-element form
(`verification-completeness.md` §2.1: command, verbatim stdout, exit code as its own
field, pinned tree SHA). Each criterion names the mutation to plant, which discharges
the §2 mutant-probe obligation, but the pre-implementation observation itself is absent
— AC-SBB-001's "the criterion must report `sha-min7: expected 0 findings … got 1`" is a
prediction, and its only backing is `probe-05-mutant-red.txt`, cited from spec.md §A.3
rather than from the criterion. §2.1's four-element form binds release-blocking
criteria; these are run-phase criteria, so this is a shortfall rather than a violation.
— Severity: **minor** — Class: **optional** — Required fix: at run phase, write the
four elements into progress.md §E.2 per criterion rather than a summary matrix.

**D5.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/spec.md`:L9 — `priority: P2 Medium`
is neither the `P0|P1|P2|P3` form nor the `High|Medium|Low|Critical` form the SSOT
schema allows; it is a hybrid of both. It passes lint because
`internal/spec/lint.go:918` checks presence only (`{"priority", fm.Priority}`), never
the enum, and the same hybrid appears in `sha-short/spec.md` and elsewhere in the
corpus, so this is inherited convention rather than a new mistake. — Severity: **minor**
— Class: **optional** — Required fix: none required for this card; `P2` alone would be
schema-exact if the value is touched for another reason.

**D6.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/spec.md`:L75 — Cites
`lint_syncsha_test.go:124` for the comment that "anticipates a narrowing to `{40}`". The
`{40}` text is on line **123** (`// Mutation that must turn it red: narrow the SHA
pattern to {40} — \`sha-short\``); line 124 is its continuation. Off by one. — Severity:
**minor** — Class: **optional** — Required fix: cite `:123`, or `:123-124` for the span.

**D7.** `.moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/spec.md`:L106-107 and
`progress.md`:L27 and `plan.md`:L42 — Three places state that `git status --short` in
this tree reports only `?? .moai/reports/t380/`. Measured now: it reports two entries,
the second being `?? .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/`. The measurement was
taken before the SPEC directory was authored and is not labelled with that timing, so it
reads as a current-state claim that is false at read time. — Severity: **minor** —
Class: **optional** — Required fix: qualify as "before authoring these artifacts", or
re-measure. `plan.md` §C says "re-confirm at run-phase entry", which will surface it.

---

## Verified-correct claims (no defect)

Recording these because they are the claims most worth doubting, and each survived
re-measurement in this tree at `3f03d9c36`:

- `commitSHATokenPattern` at `syncsha.go:63` = `^[0-9a-fA-F]{7,40}$` — correct.
- `isCommitSHAToken` at `syncsha.go:107`; `SyncSHASlotFormatRule.Check` at
  `lint_syncsha.go:83`; the predicate call at `lint_syncsha.go:103` — all three exact.
- `needsSHABackfill` at `closer.go:421`, body
  `return !isCommitSHAToken(syncSHAValueToken(value))` — exact, matching spec.md §A.
- `TestIsCommitSHAToken_LengthBand` at `syncsha_test.go:52` carrying 7/40/6/41 rows —
  exact.
- Fixture token lengths 9 / 40 / 9 / 9 for `sha-short` / `sha-full` / `sha-quoted` /
  `sha-annotated` — reproduced with the SPEC's own awk command.
- `grep -c '^func Test' internal/spec/lint_syncsha_test.go` → 3 — exact, and my
  reproduction of M2's chosen approach held it at 3.
- `sha-full` already registered in `TestSyncSHASlot_SilentOnSHA` at
  `lint_syncsha_test.go:126` — exact; the reuse-not-duplicate argument is correct.
- **§A.1 / §A.2 do not blur layers.** §A.1's "no fixture sits at the inside edge of the
  floor" is a statement about the rule-level fixture corpus; §A.2's "band coverage
  exists at the predicate level" is about `commitSHATokenPattern`'s own table. §A.2
  states the join explicitly ("Coverage of the band therefore exists at
  `isCommitSHAToken` and is absent at `SyncSHASlotFormatRule.Check`"), and my mutant
  runs confirm the mechanism: mutant 1 leaves `TestIsCommitSHAToken_LengthBand` green
  while turning the rule-level criterion red. The two statements are about different
  layers and the SPEC says so. No defect.
- **Provenance of the quoted t299 figure.** spec.md:42-43 states the 34.439s / exit 0
  figure "is quoted from t299's §E.4; it was not re-measured here", and progress.md
  §"Figures quoted from elsewhere" repeats it. The obligation is discharged.
- **Scope discipline.** No plan.md milestone edits `syncsha.go`, `closer.go`, or the
  rule's logic; M5's mutants are transient with revert mandated per mutant and
  AC-SBB-008 asserting no surviving edit. None of the four concurrent-lane files
  (`internal/template/evidence_citation_guard_test.go`, the two
  `agent-common-protocol*` rules, `manager-lead.md`) appears in any artifact except
  spec.md §C, which excludes them by name.
- **plan.md §G is labelled a prediction, not an observation**, and progress.md repeats
  the label plus "Unmeasured at plan phase". Correct handling.

---

## Answers to the two open points the SPEC flags

**M1 — `sha-below6`'s all-hex value.** Keep it. The card's axis is length, and a
value differing from `sha-min7` only in length is what isolates that axis; introducing
a non-hex character would move two variables and weaken mutant 3 as evidence. An
alphabet case is genuinely uncovered at the rule level (measured — D3), but it belongs
to a different axis and adding it here would silently widen the card's scope. Record
the residual (D3) rather than fixing it.

**M2 — the case-list home.** The chosen option works mechanically (I ran it), but it is
the wrong home for a reason the SPEC did not consider: the function's name is a
load-bearing token in another SPEC's live acceptance criterion. See D1 — this needs an
operator decision, not an implementer's.

---

## Recommendation

PASS-WITH-DEBT. The card's central claim reproduces, its bidirectionality argument is
measured-correct, and no must-pass criterion fails. Two blocking-class items should be
resolved before Implementation Kickoff Approval:

1. **Resolve D1 by operator decision** — the AC-SBB-005 shape constraint and
   AC-SBB-009's freeze on t299's acceptance.md jointly make the semantically correct
   home for the outside-band cases unreachable. Pick (a), (b), or (c) explicitly.
2. **Resolve D2** — bring the criterion count to the Tier S ceiling of 8 by merging
   AC-SBB-005 and AC-SBB-008 (both are read off one `git diff --stat`), or re-tier.

D3-D7 are optional and can be folded into run-phase edits at negligible cost; D3 in
particular is worth one sentence, because leaving the alphabet residual unstated is the
same silence that let t299's debt be misdiagnosed in the first place.
