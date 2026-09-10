# t303 — SPEC-SYNC-STRATEGY-KEY-001 run-phase verdict

- **Card**: t303
- **SPEC**: SPEC-SYNC-STRATEGY-KEY-001 (v0.2.0, Tier M, cycle_type=ddd)
- **Worktree**: `.claude/worktrees/t303`, branch `WT-sync-strategy-key`
- **Tree measured**: `63b4628a6` (all AC rows below), unless a row names another SHA
- **Base of diff-scoped probes**: `0931789b6` (the branch state at run-phase entry)
- **Raw command + output battery**: `.moai/state/verify/t303/ac-battery.txt`
- **Verdict**: **PASS** — 12/12 AC, 9/9 MUST-PASS

Shorthand: `T` = `internal/template/templates`, `D` = `T/.claude/skills/moai/workflows/sync/delivery.md`, `L` = repo root.

---

## Claim

The run phase landed all five milestones of SPEC-SYNC-STRATEGY-KEY-001. The sync delivery workflow now reads the delivery strategy exclusively from `git_strategy.{mode}.workflow`, on the canonical value domain `{github-flow, git-flow}`, with an explicit stop on an unmatched value, a `WT-*` integration-worktree route, an explicit stop on an unmatched branch in every strategy block, and a dated legacy fallback for stale configs. The retired `github.spec_git_workflow` key is gone from the shipped template and the local tree except for that one fallback mention, and the shipped-key inventory is back in parity. All 12 acceptance criteria pass; the 9 MUST-PASS criteria pass.

## Evidence

Baseline captured at run-phase entry on `0931789b6` — every RED-now figure in `acceptance.md` was re-measured and matched before any edit:

```
$ grep -rn 'spec_git_workflow' internal/template/templates/ | wc -l   →       10
$ grep -rn 'spec_git_workflow' .claude/skills .moai/config/sections/system.yaml | wc -l →       10
$ grep -c 'WT-' <D>                                                   →        0
$ grep -n 'matches no\|no defined route\|unmatched' <D> | wc -l       →        0
$ grep -n 'Default strategy\|Otherwise,' <D>
25:Default strategy (if not configured): `github_flow`
$ grep -n 'github.spec_git_workflow' internal/config/testdata/shipped_key_inventory.yaml
2505:- path: "github.spec_git_workflow"
$ go test ./internal/config/ -run TestShippedConfigKeysHaveReaders
ok  	github.com/modu-ai/moai-adk/internal/config	(cached)
```

### AC matrix (measured on `63b4628a6`)

