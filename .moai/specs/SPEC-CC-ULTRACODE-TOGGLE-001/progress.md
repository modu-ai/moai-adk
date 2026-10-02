# SPEC-CC-ULTRACODE-TOGGLE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02
plan_phase_notes: Tier M artifacts authored (spec.md, plan.md, acceptance.md). Iteration 1 plan-audit FAIL 0.75 (`.moai/reports/t1416/plan-audit.md`); v0.1.1 repair addresses D1-D10 (D11 observed: hugo rc 0, 0 WARN/ERROR at `0e7b6af5b`). RED-now cells re-measured at HEAD `0e7b6af5b` with exit codes (acceptance.md Evidence ledger); scope files identical to the original pin `c50da9c2f`. Iteration 2 plan-audit FAIL 0.86 (only D2 unresolved: verb-list pin); v0.1.2 repair replaces it with a structural `xhigh` count guard, records a rewording probe in scratch copies, applies N2-N4, and regenerates the ledger at HEAD `ff7b64f6e` (scope files identical to `c50da9c2f`). Final audit (iteration 3) pending.

Gaps (explicitly unobserved at plan time):
- Upstream facts F1-F7: F1-F6 were read through a fetch tool that summarizes pages; F7 (the `ultracode` settings entry) was read from the raw page with `curl` and tag-stripping on 2026-10-02 and matches the audit's own raw read. F1 was independently re-read by the plan-audit.
- OQ-1 (slider-toggle persistence across sessions) needs a live observation; not performed.
- `make build` and `go test ./internal/template/...` (AC-011) not run: the mirror is not yet edited.
- `moai spec lint` was run with the installed `moai` binary, which the audit found to be a strict ancestor of HEAD (build `d194083fb`); its clean result is not cited as corroboration.
- The traceability collector's `awk` verb is refused by the worktree guard; an equivalent check was run with python (see the v0.1.1 report).

## §E.2 Run-phase Evidence

