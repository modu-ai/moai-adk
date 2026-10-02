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

sync_status: audit-ready
sync_complete_at: 2026-10-02
sync_commit_sha: pending-backfill
sync_phase_actor: manager-docs (card t1416, worktree `.moai/worktrees/t1416`, branch `WT-ultracode-toggle-wording`)
sync_tree_at_measurement: HEAD `90d5bac45` plus the uncommitted `spec.md` frontmatter edit this commit carries (`status: in-progress` to `status: completed`); base of the card `c50da9c2f`
sync_sha_backfill_note: the `sync_commit_sha` slot holds the canonical `pending-backfill` placeholder in the sync commit itself (a commit cannot cite its own hash); the real SHA is backfilled in the following `docs(SPEC-CC-ULTRACODE-TOGGLE-001): backfill sync_commit_sha (t1416)` commit, which changes `progress.md` only.

b12_self_test_a: no CHANGELOG entry was written, on purpose, so no emission was attempted; informational pre-emission grep `grep -c 'SPEC-CC-ULTRACODE-TOGGLE-001' CHANGELOG.md` printed `0` (exit 1).
b12_self_test_b: informational only (nothing emitted). `ac_source=.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/acceptance.md`, tier M, state `resolved` (`test -s` true). The B12 counter printed `11` on stdout and `live=11 excluded=0 ambiguous=0` on stderr, exit 0, so a later CHANGELOG entry would cite 11 acceptance criteria (AC-001..AC-011); the count is not a RED flag.
b12_self_test_c: not applicable, no CHANGELOG entry carries file paths.
changelog_entry_position: none. The SPEC's change map (`spec.md` §3) excludes `CHANGELOG.md`, `spec.md` §4 lists historical CHANGELOG entries as out of scope, and the shared `CHANGELOG.md` is a conflict hot spot across lanes. The leader decides whether and where the entry is written (this is a user-visible docs and rule wording fix, so an `[Unreleased]` "Fixed" or "Changed" line is plausible).
frontmatter_status_transitions: `spec.md` `in-progress` to `completed` in this single sync commit (the `implemented` step is merged into it per the 3-phase close). `updated:` already read `2026-10-02`, the sync date, so that line did not change. `plan.md` and `acceptance.md` carry no `status:` field (artifact statelessness) and are untouched; `progress.md` has no frontmatter.
canary_compliance_check: not applicable (this SPEC defines no forward-looking policy that its own sync tests).

### Verification matrix (5-section evidence format)

Every row below is a command run in THIS sync run against THIS tree. Multi-file `grep -c` invocations print one `path:count` line per file in an order `ugrep` does not guarantee, so counts are re-keyed to path ids; RS is `.claude/rules/moai/workflow/dynamic-workflows.md`, RM is `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`, and docs rows are listed in the order ko, en, ja, zh. Exit codes were captured with a `; echo "exit=$?"` suffix on each command. All CJK pins ran under `LC_ALL=en_US.UTF-8` (`locale` printed `LANG="en_US.UTF-8"`, and `en_US.UTF-8` is in `locale -a`).

**Claim 1 (AC-001, AC-002, AC-003, AC-010 rule part): the rule source is corrected and the mirror is byte-identical.**

Evidence (command, then RS and RM counts, then exit):
- R1s `grep -c -E -- 'xhigh.*xhigh' RS RM`: 0, 0, exit 1 (regression guard stays 0).
- R1c `grep -c -F -- xhigh RS RM`: 1, 1, exit 0 (exactly one `xhigh` line per file).
- R1b `grep -c -F -- 'combines \`xhigh\` reasoning' RS RM`: 0, 0, exit 1 (was 1, 1 in the ledger).
- R2 `leaves the effort level unchanged`: 1, 1, exit 0. R3 `independent on/off toggle`: 1, 1, exit 0. R4 `v2.1.284`: 1, 1, exit 0.
- R5 `/effort ultracode off`: 1, 1, exit 0. R6 `--effort ultracode`: 1, 1, exit 0. R7 `starts the session at \`xhigh\``: 1, 1, exit 0.
- R8 `"ultracode": true`: 1, 1, exit 0. R10 `current session`: 1, 1, exit 0.
- R11 `grep -c -E -- 'v2.1.284.*"ultracode": true|"ultracode": true.*v2.1.284'`: 1, 1, exit 0.
- R12 `step back with \`/effort high\``: 0, 0, exit 1 (was 1, 1).
- R9 `ultrathink.` retention control: 1, 1, exit 0.
- R13 `grep -c -i -E -- slider RS RM`: 0, 0, exit 1.
- X-cmp `cmp RS RM`: empty output, exit 0.
- Diff read by eye (`git diff -U0 c50da9c2f HEAD -- .claude/rules/moai/workflow/dynamic-workflows.md`): one line (L111) changed; the new bullet names the toggle, the unchanged effort level, the off route, the launch-flag exception with the single `xhigh`, the current-session scope of the command form, the settings-key route with `v2.1.284` on the same sentence, and keeps the `ultrathink.` sentence.

**Claim 2 (AC-004): the workflows row is corrected in four locales and keeps its shape.**

