---
id: SPEC-KANBAN-WORKTREE-001
title: "Research — measurements underlying the per-card worktree lifecycle"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, worktree, research, measurements, evidence"
tier: L
---

## §A. What this file is, and how to read it

Every measured fact `spec.md`, `plan.md`, and `design.md` rely on, recorded as **command → observed output → what it establishes**, so a later reader re-runs rather than re-derives. It is a Tier L artifact added at v0.2.0 with the promotion.

Two disciplines bind it. **Measurements are time-stamped by their tree, not trusted forever** — every count below is a fact about this repository at authoring time, and a run-phase that finds a different number trusts its own measurement and records the delta. And **an output is recorded even when it contradicts the expectation that motivated the command**; §G exists because three did.

Sections §B through §G were taken at v0.2.0 promotion time on 2026-08-10. Sections §C.1, §D.4, §D.5, §E.3 and §H were added at v0.3.0 on 2026-08-11, from the same worktree; each names its own moment. All commands were run from `/Users/goos/.moai/worktrees/kanban`, whose environment at the v0.2.0 run was:

```
$ git --version
git version 2.50.1 (Apple Git-155)

$ git rev-parse --git-dir
/Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban

$ git rev-parse --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git

$ git branch --show-current
spec-kanban

$ git rev-parse --short HEAD
d39e3cdc6
```

The primary checkout is `/Users/goos/MoAI/moai-adk-go`, and the source tree read below is this worktree's checkout of the same repository.

---

## §B. This repository squash-merges

The measurement `REQ-KW-017` turns on. It is recorded first because the naive way of taking it gives the wrong answer, and the wrong answer is the reassuring one.

```
$ git log --first-parent -200 --format='%p' origin/main | awk '{print NF}' | sort | uniq -c
 200 1

$ git log --first-parent -200 --format='%s' origin/main | grep -cE '\(#[0-9]+\)$'
199

$ git rev-list --first-parent -200 origin/main | wc -l
     200
```

The window's endpoints:

```
$ git log --first-parent -5 --format='%h %s' origin/main
355250a01 docs(SPEC-CODEX-PHASE2-001): close M0 with the protocol probe and its two forced amendments (#1439)
a6471fbb1 docs(SPEC-CODEX-PHASE2-001): resolve the two M0 design forks and clear plan-audit (#1438)
5e2c83238 docs(docs-site): register the resources section, correct counts, and decentralize duplicated content (#1437)
780367f53 docs(docs-site): normalize unprefixed internal links to locale-scoped paths (#1436)
fec21959c docs(SPEC-CODEX-PHASE2-001): plan-phase artifacts for the codex Phase 2 tool surface (#1435)

$ git rev-list --first-parent -200 origin/main | tail -1 | xargs -I{} git log -1 --format='%h %ad %s' --date=short {}
5352c89ba 2026-07-29 fix(config): allow empty user.name so init/update/web run without a name (#1221)
```

**What this establishes.** Over the last 200 first-parent commits of `origin/main` — a window running 2026-07-29 to the present — **every commit has exactly one parent**, so there are **zero** merge commits, and **199 of 200** subjects carry GitHub's `(#N)` squash-merge signature. Pull requests land here as squashes.

This is the ground for `REQ-KW-017` and `design.md` §D.2. The `gh`-absent fallback lists only branches reachable as merged ancestors; a squash produces a new commit with no such ancestry, so in this repository the fallback's consequence is not a degradation but an absolute — with the pull-request observation unavailable, disposal never happens.

**Methodology caveat, recorded because the shorter form is wrong.** The obvious command is not:

```
$ git rev-list --first-parent -200 origin/main --merges | wc -l
      59

$ git rev-list --first-parent -200 origin/main --merges | tail -1 \
    | xargs -I{} git log -1 --format='%h %ad %s' --date=short {}
d91339869 2026-02-09 Merge remote-tracking branch 'origin/fix/community-bugs-batch-1'
```

`-200` bounds the **output**, not the window, so `--merges` returns the 200-most-recent *merges* — reaching back to 2026-02-09, five months beyond the intended window. The `59` is a true count of merge commits over a much longer history (local `Merge remote-tracking branch` commits, which carry no `(#N)`), and it is not the number the requirement needs. Read as a 200-commit window it would contradict the `199` above, since `59 + 199 > 200`. The parent-count form is the one to re-run.

---

## §C. The branch-prefix census

```
$ git for-each-ref --format='%(refname:short)' | sed 's|^origin/||' \
    | grep -oE '^(feat|feature|fix|chore|docs|spec)/' | sort | uniq -c | sort -rn
  64 feat/
  30 docs/
  23 chore/
  18 fix/
   3 feature/
   2 spec/

$ git for-each-ref --format='%(refname:short)' | wc -l
     491
```

**What this establishes.** The repository's dominant branch convention is `feat/` by **64 to 3**. `resolveSpecBranch` synthesizes the minority form (§D.1), so anything keyed on the synthesized name matches almost nothing — the ground for `REQ-KW-003`'s observe-don't-synthesize rule and for `design.md` §C.1.

