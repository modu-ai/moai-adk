# Acceptance — SPEC-CC-HAIKU55-STATUSLINE-001

> Verification layer. Every criterion is binary-testable; the command IS the evidence.
> All paths relative to the worktree root `.moai/worktrees/t1605`.
> Revision r2: RED-now evidence ledger added (audit D1); REQ-001/REQ-011 coverage added
> (D2/D3); AC-004 rebuilt on the wired lint verb (D4); config-invariance made
> commit-range-aware (D5); AC-012 claims narrowed to the script's actual checks (D6);
> :179 lint-scope preservation asserted (D7); `-run` anchored (D9).
> Revision r3: legal test selector `^TestAgentType$` + function name pinned (D-R2-1);
> RED-AC-005a/b/c preserve-red cells added (D-R2-2); RED-AC-004 re-measured as
> preserve-red on the current artifacts + Then clause aligned to the verb's real clean-run
> output (D-R2-3); AC-015 cond 3 backslash-escape artifact removed, rate/unit greps
> tightened to `-F` unit-bearing strings + `over 100K` threshold (D-R2-4/D-R2-5);
> AC-016 pinned string extended through the `claude_models` tail (D-R2-6); AC-012 check 4
> claim narrowed to the recipe's stale-domain blacklist (D-R2-7).

## §D.0 RED-now Evidence Ledger (P1 ACs — verification-completeness.md §2.1 two-cell adoption)

Every cell below was executed and observed on **this tree, SHA `81786284e6e5ee2fe5fbe7b498d649b27c3e0109`**
(pinned; `git rev-parse HEAD` → `81786284e6e5ee2fe5fbe7b498d649b27c3e0109`, exit 0, observed 2026-10-08,
tree clean except the untracked SPEC directory). Format per cell: command · verbatim stdout · exit code ·
why-red · green path (the milestone that flips it).

