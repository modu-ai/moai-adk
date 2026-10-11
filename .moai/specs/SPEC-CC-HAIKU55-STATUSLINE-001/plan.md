# Implementation Plan — SPEC-CC-HAIKU55-STATUSLINE-001

> Statelessness note: this artifact carries no `status:` field. Lifecycle lives in
> `spec.md` frontmatter alone.

## §A Context

- Card: t1605 (operator-directed, Class C, design-frozen). Branch `WT-haiku-docs-statusline`,
  base `81786284e`, worktree `.moai/worktrees/t1605`.
- Lane context (record): this card was leased **without a factory lease** — the serial slot
  was wedged by an ownerless-lease assigned row (t1595). Leader-dispatched; record settlement
  (`done`, merge, push) is the leader's (t1498 precedent).
- Design canon (absorbed, do not re-research):
  `.moai/reports/release-update-20261008/upstream-update-20261008.md` § GD-1 +
  `2026-10-08-haiku55-subagent-agenttype.html` (specialist report).
- All coordinates below were measured on tree t1538 @ `65e649d5f` and re-verified on this
  worktree (base `81786284e`) at plan time (2026-10-08). **Every edit starts with a re-grep
  of its anchor** — line numbers drift between trees.

## §B Known Issues

- Docs-lag: the official models-overview page lacks a Haiku 5.5 row and the official statusline
  docs omit `agentType` from the `tasks[]` field list (measured 2026-10-08). The CC changelog
  is the canon; official-docs catch-up is tracked as trailing, not blocking.
- Stale canon framing: the t1538 report described model-policy `:18`/`:26` as "alias 설명";
  the measured current text is different (anchors below). The measured lines are the edit
  anchors.
- docs-site grep drift: not every canon-named page/locale pair currently greps a Haiku row
  (plan-time measurement: `advanced/token-budget.md` carries a "Haiku (200K)" threshold-table
  row at ~:49 ko, not a model row; `context-window.md` / `how-claude-code-works.md` /
  `claude-code/_index.md` grep only in ko). Per-locale re-grep at edit time decides add vs
  reword.

## §C Pre-flight (run once, before M1)

0. Record `CARD_BASE_SHA=$(git rev-parse HEAD)` — AC-004/AC-014 measure the committed range
   from this SHA (plain `git diff` is blind to per-milestone commits).
1. `git rev-parse --short HEAD` → confirm base lineage unchanged.
2. Re-grep batch (one turn, parallel Bash):
   - `grep -n "Haiku" .claude/rules/moai/workflow/context-window-management.md` (expect row ~:17)
   - `grep -n -i haiku .claude/rules/moai/development/model-policy.md` (measured 14 lines on
     this tree: :18, :26, :34, :40, :54, :57, :70, :83, :129, :133, :143, :161, :177, :179 —
     a count that differs signals drift; the named anchors below are the load-bearing subset)
   - `sed -n '7p' .claude/rules/moai/development/prompting-best-practices.md`
   - `grep -n "statusLine" internal/template/templates/.claude/settings.json.tmpl` (expect :405 only)
   - `grep -rn agentType internal/statusline/` (expect 0 hits — nothing to collide with)
   - per-locale `grep -in haiku docs-site/content/<loc>/<page>.md` for the 8 canon pages
   - `grep -n "Haiku" README.md README.ko.md README.ja.md README.zh.md`
3. Any anchor that does not match §D's measured text → stop, re-read the surrounding block,
   re-anchor, then edit (drift handling, not a blocker).

## §D Constraints

- GEARS requirements: REQ-CC-HAIKU55-001..011 (spec.md §3) — the milestone maps cite them.
- **No-Haiku policy invariance** threads M3/M4 (coordinator-confirmed): update FACTS (context
  table, lineup, effort, rates) without re-introducing haiku routing. The measured
  DO-NOT-REVERT anchors:
  - model-policy ~:26 — `- haiku = Haiku (current generation; retired from MoAI agent routing
    per the No-Haiku policy — value remains valid for documentation/example YAML)` — the
    policy claim stays verbatim; only the "current generation" framing may name Haiku 5.5.
  - model-policy ~:179 — HaikuResidualRule lint scope (agent definitions + `claude_models`
    block, 0 haiku references) stays untouched.
