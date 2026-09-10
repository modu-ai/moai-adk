# Progress — SPEC-GO-TOOLCHAIN-SEC-002

## Status: completed (sync-phase closed)

Tier S, Class C (global change). Card t610 (Factory lane-8), branch `WT-go-1266`, base
`d3b7d438d` (local develop at card creation; local develop has since moved to `d1b61005d`).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID self-check (executed): `ID="SPEC-GO-TOOLCHAIN-SEC-002"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS` (re-run in revision 0.1.1 → `PASS`)
- ID uniqueness: only `SPEC-GO-TOOLCHAIN-SEC-001` existed under `.moai/specs/`, and `ls -d .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `No such file or directory`
- Baseline evidence: `.moai/reports/t610/baseline/` (go.mod state, CI go-version-file refs, doc mentions, toolchain versions, govulncheck ×3 toolchains + exit codes). Command record: `.moai/reports/t610/baseline/commands.md`
- RED-now ledger: acceptance.md § D.0 (E-01..E-08, pinned `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1`)
- Decisions awaiting Implementation Kickoff Approval: plan.md § D-DESIGN D1–D4
- D2 delta 1.26.6→1.26.8 (plan.md § D2): no security fixes. No 1.26.8 fix reaches the shipped binary (`CGO_ENABLED=0`; no `debug/elf` import; linux/darwin/windows targets only). go1.26.7 net/http #80927 needs unencrypted HTTP/2, which this repo does not configure (`git grep -n -E 'UnencryptedHTTP2|h2c|x/net/http2' -- '*.go' go.mod` → no output; judged on stdout, the exit value was not independently observable through the tool)
- Plan-audit iteration 1: FAIL (3 blocking, 15 non-blocking), report `.moai/reports/t610/plan-audit.md`. Revision 0.1.1 addresses B1–B3, N1–N13, and N15; N14 is kept as a recorded Gap (plan.md § D3)
- Revision 0.1.1 lint (executed): `moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `✓ No findings — all SPEC documents are valid`, exit 0
- Plan-phase Gaps:
  - the upstream issue diffs for go1.26.7/1.26.8 were not read
  - Go's `http.Server` default for unencrypted HTTP/2 was not measured, so the #80927 impact is a Gap; the `internal/web` tests in AC-GTS2-006 are a net/http regression guard, not coverage of that path
  - how `go-version-file` treats a `toolchain` directive was not verified (plan.md § D3)
  - the version half of AC-GTS2-005's RED is observed only on the installed binary, not on a build from this tree (acceptance.md E-06)
  - whether `make build` modifies tracked generated files was not observed; M3 records it
- Closed in 0.1.1: the baseline govulncheck invocation string is recorded in `commands.md` (acceptance.md E-05)

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-10
plan_revision: "0.1.2"
tier: S
req_count: 8
ac_count: 8
```

## §E.2 Run-phase Evidence

Measured by manager-develop on 2026-09-10 in `.claude/worktrees/t610` (branch `WT-go-1266`),
on the M1 commit: HEAD `41f445fa54f3a4949491bbe9fe06c68a97cf283e`, tree
`e6ec2991e4ce028b83d2be0658844ddb3813a8dd`. Shell environment carried no `GOTOOLCHAIN` or
`GOFLAGS`. Evidence directory: `.moai/reports/t610/run/` (command index: `commands.md`).
Every exit code is in a sibling `.exit` file, written from the command's own `$?`.

Diff reference at measurement time: `git rev-parse develop` → `80e9e003996b32c84b9ee6df1201c49081062f0d`,
`git merge-base develop HEAD` → `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1` (the card base, unchanged,
so `develop...HEAD` holds only this card's commits).

AC-GTS2-001 was measured first and passed before any other AC was judged.

| AC | Status | Command | Key output (verbatim) | Exit | Evidence |
|----|--------|---------|-----------------------|------|----------|
| AC-GTS2-001 | PASS | `go -C <worktree-root> version`; `go -C <worktree-root> env GOTOOLCHAIN GOMOD` | `go version go1.26.8 darwin/arm64`; `auto` / `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610/go.mod` | 0, 0 | `ac001-toolchain.txt` |
| AC-GTS2-002 | PASS | `sed -n 3p go.mod`; `grep -c '^toolchain' go.mod` | `go 1.26.8`; `0` (control `grep -c '^go ' go.mod` → `1`) | 0; 1 (no-match grep, judged on stdout `0`) | `ac002-directive.txt` |
| AC-GTS2-003 | PASS | `git diff --numstat develop...HEAD -- go.mod`; `git diff develop...HEAD -- go.mod` | `1	1	go.mod` (bytes `1 \t 1 \t go.mod \n`); single hunk `-go 1.26.4` / `+go 1.26.8` | 0, 0 | `ac003-numstat.raw`, `ac003-diff.raw` |
| AC-GTS2-004 | PASS | `GOMAXPROCS=2 timeout 590 govulncheck ./...` (no override; env `auto` + worktree go.mod recorded) | `Your code is affected by 0 vulnerabilities.`; `This scan also found 0 vulnerabilities in packages you import and 3` / `vulnerabilities in modules you require` | 0 | `ac004-govulncheck.log`, `ac004-govulncheck-env.txt` |
| AC-GTS2-004 (IDs) | PASS | `grep -c '<ID>'` on the scan log, for each of the 8 IDs | `0` for all 8; controls `affected by 0` → `1`, `modules you require` → `1` | 1 ×8 (no match); 0 ×2 | `ac004-ids.txt` |
| AC-GTS2-004 (instrument control) | observed | same `grep -c '<ID>'` on `baseline/govulncheck-auto.log` | `2` for each of the 8 IDs | 0 ×8 | `ac004-ids-control.txt` |
| AC-GTS2-005 | PASS | `make build`; `go version -m bin/moai` | build ends `go build -ldflags ... -o bin/moai ./cmd/moai`; first line `bin/moai: go1.26.8` | 0, 0 | `ac005-make-build.log`, `ac005-go-version-m.txt` |
| AC-GTS2-006 (vet) | PASS | `go vet ./...` | no diagnostics; swept-set control `go list ./...` → `140` packages | 0 | `ac006-go-vet.log`, `ac006-vet-pkgset.txt` |
| AC-GTS2-006 (tests) | PASS | `go test -count=1 -timeout 600s ./internal/web/... ./internal/update/... ./internal/goal/... ./pkg/...` | `ok` ×5: internal/web 19.832s, internal/update 6.486s, internal/goal 1.318s, pkg/models 2.612s, pkg/version 0.796s | 0 | `ac006-go-test.log`, `ac006-swept.txt` |
| AC-GTS2-007 (M1) | PASS | `git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610'` | empty stdout (0 bytes) | 0 | `ac007-complement-m1.raw` |

