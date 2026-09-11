# SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 — progress (card t637)

Plan-phase artifacts authored 2026-09-11 on HEAD `1ad0fdc09` @ `WT-acquire-branch-record`
(worktree `.claude/worktrees/t637`). Tier M. Status: completed (sync-phase closed; see §E.4).

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set for Tier M: spec.md, plan.md, acceptance.md, progress.md (this file).
- SPEC ID regex check, run as Bash in this plan pass:
  `ID="SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
  → observed output `PASS`.
- ID uniqueness: `ls .moai/specs/SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001` → "No such file or
  directory" before creation; `git grep -l "SPEC-INTEGRATION-LOCK-TARGET-SOURCE"` → no output.
- Frontmatter: 12 canonical fields plus `tier: M`; `status: draft`; `phase` is a release target
  ("v3.2.0 target").
- Requirements: REQ-ILT-001..013 in GEARS notation (no IF/THEN modality) — within the Tier M
  ceiling of 16.
- Acceptance: AC-ILT-001..014 — within the Tier M ceiling of 16; includes both lead-mandated
  fixture controls (AC-ILT-005 same tree / own branch, AC-ILT-006 other tree / other branch), the
  no-warn negative (AC-ILT-007), flag and config sources (AC-ILT-003/004), old-record
  compatibility (AC-ILT-009), and the mutation guard (AC-ILT-014, rows a-f).
- Exclusions: spec.md §E carries three `### Out of Scope — <topic>` sub-headings, each with `-`
  bullets.
- Baseline: `git diff --stat 4c99d973e origin/develop -- <12 target paths>` and
  `git diff --stat 4c99d973e HEAD -- <same paths>` both printed nothing (exit 0) in this pass.
- AC-ILT-011 base: `git grep -n "moai integration acquire" -- <5 doc files>` → 8 lines, of which
  0 contain `--card`.
- Lint: `moai spec lint .moai/specs/SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001/spec.md` (installed
  binary v3.2.0-rc.7) → exit 0, "No findings — all SPEC documents are valid" (a first run
  flagged four REQ bullets whose `shall` sat on a wrapped second line and one traceability table;
  both were reshaped, then re-run). `spec_audit` for this ID → one INFO `EraAutoDetected` (V3R6),
  no drift.
- Decisions recorded in plan.md: warning on standard error (§B1), status text shape (§B2),
  git-flow predicate (§B3), release invocation carries no card (§E2), no catalog entry covers the
  kanban-dispatch rule (§E4).

### iter-1 audit repair (v0.2.0)

plan-auditor iter1 returned FAIL 0.74 (Tier M threshold 0.80; Testability 0.55; report
`.moai/reports/t637/plan-audit-SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001-iter1.md`). Repaired on HEAD
`fe92308e3` (the audit-report commit on top of `a3b913b85`):

- D1 — acceptance.md §A.4 defines `EV` and `BIN` as absolute paths with a one-invocation preamble;
  every fixture `acquire` / `status` / `release` carries `CLAUDE_PROJECT_DIR=/tmp/t637-fx` (C3′
  included).
- D2 — every deciding command moved out of the matrix into fenced blocks CMD-ILT-001..016; §A.2
  requires `-v`, no `no tests to run`, and one `--- PASS:` line per named test. Measured in this
  pass: the AC-ILT-012 positive-control pipeline `printf '+%s\n' '<sample with a card-id token>' | grep '^+[^+]' | grep -cE 't[0-9]{3}|SPEC-|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{9}'`
  → `1` (the AC-ILT-012 positive control matches); `grep -c '^\[moai:integration-lock\] warning:'`
  → `1` on a sample holding the line, `0` on a sample without it.
- D3 / D4 — AC-ILT-007 adds github-flow + EMPTY develop and personal-mode git-flow + EMPTY develop
  cells; §D.3 row b names the first (plus absent config), new row g names the second.
- D5 — plan M5 wording carries lowercase `integration target`; AC-ILT-010 checks it
  case-sensitively.
- Every pre-existing test name cited in CMD-ILT-002/007/015/016 exists: a `git grep` for the 13
  `func <Name>(` declarations over `internal/cli` and `internal/template` found 13.

### iter-2 audit repair (v0.3.0)