It is equally the ground for **rejecting** the one-character fix: swapping the literal to `feat/` would make the 3 surviving `feature/` branches the ones that never dispose, inverting the failure rather than removing it. The rule `REQ-KW-003` adopts — recognize a card's branch by the SPEC identifier it carries — is prefix-independent and survives either ratio, which is why a later re-measurement changes the argument's weight without reopening the rule.

**Discrepancy.** `spec.md` §A.2, §D.8 and §E, `plan.md` §B, §C.8, §E and AP-13, `progress.md`, and `acceptance.md` AC-KW-003 all record **63** `feat/`, measured at v0.2.0 authoring time. The re-measurement above is **64**. The delta is one branch created between the two runs; `feature/` is unchanged at 3.

The figures are **not** chased to 64, because each is a correct record of its own measurement moment and the count is expected to keep drifting; the argument rests on the ratio's order of magnitude, and the rule `REQ-KW-003` adopts is prefix-independent and survives either number. What was done instead is attribution: every prose occurrence now names the moment it was taken and points here, so a re-runner who measures 64 reads a drifting count rather than a fabricated one. `plan.md` §C.8 already carried that attribution ("Recorded at plan time as …") and was left untouched. `acceptance.md` AC-KW-003 also carries the figure and is deliberately **not** edited — it is acceptance-criterion text, and the criterion it states does not depend on the count: it requires a `feat/`-prefixed worktree to be recognized, which holds at 63, at 64, and at any other ratio.

The ratio a run-phase acts on is the one it measures itself, per `plan.md` §C.8.

### C.1 Branch-to-identifier multiplicity, suffix shape, and prefix-freedom (v0.3.0)

The three measurements `REQ-KW-003`'s match rule and `REQ-KW-019`'s refusal turn on.

```
$ git for-each-ref --format='%(refname:short)' | sed 's|^origin/||' \
    | grep -oE 'SPEC-[A-Z0-9-]+' | sort | uniq -c | sort -rn | head -5
   5 SPEC-CODEX-PHASE2-001-
   3 SPEC-TOKEN-001-
   3 SPEC-NAVIGATOR-SYNC-003
   2 SPEC-V3R5-STATUSLINE-FMC-001
   2 SPEC-V3R5-ATOMIC-WRITE-001

$ git for-each-ref --format='%(refname:short)' | grep 'SPEC-CODEX-PHASE2-001' | sort -u
docs/SPEC-CODEX-PHASE2-001-fork-resolution
docs/SPEC-CODEX-PHASE2-001-m0-close
feat/SPEC-CODEX-PHASE2-001-run
origin/docs/SPEC-CODEX-PHASE2-001-fork-resolution
origin/docs/SPEC-CODEX-PHASE2-001-m0-close
```

**What this establishes, with one correction to the shorter reading.** The `uniq -c` count is of **occurrences after stripping `origin/`**, not of distinct branch names: the 5 resolve to **three** distinct names, two of the five being `origin/` mirrors. Three is the number that matters, and it is enough — one SPEC identifier, three branches, landing under two different type prefixes. So "the branch naming this card" is not a function, and `spec.md` §A.2's rule needed a cardinality decision it did not have at v0.2.0.

The suffix shape, over distinct post-prefix segments:

```
$ git for-each-ref --format='%(refname:short)' | sed 's|^origin/||' | sort -u \
    | grep -E '^[a-z]+/SPEC-' | awk -F/ '{print $2}' | sort -u > /tmp/segs.txt
$ wc -l < /tmp/segs.txt
      35
$ grep -cE '^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$' /tmp/segs.txt
15
```

**What this establishes.** Of 35 distinct SPEC-carrying branch segments, only **15** are a bare SPEC identifier; the other **20** carry a suffix — `-run`, `-wave5`, `-m0-close`, `-sync-sha-backfill`, `-fork-resolution`, `-pre-rebase`. A match rule requiring equality would refuse the majority of this repository's real SPEC branches, which is why the rule admits a hyphen boundary. It is equally why the rule cannot be containment: containment admits `SPEC-X-0010` and anything else embedding the identifier, and the boundary is the only thing separating the two.

The prefix-freedom check, which decides whether the hyphen boundary's residual is present or merely structural:

```
$ grep -oE '^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}' /tmp/segs.txt | sort -u > /tmp/ids.txt
$ wc -l < /tmp/ids.txt
      31
$ while read a; do while read b; do [ "$a" != "$b" ] && case "$b" in "$a"-*) \
    echo "PREFIX: $a < $b";; esac; done < /tmp/ids.txt; done < /tmp/ids.txt
$                                    # no output
```

