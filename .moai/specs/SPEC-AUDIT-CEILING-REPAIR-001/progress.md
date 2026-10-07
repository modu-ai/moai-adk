# SPEC-AUDIT-CEILING-REPAIR-001 — progress

status: implemented
card: t1560 (lane-2, run tmhxo0) · tree .moai/worktrees/t1560 · branch WT-internal-runtime-audit @ 903ccd028
evidence: this file (§G carries any ceiling records for THIS spec if the engine ever evaluates it) + .moai/reports/t1560/

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: 2026-10-07
비고: plan artifacts (spec.md · plan.md · acceptance.md · decision-index.md · references/t1500-seal-excerpt.md) authored 2026-10-07 by manager-spec (card t1560, Tier M). Revision v0.3.0 (2026-10-07): plan-audit-1 repairs integrated (verdict FAIL 0.81, `.moai/reports/t1560/plan-audit-1.md`) + leader ruling #2 (D4 fold-in). `plan_status: audit-ready` is set after the plan-auditor verdict; iteration 2 re-audits delta-scoped to the fix_scope anchors (mapping in plan.md §H).

## §E.2 Run-phase Evidence

### M1 — the eight reproduction REDs (tree `8233644f0`, code-identical to the RED baseline `903ccd028`)

Authored at M1 against the still-pristine engine per the family convention
("RED is a new test — E8 evidence required"); MP-8 basis restated: no repro
test existed at plan phase by design, so the plan-phase `go test -run
'^Test…$'` selectors matched zero tests (`ok … [no tests to run]`, exit 0) —
the expected plan-phase state, never a RED observation and never a pass.
Every RED below ran on `8233644f0` with only test files added (engine files
untouched); `go build`/`go vet` clean and the AC-ACR-012 sealed-surface
selection green before the tests were authored (baseline at 10:26 KST).