| Cell | Command | Verbatim stdout | Exit | Why red (stated) | Green path |
|------|---------|-----------------|------|------------------|------------|
| RED-AC-001a | `grep -n "Haiku 5.5 (1M)" .claude/rules/moai/workflow/context-window-management.md` | (empty) | 1 | Absence-red: the required 5.5 row does not exist yet; the criterion demands it | M3 step 1 flips: same grep returns ≥1 |
| RED-AC-001b | `grep -c "Haiku (200K)" .claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | Preserve-red baseline: the row exists today and must still exist after the edit; count 0 post-edit is the failure state | M3 step 1 preserves; re-run exits 0 with `1` |
| RED-AC-002a | `grep -n "Haiku 5.5" .claude/rules/moai/development/prompting-best-practices.md` | (empty) | 1 | Absence-red: the family list still names Haiku 4.5 only | M3 step 3 flips: ≥1 hit |
| RED-AC-002b | `grep -c "Haiku 4.5" .claude/rules/moai/development/prompting-best-practices.md` | `1` | 0 | Wrong-content-red: the stale generation name is present; AC demands 0 | M3 step 3 flips: `0` (exit 1) |
| RED-AC-003a | `grep -n "claude-haiku-5-5" .claude/rules/moai/development/model-policy.md` | (empty) | 1 | Absence-red: the 5.5 model id appears nowhere in the rule | M3 step 2 flips: ≥1 hit |
| RED-AC-003b | `grep -c "Haiku 4.5 supports neither" .claude/rules/moai/development/model-policy.md` | `1` | 0 | Wrong-content-red: the stale effort-exclusivity claim is present; AC demands 0 | M3 step 2 flips: `0` (exit 1) |
| RED-AC-004 | `moai spec lint SPEC-CC-HAIKU55-STATUSLINE-001` | `✓ No findings — all SPEC documents are valid` | 0 | Preserve-red baseline (re-measured on the r2-repaired artifacts — the r1-era `0 error(s), 4 warning(s)` output is superseded and no longer cited): the lint is clean TODAY (0 errors AND 0 warnings); the AC's failure state is any post-M3 lint regression — a warning row included, exit 0 notwithstanding — or any `HaikuResidual`-coded row appearing | M3 step 5 + M5 preserve: the same command still prints the clean-run shape |
| RED-AC-005a | `cmp .claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | (empty) | 0 | Preserve-red baseline: the mirrors are byte-identical TODAY and must STAY byte-identical; this cell pins the today-green baseline the M3 edit must not disturb — any non-zero exit post-edit is the failure state | M3 step 4 preserves: `cmp` still exits 0 |
| RED-AC-005b | `cmp .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md` | (empty) | 0 | Preserve-red baseline: same as RED-AC-005a, model-policy pair | M3 step 4 preserves: exit 0 |
| RED-AC-005c | `cmp .claude/rules/moai/development/prompting-best-practices.md internal/template/templates/.claude/rules/moai/development/prompting-best-practices.md` | (empty) | 0 | Preserve-red baseline: same as RED-AC-005a, BP pair | M3 step 4 preserves: exit 0 |
| RED-AC-007 | `grep -n "subagentStatusLine" internal/template/templates/.claude/settings.json.tmpl` | (empty) | 1 | Absence-red: no subagent statusline wiring exists in the template (measured; only the `statusLine` block at :405) | M2 flips: ≥1 hit |
| RED-AC-008 | `go test ./internal/statusline/... -run '^TestAgentType$' -v` | `testing: warning: no tests to run` + `ok ... [no tests to run]` | 0 | Empty-sweep red: the selector matches zero tests today (no `TestAgentType` function exists) — the AC's swept set is empty and asserts nothing until the test exists (runner's own `[no tests to run]` token cited) | M1 flips: swept set ≥1 function with the four subcases, all pass |
| RED-AC-009 | `grep -n "subagentStatusLine" internal/statusline/types.go` | (empty) | 1 | Absence-red: the payload is not parsed anywhere in the package, so no render path can badge it | M1 flips: the carrier type + tests exist (AC-008's sweep is non-empty) |
| RED-AC-010 | `go test ./internal/statusline/ -cover` | `ok github.com/modu-ai/moai-adk/internal/statusline 45.930s coverage: 90.8% of statements` | 0 | Baseline-red: 90.8% is the pre-M1 figure with zero agentType tests; AC-010's pass judgment needs a post-M1 figure ≥ 90.8% WITH the new tests in the swept set (RED-AC-008's empty sweep shows the current figure measures none of the new code) | M1 flips: coverage re-measured with `^TestAgentType$` tests included, ≥ 90.8% |
| RED-AC-014a | `git diff --stat 81786284e6e5ee2fe5fbe7b498d649b27c3e0109..HEAD -- internal/config/` | (empty) | 0 | Preserve-red baseline: the committed range diff is empty today and MUST stay empty through M5; a non-empty output at M5 is the failure state (commit-range-aware — plain `git diff` is blind to per-milestone commits) | M5 re-runs against the recorded `CARD_BASE_SHA..HEAD`: still empty |
| RED-AC-014b | `git diff --cached --stat -- internal/config/` | (empty) | 0 | Preserve-red baseline: the staged diff is empty today; a non-empty staged `internal/config` change at any milestone exit is the failure state | M5 re-runs: still empty |
| RED-AC-015a | `grep -c "claude-haiku-5-5" docs-site/content/ko/multi-llm/model-policy.md` | `0` | 1 | Absence-red: the canonical-locale page carries no 5.5 model id (the row at :56 is Haiku 4.5 only) | M4 flips: ≥1 |
| RED-AC-015b | `grep -cF '~967K' docs-site/content/ko/multi-llm/model-policy.md` | `0` | 1 | Absence-red: the unit-bearing auto-compact figure is stated on no MoAI surface today | M4 flips: ≥1 |
| RED-AC-015c | `grep -cF '$0.10' docs-site/content/ko/multi-llm/model-policy.md` | `0` | 1 | Absence-red: the Haiku 5.5 input rate (fixed-string) is stated on no MoAI surface today | M4 flips: ≥1 |
| RED-AC-015e | `grep -cF 'over 100K' docs-site/content/ko/multi-llm/model-policy.md` | `0` | 1 | Absence-red: the 100K surcharge threshold is stated on no MoAI surface today | M4 flips: ≥1 |
| RED-AC-015d | `grep -c "2.1.293" .claude/rules/moai/development/model-policy.md` | `0` | 1 | Absence-red: the CC v2.1.293+ requirement is stated nowhere in the rule today (the sonnet/opus entries state their versions; haiku states none) | M3 step 2 flips: ≥1 |
| RED-AC-016 | ``grep -cF 'HaikuResidualRule lint enforces 0 haiku references in agent definitions and the `claude_models` block' .claude/rules/moai/development/model-policy.md`` | `1` | 0 | Preserve-red baseline: the full scope sentence INCLUDING the `claude_models` tail exists at :179 today; the confirmed mutant (audit D7) deletes it and the tail mutant (audit D-R2-6) rewrites the tail — either failure drops the count to 0 | M3 step 5 preserves: count still 1 post-edit |