**What this establishes.** Across the 31 SPEC identifiers currently appearing on branches, **no identifier is a hyphen-delimited prefix of another**. So the residual `spec.md` §A.2.1 names — a valid identifier `SPEC-X-001-EXTRA-002` matching card `SPEC-X-001` — is structural rather than present. It is re-measured at preflight (`plan.md` §C.13), because a future SPEC identifier could create it, and `REQ-KW-019`'s refusal is what stands between it and a wrongly resolved branch.

---

## §D. The two reused symbols

### D.1 `resolveSpecBranch` — unexported, and it synthesizes the minority prefix

```
$ grep -rn 'func resolveSpecBranch\|func ResolveSpecBranch' internal/
internal/cli/worktree/shared.go:32:func resolveSpecBranch(name string) string {

$ head -1 internal/cli/worktree/shared.go
package worktree

$ sed -n '29,40p' internal/cli/worktree/shared.go
// resolveSpecBranch converts SPEC-ID patterns to branch names.
// e.g., "SPEC-AUTH-001" -> "feature/SPEC-AUTH-001"
// Regular branch names pass through unchanged.
func resolveSpecBranch(name string) string {
	if isSpecID(name) {
		return "feature/" + name
	}
	return name
}
```

**What this establishes.** The symbol is at `internal/cli/worktree/shared.go:32`, in package `worktree` under `internal/cli/`, and is **unexported** (lowercase initial, and the grep finds no exported counterpart). Its body emits the `feature/` prefix that §C measures at 3 against 64. Both facts are cited by `spec.md` §A.2 and §A.9.

### D.2 `branchMergedForCleanup` — unexported, in the command surface, and squash-blind by its own comment

```
$ grep -rn 'branchMergedForCleanup' internal/
internal/cli/session_worktree_prmerge.go:148:		if !branchMergedForCleanup(e.branch, ghAvailable) {
internal/cli/session_worktree_prmerge.go:174:// branchMergedForCleanup decides whether a branch is a cleanup candidate per
internal/cli/session_worktree_prmerge.go:179:func branchMergedForCleanup(branch string, ghAvailable bool) bool {

$ sed -n '174,193p' internal/cli/session_worktree_prmerge.go
// branchMergedForCleanup decides whether a branch is a cleanup candidate per
// REQ-SW-023. Primary path (gh available): state == "MERGED". Fallback path
// (gh absent): branch appears in `git branch --merged origin/main`. The
// fallback is squash-merge blind — squash-merged branches are NOT listed, so
// the worktree is preserved (documented via the on-entry blindness notice).
func branchMergedForCleanup(branch string, ghAvailable bool) bool {
	if ghAvailable {
		return sessionWorktreeGhPRViewState(branch) == "MERGED"
	}
	merged, err := sessionWorktreeGitBranchMerged()
	if err != nil {
		// Fail-open: a --merged query failure means we cannot confirm merge
		// state → preserve (do not risk removing an unmerged worktree).
		return false
	}
	for _, b := range merged {
		if b == branch {
			return true
		}
	}
	return false
}
```

**What this establishes.** Three facts, each load-bearing for a different requirement.

- The symbol is **unexported**, in package `cli` — the command surface.
- Its signature takes **one branch name plus one boolean** and returns **one bool**. Two pull-request identities are not recoverable from that, so the two-PR gate of `REQ-KW-007` is not expressible through it. This is why the gate is keyed on discovered identities instead (`design.md` §D.1).
- Its own comment records the `gh`-absent fallback as **squash-merge blind**, in exactly those words — the documentary half of §B's measured half, and the ground for `REQ-KW-017`.

### D.3 The import direction that forbids exporting it

```
$ grep -n 'internal/cli' cmd/moai/main.go
11:	"github.com/modu-ai/moai-adk/internal/cli"

$ go list -deps ./internal/worktree | grep moai-adk
github.com/modu-ai/moai-adk/internal/worktree

$ go list -deps ./internal/cli/worktree | grep moai-adk
github.com/modu-ai/moai-adk/internal/foundation
github.com/modu-ai/moai-adk/internal/core/git
github.com/modu-ai/moai-adk/internal/tui/internal
github.com/modu-ai/moai-adk/internal/tui
github.com/modu-ai/moai-adk/internal/worktree
github.com/modu-ai/moai-adk/internal/cli/worktree
```

**What this establishes.** `cmd/moai/main.go` imports `internal/cli`, so `internal/cli` is the command surface and the kanban command will live there — the dependency must run `internal/cli` → `internal/kanban`. A kanban package importing `internal/cli` to reach `branchMergedForCleanup` closes that loop into an import cycle the compiler refuses. This is `REQ-KW-018`'s constraint and `design.md` §F.

The second and third commands establish that an extraction — rather than an export in place — is what `REQ-KW-018` needs, and they show `internal/cli/worktree`'s dependency set, which includes both `internal/worktree` and `internal/core/git`. **They do not settle which of those is the target**, and at v0.2.0 this section was read as though they did: `go list -deps ./internal/worktree` reporting only itself is a fact about that package's dependencies, not about its purpose, and purpose is what decides a home. §D.5 settles the target on the packages' own `doc.go` files, and rejects `internal/worktree` as the L1 state guard `spec.md` §C excludes.