Run-phase actor: manager-develop (cycle_type tdd; the SPEC's RED-now pins are the RED cells). Base at run start: HEAD `0e7087faa`, branch `WT-ultracode-toggle-wording`. Pin script (all pins, one invocation each, under `LC_ALL=en_US.UTF-8`, which exists in `locale -a`): `.../scratchpad/t1416-run/pins.sh` (machine-local scratch, not a citation target; every command and its output below is reproduced inline). Judging tools: `grep` (ugrep, count lines matched per path), `cmp`, `hugo` (`/opt/homebrew/bin/hugo`), `go test`/`make build` run from this tree (the bin/moai built by `make build` carries commit `abf8ba627`-dirty; no installed `moai` build was used to judge anything). Count lines below are printed `path:count` per file in the order the tool printed them.

### M1 + M3 — rule source and template mirror (commit `abf8ba627`)

RED before (HEAD `0e7087faa`), matched the plan ledger: `[R1b]` 1,1; `[R2]` `[R3]` `[R4]` `[R5]` `[R6]` `[R7]` `[R8]` `[R10]` `[R11]` 0,0 (exit 1); `[R12]` 1,1; `[R1s]` 0,0; `[R1c]` 1,1; `[R9]` 1,1; `[R13]` 0,0.

GREEN after (HEAD `abf8ba627`), each over RS then RM:

| Pin | Command (grep -c ... RS RM) | Before | After | Exit |
|-----|-----------------------------|--------|-------|------|
| R1s | `-E 'xhigh.*xhigh'` | 0,0 | 0,0 | 1 |
| R1c | `-F xhigh` | 1,1 | 1,1 | 0 |
| R1b | `-F 'combines \`xhigh\` reasoning'` | 1,1 | 0,0 | 1 |
| R2 | `-F 'leaves the effort level unchanged'` | 0,0 | 1,1 | 0 |
| R3 | `-F 'independent on/off toggle'` | 0,0 | 1,1 | 0 |
| R4 | `-F v2.1.284` | 0,0 | 1,1 | 0 |
| R5 | `-F '/effort ultracode off'` | 0,0 | 1,1 | 0 |
| R6 | `-F -- '--effort ultracode'` | 0,0 | 1,1 | 0 |
| R7 | `-F 'starts the session at \`xhigh\`'` | 0,0 | 1,1 | 0 |
| R8 | `-F '"ultracode": true'` | 0,0 | 1,1 | 0 |
| R9 | `-F 'ultrathink.'` | 1,1 | 1,1 | 0 |
| R10 | `-F 'current session'` | 0,0 | 1,1 | 0 |
| R11 | `-E 'v2.1.284.*"ultracode": true\|"ultracode": true.*v2.1.284'` | 0,0 | 1,1 | 0 |
| R12 | `-F 'step back with \`/effort high\`'` | 1,1 | 0,0 | 1 |
| R13 | `-i -E slider` | 0,0 | 0,0 | 1 |
| X-cmp | `cmp RS RM` | empty | empty | 0 |

Mutant sanity (scratch copies of the final RS under the session scratchpad; the worktree was untouched): mutant A (second `xhigh` on the bullet line, "pairs with `xhigh` reasoning") printed `[R1s]` mutA 1 (red; final file 0); mutant J (coupling sentence on its own line) printed `[R1c]` mutJ 2 (red; final file 1). Both go red as designed.

`dynamic-workflows.md` carries a `paths:` frontmatter key, so it is not on the always-loaded surface and the `rule-authoring.md` growth statement does not apply.

### M2 — docs-site, four locales (commit `be5d56b8c`, tree `4ad5104f6`)

Order: ko of all four pages, then en, then ja and zh. Pins, per locale in the order ko, en, ja, zh (the tool's print order differed; each count is matched to its own path):

| Pin | Before | After | Exit after |
|-----|--------|-------|------------|
| W1 flag clause `--effort ultracode[^\|.。;；,，、]*xhigh` | 0,0,0,0 | 1,1,1,1 | 0 |
| W2 `"ultracode": true` | 0,0,0,0 | 1,1,1,1 | 0 |
| W3 `v2.1.284` and key on one line | 0,0,0,0 | 1,1,1,1 | 0 |
| W4 `/effort ultracode off` | 0,0,0,0 | 1,1,1,1 | 0 |
| W5 two `xhigh`, or `xhigh` before the flag | 0,0,0,0 | 0,0,0,0 | 1 |
| W6 `/effort ultracode.*/effort high` | 1,1,1,1 | 0,0,0,0 | 1 |
| W7 row shape `^\| \`/effort ultracode\` \|` | 1,1,1,1 | 1,1,1,1 | 0 |
| W8 `v2.1.284` | 0,0,0,0 | 1,1,1,1 | 0 |
| W10 current-session literal (ko, en, ja, zh) | 1 each | 1 each | 0 |
| W11 slider words over all 16 files | 0 each | 0 each | 1 |
| D2 `^\| \`/effort ultracode\` \|[^\|]*\|$` (row keeps its column count) | 1,1,1,1 | 1,1,1,1 | 0 |
| M1 `^/effort ultracode # .*xhigh` | 1,1,1,1 | 0,0,0,0 | 1 |
| M2 `^/effort ultracode #` | 1,1,1,1 | 1,1,1,1 | 0 |
| U1 `^- .*xhigh` | 1,1,1,1 | 0,0,0,0 | 1 |
| U1c lines with `xhigh` | 5,1,1,1 | 4,0,0,0 | 0 |
| U2 `/effort high` | 1,0,0,0 | 0,0,0,0 | 1 |
| U3 `/effort ultracode off` | 0,0,0,0 | 1,1,1,1 | 0 |
| U4 ko "세 가지가 함께 바뀝니다" | 1 | 0 | 1 |
| U5 ko session-boundary callout | 1 | 1 | 0 |
| C1en C1ko C1ja C1zh "is an /effort level" sentence | 1 each | 0 each | 1 |
| C2en C2ko L137 level-list wording | 1 each | 0 each | 1 |
| C3-en C3-ko C3-ja C3-zh toggle-word line | 0 each | 2, 2, 1, 1 (en, ko, ja, zh) | 0 |
| H-WF | 11,11,11,11 | 11,11,11,11 | 0 |
| H-ML | 13,13,13,13 | 13,13,13,13 | 0 |
| H-UW | 24,20,20,20 | 24,20,20,20 | 0 |
| H-CM | 23,23,18,18 | 23,23,18,18 | 0 |
| X-baseline-diff | empty | empty | 0 |

Retired pin `[W9]` (`/effort ultracode.*xhigh`) still prints 1 per locale by design: the required flag clause sits on the same row.

Hand checks of the carried audit debt (the diff in `be5d56b8c` is the evidence):
- D1: each edited row says toggle AND effort level unchanged, plus the off route, the settings key with `v2.1.284`, and the flag exception: ko `workflows.md:113`, en `workflows.md:105`, ja `workflows.md:105`, zh `workflows.md:105` (under `docs-site/content/<loc>/claude-code/agentic/`).
- D2: `grep -c -E '^\| \`/effort ultracode\` \|[^|]*\|$'` printed `1` for all four rows (table above).
- D3: pins containing CJK text ran under `LC_ALL=en_US.UTF-8`. The matched flag clause per locale (`grep -o` of the W1 pattern): ko "`--effort ultracode` 실행 플래그는 세션을 `xhigh", en "`--effort ultracode` launch flag also starts the session at `xhigh", ja "`--effort ultracode` 起動フラグはセッションを `xhigh", zh "`--effort ultracode` 启动参数会同时让会话以 `xhigh"; none contains `| . 。 ; ； , ， 、`.
- Known residual: no sentence couples ultracode to an effort level in any wording; read by eye against the full docs diff (`git diff -U0 -- docs-site/content`).

### M4 — verification

| Check | Command | Result |
|-------|---------|--------|
| AC-003 mirror parity | `cmp .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` | empty output, exit 0 |
| AC-011 build | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && make build > <scratch>/make-build.log 2>&1 && echo MAKE_BUILD_EXIT_0` | printed `MAKE_BUILD_EXIT_0`; log tail ends `catalog.yaml updated successfully (13900 bytes)` then the `go build ... -o bin/moai ./cmd/moai` line; `git status --short` showed no unexpected tracked change |
| AC-011 test | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 900 go test ./internal/template/... > <scratch>/go-test-template.log 2>&1 && echo GO_TEST_EXIT_0` | first run: NOT a pass. Go's default 10-minute test timeout fired inside the `internal/template` package (`panic: test timed out after 10m0s`, running test `TestACNA010_FailOpenOnMissingInputs`), so the line read `FAIL github.com/modu-ai/moai-adk/internal/template 600.780s`, with `ok .../agentemit 3.019s`, `ok .../commandemit 0.722s`, `[no test files]` for `scripts`; `grep -c -E '^--- FAIL'` over that log printed `0` (exit 1), so no assertion failed, but the run is a tool failure (timeout), not evidence of a pass. Machine load at that time: `uptime` printed load averages 599.21 632.42 605.09 (shared machine). Rerun (reason: timeout reclassified as TOOL_FAILURE): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 2900 go test -timeout 45m ./internal/template/ > <scratch>/go-test-template-rerun.log 2>&1 && echo GO_TEST_RERUN_EXIT_0` printed `GO_TEST_RERUN_EXIT_0`, task exit code 0; log: `ok  	github.com/modu-ai/moai-adk/internal/template	349.908s`. The two sub-packages were not re-run (they passed in the first run). |
| AC-009 hugo | `cd docs-site && /opt/homebrew/bin/hugo --minify --gc --destination <scratch>/hugo-out > <scratch>/hugo.log 2>&1 && echo HUGO_EXIT_0` | printed `HUGO_EXIT_0`; `grep -c -E 'WARN\|ERROR' hugo.log` printed `0` (exit 1); stats table: Pages 189 / 187 / 187 / 187; build output stayed in scratch |
| AC-009 URL blacklist | `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content` | no output, exit 1 |
| AC-009 Mermaid | `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content` | no output, exit 1 |
| AC-010 scope | `git status --porcelain` before the M2 commit listed only the 16 docs files; `git diff --stat c50da9c2f HEAD` lists the 2 rule files, the 16 docs files, and the plan-phase artifacts under `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/` and `.moai/reports/t1416/`; the session-handoff/moai.md/CHANGELOG/appendix `git diff --stat` printed nothing, exit 0 | as stated |

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-02
run_commits: `abf8ba627` (M1+M3 rule source and template mirror), `be5d56b8c` (M2 docs-site four locales), and the final run-record commit carrying this section and the `spec.md` frontmatter transition (its SHA is not stated here: a commit cannot cite its own hash; read it from `git log -1 -- .moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/progress.md`)
run_notes: M1 and M3 landed in one commit because the rule file and its template mirror must stay byte-identical at every commit; the status transition `draft` to `in-progress` rides the final run-record commit as the run instruction directed, not the first milestone commit.

Gaps (explicitly unobserved in this run):
- Upstream facts F1-F7 were taken from the SPEC; this run did not re-fetch the upstream pages.
- OQ-1 (slider-toggle persistence) was not observed, and no edited text mentions the slider.
- Docs wording was written inline to the native-idiom standard; the `moai-domain-humanize` pass and a native-speaker read of the ja and zh sentences were not run.
- The hugo output was judged by exit code and a `WARN|ERROR` count of 0; the rendered pages were not opened.
- A coupling stated without the token `xhigh` is judged by reading the diff only (the SPEC's known residual).
- AC-011 `go test` passed only on a rerun with a changed command (`-timeout 45m`, main package only) after the first `timeout 900 go test ./internal/template/...` run hit Go's default 10-minute test timeout under machine load ~600; the sub-packages `agentemit` and `commandemit` passed in the first run and were not repeated. No full-suite run was done (CI owns it).

Residual-risk:
- The rule sentence "choosing a different effort level does not turn it off" is an inference from F1 ("stays on at any effort level") and F3; it names no slider and no picker, but an auditor may want it checked against the live behavior.
- The four commands-page sentences (L79 in ko and en, L80 in ja and zh) carry the `v2.1.284` qualifier in prose form; if the toggle's introduction version is ever restated upstream, those four sentences plus the four workflows rows, the four ultracode-workflows pages that name it, and the rule bullet are the places to update.
- The 4-locale rows are long single table cells; they render as one row (D2 holds) but are visually dense.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
