# Research: Codify the local develop RC testbed procedure (card t281)

Synthesis of four lenses (codebase-precedent, constraints-risks, prior-SPEC-memory, release-history). Note on provenance: three lenses read the t281 worktree (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t281`); prior-SPEC-memory read the primary checkout (`/Users/goos/MoAI/moai-adk-go`). The two checkouts carry **divergent copies of CLAUDE.local.md and git-workflow-doctrine.md**, which produces the contradictions in the final section — citations below name their checkout wherever it matters.

## 1. Overall shape: this is a consolidation job, not greenfield

All four lenses converge: ~90% of the requested content already exists in tracked files, and the card's real work is (a) authoring two genuinely missing rules, (b) cross-referencing instead of duplicating, and (c) reconciling one superseded premise and one self-contradicting doc target.

Already codified, by location:

- **`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t281/CLAUDE.local.md` §4.1** (lines 300-363 in the worktree copy): the develop-branching model, six HARD disciplines (incl. #4: rc builds cut from develop on operator request via `make build VERSION=vX.Y.Z-rc.N` → clean reinstall → manual test), the §4.1.2 operational procedure (`moai cc -w develop`, `EnterWorktree(.claude/worktrees/develop)`, `git merge --no-ff`, `git push origin develop`), and the t281-reversal provenance note.
- **`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t281/.claude/rules/local/gitflow-lane-protocol.md` §9** (lines 98-109): the existing canonical rc-build runbook — exact commands, the exit-137 [HARD] rule, the bare-`go install` prohibition, recovery step.
- **`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t281/.moai/docs/version-management.md`** (lines 16-65): `-rc.N` SemVer form, ldflags injection, "Local pre-release testing needs no git tag", "Tagging is a separate, remote-facing act performed only by the release harness."
- **CC memory**: `feedback_binary_reinstall_clean` (full rm+cp procedure) and `feedback_branch_guard_worktree_forcepush` (exemption unreachability incident) at `/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`.

No prior SPEC in `.moai/specs/` owns any of the four topics — the memory is entirely doc/memory-tier (prior-SPEC-memory, negative searches confirmed).

## 2. Topic 1 — rc.N numbering for LOCAL builds

**Codified form.** `vX.Y.Z-rc.N` with the dot, `N` starting at `0`, no leading zero (`rc.1`, never `rc.01` — SemVer forbids it; `scripts/release.sh` Validation 1, lines 74-87, rejects it). Dotted `rc.9 < rc.10` orders correctly; legacy undotted `rc10 < rc9` ASCII-lexically — the failure surfaces at the tenth candidate (version-management.md; Makefile; release-history).

**Historical grounding** (release-history, from tag metadata and `git ls-remote`):
- The dotted convention was introduced 2026-08-08 by commit `456193c64` (PR #1412, "enforce SemVer 2.0.0 version grammar").
- v3.1.0 was the first line to use it: `v3.1.0-rc.0` → `-rc.1` → `-rc.2` → `v3.1.0` (2026-08-10 → 2026-08-15).
- **Reset-to-0 per release line is established by history**: v3.1.0's line started at rc.0; the in-flight v3.1.3 line also started over (rc.1 cited in CHANGELOG, rc.5 cited in SPEC-BINARY-LAG-VISIBILITY-001) — and **no v3.1.3 tag exists locally or on remote**, so the entire v3.1.3 rc line is ldflags-injection-only local builds, exactly the pattern the card wants codified.
- Legacy counter-history: v3.0.0 used undotted `rc1`…`rc12` (2026-05-22 → 2026-07-14), numbered from 1, with gaps (no rc3, no rc9) — reason undocumented.

**The genuinely missing rule — NONE found.** All three lenses that searched agree: no in-repo document states an explicit "when to increment N / when to reset to 0" policy for LOCAL (untagged) rc builds relative to the last release tag. The only precedents are the implicit progression (version-management.md line 21: `rc.0 … rc.10 → then vX.Y.Z`) and the observed tag sequences. **This rule must be authored, not transcribed.** The natural shape suggested by the evidence: N increments per candidate build within one target `vX.Y.Z`; N resets to 0 when the target X.Y.Z changes; the next release line's N starts fresh regardless of how far the previous line climbed.

**Tags stay release-harness-only — with one historical counter-precedent.** The doctrine is explicit and mechanical (version-management.md; `scripts/release.sh` is the only tag-creating path). But release-history found that `v3.1.0-rc.0` and `v3.1.0-rc.1` were **real local annotated tags** ("Local-only release candidate … NOT pushed — local testing only"), never pushed to origin — one release line of precedent predating the current "ldflags only, never tag" rule; the v3.1.3 line uses no tags at all. See contradictions §C2 for the cross-lens conflict over who created these tags.

**Ordering trap to bake into the doc.** The Makefile's own comments pin it: `VERSION` derives from `git describe --tags --abbrev=0` (a tag floor, not a build identity), and **an explicit release-candidate VERSION reads HIGHER than a later default build** — comparing `moai version` strings "reaches the opposite conclusion about which binary is newer." Monotone identity lives in the separate `BUILD_ID`. Concrete incident (SPEC-BINARY-LAG-VISIBILITY-001): an installed binary reporting `v3.1.2` was actually newer than the prior `v3.1.3-rc.5` binary. Any rc.N procedure must not order builds by version string.

## 3. Topic 2 — develop refresh after card merges

**NONE found: no documented develop-refresh/regeneration procedure exists anywhere.** Greps for `--merged`, `refresh`, `recreate`, `재생성` across CLAUDE.local.md, git-workflow-doctrine.md, git-local-workflow-doctrine.md, and gitflow-lane-protocol.md returned nothing develop-related (all lenses agree). §4.1 covers merging INTO develop, not refreshing local develop from `origin/develop` beyond the in-window absorb step. This is the second rule that must be authored.

**Supporting precedent that the criterion cannot use `git branch --merged`:**
- SPEC-WORKTREE-SQUASH-MERGE-001 (CHANGELOG line ~917; `internal/core/git/worktree.go` lines 221-389): `git branch --merged` is a reachability test; a squash merge collapses N commits into one new commit on base so the originals never become ancestors. Measured on this repo: 33 of 45 actually-merged squash branches read as unmerged (~11% reclaim). The fixed `IsBranchMerged` predicate composes reachability OR empty-diff OR squash-history signal.
- `worktree-integration.md` line 65/90: the auto-cleanup sweep judges merged-state by gh `MERGED` state with a `git branch --merged origin/<ref>` fallback explicitly annotated "squash-merge blind," and the manual path compares against "the same ref the automatic sweep uses."
- `git-local-workflow-doctrine.md` line 80: gh CLI's auto `git pull --ff-only` after a squash merge fails (local main diverges from the squash commit).
- The lane protocol already keys timing to the remote: a card worktree is discarded only **after the work is on `origin/develop`** (§6 [HARD]) — i.e., origin/develop as the reference point is the repo's existing idiom.

**Caveat on the card's premise** — see contradictions §C4: prior-SPEC-memory notes card→develop merges are mandated `--no-ff` (ancestry-preserving), so the literal claim "squash-merge era makes `git branch --merged` empty" is only partially backed for the develop lane specifically.

**Interaction with BranchGuard:** any criterion phrased as `git branch --merged` will also be denied in the primary checkout (pattern `\bgit\s+branch\b` over-matches read-only forms; recorded workaround `git show-ref --verify refs/heads/<name>` — CLAUDE.local.md §4 friction, line 363 worktree copy). Whether `--merged` specifically is denied is inferred from the pattern description, not from a quoted pattern list (`internal/hook/branch_guard.go` was not opened — constraints-risks gap).

## 4. Topic 3 — BranchGuard-safe path for develop regeneration

**Live state:** `.moai/config/sections/workflow.yaml` lines 149-156 set `branch_guard: enabled: true` (local opt-in; shipped default-false). The guard forbids `git checkout -b` / `git switch -c` / `git branch` in the primary checkout.

**Both exemptions are unreachable from tool-spawned subagents — confirmed by all three lenses that examined this:**
- The `manager-git` AgentType axis fires only for a main-thread `claude --agent manager-git` launch (subagents send no `agent_type` on PreToolUse) — SPEC-WORKTREE-BRANCH-GUARD-001/-OPTIN-001 (REQ-6 preserved the exemption).
- `MOAI_BRANCH_GUARD_EXEMPT=1` is read from the hook process's own environment, spawned before the guarded command runs — exporting it inline is a no-op (main-checkout-branch-guard.md v1.3.1 lines 93-97: "Reading a BRANCH_GUARD_VIOLATION as 'the exemption is broken' is a misdiagnosis — use a worktree instead"; detail companion lines 73-101).

**Canonical route: worktree entry — consensus across lenses.** The guard rules themselves rank it first among "two working routes"; SPEC-WORKTREE-ENTRY-STRATEGY-001 fixes the three layers (`EnterWorktree`/`ExitWorktree` canonical for current-session re-entry, `moai cc -w <name>` canonical launcher, `Agent(isolation: worktree)`); and the repo's existing practice routes all develop work through the launcher-provisioned worktree `.claude/worktrees/develop` (`moai cc -w develop` → `EnterWorktree`; raw `git worktree add` is forbidden — launcher only). **The operator-terminal route (session launched with the sentinel already in its environment) is documented but secondary** — though the recorded incident memory complicates even that route (contradictions §C5). A second, independent guard reinforces the same conclusion: the worktree-session guard denies `git -C .claude/worktrees/develop merge …` from one's own worktree — "entering is the only authorized path" (gitflow-lane-protocol.md §2, line 37).

**Implication for the doc:** the develop-regeneration procedure should be written to run from inside `.claude/worktrees/develop`, never from the primary checkout and never via `git -C`.

## 5. Topic 4 — build + clean reinstall (exit-137 avoidance)

Already codified verbatim in two tracked places plus memory — the card's ask is cross-reference, not re-invention:

```
make build VERSION=vX.Y.Z-rc.N
rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai   # or: make install
~/go/bin/moai version; echo $?                       # must be 0
```

- **[HARD] exit-137**: a bare `cp` overwrite without `rm -f` has produced exit 137 (SIGKILL) on the next invocation (stale inode / buildinfo·mmap residue), even with identical SHA256. Diagnostic signature: `bin/moai version` exits 0 while `~/go/bin/moai version` exits 137. Retry on 137. (gitflow-lane-protocol.md §9 line 107; CLAUDE.local.md known-issues — line 519 in the worktree copy, line 484 in the primary copy; memory `feedback_binary_reinstall_clean`.)
- **[HARD] bare `go install ./cmd/moai` is forbidden**: it omits the Makefile `LDFLAGS` block (Makefile line 20), so `pkg/version` compiles to defaults (`Commit="none"`, `Date="unknown"`), making the `strings <binary> | grep <sha>` binary-lag check impossible. `make install` (Makefile line 72) runs `go install $(LDFLAGS)` and is safe/equivalent.
- **Verify the installed binary, never only `bin/moai`** — the exit-0 check must target `~/go/bin/moai`.
- The rc build is cut **from develop, in the integration worktree, on operator request** (CLAUDE.local.md §4.1 discipline 4 delegates to gitflow-lane-protocol.md §9).

## 6. The superseded premise: develop is a standing REMOTE branch

Live state (constraints-risks, read-only commands): `origin/HEAD -> origin/develop`; local `develop` exists, checked out in exactly one integration worktree `.claude/worktrees/develop` at `fa8ff89ba`. CI triggers were widened to `[main, develop]` (commit `1126d13f`; six workflows: ci.yml, codeql.yml, graph-freshness.yml, lsel-leak-guard.yaml, template-neutrality-check.yaml, test-install.yml). The original card text ("develop stays local, one-shot, no push") is quoted only secondhand in reversal notes — **the card body itself was not located** by any lens.

The reversal's recorded **date and authority diverge by checkout** (contradictions §C1): the t281 worktree copies cite 2026-08-27 and contain no "2026-08-29" string; the primary checkout's CLAUDE.local.md §4.1 carries the [HARD] chain tagged "운영자 지시 2026-08-29" plus the 2026-09-01 [HARD] WT-branch-push ban. Whichever the doc cites, it should verify against the primary checkout's current §4.1.

Residual unverified risk carried by the model: **Vercel production-branch binding** for the docs site is listed 미해소 (unresolved) in §4.1.3 — how preview/production deploys react to develop-side docs-site changes is untested; docs-site-touching cards must check separately.

## 7. Doc-target placement constraints

- **CLAUDE.local.md is over budget**: 47,820 bytes (worktree copy) / 40,545 (primary) against the 40,000-char File Size Limits heuristic from `coding-standards.md`; growth >1,000 bytes triggers `rule-authoring.md`'s statement duty (byte size + cost justification incl. non-invoking sessions), with "scope first" requiring paths-scoped placement. Any §4.1 addition should be a pointer, not a procedure body.
- **SSOT / duplication risk**: gitflow-lane-protocol.md §9 already owns the rc-build procedure verbatim, and the protocol's own discipline is "절차를 여기에 다시 적지 않는다" (never re-write a procedure in a second place — "두 벌이 되는 순간 갈라진다"). Copying it into CLAUDE.local.md §4.1 or version-management.md risks divergence; cross-reference instead.
- **`moai update` wipes local-only files** (CLAUDE.local.md §2.3 [HARD]: 12 files deleted 2026-08-15) and resets `.moai/config/sections/git-strategy.yaml` to template defaults — `rc_version_format: vX.Y.Z-rc.N` is one of three keys that must be re-applied after every update. Anything under `.moai/docs/` (tracked) is safe from the wipe; `.claude/rules/local/*` and `.moai/config/*` are not.
- **git-workflow-doctrine.md is currently self-contradicting as a doc target** (contradictions §C3).
- gitflow-lane-protocol.md and `git-strategy.yaml` are deliberately NOT mirrored to `internal/template/templates/` — private workflow must not ship to the 16-language distribution.

## 8. Honest gaps (NONE found / could not determine)

- **rc.N increment/reset policy for LOCAL builds**: NONE found anywhere — must be authored (all lenses).
- **develop refresh/regeneration procedure**: NONE found anywhere — must be authored (all lenses).
- **Original t281 card body**: not located (no lens searched the todo-queue storage); premise visible only via reversal citations.
- **No prior SPEC** exists for the rc.N convention, the develop chain / integration window, or the exit-137 procedure; `.moai/specs/` greps for "2026-08-29" and "integration acquire" returned nothing.
- **v3.1.3-rc.N per-build reasoning/dates**: no records exist by design (no tags); rc.1/rc.5 are incidental citations only.
- **`v3.1.0-rc.2` push intent**: tag references PR #1512 yet is absent from origin — deliberately deleted vs never pushed could not be determined.
- **Legacy v3.0.0 rc3/rc9 gaps**: reason (skipped vs deleted) undocumented.
- **exit-137 incident's original date/machine state**: not recorded anywhere; only the [HARD] "전례가 있다" assertion; `git log --grep=137` found no dedicated commit.
- **`internal/hook/branch_guard.go` pattern list**: not opened; the `git branch --merged`-specific denial is inferred from the documented pattern description. BranchGuard behavior was not exercised live in the primary checkout (no `BRANCH_GUARD_VIOLATION` triggered).
- **CLAUDE.local.md divergence between checkouts**: worktree 47,820 vs primary 40,545 bytes with shifted line numbering (§4.1 at 300-363 vs 262-331) — the worktree copy is larger yet lacks the primary's 2026-08-29 chain tag; which copy is authoritative post-merge was not determined.
- **Full `~/.claude` memory sweep** (1100+ files) and `scripts/release.sh` git history were not exhaustively examined; a session-transcript-level decision on "worktree vs operator terminal" may exist unsearched.

### contradictions

- **C1 — Date/authority of the develop-to-remote reversal (checkout divergence).** codebase-precedent and constraints-risks (reading the t281 worktree): "no in-tree artifact dated 2026-08-29 — the tracked reversal records cite 2026-08-27"; a grep for `2026-08-29` across the worktree doc targets returned zero matches. prior-SPEC-memory (reading the primary checkout): CLAUDE.local.md §4.1 carries the [HARD] chain explicitly "dated '운영자 지시 2026-08-29'" plus a 2026-09-01 WT-push ban, and cites commit `11216d13f` for the trigger widening. Both cannot describe the same file state; the copies have diverged (sizes 47,820 vs 40,545 bytes; §4.1 at different line ranges). The card's own premise cites 2026-08-29, matching the primary — the doc codification must verify which checkout (and which date/authority) it cites rather than assuming the worktree's 2026-08-27 framing is current.
- **C2 — Who created the dotted rc tags `v3.1.0-rc.0/.1/.2`.** codebase-precedent: "git tag -l shows dotted rc tags only from the release flow (`v3.1.0-rc.0/.1/.2`)" — i.e., attributes them to the release harness. release-history (from tag metadata + `git ls-remote`): `v3.1.0-rc.0` and `-rc.1` were **local-only annotated tags** with messages "Local-only release candidate … NOT pushed — local testing only" / "Local-only RC … NOT pushed", and `-rc.2` (which references PR #1512) is also absent from origin. Under release-history's reading, two of the three dotted rc tags are counter-precedent to the card's "local rc builds never create tags" rule; under codebase-precedent's reading they are harness output. The tag messages are the primary evidence and favor release-history, but the conflict is surfaced, not adjudicated.
- **C3 — `git-workflow-doctrine.md`'s stance on the develop branch.** prior-SPEC-memory (primary copy): lines 8, 45, 344 still **forbid** develop branch creation ("❌ `develop` 브랜치 생성 (Gitflow 패턴)" / "금지"), directly contradicting CLAUDE.local.md §4.1 — and this file is a named doc target of the card. codebase-precedent (worktree copy): cites the same file's line 64 ("[HARD] 2026-08-27 개정 — git-flow … main을 갱신하는 유일한 경로는 릴리스 PR이다") and line 103 (card integration is no-PR `git merge --no-ff` → `origin/develop`) as *supporting* the develop model. Either the two checkouts hold different revisions of the doctrine, or the file internally contains both a revised git-flow section and stale prohibition lines. Either way, treating git-workflow-doctrine.md as already-consistent would be wrong; the card's "possibly" for this target hides a live contradiction that the doc work must resolve.
- **C4 — Applicability of the "squash-merge makes `git branch --merged` empty" premise.** codebase-precedent treats SPEC-WORKTREE-SQUASH-MERGE-001 as "direct precedent that a develop-refresh/recreate criterion cannot rely on `git branch --merged`." prior-SPEC-memory pushes back: the SPEC proves `--merged` misses squash-merged **card branches into base**, but CLAUDE.local.md §4.1 mandates `merge --no-ff` into develop, which **preserves ancestry** — so there is "no prior evidence about which merge shape actually reaches develop in practice" and the card's literal claim is "only partially backed." The origin/develop-based criterion is the safe conclusion under both readings (it is merge-shape-agnostic), but the doc must not assert "--merged is empty" as an established fact for the develop lane.
- **C5 — Reliability of the operator-terminal (sentinel) BranchGuard route.** constraints-risks and codebase-precedent (from the guard rule files v1.3.1): the sentinel route works when "the operator launch[es] the session with the sentinel already in its environment," listed as one of two working routes (ranked second). prior-SPEC-memory (from memory `feedback_branch_guard_worktree_forcepush`, 2026-08-12/13): the env-var exemption is effective only "when the *orchestrator sets it while spawning manager-git*," a SendMessage-resumed manager-git **still could not force-push**, and the recorded working resolutions were (a) the user typing `! git push --force-with-lease …` directly in the prompt, or (b) temporarily setting `branch_guard.enabled: false` and restoring it. The memory evidence undermines the sentinel route's reliability even in its documented form, strengthening the case that worktree entry is the only dependably sanctioned route — but the rules text and the incident memory disagree about whether the second route should be recommended at all.