- Skills loaded by the run-phase worker before any docs-site/README edit:
  `hns-oss-docs-i18n-rules` (hard i18n rules: ko-canonical chains, no body emoji,
  emphasis-marker spacing per locale, Mermaid TD-only, URL whitelist adk.mo.ai.kr),
  `hns-oss-docs-readme-sync` (README 4-file sync procedure), `hns-oss-docs-verify`
  (exit-gate recipe).
- Worktree isolation: all writes stay inside this worktree; relative paths.
- Measured edit anchors (model-policy.md, this tree):
  - `:14-18` alias list — `- haiku: Claude Haiku (fastest, lowest cost)` refreshes to name
    Haiku 5.5.
  - `:20-26` "Current model generation mapping" — the `haiku = ...` entry follows the pattern
    the `sonnet =`/`opus =` entries already model (generation, canonical id, context, price).
  - `:70` — the `[1m]`-breaker historical list ("...and Haiku 4.5, which are 200K outright")
    gains the 5.5 caveat (Anthropic API 1M; AWS-lineage alias keeps 4.5/200K).
  - `:83` — `... (Haiku 4.5 still ships 200K)` → reframe: Haiku 4.5 keeps 200K where it
    remains the alias target (Bedrock / GCP Agent Platform / Foundry); Anthropic API haiku is
    now Haiku 5.5 at 1M.
  - `:161` — effort sentence: `Haiku 4.5 supports neither` → Haiku 5.5 supports all five
    levels, default `medium`.

## §E Self-Verification

Each milestone carries its own verification (AC-XXX in acceptance.md). M1/M2 run
`go test ./internal/statusline/...`; M3/M4 run `make build`, the byte-parity `cmp`, the
per-locale greps, and `scripts/docs-i18n-check.sh`.

## §F Milestones (priority-ordered by decision-reversibility: data-model first, mechanical last)

### M1 — statusline agentType: data model + rendering + tests (Highest — new type interface)

Files (verify names by grep at edit time):
- `internal/statusline/types.go` — add a `SubagentTaskInfo` type (`agentType`, `name`,
  `type`, plus the status fields the payload carries) as pointer-nil optional, matching the
  existing `PRInfo`/`AgentInfo` conventions; add the `tasks[]` carrier to the
  subagentStatusLine input shape (a dedicated struct or an optional pointer field on
  `StdinData` — decide against the actual payload shape the script receives under the
  `subagentStatusLine` event; nil when absent).
- `internal/statusline/builder.go` — thread the tasks payload into the render path.
- Renderer (locate via `grep -rn "render" internal/statusline/*.go | grep -v _test`) — render
  an agentType badge per task row; absent/null `agentType` → no badge, row intact
  (REQ-CC-HAIKU55-010, REQ-CC-HAIKU55-011).
- New test file `internal/statusline/subagent_agenttype_test.go` — the table-driven test
  function is named **`TestAgentType`** (this exact name is what AC-008's
  `-run '^TestAgentType$'` selects) with the four subcases:
  (a) agentType present → badge rendered; (b) agentType key absent → no badge, no error;
  (c) agentType JSON null → no badge, no error; (d) empty/absent tasks[] → render succeeds.
  Run: `go test ./internal/statusline/... -run '^TestAgentType$' -v` (anchored selector on
  the legal Go test name `TestAgentType` — the bare `^AgentType$` form matches no legal test
  because Go test functions carry the `Test` prefix).
- Baseline coverage measured BEFORE this milestone: `go test ./internal/statusline/ -cover`
  (plan-time baseline: 90.8% @ 81786284e — acceptance.md RED-AC-010).
- AC map: AC-008 (REQ-CC-HAIKU55-010, REQ-CC-HAIKU55-011), AC-009 (REQ-CC-HAIKU55-010,
  REQ-CC-HAIKU55-011), AC-010.

### M2 — settings template subagentStatusLine wiring (High)

- `internal/template/templates/.claude/settings.json.tmpl` — add a `subagentStatusLine`
  block beside `statusLine` (~:405), same script form (`.moai/status_line.sh`, with the
  existing windows/non-windows platform split), same `refreshInterval` shape.
- Verify: `make build` succeeds; a rendered settings copy parses as JSON
  (`jq . <rendered>` exits 0).
