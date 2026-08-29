# SPEC-SYNC-SHA-SLOT-FORMAT-001 — Implementation Plan

Tier S. Milestones are ordered by **decision reversibility**: the two milestones carrying decisions a
reviewer might want changed come first, and the mechanical work is last.

## §A. Context

Card t299. The `sync_commit_sha` slot in `progress.md` §E.4 has no format contract, so prose occupies
it permanently and no signal is raised. The write side repairs it from a closed allowlist that does
not contain the spelling the doctrine prescribes; the read side does not check it at all.

Full measured baseline in `spec.md` §B. Every figure there was measured in this worktree at
`a6bbbf82b`; none was carried from the dispatch.

## §B. Known issues carried in

1. **The dispatch's root-cause framing was sharpened, not overturned.** The dispatch says
   `needsSHABackfill`'s allowlist is closed, which is true. The measurement adds the part that makes
   it a defect rather than a limitation: the doctrine's own prescribed spelling (`pending-backfill-*`,
   `spec-frontmatter-schema.md` §D3) is **absent** from that allowlist. The blind spot is aimed at the
   sanctioned pattern.

2. **A fourth value reader exists.** The dispatch named three (era, drift, epic render). A fourth,
   `internal/epic/status.go:261` `syncShaYAMLPattern`, is a fifth independent regex with its own
   notion of the value — the epic-render path the dispatch named is that regex, and it does not share
   `cleanFieldValue`. It is recorded and deliberately left out of scope (`spec.md` §E).

3. **A live parsing artifact, publishable today.** `moai spec audit --filter-spec SPEC-V3R6-AUDIT-MODEL-PIN-001`
   run in this tree emits `pending-backfill-sync"   # D3 self-reference exemption — …` as the
   `sync_commit_sha` detail: a stray quote plus a prose sentence, because `cleanFieldValue` trims
   quotes from the ends of the whole string and the string ends in the annotation. The §D.1 token rule
   repairs it as a side effect of being correct; no separate work item is created for it.

4. **The `--strict` cost claim was verified rather than accepted.** How, in full, is in §C step 3.
   The *demotion mechanism* verified there was correct at v0.1.0; the *arithmetic* on top of it was
   not, and is corrected at v0.2.0 — see §C step 3's second half.

5. **The lint rule does not repair the motivating case, and the SPEC now says so.** t354's
   `pending-backfill-sync` is an admitted placeholder, so the lint rule is silent on it by design;
   the closer-side inversion (M2) is what repairs it. `spec.md` §A carries the surface-by-question
   table. A run report claiming the lint rule addresses the motivating instance is wrong about its own
   deliverable.

## §C. Pre-flight (already performed; recorded so the run phase does not re-derive)

1. **Frozen baseline.** `BASELINE_SHA = a6bbbf82b`, this worktree's HEAD and `origin/develop` at
   measurement time. Every criterion in `acceptance.md` is decided against this literal, never against
   the moving ref. [HARD] Re-record it here if this branch absorbs mainline.

2. **Corpus measured.** 346 slots / 334 SHA-token / 12 not (`spec.md` §B.2), plus the 28-spelling
   placeholder family (§B.3). Commands are in the table rows; the run phase re-measures rather than
   citing these if it needs a current figure.

3. **The `--strict` cost claim, and exactly how it was verified.** The dispatch asked for the
   demotion question to be answered rather than repeated, and specifically for the finding-attachment
   half. Three separate reads:

   - `lint.go:1134` — `terminalStatusEnum` contains `completed`. Read.
   - `lint.go:218-221` — `demote` is computed **per document** from `doc.Frontmatter.Status` and
     applied to `docFindings`, the batch of every finding that document's rules produced. A rule
     reading a sibling file sets `Finding.File` to the sibling path (`lint_movingref.go:129` is the
     precedent, and it does exactly this) but its finding is still in that batch. So `progress.md`
     carrying no frontmatter status is **not** an escape from demotion: the demotion keys on the
     owning `spec.md`, and a progress.md-derived finding is demoted with it. This is the half the
     dispatch flagged as possibly changing the count; read, it does not.
   - `lint.go:61` — `--strict` escalates a warning only when `!f.Advisory`, and `applyEraDemotion`
     (line 288) sets `Advisory` on every warning of a demoted document.

   Statuses of the 12 violations' owning SPECs were measured by `grep -m1 '^status:'` per SPEC: 9
   `completed`, 2 `implemented`. Era of the two `implemented` ones was measured, not assumed, via
   `mcp__moai__spec_audit` with `project_root` set to this worktree: both return `"era":"V3R6"`,
   neither is grandfathered.

   **The twelve are not the flagged set** — this is the arithmetic step v0.1.0 of this plan skipped,
   and the plan audit's D1. REQ-SSF-005 exempts a recognized placeholder, and seven of the twelve are
   exactly that. Classifying all 346 slots with a classifier implementing §D.1 exactly:

   ```
   python3 .moai/reports/t299/grammar_check.py
     → total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
   ```

   All five flagged slots sit in `completed` SPECs, which `terminalStatusEnum` shelters.

   **Predicted: 5 total findings, 0 non-advisory.** Both `implemented` SPECs hold
   `pending-backfill-sync` and are in the exempt seven — the old prediction nominated precisely the
   two SPECs the rule must stay silent about, which is why the error was invisible from inside the
   derivation. Files and lines are named in `spec.md` §B.5.1. This remains a derivation, not a
   measurement of a rule that does not exist yet; AC-SSF-006 converts it, and its wording forbids
   tuning the rule to hit the number.