plan-auditor iter2 returned FAIL 0.88 with three blocking findings (report
`.moai/reports/t637/plan-audit-SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001-iter2.md`). Operator granted
one extra iteration (iter3 final). Repaired on HEAD `e51428db1`; acceptance.md and plan.md only,
no requirement or scope change:

- N1 — every fixture call and CMD-ILT-010(b) spells `/tmp/t637-fx-bin/moai`; the `BIN` variable is
  gone. `EV` remains, used only as a redirect/file argument (the iter2 guard probe accepted that
  form). Measured in this pass: `grep -n '\$BIN'` over acceptance.md matches only the §A.4 prose
  that explains the refusal.
- N2 — row b names only `TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn`, with the
  parenthetical corrected; new row i (absent config treated as git-flow) names
  `TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn`; §D.6 reads "rows a-i".
- N3 — the three fixture cells run `acquire` inside `( cd /tmp/t637-fx-wt/cardA && … )`; every
  command using worktree-relative paths (CMD-ILT-001..010, 014, 015, 016, §D.5) begins with
  `cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 &&` in its own invocation.
- N4 — the seam test is `TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch`, selected with
  `^(…)$`.
- N5 — CMD-ILT-014 adds a fourth count, `--card <` on the release line, expected `0`.
- N6 — the positive control is `'see t637 here'`; measured in this pass through the same pipeline
  → `1`. The earlier SPEC-ID-shaped sample string was also removed from this file.
- iter1 D13 remains **declined**. The `OwnershipTransitionUnmeasured` INFO names the
  `(none) → draft` creation commit `a3b913b85`, which has already landed; a trailer on any later
  commit does not measure that transition, and rewriting the landed commit is out of bounds. In
  addition, git only reads a custom trailer from the message's final paragraph, and this card's
  dispatch requires `🗿 MoAI` to be the final line, so a trailer cannot be both parsed and
  compliant. The INFO does not affect the lint exit status.

## §E.2 Run-phase Evidence

Run-phase by manager-develop (cycle_type=tdd), worktree `.claude/worktrees/t637`, branch
`WT-acquire-branch-record`, started on HEAD `4fcd2932e` (local develop `00ae57ad7` absorbed).
Executed in two halves because `internal/cli` compilation is a lead-gated slot: the **pre-slot
half** (everything that does not compile `internal/cli`) is done and evidenced below; the
**slot half** (every `internal/cli` test, the mutation rows against it, the fixture binary, the
fixture cells, `make build`) is handed to the lead as an ordered batch and is **not yet observed**.
Evidence files: `.moai/reports/t637/ac-evidence/`.

### E.2.1 Commits (this run)

| SHA | Subject |
|---|---|
| `2a1029aea` | test(integration): RED tests for branch provenance, card line, and git-flow fallback warning (card t637) |
| `ba0725be8` | feat(integration): record and show branch provenance, show the card, warn on git-flow caller fallback (card t637) |
| `856ee0da7` | docs(integration): documented acquire invocations carry --card <card-id> (card t637) |
| (this commit) | status draft → in-progress + this §E.2 |

The `draft → in-progress` transition rides this fourth commit rather than the first run-phase
commit; the RED commit was made before the status edit. The transition itself and its owner are
unchanged. Only `spec.md` carries a `status:` field — `plan.md` / `acceptance.md` are stateless on
the status axis (spec-frontmatter-schema § Artifact Statelessness), so nothing was added there.

### E.2.2 What was built (per plan.md)

- **Record** — `kanban.IntegrationLock.BranchSource` (`branch_source`, `omitempty`) plus the three
  constants `BranchSourceFlag` / `BranchSourceConfig` / `BranchSourceCaller`.
- **Resolution** — `resolveIntegrationTarget` returns a third value, the tier that decided the
  branch, computed from the same trimmed values (t449 order unchanged).
- **Config seam (§D)** — `config.LoadGitFlowIntegrationConfig(root) GitFlowIntegrationConfig`
  with fields `Manual`, `GitFlowWorkflow`, `DevelopBranch` and method `IsGitFlow()`.
  `LoadGitFlowDevelopBranch` now delegates to it; its test is unmodified and passes. One file
  read per `acquire`. The struct carries both predicate halves as separate fields so mutation
  rows b and g can each drop exactly one half without touching resolution.
