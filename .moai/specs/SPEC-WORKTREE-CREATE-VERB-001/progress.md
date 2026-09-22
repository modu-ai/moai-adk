# SPEC-WORKTREE-CREATE-VERB-001 — Progress

> Plan-phase artifact set authored 2026-09-22 by manager-spec (card t1070) in worktree `.claude/worktrees/t1070` (branch `WT-worktree-verb`, HEAD `0314801c2`). status: draft.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-WORKTREE-CREATE-VERB-001
phase: plan
status: draft
plan_status: audit-ready
plan_complete_at: 2026-09-22
author: manager-spec (card t1070)
baseline_tree: .claude/worktrees/t1070 @ 0314801c2 (WT-worktree-verb)
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - research.md
  - progress.md
self_check:
  spec_id_regex: PASS   # [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS  → PASS
  id_uniqueness: PASS   # 908-SPEC catalog grepped; no SPEC-WORKTREE-CREATE-VERB-001 collision
  frontmatter_schema: PASS  # canonical 12 fields present; phase: "v3.2.0 target"; status: draft
  gears_notation: PASS  # spec.md §B — Ubiquitous/When/Where patterns; no legacy IF/THEN
  out_of_scope_rule: PASS  # three "### Out of Scope —" H3 sub-headings with "-" bullets
  flat_file_rejection: PASS  # directory layout .moai/specs/SPEC-WORKTREE-CREATE-VERB-001/