```
$ grep -rn '"github.com/modu-ai/moai-adk/internal/cli/worktree"' --include='*.go' .
internal/cli/inventory_test.go:12:	wtroot "github.com/modu-ai/moai-adk/internal/cli/worktree"
internal/cli/coverage_improvement_test.go:24:	"github.com/modu-ai/moai-adk/internal/cli/worktree"
internal/cli/fang_characterization_test.go:12:	"github.com/modu-ai/moai-adk/internal/cli/worktree"
internal/cli/inventory.go:17:	wtroot "github.com/modu-ai/moai-adk/internal/cli/worktree"
internal/cli/root.go:14:	"github.com/modu-ai/moai-adk/internal/cli/worktree"
```

Confirming that both are production files rather than tests:

```
$ head -1 internal/cli/root.go
package cli

$ head -1 internal/cli/inventory.go
package cli
```

**Discrepancy — corrected in place, and it cuts toward the decision rather than against it.** `spec.md` §A.9 stated at v0.2.0 that `internal/cli/worktree` "does not import `internal/cli`, and nothing outside test files imports it from `internal/cli` either." The first half is confirmed by the `go list -deps` output above. The second half was **false as measured**: `internal/cli/inventory.go:17` and `internal/cli/root.go:14` are production files in `package cli` importing `internal/cli/worktree`. §A.9 is corrected to carry the measured fact with these two line anchors.

The conclusion §A.9 draws is unaffected and is left standing — exporting `resolveSpecBranch` in place would still compile, because the edge that would cycle (`internal/cli/worktree` → `internal/cli`) does not exist in either direction of this measurement:

```
$ go list -deps ./internal/cli/worktree | grep -x 'github.com/modu-ai/moai-adk/internal/cli'
$ echo $?
1
```

What the correction does is **strengthen** the rejection recorded in `design.md` §F: `internal/cli/worktree` is not merely adjacent to the command surface, it is a live production dependency of it, so pointing a domain package at it is a firmer edge to avoid than the original prose claimed. No requirement or criterion text depends on the false half, and none is changed.

### D.4 `IsBranchMerged` — squash-aware, `gh`-free, and inside the mechanism this SPEC adopts (v0.3.0)

The measurement that falsifies v0.2.0's rejection premise for a degraded disposal path.

```
$ grep -n 'IsBranchMerged' internal/core/git/types.go
192:	// IsBranchMerged checks whether a branch has been fully merged into
194:	IsBranchMerged(branch, base string) (bool, error)

$ grep -n 'func (w \*worktreeManager) IsBranchMerged' internal/core/git/worktree.go
233:func (w *worktreeManager) IsBranchMerged(branch, base string) (bool, error) {

$ sed -n '208,222p' internal/core/git/worktree.go
// IsBranchMerged reports whether a branch's changes relative to its merge-base
// with base are present in base's HEAD tree, irrespective of the merge strategy
// that placed them there (SPEC-WORKTREE-SQUASH-MERGE-001 REQ-WSM-001).
//
// The predicate is an ordered OR over five signals, cheapest first:
//
//	S1  reachability     git branch --merged <base> lists <branch>
//	S2  empty diff       git diff --quiet <merge-base> <branch> reports none
//	S3  rebase-merge     git cherry <base> <branch>, non-empty & all '-'
//	S4  squash-merge     synthetic-commit git cherry, non-empty & all '-'
//	S5  state check      every branch-touched path is identical in base HEAD

$ grep -rn '\bgh\b' internal/core/git/*.go | grep -v _test
$                                    # no output

$ grep -rn 'IsBranchMerged' --include='*.go' internal/ | grep -v '_test' | grep -v 'internal/core/git'
internal/cli/worktree/clean.go:82:		merged, err := WorktreeProvider.IsBranchMerged(wt.Branch, base)
internal/cli/worktree/clean.go:215:	merged, err := WorktreeProvider.IsBranchMerged(branch, base)
```

**What this establishes.** Four facts, and the first three jointly retire a premise.

- The method is **exported on the `WorktreeManager` interface** (`types.go:194`), which is the interface `spec.md` §A.2 adopts "as the mechanism rather than a new one" and §E enumerates by method name — twice, `IsBranchMerged` included.
- It is documented as reporting merge **"irrespective of the merge strategy that placed them there"**, and its signal list carries **S4, a dedicated squash-merge probe** conjoined with a state check. So the claim that no available non-`gh` path can see squash merges is false of this predicate; it is true only of the bare `git branch --merged` listing, which is this predicate's S1 alone and the neighbouring fallback's whole implementation (§D.2).
- Its package contains **zero** `gh` invocations, so nothing about it depends on the observation `REQ-KW-017` is written for the absence of.
- It is **live**, gating `moai worktree clean --stale` through two call sites, and its own doc comment calls it "a safety-critical predicate… a false positive destroys unmerged user work with no undo", carrying eight guards for that argument.

