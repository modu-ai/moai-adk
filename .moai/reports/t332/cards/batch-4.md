# t332 card sweep — batch 4

Cards: t286 t287 t288 t295 t296 t297 t300 t302 t304 t305  (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t286

**Premise (one sentence).** A "risky-command guard" regex has a bidirectional defect — a flag-order bypass (evasion) and a quoted-data false positive (over-match) coexist, per issue #1658.

**Premise verdict.** `unverified` — I located a strong candidate regex guard (`internal/hook/branch_guard.go`, `branchStatePatterns` + `quotedArgumentPattern`) that already handles the quoted-data false-positive class explicitly (a dedicated `quotedArgumentPattern`/`substituteQuotedArguments` pass collapses quoted spans before matching, with a comment citing the exact failure mode: `moai todo add "… git switch …"` being wrongly denied). Its own comments also document a flag-order-shaped gap (`git branch -vD foo` combined short flags "do not match") but explicitly label that as an **accepted, documented direction for a fail-open guard**, not a live defect. I could not confirm within budget whether issue #1658 names this guard specifically or a different "risky command" guard elsewhere in the tree (e.g. a Bash-tool-wide risk-amplifier check referenced only in doctrine, `coding-standards.md` §Bash Risk-Amplifier Doctrine, with no matching Go implementation found). The card may be describing a guard I did not find, or may be re-describing an already-accepted tradeoff as a bug.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree named for this card in `00-worktree-list.txt`)

**Claim.** No commit delivering a fix for a "risky-command guard flag-order bypass / quoted-data false positive" defect exists on either pinned ref; the card is unresolved. Whether the described defect is real, and against which specific guard, is unresolved by this sweep.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt286\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt286\b' --oneline
(no output)
$ grep -rl "risky\|dangerous" internal/hook/*.go | grep -v _test
internal/hook/branch_guard.go
internal/hook/pre_tool.go
internal/hook/types.go
```
`internal/hook/branch_guard.go:156` — `var quotedArgumentPattern = regexp.MustCompile(`'[^']*'|"[^"]*"`)` (already-shipped false-positive fix).
`internal/hook/branch_guard.go:96-102` — comment: "exotic combined short flags (`git branch -vD foo`) do not match — under-matching an obfuscated form is the documented correct direction for a fail-open guard" (labeled accepted, not a defect).

**Baseline-attribution.** File contents read at worktree HEAD `6165f9f5e`. Commit-grep run against the two pinned SHAs listed above (fetched 2026-08-30T11:16:22Z per WORKER-INSTRUCTIONS.md).

**Gaps.** Did not read GitHub issue #1658 itself (no network access in this sweep) to confirm which guard it names. Did not exhaustively search every regex-based command guard in the tree — restraint budget bounds this to `internal/hook/*.go`. Did not attempt to reproduce either claimed failure mode (bypass input, false-positive input) against `branch_guard.go` or any other candidate.

**Residual-risk.** If issue #1658 names a different guard than `branch_guard.go`, this entire premise assessment is about the wrong artifact. If it does name `branch_guard.go`, the "flag-order bypass" may be the already-accepted residual the comments describe, in which case the card may be asking to un-accept a deliberate tradeoff rather than fix a bug.

**Proposed disposition.** `needs-operator-decision` — rests on: the found candidate guard already documents the false-positive fix and explicitly accepts the flag-order gap as correct-by-design; the operator should confirm whether issue #1658 targets this guard before dispatching repair work.

**Overlap candidates.** t287 (same "guard defect via GitHub issue" pattern, adjacent issue #1659, same session/date). No other in-scope id names a guard-regex artifact.

---

### t287

**Premise (one sentence).** A "worktree guard" blocks heredoc-brace-folded dangerous commands but not an equivalent command-substitution (`$(...)`) at the same position, per issue #1659 and the lead's own-session observation.

**Premise verdict.** `unverified` — I have first-hand, directly-observed evidence bearing on this premise from *this same sweep session*: a compound Bash command (`for id in ...; do ...; done`) issued in this session was refused with the message "This session is isolated in the worktree ... Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Split it into plain, separate commands." A second, non-git compound command (multi-file `ls -d` inside a shell loop) was refused with the *identical* message even though it touched no git state at all. This confirms the card's own observation ("이 세션도 until 루프·... 포함 명령 거부됐으나"). Critically, `internal/hook/branch_guard.go:182-183` documents, in its own comments, that this behavior belongs to a SEPARATE mechanism: *"The Claude Code worktree isolation guard refuses that same shape independently"* — i.e. this repo's own `branch_guard.go` explicitly disclaims ownership of the compound-command-refusal behavior I personally observed. I found no Go source in `internal/` implementing a heredoc-vs-command-substitution guard matching the card's description. This raises a real possibility that the artifact the card wants patched is Claude Code's own harness sandbox (external to this repository, not fixable by a moai-adk-go code change), not something in this codebase — but I could not fully rule out that moai-adk-go ships its own complementary hook that ALSO does heredoc-folding (e.g. inside `internal/permission/stack.go`, which does implement an AST-based, non-regex command-substitution and IFS-word-split detector — a much stronger mechanism than "heredoc brace-folding regex" and does not obviously have the described gap).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** No delivering commit exists on either pinned ref. The premise cannot be confirmed as an in-repo defect within this sweep's budget; there is concrete evidence the described refusal behavior is (at least partly) owned by Claude Code's own harness, external to this repository.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt287\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt287\b' --oneline
(no output)
```
Directly observed in this session (tool-call error, not a file):
> "This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t332, but this command is too complex to verify that it stays inside the worktree. Refusing to run it..."
(triggered twice: once by a `for`-loop with git subcommands, once by a `for`-loop with plain `ls -d`, i.e. not git-specific.)

`internal/hook/branch_guard.go:181-183`:
```
// fails open on every uncertainty, so under-matching a deliberately obfuscated
// form is the correct direction to err. The Claude Code worktree isolation
// guard refuses that same shape independently.
```

**Baseline-attribution.** Session-observed tool refusal text from this sweep run (2026-08-30). File comment read at worktree HEAD `6165f9f5e`. Commit-grep against the two pinned SHAs.

**Gaps.** Did not read issue #1659 (no network). Did not test `internal/permission/stack.go`'s AST-based command-substitution detector against the card's specific heredoc-vs-`$(...)` bypass scenario. Did not determine whether `internal/permission/stack.go`'s guard and the sandbox-level "worktree isolation guard" I personally hit are the same mechanism or two independent layers.

**Residual-risk.** If the card in fact targets `internal/permission/stack.go` (an AST parser, not a "heredoc brace-folding regex" as described), the premise's characterization of the mechanism is wrong even if a real gap exists elsewhere in that file. If it targets the external Claude Code sandbox, this card cannot be resolved by a moai-adk-go code change at all.

**Proposed disposition.** `needs-operator-decision` — rests on: this session's own two directly-observed refusals of non-git compound commands, matching the "Claude Code worktree isolation guard" that `branch_guard.go` itself says is external to this repo's mechanism.

**Overlap candidates.** t286 (same issue-driven guard-defect pattern, adjacent issue #1658). No other in-scope id touches command-guard regexes.

---

### t288

**Premise (one sentence).** The `goal_arm` MCP wrapper misclassifies a prose (model) condition as mechanical whenever the prose text does not contain the literal word "transcript," causing the shell-command execution path to run the prose as a command and the stop-hook to exit 2 every turn-end.

**Premise verdict.** `holds` — confirmed by reading the actual classifier. `internal/cli/goal.go:37` (`parseCondition`) implements the classification rule EXPLICITLY as: `if strings.Contains(strings.ToLower(s), "transcript") { return Condition{Type: ConditionModel, ...} }` — else the ENTIRE string is treated as `ConditionMechanical` and stored as a shell `Cmd`. `internal/cli/mcp_server.go:641` confirms the MCP tool `goal_arm` calls `cond := parseCondition(conditionText) // same classifier the CLI uses` — i.e. the MCP wrapper inherits the identical single-keyword discriminator, with no additional heuristic. Any prose claim that omits the literal substring "transcript" (plausible in Korean-language or differently-worded English claims) is stored as a mechanical condition and will be run as a shell command by the stop-goal evaluator — a non-existent/invalid "command" would exit non-zero every turn, matching the card's claimed "매 턴엣드 exit 2" symptom.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** The classifier's single-keyword ("transcript") discriminator is unchanged on both pinned refs; no commit adds a broader/more robust model-vs-mechanical classification path.

**Evidence.**
```
internal/cli/goal.go:29-49 (parseCondition):
func parseCondition(s string) goal.Condition {
	s = strings.TrimSpace(s)
	if strings.Contains(strings.ToLower(s), "transcript") {
		return goal.Condition{Type: goal.ConditionModel, Claim: s}
	}
	cmd := s
	...
	return goal.Condition{Type: goal.ConditionMechanical, Cmd: cmd, ExpectExit: expect}
}

internal/cli/mcp_server.go:641:
	cond := parseCondition(conditionText) // same classifier the CLI uses

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt288\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt288\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/cli/goal.go` and `internal/cli/mcp_server.go` read at worktree HEAD `6165f9f5e`. Commit-grep against the two pinned SHAs.

**Gaps.** Did not actually arm a goal with a non-"transcript" prose condition and observe a live turn-end exit 2 (that would require a live stop-hook cycle, outside this read-only sweep's scope). Did not check `internal/goal` package's `stop-goal` evaluator to confirm it truly shells out `Cmd` unconditionally for `ConditionMechanical` (inferred from the type name and comment, not directly read).

**Residual-risk.** If the stop-goal evaluator has an additional guard that detects "this mechanical command looks like prose and refuses to run it," the exit-2 loop symptom described might not actually occur even though the classification itself is as described. I did not verify the evaluator side.

**Proposed disposition.** `keep` — rests on: `parseCondition`'s single-keyword classifier is unchanged and is shared verbatim between the CLI and the MCP wrapper, so the misclassification path the card describes is real and reachable via the MCP entry point specifically.

**Overlap candidates.** None observed among in-scope ids (memory notes list related-but-not-in-scope cards `feedback_goal_arm_mcp_worktree_split.md` and `feedback_goal_keying_worktree_unreliable.md`, but their card ids are not in `inscope-all.txt`).

---

### t295

**Premise (one sentence).** No launcher-exposed path exists to create a worktree that checks out an EXISTING branch (as opposed to creating a new one), forcing the lead to bypass `moai worktree`'s launcher discipline with a raw `git worktree add` for the gitflow `develop` integration tree.

**Premise verdict.** `holds` — confirmed at two levels. (1) The underlying git-level primitive DOES support existing branches: `internal/core/git/worktree.go`'s `Add(path, branch)` calls `branchExists()` and dispatches to `buildWorktreeAddArgs(exists, ...)`, which for `exists==true` emits `worktree add -- <path> <branch>` (checkout existing) vs. `-b <branch>` for new. (2) But the actual `moai cc -w` launch path does NOT use this existing-branch-capable function: `internal/cli/session_worktree.go:52` documents its own helper as "`sessionWorktreeGitWorktreeAdd` runs `git worktree add -b <branch> <dest>`" — hardcoded to the `-b` (always-new-branch) form, with no code path found that ever calls it with an existing branch name. This matches the card's claim exactly: the git-level capability exists somewhere in the codebase, but no launcher (`moai cc -w`, and `EnterWorktree` is a native Claude Code tool outside this repo entirely) exposes a way to invoke it for an existing branch.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Two commits matched the grep on develop but are MENTIONS, not deliveries — recorded per the false-positive warning in the instructions:
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt295\b' --oneline
daf206903 Merge WT-clocal-audit into develop — t308 CLAUDE.local.md audit (SPEC-CLOCAL-AUDIT-001)
281fde607 docs(t308): audit CLAUDE.local.md against measured reality — 20 defects fixed
```
Both commits' bodies read: "§4.1 excluded (t294/t295/t298/t303)" / "Scope: §4.1 (git-flow lane section) excluded — owned by t294/t295/t298/t303." — t295 is explicitly named as OUT OF SCOPE for that audit, i.e. this is a citation, not a landing. Confirmed by reading `git show --no-patch` on both SHAs in full.

**Claim.** t295 is unresolved; the underlying `git.WorktreeManager.Add` already supports an existing-branch code path but the `moai cc -w` launcher does not use it — the card's "no sanctioned path" claim is accurate for the actual launcher, even though the capability exists one layer down.

**Evidence.**
```
internal/core/git/worktree.go:58-73 (buildWorktreeAddArgs):
func buildWorktreeAddArgs(branchExists bool, path, branch string) []string {
	if branchExists {
		return []string{"worktree", "add", "--", path, branch}
	}
	return []string{"worktree", "add", "-b", branch, "--", path}
}

internal/cli/session_worktree.go:52-57:
	// sessionWorktreeGitWorktreeAdd runs `git worktree add -b <branch> <dest>
	...
	sessionWorktreeGitWorktreeAdd = gitWorktreeAddReal

$ grep -rln "GitWorktree\.Add\|GitWorktree\b" internal/cli/*.go | grep -v _test | grep -v deps.go
internal/cli/inventory.go
internal/cli/root.go
internal/cli/session_worktree.go
```
`session_worktree.go` — the file backing `moai cc -w` / `moai worktree` — never calls the existing-branch-capable `git.WorktreeManager.Add`; it has its own separate, hardcoded-`-b` helper.

**Baseline-attribution.** All three files read at worktree HEAD `6165f9f5e`. Commit-grep + `git show --no-patch` against the two pinned SHAs.

**Gaps.** Did not trace every call site of `sessionWorktreeGitWorktreeAdd` to rule out a hidden existing-branch code path elsewhere in `session_worktree.go` (only the doc comment and the variable wiring were read). Did not check whether `internal/cli/inventory.go` or `root.go` (which also reference `GitWorktree`) expose an existing-branch path through a different subcommand (e.g. `moai worktree recover`).

**Residual-risk.** If `inventory.go` or `root.go` DO wire `GitWorktree.Add` to an existing-branch flag somewhere, the premise would be partially falsified (a path exists, just undocumented/undiscovered). This sweep did not read those two files.

**Proposed disposition.** `keep` — rests on: `session_worktree.go`'s own doc comment hardcodes `-b <branch>` (always new branch) and no call site was found passing an existing branch, while the lower-level git manager already supports it — a real gap between capability and exposed surface.

**Overlap candidates.** t231 (in-scope — "worktree clean 앵커 소스 판독", cited directly by this card as related). t297 (same batch, also cites t231 in its own "연관" list, and both surfaced from the same 2026-08-27 gitflow-transition session). The card also names t209/t298/t303/t294 as related, none of which are in `inscope-all.txt`.

---

### t296

**Premise (one sentence).** `coding-standards.md`'s "Language Policy" section carries no "16-language programming-neutrality" content (it is about writing instruction docs in English), yet ~10 other template files cite it as the canonical location for that contract, so readers following the citation cannot reach the promised body.

**Premise verdict.** `holds` — confirmed directly. `internal/template/templates/.claude/rules/moai/development/coding-standards.md`'s `## Language Policy` section (lines 9-22) reads: "All instruction documents must be in English: CLAUDE.md, Agent definitions, ... User-facing documentation may use multiple languages: README.md, CHANGELOG.md, ..." — no mention of "16," no programming-language list, no neutrality contract. Meanwhile 7 distinct template files contain the literal phrase "16-language neutrality contract" (8 occurrences total, vs. the card's claimed "10곳" — close but not exact; the discrepancy does not affect the core claim).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** The cited section still lacks the promised content and the dangling citations still exist, on both pinned refs.

**Evidence.**
```
$ sed -n '9,22p' internal/template/templates/.claude/rules/moai/development/coding-standards.md
## Language Policy

All instruction documents must be in English:
- CLAUDE.md
- Agent definitions (.claude/agents/**/*.md)
- Slash commands (.claude/commands/**/*.md)
- Skill definitions (.claude/skills/**/*.md)
- Hook scripts (.claude/hooks/**/*.py, *.sh)
- Configuration files (.moai/config/**/*.yaml)