Evidence (ko, en, ja, zh counts; exit):
- W1 flag clause `--effort ultracode[^|.。;；,，、]*xhigh`: 1, 1, 1, 1, exit 0. W2 `"ultracode": true`: 1, 1, 1, 1, exit 0. W3 `v2.1.284` and the key on one line: 1, 1, 1, 1, exit 0. W4 `/effort ultracode off`: 1, 1, 1, 1, exit 0. W8 `v2.1.284`: 1, 1, 1, 1, exit 0.
- W5 two `xhigh`, or `xhigh` before the flag: 0, 0, 0, 0, exit 1. W6 `/effort ultracode.*/effort high`: 0, 0, 0, 0, exit 1.
- W7 row shape `^\| \`/effort ultracode\` \|`: 1, 1, 1, 1, exit 0. D2 `^\| \`/effort ultracode\` \|[^|]*\|$`: 1, 1, 1, 1, exit 0.
- W10 current-session literal (`현재 세션` / `current session` / `現在のセッション` / `当前会话`, each on its own locale file): 1 each, exit 0.
- Flag clause actually matched (`grep -o`), none containing `| . 。 ; ； , ， 、`: ko "\`--effort ultracode\` 실행 플래그는 세션을 \`xhigh", en "\`--effort ultracode\` launch flag also starts the session at \`xhigh", ja "\`--effort ultracode\` 起動フラグはセッションを \`xhigh", zh "\`--effort ultracode\` 启动参数会同时让会话以 \`xhigh".

**Claim 3 (AC-005, AC-006, AC-007): the multi-llm comment, the ultracode-workflows page, and the commands page are corrected in four locales.**

