## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-19
audited_sha: aa3f9a79b484ac4fb64f13abc02a932ec2876b89
plan_audit_verdict: PASS-WITH-DEBT 0.87 (iteration 8, delta audit, codex required) — .moai/reports/t1640/plan-audit-iter8.md (local, gitignored)
plan_audit_receipt: rcpt-9948420d62885d354aeb20dc
operator_rulings: d-20261019T032248Z-5a94 (fifth ceiling exception, delta audit; supersedes d-20261019T031546Z-4a0f), d-20261019T024810Z-fdbb (fourth ceiling exception), d-20261019T022135Z-80d3 (ceiling exception), d-20261019T021509Z-c887 (ruling), d-20261019T015400Z-20cd (ceiling exception), d-20261019T012940Z-74f1 (ceiling exception)

debts_dispose_in_run: iter7 optional items, turned by the operator ruling into run-phase fix conditions checked at sync audit
- D2: gate-check regex repeated-field bound ([^;]|;[^;])* plus fixture (xvii) in AC-WHR-021
- D3: design.md D-11 states a reverse move writes no pending line
- D4: no-token lookup wording aligned between spec.md:100 and design.md:123
- D5: emitted §E.1 gains plan_status (done here)
- D6: bundling / implementation detail removed from REQ text
- D7: iter6 optional items (pending re-key ident check, update-time spent-budget AC)
- D8: re-pin the acceptance.md document pin at Kickoff

Sections §E.2 and later are appended by the run-phase specialist; M0 only appends §E.2 (per plan.md).

## §E.2 Run-phase Evidence

### M0 — Baseline (RED-now cells)

Measured by the run-phase specialist (manager-develop, cycle_type tdd), milestone M0, card t1640. No Go command was run (a Go-test restraint window is in force); every cell below is a read-only grep, `ls`, `wc`, `git` query or `moai` CLI call, each run as its own single invocation from the worktree root `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640`.

Tree SHA for every cell of this subsection (the PARENT commit of the baseline commit that carries these lines, never the carrying commit): `8cedf9513b05482ac873346356f8a84665a416c7`. Its relation to the document pin of acceptance.md (`6f63a1494b11469cb8109caf0917e87b5a5ceb2f`): `git diff --name-only 6f63a1494b HEAD` printed exactly the six SPEC artifacts (`acceptance.md`, `design.md`, `plan.md`, `progress.md`, `research.md`, `spec.md` of this SPEC), so no production file differs between the two trees and the acceptance.md observations remain comparable.

Exit-code convention. The Bash tool prints an `Exit code N` line for a non-zero exit of a non-grep command, and prints nothing for a `grep` that exits 1. A cell is therefore marked `observed` when the tool showed the code (`Exit code 1` printed, or a command that produced output and showed no error status for exit 0), and `derived` when it is a grep with empty stdout (exit 1 derived) or a `grep -c` that printed `0` (exit 1 derived, empty stderr). Nothing here is `not observed`.

Judging build for the one CLI cell (verification-claim-integrity §2.2): `moai version` printed `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z; `git merge-base --is-ancestor 7705b365e2 HEAD` printed nothing and showed no error status (exit 0, observed), so that build is a strict ancestor of the tree and its CLI result describes older code. It was cross-checked against the tree's source: `internal/cli/worktree/root.go:41-53` registers `newNewCmd, newSyncCmd, newRemoveCmd, newCleanCmd, newSweepCmd, newRecoverCmd, newDoneCmd, newHoistCmd, newGuardSnapshotCmd, newGuardVerifyCmd, newGuardRestoreCmd` and no list verb. Every other cell is a grep, `ls` or `wc`, independent of that binary.

#### AC-WHR-001 — home-root layout absent from production code
- command: `grep -rln --exclude='*_test.go' 'worktrees/projects' internal`
- exit: 1 (derived: grep with empty stdout)
- stdout:
```
(empty)
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: no production code names the home-root layout, so the resolver that returns the home path does not exist (green path M1).
- vs acceptance.md: identical (exit 1, empty).