User-facing documentation may use multiple languages:
- README.md, CHANGELOG.md
- User guides, API documentation

$ grep -rln "16-language neutrality contract" internal/template/templates/
internal/template/templates/.moai/docs/generic-patterns-guide.md
internal/template/templates/.claude/rules/moai/development/skill-authoring.md
internal/template/templates/.claude/skills/moai/workflows/loop.md
internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md
internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md
internal/template/templates/.claude/skills/moai-workflow-loop/references/examples.md
internal/template/templates/.claude/skills/moai-workflow-loop/references/reference.md

$ grep -rn "16-language neutrality contract" internal/template/templates/ | wc -l
8

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt296\b' --oneline
(no output)
```

**Baseline-attribution.** `coding-standards.md` and all 7 citing files read at worktree HEAD `6165f9f5e`. Grep scope: `internal/template/templates/` (the distributed template root; did not scan the local `.claude/` mirror separately). Commit-grep against both pinned SHAs.

**Gaps.** Did not check whether the local (non-template) mirror `.claude/rules/...` and `.claude/skills/...` carry the same 7 dangling citations (only the template source was scanned, per the card's own framing that the template is what ships to users). Did not verify the "8 occurrences, 7 files" count against the card's claimed "10곳" beyond noting the discrepancy — could be an outdated count in the card, or a scope difference (e.g. the card may also be counting occurrences in `.claude/` local mirror or in generated/rendered files).

**Residual-risk.** The card's count (10) vs. measured count (7 files / 8 occurrences) is close but not exact — if the operator treats the exact count as load-bearing, this is a minor discrepancy to flag, not a premise failure.

**Proposed disposition.** `keep` — rests on: direct read of the cited section (no "16" or language list) plus a direct grep confirming multiple dangling citations still pointing at it.

**Overlap candidates.** None observed among in-scope ids.

---

### t297

**Premise (one sentence).** The launch-ledger (`launch.yaml` `projects[]`) grows unboundedly because every worktree that ever records a launch profile gets its own permanent entry with no reaping mechanism when the worktree is later removed, and this was requested explicitly as REQ-009 follow-up from t293's plan-auditor.

**Premise verdict.** `holds`, with a nuance the card's scope item (1) does not need — confirmed by reading the write path. `internal/profile/profile.go`'s ledger-write function (~line 555-583) already performs **duplicate-spelling dedup**: `if stored, found := lookupProjectKey(projects, key); found { key = stored }` before writing, specifically to avoid two entries for the same directory under different path spellings (own comment: "Writing the caller's spelling instead would leave two entries naming one project"). This means scope item (1) as literally stated ("중복 행을 만들지 않고 갱신") is PARTIALLY already done for same-directory-different-spelling. However, scope item (2) — reaping dead worktree entries — is confirmed still absent: no `Prune`/`Reap`/`RemoveProject`/`delete(projects...)` function was found anywhere in `internal/profile/` or `internal/cli/`. Since each worktree is a genuinely distinct directory (not a spelling variant), every worktree that ever records a profile gets a permanent, never-reaped entry — this is the growth mechanism the card actually describes, and it is real.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

The originating commit (a MENTION, not a delivery — this is the commit that itself proposed t297 as a follow-up card):
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt297\b' --oneline
2114ed981 docs(SPEC-STATUSLINE-PROFILE-RESPECT-001): sync-phase artifacts — in-progress -> implemented (t293)
```
Body: "AC-009 / M5 deferred by kickoff decision D1; follow-up card t297 queued." — confirms t297 is a request, not a resolution.

