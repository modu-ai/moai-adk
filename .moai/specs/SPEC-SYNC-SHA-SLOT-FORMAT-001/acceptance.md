# SPEC-SYNC-SHA-SLOT-FORMAT-001 — Acceptance Criteria

## §A. Falsifiability contract

[HARD] Every criterion below names two things: **the command that decides it**, and **the input that
makes it fail**. A criterion with no stated falsifying input is not accepted into this SPEC.

The reason is specific to this card. The deliverable is a guard, and a guard whose criterion cannot
fail is indistinguishable from a guard that is switched off — the criterion passes, the report reads
green, and nothing is being detected. The planted mutation is stated up front rather than left for an
auditor to discover.

### The regression triad — all three, none sufficient alone

AC-SSF-001, AC-SSF-002, and AC-SSF-003 are one instrument in three parts, and the run phase reports
them together or reports none of them:

- **(a) fires on prose** — alone, satisfied by a rule that flags everything.
- **(b) silent on a well-formed SHA** — alone, satisfied by a rule that flags nothing.
- **(c) silent on the sanctioned mid-backfill state** — alone, also satisfied by a rule that flags
  nothing, and the one the D3 window depends on.

(a) with (b) still permits a rule that has quietly closed the backfill window. (b) with (c) still
permits a switched-off rule. Only all three together separate a working guard from every degenerate
one, which is why partial reporting of this triad is treated as no report.

### [HARD] Fixture era precondition

Fixture SPEC directories live under `internal/spec/testdata/syncsha/`. Each is a minimal,
schema-valid SPEC directory whose `sync_commit_sha` line is the only thing that varies.

Every fixture MUST classify as era **V3R6** and MUST carry a **non-terminal** `spec.md` status
(`in-progress`). This is not cosmetic: `internal/spec/lint.go:220` demotes every warning on a
grandfathered or terminal-status document to advisory, and `--strict` escalates only non-advisory
warnings, so a mis-built fixture makes AC-SSF-006's `--strict` half fail for a reason unrelated to the
rule under test.

**A specific trap this precondition contains.** Era H-4 requires a *non-empty* `sync_commit_sha` as
read through `cleanFieldValue`, which maps `pending`, `<pending>`, `null`, `none`, and `tbd` to the
empty string. A fixture using one of those as its "prose" value classifies **V3R5** under H-3 and is
demoted — the criterion would then pass or fail for the wrong reason. The prose fixture therefore uses
`TBD (filled post-commit)`, taken verbatim from the corpus, which `cleanFieldValue` leaves non-empty.

## §B. Baseline anchor

All criteria are decided against the frozen literal `a6bbbf82b` recorded in `plan.md` §C, never
against `origin/develop` directly. A criterion naming a moving ref would make this SPEC an instance of
the defect `SPEC-MOVING-REF-GUARD-001` exists to catch.

## §C. Criteria

### AC-SSF-001 — a prose value is FLAGGED (MUST) — triad (a)

**Given** fixture `internal/spec/testdata/syncsha/prose/` whose `progress.md` carries
`sync_commit_sha: TBD (filled post-commit)`,
**when** `go test ./internal/spec/ -run TestSyncSHASlot_FlagsProse` runs,
**then** exactly one `SyncSHASlotFormat` finding is reported, at `progress.md` and that line, at
`warning` severity.

**Fails when:** the rule is not registered in `l.rules`, or the token split reads the whole line
instead of the first whitespace-delimited run (in which case every annotated value becomes a finding
and this fixture is no longer distinguishable).
**Mutation that must turn it red:** replace the fixture's value with `a6bbbf82b` — the finding count
must drop to 0. A criterion that stays green under this mutation is measuring nothing.

### AC-SSF-002 — a well-formed SHA is NOT flagged (MUST) — triad (b)

**Given** fixtures `internal/spec/testdata/syncsha/sha-short/` (`sync_commit_sha: a6bbbf82b`),
`sha-full/` (40 hex characters), `sha-quoted/` (`sync_commit_sha: "a6bbbf82b"`), and
`sha-annotated/` (`sync_commit_sha: a6bbbf82b   # backfilled in the following commit`),
**when** `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnSHA` runs,
**then** **0** `SyncSHASlotFormat` findings are reported across all four.

**The four fixtures are not redundant.** Each isolates one clause of the §D.1 grammar: length
tolerance, quote stripping on the token, and annotation tolerance. A single bare-SHA fixture would
stay green under a rule that rejects quotes or rejects annotations, and §B.2 measured **99 of the 334
conforming** corpus values carrying an annotation (105 of all 346 values) — so annotation intolerance
would flag nearly a third of a healthy corpus while this criterion reported success.

**Fails when:** the hex pattern is anchored to end-of-line (kills `sha-annotated`), or quote stripping
is applied to the whole line rather than the token (kills `sha-quoted`).
**Mutation that must turn it red:** narrow the SHA pattern to `{40}` — `sha-short` and `sha-quoted`
must produce findings.