| AC | Test | Command (all with `./internal/runtime/`, `-count=1`) | Verbatim RED output (assertion line) | Exit |
|---|---|---|---|---|
| AC-ACR-001 (D1) | `TestPreviousAuditedSHALatestLegacyRound` | `go test -run '^TestPreviousAuditedSHALatestLegacyRound$' -v` | `audit_counter_review_test.go:240: previous audited SHA "", want sha-rev2 (largest round strictly below the legacy latest)` | 1 |
| AC-ACR-002 (D1) | `TestPreviousAuditedSHAMixedFamilyPrior` | `go test -run '^TestPreviousAuditedSHAMixedFamilyPrior$' -v` | `audit_counter_review_test.go:264: previous audited SHA "", want sha-iter2 (convention prior below the legacy latest)` | 1 |
| AC-ACR-004 (D1 e2e) | `TestEvaluateCeilingLegacyLatestDeltaGranted` | `go test -run '^TestEvaluateCeilingLegacyLatestDeltaGranted$' -v` | `audit_ceiling_test.go:1216: outcome &{Outcome:hold Reasons:[plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: …] Blocked:true Debts:[]} (override false), want nil — the delta round is granted: count 2 reaches the tier ceiling 2 and the eligibility conditions hold` | 1 |
| AC-ACR-005 (D2) | `TestPersistOutcomeDebtAdmitCarriesDebtInventory` | `go test -run '^TestPersistOutcomeDebtAdmitCarriesDebtInventory$' -v` | `audit_ceiling_test.go:1253: §G record carries no debts= token: - 2026-10-07T10:31:23Z SPEC-ACE-ENG-001 ceiling-outcome outcome=debt-admit reasons="plan-audit ceiling reached (round count 1 >= tier ceiling 1); the verdict fails admission on the label alone and debt-admits — findings recorded as PASS-WITH-DEBT debts, entry admitted without a question (REQ-ACE-004)" evidence=…` | 1 |
| AC-ACR-013 (D3) | `TestAppendProgressRecordConcurrentSurvival` | `go test -run '^TestAppendProgressRecordConcurrentSurvival$' -count=3 -v` | run 1/3: `audit_ceiling_test.go:1303: pre-existing progress content lost:` followed by a progress.md holding only records 18/13/08/20 under the §G heading; runs 2/3 failed the same way (run 2's content additionally interleaved a partial `fusal Record` fragment — torn write) | 1 |
| AC-ACR-014 (D4 a-c) | `TestCountAuditRoundsOverflowOwnRound` | `go test -run '^TestCountAuditRoundsOverflowOwnRound$' -v` | `(a) audit_counter_review_test.go:306: count 1, want 2 (the base and the overflow suffix are two rounds)` · `(b) audit_counter_review_test.go:325: count 1, want 2 (the base never merges with iter1)` · `(c) audit_counter_review_test.go:344: count 1 (sources 3), want 3 (1 normal + 2 unparseable)` | 1 |
| AC-ACR-014 (D4 e) | `TestPreviousAuditedSHABaseRoundBaseline` | `go test -run '^TestPreviousAuditedSHABaseRoundBaseline$' -v` | `audit_counter_review_test.go:410: previous audited SHA "", want sha-base (the base is round 0, the previous audited round)` | 1 |
| AC-ACR-015 (C3) | `TestReqACSetsUnchangedReadsAcceptance` | `go test -run '^TestReqACSetsUnchangedReadsAcceptance$' -v` | `audit_ceiling_test.go:1413: an acceptance-only AC id rename verified as unchanged — the delta must be refused (fail-closed)` | 1 |
| AC-ACR-016 (D3) | `TestAppendProgressRecordInsertsAtSectionEnd` | `go test -run '^TestAppendProgressRecordInsertsAtSectionEnd$' -v` | `audit_ceiling_test.go:1340: record did not land inside the §G block:` (the record appended past the following `## §E.2` heading) | 1 |

**AC-ACR-003 keep-green guard (pre-fix run)**: `go test -run
'^TestPreviousAuditedSHAUnparseableLatestStaysEmpty$' -v ./internal/runtime/`
→ `--- PASS: TestPreviousAuditedSHAUnparseableLatestStaysEmpty (0.00s)`,
exit 0. The preserve arms of `TestCountAuditRoundsOverflowOwnRound`
(`bare_base_alone_counts_1_and_stays_latest`,
`explicit_zero_keeps_own_round_parity`) also passed inside the pre-fix RED
batch — only the (a)/(b)/(c) subtests failed.

**Measurement note (arm (e) count parenthetical)**: acceptance.md's arm (e)
RED cell says "Count=2 but the scan skips the base". The mechanically
measured count for the arm-(e) fixture (base + iter1) at `8233644f0` is
**1**, not 2 — the same base/iter1 collapse arm (b) documents as "RED today:
1" (both files parse to round 1 and dedupe into `seen[1]`). The Count=2
figure could not be reproduced on this tree; the RED was re-anchored to the
stated previous-baseline reason (the test asserts the previous-audited-SHA
behavior before the count assertions). The defect class, the fix, and the
green state are unaffected.

Full verbatim outputs: `.moai/state/verify/t1560/m1-red-batch.txt`,
`m1-red-e2e.txt`, `m1-red-concurrent.txt`, `m1-red-base-baseline.txt`
(machine-local scratch, this run).

### M2 — D1 + D4 + C3 fixes, REDs flip green (commit `c049810a2`)

`go test -count=1 -run '<M2 selectors + sealed counter/delta tests>' -v
./internal/runtime/` → exit 0, all PASS:
`TestPreviousAuditedSHALatestLegacyRound`, `TestPreviousAuditedSHAMixedFamilyPrior`,
`TestPreviousAuditedSHAUnparseableLatestStaysEmpty` (guard stays green),
`TestCountAuditRoundsOverflowOwnRound` (all five subtests),
`TestPreviousAuditedSHABaseRoundBaseline`, `TestReqACSetsUnchangedReadsAcceptance`,
`TestEvaluateCeilingLegacyLatestDeltaGranted`, plus the sealed
`TestCountAuditRounds`, `TestCountAuditRoundsDistinctN`,
`TestCountAuditRoundsLegacyPlanAuditNumbered`,
`TestPreviousAuditedSHALegacyPriorRound`,
`TestCountAuditRoundsExactHeaderAttribution`, `TestDeltaGitHelpers` —
`ok github.com/modu-ai/moai-adk/internal/runtime 9.378s`.

### M3 — D2 + D3 fixes, REDs flip green (commit `38272872c`)

`go test -count=1 -run '<persistence selectors>' -v ./internal/runtime/` →
exit 0, all 17 PASS (the D2/D3 repro REDs, the escaping RED, the two
preserve tests, and the pre-existing ceiling/trail/override family).
Concurrency discipline: `go test -count=5 -race -run
'^TestAppendProgressRecordConcurrentSurvival$|^TestAppendProgressRecordInsertsAtSectionEnd$'`
→ 10/10 `--- PASS`, `ok … 1.334s` (judged over repeated runs, not one
green).

### M4 — consistency notes (AC-ACR-010/011) and re-measurement

**AC-ACR-010 (t1500 seal, read-and-note — non-contradiction)**: the
SPEC-local verbatim excerpt `references/t1500-seal-excerpt.md` (§SEAL/§PUSH,
provenance header retained) re-read at run phase against the post-repair
diff. The seal froze card t1500's engine work at `3d7215b72` (22/22 AC +
card-review repairs). Non-contradiction holds: this repair EXTENDS the
dual-family contract the sealed card-review F4 test
(`TestPreviousAuditedSHALegacyPriorRound`, legacy-as-PRIOR) established to
its uncovered face (legacy-as-LATEST, `TestPreviousAuditedSHALatestLegacyRound`)
— the sealed test itself stays green (M2 run above), the seal's resume
points (re-review recording, factory stage path, leader push batch) are
untouched by this diff, and no sealed behavior was rewritten (the
heading-absent and §G-last record shapes are byte-identical to the sealed
append; only the §G-followed-by-a-section insertion position and the
base/overflow round identity changed, per REQ-ACR-008/009).

