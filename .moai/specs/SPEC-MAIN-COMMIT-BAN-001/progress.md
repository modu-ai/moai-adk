# Progress — SPEC-MAIN-COMMIT-BAN-001 (card t1337)

Worktree: `.moai/worktrees/t1337` (branch `WT-main-commit-ban`, base `9f738aadd`).
Cycle: TDD (plan §B). This file is created by manager-develop at run-phase entry;
§E.1 records the plan-phase signal, §E.2 the run-phase evidence.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-29
- plan-audit: iter2 PASS 0.91 (plan commit `9f738aadd`; D1-D8 settled)

## §E.2 Run-phase Evidence

Every claim below carries the command actually run, its verbatim output, and the
tree it was measured on (verification-claim-integrity §2; attribution triple per
SPEC-SYNC-PARALLEL-DOCS-001 A9).

### AC-4 RED baseline (four elements, verification-completeness §2.1)

- Tree: `9f738aadd` (the run-phase entry commit; measured BEFORE any
  implementation edit landed — the working tree was clean at session start).
- Command (single invocation, scratch probe `internal/hook/zz_red_probe_t1337_test.go`,
  deleted after capture — the committed evidence carrier is this section, which
  lands in its own commit preceding the implementation commits per VCI §2.3):

```
$ go test ./internal/hook/ -run 'TestRedProbe_CommitDenyAbsent' -count=1
--- FAIL: TestRedProbe_CommitDenyAbsent (0.41s)
    zz_red_probe_t1337_test.go:26: RED baseline: checkBranchState(git commit at primary on main) decision = "" reason = "", want "deny" — the commit-family deny does not exist on this tree
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.013s
FAIL
exit code: 1
```

- Why red (the stated reason): `checkBranchState`'s pattern set carries no
  commit-family matcher at `9f738aadd`, so a protected-branch primary-checkout
  command falls through the allow path — the commit-shaped hole the plan's RED
  cell pins.

### D8 presence test — RED then GREEN

- RED (before M1 edits, same tree): all four subtests failed — template
  workflow.yaml no `branch_guard:` key; template git-strategy no
  `lead_push_threshold`; local `deny_commits_on` nil; local threshold nil.
- GREEN (after M1): `go test ./internal/config/ -run
  '^TestTemplateConfigCarriesMainCommitBanKeys$' -count=1` → `ok
  github.com/modu-ai/moai-adk/internal/config 0.344s`.
- fallout fixed in the same milestone: `TestShippedConfigKeysHaveReaders`
  (REQ-CKH-008 anti-rot) flagged the three newly-shipped keys as untriaged →
  registered in `internal/config/testdata/shipped_key_inventory.yaml`
  (branch_guard keys W, lead_push_threshold P) → package `ok ... 4.067s`.

### ResolveHeadBranch unit tests (gitcore)

- `go test ./internal/core/... -run '^TestResolveHeadBranch' -count=1 -v` →
  OnBranch / DetachedHead / NonGitDir / EmptyDir all PASS (real git fixtures;
  the `./internal/core/...` form is used because the worktree-session guard
  refuses a command line carrying the `.../internal/core/git/` path segment).

### Protected-commit family + full hook package (M2 GREEN)

- Family: `go test ./internal/hook/ -run '^TestProtectedCommit|^TestMatchProtectedCommitCommand|^TestPreTool_ProtectedCommit' -count=1` →
  `ok github.com/modu-ai/moai-adk/internal/hook 5.265s` (26 subtests: AC-4
  headline with real `main` repo + real HEAD resolution; AC-5 non-protected;
  AC-6 worktree cwd with stub.calls==0; AC-7 empty list zero invocations
  (nil + empty); AC-8 resolver-error advisory + detached head; AC-9 both
  exemption axes; quoted-data deny; foreign-carrier allow; compound deny;
  pwsh payload deny; case-fold deny; handler-level Handle deny + inert
  default).
- Full package: `go test ./internal/hook/ -count=1 -timeout 30m` →
  `ok github.com/modu-ai/moai-adk/internal/hook 357.620s` under a held
  `hook-suite` slot lease (acquired, released after; `moai slot status` →
  `no slot leases recorded`).
- Test-harness defect found and fixed during GREEN: the first matrix run
  failed two arms (`git commit -m "git switch main"`, `pwsh -Command ...`)
  because the input builder concatenated raw JSON, corrupting commands with
  embedded quotes into unparseable tool input — the implementation was never
  at fault; `protectedCommitInput` now marshals.

### Session incident (recorded, not caused by this work)

Mid-M2 the worktree-isolation anchor of this session resolved to the t1315
merge window and refused every mutating tool call in t1337 (Bash/Write/Edit).
Coordinator repaired the anchor (root cause: parent-session contamination;
incident file `.moai/reports/t1337/anchor-contamination-incident.md`, card
t1339). All 12 in-flight files survived intact; no work was lost. Resumed and
re-verified toplevel/branch/HEAD before continuing.

## §E.3 Run-phase Audit-Ready Signal

(pending — populated at M4 close)