### AC-SSF-003 — the sanctioned mid-backfill state is NOT flagged (MUST) — triad (c)

**Given** fixtures `internal/spec/testdata/syncsha/placeholder/` (`sync_commit_sha: pending-backfill`)
and `placeholder-suffixed/` (`sync_commit_sha: pending-backfill-sync   # D3 self-reference exemption`),
**when** `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnPlaceholder` runs,
**then** **0** `SyncSHASlotFormat` findings are reported for either.

This is the criterion that keeps the D3 backfill window open. A rule failing it forbids the only
procedure by which the field can ever be populated, and it would do so while AC-SSF-001 and AC-SSF-002
both still read green.

**Fails when:** the placeholder branch is omitted from the lint predicate, or the suffixed family is
not admitted (which would flag the 24 corpus occurrences of `pending-backfill-sync` alone).
**Mutation that must turn it red:** delete the placeholder branch from the lint predicate — both
fixtures must produce findings.

### AC-SSF-004 — the closer recognizes an out-of-allowlist placeholder (MUST)

**Given** a SPEC state whose `sync_commit_sha` is `pending-backfill-sync` — the exact value carried by
`SPEC-BACKLOG-LOCK-BUDGET-001`, and the value the current four-entry allowlist mistakes for a
populated SHA,
**when** `go test ./internal/spec/ -run TestNeedsSHABackfill_OutOfAllowlistPlaceholder` runs,
**then** the backfill predicate returns `true`.

This is card t354's failure reproduced as a test rather than described.

**Fails when:** the closer's decision is still an enumerated allowlist by any spelling.
**Mutation that must turn it red:** restore the four-value `switch` — the assertion must fail.

### AC-SSF-005 — the closer's widening is strict (MUST)

**Given** the four values of the retired allowlist — `""`, `(this commit)`, `(pending)`, `<pending>` —
**when** `go test ./internal/spec/ -run TestNeedsSHABackfill_LegacySetPreserved` runs,
**then** the predicate returns `true` for every one of them.

A widening that silently drops a previously-caught case is a regression wearing a fix's clothes. This
criterion is what makes "strictly a widening" (§D.2) a measurement rather than a claim.

**Fails when:** the token split mishandles `(this commit)`, which contains a space and whose first
token is `(this`.
**Mutation that must turn it red:** make `isCommitSHAToken` return `true` for the empty string — the
`""` case must fail.

### AC-SSF-006 — corpus enforcement cost matches the §B.6 prediction (MUST)

**Given** the live corpus at `a6bbbf82b`,
**when** `go run ./cmd/moai spec lint --strict` runs in this worktree and its findings are filtered to
`SyncSHASlotFormat`,
**then** **5** findings are reported in total, of which **0** are non-advisory, and this rule
contributes **nothing** to the `--strict` exit status.

The five are named per file and line in `spec.md` §B.5.1 and are asserted individually, not only as a
count: `SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md:47` and `:259`,
`SPEC-V3R6-SPEC-ID-VALIDATION-001/progress.md:103`,
`SPEC-V3R6-SPEC-LINT-CLEANUP-001/progress.md:45`,
`SPEC-V3R6-TEST-REFACTOR-001/progress.md:149`. A count-only assertion would stay green if the rule
found five *different* slots.

**Twelve values, five findings — the other seven are exempt, not missing.** The twelve non-SHA values
of `spec.md` §B.2 include seven recognized placeholders, which REQ-SSF-005 requires the rule to stay
silent about. A run reporting 12 findings has broken AC-SSF-003, not exceeded expectations.

**This criterion reports; it does not tune.** The `5 / 0` is derived from `lint.go:61/220/284/1134`
plus a classifier run over the corpus, not measured against a rule that did not exist at authoring
time. A different number is a **finding to report to the lead**, with the divergence explained, and
never a reason to adjust the rule until the divergence is understood. The most likely honest causes:
the corpus moved between authoring and run, or the finding-attachment behavior differs from the one
§B.6 derives.

**Fails when:** the rule attaches findings to a document whose demotion path differs from the owning
`spec.md` — the non-advisory count would then be **5**, not 0, because the demotion that shelters all
five is keyed on the owning `spec.md`'s `completed` status.
**Mutation that must turn it red:** promote the finding severity to `error` — the non-advisory count
must move from 0 to 5, since `SyncSHASlotFormat` is deliberately absent from `eraDemotableCodes`
(AC-SSF-010) and an error therefore has no shelter.

### AC-SSF-007 — one predicate, two call sites (MUST)

**Given** the implementation,
**when** `grep -rn 'isCommitSHAToken' internal/spec --include='*.go' | grep -v _test` runs,
**then** exactly one definition and at least two non-test call sites are reported, one in
`closer.go` and one in the new lint file.

