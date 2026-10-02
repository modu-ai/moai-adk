# SPEC-CC-ULTRACODE-TOGGLE-001 — Acceptance Criteria

Two-cell discipline: every AC pins its RED-now state to tree `c50da9c2f` (measured 2026-10-02 in worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1416`) and states the green-path observable. All commands are single-invocation (no pipes, `&&`, or `;` inside a pinned command) and run from the worktree root unless a `cd` prefix is shown.

Evidence obligation: judge the printed `grep -c` COUNT, not the exit code — `grep -c` exits 1 on a zero count, so a printed `0` is a PASS on zero-match assertions and a printed `0` is the RED value on positive pins. Quote the verbatim printed output for every command. RED-now cells record the observed printed counts; the exit codes follow the `grep -c` contract (count 0 -> exit 1, count >= 1 -> exit 0) and were not captured separately (see Gaps in `progress.md` §E.1).

Path abbreviations:

- `RS` = `.claude/rules/moai/workflow/dynamic-workflows.md`
- `RM` = `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`
- `WF` = `docs-site/content/<loc>/claude-code/agentic/workflows.md`
- `ML` = `docs-site/content/<loc>/multi-llm/_index.md`
- `UW` = `docs-site/content/<loc>/advanced/ultracode-workflows.md`
- `CM` = `docs-site/content/<loc>/claude-code/foundations/commands.md`
- `<loc>` iterates ko, en, ja, zh — every per-locale AC is run once per locale.

## AC-001 — Rule source: the xhigh coupling is gone (REQ-001, REQ-004)

- RED-now (c50da9c2f): `grep -c 'combines `xhigh` reasoning' RS` prints `1`; same on RM prints `1`.
- Given M1 and M3 are applied When ``grep -c 'combines `xhigh` reasoning' <file>`` runs on RS and RM Then each prints `0`. When `grep -c -F 'leaves the effort level unchanged' <file>` runs on RS and RM Then each prints at least `1` (positive pin: a mutant that merely deletes the sentence fails). When `grep -c -F 'v2.1.284' <file>` runs on RS and RM Then each prints at least `1`.
- Positive-pin RED-now: `grep -c -F 'v2.1.284' RS` prints `0`.

## AC-002 — Rule source: off route and launch-flag exception (REQ-002, REQ-003)

- RED-now (c50da9c2f): ``grep -c 'step back with `/effort high`' RS`` prints `1`; `grep -c -F 'effort ultracode off' RS` prints `0`; `grep -c -F -- '--effort ultracode' RS` prints `0`; `grep -c -F '"ultracode": true' RS` prints `0`.
- Given M1 and M3 are applied When ``grep -c 'step back with `/effort high`' <file>`` runs on RS and RM Then each prints `0`. When the other three positive pins run on RS and RM Then each prints at least `1` (the off route, the launch-flag exception that legitimately keeps one `xhigh` mention, and the persistent settings-key route).
- Retention pin (REQ-003): `grep -c -F 'ultrathink.' RS` prints at least `1` after the edit — the "opener does not restore ultracode" sentence is retained. (Control: green today, MUST NOT flip.)

## AC-003 — Mirror parity (REQ-004) — control, MUST NOT flip

- Pre-state (green today at c50da9c2f): `cmp RS RM` exits 0 with no output.
- Given M1 and M3 are applied When `cmp RS RM` runs Then exit code is 0 with no output. A mutant that edits only one side flips this red.

## AC-004 — docs-site workflows row, all four locales (REQ-005, REQ-009)

- RED-now (c50da9c2f), per locale `<loc>`: `grep -c -E '/effort ultracode.*xhigh' WF` prints `1` for ko, en, ja, zh; `grep -c '/effort ultracode.*/effort high' WF` prints `1` for ko, en, ja, zh; `grep -c 'v2.1.284' WF` prints `0` for ko, en, ja, zh.
- Given M2 is applied When `grep -c -E '/effort ultracode.*xhigh' WF` runs per locale Then each prints `0`. When `grep -c '/effort ultracode.*/effort high' WF` runs per locale Then each prints `0`. When `grep -c 'v2.1.284' WF` and `grep -c -F '/effort ultracode off' WF` run per locale Then each prints at least `1` (positive pins; ko-only or two-locale edits fail the per-locale loop).
- Shape control: `grep -c -E '^\| `/effort ultracode` \|' WF` prints `1` per locale before and after (the row stays a single table row).

## AC-005 — docs-site multi-llm comment, all four locales (REQ-006)

- RED-now (c50da9c2f), per locale: `grep -c -E '^/effort ultracode # xhigh' ML` prints `1` for ko, en, ja, zh.
- Given M2 is applied When `grep -c -E '^/effort ultracode # .*xhigh' ML` runs per locale Then each prints `0`. When `grep -c -E '^/effort ultracode #' ML` runs per locale Then each prints `1` (the line still exists as a one-line comment; a mutant that deletes the line fails).

## AC-006 — docs-site ultracode-workflows page, all four locales (REQ-007)

- RED-now (c50da9c2f): en ``grep -c -F 'Reasoning effort: set to `xhigh`' UW`` prints `1`; ja ``grep -c -F 'Reasoning effort: `xhigh` に設定' UW`` prints `1`; zh ``grep -c -F 'Reasoning effort：设置为 `xhigh`' UW`` prints `1`; ko ``grep -c -F '`xhigh`로 올라갑니다' UW`` prints `1`, ``grep -c -F '/effort high`로 한 단계 내립니다' UW`` prints `1`, and `grep -c -F '세 가지가 함께 바뀝니다' UW` prints `1`.
- Given M2 is applied When each of those commands runs on its locale Then each prints `0`. When `grep -c -F '/effort ultracode off' UW` runs per locale Then each prints at least `1`. When `grep -c -F '세션 경계를 넘지 않습니다' ko-UW` runs Then it prints `1` (the session-boundary callout is retained — control).

## AC-007 — docs-site commands page, all four locales (REQ-008)

- RED-now (c50da9c2f): en ``grep -c -F 'simultaneously an `/effort` level' CM`` prints `1`; ko ``grep -c -F '동시에 `/effort` 레벨입니다' CM`` prints `1`; ja ``grep -c -F '`/effort` のレベルでもあります' CM`` prints `1`; zh ``grep -c -F '也是一个 `/effort` 等级' CM`` prints `1`.
- Given M2 is applied When each command runs on its locale Then each prints `0`. When `grep -c 'ultracode' CM` runs per locale Then each prints at least `1` (the mention is corrected, not deleted).

## AC-008 — 4-locale heading parity and no new divergence (REQ-009, REQ-010) — control

- Baseline at c50da9c2f (heading counts, `grep -c '^#' <file>`): WF 11/11/11/11 (ko/en/ja/zh); ML 13/13/13/13; UW 24/20/20/20 (already divergent, listed in `docs-site/.locale-parity-baseline`); CM ko 23, en 23, ja 18, zh 18.
- Given M2 is applied When `grep -c '^#' <file>` runs on each of the 16 docs files Then each count equals its baseline. When `git diff --stat -- docs-site/.locale-parity-baseline` runs Then the output is empty.

## AC-009 — docs-site exit gate (REQ-010) — control

- Given M2 is applied When `hugo --minify --gc` runs in `docs-site/` Then exit code is 0 and the output contains no `WARN` or `ERROR` line. When the URL-blacklist grep (`grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content`) runs Then it prints nothing. When `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content` runs Then it prints nothing. (Recipes: `hns-oss-docs-verify` §1-§3; the build is structurally green-today — a control, not a flip.)

## AC-010 — Scope control (REQ-011, REQ-012)

- Given only the change-map files are edited When `git status --porcelain` runs Then there are no entries outside the 2 rule files, the 16 docs-site files, and `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/` (the assertion is the absence of extras). When `grep -c -i 'slider' <RS>` runs Then the edited bullet makes no persistence claim about the slider toggle (reviewer reads the bullet; OQ-1 stays open).
- Out-of-scope surfaces untouched: `git diff --stat -- .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/session-handoff-examples.md .claude/rules/moai/workflow/session-handoff-format.md .claude/output-styles/moai/moai.md CHANGELOG.md` prints nothing.

## AC-011 — Template embed sanity

- Given the mirror is edited When `make build` runs from the worktree root Then exit code is 0. When `go test ./internal/template/...` runs from the worktree root Then exit code is 0. (`make build` does not regenerate the mirror — see plan §B — so this proves the embed compiles and the template package tests still pass; byte parity is AC-003.)

## Mutant probe (adoption check)

- Delete-only mutant (remove the wrong sentences, add nothing): fails AC-001/002/004/006 positive pins.
- One-side mutant (edit RS only, or ko only): fails AC-003 / the per-locale loops.
- Over-correction mutant ("never xhigh" including the launch flag): fails AC-002's `--effort ultracode` pin.
- Rewrite-in-place mutant that keeps the `/effort high` return phrasing under new words: fails AC-004's `/effort ultracode.*/effort high` zero pin.

## Quality gates

- LSP gates: N/A — prose-only change to `.md` files, no Go symbols.
- Definition of Done: AC-001..AC-011 green with verbatim output cited; open questions OQ-1..OQ-3 still recorded as open in spec.md.