**AC-ACR-011 (t1538 sealed resume point, read-and-note — non-contradiction)**:
`git show origin/WT-t1538-factory-recovery:.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md`
§봉인 기록 re-read at run phase: the remaining-gate inventory is (1) the
factory mirror-path P1 (dispatch store + binding update in one lock
section) and (2) the `^TestReview` overlay reproduction family. Both live
in factory dispatch / review-gate code — disjoint from
`internal/runtime/audit_ceiling.go` (E5 scope grep below confirms this
diff's only engine files are the four §C files).

**Re-measurement (this run, this tree, HEAD `38272872c` + M4 docs)**:

- E3 affected-package family: `go test -race -count=1 -timeout 30m
  ./internal/runtime/...` → `ok github.com/modu-ai/moai-adk/internal/runtime
  12.753s` + `ok … internal/runtime/gobin 1.415s` (full package, race on).
  `go test -count=1 -timeout 30m ./internal/auditverdict/...` → `ok … 0.289s`
  (untouched package stays green).
- AC-ACR-012's named sealed tests all pass inside the full-suite run and
  passed an explicit named selection at M2 (list above).
- Consolidated green run of all thirteen repair tests: `go test -count=1
  -run '<13 repair selectors>' -v ./internal/runtime/` → exit 0, 13/13
  `--- PASS` (`.moai/state/verify/t1560/m4-green-all.txt`).
- E4 lint/format: `go vet ./internal/runtime/... ./internal/auditverdict/...`
  clean; `golangci-lint run ./internal/runtime/...` → `0 issues.`; `gofmt -l`
  on both packages → empty.
- E5 scope grep: `git diff --name-only 903ccd028..HEAD` under `internal/`
  → exactly `internal/runtime/audit_ceiling.go`,
  `audit_ceiling_test.go`, `audit_counter.go`,
  `audit_counter_review_test.go` — no `auditverdict`, `DeltaEligible`, or
  JSON-path (`RecordCeilingOutcome`) changes.
- E6 record grammar spot-check: a no-debt outcome's §G line is built by the
  unchanged Sprintf and carries no suffix when the inventory is empty;
  `TestPersistOutcomeNoDebtRecordUnchanged` verifies the pre-repair grammar
  (`- <ts> <spec> ceiling-outcome outcome=… reasons=… evidence=…`, no
  `debts=`) for pass-through/hold, and the refusal/override shapes, green
  pre- and post-fix.
- @MX tag report: no tag changes — no new exported functions, no new
  goroutines or dangerous patterns (the §G mutex is standard in-process
  serialization per plan §D.7), no fan_in changes on tagged functions.
  Existing tags (EvaluateCeiling ANCHOR, CountAuditRounds NOTE) unchanged.

### F2 repair addendum (sync-audit-1, blocking — umask regression in the D3 atomic replace)

RED observed pre-fix at HEAD `8a8c97d2d` (new test,
`TestAppendProgressRecordNewFileModeAppliesUmask`, `//go:build darwin || linux`,
`syscall.Umask(0o077)`):
`audit_ceiling_umask_test.go:31: new progress.md mode 0644, want 0600 (0644
with the umask applied — the pre-repair os.WriteFile semantics)` — matching
the verdict's `/tmp` probe byte-for-byte. Fix at the CALL SITE
(`appendProgressRecord`): a not-yet-existing progress.md is pre-created
empty through `os.WriteFile(path, nil, 0o644)` — the kernel applies the
umask at create time, the exact pre-repair semantics — and the atomic
replace then stat-preserves that mode; `config/atomicfile.Write` itself is
UNTOUCHED (11 other callers keep their verbatim-defaultMode semantics —
zero caller impact). The existing-file arm
(`TestAppendProgressRecordExistingFileModePreserved`: 0600 stays 0600
through the replace) passed pre- and post-fix.

Mode matrix (new-file progress.md):

| umask | pre-repair `os.WriteFile` 0644 | broken (`atomicfile.Write` 0644 verbatim) | repaired (call-site pre-create) |
|---|---|---|---|
| 0022 | 0644 | 0644 | 0644 |
| 0077 | 0600 | 0644 (the regression) | 0600 |

Existing-file progress.md: mode preserved through the replace, both before
and after (0600 stays 0600) — unchanged arm.

Re-measurement: `go test -race -count=1 -timeout 30m ./internal/runtime/...`
→ `ok … 37.711s`; umask + §G pairs `-count=3 -race` → `ok … 2.615s`; vet
clean; `golangci-lint run ./internal/runtime/...` → `0 issues.`. F1/F3
untouched per leader-pending ruling; F3's aside on `persist.go`
`atomicWrite` (new files stay 0600) is observed and NOT acted on this round.

### F1 repair record (sync-audit-1, leader ruling #2 adopted — legacy-branch Atoi clamp)

RED observed pre-fix: `audit_counter_review_test.go:443: count 2, want 3
(two overflow legacy suffixes fail-count their own rounds + one normal)` —
the discarded Atoi range error clamped both overflow legacy files into
`seen[math.MaxInt]` (one merged round) with one of them elected `LatestPath`
over the legit round. Fix (audit_counter.go legacy branch): the Atoi error
routes to the fail-counted path — own round, never `LatestPath` — the same
REQ-ACR-009 semantics the convention-family overflow fix already applies;
legit small legacy numbers unchanged. Regression test:
`TestCountAuditRoundsLegacyOverflowOwnRound` (two distinct overflow legacy
suffixes + one normal → count 3, `LatestPath` = `-review-1.md`). Whole
touched-function family re-run green (18 tests: TestCountAuditRounds*,
TestPreviousAuditedSHA*, TestRoundReportDirs*, TestCountPlanAuditRounds*,
the D1 end-to-end), full package `go test -race -count=1` → `ok … 27.477s`,
vet clean, golangci-lint `0 issues.`. `evidenceRoundOf` already respected
the range error (M2), so `previousAuditedSHA` was never fail-open here —
the counter and the baseline helper now agree.

### F4 repair record (sync-audit-2, blocking — symlinked progress.md materialized as a regular file)

RED observed pre-fix at HEAD `634cb7c9b`:
`audit_ceiling_symlink_test.go:37: the progress.md symlink was replaced by
a regular file (mode -rw-r--r--)` — os.ReadFile followed the link but the
atomic rename swapped the directory entry, converting the link into a
regular file and materializing the target's content inside the tracked SPEC
directory. Fix at the call site (leader-adopted direction, F2 precedent —
restore the pre-repair write-through posture): `os.Lstat` detects the
symlink, `filepath.EvalSymlinks` resolves it once, and the read + atomic
replace act on the RESOLVED TARGET — the record lands in the target file
and the link survives as a link. A dangling link cannot be written through
and fails closed (best-effort warning upstream, never an admission change).
Product framing (one line, per the verdict's residual-risk note): the leader
chose write-through over fail-closed refusal so dotfiles-managed progress.md
workflows keep working — materialization is blocked either way. Regression
test `TestAppendProgressRecordWritesThroughSymlink`
(`internal/runtime/audit_ceiling_symlink_test.go`, darwin||linux): the
target carries both records, the path is still a symlink, the target lives
outside the SPEC directory.

### F5 repair record (sync-audit-2, blocking — double git failure read as absence)

RED observed pre-fix at HEAD `634cb7c9b`:
`audit_ceiling_test.go:1631: a double git failure (unreadable acceptance
object) admitted the delta — only the clean absence shape may read as
unchanged` — the acceptance arm equated `errA != nil && errB != nil` with
"absent at both ends", but git-show object corruption exits 128 exactly
like path absence, so an unreadable object admitted the delta fail-open.
Fix: the only-absence discriminator moved to `gitPathAbsent`
(`git ls-tree <sha> -- <path>`) — exit 0 with output = present, exit 0
with empty output = cleanly absent (the ONLY shape that may read as
unchanged-empty at both ends), non-zero = object unreadable → fail closed.
The regression fixture surgically corrupts the acceptance.md blob (the
spec.md arm still reads, so the double failure lands in the acceptance
arm): `TestReqACSetsUnchangedAcceptanceCorruptionFailsClosed`. The
Tier S both-absent case still returns true (sealed `TestDeltaGitHelpers`
stays green), and one-end-failure keeps returning false.

### §G next-heading scan fence-fold record (round-2 leader-ruled fold)

RED observed pre-fix at HEAD `634cb7c9b`:
`audit_ceiling_test.go:1651: the line immediately before the next real
heading is "", want the record` — the §G next-heading scan treated a
`## `-prefixed line INSIDE an open fenced code block as a section boundary,
so the record was inserted into the middle of the fence instead of ending
the §G block. Fix: the scan tracks fence state (`opensFence`/`closesFence`
— at least three backticks or tildes, the fence character and run length
deciding which line closes, whitespace-tailed closes allowed) while walking;
a `## ` line inside an open fence is code, not a boundary. Regression test
`TestAppendProgressRecordSkipsHeadingInsideFence`
(`internal/runtime/audit_ceiling_fence_test.go`): backtick and tilde
variants, including a ``` line inside a tilde fence as content — the record
lands immediately before the first real heading after the fence closes.

### Round-3 repair 1 record — fence state on the §G start-heading scan (leader-approved, AC wording syncing in parallel)

RED observed pre-fix at HEAD `ccd1603af`:
`audit_ceiling_fence_test.go:99: the line immediately before the next real
heading is "", want the record` — the scan that FINDS the §G heading
tracked no fence state, so a fenced example early in progress.md carrying
a §G-heading-prefixed line (`## §G Override and Refusal Record (fenced
example)`) was selected as §G and the record landed before the real
section. Fix: the heading lookup walks with the same
`opensFence`/`closesFence` state the next-heading scan already applies —
the real `## §G Override and Refusal Record` is selected. Regression test
`TestAppendProgressRecordStartHeadingSkipsFence`
(`internal/runtime/audit_ceiling_fence_test.go`). The AC wording for this
batch is being updated by manager-spec in parallel; the tests bind to the
behavior.

### Round-3 repair 2 record — round-0 family parity (leader-approved)

RED observed pre-fix at HEAD `ccd1603af`:
`audit_counter_review_test.go:491: previous SHA differs across families:
convention "" legacy "sha-round0", want sha-round0 in both (round-0 family
parity)` — evidenceRoundOf's convention branch filtered on `iterationOf >
0`, so a parsed `plan-audit-0.md` (a VALID round 0) was treated as
unparseable and skipped in the previous-round scan, while the legacy branch
honored `-review-0.md` as a valid round 0. Renaming the same (round 0,
round 2) history between families lost the baseline and could flip an
admitted delta to a final hit. Fix (audit_ceiling.go:321 region): the
convention branch parses its own numeric suffix and treats a parsed 0 as a
valid round 0 — identical round-0 validity and ordering across both
families, matching the sealed planAuditRoundFile ranking (base = iteration
0); the counter's buckets were already at parity (seen[0] both families,
untouched). `iterationOf` lost its last caller (its "1 for the bare
plan-audit.md shape" default is exactly the n=1-collapse mechanism the M2
fix removed) and is deleted — the minimum the fix needs. Regression test
`TestPreviousAuditedSHARoundZeroFamilyParity`
(`internal/runtime/audit_counter_review_test.go`): same history in both
families → same previous SHA and same count.

### Round-3 repair 3 record — ACL preservation on the atomic replace (F6, leader-approved)

RED observed pre-fix at HEAD `ccd1603af`:
`audit_ceiling_acl_test.go:46: the original's ACL did not survive the
atomic replace` — rename(2) swaps the directory entry, so the replacement
file carries the TEMP file's access-control entries (none): a
`group:_guest deny read` ACL line on progress.md vanished through the
replace, while the pre-repair os.WriteFile preserved it (the auditor's
macOS probe, reproduced in-repo). Fix at the call site (shared
`config/atomicfile` helper untouched): the atomic replace now seeds the
temp file from the original with `cp -p` (mode + ACL + xattrs) BEFORE the
new content is written — the content is then truncated over (its mtime
rides the write, so no stale-mtime exposure) and the rename lands a
replacement carrying the original's ACL. Where cp is unavailable
(Windows) or fails, the chmod fallback keeps the F2 mode posture
(existing-file mode preserved, new-file umask-adjusted 0644). Regression
test `TestAppendProgressRecordPreservesACL`
(`internal/runtime/audit_ceiling_acl_test.go`, darwin-only — chmod +a is
a macOS ACL verb): the `deny read` entry survives the record append.

### Round-4 F8 repair record — seed-failure fallback dropped the ACL and bypassed write restrictions (blocking)

RED observed pre-fix at HEAD `6eb6262fb` (PATH-stripped probes driving the
cp-unavailable fallback; verbatim, `zz_red_probe_test.go` — the probe file
is removed at commit and its contract pinned by the landed seam tests):
`zz_red_probe_test.go:25: RED-EXPECTED (pass here means no defect): a 0444
progress.md was rewritten through the fallback` and
`zz_red_probe_test.go:43: RED-EXPECTED (pass here means no defect): a
seeding failure fell back to a mode-only rewrite instead of aborting` — the
fallback restored only the POSIX mode and forced the rename, so (a) the
ACL vanished and (b) a 0444 progress.md was silently rewritten where the
pre-repair os.WriteFile returned permission denied. Fix (both faces, the
mode-only fallback branch deleted): (i) the replace verifies the original
is writable first (`os.OpenFile(path, os.O_WRONLY, 0)`) — a denial is a
clean error, the file untouched, no temp created (os.WriteFile's failure
mode restored); (ii) a seeding failure ABORTS the replace — temp removed,
original untouched, error returned (the best-effort posture holds: an
admission decision is never affected, AC-ACR-009); (iii) the mode is now
applied unconditionally (the F2 posture), never as a substitute for
metadata. Landed tests (seam `seedFileMetadataFn`, deterministic beyond
the closure):
`TestAppendProgressRecordWriteDeniedKeepsFile`,
`TestAppendProgressRecordSeedFailureAborts`
(`internal/runtime/audit_ceiling_replace_test.go`, darwin||linux).

### Round-4 edge 4 record — dangling-symlink write-through (leader-approved)

RED observed pre-fix at HEAD `6eb6262fb`:
`audit_ceiling_symlink_test.go:66: a dangling-symlink progress.md refused
the record the pre-repair write-through would have created: lstat …
progress.md: no such file or directory` — filepath.EvalSymlinks fails for
a link whose target does not exist yet, so no record was written, while
the pre-repair os.WriteFile followed the link and CREATED the target. Fix:
when Lstat confirms a symlink and EvalSymlinks fails, the link's own
referent is read (os.Readlink), its parent directory is resolved via
EvalSymlinks, and the target name joined — the record writes through to
the newly created target and the link survives. A parent that itself does
not exist fails closed (the same ENOENT the in-place write would raise).
Regression test `TestAppendProgressRecordWritesThroughDanglingSymlink`
(`internal/runtime/audit_ceiling_symlink_test.go`, darwin||linux).

### Round-4 edge 5 record — indented fence markers are not fences (leader-approved)

RED observed pre-fix at HEAD `6eb6262fb`:
`audit_ceiling_fence_test.go:140: the line immediately before the next real
heading is "", want the record` — opensFence/closesFence trimmed ALL
leading whitespace, so a backtick line indented 4+ spaces (an INDENTED CODE
BLOCK in Markdown, not a fence) opened a phantom fence whose never-closed
state hid the real §G heading after it, and the record landed at
end-of-file. Fix: a shared `fenceIndent` helper (space counts one column, a
tab advances to the next multiple-of-four column — the CommonMark rule)
bounds BOTH open and close detection to indent 0-3. Regression test
`TestAppendProgressRecordStartHeadingSkipsIndentedCodeBlock`
(`internal/runtime/audit_ceiling_fence_test.go`): the real §G is found and
the record lands inside its block.

### Round-4 gate findings 6+7 record — ACL-guaranteeing seeder and symlink-chain resolution

**Finding 6 (P1, ACL-guaranteeing clone)** — the in-flight clonefileat
seeder DROPPED the original's ACL (measured:
`TestAppendProgressRecordPreservesACL` FAIL on the working tree; a direct
probe showed clonefileat carrying the com.apple.provenance xattr but not
the `group:_guest deny read` entry), and a second factor made every
caller-side chmod an ACL destroyer: on macOS a chmod DELETES the file's
ACL, so the mode must ride the seeding and never follow it. Resolution:
the darwin seeder is restored to cp -p — the mechanism whose replace
provably carries the original's ACL (the control test the finding cited;
`TestAppendProgressRecordPreservesACL` PASS on the working tree) — and the
caller's chmod was removed entirely: the seeder contract is now "make the
temp carry the original's metadata (mode, ACL, xattrs)", per-platform.

**Finding 7 (P2, full symlink-chain resolution)** — RED observed on the
edge-4 working tree:
`audit_ceiling_symlink_test.go:111: …/alias.md was replaced by a regular
file — the chain was not followed to the end` — resolving ONE hop treated
the midlink (alias.md) as the final target and replaced it with a regular
file while the real target stayed empty. Fix: `resolveProgressPath` walks
the CHAIN (each hop's parent resolved through EvalSymlinks, the hop
re-examined until a non-symlink, cycle-guarded, 16-hop bound): a dangling
FINAL referent returns as the path the write-through creates, a cycle or a
missing intermediate directory fails closed. Regression test
`TestAppendProgressRecordWritesThroughSymlinkChain`
(`internal/runtime/audit_ceiling_symlink_test.go`, darwin||linux): both
links survive as symlinks and the record lands in the created final
target.

### Round-4 class closure — PARTIAL; darwin ACL axis is a structured blocker

Landed Go-native: **Linux closes fully** — mode (stat/chmod, applied
before the xattr family because a Linux chmod rewrites the POSIX ACL mask)
plus the extended attributes through x/sys's xattr family, POSIX ACLs
included (they ARE the system.posix_acl_access attribute);
**other platforms** carry the mode through a plain stat/chmod (the no-op
documented posture). **The darwin ACL axis is blocked on the kernel
surface, measured end to end**: clonefileat(2) copies xattrs but NOT
explicit ACLs (probe above); the raw SYS_COPYFILE trap returns EINVAL (the
libc copyfile is userspace on modern macOS); and setattrlist cannot write
the ACL — ATTR_CMN_EXTENDED_SECURITY (0x00400000) is excluded from
ATTR_CMN_SETMASK (0x51C7FF00, SDK sys/attr.h). The only remaining route is
reimplementing the kauth_filesec wire format no syscall accepts — its own
SPEC, not a repair fold-in. **Darwin therefore keeps cp -p** (the
ACL-guaranteeing seeder, finding 6) with the blocker documented in the
source. The three-axis family (`TestAppendProgressRecordPreservesAllMetadataAxes`,
PATH-stripped, umask+ACL+xattr in one seeded file) is AUTHORED and held at
`.moai/state/verify/t1560/held-audit_ceiling_axes_test.go` — it is RED on
darwin for the blocked axis and lands (green, Go-native) only when the
leader adjudicates the darwin axis; landing it red is refused. Leader
decision requested: (i) accept cp -p as the permanent documented darwin
exception, (ii) authorize a follow-up SPEC for the kauth_filesec route, or
(iii) narrow the darwin ACL axis out of the contract.

**Leader ruling received (i)**: the cp -p darwin exception is ACCEPTED —
the source comment now records it as the leader-accepted darwin exception
(kernel surface structurally blocks explicit-ACL copy; kauth_filesec
reimplementation is a follow-up-SPEC candidate). The held three-axis
family stays UNCOMMITTED at
`.moai/state/verify/t1560/held-audit_ceiling_axes_test.go` as follow-up
SPEC material, and the windows no-op posture is documented in
`progress_metadata_other.go`.

### Round-4 gate edges 6b/7b/8/7c record (gate rounds 36-37, leader-approved)

**Edge 6b (UID/GID preservation, `progress_metadata_linux.go`)** — the
seeder contract now includes ownership: `preserveOwnership` chowns the
temp to the original's uid/gid when they differ, and an impossible chown
is an error that aborts the replace (the F8 posture). RED-face honesty:
on darwin the pre-fix seeder (cp -p) already preserves the supplementary
group, so `TestAppendProgressRecordPreservesOwnership` is a preserve arm
on this machine (observed PASS pre-fix); the RED face lives in the linux
seeder (stat/chmod/xattrs copied no ownership), which this darwin lane
cannot execute — flagged as an unmeasured Gap for CI linux.

**Edge 7b/8 (Markdown-context fence detection, gate rounds 36-37)** — RED
observed pre-fix at HEAD `30dc284c2` for gate input A:
`audit_ceiling_fence_test.go:141: §G heading appears 2 times, want 1 (no
duplicate section)` — a line whose backtick fence info string contains a
backtick (and the round-37 shape, inline ```example```) was misread as an
UNCLOSED fence opener, hiding the real §G behind a phantom and
duplicating the section at end-of-file (Goldmark-verified by the gate).
Fix per the stated Markdown rules: a BACKTICK fence opener whose info
string contains a backtick is rejected (inline code, not a fence; tilde
fences unaffected); the closing fence's char+length matching was already
in place (≥ opener run, same character, whitespace tail) and the 0-3
indent window stands. Regression test
`TestAppendProgressRecordGateInputsSingleSectionG`
(`internal/runtime/audit_ceiling_fence_test.go`): gate inputs A, the
round-37 inline shape, and the closed list-item fence (a pin arm that was
green pre-fix) each leave exactly one §G with the record at its block
end.

**Edge 7c (hardlinked progress.md)** — RED observed pre-fix at HEAD
`30dc284c2` (the hardlinked mirror stopped seeing the record and
`os.SameFile` went false after the replace). Resolution per the
coordinator's framing: the IN-PLACE option is adopted — it reproduces
os.WriteFile semantics, which preserve the link relationship. When
`hardLinked` reports nlink > 1 the append writes in place through the
shared inode (truncate+write; the §G mutex keeps concurrent writers
serialized and the write-denial check enforces the permission posture);
the trade for the hardlinked shape is crash-atomicity, documented at the
branch. Platform note: nlink detection is darwin/linux (Stat_t); other
platforms keep the rename path (hardlinked progress.md unsupported
there). Regression test `TestAppendProgressRecordPreservesHardlink`
(`internal/runtime/audit_ceiling_replace_test.go`, darwin||linux): both
paths show the record, `os.SameFile` holds.

### Consolidated repair (sync-audit-5 + gate rounds 36-38) — six items, one pass

**Item 1 — F9 (High, linux default-ACL inheritance)**: the linux seeder
copied the original's xattrs but never REMOVED the temp's INHERITED
default ACL — an original with only the minimal ACL (mode bits, which
listxattr does not enumerate) replaced inside a permissive-default-ACL
directory came out WIDER than the original (auditor/codex probe: UID 65534
denied before, allowed after). Fix: the linux seeder ALWAYS writes the
original's effective access ACL — the copied extended ACL when
system.posix_acl_access exists, otherwise `setMinimalAcl` constructing the
minimal 3-entry ACL (user_obj/group_obj/other from the mode, kernel binary
form, little-endian, ACL_UNDEFINED_ID ids) via Setxattr, which REPLACES
the inherited value. Regression test
`TestAppendProgressRecordOverwritesInheritedDefaultAcl`
(`progress_metadata_linux_test.go`, //go:build linux): default-ACL
directory fixture (default blob granting other rw via Setxattr), original
mode 0640 with no extended ACL → the replaced file's access ACL is the
28-byte minimal blob with other-perm 0. **CI-linux-owned: this test
cannot execute on the darwin lane; GOOS=linux go vet compiles it clean,
and the decisive red/green run belongs to CI linux.**

**Item 6 — metadata contract sentence (leader+auditor aligned)**: "the
temp file must carry ONLY the original's metadata" — one mechanism, not
axis-by-axis. This pass closes the F2→F6→F8→F9 axis-hunting sequence: the
F8 abort posture, the F6 darwin cp -p exception (leader ruling i), and
the F9 full-overwrite seeding are one contract; the seeder is its single
mechanism. Noted for SPEC wording — the lane routes SPEC-text changes
through manager-spec if the auditor requires more than this progress
note.

**Item 2 — darwin inherited-ACL over-grant (inverse face of F6)**: RED
observed pre-fix at HEAD `e54ef43ab`: `the replaced file GAINED the
parent's inherited ACL entry the original lacked` (probe: a parent with
`group:_guest allow read,file_inherit`; the temp was created inside it and
cp -p, whose source had no ACL, left the inherited entry in place). Fix
(within the leader-accepted cp -p exception): the darwin seeder strips the
temp's ACL FIRST (chmod -N — the temp is still empty) and THEN copies the
original's data, mode, and ACL, so the replaced file's ACL is EXACTLY the
original's — inherited entries cannot survive and explicit entries are
applied after the strip. Regression test
`TestAppendProgressRecordAclExactlyOriginal`
(`audit_ceiling_acl_test.go`, darwin): replaced file carries no
`group:_guest` entry, record lands.