**Classification note (AC-006, verification-completeness.md §2.1 undecidable disposition).**
AC-006 ("`make build` succeeds after template edits") has no re-executable RED on this tree:
the build is green at baseline (observed: `make build` exit 0, tail `... -o bin/moai ./cmd/moai`,
2026-10-08 @ 81786284e) and its red requires a mutated template, which plan-phase authoring
forbids producing. Per §2.1 the criterion therefore **loses release-blocking eligibility** and is
classified a **regression-guard**: it is re-executed on the edited tree at M3 step 4 and M5
(sequenced first, per plan §F M5) and is never recorded as a pass until that re-execution
observes exit 0. It is NOT counted in the P1 release-blocking set below.

## §D AC Matrix

### AC-001 — context-window threshold table carries both Haiku rows (live + composed rule)

**Given** the edited tree, **When** running
`grep -n "Haiku 5.5 (1M)" .claude/rules/moai/workflow/context-window-management.md`
and `grep -c "Haiku (200K)" .claude/rules/moai/workflow/context-window-management.md`,
**Then** the first returns ≥1 hit with the row reading 1,000,000 tokens / **50%** / ~500,000;
the second returns exactly `1` (the 200K row unchanged at 200,000 / **90%** / ~180,000); and
`grep -n "takes the 200K row" .claude/rules/moai/workflow/context-window-management.md`
returns ≥1 hit within 15 lines after the new row's line number (provider note composes with
the existing rule — AWS-lineage Haiku 4.5 sessions keep 200K/90%).

### AC-002 — prompting BP family updated

**Given** the edited tree, **When** running
`grep -n "Haiku 5.5" .claude/rules/moai/development/prompting-best-practices.md` and
`grep -c "Haiku 4.5" .claude/rules/moai/development/prompting-best-practices.md`,
**Then** the first returns ≥1 hit at the ~:7 family list naming (Opus 5/5.5, Sonnet 5.5,
Haiku 5.5), and the second returns `0`.

### AC-003 — model-policy facts refreshed with No-Haiku anchors preserved

**Given** the edited tree, **When** running these four single commands against
`.claude/rules/moai/development/model-policy.md`:
1. `grep -c "claude-haiku-5-5"` → ≥1
2. `grep -c "retired from MoAI agent routing per the No-Haiku policy"` → ≥1 (DO-NOT-REVERT anchor intact)
3. `grep -c "Haiku 5.5 supports all five effort levels (low/medium/high/xhigh/max), default medium"` → ≥1
4. `grep -c "Haiku 4.5 supports neither"` → `0`

**Then** all four hold. (Condition 3 greps the exact sentence plan §F M3 step 2c inserts —
a line-content expectation, not a region judgment.)

### AC-004 — lint gate green + HaikuResidual scope clean (rewritten: D4/D5)

**Given** the edited tree with per-milestone commits landed, **When** running:
1. `moai spec lint SPEC-CC-HAIKU55-STATUSLINE-001` — the wired lint verb (verified live:
   prints a SEVERITY/CODE/FILE/LINE/MESSAGE table and a `N error(s), M warning(s)` footer;
   `go run . spec lint` was observed to execute nothing on this tree and is never cited again)