| AC | Sev | Command | Verbatim output | Verdict |
|----|-----|---------|-----------------|---------|
| AC-SYK-001 | MUST | `grep -c 'git_strategy.{mode}.workflow' <D>` · `grep -c 'Read \`github.spec_git_workflow\`' <D>` | `6` · `0` | PASS |
| AC-SYK-002 | MUST | `grep -rn 'spec_git_workflow' <T>/ \| wc -l` · refined `\| grep -v 'Legacy key fallback' \| wc -l` | `1` · `0` | PASS (refined) |
| AC-SYK-003 | SHOULD | `grep -c 'v3.3.0' <D>` · `grep -n 'main_direct\|feature_branch' <D>` | `1` · `39:` and `40:` only — both rows of the fallback mapping table | PASS |
| AC-SYK-004 | MUST | `grep -n '{github-flow, git-flow}' <D>` | `29:**Unmatched value stops delivery.** … report the offending value together with the canonical domain \`{github-flow, git-flow}\`. Do not create a pull request and do not push. A missing \`workflow\` subkey … is an unmatched value, not a default — there is no fallback strategy.` | PASS |
| AC-SYK-005 | MUST | `grep -c 'WT-' <D>` · `awk '/^\*\*WT-\* branch\*\*/,/^\*\*feature\/\* branch\*\*/' <D> \| grep -c 'gh pr'` | `6` · `0` | PASS |
| AC-SYK-006 | MUST | `awk '/^##### Strategy:/{s=$0} /matches no defined route/{c[s]++} END{for(k in c) print k, c[k]}' <D>` · negative probe `grep -c 'Default strategy\|Otherwise,' <D>` | `##### Strategy: github-flow 1` + `##### Strategy: git-flow 1` (exactly 2 lines) · `0` | PASS |
| AC-SYK-007 | SHOULD | `grep -c 'delivery.md' L/.claude/rules/local/gitflow-lane-protocol.md` · `grep -c 'merge --no-ff'` on that file · same on `<D>` | `2` · `0` · `1` | PASS |
| AC-SYK-008 | MUST | `grep -rn 'gitflow-lane-protocol' <T>/ \| wc -l` · `grep -n 'workflow:' <T>/.moai/config/sections/git-strategy.yaml.tmpl` · `git diff 0931789b6 -- <T>/ \| grep '^+' \| grep -oE 'SPEC-[A-Z0-9-]+-[0-9]{3}' \| wc -l` | `0` · `13:    workflow: github-flow` / `45: …` / `81: …` (no private value) · `0` | PASS |
| AC-SYK-009 | MUST | `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1` · `grep -c 'github.spec_git_workflow' internal/config/testdata/shipped_key_inventory.yaml` | `ok  github.com/modu-ai/moai-adk/internal/config  0.870s` (exit 0) · `0` | PASS |
| AC-SYK-010 | MUST | `grep -c 'github.spec_git_workflow'` · `grep -c 'automation.auto_branch'` on `<T>/.claude/skills/moai-workflow-project/schemas/tab_schema.json` | `0` · `3`; `json.load` re-parse succeeded after the edit | PASS |
| AC-SYK-011 | SHOULD | `sed -n '25p'` on template and local `doc-execution.md` | both print `- Read project configuration: git_strategy.mode, git_strategy.{mode}.workflow, conversation_language` | PASS |
| AC-SYK-012 | MUST | (1) `grep -rn 'spec_git_workflow' L/.claude/skills/ L/.moai/config/sections/system.yaml \| grep -v 'Legacy key fallback' \| wc -l` (2) `diff` tab_schema T↔L (3) `grep -c 'SPEC-SYNC-PARALLEL-DOCS-001' L/…/doc-execution.md` (4) `grep -n 'workflow:' L/.moai/config/sections/git-strategy.yaml` | (1) `0` (2) empty (3) `3` (4) `8:        workflow: git-flow` | PASS |

### Invariants (preserve criteria)

