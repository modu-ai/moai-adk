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

### M1 — RED (test-compilation-order cells)

Measured by the run-phase specialist (manager-develop, cycle_type tdd), milestone M1, card t1640. The RED is the test-compilation-order RED of plan.md §C: test files that name M1 symbols are committed before M1's first production commit, so each affected package fails to compile, and that compile failure — not any test failure — is the red. Two narrowed calls, one per affected package, each started only after both resource gates held (1-minute load ≤ 32 on `uptime`, memory free ≥ 50% on `memory_pressure`); no production file was created, `spec.md` was not touched, and no test was skipped. Work-state note: two earlier attempts of this task died on provider rate limits before committing; the four test files as committed here were finished from that attempt's partial edits and verified `gofmt -l`-clean (empty output, exit 0) before the calls below.

Tree SHA for every cell of this subsection (the PARENT commit of the RED commit that carries these lines, never the carrying commit): `440cbbd68931b57debd6df41947f011dd1f74cf7`. Exit codes: both observed (the harness printed `Exit code 1`). Both packages report undefined-symbol errors; the "no non-test Go files" shape did not appear — the test-only package's build compiles its `_test.go` files directly and names the undefined M1 symbols.

The first call was blocked once: at 21:47 KST the 1-minute load read 34.34 (> 32, memory free 86%), the gate failed, the call was stopped (not looped), and the miss was reported. The lane then authorized up to 5 re-gates ~60s apart; the first re-gate passed (21:51:50 KST, load 19.99, memory free 86%) and the second call ran. Call 1 itself ran on the earlier pass (21:46 KST, load 26.26, memory free 85%).

#### AC-WHR-002b (clauses 1-3) — stability: a key never changes once assigned

- command: `go test -run '^$' -p 1 -count=1 ./internal/worktreeroot/`
- gates at start: 21:46 KST, 1-minute load 26.26, memory free 85%
- exit: 1 (observed)
- stdout (verbatim, this run, this tree):
```
# github.com/modu-ai/moai-adk/internal/worktreeroot [github.com/modu-ai/moai-adk/internal/worktreeroot.test]
internal/worktreeroot/worktreeroot_test.go:145:12: undefined: Key
internal/worktreeroot/worktreeroot_test.go:154:12: undefined: Root
internal/worktreeroot/worktreeroot_test.go:201:15: undefined: PathFor
internal/worktreeroot/worktreeroot_test.go:226:15: undefined: PathFor
internal/worktreeroot/worktreeroot_test.go:240:15: undefined: PathFor
internal/worktreeroot/worktreeroot_test.go:496:12: undefined: Resolve
internal/worktreeroot/worktreeroot_test.go:503:16: undefined: PathFor
internal/worktreeroot/worktreeroot_test.go:535:15: undefined: Resolve
internal/worktreeroot/worktreeroot_test.go:542:19: undefined: PathFor
internal/worktreeroot/worktreeroot_test.go:587:15: undefined: Resolve
internal/worktreeroot/worktreeroot_test.go:587:15: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/worktreeroot [build failed]
FAIL
```
- tree SHA: 440cbbd68931b57debd6df41947f011dd1f74cf7
- red because: the clause tests (`TestWorktreeRoot_Key_StableAcrossResolutionMoveAndRename`, `TestWorktreeRoot_Key_RecordedKeySurvivesMoaiHomeRelocation`) cannot compile — the resolver functions they call (`Key`, `Root`, `PathFor`, `Resolve`) do not exist, so no assignment record can be read or kept and every clause is untestable (green path M1 GREEN).
- vs acceptance.md: the RB* row defers clauses (1)-(3) to "the RED commit of M1 ... (needs a Go test)" — this is that record; the "no assignment record exists" premise is M0's AC-WHR-002a grep cell.

#### AC-WHR-003a — case-insensitive collision between two repositories

