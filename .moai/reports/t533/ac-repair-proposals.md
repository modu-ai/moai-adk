# t533 — AC repair wordings drafted in run, for the lead to adjudicate

Card t533 · SPEC-CODEX-GHOST-SKILLS-MEASURE-001 · worktree `.claude/worktrees/t533` · HEAD at drafting `6a46c0edbe2dec6014c685184aa2cf9dc346cc59`

**This file proposes; it does not apply.** `acceptance.md` body content is not this agent's to edit (the run-phase agent owns `progress.md` and implementation files; AC body wording belongs to the plan-phase owner). The three plan-audit debts were closed in run **as evidence** — the corrected selectors were the ones actually executed, and their outputs are the recorded measurements. What remains is the AC *text*, which still names the defective forms. Each proposal below is written so it generalizes past this card, because two of the three defects are produced by the lane procedure and the agent environment rather than by this SPEC.

---

## Proposal 1 (N6) — replace the mtime selector; it has a measured hole

**Where**: `acceptance.md` AC-CGM-001, the second snapshot command.

**Current**

```
stat -f '%N %m' ~/.codex/* ~/.codex/.[!.]* 2>/dev/null | sort > .moai/reports/t533/codex-mtimes.before
```

**Proposed**

```
find ~/.codex -maxdepth 1 -mindepth 1 -exec stat -f '%N %m' {} + | sort > .moai/reports/t533/codex-mtimes.before
```

**Why, with the measurement.** The glob pair cannot match a name beginning with `..` — `DIR/*` excludes dot-names and `DIR/.[!.]*` requires the second character not to be a dot. On the live tree: entry selector **72**, glob selector **70**, the difference being two real `..codex-global-state.json.tmp-*` files. The replacement yields **72**. A mutant confirms the mechanism rather than only the count: creating a `..`-prefixed name in an isolated directory leaves the glob selector silent (rc=0) and turns the `find` selector red (rc=1).

**A second reason the replacement is the better form, found while reproducing the first.** Under zsh a glob with no matches aborts the command line, so if either glob happens to match nothing the selector emits **nothing at all** while the redirection still creates an empty file — and in an absence-as-PASS check an empty selector reads as "no change". `2>/dev/null` does not suppress it; the failure precedes `stat`. The `find` form has no such mode.

**Generalization worth carrying beyond this AC**: *a directory-enumeration selector built from shell globs has two failure modes that both present as silence — names the glob cannot express, and a nomatch that aborts the line. Enumerate with `find -maxdepth 1 -mindepth 1`, which has neither.*

---

## Proposal 2 (N5) — the guard's second probe needs its own mutant

**Where**: `acceptance.md` AC-CGM-001, the mutant clause.

**Proposed addition** (the existing single-mutant clause stays; this adds the second):

> **Second mutant — in-place modification.** In the same isolated directory, pin a baseline mtime into the past (`touch -t 202601010000 "$D/a"`), snapshot both probes, then modify `a` in place without changing the name set. The entry probe MUST stay silent and the mtime probe MUST turn red. Both outcomes are required: the silence proves the entry probe cannot see this event, and the red proves the mtime probe can. Record an uncaught mutant rather than deleting it — it draws the guard's boundary.

**Why.** `Then` requires *both* probes silent, so both are absence-as-PASS checks and this card's own `plan.md` D-2 binds both. Only the entry probe had ever been exercised, and the unexercised one is the only probe that catches prune's actual write shape — `os.WriteFile(cfgPath, pruned, 0o600)` (`codex_skills_prune.go:210`) rewrites `config.toml` under its existing name, which the entry probe cannot see by construction. Executed result: entry probe silent (rc=0), mtime probe red (rc=1, `1767193200` → `1788790185`).

**Known residual the wording should not overstate**: `%m` is whole-second, so an in-place write inside the same second as the baseline capture would leave the mtime probe silent. `%Fm` is available and gives nanoseconds (`1788790185.324442879`). This run captured `codex-mtimes-frac.before` with `%Fm` in addition to the AC-named `%m` artifact rather than silently substituting it.

---

## Proposal 3 (N7) — phase B must name only this card's own commits

**Where**: `acceptance.md` AC-CGM-010, phase B.

**Current**

```
git diff --name-only 6a46c0edb..HEAD | wc -l          # control, ≥ 1
git diff --name-only 6a46c0edb..HEAD -- '*.go'        # probe, empty
```

**Proposed**

```
git fetch origin develop
CARD_BASE=$(git merge-base origin/develop HEAD)       # re-derived at read time, never pinned
git diff --name-only "$CARD_BASE"..HEAD | wc -l       # control, ≥ 1
git diff --name-only "$CARD_BASE"..HEAD -- '*.go'     # probe, empty
```

with the recorded fallback: *where `origin/develop` cannot be reached, the literal branch-point SHA is a valid substitute **before** the integration absorb and invalid after it; say which case applies.*

