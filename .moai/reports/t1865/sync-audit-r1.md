verdict: FAIL
audited_sha: dfd7c40fb24ac7dcaca7540a62226563ffc5c0dc
required_backend: codex fail
convergence_overall: fail

auditor-model: claude-sonnet-5-5

# Sync audit, card t1865, SPEC-UPDATE-OWNERSHIP-REPAIR-001 (Tier L, tdd)

Overall: FAIL. Two things force it. The required codex backend returned `fail`, so a PASS is excluded. Separately, I reproduced one of codex's P1 findings myself and confirmed it by reading the code: the new classification code reads the bytes of an unregistered symlink target outside the project, including a profile-root file. That is a REQ-OWR-005 / card-t1702 hold-scope violation. It is a real defect in the code, not a codex sandbox artifact.

HEAD verified with `git rev-parse HEAD`: `dfd7c40fb24ac7dcaca7540a62226563ffc5c0dc`, which equals the requested `dfd7c40fb2`.

Judging builds:
- My own Go measurements (gofmt, `go test`) were built and run from the tree at HEAD `dfd7c40fb2`. The installed `moai` binary was not used as a measuring tool.
- The `audit-plan` verb and the `audit_multi` MCP call were served by the installed moai build `ee981a932`. The tool reported it as an ancestor of this tree's HEAD, so it is lagging. Its output (the plan and the codex dispatch) describes older code. The codex review read the diff, and the reproduction logs it cited came from tests run against this tree.

## Claim

1. Functionality: most of the 21 AC-OWR are implemented and recorded green. One requirement-bearing path (REQ-OWR-005 / the no-profile-access invariant) is violated by the unregistered-file read. FAIL on the must-pass dimension Security. Functionality is held back by the same defect and by one loss path in REQ-OWR-017.
2. The hard invariants hold: protected files and `internal/profile` are unchanged, `removalEnabled` is false in non-test code, and there are no skipped tests in the card's test files. The profile-root invariant does not hold on the unregistered-symlink path (finding F1).
3. All 12 binding run conditions have a disposition in §E.3. None is silently dropped.
4. The lane's disclosures are honest as far as I checked them. One of them (the same-clock-second prior-manifest loss) is a larger problem than "evidence loss only" (F2).

## Evidence

### Commands run and verbatim output

`git rev-parse HEAD` and `git rev-parse --show-toplevel`:
```
dfd7c40fb24ac7dcaca7540a62226563ffc5c0dc
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865
```
`moai verify audit-plan --project-root <toplevel>` (abridged, quoted): `"config_status": "ok"`, `"model": "codex"`, `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": ["codex"]`. Gates: claude off, codex required, glm off.

Hard invariants:
```
git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
invariant-exit=0
```
`removalEnabled` in non-test code (the grep for non-test `.go` files):
```
internal/cli/update_dryrun_preview.go:79:  // ... The candidate branch runs only while removalEnabled
internal/cli/update_dryrun_preview.go:84:  if removalEnabled && isCommonAssetCleanTarget(f.RelPath) && ...
internal/cli/migrate_ownership.go:161:  // removalEnabled is the removal switch (default false).
internal/cli/migrate_ownership.go:162:  removalEnabled bool
internal/cli/migrate_ownership.go:392:  if !env.removalEnabled {
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
internal/cli/migrate_project_assets.go:205:  removalEnabled: removalEnabled,
```
The only assignment is `var removalEnabled = false`. This was checked by grep, not by running every code path.

Skips in the card's test files (`grep -n "t.Skip\|SkipNow\|testing.Short"`): the only two files matched were `migrate_ownership_helpers_test.go:1008-1009` (`t.Logf("... skipped=%v", first.Skipped)` and `if first.Skipped {`, which is a run-result field and not a test skip) and `migrate_ownership_symlink_test.go:8` (a comment). There are no `t.Skip` calls.

