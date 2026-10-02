# SPEC-CC-ULTRACODE-TOGGLE-001 — Acceptance Criteria

Two-cell discipline. Every release-blocking AC carries a RED-now cell (command, verbatim stdout, exit code, tree SHA — all in the Evidence ledger below) and a green-path cell (the printed values after the work). Controls are classed regression-guard: green today, must not flip.

Measured on this worktree at HEAD `ff7b64f6e` (ledger regenerated for iteration 3; the earlier measurement tree was `0e7b6af5b`); the scope files are byte-identical to the original pin `c50da9c2f` (`git diff --stat c50da9c2f HEAD -- .claude internal docs-site` printed nothing at `ff7b64f6e`). Judge printed `grep -c` COUNTS; `grep -c` exits 1 when every file prints `0`, and exits 0 when at least one file prints non-zero — for multi-file invocations read each `path:count` line, not the exit code. The `grep` on this machine is `ugrep`, whose multi-file `-c` print order is not guaranteed to be ko, en, ja, zh; counts per path are stable, so match each count to its own path.

## Path ledger (literal paths; `<loc>` is each of ko, en, ja, zh)

| Id | Literal path |
|----|--------------|
| RS | `.claude/rules/moai/workflow/dynamic-workflows.md` |
| RM | `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` |
| WF | `docs-site/content/<loc>/claude-code/agentic/workflows.md` |
| ML | `docs-site/content/<loc>/multi-llm/_index.md` |
| UW | `docs-site/content/<loc>/advanced/ultracode-workflows.md` |
| CM | `docs-site/content/<loc>/claude-code/foundations/commands.md` |
| ko-UW | `docs-site/content/ko/advanced/ultracode-workflows.md` |

Every ledger command below is written with the literal paths (multi-file invocations list all four locale paths in the order ko, en, ja, zh). A per-locale AC is judged on each locale's own `path:count` line.

## AC-001 — Rule source: coupling removed, canonical phrases present (REQ-001, REQ-004) — release-blocking

- RED-now → green-path: `[R1b]` (old phrase) `1,1` → `0,0`; `[R2]` `leaves the effort level unchanged` `0,0` → at least `1,1`; `[R3]` `independent on/off toggle` `0,0` → at least `1,1`; `[R4]` `v2.1.284` `0,0` → at least `1,1`.
- Structural coupling guard (replaces the retired verb-list pin R1, which was a blacklist: it let "pairs with / lifts reasoning to / defaults to / best used at `xhigh`" through and false-fired on correct sentences such as the one containing "runs at" or "sets"): `[R1s]` (`xhigh` twice on one line) `0,0` must stay `0,0`, and `[R1c]` (lines containing `xhigh`) `1,1` must stay `1,1`. With `[R7]` (the launch-flag clause, at least `1,1`, in AC-002) these force exactly one `xhigh` in each rule file, inside the launch-flag clause, in any wording — a second `xhigh` on the bullet line trips `[R1s]`, a coupling sentence on a separate line trips `[R1c]`. `[R1s]` and `[R1c]` are regression-guards inside this release-blocking AC (green today by design); the RED flip is carried by `[R1b]` and the positive pins. The positive pins use the exact phrases REQ-001 mandates and plan.md M1 lists; they are byte-identical to the REQ text. Limits: this guard counts `xhigh` and does not read meaning, so it does not catch a coupling claim worded without the token `xhigh`; that class is left to review of the diff.

## AC-002 — Rule source: off route, launch-flag exception, scope, settings key (REQ-002, REQ-003, DEC-1) — release-blocking

- RED-now → green-path: `[R12]` `step back with `/effort high`` `1,1` → `0,0`; `[R5]` `/effort ultracode off` `0,0` → at least `1,1`; `[R6]` `--effort ultracode` `0,0` → at least `1,1`; `[R7]` `starts the session at `xhigh`` `0,0` → at least `1,1`; `[R10]` `current session` `0,0` → at least `1,1`; `[R8]` `"ultracode": true` `0,0` → at least `1,1`; `[R11]` `v2.1.284` and `"ultracode": true` on one line `0,0` → at least `1,1` (the settings-key route carries its version qualifier, DEC-1).
- Retention control: `[R9]` `ultrathink.` `1,1` must stay at least `1,1` (the "opener does not restore ultracode" sentence).