- **Warning** — after `AcquireIntegrationLock` succeeds, and only when
  `source == caller && gitFlow.IsGitFlow()`, one line
  `[moai:integration-lock] warning: git-flow project with no develop branch configured; the window was recorded against the caller's branch "<b>". Set git_strategy.manual.develop_branch, or pass --branch <integration-target>.`
  goes to `cmd.ErrOrStderr()`. A refused acquire returns before the check.
- **Status** — `  card:     <id>` between holder and branch when recorded; branch line gains
  ` (source: <s>)` only when `branch_source` is recorded.
- **Help** — `--branch`: "The integration target branch the merge lands on, not the card branch
  being merged (default: the configured git-flow develop branch, else the current branch)".
- **Guard** — `internal/hook/integration_lock_guard.go` not edited.

Mechanical call-site adaptation (plan.md §C allowance; assertions untouched):
`internal/cli/integration_target_test.go` lines 136, 149, 162, 175 — `branch, wt :=` →
`branch, wt, _ :=` in the four `TestResolveIntegrationTarget_*` tests.

Test-name mapping: every name acceptance.md §D.0 lists was created verbatim — no mapping change.
Additions beyond the named set: subtest `TestIntegrationAcquire_RecordsBranchSource/caller`;
`TestIntegrationAcquire_WarningIsOnStderrOnly` carries subtests `text` and `json`; kanban test
`TestIntegrationLock_BranchSourceRoundTripsAndOmitsWhenEmpty`; stream-separating helper
`runIntegrationStreams` in `integration_lock_cli_test.go`.

### E.2.3 Observed (pre-slot) — command + verbatim tail, this tree

| Item | Command | Observed | Status |
|---|---|---|---|
| RED config (HEAD `4fcd2932e` + tests) | `go test ./internal/config/ -run '^(TestLoadGitFlowDevelopBranch\|TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch)$' -count=1 -v` | `undefined: LoadGitFlowIntegrationConfig` … `FAIL … [build failed]` rc=1 (`red-config-seam.txt`) | TOOL_FAILURE-class RED (compile) |
| RED kanban | `go test ./internal/kanban/ -run '^TestIntegrationLock_BranchSourceRoundTripsAndOmitsWhenEmpty$' -count=1 -v` | `undefined: BranchSourceCaller` … `[build failed]` rc=1 (`red-kanban-branch-source.txt`) | TOOL_FAILURE-class RED (compile) |
| RED cli (type-check only) | `go vet ./internal/cli/` | `lock.BranchSource undefined` … rc=1 (`red-cli-vet.txt`) | TOOL_FAILURE-class RED (compile) |
| GREEN config seam + legacy | same config command, after M2 | 13 `--- PASS:` lines, 0 `no tests to run`, rc=0 (`green-config-seam.txt`) | PASS |
| GREEN kanban + config packages | `go test ./internal/kanban/... ./internal/config/... -count=1` | `ok internal/kanban 157.140s`, `ok internal/config 4.182s`, atomicfile ok, toolpolicy ok (`green-kanban-config-pkgs.txt`) | PASS |
| Mutation b (config side) | seam `IsGitFlow` → `return c.Manual`; seam test | `--- FAIL: …/c_manual_github-flow` rc=1; restored `clean-rc=0` (`mut-config-b.txt`) | killed |
| Mutation g (config side) | → `return c.GitFlowWorkflow` | `--- FAIL: …/d_personal_git-flow` rc=1; restored (`mut-config-g.txt`) | killed |
| Mutation i1 (config side) | absent/unreadable branch returns `{Manual:true, GitFlowWorkflow:true}` | `--- FAIL: …/e_no_file` rc=1; restored `clean-rc=0` (`mut-config-i1.txt`) | killed |
| Mutation l (supplementary, kanban) | `json:"branch_source,omitempty"` → `json:"branch_source"` | `a record with no source carries a branch_source key` / `--- FAIL` rc=1; restored `clean-rc=0` (`mut-kanban-l.txt`) | killed |
| Vet (§D.5) | `go vet ./internal/cli/... ./internal/kanban/... ./internal/config/...` | `vet-rc=0` | PASS |
| gofmt (§D.5) | `gofmt -l internal/cli internal/kanban internal/config` | no output, `gofmt-rc=0` | PASS |
| AC-ILT-011 (CMD-ILT-014 verbatim) | see acceptance.md | `8` / `0` / `1` / `0` (`ac-011.txt`); re-measured site count after absorb = 8 (plan base 8) | PASS |
| AC-ILT-012 lines 1-4 (CMD-ILT-015) | diff / added-line / neutrality / control | `diff-rc=0` / `1` / `0` / `1` | PASS (lines 5-6 in slot) |
| AC-ILT-013 line 4 (CMD-ILT-016) | `git diff --stat 1ad0fdc09 -- internal/hook/integration_lock_guard.go` | no output, rc=0 | PASS (lines 1-3 in slot) |
| catalog | `git grep -c -e kanban-dispatch -e rules/ -- internal/template/catalog.yaml` | no match, rc=1 — no rules entry after absorb; gen-catalog-hashes not run | n/a |
| Mutation dry-runs | each slot row's `sed` applied to a scratchpad copy, `diff` shown | exactly one changed line per row (a, b, c, d, e, f, g, h, i1, i2, j, k) | ready |