#### AC-WHR-002a — owner marker code absent
- command: `grep -rln --exclude='*_test.go' 'moai-project.json' internal`
- exit: 1 (derived: grep with empty stdout)
- stdout:
```
(empty)
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: no code names the `.moai-project.json` owner marker (green path M1).
- vs acceptance.md: identical.

#### AC-WHR-004 — operator key pin absent
- command: `grep -rn --exclude='*_test.go' 'worktree_project_key' internal .moai/config`
- exit: 1 (derived: grep with empty stdout)
- stdout:
```
(empty)
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: the config key `git_strategy.worktree_project_key` does not exist in code or in the live config (green path M1).
- vs acceptance.md: identical.

#### AC-WHR-005a — shared materializer builds its destination under the in-repo MoAI root
- command: `grep -n 'destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)' internal/cli/session_worktree.go`
- exit: 0 (observed: output printed, no error status)
- stdout:
```
223:	destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: `materializeSessionWorktree` still computes `destDir` from the in-repo `.moai/worktrees` root; the work replaces it with the resolver (green path M2: exit 1, empty stdout).
- vs acceptance.md: identical (line 223, same text, a tab after the colon).

#### AC-WHR-005b — launcher `-w <n> --branch` builds its destination under the legacy Claude root
- command: `grep -n 'filepath.Join(projectRoot, ".claude", "worktrees", name)' internal/cli/worktree_branch_flag.go`
- exit: 0 (observed: output printed, no error status)
- stdout:
```
109:	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: `launcherWorktreeMaterializeReal` still places the tree under `.claude/worktrees`; M2 changes only `treePath` to the resolver (green path M2: exit 1, empty stdout).
- vs acceptance.md: identical (line 109, same text).