Stray files: `git ls-files -o --exclude-standard` printed nothing before my own report. The non-Go, non-md, non-baseline diff files were also empty.

Format and small packages (TMPDIR=/private/tmp/own-fixture-d1, load average about 42):
```
gofmt-exit=0
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.423s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.917s
```

My own reproduction of codex's P1 on this tree. I ran the codex review test through `-overlay`, so nothing was written into the tree:
```
go test ./internal/cli/ -overlay /private/tmp/t1865-adversarial-review/overlay.json -run '^TestT1865AuditUnregisteredProfileSymlink$' -count=1
--- FAIL: TestT1865AuditUnregisteredProfileSymlink (0.00s)
    zz_adversarial_review_test.go:42: gate=false class=unregistered reason=no-provenance-record external-profile-bytes-read=true
    zz_adversarial_review_test.go:43: unregistered symlink read profile content outside project with removal disabled
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.343s
```
I read the test source. It builds a symlink at `.agents/skills/canary/SKILL.md` pointing into a synthetic `<profiles>/work/skills/canary/SKILL.md`. It calls `classifyOwnership` with `removalEnabled:false`, then compares `v.data` to the profile file's bytes. `v.data` is set from `os.ReadFile(abs)`.

### Code reading that explains it

- `internal/cli/migrate_ownership.go:287-294`: for an unregistered path, `unreadableKind(abs)` runs first. It uses `os.Stat`, which follows symlinks, and only catches FIFO, socket, device and irregular targets. A symlink to a regular file passes and goes on to `classifyUnregistered`.
- `internal/cli/migrate_ownership.go:348-354`: `classifyUnregistered` calls `os.ReadFile(abs)` before any root-boundary test. `EvalSymlinks` plus `withinRootBoundary` appear only later, in `disposition` (about line 386), and only for files that proved all three evidence items.
- `internal/cli/migrate_project_assets.go:297-310`: `filepath.WalkDir` passes a symlink to a file as a non-directory entry and does not follow it. No symlink skip exists before the classify call.
- At the base (`git show cdd72a814:internal/cli/migrate_project_assets.go`), registered files were already read at line 236 before the boundary check at line 291. That registered-path read existed before this card. The unregistered-file read is new in this card, because t1668 never read unregistered files.
- `internal/cli/migrate_ownership_symlink_test.go` covers only a registered symlink-escape. `migrate_ownership_noprofileaccess_test.go` has no profile-pointing symlink fixture (its symlink greps are about root resolution and a file-mode switch).

The codex log (`/private/tmp/t1865-adversarial-review/focused-probes.log`, written by the codex leg, read by me) also records three more FAILs. I re-ran only the symlink one myself.

### audit_multi (required backend gate), quoted from the call result

