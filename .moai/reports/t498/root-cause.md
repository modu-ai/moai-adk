# t498 — root cause investigation (plan-phase evidence)

Tree: `.claude/worktrees/t498`, branch `WT-codex-mirror-doctor`, base `ace1c5440` (= origin/develop tip, `HEAD..origin/develop` = 0).
Probe binary: `~/go/bin/moai` v3.2.0-rc.0, commit `e79c010b8`, mtime `Sep 3 09:16`.
Probe project: `<scratchpad>/t498probe` (throwaway, outside every repo tree).

## Claim 1 — the mirror feature is live and unconditional

`moai init t498probe --non-interactive` → rc=0, `.agents/skills` holds **24** entries, `.claude/skills` holds **24**; every entry a relative symlink (`moai -> ../../.claude/skills/moai`).
There is no non-test setter of `WithSkillMirror` / `skillMirrorDisabled` (`grep -rn --include='*.go'` → only the definition at `skill_mirror.go:114-118` and the field at `deployer.go:90-93,290`), so mirroring is not gated by `--agent`, config, or platform.

## Claim 2 — `moai update` does not repair a missing mirror (NEW, reproducible)

In the same probe project:

| step | command | `.agents/skills` count |
|---|---|---|
| baseline | (after init) | 24 |
| delete | `rm -rf .agents` | 0 (`ls: .agents: No such file or directory`) |
| routine update | `moai update --yes` → rc=0 | **0** |
| forced sync | `moai update --templates-only --force --yes` → rc=0 | **37** |

The routine run printed exactly `✓ Up to date · Skipping sync` (log 4 lines, no `Deploying templates` line). The mechanism is the version-match short-circuit: `runTemplateSyncWithProgress` returns `syncSkipped`, and `internal/cli/update.go:509-523` returns early — **before** the "Deploy Templates" step that owns `deployWithMirrorNotice` (`update_template_sync.go:341`). Mirror creation lives inside `DeployWithResult` (`deployer.go:290-292`), so no deploy ⇒ no mirror, and no notice either.

Consequence: on a version-matched project a missing, stale, or broken mirror is invisible and self-perpetuating. This is the standing gap the doctor check is for.

## Claim 3 — why THIS repository has no mirror

Measured absent in three trees: primary checkout `/Users/goos/MoAI/moai-adk-go/.agents`, `.claude/worktrees/develop/.agents`, and this worktree — all `No such file or directory`.

The cause is **not** a deliberate exclusion of the feature. No rule, doctrine, or doc excludes it: `grep -rn '\.agents/' --include='*.md' .claude/rules .moai/docs CLAUDE.md CLAUDE.local.md AGENTS.md` → no matches. Nothing in code gates it (Claim 1). The mirror is created only by a deploy, and no mirror-bearing deploy has left a result in this tree.

A related divergence WAS found, and it is a candidate defect rather than the cause:

- this repo `.gitignore:133` — blanket `.agents/`, introduced by `16b5a6c9a` ("chore(gitignore): migration 백업 + Codex CLI 아티팩트 무시"), which predates the mirror feature (`9c94c6b7a`, 2026-08-22, in `origin/main`).
- the shipped template `.gitignore` — a documented policy block ("Skill Mirror (build product, not source)") ignoring only `.agents/skills/moai*` and stating "the `.agents/` root itself is NOT [ignored]: entries you create there, and source files placed there later, stay tracked".

So the repository ignores more than the policy it ships intends to.

## Gaps (explicitly NOT observed)

- Whether a mirror-carrying binary ever ran a deploy in this repository is **indeterminate**. `.claude/settings.json` mtime `Aug 27 14:25` shows a deploy ran around then, and the feature landed Aug 22, but the binary in use on that date is unrecoverable — "never created" and "created then removed" cannot be separated from the evidence available. Nothing in `CleanMoaiManagedPaths` touches `.agents`.
- Windows copy-fallback behaviour was not exercised; the probe ran on darwin, where symlink creation succeeds.
- `.moai/manifest.json` in the primary checkout reads `version: v2.0.0-dirty`, `deployed_at: 2026-02-07` — read but not used as a basis for any claim above.

## Claim 4 — doctor carries no mirror check (absence claim re-verified positively)

`grep -rn '\.agents' --include='*.go' internal/cli/` returns **5 hits, all in `_test.go` files** — zero in any non-test file across the entire doctor surface. The full `DiagnosticCheck` name inventory (`grep -rhn 'DiagnosticCheck{Name:'`) carries no mirror row.

`doctor_codex.go`'s skill logic is a different subject: `classifyCodexSkillPath` / `codexStaleSkillFinding` read the **user-layer** `~/.codex/config.toml` `[[skills.config]]` registrations (via `resolveCodexHomeDir`), not the project-local `.agents/skills` mirror.

The mirror's own observable state (`MirrorModeFailed` / `Skipped` / `Copy`) is consumed only at deploy time — `internal/mirrornotice` is imported by exactly two non-test callers, `internal/core/project/initializer.go:427` (init warnings) and `internal/cli/mirror_notice.go:43` (update stderr). A user who never re-deploys never sees it again. Confirmed.

## Residual risk

- The natural home for a new check is `checkCodexWiring` (`doctor_codex.go:86`), which is advisory, fail-open, and READ-only, and carries an explicit un-nagging invariant: a claude-only project on a codex-less machine stays silent. A mirror check added there must preserve that invariant, or every claude-only user gains a permanent new row.
- `.agents/` is gitignored in this repo, so a mirror created here would be untracked — the fix for the absence is a deploy, not a commit, and it changes nothing that CI can see.
