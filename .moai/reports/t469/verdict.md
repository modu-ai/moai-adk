# Card t469 — Sync-Phase Completion Verdict

Card: t469 — AC-HWD-015 strip-aware mirror amendment (SPEC-ACHWD-STRIP-EXEMPT-001 wrapper over SPEC-HOOK-WIRING-DRIFT-001 v0.4.0)
Branch: `WT-achwd-strip-exempt` · Sync commit: `17b447240`
Date: 2026-09-03

## Claim

1. SPEC-ACHWD-STRIP-EXEMPT-001 is closed (`in-progress → completed`) on a single
   sync commit carrying the 3-phase close.
2. SPEC-HOOK-WIRING-DRIFT-001 is re-closed to `completed` (v0.4.0 amendment
   re-close, the path its `## Amendments` `re_close_path` row declared). No body
   content and no other frontmatter field was touched — the record was already
   written consistently by the amendment commit `820db6cf9`.
3. CHANGELOG carries one `[Unreleased]` Added row for the wrapper SPEC, placed
   directly above the t216 row it amends; no second row for the re-close.
4. The §I.4/AC-HWD-015 strip-aware mirror check exits 0 on the final tree.
5. Both spec lints report 0 errors on the final tree.

## Evidence

All commands run from worktree `t469`, recorded verbatim in
`.moai/reports/t469/sync-evidence/` (exported, tracked path):

| Check | Command | Observed | File |
|---|---|---|---|
| wrapper lint | `go run ./cmd/moai spec lint .moai/specs/SPEC-ACHWD-STRIP-EXEMPT-001/spec.md` | `0 error(s), 4 warning(s)` (1 `ModalityMalformed` on REQ-ASE-001 + 3 `CoverageIncomplete`, exit 0) | `sync-evidence/lint-wrapper.txt` |
| HWD lint | `go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md` | `0 error(s), 18 warning(s)` (4 `ModalityMalformed` + 14 `CoverageIncomplete`, exit 0) — identical composition to the run-phase A5b measurement, both classes pre-existing | `sync-evidence/lint-hwd.txt` |
| mirror check | AC-HWD-015 perl command (program re-expressed as `sync-evidence/t469-mirror-check.pl` — **semantically equivalent, not byte-identical**: whitespace and a trailing `;` differ, all four regex substitution bodies character-identical; the inline one-liner form was refused by the worktree session guard as unverifiable) | no output, `exit=0` (re-observed by the sync-auditor at `a28706cff`; invocation + output + exit recorded in the evidence file) | `sync-evidence/mirror-check.txt` |

Wrapper warning profile note (not smoothed): the wrapper SPEC has its own
warning profile — 1 `ModalityMalformed` (REQ-ASE-001's lead clause breaks the
linter's subject matcher) + 3 `CoverageIncomplete` (the known cross-file
spec.md↔acceptance.md resolution limitation; this Tier S SPEC carries its ACs
inline in spec.md §3). Pre-existing from plan phase, unchanged by sync.

Sync-phase commit contents (`git show --stat`): wrapper `spec.md` (frontmatter
status only), wrapper `progress.md` (§E.4 populated), HWD `spec.md` (frontmatter
status only), `CHANGELOG.md` (one row), plus this report's `sync-evidence/`
artifacts.

## Baseline-attribution

- All three commands measured in this run, on this tree, at the sync-commit
  HEAD. Verbatim outputs at the `sync-evidence/` paths above.
- Run-phase A5 measurements (§E.1 of wrapper progress.md, HEAD `820db6cf9`)
  were re-measured, not carried: lint composition and mirror exit are identical
  across the two trees, as expected — sync touched frontmatter `status:` and
  prose in CHANGELOG/progress.md only, none of it lint- or mirror-relevant.

## Gaps

- ~~sync-auditor review not yet performed~~ — **performed 2026-09-03**; score
  row filled below, full audit section appended at the end of this file.
  | Auditor | Score | Verdict path |
  |---|---|---|
  | sync-auditor | **PASS-WITH-DEBT — 91/100** (harmonic; no blocking defect) | this file, § Sync-Auditor Audit Record |
