# SPEC-PRECOMMIT-VET-MONOREPO-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-04
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
baseline_sha: b9298de32
reference_patch: t312-precommit-vet @ b6f478b1a (read + judged; adoption-by-re-authoring, spec.md §A)
twin_pair:
  - internal/cli/hook_install_precommit.go   # constant preCommitHookContent, vet block :78-103
  - internal/template/templates/.git_hooks/pre-commit   # vet block :35-59
```

Plan-phase notes for the auditor:

- **Approach settled, no OPEN decision**: adopt reference patch `b6f478b1a`'s logic (module-root walk → MODROOTS grouping → per-module subshell vet → module-root naming in the failure message), re-authored in place with generalized comments. The Implementation Kickoff Approval gate is unchanged and still mandatory before run-phase entry.
- **Local repro impossibility is structural**: this repository's `go.mod` is at the repo root, so the defect cannot be reproduced against this tree. The RED mechanism is the two contrast tests (AC-PVM-002/003) with a `submod/`-only fixture inside `t.TempDir()`, run through the existing `runPreCommitHook` harness.
- **RED ordering is load-bearing**: the RED cells must be captured at M1's opening act against the UNCHANGED constant pair, before the twin edit lands (plan.md §F M1; verification-completeness §2). AC-PVM-003's RED discriminates via the stderr naming assertion, not the exit code (exit 1 occurs on old for the wrong reason — stated in acceptance.md so the RED is red for the right stated reason).
- **Release coordination (one line, for the lead)**: t230's landing precondition is satisfied (`32d2221fa` + `539349c5b` are develop ancestors); the remaining "at least one release must pass after t230's landing before this ships" is a deployment-time concern owned by release card t204 and does not block this SPEC's phases (AC-PVM-012).

## §E.2 Run-phase Evidence

> Every row carries the attribution triple per VCI §2: (a) the exact command, (b) the verbatim observed output, (c) the tree SHA of the run. RED cells were captured against the PRE-EDIT constant pair at M1's opening act, before the twin edit landed; GREEN cells against the post-edit working tree (HEAD `bfdf833e8` + the run-phase diff that becomes the run commit).

### M1 pre-flight baseline (measured before the first change)

| Check | Command | Observed |
|---|---|---|
| Branch + HEAD | `git branch --show-current && git rev-parse --short HEAD` | `WT-precommit-vet-module`, `bfdf833e8` |
| Build | `go build ./...` | exit 0 |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Pre-change guards | `go test -count=1 ./internal/cli/ -run 'TestPreCommit'` | `ok  github.com/modu-ai/moai-adk/internal/cli  15.537s` (exit 0) |
| Lint baseline | `golangci-lint run --timeout=2m ./internal/cli/...` | `0 issues.` (exit 0) |

### M1 opening act — RED cells (captured against the PRE-EDIT constant pair)

Command (single invocation, both tests): `go test -count=1 -run 'TestPreCommitHook_Submodule' ./internal/cli/...`
Exit code: `1`. Tree SHA: `bfdf833e8` (constant pair untouched; only the two new tests present in the working tree). Fixture: repo inside `t.TempDir()` whose only `go.mod` lives in `submod/` (`module sample`, go 1.21); staged gofmt-clean `.go` under `submod/`; shim PATH with go present and moai absent.

Verbatim output:

```text
--- FAIL: TestPreCommitHook_SubmodulePassesClean (0.34s)
    hook_install_precommit_test.go:627: expected exit 0 (clean submodule file passes), got 1
        stderr:

        [pre-commit] FAILED: go vet reported issues in the staged packages.
        [pre-commit] Hint: run go vet on the affected packages, fix, then re-commit.
        [pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit
--- FAIL: TestPreCommitHook_SubmoduleVetBlocks (0.27s)
    hook_install_precommit_test.go:653: expected 'module root: submod' in stderr, got:

        [pre-commit] FAILED: go vet reported issues in the staged packages.
        [pre-commit] Hint: run go vet on the affected packages, fix, then re-commit.
        [pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.709s
```

RED reasons (each red for the right stated reason):
- `TestPreCommitHook_SubmodulePassesClean` — RED via **exit code** (got 1, expected 0): the defect itself — a gofmt-clean, vet-clean staged file is blocked because root-level `go vet ./submod` dies on module resolution (`go: cannot find main module`), misreported as a vet finding.
- `TestPreCommitHook_SubmoduleVetBlocks` — RED via the **stderr discriminator**: the OLD constant also exits 1 on this fixture (for the wrong reason — module-resolution failure), so the exit-code assertion alone is green-on-old; the missing `module root: submod` line is what makes this a true contrast.

### M1/M2 two-cell adoption record (GREEN flip, post-edit)

| Cell | Command | Verbatim output | Exit | Tree |
|---|---|---|---|---|
| GREEN AC-PVM-001 (twin identity) | `go test -count=1 -run 'TestPreCommitTemplateMatchesConstant' ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  1.086s` | 0 | `bfdf833e8` + run diff |
| GREEN AC-PVM-002 (clean pass) | `go test -count=1 -run 'TestPreCommitHook_SubmodulePassesClean' ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  1.370s` | 0 | `bfdf833e8` + run diff |
| GREEN AC-PVM-003 (vet block + naming) | `go test -count=1 -run 'TestPreCommitHook_SubmoduleVetBlocks' -v ./internal/cli/` | `=== RUN   TestPreCommitHook_SubmoduleVetBlocks` / `--- PASS: TestPreCommitHook_SubmoduleVetBlocks (0.39s)` | 0 | `bfdf833e8` + run diff |

The AC-PVM-003 PASS is the observation that stderr contained `module root: submod` for the blocking case — the assertion inside the test is exactly that check.

### AC matrix (12 rows)

| AC | Status | Verification (command → observed, this run, tree SHA) |
|---|---|---|
| AC-PVM-001 (twin identity, guard) | PASS | `go test -count=1 -run 'TestPreCommitTemplateMatchesConstant' ./internal/cli/` → `ok ... 1.086s`, exit 0 (`bfdf833e8`+diff) |
| AC-PVM-002 (submodule clean pass) | PASS | RED cell above (`bfdf833e8`, pre-edit, exit 1) → GREEN `ok ... 1.370s`, exit 0 (post-edit) |
| AC-PVM-003 (submodule vet block + naming) | PASS | RED cell above (stderr discriminator absent, pre-edit) → GREEN `--- PASS: ... (0.39s)`, exit 0 (post-edit) |
| AC-PVM-004 (repo-root fallback guard) | PASS | guards batch → `ok ... 2.235s`, exit 0; `TestPreCommitHook_GoVetBlocks` body unmodified |
| AC-PVM-005 (build tags before cd) | PASS | `grep -n 'BT_TAGS=""'` → Go const :95, template :52; `grep -n 'cd "\$_mr"'` → Go const :145, template :102. 95<145 and 52<102, no ORDER-FAIL |
| AC-PVM-006 (subshell isolation, both surfaces) | PASS | `grep -c '( cd "\$_mr" && go vet' <both files>` → `internal/template/templates/.git_hooks/pre-commit:1`, `internal/cli/hook_install_precommit.go:1` |
| AC-PVM-007 (gofmt guard) | PASS | guards batch → `ok ... 2.235s`, exit 0 |
| AC-PVM-008 (skip-bypass guard) | PASS | guards batch → `ok ... 2.235s`, exit 0 |
| AC-PVM-009 (toolchain-absent / no-staged-Go guards) | PASS | `go test -count=1 -run 'TestPreCommitHook_GoVetBlocks|TestPreCommitHook_GofmtBlocks|TestPreCommitHook_SkipBypass|TestPreCommitHook_ToolchainAbsent|TestPreCommitHook_NoStagedGo' ./internal/cli/` → `ok  github.com/modu-ai/moai-adk/internal/cli  2.235s`, exit 0; both bodies unmodified |
| AC-PVM-010 (package sweep, ≥600s, rc unpiped) | **FAIL — inherited red, not authored by this card** | `go test -count=1 -cover ./internal/cli/...` → exit **1**: `FAIL  github.com/modu-ai/moai-adk/internal/cli  462.872s` with exactly ONE failing test, `TestBinaryLag_DoctorCheckNameSetIsUnchanged` (`binary_lag_test.go:205: this SPEC added doctor check name "Hook Delivery"; REQ-BLV-009 rewires the existing "Binary Freshness" item and registers no new name`). All 16 subpackages `ok` (e.g. `ok .../internal/cli/worktree 10.857s coverage: 87.1%`, `ok .../internal/cli/update/deploy 6.238s coverage: 91.6%`). Attribution of the red: commit `2b4582a43` (`feat(SPEC-UPDATE-HOOK-DELIVERY-001): M3 hook-delivery doctor check — read-only detect+guide, card t466`) registered `{"Hook Delivery", ...}` in `internal/cli/doctor.go:230` without adding it to `namesAddedAfterBaseline` in `binary_lag_test.go`. The guard's inputs (`os.ReadFile("doctor.go")` + `git show lagBaselineSHA:internal/cli/doctor.go`) are untouched by this card's 4-file diff — the red is present on the baseline tree and reproduces deterministically in isolation (exit 1, 0.05s). Repair belongs to the t466/BinaryLag axis, out of this card's scope (drive-by fix deliberately not made). |
| AC-PVM-011 (cross-platform build) | PASS | `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (rc observed directly, unpiped) |
| AC-PVM-012 (release-coordination line) | PASS | `sed -n '/^## §E.3/,/^## §E.4/p' .moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/progress.md \| grep -c 'release card t204'` → `1` (measured after §E.3 was authored, on the working tree that becomes the run commit) |

### Sweep and quality gates (run-phase tree)

| Check | Command | Observed |
|---|---|---|
| gofmt (repo-wide) | `gofmt -l .` | 0 lines, exit 0 |
| gofmt (touched files) | `gofmt -l internal/cli/hook_install_precommit.go internal/cli/hook_install_precommit_test.go` | empty, exit 0 |
| go vet | `go vet ./internal/cli/...` | 0 output lines, exit 0 |
| Lint delta | `golangci-lint run --timeout=2m ./internal/cli/...` (post-change) | `0 issues.` exit 0 — zero new findings vs the pre-flight baseline (`0 issues.`) |
| Coverage (subpackages) | from the `-cover` sweep | 16 subpackages measured, 80.9%–100.0% (e.g. worktree 87.1%, deploy 91.6%, plan 95.0%) |
| Coverage (main `internal/cli` package) | — | **Unmeasured** — the package FAILs on the inherited red (see AC-PVM-010), and `go test -cover` prints no coverage line for a failing package. Named gap, not a claim. |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete_with_inherited_red
run_complete_at: 2026-09-04
run_baseline_head: bfdf833e8   # GREEN/sweep measurements taken on this HEAD + the run diff (identical content to the run commit)
twin_edit_commit: single atomic commit (both surfaces + tests + SPEC artifacts)
inherited_red:
  test: TestBinaryLag_DoctorCheckNameSetIsUnchanged
  authored_by: 2b4582a43 (SPEC-UPDATE-HOOK-DELIVERY-001, card t466)
  owner_of_repair: t466 / BinaryLag axis — NOT this card
ac_result: 11 PASS / 1 FAIL (AC-PVM-010, inherited red only)
```

**Run-phase close record (manager-develop).**

- **What changed**: the vet block inside `preCommitHookContent` (`internal/cli/hook_install_precommit.go`) and its template twin (`internal/template/templates/.git_hooks/pre-commit`) were re-authored in lockstep after reference patch `b6f478b1a`'s logic — `_moai_module_root()` upward walk, `MODROOTS` grouping, per-module `( cd "$_mr" && go vet $BT_TAGS $PKGS )` subshell, build-tags read hoisted before any `cd`, failure message + hint naming the module root. Two contrast tests added to `internal/cli/hook_install_precommit_test.go`. Nothing outside the vet block changed (gofmt block, skip-bypass, toolchain-absent guards, heavy-gate block, installer machinery — untouched; verified by the 5 guard tests staying green and the diff touching only 4 files).
- **Two-cell adoption**: RED captured at M1's opening act against the pre-edit constant pair (verbatim above, exit 1, tree `bfdf833e8`), GREEN after the twin edit (exit 0). Both cells recorded with command + output + exit code + tree.
- **The one red in the sweep is inherited, with the evidence chain in the AC-PVM-010 row**: `TestBinaryLag_DoctorCheckNameSetIsUnchanged` fails because commit `2b4582a43` (card t466) registered the "Hook Delivery" doctor check name without extending the BinaryLag guard's allowlist. This card's diff cannot reach either of that test's inputs. The repair is a one-line allowlist addition owned by the t466/BinaryLag axis; making it inside this card would be a drive-by fix across SPEC boundaries, so it is reported, not made. Until it lands, every card branching from current develop sees this same red in its `internal/cli` full sweep — the lead should route it to the t466 axis.
- **Gaps (explicitly NOT observed)**: (1) main `internal/cli` package coverage is unmeasured — suppressed by the inherited red; (2) the full sweep was measured with `-cover` appended to the AC-PVM-010 command form (same swept set, exit semantics unchanged); (3) the GREEN and sweep cells were measured on the working tree at HEAD `bfdf833e8` carrying the run diff — the run commit's tree is that exact content, so the commit itself re-witnesses the measured state.
- **Residual risk**: the unquoted `$MODROOTS`/`$PKGS` word-splitting (space-in-paths) is inherited from the previous single-invocation form and stays out of scope per spec.md §E.1 — a future quoting pass would touch the same twin pair and deserves its own card.
- **§E commands run this phase (exit codes)**: pre-flight `go build ./...` (0), windows build (0), `TestPreCommit` baseline (0), lint (0); RED `go test -run 'TestPreCommitHook_Submodule' ./internal/cli/...` (1, intended); GREEN AC-001/002/003 (0, 0, 0); sweep `go test -count=1 -cover ./internal/cli/...` (**1** — inherited red); `gofmt -l .` (0); `go vet ./internal/cli/...` (0); builds (0, 0); guards batch (0); post-change lint (0); isolated inherited-red repro (1).
- **Release coordination (for the lead to hand to the release card)**: t230's landing precondition is satisfied (`32d2221fa` + `539349c5b` are develop ancestors); the remaining "at least one release must pass after t230's landing before this ships" is a deployment-time concern owned by **release card t204** and does not block this SPEC's phases.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-09-04
sync_commit_sha: f964c73ed   # sync commit measured SHA — D3 two-commit pattern backfill (a commit cannot cite its own hash)
synced_artifacts:
  - CHANGELOG.md   # one [Unreleased]/Added entry: the monorepo vet fix, user-facing only
  - spec.md   # frontmatter status+updated only: in-progress → completed (the 3-phase close transition rides this sync commit)
  - progress.md   # this §E.4 signal
untouched_by_sync: §E.3 close record (run-phase-owned, byte-unchanged); internal/cli/** and internal/template/** (run-phase surfaces, closed at 45c60e1a3)
ac_matrix: §E.2 (11 PASS / 1 FAIL — AC-PVM-010, inherited red owned by the t466/BinaryLag axis; unchanged by sync)
gaps: main internal/cli package coverage unmeasured (suppressed by the inherited red — named in §E.2/§E.3, still unmeasured at sync); sync commit not pushed (lead owns integration)
```

**Sync-phase close record (manager-docs).** This single sync commit carries: the CHANGELOG entry, the `spec.md` frontmatter `in-progress → completed` transition (`status`+`updated` only — body content untouched), and this §E.4 signal. `plan.md` / `acceptance.md` carry no `status:` field to transition (artifact status-statelessness). The AC-PVM-012 close-record check (`sed -n '/^## §E.3/,/^## §E.4/p' … | grep -c 'release card t204'` ≥ 1) re-measured on the post-edit working tree before commit — §E.3 is byte-unchanged, so the count is unaffected by the §E.4 rewrite. The inherited-red note and the coverage gap are recorded in §E.2/§E.3 and are deliberately NOT changelog material.