2. `git diff --stat $CARD_BASE_SHA..HEAD -- .claude/agents/ .moai/config/` where
   `$CARD_BASE_SHA` is the SHA recorded at pre-flight (plan §C step 0)

**Then** (1) the lint exits 0 with the clean-run output shape `✓ No findings — all SPEC
documents are valid` (the shape the verb actually prints when clean — observed). If the verb
prints a findings table instead, that table is acceptable **only if it carries 0 errors and
0 warnings** — a table with any warning, exit 0 notwithstanding, fails this AC — and zero
rows whose CODE column names the HaikuResidual rule (grep the output for the code stem
`HaikuResidual` → `0`; this greps the structured CODE column, never a file path — the SPEC's
own path contains HAIKU55 and would false-positive a path-text grep); and (2) the
`git diff --stat $CARD_BASE_SHA..HEAD -- .claude/agents/ .moai/config/` output is empty
(no agent-definition or config file modified across the committed range).

### AC-005 — live↔mirror byte parity (3 rules)

**Given** the edited tree, **When** running
`cmp .claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`
and the equivalent `cmp` for `development/model-policy.md` and
`development/prompting-best-practices.md`, **Then** all three exit 0 (byte-identical).

### AC-006 — `make build` succeeds after template edits [regression-guard — see §D.0 note]

**Given** the edited templates (go:embed), **When** running `make build` on the edited tree
(at M3 step 4, and again sequenced FIRST at M5 before any parallel read-only batch),
**Then** the build exits 0. Evidence: verbatim exit status + bounded tail. Baseline observed
green at 81786284e (§D.0). Not release-blocking per the §2.1 undecidable disposition; never
recorded as a pass from the baseline observation alone.

### AC-007 — subagentStatusLine wiring present in the template

**Given** the edited `internal/template/templates/.claude/settings.json.tmpl`, **When**
running `grep -n "subagentStatusLine" internal/template/templates/.claude/settings.json.tmpl`,
**Then** ≥1 hit exists as a sibling block of `statusLine` routing to `.moai/status_line.sh`,
and a rendered settings copy parses as JSON (`jq . <rendered>` exits 0).

### AC-008 — statusline unit tests: agentType present / absent / null (anchored selector)

**Given** `internal/statusline/subagent_agenttype_test.go` exists with the test function
named `TestAgentType` (plan M1), **When** running
`go test ./internal/statusline/... -run '^TestAgentType$' -v` (anchor + legal-Go-test name —
`^AgentType$` matches no legal test, which lacks the `Test` prefix), **Then** the runner
reports a non-empty swept set (no
`[no tests to run]` token) and all tests pass, covering: (a) agentType present → badge
rendered; (b) agentType key absent → no badge, no error; (c) agentType JSON null → no badge,
no error; (d) empty/absent `tasks[]` → render succeeds.

### AC-009 — rendered-statusline observation (payload marking specified)

**Given** the built binary, **When** piping a synthesized subagentStatusLine stdin JSON into
the statusline entry point — the payload carries a `tasks` array whose first row has
`"agentType": "manager-develop"` and whose second row omits the `agentType` key — **Then**
the rendered output shows the agentType badge on the first row and no badge — row still
present — on the second. Evidence: verbatim rendered line(s) captured in the run report.
(The `tasks[]` carrier field name is the canon-named marking; RED-AC-009 records that no
parsing exists today.)

### AC-010 — Go build + package suite green, coverage not regressed

**Given** all Go edits, **When** running `go build ./...` and
`go test ./internal/statusline/ -cover`, **Then** both exit 0 and the post-M1 coverage
figure is ≥ **90.8%** (the baseline pinned in RED-AC-010, measured @ 81786284e with the
agentType tests absent; the post-M1 figure must measure WITH the `^TestAgentType$` tests in the
swept set).

### AC-011 — docs-site Haiku 5.5 rows per page per locale