**Claim.** No pruning/reaping mechanism for stale `projects[]` ledger entries exists on either pinned ref; ledger growth is unbounded as worktrees accumulate and are removed without a corresponding ledger cleanup.

**Evidence.**
```
internal/profile/profile.go:561-579 (write path, dedup already present):
	if key := normalizeProjectKey(projectRoot); key != "" {
		projects, ok := existing[projectsKey].(map[string]any)
		if !ok {
			projects = make(map[string]any)
		}
		if stored, found := lookupProjectKey(projects, key); found {
			key = stored
		}
		projects[key] = name
		existing[projectsKey] = projects
	}

$ grep -rn "func.*[Pp]rune\|func.*[Rr]eap\|func.*[Rr]emoveProject\|delete(projects" internal/profile/*.go internal/cli/*.go | grep -v _test
internal/cli/chain.go:474:func newChainPruneCmd() *cobra.Command
internal/cli/chain.go:487:func runChainPrune(...)
```
(The only "prune" hits are `chain.go`'s unrelated chain-state prune command, not the launch ledger.)

**Baseline-attribution.** `internal/profile/profile.go` (full write function, lines ~549-614) and a repo-wide grep of `internal/profile/*.go` + `internal/cli/*.go` for prune/reap functions, read/run at worktree HEAD `6165f9f5e`.

**Gaps.** Did not measure an ACTUAL ledger's current row count / growth rate (would require access to a real `~/.moai/launch.yaml` across many worktree lifecycles, out of this sweep's read-only repo-tree scope). Did not check whether `moai worktree done` or `moai worktree clean` (the disposal commands) happen to call any ledger-cleanup code not matched by my prune/reap keyword grep.