**Discrepancy against v0.2.0, and it is a defect of method rather than of outcome.** `spec.md` §A.4.1 at v0.2.0 rejected every degraded path on squash-blindness. That premise is retired here. The outcome does **not** change: the gate requires at least two pull-request identities, and this predicate's subject is a branch — a branch count is not a pull-request count in either direction (§C.1 measures three branches for one card), and a content question is not an identity question. `spec.md` §A.4.1 and `design.md` §D.2 are rewritten to argue the outcome on arity. Both files' conclusions stand; both files' reasons are replaced.

**Methodology note, recorded because it is the reusable part.** The falsified premise concerned a predicate the SPEC had already named twice in its own adoption list. A grep for `IsBranchMerged` across the six artifacts returns two hits, both bare enumerations — and **zero** in §A.4, §A.4.1, §A.9, `design.md` §D or `research.md` §D, which are the sections that decided merge detection. Naming a mechanism in an adoption list is not the same as having read it.

### D.5 The extraction target, chosen by reading `doc.go` (v0.3.0)

```
$ sed -n '1,6p' internal/worktree/doc.go
// Package worktree provides working tree state guard primitives for the MoAI
// orchestrator. It captures Snapshots of working tree state, computes Divergence
// between pre/post states, logs divergences to .moai/reports/worktree-guard/,
// and writes SuspectFlags when an Agent(isolation: "worktree") response shows
// an empty worktreePath.

$ sed -n '1,7p' internal/core/git/doc.go
// Package git provides Git repository operations for MoAI-ADK.
//
// It implements three main interfaces:
//   - Repository: read-only operations on a Git repository
//   - BranchManager: branch lifecycle and conflict detection
//   - WorktreeManager: Git worktree management for parallel development

$ go list -deps ./internal/core/git | grep moai-adk
github.com/modu-ai/moai-adk/internal/foundation
github.com/modu-ai/moai-adk/internal/core/git

$ ls internal/kanban
ls: internal/kanban: No such file or directory
```

**What this establishes.** `internal/worktree` is the **L1 worktree state guard** — the mechanism `spec.md` §C excludes by name so an implementer does not wire it by mistake. v0.2.0 selected it as the branch-derivation home on the strength of its dependency count and cited `doc.go:7`, one line below the sentence above. Its leaf status is true and was the wrong criterion.

`internal/core/git` declares branch lifecycle among its subjects, carries `internal/foundation` as its only internal dependency (so it imports neither consumer and cannot import kanban), and is already imported by `internal/cli/worktree` per §D.3's dependency list. It is also the home of `WorktreeManager` and `IsBranchMerged` (§D.4), so the extraction consolidates. This is the target `REQ-KW-018` names at v0.3.0.

**Discrepancy.** `spec.md` §A.9 at v0.2.0 also wrote "the leaf both consumers already reach". The fourth command shows `internal/kanban` does not exist, so one of the two consumers reached nothing. Corrected in `spec.md` §A.9.1 and `design.md` §F.

---

## §E. The two lock substrates

### E.1 The in-process one, and its own comment forbidding promotion

```
$ sed -n '12,26p' internal/lockfile/lockfile_windows.go
// @MX:NOTE: [AUTO] Windows in-process-mutex limitation preserved verbatim from the
// pre-migration internal/cli/team_spawn_lock_windows.go (SPEC-AGENT-TEAM-RETIRE-001
// REQ-ATR-001 — behavior preservation is the contract; do NOT silently "upgrade"
// this to LockFileEx).
//
// fileLocks holds in-process mutexes keyed by absolute file path.
// Windows lacks portable advisory file locks (no fcntl/flock equivalent in stdlib),
// so we fall back to process-local mutexes. This means:
//   - Concurrent writes within the SAME process are serialized (safe for tests and
//     tmux teammates that run as separate processes within the same OS user session)
//   - Concurrent writes across DIFFERENT OS processes are NOT protected
//     (acceptable limitation: ClaimTask is primarily exercised by tmux-based team
//     workflows, which are macOS/Linux-only; Windows users run solo mode)
var fileLocks = map[string]*sync.Mutex{}
```

**What this establishes.** `internal/lockfile` is a `map[string]*sync.Mutex` on Windows whose own comment states cross-process writes are **not** protected and forbids upgrading it. Kanban sessions are distinct OS processes, so it would hold nothing in production while passing every same-process test — the measured reason `REQ-KW-013` selects the other family (`design.md` §E.1), and the reason `AC-KW-014` demands a **separate-process** contention test rather than goroutines.

### E.2 The cross-process one, and the stale-lock gap on Windows

```
$ sed -n '1,8p' internal/spec/lock_windows.go
//go:build windows

// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 — Windows per-SPEC close lock.
// Windows lacks fcntl-style advisory flock; we use atomic-create-file (O_CREATE|O_EXCL)
// per design.md §D.2 fallback. Stale lock detection (PID + timestamp embedded) is a
// post-MVP enhancement; M1 leaves stale-lock cleanup as a known-issue requiring
// manual `del .moai/state/spec-close-*.lock`.
package spec
```