**TDD evidence note.** Under the TDD result contract a compile failure is TOOL_FAILURE, not
EXPECTED_RED. The pre-slot RED above proves the tests preceded the implementation (commit
`2a1029aea` precedes `ba0725be8`) but is not assertion-level RED. Assertion-level RED for the
config and kanban tests is the killed mutation rows b/g/i1/l above. For the `internal/cli` tests,
assertion-level RED is supplied by the slot-batch mutation rows (a pre-implementation run of those
tests would require a temporary worktree at `2a1029aea`, which the dispatch forbids).

### E.2.4 Plan-audit iter3 optional findings

- **O1 (row i wording → two mutants).** acceptance.md is not edited by run-phase (ownership).
  Row i is executed as **two separate mutants**, each against the same named test
  `TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn`:
  **i1** — the absent/unreadable-file branch of `LoadGitFlowIntegrationConfig` returns a
  git-flow-flagged config (`internal/config/loader_integration_branch.go`);
  **i2** — the warning condition drops the whole predicate
  (`if source == kanban.BranchSourceCaller {`, `internal/cli/integration.go`).
  Both are in the slot batch; i1 is additionally observed killed by the config seam test above.
- **O2 (stale evidence in fixture cells).** The slot batch runs each fixture cell with one added
  step right after the preamble's `mkdir -p "$EV" &&`: `rm -f` of that cell's four evidence files
  by explicit name (no glob — an unmatched zsh glob aborts the whole command line). If the binary
  is absent, the chain stops at `test -x`, the subsequent `grep -c` reads a missing file and prints
  no count, so no earlier run's number can be re-counted. This is the only deviation from the
  verbatim CMD-ILT-011..013 text.

### E.2.5 Awaiting the lead-gated compile slot (NOT observed at run-phase write time; resolved in §E.3)

AC-ILT-001..010 (Go halves), AC-ILT-004/005/006 fixture halves (C3′/C1′/C2′), AC-ILT-012 lines
5-6 (template test, `make build`), AC-ILT-013 lines 1-3 (cli / config / hook invariant runs), and
AC-ILT-014 rows a-i on the `internal/cli` side (plus supplementary j, k, l). The ordered batch is
in the run-phase completion report returned to the orchestrator.

**Resolved in §E.3** — the lead-gated slot ran on 2026-09-12 (`.moai/reports/t637/slot-run.md`);
every item listed above is observed PASS there, and none remains open.

### E.2.6 Residual risks named for the lead

- **CLAUDE.local.md divergence (plan.md §E3).** The three edits are line-local (370, 389, 403 on
  this tree). The primary checkout's uncommitted 734-line working copy has these lines at
  338/357/371; a collision surfaces when that work is committed on `main` or when the release PR
  carries `develop`'s copy to `main`, not at this card's `develop` merge.
- `WarningIsOnStderrOnly/text` asserts stdout is exactly the acquired line. If the settings-drift
  precondition ever reports a non-clean verdict in a `t.TempDir()` scratch repo, it writes to
  stdout and the test fails for a reason unrelated to the warning; not observed yet (slot).

## §E.3 Run-phase Audit-Ready Signal

The lead-gated `internal/cli` compile slot closed the run-phase evidence gap left open in §E.2.5.
Evidence: `.moai/reports/t637/slot-run.md` (lane-7, 2026-09-12), backed by
`.moai/reports/t637/ac-evidence/`.