**Residual-risk.** If a differently-named function (not matching "prune/reap/remove/delete") performs cleanup on worktree disposal, my keyword grep would miss it — the absence claim rests on the specific keyword set searched, not an exhaustive read of every disposal code path.

**Proposed disposition.** `keep` — rests on: no prune/reap function found anywhere in the two most relevant packages, confirming scope item (2) — the reaping half of the card — is genuinely unaddressed, even though the dedup half (scope item (1)) is partially already shipped.

**Overlap candidates.** t231 (in-scope — worktree-clean anchor-source reading, cited directly by this card). t295 (same batch, both surfaced 2026-08-27, both cite t231). The card also names t293/t209 as related; neither is in `inscope-all.txt`.

---

### t300

**Premise (one sentence).** A "baseline-first" acceptance criterion (AC-GF-022, requiring the baseline measurement to precede the first implementation commit) was violated when both landed in the same commit and then became permanently unverifiable after a squash merge, and this card exists to prevent recurrence of that pattern in future SPECs.

**Premise verdict.** `holds` — this card's own premise is directly corroborated by the commit that spawned it, read in full. `git show --no-patch` on the matched commit shows the operator-decision record stating the exact facts the card restates: "7f2e9e77d carries m5-baseline.md (+67) together with the M5 implementation files in one commit; that commit is unreachable from origin/main and origin/develop (both ancestor checks exit 1; PR #1648 squash 6786c3fa4 carries the artifacts into history)... Ordering is permanently unverifiable, not deferred... Recurrence prevention: card t300." This is a first-party, already-landed acknowledgment of the exact defect t300 describes and an explicit statement that t300 was created to prevent recurrence — the premise needs no further independent verification beyond this citation, since the citation IS the origin of the claim.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