Call: `mcp__moai__audit_multi` with `project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, `target=baseBranch`, no `gates` argument.
- `per_backend_verdicts[0].backend`: `codex`; `gate`: `required`; `verdict`: `fail`
- `overall_verdict`: `fail`
- `disagreement_flag`: `null`; `participant_count`: 1; `residual_risk_note`: `required-backend FAIL: codex`; `fail_open_backends`: `[]`
- `audit_receipt`: `rcpt-7770ccadc2f78594b050e777`
- `plan_source`: `config`; `gate_unmet`: not present in the returned object
- `build_commit`: `ee981a932`, with `build_lag` reporting the installed binary as an ancestor of HEAD `dfd7c40fb`.

The codex failure is NOT the "no go files to analyze" sandbox failure the brief anticipated. Codex ran Go probes on HEAD `dfd7c40fb2` (it states "additional 4 repro FAIL, 8 pre-existing related tests PASS, manifest and backup packages all PASS") and returned four substantive findings, listed below with my disposition. The gate result is therefore a real `fail` and is not softened here.

Codex findings, with my disposition:
1. P1, `migrate_ownership.go:350`: an unregistered symlink reads profile bytes outside the project. Reproduced by me. Confirmed by code. Accepted as BLOCKING (F1).
2. P1, `update_prior_manifest.go:110`: two runs in the same clock second overwrite the older prior manifest. Reproduced in the codex log (`same-run-dir=true ... old-generation-preserved=false ... prior-gens-after-current-exclusion=0`). The lane disclosed it. I did not re-run it. Disposition in F2.
3. P2, `migrate_ownership.go:372`: a user-root counterpart FIFO makes the verdict hang before the profile-keep gate returns. Codex-reproduced (`blocked-until-FIFO-closed=true`). I did not re-run it. By reading, `disposition` sets `v.counterpart = env.counterpartOf(rel)` before the profile gate, and no kind check covers the counterpart. This is distinct from the lane's disclosed "swapped between check and read" gap. SHOULD-FIX (F3).
4. P2, `migrate_ownership_emptydir.go:197`: an empty managed `resources/` next to a live `SKILL.md` is not pruned. Reproduced by codex. By reading, `pruneSkillDirTree` returns whenever the skill tree holds any file. REQ-OWR-008 (spec.md:69) speaks only of a skill directory that is itself empty, and `grep -n -i "nested\|sibling"` on spec.md and acceptance.md returns nothing. Conservative design choice, not a SPEC violation. OPTIONAL (F4).

## Baseline-attribution

- All command output above was observed in this run against tree HEAD `dfd7c40fb24ac7dcaca7540a62226563ffc5c0dc`, in the worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, with `TMPDIR=/private/tmp/own-fixture-d1`.
- Diff base for the card: `cdd72a814` (`git diff --stat cdd72a814 HEAD`: 54 files, 17803 insertions, 168 deletions).
- The AC matrix (14 release-blocking, 7 regression guard), the binding-condition table, and the lane's scoped-run figures (63 top-level PASS, 0 FAIL/SKIP) were read from `progress.md` §E.3 and §E.4. They are the lane's records and I did not re-observe them. I re-ran only the manifest and backup packages and one overlay test.
- Judging-build coordinates are stated at the top.

## Binding run conditions (12), re-read in §E.3

| Condition | Disposition found | My read |
|---|---|---|
| N4-remainder | disposed, blind spots disclosed | disposed |
| N14 | disposed (chunk 1, newest generation with records decides) | disposed |
| N15 | disposed, narrowed to `TemplateHash`, leader-approved | disposed |
| N16 | open and disclosed, "no profile-root-opening code added" | disposed as disclosed, but F1 shows a card-added route that does read through a user symlink; N16's own claim ("nothing opens or reads a profile root") is not true on that route |
| N17 | disposed (header comments) | disposed |
| N18 | disposed as SPEC wording debt | disposed |
| N19 | disposed | disposed |
| N20 | disposed | disposed |
| N21 | disposed as SPEC wording debt | disposed |
| N22 | disposed | disposed |
| N23 | disposed | disposed |
| N24 | confirmed limit, disclosed | disposed |

All 12 are present. No condition is silently dropped. I did not re-measure each test cell behind the dispositions.

## RED/arrival cells and test honesty

- The release-blocking list (14: AC-OWR-01..07, -11..14, -16, -20, -21) and the regression-guard list (7: AC-OWR-08, -09, -10, -15, -17, -18, -19) match the brief. The matrix text says regression guards were green on arrival and red under mutants, and it states "this classification is a record of fact; the final judgment is the audit's".
- I read the matrix and the dispositions. I did NOT open each A/B/C RED cell body in §E.2 (about 3,000 lines), so "each release-blocking AC was red for the stated reason before green" is accepted on the recorded labels, not re-verified. See Gaps.
- Vacuity probe. The AC-OWR-08 no-profile-access guard (a source scan plus a differential canary) cannot see the unregistered-symlink route, because no fixture places a profile-pointing symlink under a common asset root. This is a mutant that satisfies the guard while violating REQ-OWR-005. It confirms the lane's N16 note that the scan "cannot see" run-time-assembled access, and it adds a second blind spot that does not need an assembled path.
- AC-OWR-18 (c) `SharedRunDirectory` asserts the lossy same-second behaviour as the expected outcome. The test encodes the loss and so cannot catch it. Codex noted that the existing 1.1 second wait in tests hides the collision elsewhere. I did not read those lines myself.

## Score (flat weighted model; harmonic scoring left to the verdict stage)

| Dimension | Score | Verdict | Basis |
|---|---|---|---|
| Functionality (40%) | 78/100 | FAIL | 21 AC recorded green by the lane; F1 and F2 mean REQ-OWR-005 and REQ-OWR-017 are not fully met |
| Security (25%) | 55/100 | FAIL (must-pass) | F1: a card-added read of bytes outside the project through a symlink; reproduced |
| Craft (20%) | 82/100 | PASS | gofmt clean, no skips, two small packages green; package-wide test, coverage, `-race` not run (carried to CI) |
| Consistency (15%) | 85/100 | PASS | matches the file's style; the lane's evidence format is thorough; the SPEC-wording debts are disclosed, bodies untouched |

The Security must-pass dimension fails independently, so the overall verdict is FAIL regardless of the other scores.

## Findings

- **F1 [BLOCKING]** `internal/cli/migrate_ownership.go:348-354` (`classifyUnregistered`), reached from `:287-294` of `classifyOwnershipCore`. An unregistered symlink under a common asset root that targets a regular file outside the project has its target bytes read (`os.ReadFile` follows the link) before any root-boundary check, even with `removalEnabled=false`. A symlink pointing into a profile `skills` or `agents` tree reads a profile-root file, which REQ-OWR-005 forbids and which the card-t1702 hold scope protects. Reproduced by me (output above). The AC-OWR-08 guard has no fixture for it. Required fix: in the unregistered branch, `Lstat` the path first. If it is a symlink, or its `EvalSymlinks` result is not within `rootResolved`, return `unregistered` with keep reason `symlink-escape` (or `non-regular-file`) without calling `os.ReadFile`. Add a test-first case to the no-profile-access test: an unregistered symlink to a profile-shaped regular file must produce `v.data == nil` and a keep reason. Decide whether the registered-file path (pre-existing at the base, `ReadFile` before the boundary check) gets the same ordering; the registered case is not card-added, so it is optional here.
- **F2 [SHOULD-FIX, may be BLOCKING under a strict reading of REQ-OWR-017]** `internal/cli/update_prior_manifest.go` (`ensureRunBackupDir` and `preservePriorManifest`, around line 110), with `defs.BackupTimestampFormat = "20060102_150405"` at one-second resolution. Two updates in the same clock second share a run directory, and the later `manifest.prior.json` replaces the earlier one. REQ-OWR-017 says the lookup "shall not delete, prune or shorten the retention of any run backup". The lane disclosed this as "evidence loss only (verdicts lean toward keeping)", and the lean-toward-keeping direction is correct because the loss makes the repair keep files, not remove them. Codex rates it P1. Whether the overwrite counts as a deletion is a reading of the SPEC that the leader should settle. Required fix: make the per-run backup directory unique (for example a numeric suffix on collision, created with `O_EXCL`/`Mkdir`), or write the prior manifest only when the file is absent. Replace the AC-OWR-18 (c) `SharedRunDirectory` expectation with a no-wait consecutive-run test that expects both generations to survive.
- **F3 [SHOULD-FIX]** `internal/cli/migrate_ownership.go:372` (`disposition`: `v.counterpart = env.counterpartOf(rel)` before the profile-reach gate). A FIFO placed at the user-root counterpart of a registered managed file makes the verdict wait for a writer, though the profile gate would have kept the file anyway. Codex reproduced it. I did not re-run it, so this rests on codex's log and my reading of the ordering. Required fix: kind-check the counterpart path before reading it, or move the profile-keep gate ahead of the counterpart lookup. Test a counterpart FIFO as well as a project FIFO.
- **F4 [OPTIONAL]** `internal/cli/migrate_ownership_emptydir.go:197` (`pruneSkillDirTree`). An empty managed nested directory is left in place when a live sibling file exists elsewhere in the skill tree. REQ-OWR-008 as written concerns only a skill directory that is itself empty, so this is not a SPEC miss. Leave it unless the leader wants nested directories pruned independently.
- **F5 [MINOR]** The SPEC-wording debts (REQ-OWR-006/AC-OWR-10, AC-OWR-21 command path and case (o), AC-OWR-20 (b), AC-OWR-06 (b), AC-OWR-08 glob, AC-OWR-12 referent) are disclosed and the SPEC bodies were not edited. The AC-COLLECTOR-ANCHOR-001 citation makes the B12 counter print 22 against 21 AC-OWR; the CHANGELOG says 21. The disclosures are honest and recorded; carry them forward as debt to whoever next edits the SPEC.
- **F6 [MINOR]** The `update.go` product changes (dry-run takes no update lock; dry-run skips the binary installer) are accepted under REQ-OWR-011 and disclosed. The warning line for a concurrent update is left as follow-up.

## Gaps (not observed)

- The package-wide `internal/cli` test, package coverage, and `-race` were not run by me or, per the lane, by the lane. No Windows test execution; build and vet only, per the lane.
- I did not open each RED cell body in §E.2, nor each mutant result, nor the AC-OWR-18 and AC-OWR-15 test bodies. "Red for the stated reason before green" is accepted from the recorded matrix.
- I did not re-run the 63 top-level lane PASS set. I re-ran only `internal/manifest`, `internal/cli/update/backup`, and the one overlay test, so the "63 PASS, 0 FAIL" figure is the lane's, not mine.
- I did not re-run codex's F2, F3 and F4 probes, nor test a case-sensitive file system, nor measure `internal/cli/testdata/profile_callsite_baseline.txt` against a fresh call-site scan.
- The frozen call-site baseline exception row count (one) was not independently counted. I checked that `internal/profile` is byte-unchanged, not that the baseline file's row set is correct.
- Machine load was high (load average about 42), so I ran no more than three go commands, all sequential and anchored. `go build` and `go vet` for the whole tree were not re-run by me.
- No refused tool call occurred in this audit.

## Residual-risk

- Even with F1 fixed, the broader class "a code path reads a project file that a user symlink redirects outside the project" is only guarded for the paths the tests happen to cover. The registered-file path reads before the boundary check and was that way at the base.
- The hard invariant "no code opens or reads a profile root" is enforced by a source scan and a differential canary that cannot see run-time-assembled paths (N16) or a symlink redirect (F1). Treat the invariant as only partly mechanically guarded.
- A FIFO swapped between the kind check and the read is uncovered (disclosed). Time-of-check races are not tested.
- The codex review ran three rounds on the lane's side and a fourth fresh one here; each found a new defect class. A further pass may find more.
- The installed binary lags HEAD (`ee981a932` is an ancestor), so any CLI behaviour I did not build from the tree was not judged here.

## Iteration history

- Plan-audit rounds 1 to 4 (FAIL 0.68, FAIL 0.76, then 0.92 PASS-WITH-DEBT in the delta round): referenced from the §E.1 binding conditions, not re-read.
- Sync audit: this is the first sync-audit round recorded in `.moai/reports/t1865/`, to my reading.
- Verdict path forward: fix F1 test-first, decide F2's reading of REQ-OWR-017 and fix it, fix F3, and re-run the codex gate. Re-audit scope: the defect delta F1 to F3.

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-7770ccadc2f78594b050e777
