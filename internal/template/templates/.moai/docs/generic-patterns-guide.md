# Generic Patterns Guide

> Externalized generic operational patterns for MoAI-ADK projects. This guide
> distills 4 patterns that are useful across user projects regardless of
> maintainer-specific local doctrine. Each pattern is presented in
> user-audience-neutral prose — if your team adopts the pattern, customize the
> specifics to your environment.
>
> Created per the local-namespace consolidation policy.
> Audience: any MoAI-ADK user project (template-distributed)

---

## Multi-Session Race Mitigation Procedure

When 2+ Claude Code sessions operate on the same project root + same memory
hash (`~/.claude/projects/{project-hash}/memory/`), the sessions may both
consume the same paste-ready resume and attempt the same `/moai <subcommand>`
work in parallel. The git working tree is shared; the memory file is shared.

### Failure mode

Without a pre-spawn fetch, the second session works on a stale baseline and
may produce:

- Duplicate commits (both sessions create the same SPEC commit, one is orphaned)
- Conflicting frontmatter edits (both update `status:` to `in-progress`,
  one overrides the other)
- CHANGELOG entry races (both append the same SPEC entry, manual cleanup needed)
- Push range mismatch (push succeeds with an intermediate commit from the
  other session, detected retrospectively when `git log` shows a mystery commit)

### Defense-in-depth mitigation (4 layers)

**Layer 1 — Pre-spawn fetch obligation (HARD)**: Before spawning any
implementation `Agent()` that will commit or modify shared working-tree files,
execute the canonical 2-command parallel batch:

```bash
# 1. Fetch latest origin/main without merging
git fetch origin main

# 2. Count divergence between local HEAD and origin/main
git rev-list --count --left-right origin/main...HEAD
```

Interpretation matrix:

| Output | Meaning | Action |
|--------|---------|--------|
| `0 N` | Local ahead by N (clean — your commits not yet pushed) | Proceed normally |
| `0 0` | Synced (local == origin/main) | Proceed normally |
| `N 0` | Origin ahead by N — parallel session race detected | STOP, surface via AskUserQuestion: rebase / inspect / abort |
| `N M` | Diverged (both ahead) | STOP, MUST resolve before spawn |

This rule is defined in `.claude/rules/moai/core/agent-common-protocol.md`
§ Pre-Spawn Sync Check.

**Layer 2 — L2/L3 worktree opt-in (recommended for known multi-session
patterns)**: If your team works with 2+ sessions on the same project, use
`/moai plan --worktree` or `moai worktree new SPEC-XXX --base origin/main`
to materialize a per-SPEC working tree. Memory is still shared, but git
working trees are separated → race conflict surface drops to zero.

**Layer 3 — Single-session paste-ready discipline (user-level rule)**: The
paste-ready resume message emitted by the orchestrator at session end should
be pasted into one session only. Other sessions on the same project should
either work on different SPEC IDs or operate in read-only mode (Explore,
diagnostic agents).

**Layer 4 — Retrospective detection (information signal)**: If `git log
--oneline` shows a mystery commit (one you did not author in the current
session for an unexpected SPEC ID), assume a parallel session race occurred.
The commit is already on main, so no recovery is needed — but record the
event so future sessions know to expect the same pattern.

### When this pattern applies to your project

- Your team uses 2+ Claude Code sessions on the same project simultaneously
- You see mystery commits in `git log` you did not author
- You experience occasional duplicate SPEC artifact edits
- You want to formalize multi-developer coordination via worktree isolation

### When to skip

- Single-developer project with strict single-session discipline
- Strict pull-request workflow where every change goes through PR review (the
  PR merge step naturally absorbs the race)

---

## Hook Setup Procedure for New Machines

When a developer clones your project on a new machine, the git hooks under
`.git/hooks/` do NOT travel via `moai update` (git infrastructure is
local-only). If your team uses pre-push hooks, pre-commit hooks, etc., each
machine needs manual setup.

### Example: warn-only pre-push hook for main branch protection

If your team uses a "1-person OSS" or "Hybrid Trunk" workflow where direct
push to main is permitted but warned, create a warn-only pre-push hook:

```bash
cat > .git/hooks/pre-push <<'EOF'
#!/bin/bash
while read local_ref local_sha remote_ref remote_sha; do
  if echo "$remote_ref" | grep -qE "refs/heads/main$"; then
    echo "[warn] Direct push to main detected. 5s delay before push (Ctrl+C to abort)." >&2
    sleep 5
  fi
done
exit 0
EOF
chmod +x .git/hooks/pre-push
```

### Verification (dry-run)

```bash
# Main push: should warn + 5s sleep + exit 0
echo "refs/heads/main 0000 refs/heads/main 0000" | .git/hooks/pre-push

# Feature branch push: should be silent + exit 0
echo "refs/heads/feat/test 0000 refs/heads/feat/test 0000" | .git/hooks/pre-push
```

### When to apply

- Your team has multiple developers cloning the repo on different machines
- You want a soft warning before main pushes (not enforcement — branch
  protection on the GitHub side handles enforcement)
- You explicitly adopt a Hybrid Trunk / 1-person OSS workflow

### When to skip