The originating (mention) commit:
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt300\b' --oneline
69891ce99 docs(t279): record AC-GF-022 ordering deviation at source (operator decision A, card t279)
```
This commit records the deviation and creates the follow-up card; it does NOT implement the recurrence-prevention procedure t300 asks for (scope items 1-3: procedural guard for baseline-first ACs, squash-merge interaction, sweep of other SPECs for the same pattern).

**Claim.** No commit implements a procedural or mechanical guard against future baseline-in-same-commit ACs; t300's recurrence-prevention work remains undone.

**Evidence.**
```
$ git show --no-patch --format='%H%n%s%n%b' 69891ce99
69891ce99773563ec17e10961737b5284582fe18
docs(t279): record AC-GF-022 ordering deviation at source (operator decision A, card t279)
Operator decision A (2026-08-27): keep status: completed, record the AC-GF-022
ordering deviation in place at the AC definition in acceptance.md. Facts
re-measured in this tree before writing: both m5 artifacts exist; 7f2e9e77d
carries m5-baseline.md (+67) together with the M5 implementation files in one
commit; that commit is unreachable from origin/main and origin/develop (both
ancestor checks exit 1; PR #1648 squash 6786c3fa4 carries the artifacts into
history). Ordering is permanently unverifiable, not deferred; post-artifact
existence/content remains verifiable. Recurrence prevention: card t300.
Cross-references the progress.md §E.4 open_followups row.
```

**Baseline-attribution.** Full commit message of `69891ce99` read directly via `git show`, against worktree HEAD `6165f9f5e`. Commit-grep against both pinned SHAs.

**Gaps.** Did not read `SPEC-V3R6-GRAPH-FRESHNESS-001`'s `acceptance.md` directly to see the current state of the AC-GF-022 deviation note. Did not sweep other SPECs' acceptance.md files for the same baseline-first-AC pattern (that IS scope item (3) of the card itself — explicitly out of this sweep's bounded-depth budget, consistent with the Restraint clause).

**Residual-risk.** None specific beyond the general risk that a procedural fix (documentation-only, per the card's own "주의 - 대표 mutant" warning about doc-only non-fixes) might land without an actual verification mechanism, which is exactly the failure mode the card itself warns against.

**Proposed disposition.** `keep` — rests on: the originating commit itself both confirms the defect and explicitly names t300 as the not-yet-done recurrence-prevention follow-up.

**Overlap candidates.** t291 named by the card itself ("F5" squash-provenance orphaning — same root cause) — not in `inscope-all.txt`. t279 (the originating SPEC/card) — also not in `inscope-all.txt`. No in-scope id overlaps.

---

### t302

**Premise (one sentence).** `.claude/workflows/sync-audit-4dim.js` states two contradictory things about its own verdict authority in the same file — the header says its verdict was PROMOTED to binding on the happy path, while the `meta.description` field still says it is "an execution vehicle, NOT the binding sync-auditor verdict owner."

**Premise verdict.** `holds` — confirmed by reading the file directly (both cited passages are present verbatim, in the LOCAL, currently-loaded copy at `.claude/workflows/sync-audit-4dim.js`, i.e. the exact file this session's own skill listing describes with matching text). Header (lines 4-10): "SPEC-AUDIT-SNAPSHOT-001 (A3) PROMOTED its verdict to BINDING on the happy path: where the verdict is PASS with all four dims above their floor, not INCOMPLETE, and no contested finding, the orchestrator treats this workflow's harmonic-mean verdict as the binding sync-phase verdict and does NOT spawn the cold `sync-auditor` subagent." `meta.description` field (line ~50): "Sync-phase 4-dimension quality read (Functionality/Security/Craft/Consistency) — parallel read-only judges + in-script harmonic-mean verdict; **execution vehicle, NOT the binding sync-auditor verdict owner**." These two statements directly conflict on the exact question the card names: whether this script's verdict is binding. The negation in `meta.description` also independently confirmed at session start — this exact string appears in this session's own available-skills listing for `sync-audit-4dim`, meaning the contradiction is externally visible, not just an internal file artifact.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** No commit resolves the header-vs-description contradiction in `sync-audit-4dim.js` on either pinned ref.

**Evidence.**
```
$ sed -n '1,12p' .claude/workflows/sync-audit-4dim.js
// sync-audit-4dim.js — 4-dimension sync-phase quality verdict (Context → Judge → Verdict)
//
// VERDICT SCOPING (what this workflow IS and is NOT):
//   This is an EXECUTION VEHICLE for a skeptical 4-dimension quality read. SPEC-AUDIT-SNAPSHOT-001
//   (A3) PROMOTED its verdict to BINDING on the happy path: where the verdict is PASS with all
//   four dims above their floor, not INCOMPLETE, and no contested finding, the orchestrator treats
//   this workflow's harmonic-mean verdict as the binding sync-phase verdict and does NOT spawn the
//   cold `sync-auditor` subagent. The cold auditor remains the FALLBACK verdict owner for the
//   failure modes (INCOMPLETE / dim-0 / contested finding)...