**What this establishes.** The `internal/spec` family is genuinely cross-process on Windows (atomic create, `O_CREATE|O_EXCL`) — hence the reuse — but performs **no stale-lock detection**, with manual `del` as the documented workaround and PID-plus-timestamp embedding named as the post-MVP enhancement.

That gap is what `REQ-KW-014` escapes. Because this SPEC's recovery path is itself a holder change requiring the card's lock, inheriting the limitation would ship a deadlock: a dead holder's artifact blocks the very operation that recovers the card (`design.md` §E.2). The comment's own suggested mechanism — an embedded identity — is what the escape adopts, gated on a positive absence probe rather than on age.

### E.3 The Unix half, and why the deadlock is platform-asymmetric (v0.3.0)

```
$ sed -n '14,40p' internal/spec/lock_unix.go
// flockSpecLock holds an open file descriptor with flock(LOCK_EX|LOCK_NB) held.
type flockSpecLock struct {
	fd int
}

func (f *flockSpecLock) release() error {
	if f == nil || f.fd == 0 {
		return nil
	}
	// Close releases the flock atomically.
	err := unix.Close(f.fd)
	f.fd = 0
	return err
}

// acquireSpecCloseLockImpl opens lockPath O_CREAT|O_RDWR and applies a
// non-blocking exclusive flock. Returns ErrSpecCloseLockHeld on contention.
func acquireSpecCloseLockImpl(lockPath string) (specCloseLockImpl, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	...
}

$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/state/spec-close-*.lock | wc -l
      14

$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/state/spec-close-*.lock | sort -k6,7 | head -1
-rw-r--r--@ 1 goos  staff  0 May 30 12:37 …/spec-close-SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001.lock
```

**What this establishes.** The Unix implementation holds the lock on an **open descriptor**, and its own comment records that closing releases it atomically — so the kernel releases the lock when the holding process exits, and the release path never unlinks the file. The consequence is directly observable: the primary checkout's `.moai/state/` holds **14** zero-length `spec-close-*.lock` artifacts, the oldest dated **2026-05-30**, and every one of them is inert. A file survives; a lock does not.

So the stuck-card deadlock `REQ-KW-014` escapes is **Windows-only**. That is not a softening of the requirement — the escape stands unchanged — but it changes how the requirement must be argued and how it must be judged. `spec.md` §A.6.1 states the asymmetry rather than describing the hazard as universal, and `AC-KW-015`'s rows record the platform they ran on, because a green macOS run exercises only the half where the artifact cannot block.

**This is also the ground for the race repair.** On Unix an artifact can be released and re-created by a live acquirer with no visible change to the path, and on Windows the same sequence is an unlink of a valid lock. A clearing act that inspects, probes, then unlinks has no way to notice either. `REQ-KW-014` at v0.3.0 conditions the removal on the artifact still being the one inspected (`design.md` §E.2.1).

---

## §F. The liveness inputs

### F.1 The registry fields the probe resolves through

```
$ sed -n '82,96p' internal/session/registry.go
// Entry is a single row in the active-sessions registry.
//
// Schema is frozen per REQ-COORD-002 and REQ-COORD-024. Any modification
// requires a follow-up SPEC superseding REQ-COORD-024.
type Entry struct {
	SessionID     string    `json:"session_id"`
	SpecID        string    `json:"spec_id"`
	Phase         string    `json:"phase"`
	StartedAt     time.Time `json:"started_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	PID           int       `json:"pid"`
	Host          string    `json:"host"`
	CWD           string    `json:"cwd"`
}
```

**What this establishes.** The registry already carries exactly the three fields `REQ-KW-011`'s probe needs — `SessionID` to key on, `PID` to probe, `Host` to decide whether the probe is meaningful on this machine. No field is added anywhere, by this SPEC or to the card record.

It also shows `LastHeartbeat` present and deliberately **unused** as a decider (`design.md` §B.2): a stopped heartbeat is ambiguous between a dead session and a stalled writer, an absent process is not. And the frozen-schema note is why this SPEC consumes the registry rather than extending it.

### F.2 The stall threshold has no second constant to compare against

```
$ grep -rn '14400' --include='*.go' internal/ pkg/
$ echo $?
1