## AC-003 — Mirror parity (REQ-004) — regression-guard

- `[X-cmp]` `cmp RS RM` exit `0`, empty stdout today; must stay exit `0` after the edit. A mutant that edits only one side flips it.

## AC-004 — docs-site workflows row, four locales (REQ-005, REQ-009, REQ-011) — release-blocking

- RED-now → green-path, per locale ko/en/ja/zh: `[W1]` `--effort ultracode` followed by `xhigh` within one clause (`--effort ultracode[^|.。;；,，、]*xhigh`: no table-cell break, period, semicolon, or comma between them, which REQ-005 requires of the flag clause) `0,0,0,0` → at least `1` each; `[W2]` `"ultracode": true` `0,0,0,0` → at least `1` each; `[W3]` `v2.1.284` and `"ultracode": true` on one line `0,0,0,0` → at least `1` each; `[W4]` `/effort ultracode off` `0,0,0,0` → at least `1` each; `[W8]` `v2.1.284` `0,0,0,0` → at least `1` each; `[W6]` `/effort ultracode` … `/effort high` `1,1,1,1` → `0,0,0,0`.
- Row-xhigh discipline (replaces the retired pin `[W9]` `/effort ultracode.*xhigh`, which the required flag clause would legitimately match): `[W5]` (`xhigh` twice, or `xhigh` before `--effort ultracode`) `0,0,0,0` must stay `0,0,0,0` — together with `[W1]` this permits exactly one `xhigh` on the row, located in the flag clause. A mutant that keeps the old "Combines `xhigh` …" opening and appends the flag clause trips `[W5]`; a mutant that drops the flag clause trips `[W1]`; a mutant that moves the `xhigh` claim to another sentence, or behind a semicolon, a period, or a comma, trips `[W1]`. Limit: a relocation that stays in one comma-free clause ("… launch flag also starts the session and ultracode always runs at `xhigh`") still satisfies `[W1]`; that class is left to review of the diff (probe row G3 below).
- Controls (regression-guard): `[W7]` row shape `1,1,1,1` stays; `[W10-ko]`, `[W10-en]`, `[W10-ja]`, `[W10-zh]` the current-session scope literal (`현재 세션` / `current session` / `現在のセッション` / `当前会话`) stays `1` each (the existing "current session only" statement is retained, not rewritten away).

## AC-005 — docs-site multi-llm comment, four locales (REQ-006) — release-blocking

- RED-now → green-path: `[M1]` `^/effort ultracode # .*xhigh` `1,1,1,1` → `0,0,0,0`. Control: `[M2]` `^/effort ultracode #` `1,1,1,1` stays (a mutant that deletes the line fails).

## AC-006 — docs-site ultracode-workflows page, four locales (REQ-007) — release-blocking

- RED-now → green-path: `[U1]` any list bullet carrying `xhigh` (`^- .*xhigh`, label-independent; today `1,1,1,1` and the single match per page is the effects bullet — ko L67, en L112, ja/zh L109) `1,1,1,1` → `0,0,0,0`; `[U1c]` lines containing `xhigh` per page (the effects-bullet mention removed, nothing added) `5,1,1,1` → `4,0,0,0` (ko keeps four unrelated lines — L144 and the L151-153 table rows — so its expected value is 4, not 0; this pin also catches the claim reworded as prose or relabelled); `[U2]` `/effort high` `1,0,0,0` → `0,0,0,0`; `[U3]` `/effort ultracode off` `0,0,0,0` → at least `1` each; `[U4]` ko "three things change together" sentence `1` → `0`.
- Control: `[U5]` ko session-boundary callout `1` stays `1`. Heading counts are held by AC-008.

## AC-007 — docs-site commands page, four locales (REQ-008) — release-blocking

- RED-now → green-path: `[C1en]`, `[C1ko]`, `[C1ja]`, `[C1zh]` the "is an `/effort` level" sentence `1` each → `0` each; `[C2en]`, `[C2ko]` the L137 level-list wording `1` each → `0` each; `[C3-en]`, `[C3-ko]`, `[C3-ja]`, `[C3-zh]` a line carrying `ultracode` and the locale's toggle word (`toggle` / `토글` / `トグル` / `开关`) `0` each → at least `1` each. (The earlier `grep -c ultracode` pin is retired: it was already `3,3,2,2` today.)