**Why the literal SHA is wrong here even though pinning is usually right.** Pinning `6a46c0edb` is correct as an *anchor* — it is the address at which the measurement was taken. The defect is not the pin, it is the **range's attribution**. The lane procedure absorbs `git merge origin/develop` into the card worktree and re-measures in the merged tree, and from that moment `6a46c0edb..HEAD` contains other lanes' commits, among which `.go` changes are essentially certain. A card that changed no Go line would then be judged FAIL by an AC measuring somebody else's work.

`git merge-base origin/develop HEAD` re-derives the correct left endpoint in **both** states, which is why it generalizes: before the absorb it resolves to the branch point (verified today — it returns exactly `6a46c0edbe2dec6014c685184aa2cf9dc346cc59`), and after the absorb it resolves to the absorbed develop tip, leaving the range equal to the card's own contribution either way.

Under the moving-ref predicate this is remedy **R4**, not a pin: the claim is *"relative to whatever mainline currently is, this card changed no Go file"*, so the moving ref is the claim's subject. The wording therefore leads with the command a reader must run, and treats any recorded value as a dated reference. Pinning it would destroy the very property the AC is trying to assert.

**Generalization — this is a lane-procedure defect, not a t533 defect.** *Any acceptance criterion that measures "what this card changed" against a literal base SHA becomes false the moment the lane absorbs the integration branch. Such criteria should derive their left endpoint with `git merge-base <integration-ref> HEAD` at read time. This applies to every card that carries a scope-limiting AC (`no Go changes`, `no template changes`, `only these paths`), not only to this one.*

---

## Proposal 4 (new in run, same family as N6) — phase A's selector is blind inside untracked directories

**Where**: `acceptance.md` AC-CGM-010, phase A probe and control.

**Current**: `git status --porcelain | grep -c '\.go$'`

**Proposed**: `git status --porcelain --untracked-files=all | grep -c '\.go$'`, with the control taken from the same expanded listing.

**Why, with the measurement.** `git status --porcelain` collapses an untracked directory to a single line. Today the whole working tree reported as just two lines — `?? .moai/reports/t533/` and `?? .moai/specs/SPEC-…/` — so a `.go` file created inside either directory would have been invisible to the probe. Demonstrated with a needle proven present by other means: against the collapsed listing `grep -c '\.md$'` returns **0** although **5** `.md` files exist; against `--untracked-files=all` the same needle returns **5**. The `.go` claim itself is unaffected (the expanded selector also returns 0), but the selector as written could not have detected a violation.

This is the same shape as N6 — a selector whose declared scope exceeds its real coverage, failing silently in the absence-as-PASS direction.

---

## Proposal 5 (new in run) — correct the stated cause of the broken instrument in §F/§F.2

**Where**: `spec.md` §F and §F.2, and `acceptance.md` AC-CGM-009.

**What is wrong.** The documents attribute the silent give-up on `~/.zsh_history` to **locale and file encoding** — "default-locale `grep` treats a `Non-ISO extended-ASCII` file as binary and quietly bails". Measured today, that attribution is false. The real cause is the **agent shell's `grep` shell function**, injected by the Claude Code shell snapshot, which routes to `ugrep` with `-I` (skip files classified as binary) and `--ignore-files`. Evidence:

- In the agent shell: `grep -c 'moai' ~/.zsh_history` → `rc=1 len=0` (nothing printed).
- `type grep` in that shell → *"grep is a shell function"*, whose body execs `ugrep … -I --ignore-files …`.
- Shell functions are not exported to child shells, so the identical line inside `zsh measure.sh` → `rc=0 len=4 out=[1071]`.
- `/usr/bin/grep -c 'moai' ~/.zsh_history` → `rc=0 out=[1071]`. Locale is unchanged across all three (`LANG=C.UTF-8`, `LC_ALL` unset).

**What survives unchanged**, and should be said so plainly: the *lesson* is right and the *prescription* works. A bailed grep prints nothing and is indistinguishable from an empty corpus; `LC_ALL=C grep -a` restores a count — because `-a` overrides the wrapper's `-I`, not because the locale mattered. And the substantive result is independently confirmed on the real binary: `/usr/bin/grep -c 'codex-skills'` → `0`, `'moai clean'` → `0`, control `'moai'` → `1071`.

**Why the correction is worth making rather than absorbing quietly.** The stated cause changes who is affected and what the fix is. As written, the hazard reads as a property of certain *files* that anyone would hit; measured, it is a property of *this agent environment's* shell wiring, invisible to a human running the same command in their own terminal — and it silently extends to `--ignore-files`, which makes a recursive `grep` skip gitignored paths. That is a materially larger hazard than the one recorded, and it applies to every grep-based absence claim any agent makes in this repository, not to this card alone.

**Generalization**: *before resting an absence claim on `grep`, establish which binary actually ran (`type grep`). In this agent environment `grep` is a wrapper that skips binary-classified and ignore-listed files; a control that differs only in needle cannot detect it, because both probe and control go through the same wrapper. A control that differs in **binary** (`/usr/bin/grep`) can.*