#### AC-WHR-010 — `moai worktree list` verb absent
- command: `moai worktree list --json`
- exit: 1 (observed: the tool printed `Exit code 1`)
- stdout:
```
(empty)
```
- stderr (verbatim words; the tool's box padding and trailing spaces are squeezed):
```
ERROR
Unknown flag: --json.
Try --help for usage.
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7 (judging build `moai_cp/20261007_195927-1888-g7705b365e2`, a strict ancestor of the tree; see the provenance paragraph above, source cross-check `root.go:41-53`)
- red because: the `list` verb does not exist, so the parent command rejects the flag (green path M4, run by a binary built from the run-phase tree).
- vs acceptance.md: identical (exit 1, same error words, same judging build).

#### AC-WHR-012b — move-list writer and reader absent
- command: `grep -rln --exclude='*_test.go' "move-list\|movelist\|move_list" internal`
- exit: 1 (derived: grep with empty stdout)
- stdout:
```
(empty)
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: neither a writer nor a reader of `move-list.jsonl` exists (green path M1 reader, M6 writer).
- vs acceptance.md: identical.

#### AC-WHR-013 — shared classifier file absent
- command: `ls internal/cli/worktree/classify.go`
- exit: 1 (observed: the tool printed `Exit code 1`)
- stdout:
```
(empty)
```
- stderr:
```
ls: internal/cli/worktree/classify.go: No such file or directory
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: the shared classifier does not exist (green path M1 function, M3/M4/M6 consumers).
- vs acceptance.md: identical.

#### AC-WHR-019 (first cell) — update consolidation's destination expression is the in-repo root
- command: `grep -c 'Join(root, sessionWorktreeSubdir)' internal/cli/update_worktree_migration.go`
- exit: 0 (observed: output printed, no error status)
- stdout:
```
1
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: the update-time consolidation's destination (`:88`, `newRoot := filepath.Join(root, sessionWorktreeSubdir)`, see the `sessionWorktreeSubdir` sweep below) is the in-repo `.moai/worktrees` root (green path M6: exit 1, stdout `0`).
- vs acceptance.md: identical (`1`, exit 0).

#### AC-WHR-025 — home-root location absent from the guidance, with its baseline and template cells
- command: `grep -c 'worktrees/projects' AGENTS.md`
- exit: 1 (derived: grep -c printed `0`, empty stderr)
- stdout:
```
0
```
- tree SHA: 8cedf9513b05482ac873346356f8a84665a416c7
- red because: the home-root location text is absent from the repository guidance (green path M7: both counts at least 1).
- vs acceptance.md: identical.
- supplementary (same acceptance cell, template mirror): command `grep -c 'worktrees/projects' internal/template/templates/AGENTS.md.tmpl`; exit 1 (derived: grep -c printed `0`, empty stderr); stdout `0`; tree SHA as above; identical to acceptance.md and research.md §R15.
- supplementary (budget baseline, not the red): command `wc -c AGENTS.md`; exit 0 (observed: output printed, no error status); stdout `   23780 AGENTS.md`; tree SHA as above; equals the ceiling of clause (c), identical to research.md §R15.

#### Legacy-reader inventory re-count (research.md §R3), tree 8cedf9513b05482ac873346356f8a84665a416c7
All commands exit 0 (observed: output printed, no error status). The count column is the number of stdout lines, counted by reading.

Site check. Command: `grep -rn -E --include='*.go' --exclude='*_test.go' 'func (\([a-zA-Z]+ \*?[A-Za-z]+\) )?(cardIDFromPath|cardIDFromRegisteredWorktree|autoOwnerWorktrees|DefaultWorktreeBases|resolveLaunchWorktreePath|resolveWorktreeLaunchTarget|seedAdmittedWorktreeHooksFor|ccWorktreeWriterPrecheck|resolveCodexWorktreeDir|validateWorktreeL2Value|cleanupMoaiWorktrees|isL1WorktreePath|sweepTierOf|runUpdateWorktreeMigration|RecordCardWorktree|PreviousCardWorktree|factoryCardWorktreeDir|factoryRefuseForeignWorktree|factoryCarryPreviousWorktree|removeWorktree)\(' internal`. Count: 20 (the 13 rows of §R3 name these 20 functions). Stdout, verbatim:
```
internal/homestate/card_worktree.go:20:func (f *FactoryDB) RecordCardWorktree(ctx context.Context, runID, cardID, path, actor string, now time.Time) (Card, error) {
internal/homestate/card_worktree.go:73:func (f *FactoryDB) PreviousCardWorktree(ctx context.Context, cardID, excludeRunID string) (string, bool, error) {
internal/cli/factory_card.go:499:func factoryCardWorktreeDir(root, cardID string) string {
internal/cli/factory_card.go:512:func factoryRefuseForeignWorktree(root, cardID string) error {
internal/cli/factory_card.go:528:func factoryCarryPreviousWorktree(ctx context.Context, db *homestate.FactoryDB, root, runID, cardID string) (string, error) {
internal/cli/launch_chain.go:60:func resolveLaunchWorktreePath(target string) (string, bool) {
internal/cli/session_worktree.go:1014:func ccWorktreeWriterPrecheck(read worktreeRead) error {
internal/cli/todo_auto.go:88:func autoOwnerWorktrees(root, cardID string) []string {
internal/cli/factory_anchor_guard.go:101:func resolveWorktreeLaunchTarget(root, value string) string {
internal/cli/launcher.go:449:func cleanupMoaiWorktrees(projectRoot string) string {
internal/cli/launcher.go:553:func removeWorktree(projectRoot, worktreeName string) error {
internal/cli/launcher.go:940:func validateWorktreeL2Value(read worktreeRead, warn io.Writer) error {
internal/cli/update_worktree_migration.go:236:func runUpdateWorktreeMigration(root string, dryRun bool, out io.Writer) error {
internal/cli/codex_launcher.go:474:func resolveCodexWorktreeDir(projectRoot, value string) (string, error) {
internal/cli/codex_hooks_seed.go:62:func seedAdmittedWorktreeHooksFor(read worktreeRead, warn io.Writer) {
internal/cli/worktree/sweep.go:574:func sweepTierOf(path string) string {
internal/cli/worktree/done.go:230:func isL1WorktreePath(targetPath string) bool {
internal/hook/session_start_factory.go:231:func cardIDFromRegisteredWorktree(dir string) string {
internal/hook/session_start_record.go:223:func cardIDFromPath(dir string) string {
internal/factory/status_read.go:63:func DefaultWorktreeBases(primaryRoot string) WorktreeBases {
```
Comparison with §R3: all 20 functions exist at the cited files, and every cited range still contains its function declaration, except that `cardIDFromPath` is declared at line 223 while §R3 row 1 cites `226-241` (the cited range starts three lines after the declaration, so it reads as the body range; no material drift) and `resolveWorktreeLaunchTarget` (101, cited `98-106`) and `ccWorktreeWriterPrecheck` (1014, cited `1010-1033`) sit inside their cited ranges. The bodies and the interior line numbers of §R3 rows 4 (loop `:117-143`), 10 (`:87,105-108`) and 13 (call sites `:1152`, `:1162`, `:1450`, `:1455`) were not re-read in M0 (see Gaps).

Literal-pair sweep (the AC-WHR-007 patterns, production Go only):
- command: `grep -rn --include='*.go' --exclude='*_test.go' '".claude", "worktrees"' internal`; count: 10; stdout:
```
internal/cli/launch_chain.go:71:	return filepath.Join(root, ".claude", "worktrees", target), true
internal/cli/factory_anchor_guard.go:105:	return filepath.Join(shortWorktreeRoot(root), ".claude", "worktrees", value)
internal/cli/codex_hooks_seed.go:73:		tree = filepath.Join(root, ".claude", "worktrees", value)
internal/cli/launcher.go:456:	localBase := filepath.Join(projectRoot, ".claude", "worktrees")
internal/cli/launcher.go:956:		acceptedPrefixes = append(acceptedPrefixes, filepath.Join(root, ".claude", "worktrees"))
internal/cli/todo_auto.go:90:		filepath.Join(root, ".claude", "worktrees", cardID),
internal/cli/worktree_branch_flag.go:109:	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)
internal/cli/worktree/done.go:237:		filepath.Join(mainRoot, ".claude", "worktrees"),
internal/hook/worktree_create.go:41:var agentWorktreeParentDir = filepath.Join(".claude", "worktrees")
internal/factory/status_read.go:64:	bases := WorktreeBases{Claude: filepath.Join(primaryRoot, ".claude", "worktrees")}
```
- command: `grep -rn --include='*.go' --exclude='*_test.go' '".moai", "worktrees"' internal`; count: 3; stdout:
```
internal/cli/todo_auto.go:91:		filepath.Join(root, ".moai", "worktrees", cardID),
internal/cli/launcher.go:957:		acceptedPrefixes = append(acceptedPrefixes, filepath.Join(root, ".moai", "worktrees"))
internal/cli/worktree/done.go:238:		filepath.Join(mainRoot, ".moai", "worktrees"),
```
- command: `grep -rn --include='*.go' --exclude='*_test.go' 'Base(parent) == "worktrees"' internal`; count: 1; stdout:
```
internal/hook/session_start_factory.go:241:		if filepath.Base(parent) == "worktrees" {
```
- command: `grep -rn --include='*.go' --exclude='*_test.go' 'sessionWorktreeSubdir' internal`; count: 6 (one comment, one constant declaration, four uses); stdout:
```
internal/cli/session_worktree.go:46:// sessionWorktreeSubdir is MoAI's harness-neutral worktree landing directory
internal/cli/session_worktree.go:48:const sessionWorktreeSubdir = ".moai" + string(filepath.Separator) + "worktrees"
internal/cli/session_worktree.go:223:	destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)
internal/cli/factory_card.go:500:	return filepath.Join(root, sessionWorktreeSubdir, cardID)
internal/cli/codex_launcher.go:491:		modern := filepath.Join(projectRoot, sessionWorktreeSubdir, value)
internal/cli/update_worktree_migration.go:88:	newRoot := filepath.Join(root, sessionWorktreeSubdir)
```
Sweep reading: 14 literal-pair lines (10 + 3 + 1) and 4 `sessionWorktreeSubdir` uses, 18 production lines in all; `internal/hook/worktree_create.go:41` is the agent-parent constant that decision `dab5` keeps. The legacy-path builders of §R3 rows 2, 3, 4, 7, 8, 9, 10 and 13, and three of the four row-5 files (`launch_chain.go`, `factory_anchor_guard.go`, `codex_hooks_seed.go`), appear in this output at the cited files; `ccWorktreeWriterPrecheck` (row 5) and `cardIDFromPath` (row 1) do not appear in this output (their bodies were not read in M0, so how they spell the path is not claimed here). This is the baseline for the AC-WHR-007 allow-list sweep at M3 and is not itself a gate cell.

#### Differences from acceptance.md / research.md §R15
None. All ten cells (and the two supplementary AC-WHR-025 cells) reproduced the §R15 values on `8cedf9513b05482ac873346356f8a84665a416c7` byte for byte (stdout and exit code). The §R3 re-count found no missing reader and no site moved out of its cited range (one start-line offset of 3 lines for `cardIDFromPath`, immaterial).

#### Five-section summary (verification-claim-integrity §3)

**Claim.** (1) The ten RED-now cells of AC-WHR-001, 002a, 004, 005a, 005b, 010, 012b, 013, 019 (first cell) and 025 are red on the pre-change tree for the stated reasons, with the values acceptance.md quotes. (2) The 20 legacy-reader functions of research.md §R3 and the literal-pair sites listed above exist on this tree. (3) Nothing outside `progress.md` was touched by M0.

**Evidence.** The commands and verbatim stdout above, each run as its own single invocation in this run; exit codes as marked observed or derived; judging-build provenance and its source cross-check in the provenance paragraph.

**Baseline-attribution.** Measured in this run against tree `8cedf9513b05482ac873346356f8a84665a416c7` (the parent of the baseline commit); the six-file SPEC-only difference to the acceptance.md pin `6f63a1494b11469cb8109caf0917e87b5a5ceb2f` is shown by `git diff --name-only 6f63a1494b HEAD`. The one tool measurement (`moai worktree list --json`) was judged by build `moai_cp/20261007_195927-1888-g7705b365e2`, a strict ancestor of the tree (`git merge-base --is-ancestor`, exit 0), with the tree's own source read as cross-check.

**Gaps.** (a) AC-WHR-021 is not recorded here by design (M-1 owns it). (b) Every other RB* cell needs a Go test and is recorded at the RED commit of the milestone that owns it. (c) Six exit codes are derived from stdout, not shown by the tool: the four empty-stdout greps (AC-WHR-001, 002a, 004, 012b) and the two `grep -c` cells that printed `0` (AC-WHR-025 on `AGENTS.md` and on the template); none is `not observed`. (d) §R3 row 8's "`git worktree list | grep -c '/worker-'` prints 0" was not re-counted: it is a piped command outside the single-invocation form and `git worktree list` is a roughly 200-line listing; the claim stays as measured at `e5950b354f`. (e) The interior line numbers of §R3 rows 4, 10 and 13 were not re-read. (f) No Go command was run (restraint window), so no build, vet or test result is claimed; `GOOS=windows` build is not part of M0.

**Residual-risk.** The grep red cells prove the literals are absent, not that a differently spelled implementation will satisfy the green path (acceptance.md states this per cell, and each green path carries a Go test). The `moai worktree list --json` cell was produced by a build older than the tree; the source cross-check (`root.go:41-53`) closes that for this tree. Exit codes derived from stdout hold only while the harness behavior stated in acceptance.md (§B header) holds. The `.moai/reports/*` files of M-1 are gitignored and local, so the tracked `progress.md` is the only carrier of these cells.

## §G Override and Refusal Record

- 2026-10-19T03:29:15Z SPEC-WORKTREE-HOME-ROOT-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter8.md
- 2026-10-19T03:33:35Z SPEC-WORKTREE-HOME-ROOT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter8.md
- 2026-10-19T03:34:01Z SPEC-WORKTREE-HOME-ROOT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter8.md