- ~~`sync_commit_sha` backfill pending~~ — **resolved**: backfill commit
  `85296e031` landed (placeholder `pending-backfill-sync` → `17b447240`,
  diff verified by the auditor).
- No Go packages, tests, `go vet`, or build were in scope: this run's file
  deltas are `.moai/specs/**`, `CHANGELOG.md`, and `.moai/reports/**` only
  (scope statement in wrapper progress.md §E.1).

## Residual-risk

- **Carried debt from SPEC-HOOK-WIRING-DRIFT-001 plan.md §I.5** (recorded in
  wrapper progress.md "Carried Debt"): ① normalization over-absorption at
  parenthetical granularity; ② the internal-date class is NOT normalized — a
  future date-mandated template strip will report a false-FAIL; ③ the
  (ii-bare) absorption gap — a bare forbidden token inserted into the template
  copy passes the mirror check and is closed only by AC-HWD-016's
  template-side neutrality scan.
- **Plan-audit was PASS-WITH-DEBT 0.80** (Tier M boundary value) — the audit
  itself flagged D1–D5, all patched at `8eb5b9102`; the residual is that the
  margin was exactly at threshold, so any future re-audit of §I could tip
  either way without new facts.
- The re-close of SPEC-HOOK-WIRING-DRIFT-001 relies on its own frontmatter
  `status:` flip only; the drift detector will read the re-close commit as the
  wrapper's close subject — the commit subject names the wrapper ID, and the
  amendment's `re_close_path` row is the record tying the two closes together.

---

## Sync-Auditor Audit Record (2026-09-03, independent re-execution)

Auditor: sync-auditor · Tree: worktree `t469`, HEAD `85296e031` (clean) ·
Every command below was executed by the auditor in this run, on this tree,
unless marked "reasoned". Verdict: **PASS-WITH-DEBT — 91/100**. No blocking
defect; the must-pass firewall (Functionality, Security) passes on both
dimensions independently.

### Dimension Scores (harmonic mean = 91)

| Dimension | Weight | Score | Verdict | Evidence (verbatim, this run) |
|---|---|---|---|---|
| Functionality | 40% | 95 | PASS | Criterion command (lines 566-569 of acceptance.md, extracted and run) → no output, exit 0; `.pl` form → no output, exit 0; AC-ASE-001 payload diff exactly the declared A1-A5 set, `git show --stat 820db6cf9` = 4 files, all `.moai/specs/**`; HWD frontmatter at HEAD = 0.4.0/completed/`amendment_of` self-ref; `4d57b3dcf` verified as the 0.3.0 close commit; `a239cf050` confirmed ancestor (`git merge-base` = itself). Dock: AC-ASE-003's parenthetical mispredicts warning composition (D4). |
| Security | 25% | 95 | PASS | Mutant B (bare `(SPEC-AUDIT-FAKE-001 A9)` appended to line 275 of the template copy): mirror check → exit 0 (absorbed, as recorded); neutrality grep → `1` hit; `go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` → `--- FAIL ... class=C1-spec-id-prefix match=SPEC-AUDIT-FAKE-001` — the compensating control is live and fires on the absorbed shape. After restore: mirror exit 0, leak test `ok` exit 0. The check itself is guard-safe (plain interpreter+script, file reads only, no eval/shell-out); the un-normalized internal-date class fails SAFE (false-FAIL direction). Dock: parenthetical-granularity over-absorption remains a real blind spot at three-file scope (carried debt 1). |
| Craft | 20% | 85 | PASS (with findings) | Both lints re-run match recorded profiles exactly: wrapper `0 error(s), 4 warning(s)`, HWD `0 error(s), 18 warning(s)` (4 `ModalityMalformed` @121/132/158/183 + 14 `CoverageIncomplete` — same lines as the recorded output), exit 0 both. Docks: D1 (byte-for-byte overclaim), D2 (0-byte evidence file), D3 (§E.2/§E.3 placeholders in a completed SPEC). |
| Consistency | 15% | 90 | PASS | CHANGELOG: exactly one new row (line 12), directly above the intact t216 (line 14) and t456 (line 16) rows; the row's line-275 claim verified on both copies and its 47/24 class figures re-measured exactly (`DIFF count: 47`, `NOTEMPLATE count: 15`, `pairs with forbidden classes: 24`, via the plan-phase scratch scripts re-run). Drift audit (`moai spec audit --json`): both SPECs carry only `EraAutoDetected` INFO (H-4, V3R6); the repo's only 2 MUST-FIX findings are on unrelated SPECs (CI-FLAKE-SERIES, FULL-SUITE-DOCTRINE — pre-existing). Dock: D5 (wrapper transition rides a commit naming the other SPEC). |