## AC-008 — 4-locale heading parity and no new divergence (REQ-009, REQ-010) — regression-guard

- `[H-WF]` 11/11/11/11, `[H-ML]` 13/13/13/13, `[H-UW]` 24/20/20/20 (already divergent), `[H-CM]` 23/23/18/18 (already divergent) must stay identical after the edit. `[X-baseline-diff]` (`git diff --stat -- docs-site/.locale-parity-baseline`) empty today, must stay empty.

## AC-009 — docs-site exit gate (REQ-010) — regression-guard

- Observed on the unedited tree at `0e7b6af5b` (iteration 1 run; the scope files are identical at `ff7b64f6e`, and the plan-audit iteration 2 reproduced exit 0 / 0 `WARN|ERROR` on its own tree): `hugo --minify --gc --destination <scratch>` run in `docs-site/` → exit `0`, `grep -c -E 'WARN|ERROR'` over its output printed `0`. URL-blacklist grep (`grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content`) printed nothing, exit `1`. Mermaid grep (`grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content`) printed nothing, exit `1`. After the work all three must repeat these values.

## AC-010 — Scope and slider control (REQ-011, REQ-012) — regression-guard

- `[R13]` `slider` over RS and RM `0,0`, and `[W11]` slider words (`slider|슬라이더|スライダー|滑块`) over all 16 edited docs-site files `0` each, must stay `0` after the edit (the edited text does not mention the slider, so it cannot assert slider persistence).
- Given only the change-map files are edited When `git status --porcelain` runs Then no entry lies outside the 2 rule files, the 16 docs-site files, and `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/` (the assertion is the absence of extras). When `git diff --stat -- .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/session-handoff-examples.md .claude/rules/moai/workflow/session-handoff-format.md .claude/output-styles/moai/moai.md CHANGELOG.md .moai/docs/session-handoff-appendix.md` runs Then the output is empty. Line-level scope inside the 18 edited files is judged by reading the diff against the §3 change map; no command asserts it.

## AC-011 — Template embed sanity (REQ-004) — regression-guard

- Given the mirror is edited When `make build` runs from the worktree root Then the exit code is 0. When `go test ./internal/template/...` runs from the worktree root Then the exit code is 0. (`make build` does not regenerate the mirror — plan §B; byte parity is AC-003.)

## Mutant probe (adoption check)

Named mutants and the pin each one trips (by design): delete-only (remove the wrong sentences, add nothing) fails the positive pins `[R2]`..`[R8]`, `[W1]`..`[W4]`, `[U3]`, `[C3-*]`; one-side (edit RS only, or ko only) fails `[X-cmp]` / the per-locale pins; over-correction ("never xhigh" including the launch flag) fails `[R7]` and `[W1]`; unqualified settings key fails `[R11]`, `[W3]`; slider claim fails `[R13]`, `[W11]`.

Rewording probe, run in scratch copies under the session scratchpad (the worktree was not touched): the real line-111 bullet with the old coupling and return sentences replaced by a correct draft (positive phrases R2-R11 present, one `xhigh` in the launch-flag clause), plus one extra coupling sentence per mutant. Result of the NEW pin set (`[R1b]`, `[R12]`, `[R1s]`, `[R1c]`, `[R2]`..`[R8]`, `[R10]`, `[R11]`, `[R9]`):