| Invariant | Command | Verbatim output | Verdict |
|---|---|---|---|
| Step 3.1 Tier Route A/B gate untouched | `git diff 0931789b6 -- <D> \| grep -cE '^[+-].*(Route A\|Route B\|Tier-based Route gate)'` | `0` — and the gate text reads verbatim at `<D>:47-51` | PASS |
| Template neutrality guard | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok  github.com/modu-ai/moai-adk/internal/template  0.524s` (exit 0) | PASS |
| Adjacent template guards | `go test ./internal/template/ -run 'TestInternalContentLeak\|TestRuleTemplateMirror\|TestCommandsAudit' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/template  0.249s` (exit 0) | PASS |
| Embedded templates regenerated | `make build`; then `git status --short` | exit 0; empty status | PASS |
| Local-only mirror content preserved | `diff` T↔L on delivery.md and doc-execution.md | delivery.md: only the 2-line footer; doc-execution.md: only the `SPEC-SYNC-PARALLEL-DOCS-001 A5` block | PASS |

### Commits (all on `WT-sync-strategy-key`, none pushed)

| Milestone | SHA | Subject |
|---|---|---|
| M1 | `c374d9605` | `feat(sync): read git_strategy.{mode}.workflow, clean the value axis, route WT-* lanes (t303)` |
| M2 | `eb31593d1` | `feat(project): point auxiliary consumers at the canonical git-strategy keys (t303)` |
| M3 | `1c9295a07` | `feat(config): retire github.spec_git_workflow from the shipped template (t303)` |
| M4 | `63b4628a6` | `chore(sync): sync local mirrors and repo-local surfaces to the canonical key (t303)` |
| M5 | (this commit) | `docs(SPEC-SYNC-STRATEGY-KEY-001): run-phase evidence + status in-progress (t303)` |

## Baseline-attribution

Every figure above was measured in this run, against this tree, at the SHA named in the row. The RED-now baselines quoted under Evidence were re-measured at `0931789b6` at run-phase entry — they were not carried over from `acceptance.md`'s `d29b8942e` measurement, and they matched it in all six probes. The diff-scoped probes (AC-SYK-008) are pinned to the fixed base SHA `0931789b6`, never to a branch name.

`make build` was run and produced no tracked-file change. That is not an unverified assumption: `internal/template/scripts/gen-catalog-hashes.go:22` states "For skill entries: hash only the root SKILL.md or skill.md (not sub-files)", and lines 127-134 implement it, so the three sub-file edits (delivery.md, doc-execution.md, tab_schema.json) cannot move a catalog hash. No `catalog.yaml` cascade is due.

## Gaps — explicitly NOT observed

1. **Full test suite not run.** Only `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` and three targeted `internal/template` guards were executed, per the machine-load discipline. A cross-package regression outside those two packages would not have been seen here. The change touches no Go production code, so the exposure is limited to testdata parity — which is the one thing that was tested.
2. **Cross-platform build not run.** No `GOOS=windows go build`; the change adds no Go code, so there is nothing platform-conditional to compile.
3. **Runtime behavior of the rewritten skill text not exercised.** `delivery.md` is instruction text read by an agent at sync time, not executable code. No `/moai sync` run was performed against a `git-flow` config or a `WT-*` branch, so the routes are verified as *text* (greps, per-block awk), not as *observed behavior*. The first real exercise will be the next sync on this branch.
4. **`golangci-lint` not run.** No Go source changed; the only Go-adjacent edit is a testdata YAML row removal.
5. **Template-neutrality CI guard not observed on CI.** It was run locally (exit 0). Its trigger is `pull_request: branches: [main]`, and under this repo's git-flow model card branches carry no PR — see Residual-risk 4.
6. **Nothing pushed.** No `git push`, no PR, no integration into `develop`. `l44_pre_commit_fetch` / `l44_post_push_fetch` are both `not-run` because no push occurred.

## Residual-risk

1. **B5 — CLAUDE.local.md recipe now points at an out-of-domain value (operator-surfaced, NOT edited).** `CLAUDE.local.md` §2.3 lines 180-181 and 186 instruct the operator to re-apply `workflow: gitflow` after `moai update`, and line 181's check (`grep -n 'workflow: gitflow' … || echo 'REVERTED — 재적용 필요'`) will now print `REVERTED` against the correct `git-flow` value — a false alarm that invites re-introducing the exact token the new Step 3.0 stops on. `CLAUDE.local.md` is user-owned; per the run-phase ownership boundary it was left untouched. **Recommended operator edit: `gitflow` → `git-flow` at lines 181 and 186.**
2. **The local `git-flow` value is non-durable.** `moai update` wipes `.moai/config` wholesale, so `git_strategy.manual.workflow: git-flow` reverts to the template default `github-flow` on every update. This is pre-existing (plan.md B1), not introduced here; the new explicit stop converts what used to be a silent wrong-strategy into a visible stop, which is an improvement but still a manual re-apply.
3. **OBS-1 — adjacent schema paths are wrong, and were left wrong.** `tab_schema.json` L517 / L727 bind `git_strategy.personal.auto_branch` and `git_strategy.team.auto_branch`, both missing the `.automation` level that `internal/config/types.go:58,158` and `git-strategy.yaml.tmpl:26-27` actually use. The question this SPEC rebound uses the correct path (`git_strategy.{mode}.automation.auto_branch`), so the schema now carries two spellings of the same key. Pre-existing defect, outside this SPEC's scope envelope — not repaired, reported.
4. **OBS-2 — the rebound question overlaps its per-mode siblings.** The rebound question sits in batch 3.10, whose condition is `git_strategy.mode != manual`, i.e. personal or team — and both of those modes already ask `auto_branch` in their own batches (L517, L727). REQ-SYK-010 prescribed this rebind explicitly and it was executed as written; the overlap is reported rather than resolved unilaterally. A follow-up may prefer to drop the question entirely and fix the two sibling paths instead.
5. **OBS-3 — the neutrality CI guard may never fire for this card.** `.github/workflows/template-neutrality-check.yaml` triggers on `pull_request: branches: [main]`. Under the repo's git-flow model (CLAUDE.local.md §4.1) card branches carry no PR and integrate into `develop` directly, so this guard's protection for this change rests on the local run (exit 0) plus whatever fires on the eventual release PR.
6. **The `{mode}` placeholder in `tab_schema.json` is interpreted, not resolved.** No Go code reads `tab_schema.json` (`grep -rn 'tab_schema' --include='*.go'` returns only a leak-test path reference); the interviewing agent reads it. `field: "git_strategy.{mode}.automation.auto_branch"` therefore relies on the agent substituting the active mode — the same convention `delivery.md` uses in prose. If a future consumer parses this file mechanically, the placeholder will need resolving.