amendment_note: "iter-1 plan-audit PASS-WITH-DEBT 0.80 (2026-09-22, .moai/reports/plan-audit/SPEC-WORKTREE-CREATE-VERB-001-iter1.md); D2-D5 micro-amendment applied per re-delegation — REQ-WCV-009 added (path-escape + plain-dir refusal, D3), REQ-WCV-005 gate-independence clause (D2), research.md Fact 1 inventory completed (D4), REQ-WCV-004 relabeled Event-driven (D5)"
open_items: []
```

The plan-phase decision marker was resolved at run kickoff after the live-Codex observation and operator relay below. Option 가 is now fixed.

## §E.2 Run-phase Evidence

### M1 live observation, decision, and cross-harness handoff

- **Operator relay received 2026-09-22.** Exact directive: `card: t1070`, `spec: SPEC-WORKTREE-CREATE-VERB-001`, `cmd: /moai run SPEC-WORKTREE-CREATE-VERB-001`, worktree `.claude/worktrees/t1070`, branch `WT-worktree-verb`, handoff commit `df26e519d`; the operator stated that the existing-tree `moai codex -w` entry observation satisfies M1 in place and that agent-24 would terminate while keeping the sole worktree copy.
- **Relay character:** this was not a direct reply in the source Claude terminal. The preserved chain is operator → Claude lane `agent-24` / session `86225c1e-006e-40d0-b804-96577cb8d022` / PID `57105` → Codex lane source session `01a0c58d-0153-74d0-9948-8c7ee59299c9`. The operator later confirmed agent-24 removal.
- **Takeover readback:** `ps -p 57105 -o pid=,stat=,etime=,command=` produced no row and exit 1; `git status --short` produced no output; branch was `WT-worktree-verb`; HEAD was `df26e519d4a7a9da947e7b114d45f38f98ea9170`; `develop...HEAD` was `0 1`. The stale L1 lock left by the terminated process was then removed with `git worktree unlock` without removing the tree.
- **Attribution caveat:** `moai session current --json` read the project-wide side channel as `31a1b5bf-162d-4a12-9604-a3051de1a6fa`, conflicting with the conversation's canonical source-session attribution above. This side-channel value is therefore not used as handoff authority.
- **Decision:** option 가, a new `moai worktree new <name>` adapter, is selected. Existing-tree entry is already provided by `moai codex -w`; provisioning must also serve scripts, factory leads, and t1082, so a Codex-only `--create` surface would leave the harness-neutral requirement open.
- **Implementation boundary:** reuse `materializeSessionWorktree`; no second `git worktree add`; no retired `--base`, `--from-current`, BODP, tmux, or auto-entry behavior.

### M2-M4

- RED, command:
  ```text
  $ GOCACHE=/tmp/t1070-red-cache go test ./internal/cli/worktree -run '^TestNew_' -count=1
  internal/cli/worktree/new_test.go:15:9: undefined: WorktreeCreator
  internal/cli/worktree/new_test.go:33:9: undefined: newNewCmd
  FAIL github.com/modu-ai/moai-adk/internal/cli/worktree [build failed]
  exit=1
  ```
- The first live scratch-repository probe found a real acceptance defect: an empty plain directory was accepted by Git and converted into a worktree (`plain_exit=0`). `TestWorktreeNew_RefusesPlainDirectoryBeforeGit` then reproduced it as RED: `plain-directory collision error = <nil>`, exit 1. The shared materializer now rejects every pre-existing destination with `os.Lstat` before its sole add seam.
- Final scoped regression, command and output:
  ```text
  $ unset MOAI_SESSION_WORKTREE MOAI_WORKTREE_BASE_BRANCH MOAI_KANBAN_ID MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && GOCACHE=/tmp/t1070-final-test-cache go test ./internal/cli/worktree ./internal/config ./internal/template -count=1
  ok github.com/modu-ai/moai-adk/internal/cli/worktree 7.390s
  ok github.com/modu-ai/moai-adk/internal/config 12.997s
  ok github.com/modu-ai/moai-adk/internal/template 62.094s
  ```
- Existing entry and gate regression, command and output:
  ```text
  $ GOCACHE=/tmp/t1070-final-cli-cache go test ./internal/cli -run '^(TestWorktreeNew_.*|TestSessionWorktree.*|TestCodexLaunchVerb_Worktree.*)$' -count=1
  ok github.com/modu-ai/moai-adk/internal/cli 2.388s
  ```
- Independent audit delta for REQ-WCV-003 / AC-WCV-002:
  ```text
  $ GOCACHE=/tmp/t1070-delta-red-cache go test ./internal/cli/worktree -run '^TestNew_RejectsMissingNameBeforeCreation$' -count=1
  --- FAIL: TestNew_RejectsMissingNameBeforeCreation (0.00s)
      new_test.go:58: missing-name error = "accepts 1 arg(s), received 0", want expected argument <name>
  FAIL
  exit=1

  $ GOCACHE=/tmp/t1070-delta-green2-cache go test ./internal/cli/worktree ./internal/cli -run '^(TestNew_RejectsMissingNameBeforeCreation|TestWorktreeNew_MissingNameFangErrorNamesExpectedArgument)$' -count=1
  ok github.com/modu-ai/moai-adk/internal/cli/worktree 0.146s
  ok github.com/modu-ai/moai-adk/internal/cli 0.728s

  $ /tmp/t1070-moai-delta worktree new
  ERROR
  Expected argument <name>, received 0.
  exit=1
  ```
  The first independent audit therefore failed on one merge-blocking diagnostic-contract defect. The custom argument validator and root/Fang stderr regression test close that defect; delta re-audit is running.
- Post-fix scoped regression, command and output:
  ```text
  $ unset MOAI_SESSION_WORKTREE MOAI_WORKTREE_BASE_BRANCH MOAI_KANBAN_ID MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && GOCACHE=/tmp/t1070-delta-final-test-cache go test ./internal/cli/worktree ./internal/config ./internal/template -count=1
  ok github.com/modu-ai/moai-adk/internal/cli/worktree 5.803s
  ok github.com/modu-ai/moai-adk/internal/config 12.021s
  ok github.com/modu-ai/moai-adk/internal/template 60.417s

  $ GOCACHE=/tmp/t1070-delta-final-cli-cache go test ./internal/cli -run '^(TestWorktreeNew_.*|TestSessionWorktree.*|TestCodexLaunchVerb_Worktree.*)$' -count=1
  ok github.com/modu-ai/moai-adk/internal/cli 2.027s
  ```
- Coverage, command and output:
  ```text
  $ GOCACHE=/tmp/t1070-cover-cache go test ./internal/cli/worktree -coverprofile=/tmp/t1070-worktree-cover.out -count=1
  ok github.com/modu-ai/moai-adk/internal/cli/worktree 4.058s coverage: 87.3% of statements
  newNewCmd 90.9%
  validateNewWorktreeName 80.0%
  ```
- Static quality: `go vet ./internal/cli/worktree ./internal/cli ./internal/config` produced no output, exit 0; `golangci-lint run` with isolated Go and linter caches printed `0 issues.`, exit 0; `git diff --check` produced no output, exit 0.
- Strict SPEC lint returned only the pre-existing informational `OwnershipTransitionUnmeasured` item for plan commit `df26e519d`; exit 0.
- Rule deployment parity: `cmp -s` returned 0 for each project/template pair: `kanban-dispatch.md`, `kanban-dispatch-detail.md`, `worktree-integration.md`, and `branch-origin-protocol.md`.
- Final live scratch repository `/private/tmp/t1070-live2.LF5jea`: `moai worktree new probe-card` created `.claude/worktrees/probe-card` on branch `probe-card`, with primary and worktree HEAD both `b881ff2ef96b5701ea86d0807084140dc6a4babc`; duplicate tree, pre-existing plain directory, existing branch, and `../escape` each exited 1. `git worktree list --porcelain` contained only the primary tree and `probe-card`; the plain directory remained empty and unregistered.
- Full `go test ./internal/cli -count=1 -timeout 120s` was attempted and is **not a pass**: existing environment-sensitive `TestBinaryLag_OneSeamServesBothSurfaces` and `TestCodexLive_ReviewStartBaseBranchIsNotRejected` failed, doctor/network subprocesses remained active, and the package hit the 120-second timeout. The acceptance-scoped tests above pass independently; full-suite CI remains a gap.

## §E.3 Run-phase Audit-Ready Signal

- Status: audit-ready.
- Requirements: REQ-WCV-001..009 implemented.
- Acceptance: AC-WCV-001..008 have direct automated or live evidence above.
- Changed behavior is limited to the explicit creation verb, pre-add collision refusal in the shared materializer, and the doctrine/template update. Existing launcher entry remains resolve-only for Codex and default-OFF auto-entry remains unchanged.
- Independent sync audit: initial verdict FAIL on the missing-argument diagnostic contract; the finding was fixed with RED/GREEN and actual-binary evidence above. Delta re-audit returned binding PASS with no merge-blocking finding.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec_id: SPEC-WORKTREE-CREATE-VERB-001
card: t1070
sync_commit_sha: 895760239799d75700fcce0f853d39dec453d7ed
phase: sync
status: completed
sync_status: audit-ready
independent_verdict: PASS
merge_blocking_findings: 0
initial_finding:
  id: F1
  severity: P2
  requirement: REQ-WCV-003 / AC-WCV-002
  disposition: CLOSED
optional_debt:
  - "F2: root-to-real-Git automated E2E is absent; actual-binary live smoke covers the path and the auditor judged it non-blocking."
artifacts:
  - .moai/reports/t1070/verdict.md
```

- Independent delta audit rebuilt the binary and observed both zero-argument and two-argument calls exit 1 with `Expected argument <name>, received N.`
- The auditor independently reran the changed unit/Fang tests, relevant worktree/config/CLI regressions, `go vet`, gofmt, and `git diff --check`; all passed.
- Full `internal/cli` and repository-wide CI remain explicitly outside the local PASS claim; the earlier attempted unfiltered package run failed on environment-sensitive tests and timed out, as recorded above.
- Lifecycle amendment close: reopen commit `e142a2fd7` recorded the skipped-edge defect as `completed → in-progress`; this sync commit re-applies `in-progress → completed`. The preserved `sync_commit_sha` points to the original evidence-bearing close because a commit cannot cite its own SHA.