$ grep -n '14400' .claude/skills/moai/workflows/factory.md
96:- **Bound it with the flags, never with prose.** The preset arms with `--max-turns 0 --max-duration 14400` — infinite turns, a four-hour wall clock. …
```

**What this establishes.** The four-hour wall-clock bound the 21600-second default is reasoned against exists **only as prose** in a workflow document; nothing in Go declares it. A criterion asserting a relationship between two constants would therefore be unsatisfiable, which is why `REQ-KW-010` is checked as a configuration property (the shipped default reads 21600; a non-positive value is refused by name) and the relationship to 14400 is documentation-anchored, with the resulting weakness named in `spec.md` §A.5 rather than hidden.

`plan.md` §C.5 re-runs this grep, and `plan.md` AP-3 names the route back into the unsatisfiable form.

### F.3 The established two-pull-request precondition

```
$ grep -n 'BOTH run PR AND sync PR' .claude/rules/moai/workflow/spec-workflow.md
437:- Pre-condition: BOTH run PR AND sync PR are in MERGED state (verify via `gh pr view <PR>`)
```

**What this establishes.** The repository already states the same precondition for the same act, and states it against pull-request **numbers** rather than a branch name. `REQ-KW-007` adopts that form rather than inventing one, which is also the measured reason a branch-keyed predicate cannot serve (§D.2).

---

## §G. The predecessor carried no clean-tree assignment fence

```
$ ls .moai/state/kanban-source/SPEC-KANBAN-MULTISESSION-001/
acceptance.md  design.md  plan.md  progress.md  research.md  spec.md

$ grep -rniE 'clean tree|clean-tree|working tree is clean|tree is clean|concurrent assignment' \
    .moai/state/kanban-source/SPEC-KANBAN-MULTISESSION-001/