### Checks executed (all on tree `85296e031` unless noted)

1. **Criterion machine check (AC-ASE-002)** — acceptance.md lines 566-569
   extracted verbatim to `/tmp`, run via bash → no output, exit 0;
   `sync-evidence/t469-mirror-check.pl` run on the three files → no output,
   exit 0. Both forms, this run.
2. **Extraction fidelity** — whitespace-normalized diff of the criterion
   one-liner vs the `.pl` body: whitespace-only differences (`my $rc=0` vs
   `my $rc = 0`, `(@ARGV){` vs `(@ARGV) {`, unspaced `open`/list forms,
   trailing `;`); **all four regex bodies character-identical**. Not
   byte-identical → D1.
3. **Mutant A round-trip (dispatched check 2)** — editorial non-token line
   appended to local `hook-independence.md` → `MISMATCH
   .claude/rules/moai/development/hook-independence.md`, exit 1 → restored →
   `git status --porcelain` empty. (An earlier execution of the same mutant
   also observed exit 1 before a mid-run reset; the coordinator restored that
   edit.)
4. **Mutant B (absorbed shape, auditor-added)** — see Security row above;
   also confirms acceptance.md's (ii-bare) record ("neutrality scan observed a
   non-zero SPEC-ID count on the mutant template copy") with the auditor's own
   measurement, closing the ambiguity the plan-summary's "7" attribution left.
5. **AC-ASE-001 (payload fidelity + scope)** — `git diff 820db6cf9^ 820db6cf9`
   per file: acceptance.md = only the AC-HWD-015 block; spec.md = only
   frontmatter (version/status/amendment_of) + HISTORY Amendments sub-section
   + 0.4.0 row + REQ-HWD-013 reword + §F H3; commit stat = 4 files, all under
   `.moai/specs/**`. Payload texts compared against plan.md §I.4 A1-A3 —
   character-identical (A1 HISTORY row, Amendments table, A2 reword +
   blockquote, A3 §F text).
6. **Frontmatter lifecycle (dispatched check 4)** — `git show 4d57b3dcf
   --stat`: the t216 sync close of HWD ("status in-progress -> completed"),
   confirming `prior_completed_sha`; `git merge-base a239cf050 85296e031` =
   `a239cf050` (ancestor confirmed); wrapper status chain draft (`dd4f3d984`)
   → in-progress (`820db6cf9`) → completed (`17b447240`); backfill diff
   `17b447240..85296e031` shows `pending-backfill-sync` → `17b447240`.
7. **No retroactive re-judgment (dispatched check 5)** — covered by check 5's
   diffs: no other AC edited, no other requirement text changed.
8. **CHANGELOG (dispatched check 6)** — see Consistency row above.
9. **Lint (dispatched check 7)** — see Craft row above; composition matches
   the recorded profiles and progress.md §E.1's composition note exactly.
10. **Drift audit (auditor-added)** — `go run ./cmd/moai spec audit --json`:
    both SPECs clean of lifecycle drift (only EraAutoDetected INFO). The
    re-close shape the Residual-risk section worried about did not materialize
    as a drift finding on this tree.

### Defects (all non-blocking; no must-pass dimension failed)