- AC map: AC-007. (AC-010's coverage re-measure belongs to M1/M5, not here.)

### M3 — rules docs: 3 live files + 3 template mirrors (High — policy-sensitive edits)

1. `.claude/rules/moai/workflow/context-window-management.md` — add the
   `| Haiku 5.5 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |` row beside the
   existing `| Haiku (200K) | 200,000 tokens | **90%** | ~180,000 tokens |` row (~:17) + a
   one-sentence provider note wired into the table's existing rule ("a session that matches
   both a 1M row and the 200K-sessions row takes the 200K row" — AWS-lineage Haiku 4.5
   sessions keep 200K/90%). The new row composes with the rule; it never bypasses it
   (REQ-CC-HAIKU55-002).
2. `.claude/rules/moai/development/model-policy.md` — per the §D anchors: `:14-18` alias
   description, `:20-26` generation-mapping `haiku =` entry, `:70` historical-200K list
   caveat, `:83` "still ships 200K" reframe, `:161` effort sentence (REQ-CC-HAIKU55-003;
   No-Haiku anchors preserved). The entry MUST carry these AC-015 anchor clauses verbatim
   (AC-anchor/plan-instruction pair — update both together if reworded):
   - a. the model id `claude-haiku-5-5` and the requirement `2.1.293` somewhere in the entry;
   - b. the provider-split clause: "... resolves `haiku` to Haiku 4.5 (200K)" (naming that
     AWS Bedrock / GCP Agent Platform / Foundry keep Haiku 4.5);
   - c. the effort sentence: "Haiku 5.5 supports all five effort levels
     (low/medium/high/xhigh/max), default medium" (replaces "Haiku 4.5 supports neither");
   - d. the thinking clause (CORRECTED FACT — replaces the r3-era "adaptive reasoning
     cannot be disabled" clause, which is factually wrong per the official model-config
     page): insert the exact sentence "Haiku 5.5's adaptive reasoning can be disabled with
     `thinking: {"type": "disabled"}` at `high` effort or below (low/medium/high); effort is
     the better lever for controlling reasoning depth." — AC-015 condition 4 greps the
     fixed-string fragment `disabled with `thinking: {"type": "disabled"}` at `high` effort
     or below`, so this sentence's wording is AC-anchor-paired (update both together if
     reworded). Insert the sentence as ONE line — the grep is line-based, so a wrap inside
     the fragment breaks the match (validated: the fragment matches single-line content,
     misses multi-line wraps).
3. `.claude/rules/moai/development/prompting-best-practices.md:7` — family list →
   (Opus 5/5.5, Sonnet 5.5, Haiku 5.5) (REQ-CC-HAIKU55-004).
4. Mirror all three byte-identically into
   `internal/template/templates/.claude/rules/moai/{workflow/context-window-management.md,development/model-policy.md,development/prompting-best-practices.md}`,
   then `make build` (go:embed) (REQ-CC-HAIKU55-005; AC-006 regression-guard re-execution).
5. No-Haiku invariance check (REQ-CC-HAIKU55-008, REQ-CC-HAIKU55-003): run
   `moai spec lint SPEC-CC-HAIKU55-STATUSLINE-001` — 0 errors, 0 warnings, zero
   `HaikuResidual`-coded rows (AC-004); `git diff --stat $CARD_BASE_SHA..HEAD -- .claude/agents/ .moai/config/`
   empty (committed range, not working tree); model-policy :26 policy phrase AND the :179
   HaikuResidualRule scope sentence both still grep ≥1 (AC-003 cond 2, AC-016).
- AC map: AC-001, AC-002, AC-003, AC-004, AC-005, AC-006, AC-015, AC-016.

### M4 — docs-site 8 pages x 4 locales + README 4-locale + exit gates (Medium — mechanical once M3 facts are frozen)