**Given** the edited docs-site, **When** running per page and locale
`grep -c "claude-haiku-5-5\|Haiku 5.5" docs-site/content/<loc>/<page>.md`, **Then** each of
the 8 canon pages carries ≥1 Haiku 5.5 hit in every locale where the page carries its
Haiku-bearing row after M4 (ko complete in all 8; en/ja/zh follow the ko-canonical chain —
a locale with no corresponding row pre-edit is either updated per the chain or recorded as a
0-hit finding in the run report, never silently skipped).

### AC-012 — docs-site i18n gates: parity via the script + explicit body checks (narrowed: D6)

**Given** all docs-site edits, **When** running:
1. `scripts/docs-i18n-check.sh` — covers file-path parity, frontmatter title, H1, and
   glossary terms ONLY (script read in full: :6-11, :93-241). Expected: exit 0.
2. Body-emoji scan — a perl-based scan (NOT the recipe's `grep -P` form: macOS
   `/usr/bin/grep` rejects `-P` with `grep: invalid option -- P`, exit 2, which an
   empty-output read mistakes for 0 hits). Run once per locale over the pages THIS SPEC
   modifies (the plan M4 page list, 8 pages; `<loc>` ∈ ko/en/ja/zh; plain explicit paths —
   brace/computed forms are guard-refused in a worktree-isolated session):

   `perl -CSD -ne 'print "$ARGV:$.:$_" if /[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]/' docs-site/content/<loc>/multi-llm/model-policy.md docs-site/content/<loc>/multi-llm/_index.md docs-site/content/<loc>/advanced/token-budget.md docs-site/content/<loc>/cost-optimization/prompt-caching.md docs-site/content/<loc>/claude-code/context-memory/context-window.md docs-site/content/<loc>/claude-code/foundations/commands.md docs-site/content/<loc>/claude-code/foundations/how-claude-code-works.md docs-site/content/<loc>/claude-code/_index.md`

   **Execution guard** (report-not-verdict, verification-completeness §1.1): any checker
   usage/option error — exit ≥ 2, or stderr carrying `invalid option` / `usage:` — FAILS the
   check; it is never read as 0 hits. The checker must work identically on macOS and Linux
   (perl one-liner; an `rg` fallback `rg -n '[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]' <files>`
   is acceptable if perl is absent, same guard applies).

   **Scope**: the 32 in-scope files above only. A tree-wide 0-hit is NOT required —
   out-of-scope pages this card does not touch (e.g. statusline output-example pages such as
   `ja/advanced/statusline.md`, which carry emoji by design) are not in this gate.

   **Allow-list** (mirroring the recipe's own allowances) — a scan candidate line passes if
   it is one of: (i) the session-handoff cut-line marker lines (`✂──── ... ────✂`);
   (ii) `Generated with [Claude Code]` attribution lines; (iii) ✓/✗/decorative symbols the
   recipe permits. **Pass condition**: after removing allow-listed lines, the candidate set
   is empty — any emoji-bearing body line in an in-scope page that is none of the three
   classes fails the check (counter-example that flips it red: a `✅` in a new body
   paragraph).

   **Observed baseline (measured on this tree, 2026-10-08, base `81786284e`)**: the scan
   returns exactly 8 candidate lines — the ✂ cut-line markers at
   `ko/advanced/token-budget.md:536,549`, `en/...:512,525`, `ja/...:410,423`,
   `zh/...:396,409` — all allow-list class (i); every other in-scope page is 0-hit. Exit 0
   on all four locale runs.
3. Mermaid direction grep — `grep -rnE "graph (LR|RL)|flowchart (LR|RL)" docs-site/content/`
   (counter-example: a `flowchart LR` block). Expected: exit 1, 0 hits.
4. URL stale-domain blacklist — the hns-oss-docs-verify recipe's URL grep over
   `docs-site/content/`, which blacklists the recipe's named known-stale domains (a
   blacklist, NOT a whitelist — a novel external domain such as example.com is NOT detected
   by this command; the adk.mo.ai.kr whitelist itself remains an i18n-rules authoring
   discipline enforced at review, not by this grep). Counter-example that flips it red: any
   of the recipe's blacklisted stale domains appearing in content. Expected: 0 hits.

