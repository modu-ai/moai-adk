# t1812 progress — TestCodexReviewGate macOS /private/var cwd mismatch

Card: t1812 (Class B, run → sync). Lane: lane-38. Tree: `.moai/worktrees/t1812`, branch `WT-codex-gate-tmpdir-symlink`, base `3a8689d6b`.

## §1 Cause determination (recorded BEFORE any repair)

**Verdict: the test expectation is wrong; the production path resolution is intentional. No code regression.**

### Reproduction (this run, this tree)

Command: `go test ./internal/cli -run 'TestCodexReviewGatePrimaryCheckoutSkip|TestCodexReviewGate_TreeScopeRequestShapeUnchanged|TestCodexReviewGate_StaleEnvKeepsTreeScope' -count=1`

Observed (verbatim, trimmed to the deciding lines):

```
--- FAIL: TestCodexReviewGatePrimaryCheckoutSkip
    codex_review_gate_primary_scope_red_test.go:84: ... got dir "/private/var/folders/.../001" want "/var/folders/.../001"
--- FAIL: TestCodexReviewGate_TreeScopeRequestShapeUnchanged
    codex_review_scope_test.go:87: thread/start cwd = "/private/var/folders/.../001", want the resolved tree "/var/folders/.../001"
--- FAIL: TestCodexReviewGate_StaleEnvKeepsTreeScope
    codex_review_scope_test.go:280: tree-scope cwd = "/private/var/folders/.../001", want the resolved tree "/var/folders/.../001"
```

### Why

- `t.TempDir()` on macOS returns `/var/folders/...`, where `/var` is a symlink to `/private/var`.
- The three tests compare the gate's cwd / row dir against the raw fixture root (`f.root`, `f.primary`, `root`) — the unresolved spelling — while the assertion message itself says "the resolved tree".
- The producer anchors the tree scope on the git toplevel, which git reports symlink-resolved. Sibling tests in `internal/cli/codex_review_ownership_test.go` (lines 367-376, 466-477) already resolve the expected value with `filepath.EvalSymlinks` before comparing, with the comment "resolve before comparing (macOS /var symlink)". `reviewScopeGitPath` / `reviewScopeEvalPath` (`codex_review_scope.go`) are deliberate symlink-resolution (card t1404 N4).
- So the three tests are the only comparison sites that never resolved the expected side. Linux CI (no `/var` symlink) cannot see it.

### Gaps

- Not measured: Linux/Windows run of these tests (this lane runs on darwin only; the change is a no-op where `EvalSymlinks` is the identity).
- Not re-checked here: whether the failure also reproduces at the card's cited base `e04ebd5ca`; this run measured the current tree `3a8689d6b` only.
- Refusal: one compound `cd … && wc && go test … > file; …` Bash command was refused by the worktree-session guard ("too complex to verify"); re-run as plain separate commands, results above.

## §2 Pre-run gitconfig line count