plan.md:201:- D3's recovery: release the holder, leave the column, and gate re-dispatch on M3's orphan classification — immediate for a clean tree, human-cleared for a dirty one (REQ-KM-049, REQ-KM-050).
acceptance.md:159:**AC-KM-061** (REQ-KM-050) — *Given* a released card whose orphaned worktree is clean, *when* the lead evaluates it for dispatch, *then* it is dispatchable and the target tree is that same worktree. …
```

**What this establishes.** Two hits across all six predecessor files, and **both are orphan classification** — the re-dispatch gate, which this SPEC keeps as `REQ-KW-012`. Neither is an assignment fence, and no exclusion mechanism for concurrent assignment appears anywhere in the predecessor.

So `REQ-KW-013`'s lock is an **addition**, closing a gap that was open in the predecessor, not a replacement for a weaker mechanism. The distinction matters for review: "we replaced a weak fence with a strong lock" invites a comparison, while "there was no exclusion and now there is" invites the question of what else the split inherited as absent — which is the question that found the sibling's own F1 defect.

**Discrepancy, already consumed.** v0.1.0's §A.6 asserted the predecessor carried such a fence and that this SPEC rejected it. That claim is false, and it is corrected in `spec.md` HISTORY and §A.6 at v0.2.0. The clean-tree fence is still **rejected as a design** (`design.md` §B.1) — it is simply rejected as an option nobody had adopted, rather than removed from an inheritance.

---

## §H. The `worker-` sweep, and the frontmatter cycle (v0.3.0)

### H.1 What `moai cc` removes, and what it leaves

```
$ grep -n 'cleanupMoaiWorktrees' internal/cli/launcher.go
227:	worktreeMsg := cleanupMoaiWorktrees(root)
478:// cleanupMoaiWorktrees removes moai-related git worktrees from both the
481:func cleanupMoaiWorktrees(projectRoot string) string {

$ sed -n '486,504p' internal/cli/launcher.go
	// 1. Local Claude Native worktree path.
	localBase := filepath.Join(projectRoot, ".claude", "worktrees")
	...
	// 2. Global ~/.moai/worktrees/*/ paths (MoAI worktree migration target).
	if homeDir, err := os.UserHomeDir(); err == nil {
		globalBase := filepath.Join(homeDir, ".moai", "worktrees")
		if entries, err := os.ReadDir(globalBase); err == nil {
			for _, entry := range entries {
				if entry.IsDir() { basePaths = append(basePaths, …) }

$ sed -n '531,536p' internal/cli/launcher.go
		workerName := filepath.Base(worktreePath)
		if !strings.HasPrefix(workerName, "worker-") {
			continue
		}
		for _, base := range basePaths {
```

**What this establishes, and it inverts the shorter reading.** `applyCCMode` calls this unconditionally at `:227`, so it runs on every `moai cc` launch. The function assembles two kinds of base path — `.claude/worktrees/` and each directory under `~/.moai/worktrees/`, the second being this SPEC's L2 home — which reads as a threat to every card tree.

It is not, and the ordering is why: the **`worker-` prefix filter is applied before the base-path loop**, so it gates both bases equally. Directories under `~/.moai/worktrees/` are enumerated only as *containers to scan*, never removed themselves. A card worktree whose base name is its SPEC identifier does not match the filter and is skipped entirely.

A second layer confirms it. Removal goes through `removeWorktree`, which omits `--force` by a documented decision recorded at `:575` — so even a matching tree holding uncommitted work is kept, and reported as kept.

**So the finding is a naming prohibition rather than a hazard**: `REQ-KW-003` requires the card worktree's base name to be the SPEC identifier and never a `worker-` prefix, because adopting that convention — which is what the team-worktree code suggests — would place every card tree inside the sweep radius of a routine command, silently for a clean tree. `SPEC-KANBAN-BOOTSTRAP-001` records the same constraint from its side and assigns it here.

### H.2 The mutual `dependencies:` declaration

```
$ grep -n '^dependencies:' .moai/specs/SPEC-KANBAN-WORKTREE-001/spec.md \
    .moai/specs/SPEC-KANBAN-BOARD-001/spec.md
SPEC-KANBAN-WORKTREE-001/spec.md:15:dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001]
SPEC-KANBAN-BOARD-001/spec.md:15:dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-WORKTREE-001]
```

**What this establishes.** Each SPEC names the other, which is the cycle `spec.md` §A.4 forbids and the bootstrap edge was deliberately shaped to avoid. It is new: the board sibling promoted this SPEC out of `related_specs:` at its own v0.3.0, because `REQ-KB-020` consumes `REQ-KW-003`.

Both consumptions are real, and they differ in kind. This SPEC's is a **landing** dependency — `REQ-KW-002` gates on the card record's holder field existing, and nothing else substitutes for it. The board's is a **contract** dependency — `REQ-KB-020` consumes an identification rule readable from `spec.md` §A.2 and §A.2.1 with no code of this SPEC's having landed, exactly as this SPEC consumes `REQ-KS-006`. The declared edge belongs on the landing dependency, so the board's entry is the one to move.

**Not repaired here, deliberately.** `AC-KW-001` carries the observation and records that it currently fails. A SPEC that deletes its own dependency entry to break a cycle its sibling created leaves a real prerequisite undeclared and no record of the trade — which is the failure mode `design.md` §G exists to prevent.

---

## §I. Out of Scope

### Out of Scope — what this file does not measure

- The runtime behavior of any worktree-lifecycle implementation. Nothing here exists yet; every measurement is of the repository, not of the feature.
- Whether either prerequisite has landed. `internal/kanban` does not exist and `SPEC-KANBAN-BOARD-001` is an unlanded draft; `plan.md` §C commands 1 and 2 are the gates, and `plan.md` §B records the plan-time state.
- The correctness of `SPEC-KANBAN-RENAME-001`'s rename mapping. Only whether it landed is measured, and only at preflight.
- Windows behavior of either lock substrate. §E reads their source and their documented contracts; no Windows execution was performed, and `AC-KW-014` / `AC-KW-015` are judged on the platform whose implementation lacks stale detection.

### Out of Scope — measurements deferred to run-phase

- The branch-guard baseline (the primary checkout's branch and HEAD). `spec.md` §A.8 records `main` @ `b59a8ba7d` at v0.1.0 authoring time; `origin/main`'s tip has since moved to `355250a01`. The baseline is **re-read at run-phase and never carried forward** from any document, per `plan.md` §C.3 — which is why the drift is noted here rather than propagated into the requirement.
- Each mirrored template pair's byte-identical-versus-sanitized classification. It is time-varying and is re-measured at run-phase per `spec.md` §D.10; `plan.md` §D records the plan-time picture as a starting point, not a standing fact.
- The branch-prefix ratio (§C) and the squash-merge census (§B). Both are re-run at preflight (`plan.md` §C.8, §C.11); a changed measurement reopens the section it grounds rather than halting the milestone.
- The branch-identifier multiplicity and prefix-freedom census (§C.1). Re-run at `plan.md` §C.13. The prefix-freedom half is the one to watch: a future SPEC identifier can turn §A.2.1's structural residual into a present one.
- The `worker-` sweep's two properties (§H.1) — filter-before-base-loop, and non-force removal. Re-run at `plan.md` §C.15; a change to either reopens `spec.md` §A.2.2's survival conclusion.
- The frontmatter cycle (§H.2). Re-run at `plan.md` §C.16, and expected to still fail until the board sibling is corrected. It is surfaced, never repaired from this side.
- Windows execution of the clearing operation's concurrent re-acquisition row. Nothing here was run on Windows; §E.3 establishes only that the Unix half cannot exhibit the defect, which is why `AC-KW-015` requires its rows to record their platform.

---

## §J. Cross-references

- `spec.md` §A.2, §A.2.1, §A.2.2, §A.3.1, §A.4, §A.4.0, §A.4.1, §A.5, §A.6, §A.6.1, §A.7, §A.7.2, §A.9, §A.9.1 — the context each measurement supports, and §B, the requirements it grounds.
- `design.md` §B … §G — the decisions each measurement forced, with the alternatives it eliminated.
- `plan.md` §B (known issues) and §C (the sixteen preflight commands that re-run these measurements).
- `acceptance.md` AC-KW-001, AC-KW-002, AC-KW-010, AC-KW-011, AC-KW-014, AC-KW-015, AC-KW-017, AC-KW-018, AC-KW-019, AC-KW-023 — the criteria that consume them.
- `SPEC-KANBAN-BOARD-001` `research.md` §G, §I — the sibling's measurements of the same two lock substrates and of the disclaimer pair whose shape `design.md` §G is written against.