- command, gates, exit, run: identical to the AC-WHR-002b cell (one invocation, `./internal/worktreeroot/`, exit 1); the verbatim stdout above is this cell's evidence.
- tree SHA: 440cbbd68931b57debd6df41947f011dd1f74cf7
- red because: `TestWorktreeRoot_Key_CaseInsensitiveCollisionSuffix` cannot compile on the same undefined resolver symbols; no collision logic or home layout exists to test (the acceptance.md premise "grep -rln --exclude='*_test.go' 'worktrees/projects' internal prints nothing" is M0's AC-WHR-001 cell).
- vs acceptance.md: the RB* row defers to the RED commit of M1 — recorded here.

#### AC-WHR-003b — third-owner and unreadable-marker refusal

- command, gates, exit, run: identical to the AC-WHR-002b cell; the verbatim stdout above is this cell's evidence.
- tree SHA: 440cbbd68931b57debd6df41947f011dd1f74cf7
- red because: the refusal tests (`TestWorktreeRoot_Key_RefusesWhenTheSuffixedFolderHasAThirdOwner` at `worktreeroot_test.go:496` calling the undefined `Resolve`, `...RefusesAnUnreadableMarker` at `:535`) cannot compile; refusal over a third-owned or unreadable marker is unreachable (green path M1 GREEN).
- vs acceptance.md: the RB* row defers to the RED commit of M1 — recorded here.

#### AC-WHR-023 — spelling-safe equality, resolver `Same` clause

- command, gates, exit, run: identical to the AC-WHR-002b cell; the verbatim stdout above is this cell's evidence.
- tree SHA: 440cbbd68931b57debd6df41947f011dd1f74cf7
- red because: the package build fails on undefined resolver symbols and therefore `TestWorktreeRoot_Same_SpellingSafeEquality` cannot compile; `Same` is not in the printed error list because the compiler stopped at 10 errors (`too many errors` at `worktreeroot_test.go:587`) — its undefined status past the cap is a file-read observation (defined at `:737`, referenced from `:753`), not observed compiler output. The compile failure itself is the red (plan.md §C).
- vs acceptance.md: the RB* ownership of 023's `Same` clause at M1 is plan.md §C M1; the acceptance row's other equality sites are later milestones' cells (`isL1WorktreePath` M3, `internal/session/anchor.go` M6, per research.md §R13).

#### Second observation — `internal/cli/worktree` (classifier and move-list readers; no RB* cell of this commit)

- command: `go test -run '^$' -p 1 -count=1 ./internal/cli/worktree/`
- gates at start: 21:51:50 KST, 1-minute load 19.99, memory free 86%
- exit: 1 (observed)
- stdout (verbatim, this run, this tree):
```
# github.com/modu-ai/moai-adk/internal/cli/worktree [github.com/modu-ai/moai-adk/internal/cli/worktree.test]
internal/cli/worktree/classify_test.go:79:33: undefined: ClassifyContext
internal/cli/worktree/classify_test.go:97:10: undefined: ClassifyContext
internal/cli/worktree/classify_test.go:109:58: undefined: WorktreeRecord
internal/cli/worktree/classify_test.go:112:9: undefined: WorktreeRecord
internal/cli/worktree/classify_test.go:168:11: undefined: Classify
internal/cli/worktree/classify_test.go:185:12: undefined: WorktreeRecord
internal/cli/worktree/classify_test.go:195:6: undefined: Classify
internal/cli/worktree/classify_test.go:216:9: undefined: ClassifyContext
internal/cli/worktree/movelist_test.go:197:51: undefined: MoveRequester
internal/cli/worktree/movelist_test.go:219:25: undefined: MoveRequester
internal/cli/worktree/classify_test.go:216:9: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree [build failed]
FAIL
```
- tree SHA: 440cbbd68931b57debd6df41947f011dd1f74cf7
- red because: the classifier and move-list reader tests cannot compile — `ClassifyContext`, `WorktreeRecord`, `Classify` (classify.go) and `MoveRequester` are undefined; past the compiler's cap (`ParseWorktreeList`, `TreeIdentity`, `TreeIdent`, `EnsureInstanceToken`, `MovedTo`, `MovedBy`) the same holds by file-read, not by observed output. This observation owns no RB* cell here: AC-WHR-013's and 012b's observed cells are M0's greps, AC-WHR-014b is recorded at M3's RED (research.md §R13, M1 row).
- vs acceptance.md: consistent with the M0 grep cells (classifier file absent, move-list code absent) — this is the compile-level red of the same green path.

Coverage note (one line per the milestone plan): the RED files also carry test coverage for AC-WHR-001 (`PathFor` home-root form), 002a (key derivation and owner marker), 003 (the collision-and-refusal family), 004 (operator pin), 022 (`Candidates` across three roots) and the non-`Same` clauses of 023 (`Candidates` two-spelling collapse, classifier spelling-safety, case-sensitive distinctness); their green path is the M1 GREEN production commit, and AC-WHR-022's own observed cell is recorded at M3 per research.md §R13.

#### Five-section summary (verification-claim-integrity §3)

**Claim.** (1) The four RB* cells M1 owns (AC-WHR-002b clauses 1-3, 003a, 003b, 023 `Same`) are red on the pre-production tree: each affected package fails to compile solely on undefined M1 symbols, observed verbatim above. (2) The `internal/cli/worktree` observation red is the same shape for the classifier and move-list readers. (3) No production file, `spec.md`, or other SPEC body was touched by this RED commit; the four test files are `gofmt`-clean and 39 test functions strong (1 + 12 + 7 + 19).

**Evidence.** The two command invocations and their verbatim stdout above, each run as its own single invocation in this run against tree `440cbbd68931b57debd6df41947f011dd1f74cf7`; both exit codes observed (`Exit code 1`); gate readings recorded per call; the 21:47 gate miss and the lane's re-gate authorization recorded in the preamble.

**Baseline-attribution.** Measured in this run, in this worktree (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640`), against HEAD `440cbbd68931b57debd6df41947f011dd1f74cf7` — the parent of the commit carrying these lines (plan.md §C pin rule). No tool build is involved (the `go` toolchain is external tooling, not a tree-built measurement tool; verification-claim-integrity §2.2 does not bind it).

**Gaps.** (a) The compiler caps each package's error list at 10 (`too many errors`), so identifiers past the cap — `Same`, `Candidates`, `Candidate` in the resolver package; `ParseWorktreeList`, `TreeIdentity`, `TreeIdent`, `EnsureInstanceToken`, `MovedTo`, `MovedBy` in `internal/cli/worktree` — are not individually observed as errors; their undefined status rests on file reads (surface headers and reference sites) and will be re-observed at the M1 GREEN compile. (b) `-run '^$'` ran no tests in either package (the build failed before any test binary); no test behavior, only compilation, is claimed. (c) No `go build`/`go vet` was run (dispatch constraint), so package-level buildability beyond the two named packages is carried by M0's "nothing changed" baseline plus the commit diff, not re-observed here. (d) The M1 RED commit leaves the `spec.md` `draft -> in-progress` flip to M1's first production commit (plan.md §C), so the flip is not observed here by design.

**Residual-risk.** An undefined symbol hiding past the 10-error cap that is NOT an M1 symbol would make the red wider than "solely M1 symbols"; the static collision scan (production-file grep for every surface name; helper-name scan across the package's existing 30 test files) found none, but that scan is a read, not a compile. The RED proves absence of the M1 surface, not that the tests themselves compile once the surface exists — first proof of test correctness is the M1 GREEN run, and any latent test defect surfaces there, not here.

### M1 — GREEN (exit evidence)

Measured by the run-phase specialist (manager-develop, cycle_type tdd), milestone M1, card t1640. The M1 exit evidence of plan.md §C (the test-compilation-order rule): `go build` of every package the milestone touched, tests excluded, green; then `go test` of the touched packages compiling and passing. Every call below ran as its own single invocation from the worktree root, each started only after both resource gates held (1-minute load ≤ 32 on `uptime`, memory free ≥ 50% on `memory_pressure`); one gate failure (22:35:20 KST, load 37.14) stopped the sequence under the lane's bounded-patience authorization and passed on re-gate 1 at 22:35:42 KST (load 28.26).

Tree SHA for every cell of this subsection: `9663d75c98` (the M1 production commit; the measurements ran on the working tree carrying exactly that content immediately before the commit landed — no file changed between the last call and the staging).

- command: `go build ./internal/worktreeroot/ ./internal/cli/worktree/`
- gates: 22:25:27 KST, load 15.88, memory free 86%
- exit: 0 (observed, second attempt; attempt 1 failed on a `const` initialized from `filepath.Join` — the specialist's own defect, fixed, no behavior change)
- command: `go test -p 1 -count=1 -run Key ./internal/worktreeroot/` — exit 0, `ok ... 12.500s` (22:33:31, load 19.28, mem 86%)
- command: `go test -p 1 -count=1 -run PathFor ./internal/worktreeroot/` — exit 0, `ok ... 0.450s` (22:33:59, load 29.79, mem 86%)
- command: `go test -p 1 -count=1 -run Candidates ./internal/worktreeroot/` — exit 0, `ok ... 0.497s` (22:34:20, load 27.25, mem 86%)
- command: `go test -p 1 -count=1 -run Same ./internal/worktreeroot/` — exit 0, `ok ... 0.083s` (22:34:33, load 23.04, mem 86%)
- command: `go test -p 1 -count=1 -run TestClassify ./internal/cli/worktree/` — exit 0, `ok ... 0.592s` (22:34:46, load 28.19, mem 86%)
- command: `go test -p 1 -count=1 -run ParseWorktreeList ./internal/cli/worktree/` — exit 0, `ok ... 0.461s` (22:35:42, load 28.26, mem 85%)
- command: `go test -p 1 -count=1 -run TestTreeIdent ./internal/cli/worktree/` — exit 0, `ok ... 0.766s` (22:35:55, load 25.64, mem 85%)
- command: `go test -p 1 -count=1 -run TestEnsureInstanceToken ./internal/cli/worktree/` — exit 0, `ok ... 0.965s` (22:36:10, load 21.30, mem 86%)
- command: `go test -p 1 -count=1 -run TestMovedTo ./internal/cli/worktree/` — exit 0, `ok ... 4.810s` (22:36:23, load 19.79, mem 86%)
- command: `go test -p 1 -count=1 -run TestMovedBy ./internal/cli/worktree/` — exit 0, `ok ... 4.275s` (22:36:47, load 24.78, mem 86%)
- command: `go test -p 1 -count=1 ./internal/config/` — exit 1 first run: the REQ-CKH-008 shipped-key anti-rot guard failed on the new `git_strategy.worktree_project_key` missing from the triage inventory; the inventory entry was added (841 -> 842, class W, evidence the two readers) and the re-run passed — exit 0, `ok ... 7.679s` (22:38:44, load 14.58, mem 86%)

Notes. `-run TestClassify` necessarily subsumes the Listing, SpellingSafe and CaseDiffering families (Go's prefix matching; the no-pipe constraint forbade alternation), so the ten calls cover all 38 RED test functions plus TestMain. The M1 tests passed with ZERO corrections: the RED files were satisfied exactly as committed. One non-test fix rode the same commit: the production loader file is named `loader_worktree_project.go` because the credentials gitignore pattern `*_key.*` (`.gitignore:199`) would have silently excluded the literal name `loader_worktree_project_key.go` from the commit.

#### Five-section summary (verification-claim-integrity §3)

**Claim.** (1) M1 exits GREEN: both packages build (tests excluded) and all 38 M1 tests pass, with `internal/config` (field, loader, inventory) green on its own suite. (2) The spec.md `draft -> in-progress` flip landed in this first production commit of M1. (3) No test was weakened, skipped or corrected; no production behavior was relaxed to pass a test.

**Evidence.** The twelve command invocations above with their verbatim one-line `ok` results and gate readings, this run, each gated before start; the build's failed first attempt is recorded beside its fix.

**Baseline-attribution.** Measured against tree `9663d75c98` (content-identical working tree immediately pre-commit), in worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640`, 2026-10-19 22:25-22:39 KST. The `go` toolchain is external tooling (verification-claim-integrity §2.2 does not bind it).

**Gaps.** (a) The package-wide `go test` of the two packages did NOT run: the `go-test-heavy` slot was held (lane-55, card t1892, bound to 2026-10-19T13:58:30Z, no `--force` taken), so the verification ran as the ten narrowed family calls above; the package-wide run is deferred to the 35e9 landing re-measure per the lane's instruction. (b) `golangci-lint` and coverage were not run at this milestone (not part of the M1 exit contract of plan.md §C; CI owns the repository-wide verdict, PENDING at report time). (c) `internal/template`'s suite was not run; the template edit's safety rests on two reads — the embed manifest is a file-set allowlist (same file, new content) and `catalog.yaml` carries zero git-strategy entries — CI re-verifies. (d) The `-run` families overlap as noted (TestClassify covers five), which the record states rather than hides.

**Residual-risk.** The narrowed families exercised the tests by name; a test accidentally excluded from every family would be unexercised here — the family set was derived from the test-function inventory (38 names mapped to 10 patterns) and cross-checked against the counts, but the derivation is a read, not a mechanical proof; the deferred package-wide run closes it. The RED-to-GREEN path was single-pass, so no mutation testing of the new production logic occurred; the sync audit's independent verification is the second pair of eyes.

### M2 — RED (AC-WHR-006 cell and supporting behavioral reds)

Measured by the run-phase specialist (manager-develop, cycle_type tdd), milestone M2, card t1640. **Branch Y of D-5 is the selected arm**: measurement PM-3 recorded FAIL (operator ruling bd72 authorized lane execution; the one-line record with verbatim stderr is at `.moai/reports/t1640/pre-kickoff-measurements.md` — claude 2.1.303 rejects the absolute home-root path at argument validation, `Invalid worktree name`, exit 1 — and the tracked selector copy the tests parse is `internal/cli/testdata/pm3-record-t1640.md`). Per REQ-WHR-006 branch Y, a launcher short name with no existing tree passes through unchanged and creates nothing; an existing tree is entered.

The red here is BEHAVIORAL, not a compile failure: M2's tests name only M1 and existing symbols (research §R13, M2 row — "tests name: resolver (M1); existing materializer and RegisterLauncherWorktree"), so the package compiles and the tests fail on their assertions against the pre-change tree — the same red the plan §C rule exists to guarantee, one step later in the toolchain. Five tests fail; two pass by design and pin branch Y's pass-through against a branch-X drift. Tree SHA for every cell: `5d4817f2eda7e1a1b6b4c712fd87c28afe5eba69` (the parent of the M2 RED commit). Every call gated (1-minute load ≤ 32, memory free ≥ 50%); exit codes observed (`Exit code 1`).

#### AC-WHR-006 — launcher short name, branch-Y arm and both-branches clause (RB*)

- command: `go test -p 1 -count=1 -run 'TestLauncherShortName' ./internal/cli/`
- gates at first run: 23:06:58 KST, load 15.08, memory free 82%; re-run after two fixture defects in the new tests were fixed (a legacy-tree fixture colliding with its own created-nothing assertion; the glm path's missing API-key guard — `config.DefaultGLMEnvVar` set): 23:08:20 KST, load 12.71, memory free 82%
- exit: 1 (observed)
- stdout (verbatim, decisive failures of the re-run):
```
--- FAIL: TestLauncherShortName_HomeRootTree_IsEntered (0.11s)
    m2_launcher_short_name_test.go:138: backend argv "--worktree t1" does not name the existing home-root tree /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestLauncherShortName_HomeRootTree_IsEntered2841815811/002/worktrees/projects/repo/t1; the launcher must enter it
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.694s
FAIL
```
- red because: a short name that names an existing home-root tree is passed through as the bare short name, so the backend resolves it against `.claude/worktrees/<name>` and would create a duplicate there (launcher.go:913's comment is the current contract) — that fails AC-WHR-006's both-branches clause ("with a tree already present at the home root ... the launcher enters that tree"). The branch-Y no-tree arm passes and is kept as the pin: the backend receives the short name unchanged and the launcher creates nothing, which branch Y preserves (green path M2 GREEN).
- vs acceptance.md: the RB* row defers to the RED commit of M2 with a stub backend — this is that record; the test selects its branch from the recorded PM-3 line (`FAIL`, cited to bd72 above) and fails on an absent or GAP line.

Supporting behavioral reds observed in the same run set (each `Exit code 1`, gates recorded, one call per family):

- command: `go test -p 1 -count=1 -run 'TestMaterializeSessionWorktree_CreatesAtHomeRoot' ./internal/cli/` — 23:05:28 KST, load 23.60, mem 81% — `destination = ".../repo/.moai/worktrees/WT-m2a", want the home root ".../worktrees/projects/repo/WT-m2a"` plus the in-repo-tree, registration and safe.directory clauses (AC-WHR-005a's Go clause; its M0 grep cell is the recorded RED-now).
- command: `go test -p 1 -count=1 -run 'TestWorktreeNew_PathReachesTheSharedMaterializer' ./internal/cli/` — 23:05:58 KST, load 20.21, mem 82% — `moai worktree new created ".../repo/.moai/worktrees/WT-m2new", want the home root` (AC-WHR-005a's second entry path through the shared seam).
- command: `go test -p 1 -count=1 -run 'TestLauncherBranchMaterialize_HomeRootBranchAndHead' ./internal/cli/` — 23:06:22 KST, load 15.92, mem 82% — `tree missing at the home root .../worktrees/projects/repo/m2b` (AC-WHR-005b's Go clause; its M0 grep cell is the recorded RED-now).
- command: `go test -p 1 -count=1 -run 'TestM2_NoLegacyDestinationConstructions' ./internal/cli/` — 23:06:43 KST, load 15.11, mem 82% — the count-first scan printed the expected-site list, then: `worktree_branch_flag.go constructs ".claude", "worktrees" 1 time(s)`, and both destination functions read `sessionWorktreeSubdir` / the literal pair instead of `worktreeroot.PathFor`.

#### Five-section summary (verification-claim-integrity §3)

**Claim.** (1) The AC-WHR-006 RB* cell is red for the stated reason: the existing-home-root-tree entry clause fails against the pre-change tree, while the branch-Y no-tree pass-through holds and is pinned. (2) The AC-WHR-005a/005b Go clauses are red (their M0 grep cells stay the recorded RED-now observations; these are the behavioral companions whose green the M2 GREEN commit delivers). (3) The tests carry no M2-built symbol references, so the red is behavioral per research §R13's M2 row; spec.md needed no frontmatter change (already in-progress).

**Evidence.** The five narrowed invocations above with verbatim decisive output and per-call gate readings; the selector parse fixed to the tracked PM-3 copy citing bd72; the glm key and fixture-collision corrections are recorded as the specialist's own attempt-1 defects (attempt 2 of the 3-attempt bound).

**Baseline-attribution.** Measured in this run, this worktree, against HEAD `5d4817f2eda7e1a1b6b4c712fd87c28afe5eba69` — the parent of the M2 RED commit. The `go` toolchain is external tooling (§2.2 does not bind it).

**Gaps.** (a) `TestLauncherShortName_NoExistingTree_PassesThrough`, `..._LegacyTree_PassesThrough` and `..._Glm_NoExistingTree_PassesThrough` PASSED in the RED run — by design: they pin branch Y's preserved pass-through and would fail against a branch-X implementation; they are regression pins, not red cells. (b) The stub-backend assertion is on argv and created-nothing, not on the real backend's acceptance of absolute paths — PM-3 measured that acceptance as FAIL, and entry strategy for the real backend is REQ-WHR-025/t1879's domain. (c) No package-wide run (go-test-heavy discipline continues).

**Residual-risk.** The scan-based 005a clause reads function bodies by name; a rename would silently empty it (the plan's own grep cells at M0/M3 carry the count-first form). The AC-006 selector depends on the tracked testdata copy staying in sync with the gitignored source record — the copy cites bd72 and its source, and the §E.2 citation here ties them; a divergence would surface as the tests' own selector failing on a malformed line.

## §G Override and Refusal Record

- 2026-10-19T03:29:15Z SPEC-WORKTREE-HOME-ROOT-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter8.md
- 2026-10-19T03:33:35Z SPEC-WORKTREE-HOME-ROOT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter8.md
- 2026-10-19T03:34:01Z SPEC-WORKTREE-HOME-ROOT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640/.moai/reports/t1640/plan-audit-iter8.md