- All 14 acceptance criteria (AC-ILT-001..014) observed PASS in the slot: the 10 named
  `internal/cli` tests plus subtests (slot-001..010), the pre-existing t449 regression set
  (slot-016a, 10 tests), the config seam (slot-016b, 13 tests, run outside the slot), the
  template-neutrality and hook-integration-lock invariant runs (slot-015, slot-016c), and
  `golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/config/...` at
  `0 issues.`.
- All 13 named mutation rows (a-l, with row i split into i1/i2 per plan-audit iter3 O1) turned
  their targeted test(s) red; every mutant was restored from a `cp` backup and verified `cmp`
  byte-identical, then the full AC set was re-run clean (14 PASS, 0 FAIL,
  `.moai/reports/t637/ac-evidence/restored-pass.txt`).
- Fixture cells C1′/C2′/C3′ (`/tmp/t637-fx-bin/moai`, `CLAUDE_PROJECT_DIR=/tmp/t637-fx`)
  reproduced the git-flow caller-fallback warning (C1′, one warning line naming `WT-a`), the
  cross-tree mismatch made visible (C2′, `card: tB` next to `branch: WT-a (source: caller)`), and
  the configured-develop-branch silent path (C3′, `branch_source: config`, zero warnings). The
  real integration-lock file's sha256 was identical before and after every cell
  (`063ae13b46b700f78ae34e9418bac488933033bf` ×6) — no fixture touched the live window.
- `make build` exited 0; `internal/template/catalog.yaml` is unchanged (`git diff --stat` empty)
  after the Makefile's own `gen-catalog-hashes --all` run.
- Gaps carried forward from the slot report: `go test ./...` was not run (out of scope; CI is the
  full-suite judge per spec.md §D), the `make build` binary was not itself used for the fixture
  cells (a separate LDFLAGS-free build was), and cross-platform (windows/linux) builds were not
  checked locally.
- Residual risk carried forward: the `CLAUDE.local.md` line-number edits (370/389/403 on this
  tree) diverge from the primary checkout's uncommitted working copy (338/357/371); a collision
  is possible when that work lands on `main` or a release PR carries `develop`'s copy across, not
  at this card's `develop` merge.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-09-12"
sync_commit_sha: 38c058c09
sync_status: completed
b12_self_test_a: "grep -c 'SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001' CHANGELOG.md -> 0 (pre-emission, before this commit's edit)"
b12_self_test_b: "grep -oE 'AC-ILT-[0-9]+' acceptance.md | sort -u | wc -l -> 14; CHANGELOG entry cites '14 acceptance criteria'"
b12_self_test_c: "ls internal/cli/integration.go internal/kanban/integration_lock.go internal/config/loader_integration_branch.go .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/harness/hns-release-specialist.md CLAUDE.local.md -> all resolve"
changelog_entry_position: "Unreleased > Fixed, appended after the SPEC-SYNC-GATE-FAILSTATE-001 entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (status + updated only, this commit)"
  plan_md: "stateless on the status axis — no transition (spec-frontmatter-schema § Artifact Statelessness)"
  acceptance_md: "stateless on the status axis — no transition"
  progress_md: "not a status-bearing artifact — this §E.4 block is the sync signal"
canary_compliance_check:
  applicable: false
  reason: "this SPEC defines no forward-looking policy that its own sync tests"
```

Docs-site: `grep -rln 'moai integration' docs-site/content` returned no matches — the docs-site
carries no page describing `moai integration`, so no docs-site change is made by this sync.

MX tags: this SPEC's run-phase does introduce new exported symbols —
`config.LoadGitFlowIntegrationConfig`, `config.GitFlowIntegrationConfig`,
`(GitFlowIntegrationConfig).IsGitFlow`, and `kanban.BranchSourceFlag` /
`kanban.BranchSourceConfig` / `kanban.BranchSourceCaller`. `LoadGitFlowIntegrationConfig` has 2
non-test callers (`LoadGitFlowDevelopBranch` and `internal/cli/integration.go`'s acquire path),
so fan_in < 3 and no @MX:ANCHOR is required; its exported godoc already states the contract, so no
@MX:NOTE is added either. No dangerous pattern (goroutine, complexity >= 15) was introduced. No
@MX annotation change is made by this sync commit.

`sync_commit_sha` is recorded as `pending-backfill` in this commit — a commit cannot cite its own
hash — and is backfilled in a following commit, per the pattern already used by
SPEC-SYNC-GATE-FAILSTATE-001 (CHANGELOG.md, same convention).
