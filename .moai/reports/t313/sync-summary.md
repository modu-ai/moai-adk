# SPEC-WORKTREE-BASEREF-001 — Sync-Phase Summary (card t313)

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t313` · branch `WT-worktree-baseref`
Pre-sync HEAD: `023f7fd57` (re-read immediately before staging and again before committing; unchanged)
Agent: `manager-docs`, sync phase. Run-phase and sync-audit were both complete and PASSED before this work began.

---

## Verdict

Sync complete. Documentation, the SPEC lifecycle close, and one sync commit. **No production code touched, no docs repair needed** — the M6 documentation review found the shipped docs accurate and complete.

---

## 1. Changes per file

### `CHANGELOG.md` — line 12 (+ blank separator at line 13)

One new entry at the top of `## [Unreleased]` → `### Added`, matching the file's existing entry style (bold SPEC link, dense single-paragraph prose, run-phase milestone commit chain, audit verdict, sync-commit disposition, `🗿 MoAI` trailer). User-visible substance: the new optional `git_strategy.worktree_base_branch` setting, empty by default (behaviour unchanged when unset), its two consumers, the shared resolvability predicate, and the `moai doctor` / web-console surfaces. No other CHANGELOG line was modified.

### `.moai/specs/SPEC-WORKTREE-BASEREF-001/spec.md` — line 5

`status: in-progress` → `status: completed`. Frontmatter only. `updated:` already read `2026-08-27` (today) and is correct as-is; no other frontmatter field and no body line was touched (`plan.md` and `acceptance.md` were not modified at all).

### `.moai/specs/SPEC-WORKTREE-BASEREF-001/progress.md` — lines 107-150

`## §E.4 Sync-phase Audit-Ready Signal` — the `_<pending sync-phase>_` placeholder replaced with the full YAML audit-ready signal block: sync verdict + report path, the B12 self-tests (a/b/c), the frontmatter transition record, the documentation-review verdict, the six debt findings F1-F6 with dispositions, the five explicitly-unchecked gaps, residual risk, the sync file list, push state, and the t316 boundary confirmation. Sections §E.1 / §E.2 / §E.3 were not touched.

### `.moai/reports/t313/sync-audit.md` — newly tracked

The sync-auditor's report was untracked; it is committed verbatim, unmodified.

### `.moai/reports/t313/sync-summary.md` — this file

Committed in the same sync commit.

**No documentation file was changed.** See §3 below.

---

## 2. Status transition

`in-progress → implemented → completed`, merged into the single sync commit per the 3-phase close convention (`spec-frontmatter-schema.md` § Status Transition Ownership Matrix). Applied to `spec.md` only — it is the only artifact of this SPEC carrying a frontmatter block (`plan.md`, `acceptance.md`, and `progress.md` have none, so no transition is applicable to them).

## 3. Documentation review (task 5) — M6 commit `8c46460ff`

**Verdict: accurate and complete. No repair needed, and none was made.**

The M6 commit added a nine-line `git_strategy.worktree_base_branch` section to `.claude/rules/moai/workflow/worktree-integration.md` (after the existing `baseRef` discussion, which it leaves unchanged) and mirrored it to `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`. Five claims were re-verified against the shipped code in this tree:

| Documented claim | Verified against | Result |
|---|---|---|
| The key lives at the root of `git_strategy` in `.moai/config/sections/git-strategy.yaml` | the local config file + `internal/config/types.go:174` | correct |
| The session-start half fires from the primary checkout only | `internal/hook/worktree_base_branch.go:91-100` — the `WorktreeBaseBranchInPrimaryCheckout()` gate precedes the config read | correct |
| The `git worktree add` half is honoured from any working tree | `internal/cli/session_worktree.go:215-221` — no checkout discriminant on this path | correct |
| An unresolvable value falls back to the no-operand form | `session_worktree.go:216-218` (`base = ""` on a failed predicate) + `gitWorktreeAddArgs`, which appends the operand only when non-empty | correct |
| `moai doctor --check 'Worktree Base Branch'` reports without writing | `internal/cli/doctor_worktree_base.go` — read-only, and the exact diagnostic name matches | correct |
| The shipped template default is empty | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:12` → `worktree_base_branch: ""` | correct |

Mirror parity: `diff -q` between the local rule and the template mirror is clean (byte-identical).

`docs-site` gap check: `docs-site` does not enumerate individual `git_strategy` keys anywhere (a repository-wide grep finds no `worktree_root` or `github_username` page mention either); the web-console page describes the Git & Worktree group at group level. Omitting this key there is therefore consistent with how every sibling key is treated, not a gap. The user-facing field strings do ship in all four locales via `internal/web/assets/i18n.js`.

## 4. B12 CHANGELOG emission self-tests

| Test | Command | Result |
|---|---|---|
| (a) pre-emission duplicate grep | `grep -c 'SPEC-WORKTREE-BASEREF-001' CHANGELOG.md` | `0` before the write — no duplicate entry from a parallel session |
| (b) AC count match | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `16` (non-zero, so not a vacuous comparison); the CHANGELOG entry states 16 criteria, AC-WBR-001..016 |
| (c) file-path verification | `ls -1 <12 cited paths>` | all 12 resolved |

## 5. Sync commit

**SHA: `de2042416`** — recorded below after the commit landed.

`progress.md` §E.4 carries `sync_commit_sha: pending-backfill-sync`. A commit cannot name its own hash, so the placeholder is the established pattern (`spec-frontmatter-schema.md` § SHA placeholder backfill exemption). **Backfill method: a second, follow-up commit** — not an amend, so the sync commit's own SHA stays stable and the audit trail shows both steps.

Staging was by explicit pathspec (`CHANGELOG.md`, `.moai/specs/SPEC-WORKTREE-BASEREF-001/`, `.moai/reports/t313/`); no `git add -A`, no `git add .`, no `git commit -a`. `git status --short` was re-read immediately before staging, and `git rev-parse --short HEAD` + `git branch --show-current` immediately before committing — both matched `023f7fd57` / `WT-worktree-baseref`.

**Not pushed. No PR opened.** Integration is the orchestrator's call.

---

## Gaps

Stated so they are not mistaken for passes.

1. **The sync-audit's five own gaps are carried forward unchanged**, not re-closed by this phase: the AC-WBR-012 mutation is attributed to the implementer rather than independently reproduced; all measurement is darwin/arm64 (no Windows or Linux run — CI covers this on the PR head); concurrency (two simultaneous SessionStarts) was reasoned about but not stress-tested; the web console was never driven through a browser; and inherited G2 (`EnterWorktree`'s actual `refs/remotes/origin/HEAD` read) remains inferred from behaviour. Full statements in `.moai/reports/t313/sync-audit.md` § Coverage statement.
2. **This phase ran no tests and no build.** The sync change set is markdown-only (CHANGELOG + two SPEC artifacts + two reports), so there is nothing for a Go toolchain to verify. The green test/vet/lint evidence in this SPEC belongs to the run phase and the sync audit — this summary does not re-claim it as a fresh measurement.
3. **The six debt findings F1-F6 are recorded, not repaired.** All are optional and non-blocking per the auditor. F1, F3, and F6 are criterion-text repairs in `acceptance.md`, whose body `manager-docs` may not edit — repairing them requires re-delegation to `manager-spec`, which was out of this dispatch's scope. F2 and F4 are guard-strength follow-ups (the auditor names them the highest-leverage pair); F5 is optional `--` hardening on three git invocations.
4. **The sync commit SHA is self-referential.** `progress.md` §E.4 carries the `pending-backfill-sync` placeholder at the moment of the sync commit; the real SHA lands in the follow-up backfill commit recorded in §5 above.
5. **`docs-site` was checked for a coverage gap, not audited.** The check established that no sibling `git_strategy` key is documented per-key there; it did not audit the docs-site pages for any other staleness.