Evidence (ko, en, ja, zh unless stated):
- M1 `^/effort ultracode # .*xhigh`: 0, 0, 0, 0, exit 1. M2 `^/effort ultracode #`: 1, 1, 1, 1, exit 0.
- U1 `^- .*xhigh`: 0, 0, 0, 0, exit 1. U1c `grep -c -F -- xhigh`: 4, 0, 0, 0, exit 0 (ko keeps four unrelated `xhigh` lines by design). U2 `/effort high`: 0, 0, 0, 0, exit 1. U3 `/effort ultracode off`: 1, 1, 1, 1, exit 0.
- U4 ko "세 가지가 함께 바뀝니다": 0, exit 1. U5 ko session-boundary callout: 1, exit 0 (retained).
- C1en, C1ko, C1ja, C1zh (the old "is an /effort level" sentences): 0 each, exit 1. C2en, C2ko (the old L137 level-list wording): 0 each, exit 1.
- C3-en 2, C3-ko 2, C3-ja 1, C3-zh 1, each exit 0 (a line with `ultracode` and the locale's toggle word).
- ko `ultracode-workflows.md` L65-L70 read by eye: the lead sentence now says "두 가지가 함께 바뀝니다" and exactly two effects bullets follow; the en page (L111-L115) lists two effects, then the toggle and off-route paragraph.

**Claim 4 (AC-008, AC-009, AC-010 docs part): no new divergence, no URL or Mermaid regression, scope held.**

Evidence:
- H-WF (11, 11, 11, 11), H-ML (13, 13, 13, 13), H-UW (24, 20, 20, 20), H-CM (23, 23, 18, 18) read per path from `grep -c '^#'`, all exit 0, identical to the AC-008 baselines.
- `git diff --stat c50da9c2f HEAD -- <session-handoff trio> .claude/output-styles/moai/moai.md CHANGELOG.md .moai/docs/session-handoff-appendix.md docs-site/.locale-parity-baseline internal/template/templates/.moai/docs/session-handoff-appendix.md`: empty output, exit 0.
- `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content`: no output, exit 1. `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content`: no output, exit 1.
- W11 slider words over the 16 docs files: 0 for each file, exit 1.
- `git diff --name-only c50da9c2f HEAD` lists the 2 rule files, the 16 docs-site files, the four SPEC artifacts, and the four plan-phase records under `.moai/reports/t1416/`; no other path.
- `git status --porcelain` before staging printed only ` M .moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/spec.md`.

**Claim 5 (carried audit debt D1-D3 hold on the final diff).** Checked by hand against `git diff -U0 c50da9c2f HEAD -- docs-site/content`:
- D1 (toggle and effort-unchanged statements in the docs rows): each of the four rows says it is a toggle, that turning it on or off leaves the effort level unchanged, the off route, the settings key with `v2.1.284`, and the `--effort ultracode` exception. Resolved.
- D2 (row column count): every row is a two-cell row (`grep` D2 pin 1, 1, 1, 1) under a two-column header (`| 항목 | 설명 |`, `| Item | Description |`, `| 項目 | 説明 |`, `| 条目 | 说明 |` with a `|------|------|` separator, read in the files). Resolved.
- D3 (locale prerequisite of W1): the pins ran under `LC_ALL=en_US.UTF-8`, shown above. Resolved.

**Claim 6 (lifecycle tooling).**

Evidence:
- `moai version` printed `moai-adk v3.2.0-rc.24`, commit `c50da9c2f`, built 2026-10-02T05:34:44Z, exit 0. `git merge-base --is-ancestor c50da9c2f HEAD` exit 0, so the installed binary is a strict ancestor of HEAD (the card base, lacking every commit of this card). The MCP server build `d194083fb` is likewise an ancestor of HEAD (`is-ancestor` exit 0).
- `moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001` (run with `spec.md` already carrying `status: completed`): `✓ No findings — all SPEC documents are valid`, exit 0. Not cited as corroboration (lagging binary).
- `mcp__moai__spec_audit` with `filter_spec` and this worktree as `project_root` (before this section existed): `total_specs 1, grandfathered 0, modern_era_clean 1`, one `EraAutoDetected` INFO finding (H-5), no drift finding. Not cited as corroboration (lagging server build).
- Re-run after this section was written (uncommitted tree, `sync_commit_sha` still the placeholder): `moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001` printed `✓ No findings — all SPEC documents are valid`, exit 0; `mcp__moai__spec_audit` (same scoping) printed `modern_era_clean 1, grandfathered 0` with one INFO `EraAutoDetected` finding, `heuristic_matched` now `H-4 (§E.2 + §E.4 + sync_commit_sha)`, and no drift finding. Same lag caveat applies.

**Baseline-attribution.** All rows were measured in this run against the tree named in `sync_tree_at_measurement`, with `grep` (ugrep), `cmp`, and `git` from this machine. The run-phase build and test results are NOT re-measured here: they are cited from §E.2 (`make build` exit 0; `go test` of `internal/template/` exit 0 on a rerun with `-timeout 45m` after a first run hit Go's default 10-minute timeout under machine load), and this sync run did not re-run `make build`, `go test`, or the hugo build.

**Gaps (explicitly unobserved).**
- The AC-011 build and test and the AC-009 hugo exit gate were not re-run in this sync run; only the §E.2 record exists, and the first `go test` run in that record was a timeout, not a pass.
- No rendered docs page was opened, and no native-speaker or `moai-domain-humanize` pass was run on the ko, ja, and zh sentences (carried from §E.3).
- Upstream facts F1-F7 were not re-fetched in this run.
- OQ-1 (slider-toggle persistence) was not observed (see below).
- A coupling stated without the token `xhigh` is judged only by reading the diff (the SPEC's known residual); the diff was read and no such sentence was found.
- `ugrep` print order and the first command batches: a first rule-source batch printed no exit codes and was repeated with exit capture; a first docs batch failed on zsh word splitting (`No such file or directory`, exit 2) and was repeated with array expansion. Only the repeated, exit-captured results are cited.
- `moai spec lint` and the audit ran on binaries and a server older than HEAD, so a clean result there says only that the older checks reported nothing.

**Residual-risk.**
- The rule sentence "choosing a different effort level does not turn it off" is an inference from upstream F1 and F3 (carried from §E.3).
- The upstream version qualifier `v2.1.284` appears in 1 rule bullet (RS and RM), 4 workflows rows, 4 commands sentences, and the ultracode-workflows pages; if upstream restates it, those are the places to update.
- The `/effort [low|medium|high|xhigh|max|ultracode|auto]` syntax rows in the model-policy and commands pages stay as written (out of scope, command syntax still accepted per `spec.md` §4).

### Carried audit debt and disposition

- D1, D2, D3: resolved, checked by hand (Claim 5).
- D4 (cosmetic, known debt, not fixed here): `spec.md` L50 says the evidence baseline was measured at HEAD `0e7b6af5b`, while `acceptance.md` L5 and L115 state the regenerated ledger tree is `ff7b64f6e`; and the `spec.md` HISTORY list reads v0.1.0, v0.1.2, v0.1.1 (out of order). Both live in the `spec.md` body, which manager-docs may not edit, so they are reported to the leader as known debt (a `manager-spec` correction would be a non-transition frontmatter-or-body fix; it has no behavioural effect).

### Open question carried forward

OQ-1 (does flipping the slider's Ultracode toggle persist across sessions) remains OPEN. It needs a live observation (toggle in a fresh session, grep the settings files, start a new session) that no run performed. The SPEC forbids claiming it (REQ-011): the edited text neither asserts slider persistence nor mentions the slider, confirmed by pins R13 (0, 0) and W11 (0 for all 16 files) above. OQ-3 (the local changelog snapshot predates 2.1.284) also remains as recorded in `spec.md`.

### Where the SPEC and reality disagreed

- `acceptance.md` AC-010 says `git status --porcelain` shows no entry outside the 2 rule files, the 16 docs files, and `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/`; the branch also carries four committed plan-phase records under `.moai/reports/t1416/` (plan-audit iterations and the decision record). They are the card's evidence path, not scope creep, but the AC text does not list them.
- `plan.md` lists M1 (rule source) and M3 (mirror) as separate milestones; they landed in one commit because the pair must stay byte-identical at every commit (recorded in §E.3).
- No CHANGELOG entry was written even though the lifecycle normally closes with one; this follows the SPEC's change map and is left to the leader (see `changelog_entry_position`).