$ sed -n '47,51p' .claude/workflows/sync-audit-4dim.js
export const meta = {
  name: 'sync-audit-4dim',
  description: 'Sync-phase 4-dimension quality read (Functionality/Security/Craft/Consistency) — parallel read-only judges + in-script harmonic-mean verdict; execution vehicle, NOT the binding sync-auditor verdict owner',

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt302\b' --oneline
(no output)
```

**Baseline-attribution.** File read directly from this worktree's live `.claude/` tree (not the template source) at HEAD `6165f9f5e`, matching the card's own citation of local line numbers (":53", ":4-10"). Commit-grep against both pinned SHAs.

**Gaps.** Did not check `internal/runtime.FourDimVerdict.IsBinding()` (the header's cited "mechanical predicate") to verify the fallback conditions (INCOMPLETE / dim-0 / contested) are actually observable in code, as scope item (2) requests. Did not check `sync.md:81` or `sync-auditor.md` for alignment (scope item (3)). Did not check whether `internal/template/templates/.claude/workflows/sync-audit-4dim.js` (the template source, distinct from this local copy) carries the same contradiction or has already been fixed there — only the local copy was read.

**Residual-risk.** If the template source has already been corrected and only the local copy is stale, `make build` + re-sync would resolve this without new authoring work — a different disposition than a fresh code/doc fix. This was not checked.

**Proposed disposition.** `keep` — rests on: direct verbatim read of both contradicting passages in the same live file, plus independent corroboration that the "NOT binding" phrasing is externally visible (this session's own skill listing).

**Overlap candidates.** None observed among in-scope ids (SPEC-AUDIT-SNAPSHOT-001 is named but no in-scope card id references it directly in the visible batch text).

---

### t304

**Premise (one sentence).** Six of the 55 package paths cited across `.moai/project/codemaps/*.md` (`internal/design`, `internal/evaluator`, `internal/factory`, `internal/migrate`, `internal/research`, `internal/state`) name packages that do not exist in the tree, and this was already true at the codemaps' own stamped-baseline commit (not new drift).

**Premise verdict.** `holds` — confirmed directly. All six named paths were checked individually and none exist in the current worktree tree:
```
$ ls -d internal/design internal/evaluator internal/factory internal/migrate internal/research internal/state
ls: internal/design: No such file or directory
ls: internal/evaluator: No such file or directory
ls: internal/factory: No such file or directory
ls: internal/migrate: No such file or directory
ls: internal/research: No such file or directory
ls: internal/state: No such file or directory
```
And a grep confirms at least one codemaps file (`modules.md`) still cites them:
```
$ grep -rln "internal/design\|internal/evaluator\|internal/factory\|internal/migrate\|internal/research\|internal/state" .moai/project/codemaps/
.moai/project/codemaps/modules.md
```

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** The six nonexistent-package citations are still present in `.moai/project/codemaps/modules.md` on both pinned refs; no correction commit exists.

**Evidence.** (as above — directory-absence check + citation grep, both run against this worktree's HEAD)
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt304\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt304\b' --oneline
(no output)
```

**Baseline-attribution.** Directory-existence checks and codemaps grep run directly against this worktree's live tree at HEAD `6165f9f5e` (a reasonable proxy for the pinned develop tree since the card's own claim is that this defect predates and postdates the recent drift window — I did not separately check out the pinned develop tree to re-verify absence there, relying instead on the card's own git-ls-tree citation at the stamped baseline `a995e58fa`, which I did not independently re-run).

**Gaps.** Did not independently re-run `git ls-tree -d a995e58fa -- internal/design ...` (the card's own cited baseline-verification command) — I verified only against the current worktree tree, not the historical baseline commit. Did not identify which specific line, in which of the (possibly multiple) codemaps files, cites each of the six paths — only confirmed `modules.md` as one hit. Did not check the other 49 (of 55) cited paths for accuracy.

**Residual-risk.** If the current worktree tree differs from the stamped baseline in a way that removed these six packages AFTER the codemaps were generated (rather than them being absent from the start, as the card claims), my current-tree-only check would still show "absent" but for a different reason than the card's history claim — this would not change the practical disposition (codemaps are still wrong) but would change the "not new drift" framing.

**Proposed disposition.** `keep` — rests on: direct confirmation that all six named packages are absent from the tree and at least one codemaps file still cites them.

**Overlap candidates.** t291 (named directly by the card as "직교" — orthogonal but related, SPEC-STAMP-REACHABILITY-001) — not in `inscope-all.txt`. No in-scope id overlap observed.

---

### t305

**Premise (one sentence).** The statusline warm-render path spends ~93% of its ~236ms wall time on 5 serialized `git` subprocess spawns (measured in `.moai/reports/t215/profiling.md`), and two specific, quantified optimizations (deduping a repeated `rev-parse` call, and trimming the status/rev-list spawn set) could recover 40-77% of that time — this card asks to re-measure and apply them in the current tree.

**Premise verdict.** `holds` — the cited profiling report exists and its figures match the card's claims exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** `.moai/reports/t215/profiling.md` substantiates every specific number the card cites; no commit yet applies the two optimization candidates (deduped `rev-parse`, reduced status/rev-list spawn count) on either pinned ref.

**Evidence.**
```
$ grep -n "93%\|236ms\|spawn\|rev-parse" .moai/reports/t215/profiling.md
7:The statusline warm-path wall time is **~236 ms per render on this machine/tree**, and **≈100% of it is mechanically attributed**: ~93% is the five serialized git subprocess spawns on the render path, ~7% is Go process boot...
51:git rev-parse --git-dir:  median=29.5ms  ← NewRepository spawn 1/2 (manager.go:43)
52:git rev-parse --show-toplevel: median=29.1ms  ← NewRepository spawn 2/2 (manager.go:48)
87:| Builder init: 2× `git rev-parse` | 58.7 ms | ~25% |
88:| Git status collection: 3 spawns (`symbolic-ref`, `status --porcelain`, upstream `rev-list`) | 161.8 ms | ~68% |

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt305\b' --oneline
(no output)
```
The card's two quantified follow-up candidates match the report's own numbers: "-58.7ms" (Builder init dedup) = the report's `58.7 ms / ~25%` row; "-36~-123ms" (status/rev-list trim) is within the report's `161.8 ms / ~68%` row's range.

**Baseline-attribution.** `.moai/reports/t215/profiling.md` read at worktree HEAD `6165f9f5e`. Commit-grep against both pinned SHAs.

**Gaps.** Did not re-run the profiling benchmark (`internal/statusline/profile_bench_test.go`) myself to check whether the figures still hold in THIS tree/session's load window — the card itself explicitly flags this as required first-step work ("이 트리에서 먼저 재측정할 것"), which is why I did not attempt it (out of restraint-budget scope for a read-only sweep). Did not verify `internal/statusline/manager.go:43,48` (the cited `NewRepository` call sites) actually still contain the described duplicate `rev-parse` calls — only the profiling-report citation was read, not the current source.

**Residual-risk.** If `manager.go` has already been partially refactored since the t215 report was written, the specific line numbers/duplication the card describes might have shifted or already been partly addressed — this sweep did not check the current state of `manager.go` itself, only the historical report.

**Proposed disposition.** `keep` — rests on: the profiling report's own numbers substantiate every figure the card cites, and no commit matching this card exists on either pinned ref.

**Overlap candidates.** None in-scope (t215 and t211 are named by the card but neither is in `inscope-all.txt`).