**Item 3 — `hop/../actual.md` resolution order (two gate repros)**: RED
observed pre-fix at HEAD `e54ef43ab`: `the real actual.md does not carry
the record` — filepath.Join pre-cleans `..` BEFORE symlink resolution, so
`hop/../actual.md` (hop → other-dir) selected actual.md in the LINK's own
directory and materialized a stray there while the real file went
untouched. Fix: `resolveProgress` applies the referent COMPONENT-WISE in
filesystem order — the base is the link's parent resolved (relative
referents) or the filesystem root (absolute referents), each named
component resolves through the same walk before any later `..` pops it,
and cycles/depth fail closed (depth cap 8; the vestigial 16-hop loop whose
counter never changed is gone, staticcheck SA4008). Regression test
`TestAppendProgressRecordResolvesDotDotThroughSymlink`
(`audit_ceiling_symlink_test.go`, darwin||linux): the record lands in the
REAL actual.md, no stray materializes, both symlinks survive. (Probe note:
the test fixture itself initially pre-cleaned the referent with
filepath.Join before creating the symlink — the kernel stores the raw
string; the fixture now preserves it.)

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-07
run_commit_sha: 38272872c
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable (card-dedicated worktree, single-writer lane; branch/HEAD re-read before each commit per the staleness rule)
l44_post_push_fetch: pending (push is the leader's batch — lane does not push)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; go vet clean; gofmt clean)
cross_platform_build: not-run-in-lane (CI matrix owns the darwin/windows verdict; `go vet` compiled both engine packages clean locally)
total_run_phase_files: 4 (the plan §C set: audit_ceiling.go, audit_counter.go, audit_counter_review_test.go, audit_ceiling_test.go) + progress.md/spec.md run-phase records
m1_to_m4_commit_strategy: one commit per milestone (M1 932f1bed0 REDs · M2 c049810a2 D1+D4+C3 · M3 38272872c D2+D3 · M4 this record); CI run on the integrated branch owns the repository-wide test verdict — PENDING at report time
비고: AC-ACR-004's tier-ceiling fixture and AC-ACR-015's git fixture measured
slower under the race detector (2-3 s each) — no flake observed across the
repeated concurrency runs. The arm-(e) Count=2 parenthetical in
acceptance.md could not be reproduced (measured 1, see §E.2 measurement
note) — recorded as an observation, not an AC failure.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop owns this section.>_

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete (manager-docs sync-phase deliverables done; spec.md transitions to `implemented` on this sync commit — `completed` held for its own gate per the dispatching leader's instruction, card t1560)
- sync_complete_at: 2026-10-07
- sync_commit_sha: 14a6956a8 (sync commit `docs(SPEC-AUDIT-CEILING-REPAIR-001): sync-phase CHANGELOG entry, implemented transition, §E.4 signal (card t1560)`; backfilled in this follow-up commit — the sync commit could not cite its own hash, per the D3 backfill window, spec-frontmatter-schema § SHA placeholder backfill exemption)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Fixed` — first entry (SPEC-AUDIT-CEILING-REPAIR-001; five repair behaviors: D1 legacy-family latest delta round, D2 debt inventory in §G/trail records, D3 atomic §G append + insertion position, D4 round counting base/overflow identity, REQ-ACR-010 both-definition-files delta gate)
- frontmatter_status_transitions:
  - spec.md: in-progress → implemented (this sync commit)
  - plan.md: no `status:` field (Artifact Statelessness) — `updated:` confirmed 2026-10-07, no byte change needed
  - acceptance.md: no `status:` field (Artifact Statelessness) — `updated:` confirmed 2026-10-07, no byte change needed
  - progress.md: status line in-progress → implemented (this sync commit)
- updated_field_refresh: 2026-10-07 (spec.md/plan.md/acceptance.md already dated 2026-10-07 from plan/run phase — confirmed present, no byte change)
- b12_self_test_a (pre-emission grep): `grep -c 'SPEC-AUDIT-CEILING-REPAIR-001' CHANGELOG.md` = 0 (pre-emission, exit 1) → no duplicate entry, emission safe
- b12_self_test_b (AC count match): ac_source=acceptance.md (tier: M); counter live=18 — 16 declared criteria AC-ACR-001..016 (`grep -c '^- \*\*AC-ACR-'` = 16 matrix rows) + 2 prose-example tokens `AC-R-001`/`AC-R-002` at acceptance.md:218 inside AC-ACR-010's fixture prose (example identifier shapes, declare no criterion, ownerless); CHANGELOG entry cites "16 acceptance criteria AC-ACR-001..016" matching the declared set; no AMBIGUOUS halt (both extra tokens uniformly unmarked → live by the mechanical rule, excluded from the cited count by inspection); `[REF]` marker placement is an acceptance.md body edit — manager-spec's surface, recorded as observation, not performed here
- b12_self_test_c (file path verification): paths claimed in the CHANGELOG entry verified — `.moai/specs/SPEC-AUDIT-CEILING-REPAIR-001/spec.md`, `.moai/specs/SPEC-AUDIT-CEILING-REPAIR-001/progress.md` (§E.2 evidence section present)
- canary_compliance_check: N/A (this SPEC defines no forward-looking policy with its own sync tests)
- mx_tag_validation: sync sub-step — no tag changes (§E.2 M4: no new exported functions, no new dangerous patterns; EvaluateCeiling ANCHOR + CountAuditRounds NOTE unchanged)
- 비고: progress.md carries a duplicate empty `## §E.3` placeholder below the filled §E.3 (lines 162-164) — §E.2/§E.3 are manager-develop's surface, left untouched per ownership; era classification reads literal heading presence so the duplicate is inert for lint/audit.
