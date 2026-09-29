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

### M3 doctrine verification (final tree)

- Template guard family: `go test ./internal/template/ -run
  '^(TestTemplateNeutralityAudit|TestSanitizedPairParity|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestTemplateNoInternalContentLeak)$'
  -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 1.187s`.
- AC-16 worktree sentinel set: `commit-dead` hits all 5 expected files;
  `branch -f main origin/main` hits the chain doc; `lead_push_threshold` hits
  the chain doc + AGENTS.local.md; `deny_commits_on` hits workflow.yaml +
  types.go; `commit-protected` hits the template AGENTS.md.tmpl. Every probe
  non-zero.

### M4 rehearsal (AC-13) — observed, not assumed

Scratch clone at the primary's real state class; both refusal classes
observed verbatim (modified-tracked-set first, untracked-overwrite after the
tracked set clears; bare-SHA switch needs `--detach`); the drafted
`git diff develop --stat` EMPTY precondition was DISPROVEN by observation
(untracked develop-adds count as deletions) and the procedure doc corrected to
the sound two-check form in the same commit as this update. Frozen-ref
invariant: `git rev-parse main` = `c8f245c2c9a58083518f0b7cbebff0542848e033`
before AND after the rehearsal. Full log:
`.moai/reports/t1337/rehearsal-main-residue.md`.

### Final verification batch (M4, all observed this run, this tree)

| Claim | Command | Observed |
|---|---|---|
| scoped suites | `unset MOAI_* && go test ./internal/hook/ ./internal/config/... ./internal/core/... -count=1 -timeout 30m` | ok hook 466.145s · config 13.361s · config/atomicfile · config/toolpolicy · core/git 141.968s · core/project · core/quality — exit 0, under a held `hook-suite` lease |
| vet | `go vet ./internal/hook/ ./internal/core/... ./internal/config/...` | no output, exit 0 |
| windows build | `GOOS=windows go build ./...` | no output, exit 0 |
| lint (CI version) | `golangci-lint run internal/hook/... internal/config/...` (v2.1.6 verified) | `0 issues.`, exit 0 |
| lease release | `moai slot release --resource hook-suite` + `moai slot status` | released; `slot hook-suite: free` |

Command-form substitution (guard-measured, recorded): the batch uses
`./internal/core/...` — the worktree-session guard refuses a command line
carrying the `.../internal/core/git/` path segment; the `.../core/...` form is
a strict superset of the acceptance batch's `core/git` target.

### Lead-run pending checklist (lane did NOT execute; exact commands)

```
grep -l 'commit-dead' /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md    # AC-3/AC-16 primary probe — post-disposition, REQ-4.6
git rev-parse main            # primary HEAD state, AC-13 falsifiable read (expect c8f245c2c9a58... pre-disposition)
```

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-09-29
- ACs: AC-1..AC-15 lane-verified PASS (matrix below); AC-3/AC-15/AC-16 carry
  PENDING-LEAD primary probes with the exact commands above (the primary-only
  CLAUDE.local.md §4.1 clause is the operator's disposition-time step per
  REQ-1.4/REQ-4.6 — never a lane edit)
- AC-4 RED→GREEN: RED observed at 9f738aadd (four elements in §E.2), GREEN by
  M2 (family selector ok 5.265s; full hook package ok 466.145s)
- Evidence paths: this file (committed), `.moai/reports/t1337/rehearsal-main-residue.md`
  (local-only content, untracked), `/tmp/t1337-switch-refusal*.txt` (scratch —
  known loss, deciding lines transcribed into the rehearsal log)
- Commits (branch WT-main-commit-ban): 596962243 (bookkeeping+RED evidence) →
  580c64cd1 (M1) → 118076418 (M2) → 18fbd4212 (M3) → this commit (M4)
- Gaps: full-suite `go test ./...` is CI's verdict (lane-local scoping); the
  real-primary disposition execution is the operator's post-merge act; the
  primary-run probe set is lead-executed
- Residual-risk: the worktree-session guard's command-string parsing rejected
  two legitimate forms this run (core/git path segment; computed sed arg) —
  lanes needing those forms must use the documented substitutes; the anchor
  contamination incident (card t1339) could recur to any concurrently spawning
  lane

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-09-29
- sync_commit_sha: 2e5de5413
- Sync commit scope (the single 3-phase close commit): spec.md frontmatter
  `in-progress → completed` (merged transition per the Status Transition
  Ownership Matrix — no separate Mx chore commit), this §E.4 section, and the
  CHANGELOG.md Unreleased entry. No `.go` file is touched at sync; run-phase
  evidence stands at `815ba7e8b`.
- CHANGELOG judgment — entry ADDED under `## [Unreleased] → ### Added` (top
  position, newest-first like the t1312/t1311 entries): the commit deny is a
  shipped-code behavior change (new deny class behind
  `workflow.branch_guard.deny_commits_on`, ships `[]`) and
  `git_strategy.manual.lead_push_threshold` is a shipped key — both inside the
  file's established scope (the card t1312 session-record retention entry is
  the same shape). Pre-emission duplicate check:
  `grep -c 'SPEC-MAIN-COMMIT-BAN-001' CHANGELOG.md` → `0` before drafting.
  Entry content verified against the tree: symbol names
  (`matchProtectedCommitCommand`, `checkProtectedCommit`, `ResolveHeadBranch`)
  and every cited path read back from the implementation files, not from the
  plan.
- User-doc impact — ZERO wrong-text hits; no user-doc edit required:
  - `grep -rn 'already-checked-out\|already checked out'` over `README.md`
    `README.ko.md` `README.ja.md` `README.zh.md` + `docs-site/content` → 0
    matches (the old AGENTS.md §2 wording is maintainer-internal; the template
    AGENTS.md clause is generic, not user-doc surface).
  - docs-site `advanced/config-sections.md` (4 locales) documents the
    BranchGuard's legacy deny set (`switch` / `checkout` / `reset --hard` /
    `stash` / `rebase`) — still accurate: none of that text becomes false, and
    the commit deny ships inactive (`deny_commits_on: []`). The new key's
    documentation ships inline in the template `workflow.yaml` itself.
    `advanced/autonomous-loops.md` mentions the BranchGuard only as a
    pattern-sibling reference — unaffected.
  - Coverage of the new keys in docs-site would be a docs-completeness
    follow-up (4-locale same-PR obligation), not a sync correctness blocker —
    left to the lead's discretion.
- Codemaps: none exist in this tree — `ls codemaps` → absent (verified during
  t1315, re-verified this sync). Nothing to rotate.
- Lead-run pending checklist (PENDING-LEAD — primary-only probes, LEAD/OPERATOR
  owned; the lane did NOT execute them and MUST NOT: the primary checkout is
  outside this worktree and the probes are disposition-coupled):
  - `grep -l 'commit-dead' /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md` —
    AC-3/AC-16 primary probe; expected hit POST-disposition only (the
    CLAUDE.local.md §4.1 clause is the operator-side step per REQ-1.4/REQ-4.6).
  - `git rev-parse main` (run at the primary checkout) — AC-13 falsifiable
    read; expected `c8f245c2c9a58083518f0b7cbebff0542848e033` pre-disposition.
  - Both rows stay lead-owned past this SPEC's close: AC-3's full pass and
    AC-16's primary set complete at disposition time, not at sync time.
- sync_commit_sha backfill note: recorded as the `pending-backfill-sync`
  placeholder in the sync commit `2e5de5413` itself (a commit cannot reference
  its own hash; SHA-placeholder backfill exemption,
  spec-frontmatter-schema.md § D3) and backfilled to `2e5de5413` in this
  follow-up commit.
