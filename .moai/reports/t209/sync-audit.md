# Sync-phase audit — SPEC-WORKTREE-REAPER-001 (card t209)

- Position: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`, branch `WT-worktree-reaper`, HEAD `301841e0f` (matches the dispatched HEAD).
- Scope audited: `git diff origin/main...HEAD` (37 files), acceptance.md (28 criteria), spec.md (25 requirements), progress.md §E.1–§E.4.

## Verdict: **FAIL** — harmonic mean **0.778** against threshold 0.85

Must-pass firewall also trips independently: **Functionality 0.72 < 0.85**.

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 0.72 | FAIL | `go build ./...` exit 0; `go test ./internal/session/...` → `ok … 17.792s`; `go test ./internal/cli/worktree/...` → `ok … 13.306s`. But the M1 git fallback is unreachable in practice: replaying `gitBranchMergedReal`'s parser against real `git branch --merged origin/main` yields `parsed entries = 149 / still '+' prefixed = 119 / matchable = 30`, sample unmatchable `"+ WT-agent-toml-dual"`. See F1. |
| Security (25%) | 0.82 | PASS | No shell interpolation (all `exec.Command` argv form); no secret handling introduced; destructive path is guarded by merge → dirty → anchor → refusal → ignored-content, each fail-closed. Weakened by F2 (silent fail-open on the authoritative lock source) and F5/F6 (allowlist classifies a credential-bearing file and the audit-evidence dir as costless). |
| Craft (20%) | 0.80 | PASS | `go vet ./...` exit 0; `GOOS=windows go vet ./...` exit 0. progress.md §E.2/§E.3 is unusually disciplined (verbatim commands, an explicit re-measured closure of the `internal/cli` gap, a `[HARD]` note that a wrapper `exit=0` was distrusted in favour of the log). Deducted for F1's root cause being a test-seam blind spot and for F4 (new load-bearing predicate untested). |
| Consistency (15%) | 0.78 | PASS | 8 docs-site edits verified row-by-row against `internal/cli/worktree/clean.go:37,42` — `--base` default `origin/main` and `--json` "removes nothing, overrides `--yes`" are stated correctly in all 4 locales. Template mirror byte-identical (`diff` → `MIRROR_IDENTICAL`). Deducted for F3 (guard asymmetry between the two sweeps), F1's unsupported CHANGELOG claim, and F7/F8. |

**Answers to the seven directed attack points** (5 of 7 clean):

1. Three-valued merge detection — **logic correct, data path broken.** `branchMergedForCleanup` (`session_worktree_prmerge.go:342-367`) returns on a determinate gh answer without touching git; no OPEN/CLOSED/DRAFT can reach the fallback. But the fallback itself cannot answer — F1.
2. Anchor guard fail-closed — **every `lockAnchorVerdict` return traced and correct** (`anchor_lock.go:128-147`): unreadable reason → anchored; `!determined` → anchored; alive → anchored; only a positively-confirmed dead pid returns false, and the registry still gets its say. One fail-open remains one level up — F2.
3. `--json` overrides `--yes` — **holds.** `runClean` (`clean.go:64-71`) branches to `reportStaleWorktrees` before `apply` is ever consulted; `reportStaleWorktrees` has no `Remove` call. Caveat F7 (it does call `Prune`).
4. `--base origin/main` vs the base-branch guard — **holds, and is not exploitable.** `isBaseBranch("feature/main","origin/main")` → `false` (branch ≠ base, and ≠ trailing segment `main`). Over-breadth is only in the keep direction. Untested — F4.
5. Regenerable allowlist — **in-tree and fail-closed, confirmed by measurement.** `regenerableIgnoredPaths` at `session_worktree_prmerge.go:242-251`; run against this worktree's real `git status --porcelain --ignored` (13 `!!` entries), `.claude/agent-memory/`, `.moai/reports/session-*.md` and `.moai/specs/…/.moai/` all classify irreplaceable → preserve.
6. Docs-site rows — **accurate.** Verified against the code, all 4 locales, both file families.
7. Windows probe — **matches the fail-closed contract.** `probeProcessLiveness` returns `(true, true)`; in `lockAnchorVerdict` both `alive` and `!determined` lead to anchored, so the mislabel (`determined=true` for a value that is really undetermined) is behaviourally inert. Consequence worth stating: on Windows a locked tree is never reapable. F9.

## Findings

### F1 — [Critical] [blocking] `internal/cli/session_worktree_prmerge.go:432-446` — the git fallback cannot report any live worktree's branch as merged, so M1's headline fix does not work in the case it was written for

`gitBranchMergedReal` strips only the `*` current-branch marker. Git also prefixes every branch **checked out in a linked worktree** with `+`, and the sweep's candidates are by definition checked out in a worktree. Measured on this repository (replay of the exact parser body against real stdout):

```
parsed entries     = 149
still '+' prefixed = 119 (never equal e.branch -> read as NOT MERGED)
matchable          = 30
sample unmatchable = "+ WT-agent-toml-dual"
sample matchable   = "WT-docs-redesign"
```

The 30 matchable entries are branches whose worktrees no longer exist — i.e. never candidates. So `branchMergedForCleanup` returns `mergeStateNotMerged` for essentially every tree it is asked about on the gh-no-answer path.

That path is exactly the one the CHANGELOG advertises as fixed: *"`gh pr view` errors on the ordinary case of a merged PR whose head branch was deleted on the remote … only the no-answer case consults `git branch --merged origin/main`."* gh errors → `ok=false` → fallback → `+`-prefixed line → no match → preserved. The original symptom (worktree preserved forever) still reproduces, now via a different mechanism.

Why every AC still passed: `AC-WR-002` / `TestPRMergeCleanup_GhNoAnswerConsultsGitFallback` swaps the `sessionWorktreeGitBranchMerged` seam and returns fabricated clean branch names. Nothing in the suite exercises `gitBranchMergedReal`'s parser (`grep` over `internal/cli/*_test.go` finds only seam-swap references). The line is byte-identical in `origin/main`, but M1 is the milestone that made this parser load-bearing and asserted it "decides".

**Required fix:** strip the worktree marker as well as the current-branch marker — e.g. `line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "*+ "))`, or switch to `git for-each-ref --merged origin/main --format='%(refname:short)' refs/heads/`, which emits no markers. Add a table-driven unit test over the raw parser covering `"* main"`, `"+ WT-x"`, and `"  WT-y"`, plus one criterion that binds the real function rather than the seam.

### F2 — [High] [blocking] `internal/cli/worktree/clean.go:373-377` — an unreadable worktree porcelain silently drops the authoritative anchor source, with no notice, on both `clean` sweeps

`worktreeLockStates` returns `nil` on any `gitWorktreeCmd` error. Every tree then gets a zero-value `LockInfo`, `lockAnchorVerdict` reports "no opinion", and the decision falls back to the registry alone — the source this SPEC measured at **1 of 5** live anchors (`spec.md` REQ-WR-010). No notice is emitted, so the operator cannot tell the degraded run from a healthy one, and `--yes` will remove a lock-anchored tree, killing a live session's shell — precisely the harm M2 exists to prevent.

The two sweeps disagree about the same failure: `prMergeCleanup` aborts the whole sweep when `git worktree list` fails (`session_worktree_prmerge.go:157-163`, preserve-all), and `AC-WR-016` codifies that as "a porcelain parse failure removes nothing" — but that criterion tests `prMergeCleanup` only. `clean.go` takes the opposite branch. The in-code comment calls it "fail-open on the READ, fail-closed on the DECISION", which does not hold: dropping the authoritative input silently changes the decision.

**Required fix:** make `worktreeLockStates` return `(map, error)`; on error, either abort the sweep as `prMergeCleanup` does, or mark every candidate `Anchored: "undetermined"` with a keep reason. Extend `AC-WR-016` to cover `cleanStaleWorktrees` and `cleanMergedWorktrees`.

### F3 — [High] [blocking] `internal/cli/worktree/clean.go:329-363` + `.claude/rules/moai/workflow/worktree-integration.md` — `clean --stale --yes` has no ignored-content guard, and this change newly documents it as the disposal path for the trees the other sweep cannot reach

REQ-WR-024/P2 exists because `git status --porcelain` and non-forced `git worktree remove` both disregard ignored files, so nothing else stands between a sweep and the destruction of `.claude/agent-memory/`. That guard was added to `prMergeCleanup` only. `staleKeepReason` calls `worktreeHasLocalChanges`, which runs `git status --porcelain` **without** `--ignored` (`clean.go:358`); `grep -n "ignored" internal/cli/worktree/clean.go` returns nothing.

Consequence, measured on this tree: `.claude/agent-memory/` appears as `!! .claude/agent-memory/` and therefore reports `dirty=no`. A merged, unanchored tree whose only ignored content is agent memory is classified removable by `--stale` and destroyed by `--yes`.

The same commit adds a rule-doc section instructing operators to use exactly that command for the non-`WT-` population — the population `spec.md` §G measured as carrying agent memory in 5 trees. REQ-WR-019 shared the *anchor* decision across all three sweeps; the *ignored-content* decision was left un-shared, and the new documentation points users at the unguarded one.

**Required fix:** lift `ignoredContentAllowsRemoval` / `regenerableIgnoredPaths` into a shared package (alongside `internal/session`'s anchor decision) and apply it in `classifyStaleWorktrees`, surfacing it as a fourth predicate on `staleCandidate` so `--json` reports it. Until then, remove or qualify the rule-doc recommendation.

### F4 — [Medium] [blocking] `internal/cli/worktree/clean.go:291-302` — the trailing-segment base guard has no test, though the CHANGELOG calls it load-bearing

The CHANGELOG states the `origin/main` default "would have silently disabled" the base-branch guard, and that the trailing-segment comparison is what keeps it standing. `grep isBaseBranch internal/cli/worktree/*_test.go` returns nothing: `TestCleanStale_BaseDefaultsToOriginMain` (AC-WR-019) asserts only which base string is passed to `IsBranchMerged`. The predicate protecting a second checkout of `main` from being judged merged-into-`origin/main` is unbound by any criterion.

**Required fix:** add a table test over `isBaseBranch` covering `("main","origin/main")→true`, `("feature/main","origin/main")→false`, `("main","main")→true`, `("","origin/main")→false`, and add a `--stale` case with a worktree on `main` asserting `keep_reason == "checked out on the base branch"`.

### F5 — [Medium] [optional] `internal/cli/session_worktree_prmerge.go:242-251` — the allowlist classifies `.moai/state` as regenerable, which includes the repository's own audit-evidence store

`.moai/state/verify/<session>/` is where `agent-common-protocol.md` § Parallel Execution requires verification evidence to be persisted, "so the cited path still resolves at audit time". Measured here: `!! .moai/state/verify/` is ignored, matches the `.moai/state` prefix, and this SPEC's own §E.2 cites `.moai/state/verify/t209/ac-results.txt` — 18 files, present now, and classified costless by the sweep that would delete this tree.

Not a defect in the allowlist's stated derivation rule (the paths were enumerated from measurement), but the classification contradicts a standing doctrine. **Required fix:** narrow the entry to `.moai/state/config-cache.json`, `.moai/state/context-usage.json`, `.moai/state/goal`, `.moai/state/handoff`, or exclude `.moai/state/verify` explicitly.

### F6 — [Low] [optional] `internal/cli/session_worktree_prmerge.go:245` — `.claude/settings.local.json` is allowlisted as costless while `CLAUDE.local.md` §2 describes it as carrying per-machine API tokens

Regenerable in principle (SessionStart rewrites it from `~/.moai/.env.glm`), so the classification is defensible — but an allowlist entry whose loss is asserted to "cost nothing" for a credential-bearing file deserves the reason recorded inline. **Required fix:** one-line comment naming the regenerator.

### F7 — [Low] [optional] `internal/cli/worktree/clean.go:309`, `clean.go:37`, 8 docs-site rows, rule doc — the "`--json` removes nothing" claim is imprecise: `reportStaleWorktrees` calls `WorktreeProvider.Prune()` first

`git worktree prune` deletes administrative records for worktrees whose directories are currently missing — a real mutation of shared repository state, and the reason this audit was instructed not to run the command. progress.md §E.3 Gap 2 discloses it honestly; the user-facing surfaces (flag help, 8 docs-site rows, rule doc) do not. **Required fix:** either skip `Prune` on the `--json` path (stale records can be reported as such), or qualify the claim to "removes no worktree" in the flag help and the 4 locales.

### F8 — [Low] [optional] `internal/cli/session_worktree_prmerge.go:104-139` vs `internal/session/anchor_lock.go:68-89` — two independent parsers for the same `git worktree list --porcelain` lock lines

`parseWorktreeList` and `ParseWorktreeLocks` duplicate the `locked` / `locked <reason>` cases. A future divergence would make the two sweeps disagree about the same lock with no test binding them together. **Required fix:** have `parseWorktreeList` consume `session.ParseWorktreeLocks`, or add a shared-fixture test asserting both agree.

### F9 — [Low] [optional] `internal/session/anchor_pid_windows.go:24-27` — the Windows probe reports `determined=true` for a value it cannot determine

Behaviourally inert (both `alive` and `!determined` route to anchored), and the fail-closed direction is right. The doc-comment claim "existing, definitively" is nonetheless an unobserved claim in a file whose whole purpose is to keep "I do not know" distinguishable from "dead". Operational consequence worth recording where operators will read it: on Windows a locked worktree is never reapable, and stale locks must be cleared by hand. **Required fix:** return `(true, false)` (same behaviour, honest label) and note the Windows consequence in the rule doc.

### F10 — [Low] [optional] `internal/session/anchor_lock.go:154` — `strings.Index(reason, "pid ")` matches inside a word

A lock reason containing e.g. `rapid 1` yields pid `1`, whose liveness probe answers about an unrelated process. Claude Code controls the reason format today, so this is latent rather than live. **Required fix:** anchor the match on a word boundary (`" pid "` / `"(pid "` / prefix), and treat a non-boundary match as unreadable → anchored.

### F11 — [Low] [optional] `internal/template/templates/.moai/config/sections/workflow.yaml:35` — a stale comment states `auto_cleanup` is "declared but not read — no code path consumes"

`prMergeCleanup` reads it (`session_worktree_prmerge.go:150`). The local twin was rewritten in this change with an accurate account; the template twin still carries the false one. Pre-existing, but the two now contradict each other in the same commit's subject area. **Required fix:** correct the template comment (Template-First mirror).

### F12 — [Info] [optional] `.moai/config/sections/workflow.yaml:132` — the safety-critical `auto_cleanup: false` lives inside a path `moai update` wipes

`CLAUDE.local.md` §2.3 records `.moai/config` as a managed root deleted wholesale by `CleanMoaiManagedPaths` before redeploy. The template default is also `false` (verified: `internal/template/templates/.moai/config/sections/workflow.yaml:41`), so the value survives an update by coincidence rather than by protection — but the Korean rationale block explaining *why* this repository must keep it off will be silently deleted. **Required fix:** none in code; consider recording the rationale in the SPEC's §G or a local-only rule file.

## Gaps — what this audit did NOT observe

1. **The 28-criterion AC battery was not re-executed.** Every test-shaped criterion in `acceptance.md` targets `./internal/cli/`, which the dispatch prohibited running (13-minute package). progress.md's claim "every criterion returned 1" is therefore **unverified by me**; I verified the two packages I was authorised to run (`internal/session`, `internal/cli/worktree`) and read the new `internal/cli` test bodies statically. F1 is a criterion-design defect I found by reading, not by re-running.
2. **`clean --stale --json` was never executed against a real repository** — prohibited (it prunes shared state). F3 and F7 rest on code reading plus the measured `git status --porcelain --ignored` output of this worktree, not on the command's actual output.
3. **Windows runtime behaviour** — `GOOS=windows go vet ./...` (exit 0) proves compilation only. Same gap progress.md §E.3 Gap 1 declares.
4. **`golangci-lint` was not run** — not in the authorised command set. Craft's lint limb is unmeasured; only `go vet` was observed.
5. **Sibling worktrees were not inspected.** The isolation guard refuses `git -C` into them, so the 148-tree population is evidenced from `git worktree list` output shared across the repository, not from filesystem checks. This is the same positional limit progress.md §E.3 Gap 3 records, and it is why F3's blast radius (how many trees hold only agent memory) is stated as a mechanism rather than a count.
6. **No CI verdict was read.** The branch is unpushed; the full-suite verdict in a clean environment remains outstanding and is the stronger evidence for AC-WR-022.
7. **F1's fix was not attempted or measured.** I confirmed the defect; I did not confirm that stripping `+` makes the fallback produce correct positives end-to-end.

## Residual risk

- F1 fails in the **preserve** direction — no data is lost, and `auto_cleanup` is `false` in this repository, so nothing is sweeping today. The risk is that the SPEC closes as `completed` while its primary symptom is intact, and the next operator re-enables `auto_cleanup` on the belief that it was fixed.
- F3 fails in the **destroy** direction and is reachable by hand today (`moai worktree clean --stale --yes`), independent of the `auto_cleanup` toggle. It is the finding with real loss potential.
- The check→act race (EC-13) is wider in `cleanStaleWorktrees` than in `prMergeCleanup`: classification of every tree completes before any removal begins, so tree N's anchor state was read before trees 1..N-1 were removed. progress.md §E.3 discloses the race generally; the asymmetry between the two sweeps is not called out.
- My four dimension scores are judgements over a defect set I gathered read-only in one pass; a re-audit that runs the `internal/cli` battery could surface criteria that fail for reasons unrelated to F1–F12.
