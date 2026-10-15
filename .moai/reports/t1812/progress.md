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

The skip is ONE anchored regex `^(…)$` carrying the 7 names plus the SessionWorktree exact-match names (41, from `go test ./internal/cli -list SessionWorktree`, which includes `TestEnterSessionWorktree_ProfileSubcommand`), as the dispatch required. Remeasure scope `./internal/cli`, `-timeout=150m`, output tee'd, `wc -l ~/.gitconfig` read immediately before and after.

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

Both sides show the same 7 top-level failures with the same messages (the A side prints each message before its `--- FAIL` header because it came from a `-v` run; the B side from a plain run).