- Your team uses strict PR-only workflow (the GitHub branch protection
  handles all main-push prevention)
- Your team uses commit signing as the primary discipline (different
  mechanism)

---

## Settings Intent Doctrine

The following Claude Code settings keys are commonly customized at the
project level. This section explains the intent so your team can decide
whether to override the template defaults.

### `defaultMode`

- **Template default**: unset (Claude Code uses `"default"` mode — prompts
  before each tool use)
- **Common customization**: `"acceptEdits"` (auto-accept file edits) or
  `"bypassPermissions"` (skip all permission prompts)
- **When to customize**: developer is comfortable with the agent's edits and
  wants faster iteration. Increases velocity at the cost of permission
  visibility per change.
- **When NOT to customize**: production deployments, security-sensitive code,
  shared multi-developer environments where one developer's auto-accept
  affects review discipline of others.

### `teammateMode` (runtime-managed)

- **Template default**: unset (no teammate mode active)
- **Common customization**: `"tmux"` for Agent Teams + tmux split-pane;
  `"claude"` for Claude-only teammates; `"glm"` for GLM cost-optimization
- **When to customize**: project actively uses Agent Teams or CG Mode for
  parallel/cost-optimized work.
- **Important**: This key is typically set in `.claude/settings.local.json`
  (per-machine, not committed to git) and modified by runtime commands
  (`moai cg`, `moai glm`, SessionStart hook). Do not add it to the project
  `.claude/settings.json` template.

### `env.PATH`

- **Template default**: unset
- **Common customization**: project-specific PATH prepend (e.g., for
  language-specific binaries like Go's `$HOME/go/bin`)
- **When to customize**: project requires specific tool versions or
  custom-built binaries not on the system PATH.
- **Important**: Use environment variable expansion (`$HOME`, etc.) — never
  hardcode absolute user-specific paths in committed settings. This makes
  the settings portable across team members.

### Operating principle

These 3 settings are documented here because user projects commonly need to
decide whether to customize them. The template ships with safe defaults
(unset / Claude Code defaults). Your team adopts customizations as needed
without inheriting maintainer-specific local doctrine.

---

## Late-Branch Phase D Recovery Procedure

When a long-running branch (e.g., `feat/SPEC-XXX`) needs to integrate
upstream changes from main and the local working tree has uncommitted
modifications or dirty state, use this 5-step recovery sequence to safely
realign without losing work.

### Sequence

```bash
# 1. Preserve orphan commits (if you have local commits not yet pushed
#    that you want to keep around as a safety branch)
git branch save-orphan-$(date +%Y-%m-%d) <latest-local-commit-sha>

# 2. Stash dirty working tree (including untracked files)
git stash push --include-untracked -m "phase-d-$(date -u +%Y%m%dT%H%M%SZ)"

# 3. Fetch latest origin/main
git fetch origin main

# 4. Safe reset (prefer --keep over --hard for working-tree safety)
git reset --keep origin/main

# 5. Restore stashed files (with explicit checkout fallback if pop is partial)
git stash pop || git checkout stash@{0} -- <missing-paths>
```

### Why `--keep` instead of `--hard`

`git reset --keep` is safer than `--hard`: it refuses to reset if working
tree has uncommitted changes (preventing data loss). Combined with the prior
`git stash push`, the working tree is clean at step 4 so `--keep` succeeds
with the same effect as `--hard`.

Claude Code sandboxes also typically block `git reset --hard` as a
destructive command. Using `--keep` avoids this restriction.

### When `git stash pop` produces partial output

`git stash pop` may silently skip some files even when no conflict exists.
The "silent skip" pattern (lines 14-15 of step 5 fallback) addresses this:

```bash
# Diagnose what stash contains vs what was restored
git stash show --stat stash@{0}

# Explicitly checkout any missing paths from the stash
git checkout stash@{0} -- <missing-path-1> <missing-path-2>

# Unstage if needed
git restore --staged <paths>

# Cleanup
git stash drop stash@{0}
```

### When to apply this pattern

- Long-running feature branch (multiple weeks active) that needs to
  integrate upstream main changes
- Working tree has uncommitted modifications you want to preserve
- Local commits exist that aren't pushed yet (orphan branch safety net)

### When to skip

- Clean working tree + simple `git pull --rebase origin main` works
- Short-lived feature branch (`git rebase main` directly is sufficient)
- Strict PR workflow where local branches are rebased before PR creation

---

## Cross-References

- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check —
  canonical Layer 1 rule for race mitigation
- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume
  Pattern — L2/L3 worktree as race-elimination alternative
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory
  Convention — namespace separation contract
- `.claude/rules/moai/development/coding-standards.md` § Language Policy —
  16-language neutrality contract

---

Provenance: This guide externalizes 4 generic patterns from the maintainer-only
project doctrine (Multi-Session Race Mitigation, Hook Setup, Settings Intent,
Late-Branch Recovery). Each pattern is reframed for user-audience neutrality —
maintainer-specific elements (personal paths, 1-person OSS policy specifics,
maintainer-machine assumptions) are generalized as "if you adopt this policy"
or "your project hash" placeholders. The maintainer file (if any) remains the
source of detailed local doctrine; this guide is the template-distributed
generalized derivative.