1. Load `hns-oss-docs-i18n-rules` + `hns-oss-docs-readme-sync` before the first edit.
2. Per page, per locale (ko first — canonical chain ko → en → ja/zh): re-grep the Haiku
   anchor, add or reword to include the Haiku 5.5 row (`claude-haiku-5-5`, 1M,
   provider-split alias, v2.1.293+) (REQ-CC-HAIKU55-006). Plan-time measured anchors:
   - `multi-llm/model-policy.md` (~:56 ko model row) — 4/4 locales grep a Haiku row; the ko
     page ALSO gains a compact Haiku 5.5 facts note (REQ-CC-HAIKU55-001, AC-015): auto-compact
     ~967K, rates input $0.10 / output $0.50 per Mtok with the surcharge pair over 100K:
     $0.50 / $2.50, provider-split alias — the en/ja/zh chain derives it (the phrases
     `~967K`, `$0.10`, `$0.50`, `$2.50`, `over 100K` are AC-015's grep anchors — keep them
     verbatim)
   - `multi-llm/_index.md` — 4/4 locales
   - `advanced/token-budget.md` (~:49 ko "Haiku (200K)" threshold row) — the threshold table
     gets the 5.5 row mirroring the M3-1 rule edit; 0 model-row hits today
   - `cost-optimization/prompt-caching.md` (~:155 ko) — 4/4 locales
   - `claude-code/context-memory/context-window.md` (~:57 ko model row) — ko only; other
     locales re-grep and follow the ko chain
   - `claude-code/foundations/commands.md` (~:94 ko) — ko+en hits; ja/zh re-grep
   - `claude-code/foundations/how-claude-code-works.md` (~:65 ko) — ko only; re-grep
   - `claude-code/_index.md` (~:21 ko) — ko only; re-grep
3. README 4-locale set (REQ-CC-HAIKU55-007): plan-time grep finds only
   `| Haiku | glm-5.3-flash | 1M |` GLM-alias rows (~:698 ko) — re-grep for a
   "Haiku 4.5"/`claude-haiku-4-5` generation row at edit time; where a model table names a
   Haiku generation, add the Haiku 5.5 row; where none exists, record the 0-hit finding in
   the run report instead of inventing a row. Keep section-order parity across the 4 files
   (`hns-oss-docs-readme-sync` checklist).
4. Exit gates: `scripts/docs-i18n-check.sh` → exit 0 (this validates parity/title/H1/glossary
   ONLY); PLUS the three explicit body checks of acceptance AC-012 — the
   `hns-oss-docs-verify` emoji scan, the Mermaid-direction grep
   (`grep -rnE "graph (LR|RL)|flowchart (LR|RL)" docs-site/content/` → 0 hits), and the URL
   whitelist/blacklist grep — each with its 0-hit expectation; hugo build warning-free per
   `hns-oss-docs-verify` where the site is built locally.
- AC map: AC-011, AC-012, AC-013, AC-015 (ko facts note + chain derivation).

### M5 — close-out (Medium — build first, then the read-only batch)

1. `make build` runs FIRST, alone (build class — it regenerates templ/catalog sources the
   batch reads; not batch-safe with them per verification-batch-pattern.md).
2. Then one parallel read-only verification turn: `go test ./internal/statusline/ -cover`,
   `go build ./...`, 3x `cmp` byte-parity, per-locale greps, `scripts/docs-i18n-check.sh`,
   AC-004's lint + committed-range checks (`git diff --stat $CARD_BASE_SHA..HEAD -- internal/config/`
   AND `git diff --cached --stat -- internal/config/`, both empty), AC-016's :179 count.
   Conventional commits per the transition matrix; run-phase commits follow
   `feat(SPEC-CC-HAIKU55-STATUSLINE-001): M<n> ...` with the `Authored-By-Agent:` trailer.
- AC map: AC-014, AC-010 (coverage re-measure), AC-006 (re-execution).

## §G Anti-Patterns

- Editing a target surface from plan phase (this plan only reads and records).
- Trusting a line number without re-grepping it at edit time.
- Re-wording the No-Haiku policy claims while refreshing lineup facts (M3 guard; §D anchors).
- Adding a Haiku 5.5 row to a README/docs table that carries no Haiku generation row
  (invented content — record the 0-hit finding instead).
- Emoji in docs-site bodies; Mermaid `LR`/`RL`; non-whitelisted URLs (i18n rules skill).
- Parsing `agentType` with a non-pointer field (breaks the absent-key nil fallback).
- `go test ./...` full-suite in the card worktree (structural red — affected packages + CI
  only, per lane verification-load doctrine).

## §H Cross-References

- spec.md §3 REQ-CC-HAIKU55-001..011 · acceptance.md AC-001..016
- Skills: `hns-oss-docs-i18n-rules`, `hns-oss-docs-readme-sync`, `hns-oss-docs-verify`
- Related cards/SPECs: SPEC-AGENT-ARCH-V2-001 (No-Haiku policy), t1600 (future badge consumer)