`wc -l ~/.gitconfig` before the narrow 3-test run: 484 (the leader's baseline was 482).

## §3 Repair (run)

Three expected-side comparisons now resolve the fixture root with `filepath.EvalSymlinks` before comparing (same pattern as `codex_review_ownership_test.go`): `codex_review_scope_test.go` (TreeScopeRequestShapeUnchanged, StaleEnvKeepsTreeScope), `codex_review_gate_primary_scope_red_test.go` (PrimaryCheckoutSkip). No production file changed.

Evidence (this run, tree `3a8689d6b` + the 3 edits):
- `go test ./internal/cli -run 'TestCodexReviewGate' -count=1 -v` → 21 PASS lines incl. the 3 target tests, `ok ... 8.087s`.
- `gofmt -l internal/cli` and `go vet ./internal/cli` → no output.
- `go test ./internal/template -count=1` → `ok ... 243.716s`.
- `wc -l ~/.gitconfig` → 484 before and after (leader baseline 482; the +2 predates this card's runs).

## §4 Sync-scope subset: `go test ./internal/cli -run 'Mirror|Parity|Doc'` = FAIL (pre-existing, 7 tests)

Failing: `TestCheckCodexWiring_{DanglingMirrorEntriesCounted,MirrorAbsentAdvisesRedeploy,MirrorFindingParticipatesInTailDrop,MirrorSummaryWidth,MirrorUnreadableIndeterminate,MirrorUsesExistingRowTwoRegisters}` and `TestHookEntryParity_NoDivergenceEitherDirection` (template `UserPromptSubmit/handle-user-prompt-submit.sh timeout=5` vs project `timeout=15`).

A/B: with my two test files stashed (tracked edits only), the same 7 tests, run by name, fail with identical messages. So they are independent of this card's change. Not diagnosed further: cause (Codex mirror = host `~/.codex` state is a candidate; hook parity = project settings vs template drift) — a separate card, not t1812.

Gaps: the full `./internal/cli` suite (SessionWorktree skip) was not run — the go-test-heavy slot turn is lane-38's last; the change is test-only in 2 files covered by the targeted runs above. The subset's remaining tests other than the 7 passed in the same run (only the 7 `--- FAIL` top-level names appear).

## §5 Landing record: carried-failure `-skip` (leader ruling, d8 cause units)

Leader ruling (cross-session message to lane-38): the 7 failures may go into the remeasure `-skip`; brain interpretation d8 (cause unit proven by A/B, cap 3) fits. Two cause units, both already carded, no new card:

1. `TestCheckCodexWiring_Mirror*` x6 — card t1813; appears to depend on the real HOME `.agents`/`.codex` state (the leader's reading; not diagnosed by this lane).
2. `TestHookEntryParity_NoDivergenceEitherDirection` — resolves with the t1775 landing.

The skip is ONE anchored regex `^(…)$` carrying the 7 names plus the SessionWorktree exact-match names (43 — counted with `go test ./internal/cli -list SessionWorktree | grep -c '^Test'`; an earlier draft of this paragraph said 41 without counting, which was wrong — from `go test ./internal/cli -list SessionWorktree`, which includes `TestEnterSessionWorktree_ProfileSubcommand`), as the dispatch required. Remeasure scope `./internal/cli`, `-timeout=150m`, output tee'd, `wc -l ~/.gitconfig` read immediately before and after.

### A side — this card tree WITH the fix (HEAD `3ac641fc2`), `go test ./internal/cli -run 'Mirror|Parity|Doc' -v`, failure lines verbatim

```
    doctor_codex_test.go:1267: absent mirror on a wired project status = ok, want Warn: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy (0.01s)
    doctor_codex_test.go:1292: dangling entries status = ok, want Warn: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_DanglingMirrorEntriesCounted (0.01s)
    doctor_codex_test.go:1437: Detail does not record the indeterminate condition: ""
--- FAIL: TestCheckCodexWiring_MirrorUnreadableIndeterminate (0.01s)
    doctor_codex_test.go:1467: premise broken — expected exactly one mirror finding: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
    doctor_codex_test.go:1467: premise broken — expected exactly one mirror finding: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MirrorSummaryWidth (0.02s)
    --- FAIL: TestCheckCodexWiring_MirrorSummaryWidth/absent (0.01s)
    --- FAIL: TestCheckCodexWiring_MirrorSummaryWidth/dangling (0.01s)
    doctor_codex_test.go:1523: premise broken — the joined summaries did not exceed 113 runes, so no tail-drop occurred: "~/.codex/config.toml: 1 stale skill entry"
    doctor_codex_test.go:1553: premise broken — expected the lone mirror finding: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop (0.02s)
    --- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop/tail_drop_keeps_the_lead_and_detail_keeps_everything (0.01s)
    --- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop/lead_summary_exception_stays_unreachable_for_mirror_summaries (0.01s)
    doctor_codex_test.go:1578: summary register empty of the mirror observation: "wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical)"
    doctor_codex_test.go:1581: detail register carries no fuller mirror text: ""
--- FAIL: TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters (0.01s)
    hook_entry_parity_test.go:172: template-only (registered in the template, missing from this project): UserPromptSubmit/handle-user-prompt-submit.sh timeout=5 async=false
    hook_entry_parity_test.go:175: project-only (registered in this project, absent from the template): UserPromptSubmit/handle-user-prompt-submit.sh timeout=15 async=false
    hook_entry_parity_test.go:178: hook-entry parity: 1 template-only, 1 project-only
--- FAIL: TestHookEntryParity_NoDivergenceEitherDirection (0.00s)
```

### B side — base test files (the two t1812 test edits stashed; production identical), the 7 names run by `-run`, verbatim

```
--- FAIL: TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy (0.01s)
    doctor_codex_test.go:1267: absent mirror on a wired project status = ok, want Warn: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_DanglingMirrorEntriesCounted (0.01s)
    doctor_codex_test.go:1292: dangling entries status = ok, want Warn: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MirrorUnreadableIndeterminate (0.01s)
    doctor_codex_test.go:1437: Detail does not record the indeterminate condition: ""
--- FAIL: TestCheckCodexWiring_MirrorSummaryWidth (0.02s)
    --- FAIL: TestCheckCodexWiring_MirrorSummaryWidth/absent (0.01s)
        doctor_codex_test.go:1467: premise broken — expected exactly one mirror finding: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
    --- FAIL: TestCheckCodexWiring_MirrorSummaryWidth/dangling (0.01s)
        doctor_codex_test.go:1467: premise broken — expected exactly one mirror finding: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop (0.02s)
    --- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop/tail_drop_keeps_the_lead_and_detail_keeps_everything (0.01s)
        doctor_codex_test.go:1523: premise broken — the joined summaries did not exceed 113 runes, so no tail-drop occurred: "~/.codex/config.toml: 1 stale skill entry"
    --- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop/lead_summary_exception_stays_unreachable_for_mirror_summaries (0.01s)
        doctor_codex_test.go:1553: premise broken — expected the lone mirror finding: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters (0.01s)
    doctor_codex_test.go:1578: summary register empty of the mirror observation: "wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical)"
    doctor_codex_test.go:1581: detail register carries no fuller mirror text: ""
--- FAIL: TestHookEntryParity_NoDivergenceEitherDirection (0.01s)
    hook_entry_parity_test.go:172: template-only (registered in the template, missing from this project): UserPromptSubmit/handle-user-prompt-submit.sh timeout=5 async=false
    hook_entry_parity_test.go:175: project-only (registered in this project, absent from the template): UserPromptSubmit/handle-user-prompt-submit.sh timeout=15 async=false
    hook_entry_parity_test.go:178: hook-entry parity: 1 template-only, 1 project-only
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.362s
```

### Two more names in the same `-skip` (leader message, no A/B run by this lane)

`TestRunWeb_AutoEntryOn_SuppressesAdvisory` and `TestWorktreeNew_WiresTheSharedMaterializer` are added to the one regex at the leader's instruction: the leader states they write to the real `~/.gitconfig` (confirmed on the leader's side; cause count not computed). This lane did NOT run them and did NOT verify that claim — they are skipped on the leader's word, as a Gap, not as a measured fact. Both names exist (`go test ./internal/cli -list` printed both). Final skip regex = 43 SessionWorktree names + 7 carried + these 2 = 52 names (51 `|` separators in the saved pattern file).

### First remeasure result and the two names added after it

First remeasure (started 21:05Z, ended about 22:05Z; skip = 52 names) exited 1: the remeasure's own record message named one failing test, `TestCountCodexAgentTOMLs_BadPatternRoot`, and said the record is invalid until a green re-measure. Evidence for the cause, observed this run: run alone it fails with `codex_readiness_test.go:545: count = 6, want 0`; `countCodexAgentTOMLs` (`internal/cli/codex_readiness.go:219`) counts the real HOME's `~/.codex/agents/*.toml` before the project root, and `ls ~/.codex/agents/*.toml | wc -l` printed 6. No A/B by this lane; the mechanism was read from the source and the HOME count.

Leader ruling (cross-session message): known family, card t1817 owns the cause, proven by A/B on t1768; the cause count stays within the db7d cap of 3 (units: Mirror/HOME state, hook parity via t1775, real-HOME codex agents), so an exact-match skip of `TestCountCodexAgentTOMLs` and `TestCountCodexAgentTOMLs_BadPatternRoot` is approved, plus one more remeasure with the `-json` output kept in a file and the failing-test count taken with jq. Final skip regex = 54 names.

`~/.gitconfig` around the first remeasure: before sha b68571080eb1d2cf67a10e381bfd5d94cb3d1e7b3967470ffba896b2bc628e14 (486 lines, snapshot about a minute after the run started); after sha 3596c79456fdf394c7a670f1820dcf8c82adc1612f96405ebd52de5a74ee2581 (489 lines). The three added lines are `directory = …/.moai/worktrees/t1757`, `…/t1800`, `…/t1822` (the leader confirms these are other cards' tree registrations).

Both sides show the same 7 top-level failures with the same messages (the A side prints each message before its `--- FAIL` header because it came from a `-v` run; the B side from a plain run).

## §6 Landing remeasure history, lane-4 takeover (2026-10-16 / 17)

Lane-4 took the card over from lane-38 (queue record: owner lane-4, stage merge-ready; lease re-taken after an expiry via the 8859 sequence). All runs: `moai integration remeasure -- go test -json -count=1 -timeout=150m [-skip …] ./internal/cli/`, TMPDIR/GOTMPDIR/HOME unset. Machine-local scratch holds the outputs; the integration record is the durable one.

| # | Window (UTC) | Tree state | Skips | Result |
|---|---|---|---|---|
| 1 | 13:03:27–13:51:43 | 3bac1c8f8 (no main absorb), 54-name skip from §5 | 54 names | exit 1 — `TestCheckCodexWiring_CopyModeDetailOnly` (9th Codex wiring name) |
| 2 | 13:53:40–14:31:50 | 85496a5f4 (main a4a06d96e absorbed, incl. t1813 99f1ebd07) | none | exit 1 — `TestManagedTUINeverReachedWithoutOptIn/opted_in_probes_once` |
| 3 | 14:32:38–15:11:29 | 85496a5f4 | one subtest | exit 1 — `TestAuditLagUsesBinlagSeam` |
| 4 | 15:12:12–15:44:55 | 85496a5f4 | `^(TestAuditLagUsesBinlagSeam\|TestManagedTUINeverReachedWithoutOptIn)$` | exit 0, valid for tree 8b3c8a0ac063 (base a4a06d96e4bd) |
| 5 | 18:57:12–19:27:36 | 962d18b5e (main fb55b8029 absorbed, incl. t1792 + t1823) | none | **exit 0, valid for tree 9fe0280c39fb (base fb55b8029a4a)** |

Decisions (leader window lane-17, facts passed by message; peer answers are not operator approvals):
- Run 1 red: cause unit = real-HOME state; main already carried t1813 (99f1ebd07), our base did not. A/B this session: the test fails in this tree, passes in `.claude/worktrees/main`. Decision: absorb main, drop all Codex skips.
- Run 2 red / run 3 red: both failed identically in the main tree and this tree (A/B same session), card diff (two `codex_review*_test.go` files) overlaps neither. Messages: `managed_codex_tui_test.go:2120: probe=0 tui spawns=0 direct door=0, want 1/0/0`; `mcp_build_identity_test.go:809/814` want-map drift (`local_main_resync.go` 414→435, 417→438, `integration.go` 287→288). Both are fixed by card t1823 (f37be81f7 and the want-map commit). Run 4 therefore skipped those two exact names (the whole TUI test, since `-skip` takes one regex).
- Run 5 needed no skips after t1823 landed on main (fb55b8029).

Gaps:
- Run 4 skip list was two names; the TUI skip also removed the three passing subtests of that test. Superseded by run 5 (no skips).
- `~/.gitconfig` line counts (observed only, cause not separated, other sessions also write): 512→573 (run 1), 573→582 (run 2), 584→595 (run 3), 595→596 (run 4), 598→599 (run 5).
- The `-json` stream is held by the integration record; the lane stored only its stdout summary line.
- Refusals recorded: the permission classifier denied `mcp__moai__factory_complete` (reason "Security Test Removal", twice: after run 4 and after run 5, the second time with no skips) and `mcp__moai__factory_renew` (no reason given). Neither was routed around (no CLI, no peer). Source: this session's tool results.

Residual-risk:
- main moved on after run 5 (0bd1da497, statusline t1822 only). `complete` may refuse with "integration branch moved past the record's base"; then absorb main again and re-measure before completing.
- This file is the only edit since the merge `962d18b5e` and is deliberately NOT committed: the remeasure record is tree-keyed, so a commit here would invalidate tree 9fe0280c39fb. Commit it only together with, or after, the landing decision.

Waiting state: card t1812 is at merge-ready, lease expired/expiring, waiting for the operator to allow `mcp__moai__factory_complete` and `mcp__moai__factory_renew` in the lane-4 session. Tree `.moai/worktrees/t1812`, branch `WT-codex-gate-tmpdir-symlink`, HEAD 962d18b5e. Do not remove the tree.

## §7 Re-lease into run tn1p4e (lane-4, 2026-10-17)

Leader notice (cross-session): run tmvgvn retired 10-17 10:01Z, so stage/complete/renew fail "factory card not found" on tn1p4e; re-lease with `moai factory next --card t1812` (card t1825). No operator hold found on t1812 (status row and memory notes read; absence of a hold marker is not proof of none).

Observed this session:
- `moai factory next --card t1812` from the primary checkout: leased, `stage=-`, record restarts (tmvgvn history not carried over).
- The lease expired at 10:45:22Z before work resumed; `mcp__moai__factory_renew` returned `refused-expired`. Recovery per rule 8859: `moai factory stage t1812 run` (answered "lease expired", card back to assigned) → `moai factory next --card t1812` from the primary checkout (from inside the worktree it errors) → `moai factory stage t1812 run` → `run v7 lease renewed`, lease until 11:31:13Z.
- The earlier classifier denials of factory_complete/factory_renew were NOT reproduced for renew (it executed and returned a result); factory_complete not yet attempted in this run.

Gaps: sync-audit and the landing remeasure for tn1p4e are not re-done yet; run 5 (tree 9fe0280c39fb, §6) was recorded against tmvgvn and its validity under tn1p4e is unverified.