§B.4 measured four readers already holding three different notions of a value; this criterion is what
stops the count becoming five.

**Fails when:** the lint rule re-implements the hex test locally instead of calling the shared
predicate.
**Mutation that must turn it red:** inline a second copy of the pattern in the lint file — the
call-site grep must show the lint file no longer referencing the predicate.

### AC-SSF-008 — registry addition is one line (SHOULD)

**Given** the diff against `a6bbbf82b`,
**when** `git diff a6bbbf82b -- internal/spec/lint.go` runs,
**then** the change to the `l.rules` slice is a single added rule entry (with its comment block), and
no existing rule entry is reordered or removed.

Following the `MovingRefUnpinnedRule` file-layout precedent keeps the shared-file footprint to one
line, which is what makes the expected collision with card t357's M2 edit to the same slice a
one-line, mechanically resolvable conflict rather than a merge to reason about.

**Fails when:** rule logic is added inline to `lint.go` instead of a dedicated
`internal/spec/lint_syncsha.go`.

### AC-SSF-009 — cleanFieldValue and era classification are untouched (MUST)

**Given** the diff against `a6bbbf82b`,
**when** `git diff --stat a6bbbf82b -- internal/spec/era.go` runs,
**then** the output is empty.

REQ-SSF-008: the `(this commit)`-as-value behavior is load-bearing for era heuristics H-3 and H-4, and
this SPEC's §B.6 prediction itself depends on the current classification of the two `implemented`
SPECs. Changing the normalizer would move an unmeasured set of SPECs between eras.

**Fails when:** the token-splitting logic is implemented by editing `cleanFieldValue` rather than
alongside it.

### AC-SSF-010 — `SyncSHASlotFormat` is absent from `eraDemotableCodes` (MUST) — decides REQ-SSF-007

**Given** the implementation,
**when** `grep -A6 'var eraDemotableCodes' internal/spec/lint.go` runs,
**then** the map contains exactly `MissingExclusions` and `FrontmatterInvalid`, and **no**
`SyncSHASlotFormat` entry.

The map demotes **errors** (`lint.go:284` gates on `f.Severity == SeverityError`), while a warning is
already made advisory by the branch at `lint.go:288`. An entry for a warning-severity code would be
inert, and an inert entry in a policy map reads to a later maintainer as intent that was never meant.

**[HARD] This criterion decides the requirement as written; it does not decide the contingency.**
REQ-SSF-007's justification holds only while the rule's severity is `warning` at the `Finding` level.
If card t357 M2 promotes at that level rather than at `Report.HasErrors`, the map becomes the only
remaining shelter for the five findings on closed history and this criterion would be enforcing the
wrong thing. The lane's instruction in that case is to **stop and report**, not to add the entry and
not to weaken this criterion — see `spec.md` REQ-SSF-007 and `plan.md` §C.4.

**[HARD] The contingency was RESOLVED by observation on 2026-08-29, and the criterion is unchanged.**
The t357 lane reports, through the lead, that t357 M2 sets `SeverityError` directly at its own rule's
emission site with no global warning-to-error promotion — so this rule's `Finding.Severity` stays
`warning`, the shelter holds, and REQ-SSF-007's premise is intact. This is a **reported** observation,
not one verified from this tree (no t357 SPEC directory exists here). The resolution **strengthens**
this criterion's premise: the map entry would still be inert, so its absence is still what the map
should show. It does not relax the criterion, and the "stop and report" instruction above is now
historical.

**Fails when:** an implementer adds the code to the map "for symmetry" with the demotion the rule
already receives.
**Mutation that must turn it red:** add `"SyncSHASlotFormat": true` to the map — the grep must show a
third entry.

## §D. Definition of Done

- [ ] AC-SSF-001, AC-SSF-002, AC-SSF-003 all reported together, each with its stated mutation run and
      its red result recorded (the triad is reported whole or not at all).
- [ ] AC-SSF-004 and AC-SSF-005 green; the legacy four preserved by measurement.
- [ ] AC-SSF-006 reported with its measured numbers, and any divergence from **5 total / 0
      non-advisory** explained rather than absorbed; the five findings named per file and line, not
      only counted.
- [ ] AC-SSF-007, AC-SSF-008, AC-SSF-009, AC-SSF-010 green.
- [ ] If card t357 M2 landed before this card, the layer it promotes at is read and reported, and the
      REQ-SSF-007 contingency (`spec.md` REQ-SSF-007) is surfaced to the lead rather than resolved by
      the lane.
- [ ] `go test ./internal/spec/...` and `go vet ./internal/spec/...` green; no other package's tests
      regressed among those the change can affect.
- [ ] The twelve corpus values are unchanged — `git diff a6bbbf82b -- .moai/specs` shows no
      modification to any `sync_commit_sha` line outside this SPEC's own directory.