**Then** all four hold. (Check 1 alone proves parity, not the three body constraints —
checks 2-4 carry them explicitly.)

### AC-013 — README 4-locale sync

**Given** the README edits, **When** running `grep -n "Haiku 5.5" README.md README.ko.md README.ja.md README.zh.md`,
**Then** every README that gained a Haiku 5.5 row has it in all 4 files at the
section-parity position; the 0-hit escape of §D.5 applies ONLY to this AC's enumerated case
(a README with no Haiku-generation table — plan-time measurement: only
`| Haiku | glm-5.3-flash | 1M |` rows exist).

### AC-014 — config-code invariance, commit-range-aware (rewritten: D5)

**Given** the whole change set committed per-milestone with `CARD_BASE_SHA` recorded at
pre-flight, **When** running
`git diff --stat $CARD_BASE_SHA..HEAD -- internal/config/` and
`git diff --cached --stat -- internal/config/`,
**Then** both outputs are empty (`defaults.go` `DefaultSpeedModel = "haiku"` and
`envkeys.go` untouched across the committed range AND the staging area — plain `git diff`
sees neither).

### AC-015 — REQ-CC-HAIKU55-001 fact bundle stated on MoAI surfaces (new: D2)

**Given** the M3/M4 edits, **When** running each of these single commands:
1. `grep -c "claude-haiku-5-5" .claude/rules/moai/development/model-policy.md` → ≥1 (model id)
2. `grep -c "2.1.293" .claude/rules/moai/development/model-policy.md` → ≥1 (CC version requirement)
3. ``grep -cF 'resolves `haiku` to Haiku 4.5 (200K)' .claude/rules/moai/development/model-policy.md`` → ≥1 (provider split — exact clause plan M3 step 2b inserts; bare backticks, fixed-string match — the r2 `` \` `` escape artifact is repaired; the quoting form is validated on this tree by RED-AC-016, which matches the same file with the same form)
4. ``grep -cF 'disabled with `thinking: {"type": "disabled"}` at `high` effort or below' .claude/rules/moai/development/model-policy.md`` → ≥1 (thinking-disable path — the corrected fact per the official model-config page: Haiku 5.5 thinking CAN be disabled, at `high` effort or below (low/medium/high), effort being the better lever; exact fixed-string fragment of the sentence plan M3 step 2d inserts. The r3-era "adaptive reasoning cannot be disabled" clause was factually wrong and is retracted)
5. `grep -c "claude-haiku-5-5" docs-site/content/ko/multi-llm/model-policy.md` → ≥1 (ko canonical page row)
6. `grep -cF '~967K' docs-site/content/ko/multi-llm/model-policy.md` → ≥1 (auto-compact figure, unit-bearing — a bare `967` mutant such as `967 tokens` does not pass)
7. `grep -cF '$0.10' docs-site/content/ko/multi-llm/model-policy.md` → ≥1 and
   `grep -cF '$0.50' docs-site/content/ko/multi-llm/model-policy.md` → ≥1 and
   `grep -cF '$2.50' docs-site/content/ko/multi-llm/model-policy.md` → ≥1 (both rate pairs, fixed-string — `$0x10`-class dot mutants do not pass) and
   `grep -cF 'over 100K' docs-site/content/ko/multi-llm/model-policy.md` → ≥1 (surcharge threshold pinned)

**Then** all hold (condition 3: the sentence with bare backticks `resolves `haiku` to Haiku 4.5 (200K)`
must appear; conditions 6-7: the ko facts note — which the en/ja/zh chain derives — carries
the unit-bearing auto-compact figure and the base rate pair, the surcharge pair, and the
100K threshold). Conditions 5-7 are scoped to the Haiku 5.5 facts note M4 adds to the ko
page (the greps run against the page file; the matched strings must sit inside that facts
note, not elsewhere on the page). RED-now: RED-AC-015a, b, c, d, e.

### AC-016 — HaikuResidualRule scope sentence preserved at model-policy ~:179 (new: D7)

**Given** the M3 edits, **When** running
``grep -cF 'HaikuResidualRule lint enforces 0 haiku references in agent definitions and the `claude_models` block' .claude/rules/moai/development/model-policy.md``,
**Then** the count is 1 (the exact scope sentence INCLUDING the `claude_models` tail,
observed at :179 in RED-AC-016, survives the lineup refresh verbatim). This closes both
mutants: the deletion mutant (audit D7) and the tail-rewrite mutant (audit D-R2-6) — each
drops the count to 0 — while AC-003 and AC-004 alone would still pass.

## §D.1 Severity

- P1 (release-blocking, RED-now cells in §D.0): AC-001, AC-002, AC-003, AC-004, AC-005,
  AC-007, AC-008, AC-009, AC-010, AC-014, AC-015, AC-016
- P2 (not release-blocking): AC-006 (regression-guard — §2.1 undecidable disposition, §D.0
  note), AC-011, AC-012, AC-013 (locale-chain completeness; §D.5's narrowed escape applies
  only to its two enumerated cases)

## §D.2 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-001 | REQ-CC-HAIKU55-002 | M3 |
| AC-002 | REQ-CC-HAIKU55-004 | M3 |
| AC-003 | REQ-CC-HAIKU55-003 | M3 |
| AC-004 | REQ-CC-HAIKU55-008 | M3 |
| AC-005 | REQ-CC-HAIKU55-005 | M3 |
| AC-006 | REQ-CC-HAIKU55-005 | M3, M5 |
| AC-007 | REQ-CC-HAIKU55-009 | M2 |
| AC-008 | REQ-CC-HAIKU55-010, REQ-CC-HAIKU55-011 | M1 |
| AC-009 | REQ-CC-HAIKU55-010, REQ-CC-HAIKU55-011 | M1 |
| AC-010 | REQ-CC-HAIKU55-010 | M1, M5 |
| AC-011 | REQ-CC-HAIKU55-006 | M4 |
| AC-012 | REQ-CC-HAIKU55-006 | M4 |
| AC-013 | REQ-CC-HAIKU55-007 | M4 |
| AC-014 | (config invariance, spec §1) | M5 |
| AC-015 | REQ-CC-HAIKU55-001 | M3, M4 |
| AC-016 | REQ-CC-HAIKU55-003, REQ-CC-HAIKU55-008 | M3 |

## §D.3 Edge Cases

- `agentType` present but empty string → treated as absent (no badge).
- `tasks[]` with 100+ rows → rendering bounded (follow the segment's existing row-cap
  convention if one exists).
- Locale page where the Haiku row was reworded between plan and edit → re-grep decides
  add-vs-reword; the AC checks the 5.5 fact, not the exact sentence.
- AC-015 condition 3's prescribed clause: if the M3 edit phrases the provider split
  differently, the AC anchor and the plan M3 step 2b sentence are updated TOGETHER in the
  same revision (an AC-anchor/plan-instruction pair, never one alone).

## §D.4 Quality Gates

- TRUST 5: Tested (AC-008/010), Readable (badge naming matches package conventions),
  Unified (gofmt via build), Secured (display-only parsing of the trusted stdin shape,
  pointer-nil decoding), Trackable (conventional commits per milestone).
- Sync gate: `hns-oss-docs-verify` recipe for docs surfaces (AC-012 checks 2 and 4 cite its
  inlined scan steps).

## §D.5 Definition of Done

All P1 ACs green with verbatim command output in the run evidence, each traceable to its
§D.0 RED-now cell; AC-006 re-executed green on the edited tree (never recorded as a pass
from the baseline); P2 ACs AC-011/AC-012 green outright. The only 0-hit escape in this SPEC
is the one AC-013 itself enumerates: a README whose model tables carry no Haiku-generation
row (the finding is recorded in the run report and the AC closes on that record) — and the
AC-011 per-locale case of a locale with no corresponding pre-edit row (chain-updated or
recorded, never silently skipped). REQ-006's ko-canonical completeness is NOT escapable:
AC-011's "ko complete in all 8" has no escape. spec.md transitioned per the ownership
matrix; §E.2/§E.3 populated by the run phase.