AC-GTS2-006 swept set: 5 packages selected (`go list` over the four patterns printed exactly
`internal/web`, `internal/update`, `internal/goal`, `pkg/models`, `pkg/version`), 5 `ok` lines,
0 lines matching `no test files|no tests to run|^FAIL|^---`. No `-run` selector was used.
`internal/cli` was excluded: no lead slot approval was given, so its tests were not run.

`make build` and tracked files: `git status --porcelain --untracked-files=no` before and after
`make build` both printed ` M .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/progress.md` (the
orchestrator's uncommitted §F section), and `cmp` of the two captures exited 0. The build's
`templ generate` and `gen-catalog-hashes.go --all` steps changed no tracked file, so nothing
regenerated needs excluding from the commit.

Instrument notes: the first `ac002-directive.txt` capture passed `--` to BSD `sed`, which read it
as a filename (`sed_exit=1`); the file was overwritten by a clean capture. govulncheck ran under
`timeout 590`, not 900, because the Bash tool caps a foreground call at 600 s; the scan exited 0
inside that bound. AC-GTS2-007 is re-checked after the evidence commit
(`ac007-complement-m3.raw`), recorded in the commit report.

## §E.3 Run-phase Audit-Ready Signal

- M1 commit: `41f445fa5` (`go.mod` line 3 + spec.md `status: draft → in-progress`).
- M2–M3 evidence commit: this commit (SHA not self-referenceable; see `git log develop..HEAD`).
- Verification split: the lane ran `go vet ./...`, `make build`, govulncheck, and the tests of 5
  selected packages. The all-package verdict is NOT a lane claim: it belongs to the lead's batch
  full run on the develop tip and to CI on the lead's `origin/develop` push after M4. The lane did
  not run `go test ./...` and did not run `internal/cli` tests.
- Not done by the lane: no push, no binary install (`~/go/bin/moai` untouched), no M4/M5 work.

```yaml
run_complete_at: 2026-09-10
run_commit_sha: 41f445fa5
run_status: audit-ready
ac_pass_count: 7
ac_fail_count: 0
ac_scope: "AC-GTS2-001..007 (AC-GTS2-008 is sync-phase, M5)"
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (no push by the lane; lead batch-pushes origin/develop)
l44_post_push_fetch: not-applicable (no push)
new_warnings_or_lints_introduced: "go vet 0 diagnostics; golangci-lint not run (no Go source change)"
cross_platform_build:
  darwin_arm64: "make build exit 0, bin/moai go1.26.8"
  windows_linux: not-run (CI matrix owns it)
total_run_phase_files: 2 tracked edits (go.mod, spec.md) + evidence under .moai/reports/t610/run/ + progress.md
m1_to_mN_commit_strategy: "M1 directive commit, then one M2-M3 evidence commit"
full_suite_verdict_owner: "lead batch full run + CI on origin/develop (after M4)"
```

## §E.4 Sync-phase Audit-Ready Signal

Measured by manager-docs on 2026-09-10 in `.claude/worktrees/t610` (branch `WT-go-1266`), before
the sync commit. Evidence directory: `.moai/reports/t610/sync/` (full command/output/exit
detail: `ac008-evidence.md`).

**AC-GTS2-008 result table** (grep-based, unpiped exit codes):

| File | `1.26.4` count | exit | `1.26.8` count | exit |
|------|---------------:|------|---------------:|------|
| `.moai/project/product.md` | 0 | 1 (no-match) | 2 | 0 |
| `.moai/project/structure.md` | 0 | 1 (no-match) | 1 | 0 |
| `.moai/project/codemaps/overview.md` | 0 | 1 (no-match) | 1 | 0 |
| `.moai/project/codemaps/modules.md` | 0 | 1 (no-match) | 1 | 0 |
| **Total** | **0** | — | **5** | — |

Control (proves the grep instrument reads the token): `git show
403bac94b339c19ebcee29e12b1a722cfd3d69ee:.moai/project/product.md` →
`control-product-at-403bac94b.md`, then `grep -c '1\.26\.4'` on that pre-edit capture → `2`
(exit 0), matching spec.md § D.0 E-08's baseline for `product.md`.

Verdict: AC-GTS2-008 PASS.

**Files changed this sync phase:**
- `.moai/project/product.md` (2 mentions: lines 244, 300)
- `.moai/project/structure.md` (1 mention: line 132)
- `.moai/project/codemaps/overview.md` (1 mention: line 6)
- `.moai/project/codemaps/modules.md` (1 mention: line 6)
- `CHANGELOG.md` (`[Unreleased]` → `### Fixed`, one new entry citing SPEC-GO-TOOLCHAIN-SEC-002)
- `.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/spec.md` (frontmatter `status: in-progress → completed`)
- `.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/progress.md` (this file: §E.4 + status line)
- `.moai/reports/t610/sync/` (this evidence directory)

CHANGELOG pre-emission self-test (B12): `grep -c 'SPEC-GO-TOOLCHAIN-SEC-002' CHANGELOG.md` → `0`
(exit 1, no match) before emission — no duplicate entry risk. AC count self-test: `grep -oE
'AC-GTS2-[0-9]+' acceptance.md | sort -u | wc -l` → `8`, matching the CHANGELOG entry's "8
acceptance criteria" claim.

```yaml
sync_complete_at: 2026-09-10
sync_commit_sha: pending-backfill-sync
sync_status: audit-ready
b12_self_test_a: "grep -c SPEC-GO-TOOLCHAIN-SEC-002 CHANGELOG.md -> 0 (pre-emission), no duplicate"
b12_self_test_b: "AC-GTS2-[0-9]+ distinct count in acceptance.md -> 8, matches CHANGELOG claim"
b12_self_test_c: "ls verified: product.md, structure.md, codemaps/overview.md, codemaps/modules.md all exist"
changelog_entry_position: "[Unreleased] -> ### Fixed (first entry under that heading)"
frontmatter_status_transitions.spec_md: "in-progress -> completed (status + updated only)"
canary_compliance_check: "not applicable — this SPEC defines no forward-looking policy that its own sync tests"
```

## §F Phase 4 Mode Selection

Implementation Kickoff Approval: passed 2026-09-10 — lead kickoff (delegated conditions: plan-audit-2 PASS, 0 blocking, one-sentence change) plus operator confirmation in the lane-8 session ("진행").

Input parameters:
- tier: S
- scope: 1 file changed (go.mod line 3), plus evidence files under `.moai/reports/t610/`
- domain count: 1 (build toolchain)
- file language mix: go.mod directive only, no Go source
- concurrency benefit: none (single edit, sequential verification)
- Agent Teams: not requested

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | The change is one line, but the verification ledger (AC-001..007 evidence files) is run-phase work owned by manager-develop |
| serial | yes | One manager-develop spawn covers M1–M3 in order |
| fanout | no | Single domain, nothing to parallelize |
| sweep | no | One file, not a high-volume mechanical transform |

Decision: serial

Justification: a one-line directive change with a fixed sequential verification chain (toolchain gate → govulncheck → build → vet → scoped tests) has no parallel split. Verification scope follows the lead's rule change of 2026-09-10: the go directive reaches every package, so the lane shows only that the toolchain switch does not break the build (`go build ./...`, `go vet ./...`, govulncheck, a few representative package tests including `internal/web`). The all-package verdict belongs to the lead's batch full run on the develop tip and to CI. `internal/cli` tests are not run by the lane.