| Rule-source mutant | New pin set | Tripped by | Old verb-list R1 (retired) |
|---|---|---|---|
| D correct draft (control) | PASS | — | 1 (false positive: the verb "runs at" in the draft's own sentence "at whichever effort level the session runs at") |
| A "pairs with `xhigh` reasoning ..." | FAIL | `[R1s]` | 1 (fired on the draft's "runs at", not on A's wording) |
| B "lifts reasoning effort to `xhigh` ..." | FAIL | `[R1s]` | 1 (same cause) |
| C "defaults to `xhigh` effort ..." | FAIL | `[R1s]` | 1 (same cause) |
| E "best used at `xhigh`; stays on at that level" | FAIL | `[R1s]` | 1 (same cause) |
| F "forces `xhigh`" | FAIL | `[R1s]` | 1 |
| J coupling on its own new line | FAIL | `[R1c]` (two `xhigh` lines) | 1 (same cause) |
| FP correct line containing "sets" | PASS | — | 1 |
| OLD (today's line 111) | FAIL | `[R1b]`, `[R12]`, `[R2]`..`[R8]`, `[R10]`, `[R11]` | 1 |

The "Old verb-list R1" column shows the retired pin printed `1` on every row, including the correct draft, because the draft carries the verb "runs at" (checked: only the `runs at` alternative matches, and rewording that phrase makes it print `0`). It therefore could not separate right from wrong on these rows; the plan-audit's own mutants A, B, C, E printed `0` for the old pin only because they lacked that phrase, which is the blacklist hole it reported. Either way the pin is dropped, not restated.

| Docs-row mutant (synthetic row) | `[W1]` (new bound, want at least 1) | `[W5]` (want 0) | Result |
|---|---|---|---|
| I correct (flag clause carries the only `xhigh`) | 1 | 0 | PASS |
| ko variant with `。`/no punctuation in the flag clause | 1 | 0 | PASS |
| G semicolon relocation ("`--effort ultracode` to launch; ultracode forces `xhigh`") | 0 | 0 | FAIL (caught) |
| G2 comma relocation | 0 | 0 | FAIL (caught) |
| H claim in the next sentence | 0 | 0 | FAIL (caught) |
| zh `。` relocation | 0 | 0 | FAIL (caught) |
| G3 comma-free relocation ("... also starts the session and ultracode always runs at `xhigh`") | 1 | 0 | PASSES — known residual, review of the diff |

| UW synthetic page | `[U1]` `^- .*xhigh` (want 0) | `[U1c]` page `xhigh` lines (want 0 for en/ja/zh) | Result |
|---|---|---|---|
| correct (bullet removed, off route named) | 0 | 0 | PASS |
| relabelled bullet ("- Reasoning level: set to `xhigh`") | 1 | 1 | FAIL (caught by both; the old label-keyed U1 would have printed 0) |
| prose form ("Reasoning effort is raised to `xhigh` ...") | 0 | 1 | FAIL (caught by `[U1c]`) |
| original (today) | 1 | 1 | FAIL |

Vacuity check: every positive pin above is `0` today (RED) per the ledger; zero pins that are `0` today (`[R1s]`, `[W5]`, `[R13]`, `[W11]`) are regression-guards, not RED flips; the controls are the entries marked regression-guard.

## Quality gates

- LSP gates: N/A — prose-only change to `.md` files, no Go symbols.
- Definition of Done: AC-001..AC-011 green with verbatim output cited; OQ-1 and OQ-3 still recorded as open in spec.md.

## Evidence ledger (RED-now cells; tree `ff7b64f6e`, scope files identical to `c50da9c2f`)

Each entry: the command, its verbatim stdout (one `path:count` line per file for multi-file invocations), and its exit code. Entries `[R*]` run on RS then RM; `[W*]`, `[M*]`, `[U*]`, `[C*]`, `[H-*]` on the four locales. The retired pins `[W9]` (and the verb-list `[R1]`, no longer in the ledger) are explained in AC-001 / AC-004. Regenerated at iteration 3 with the pin set of this document; the previous ledger tree was `0e7b6af5b` (same scope content).

```
[R1s] grep -c -E -- 'xhigh.*xhigh' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R1c] grep -c -F -- xhigh .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:1
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
    exit: 0
[R1b] grep -c -F -- 'combines `xhigh` reasoning' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:1
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
    exit: 0
[R2] grep -c -F -- 'leaves the effort level unchanged' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R3] grep -c -F -- 'independent on/off toggle' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R4] grep -c -F -- v2.1.284 .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R5] grep -c -F -- '/effort ultracode off' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R6] grep -c -F -- '--effort ultracode' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R7] grep -c -F -- 'starts the session at `xhigh`' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R8] grep -c -F -- '"ultracode": true' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R9] grep -c -F -- ultrathink. .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:1
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
    exit: 0
[R10] grep -c -F -- 'current session' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R11] grep -c -E -- 'v2.1.284.*"ultracode": true|"ultracode": true.*v2.1.284' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[R12] grep -c -F -- 'step back with `/effort high`' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:1
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
    exit: 0
[R13] grep -c -i -E -- slider .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      .claude/rules/moai/workflow/dynamic-workflows.md:0
      internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
    exit: 1
[W1] grep -c -E -- '--effort ultracode[^|.。;；,，、]*xhigh' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
    exit: 1
[W2] grep -c -F -- '"ultracode": true' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
    exit: 1
[W3] grep -c -E -- 'v2.1.284.*"ultracode": true|"ultracode": true.*v2.1.284' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
    exit: 1
[W4] grep -c -F -- '/effort ultracode off' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
    exit: 1
[W5] grep -c -E -- 'xhigh.*xhigh|xhigh.*--effort ultracode' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
    exit: 1
[W6] grep -c -E -- '/effort ultracode.*/effort high' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:1
      docs-site/content/en/claude-code/agentic/workflows.md:1
      docs-site/content/ja/claude-code/agentic/workflows.md:1
      docs-site/content/zh/claude-code/agentic/workflows.md:1
    exit: 0
[W7] grep -c -E -- '^\| `/effort ultracode` \|' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:1
      docs-site/content/en/claude-code/agentic/workflows.md:1
      docs-site/content/ja/claude-code/agentic/workflows.md:1
      docs-site/content/zh/claude-code/agentic/workflows.md:1
    exit: 0
[W8] grep -c -E -- v2.1.284 docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
    exit: 1
[W9] grep -c -E -- '/effort ultracode.*xhigh' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:1
      docs-site/content/en/claude-code/agentic/workflows.md:1
      docs-site/content/ja/claude-code/agentic/workflows.md:1
      docs-site/content/zh/claude-code/agentic/workflows.md:1
    exit: 0
[W10-ko] grep -c -F -- '현재 세션' docs-site/content/ko/claude-code/agentic/workflows.md
    stdout:
      1
    exit: 0
[W10-en] grep -c -F -- 'current session' docs-site/content/en/claude-code/agentic/workflows.md
    stdout:
      1
    exit: 0
[W10-ja] grep -c -F -- 現在のセッション docs-site/content/ja/claude-code/agentic/workflows.md
    stdout:
      1
    exit: 0
[W10-zh] grep -c -F -- 当前会话 docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      1
    exit: 0
[W11] grep -c -i -E -- 'slider|슬라이더|スライダー|滑块' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md docs-site/content/ko/multi-llm/_index.md docs-site/content/en/multi-llm/_index.md docs-site/content/ja/multi-llm/_index.md docs-site/content/zh/multi-llm/_index.md docs-site/content/ko/advanced/ultracode-workflows.md docs-site/content/en/advanced/ultracode-workflows.md docs-site/content/ja/advanced/ultracode-workflows.md docs-site/content/zh/advanced/ultracode-workflows.md docs-site/content/ko/claude-code/foundations/commands.md docs-site/content/en/claude-code/foundations/commands.md docs-site/content/ja/claude-code/foundations/commands.md docs-site/content/zh/claude-code/foundations/commands.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:0
      docs-site/content/en/claude-code/agentic/workflows.md:0
      docs-site/content/ja/claude-code/agentic/workflows.md:0
      docs-site/content/zh/claude-code/agentic/workflows.md:0
      docs-site/content/ko/multi-llm/_index.md:0
      docs-site/content/en/multi-llm/_index.md:0
      docs-site/content/ja/multi-llm/_index.md:0
      docs-site/content/zh/multi-llm/_index.md:0
      docs-site/content/ko/advanced/ultracode-workflows.md:0
      docs-site/content/en/advanced/ultracode-workflows.md:0
      docs-site/content/ja/advanced/ultracode-workflows.md:0
      docs-site/content/zh/advanced/ultracode-workflows.md:0
      docs-site/content/ko/claude-code/foundations/commands.md:0
      docs-site/content/en/claude-code/foundations/commands.md:0
      docs-site/content/ja/claude-code/foundations/commands.md:0
      docs-site/content/zh/claude-code/foundations/commands.md:0
    exit: 1
[M1] grep -c -E -- '^/effort ultracode # .*xhigh' docs-site/content/ko/multi-llm/_index.md docs-site/content/en/multi-llm/_index.md docs-site/content/ja/multi-llm/_index.md docs-site/content/zh/multi-llm/_index.md
    stdout:
      docs-site/content/ko/multi-llm/_index.md:1
      docs-site/content/en/multi-llm/_index.md:1
      docs-site/content/ja/multi-llm/_index.md:1
      docs-site/content/zh/multi-llm/_index.md:1
    exit: 0
[M2] grep -c -E -- '^/effort ultracode #' docs-site/content/ko/multi-llm/_index.md docs-site/content/en/multi-llm/_index.md docs-site/content/ja/multi-llm/_index.md docs-site/content/zh/multi-llm/_index.md
    stdout:
      docs-site/content/ko/multi-llm/_index.md:1
      docs-site/content/en/multi-llm/_index.md:1
      docs-site/content/ja/multi-llm/_index.md:1
      docs-site/content/zh/multi-llm/_index.md:1
    exit: 0
[U1] grep -c -E -- '^- .*xhigh' docs-site/content/ko/advanced/ultracode-workflows.md docs-site/content/en/advanced/ultracode-workflows.md docs-site/content/ja/advanced/ultracode-workflows.md docs-site/content/zh/advanced/ultracode-workflows.md
    stdout:
      docs-site/content/ko/advanced/ultracode-workflows.md:1
      docs-site/content/en/advanced/ultracode-workflows.md:1
      docs-site/content/ja/advanced/ultracode-workflows.md:1
      docs-site/content/zh/advanced/ultracode-workflows.md:1
    exit: 0
[U1c] grep -c -F -- xhigh docs-site/content/ko/advanced/ultracode-workflows.md docs-site/content/en/advanced/ultracode-workflows.md docs-site/content/ja/advanced/ultracode-workflows.md docs-site/content/zh/advanced/ultracode-workflows.md
    stdout:
      docs-site/content/ko/advanced/ultracode-workflows.md:5
      docs-site/content/en/advanced/ultracode-workflows.md:1
      docs-site/content/ja/advanced/ultracode-workflows.md:1
      docs-site/content/zh/advanced/ultracode-workflows.md:1
    exit: 0
[U2] grep -c -F -- '/effort high' docs-site/content/ko/advanced/ultracode-workflows.md docs-site/content/en/advanced/ultracode-workflows.md docs-site/content/ja/advanced/ultracode-workflows.md docs-site/content/zh/advanced/ultracode-workflows.md
    stdout:
      docs-site/content/ko/advanced/ultracode-workflows.md:1
      docs-site/content/en/advanced/ultracode-workflows.md:0
      docs-site/content/ja/advanced/ultracode-workflows.md:0
      docs-site/content/zh/advanced/ultracode-workflows.md:0
    exit: 0
[U3] grep -c -F -- '/effort ultracode off' docs-site/content/ko/advanced/ultracode-workflows.md docs-site/content/en/advanced/ultracode-workflows.md docs-site/content/ja/advanced/ultracode-workflows.md docs-site/content/zh/advanced/ultracode-workflows.md
    stdout:
      docs-site/content/ko/advanced/ultracode-workflows.md:0
      docs-site/content/en/advanced/ultracode-workflows.md:0
      docs-site/content/ja/advanced/ultracode-workflows.md:0
      docs-site/content/zh/advanced/ultracode-workflows.md:0
    exit: 1
[U4] grep -c -F -- '세 가지가 함께 바뀝니다' docs-site/content/ko/advanced/ultracode-workflows.md
    stdout:
      1
    exit: 0
[U5] grep -c -F -- '세션 경계를 넘지 않습니다' docs-site/content/ko/advanced/ultracode-workflows.md
    stdout:
      1
    exit: 0
[C1en] grep -c -F -- 'simultaneously an `/effort` level' docs-site/content/en/claude-code/foundations/commands.md
    stdout:
      1
    exit: 0
[C1ko] grep -c -F -- '동시에 `/effort` 레벨입니다' docs-site/content/ko/claude-code/foundations/commands.md
    stdout:
      1
    exit: 0
[C1ja] grep -c -F -- '`/effort` のレベルでもあります' docs-site/content/ja/claude-code/foundations/commands.md
    stdout:
      1
    exit: 0
[C1zh] grep -c -F -- '也是一个 `/effort` 等级' docs-site/content/zh/claude-code/foundations/commands.md
    stdout:
      1
    exit: 0
[C2en] grep -c -F -- 'plus `auto`, and `ultracode`' docs-site/content/en/claude-code/foundations/commands.md
    stdout:
      1
    exit: 0
[C2ko] grep -c -F -- '그리고 워크플로우 오케스트레이션을 켜는 `ultracode`가 있습니다' docs-site/content/ko/claude-code/foundations/commands.md
    stdout:
      1
    exit: 0
[C3-en] grep -c -i -E -- 'ultracode.*toggle|toggle.*ultracode' docs-site/content/en/claude-code/foundations/commands.md
    stdout:
      0
    exit: 1
[C3-ko] grep -c -i -E -- 'ultracode.*토글|토글.*ultracode' docs-site/content/ko/claude-code/foundations/commands.md
    stdout:
      0
    exit: 1
[C3-ja] grep -c -i -E -- 'ultracode.*トグル|トグル.*ultracode' docs-site/content/ja/claude-code/foundations/commands.md
    stdout:
      0
    exit: 1
[C3-zh] grep -c -i -E -- 'ultracode.*开关|开关.*ultracode' docs-site/content/zh/claude-code/foundations/commands.md
    stdout:
      0
    exit: 1
[H-WF] grep -c '^#' docs-site/content/ko/claude-code/agentic/workflows.md docs-site/content/en/claude-code/agentic/workflows.md docs-site/content/ja/claude-code/agentic/workflows.md docs-site/content/zh/claude-code/agentic/workflows.md
    stdout:
      docs-site/content/ko/claude-code/agentic/workflows.md:11
      docs-site/content/en/claude-code/agentic/workflows.md:11
      docs-site/content/ja/claude-code/agentic/workflows.md:11
      docs-site/content/zh/claude-code/agentic/workflows.md:11
    exit: 0
[H-ML] grep -c '^#' docs-site/content/ko/multi-llm/_index.md docs-site/content/en/multi-llm/_index.md docs-site/content/ja/multi-llm/_index.md docs-site/content/zh/multi-llm/_index.md
    stdout:
      docs-site/content/ko/multi-llm/_index.md:13
      docs-site/content/en/multi-llm/_index.md:13
      docs-site/content/ja/multi-llm/_index.md:13
      docs-site/content/zh/multi-llm/_index.md:13
    exit: 0
[H-UW] grep -c '^#' docs-site/content/ko/advanced/ultracode-workflows.md docs-site/content/en/advanced/ultracode-workflows.md docs-site/content/ja/advanced/ultracode-workflows.md docs-site/content/zh/advanced/ultracode-workflows.md
    stdout:
      docs-site/content/ko/advanced/ultracode-workflows.md:24
      docs-site/content/en/advanced/ultracode-workflows.md:20
      docs-site/content/ja/advanced/ultracode-workflows.md:20
      docs-site/content/zh/advanced/ultracode-workflows.md:20
    exit: 0
[H-CM] grep -c '^#' docs-site/content/ko/claude-code/foundations/commands.md docs-site/content/en/claude-code/foundations/commands.md docs-site/content/ja/claude-code/foundations/commands.md docs-site/content/zh/claude-code/foundations/commands.md
    stdout:
      docs-site/content/ko/claude-code/foundations/commands.md:23
      docs-site/content/en/claude-code/foundations/commands.md:23
      docs-site/content/ja/claude-code/foundations/commands.md:18
      docs-site/content/zh/claude-code/foundations/commands.md:18
    exit: 0
[X-cmp] cmp .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md
    stdout:
      (empty)
    exit: 0
[X-baseline-diff] git diff --stat -- docs-site/.locale-parity-baseline
    stdout:
      (empty)
    exit: 0
```