4. **Merge-conflict expectation, recorded so it is not diagnosed twice.** Card **t357 M2** also edits
   the `l.rules` registry slice in `internal/spec/lint.go`. A one-line conflict there is **expected**
   and mechanically resolvable — both sides add a rule entry, and the resolution keeps both. This is
   carried from the lead's dispatch and is **not** independently verified here (no t357 SPEC directory
   exists in this tree to read). Treat it as an expectation, not a measurement; AC-SSF-008 is what
   keeps this card's footprint on that file to the one line that makes the resolution trivial.

   t357 M2 also raises `--strict` behavior by promoting warnings to errors. **This affects a decision
   here, not only a number** — the plan audit's D3, and the correction to how v0.1.0 of this plan
   framed it. Two layers are possible and they differ in consequence:

   | t357 promotes at | Consequence for this card |
   |---|---|
   | `Report.HasErrors` | `Finding.Severity` stays `warning`; the shelter holds; §B.6 and REQ-SSF-007 both stand |
   | `Finding.Severity` | the five findings become errors; `eraDemotableCodes` is then the ONLY shelter, and REQ-SSF-007 forbids it — five findings on closed history become hard `--strict` errors |

   **Lane instruction if t357 lands first:** read which layer it promotes at and **stop and report to
   the lead** in the second case. Do not add `SyncSHASlotFormat` to `eraDemotableCodes` on the lane's
   own authority — that would satisfy REQ-SSF-007's inverse without the requirement having been
   changed, and the choice between sheltering the five and keeping the map errors-only is the
   operator's. Do **not** attempt to verify t357 from this tree; no SPEC directory for it exists here,
   and this row is carried from the dispatch as an expectation.

## §D. Constraints

- **No corpus repair.** `git diff a6bbbf82b -- .moai/specs` must show no modification to any
  `sync_commit_sha` line outside this SPEC's own directory (AC-SSF-009's sibling in the DoD).
- **`SPEC-BACKLOG-LOCK-BUDGET-001` (t354) is not touched.** Reasoning in `spec.md` §D.4 and §E.
- **`cleanFieldValue` and era classification unchanged** (REQ-SSF-008, AC-SSF-009).
- **Template-First.** No `.claude/` or `.moai/` distributed file is expected to change. **If** the
  format contract turns out to need a doc edit, `spec-frontmatter-schema.md` §D3 is the target, and
  the local copy and `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md`
  are **byte-identical** at `a6bbbf82b` (verified: `diff` → exit 0, both 24677 bytes) — so both change
  in the same run, followed by `make build`. Note the template mirror is subject to the neutrality
  guard: a SPEC ID or an internal date must not enter the template copy.
- **Verification scope.** `go test ./internal/spec/...` plus any other package the change touches.
  Do **not** run the full local suite; push and let CI decide the whole tree.
- Priority labels only; no time estimates.

## §E. Self-verification

Before returning, the run phase confirms: the triad (AC-SSF-001/002/003) reported **together** with
each stated mutation actually run and its red result recorded; AC-SSF-006's number reported as
measured with any divergence from 2 explained rather than absorbed; `git diff --stat a6bbbf82b -- internal/spec/era.go`
empty; the `l.rules` diff one entry.

## §F. Milestones

### M1 — The grammar and the shared predicate (Priority High)

The decision most likely to be revised, so it lands first and alone.