- **D1** [minor] [optional] **[REPAIRED 2026-09-03 — lead adjudication]** verdict.md Evidence table — "program extracted
  byte-for-byte" is a documentation overclaim: the `.pl` is semantically
  identical (regex bodies character-identical, both forms exit 0) but differs
  in whitespace and a trailing `;`. Required fix: reword to "semantically
  identical, whitespace-normalized" or re-extract verbatim. → Evidence-table
  wording lowered to semantic equivalence with the measured delta named.
- **D2** [minor] [optional] **[REPAIRED 2026-09-03 — lead adjudication]** `sync-evidence/mirror-check.txt` is 0 bytes — it
  persists "no output" only; the "exit=0" half of the Observed cell is not in
  the cited file. Required fix: record `exit=0` in the file (or state in
  verdict.md that the exit was observed in-session). → file re-written from a
  fresh auditor re-execution at `a28706cff` (invocation + observed no output +
  `mirror-check-exit=0`; output emptiness independently observed via
  `wc -c` → `0`).
- **D3** [minor] [optional] wrapper progress.md §E.2/§E.3 remain
  `_<pending run-phase>_` in a completed SPEC, and the run-phase A5a/A5b
  evidence lives under §E.1 (the plan-phase section). Era classification is
  unaffected (the literal §E.2/§E.4 headings exist; H-4 matched — measured).
  Required fix: move the A5 outputs to §E.2/§E.3 or annotate the placement.
- **D4** [minor] [optional] AC-ASE-003's parenthetical says "18 pre-existing
  `CoverageIncomplete`" — actual composition is 4 `ModalityMalformed` + 14
  `CoverageIncomplete`. The §E.1 composition note documents the imprecision;
  the criterion's core (0 errors) verified. Required fix: reword the AC
  parenthetical at the next amendment opportunity.
- **D5** [minor] [optional] the wrapper's `draft → in-progress` transition
  rides `820db6cf9`, whose subject names SPEC-HOOK-WIRING-DRIFT-001 — the
  wrapper's own ID appears in no commit subject until the sync commit. The
  close-subject full-ID mandate's concern (drift-detector subject extraction)
  did not materialize on this tree (drift audit clean — measured), so this is
  a hygiene note, not a defect claim against the mandate.

Lead adjudication on this findings list (2026-09-03): D1/D2 repaired in this
record update; D3/D4/D5 record-only — no repair, left as recorded. The
dimension scores and the 91 harmonic verdict reflect the audited sync-commit
state `85296e031`; the repairs do not retroactively re-score it. The Mutant B
compensating-control observation (Security row, check 4) stands as the card's
non-vacuity evidence.

### Reasoned, not executed

- AC-ASE-003 wording imprecision (D4) — classified from the lint output
  comparison, not a separate command.
- §E.2/§E.3 placement (D3) — classified from the Section Map
  (spec-frontmatter-schema.md) against the progress.md content.
- The carried-debt ledger's completeness — accepted from the plan-audit record
  (PASS-WITH-DEBT 0.80, boundary) plus this audit's own mutant verification of
  debt 3's compensating control; debts 1-2 (parenthetical over-absorption,
  internal-date class) were not independently re-mutanted — their shapes are
  re-stated from §I.5 and are consistent with the regex bodies the auditor
  read character-by-character.
- Plan-phase measurements at `@a1d7598ac` (pre-amendment tree) are cited from
  plan-summary.md and not re-measured (the pre-amendment tree no longer
  exists); everything load-bearing for the verdict was re-measured on
  `85296e031`.

### Residual-risk (auditor's own)

- The verdict's 91 rests on a three-file scope; the fleet-wide strip-aware
  invariant is declared out of scope (§F) with 47 differing pairs / 24
  token-bearing — the class hazard is open until the follow-up SPEC or
  doctrine owner picks it up.
- AC-HWD-016's neutrality scan is a criterion, not a gate; the mechanical
  control behind it (`TestTemplateNoInternalContentLeak`) is live (verified),
  but nothing re-runs the mirror check itself on a schedule — it is
  criterion-scoped. If a future template strip lands without the pair being
  re-judged, the drift is silent until the next audit.