- Define `isCommitSHAToken(token string) bool` and the token-splitting helper in `internal/spec`
  (new file `internal/spec/syncsha.go`, or alongside the lint rule — placement is the implementer's
  call provided AC-SSF-007's single-definition property holds).
- Grammar per `spec.md` §D.1: first whitespace-delimited run is the token; single leading/trailing
  quote stripped from the **token**, not the line; `[0-9a-fA-F]{7,40}` is a SHA;
  `pending-backfill(-[A-Za-z0-9-]+)?` is the canonical placeholder family.
- Unit tests for the grammar itself, including the two limitations recorded in `spec.md` §D.1 (L1
  hex-word false negative, L2 no reachability) asserted as known behavior rather than left implicit.

**Reviewable decisions in this milestone**: the canonical placeholder spelling; the first-token rule
versus a separator enumeration; the 7-40 length band.

### M2 — Closer write-side inversion (Priority High)

- `needsSHABackfill` becomes `!isCommitSHAToken(token(value))`. The four-value `switch` is removed.
- AC-SSF-004 (t354's failure reproduced) and AC-SSF-005 (legacy four preserved) land here.
- `mx_commit_sha` shares this predicate and inherits the widening. It is measured, not assumed: 85
  slots, 9 non-SHA, of which three are deliberate declarations (`<NA>`, `_<pending Mx-phase>_`, and a
  `_(not applicable …)_` sentence) that a later close would overwrite with `(this commit)`. The
  behavior is **accepted** in `spec.md` §E on a measured basis — all three owning SPECs are already
  `completed`, so the live blast radius is zero. State the inheritance in the run report; do not add
  an `mx` guard, which would be scope this card did not take.

**Reviewable decision**: inverting an enumeration into a positive test changes behavior for every
unenumerated value at once. AC-SSF-005 is the guard that makes the change a widening rather than a
substitution; if it fails, stop rather than adjusting the criterion.

### M3 — Lint read-side rule (Priority Medium)

Mechanical once M1 and M2 are settled. Follows the `MovingRefUnpinnedRule` layout exactly:

- New `internal/spec/lint_syncsha.go` + `internal/spec/lint_syncsha_test.go`.
- Exactly **one** entry added to the `l.rules` slice in `lint.go`, with a comment block recording the
  `warning` severity choice and the deliberate absence from `eraDemotableCodes` (`spec.md` §D.3).
- The rule reads the SPEC's own sibling `progress.md` — `SPECDoc.Body` carries `spec.md` alone, so a
  body-only rule would see none of the corpus.
- Fixtures under `internal/spec/testdata/syncsha/`, all satisfying the era precondition in
  `acceptance.md` §A. The prose fixture uses `TBD (filled post-commit)` specifically, for the reason
  stated there.
- AC-SSF-001, -002, -003, -007, -008 land here.

### M4 — Corpus measurement and report (Priority Medium)

- Run `moai spec lint --strict`, filter to `SyncSHASlotFormat`, record total and non-advisory counts.
- AC-SSF-006 (expected **5 total / 0 non-advisory**). **Report the numbers; do not tune to them.**
- Record **two distinct inventories** in `progress.md` §E.2, because they carry different decisions:
  1. **The five findings** (`spec.md` §B.5.1), per file and line — the observed-not-repaired list
     behind the operator decision on `completed`-SPEC history.
  2. **The seven exempt placeholders** — not findings, and not a defect list. Each is a slot still
     owed a real SHA, so this is the inventory of SPECs whose own close will populate them (t354's
     among them, per `spec.md` §D.4). Useful to the lead for scheduling, never a work item for this
     lane.

## §G. Anti-patterns

- **Enumerating placeholder spellings on the closer side.** The set is open — 28 spellings measured
  today (`spec.md` §B.3) — and every new suffix silently opts a SPEC out of repair. This is the defect
  being fixed; re-introducing it as "just one more case" undoes the card.
- **Repairing the corpus while passing through it.** Twelve values are visible and each looks like a
  one-line fix. Nine belong to closed history and two belong to other open cards.
- **Reporting one leg of the triad.** Any single leg is satisfied by a degenerate rule; a partial
  report reads as evidence and is not.
- **Adjusting the rule until AC-SSF-006 reports 5 / 0.** The prediction is derived, and a divergence
  is information about the derivation. v0.1.0 predicted `12 / 2` from the same derivation with one
  step missing; the lesson is that the number is an output of reasoning that can be wrong, not a
  target.
- **Reading "12 non-SHA values" as "12 findings".** Seven are exempt placeholders. A run reporting 12
  findings has broken AC-SSF-003.
- **Adding `SyncSHASlotFormat` to `eraDemotableCodes`.** That map demotes errors; the entry would be
  inert, and an inert entry in a policy map reads as intent to a later maintainer.
- **Editing `cleanFieldValue` because the token logic "belongs" there.** Era classification depends on
  its current behavior, including this card's own §B.6 prediction.

## §H. Cross-references

- `spec.md` §B — the measured baseline every figure here refers to.
- `acceptance.md` §A — the falsifiability contract and the fixture era precondition.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § SHA placeholder backfill exemption (D3).
- `SPEC-MOVING-REF-GUARD-001` — file layout, `warning` severity, and `eraDemotableCodes` precedent.
